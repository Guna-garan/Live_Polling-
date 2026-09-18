package vote

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"livepoll/internal/poll"
	"livepoll/internal/realtime"
)

var (
	ErrInvalidVoterID = errors.New("voterId is required")
	ErrPollClosed     = errors.New("this poll is closed")
	ErrPollExpired    = errors.New("this poll has ended")
	ErrInvalidOption  = errors.New("optionId does not belong to this poll")
)

type Service struct {
	repo     *Repository
	polls    *poll.Repository
	counters *realtime.Counters
}

func NewService(repo *Repository, polls *poll.Repository, counters *realtime.Counters) *Service {
	return &Service{repo: repo, polls: polls, counters: counters}
}

// Cast implements the full voting pipeline from section 13/37 of the
// spec, in strict order:
//  1. Load the poll and validate it exists, is active, and isn't
//     expired (lazily closing it if it just expired).
//  2. Validate the option belongs to the poll and a voter ID was given.
//  3. Insert the vote into MongoDB — the unique index is what actually
//     stops duplicate votes, even under a race.
//  4. Only once that insert succeeds: HINCRBY the Redis counter and
//     publish the update event.
//
// If any validation step fails, or the Mongo insert is rejected as a
// duplicate, Redis is never touched — an invalid or duplicate vote must
// never move the counters.
func (s *Service) Cast(ctx context.Context, pollID primitive.ObjectID, req CastVoteRequest) (*CastVoteResponse, error) {
	voterID := strings.TrimSpace(req.VoterID)
	if voterID == "" {
		return nil, ErrInvalidVoterID
	}

	p, err := s.polls.FindByID(ctx, pollID)
	if err != nil {
		return nil, err
	}

	if p.Status != poll.StatusActive {
		return nil, ErrPollClosed
	}
	if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
		_ = s.polls.MarkExpired(ctx, pollID)
		return nil, ErrPollExpired
	}
	if !p.HasOption(req.OptionID) {
		return nil, ErrInvalidOption
	}

	v := &Vote{PollID: pollID, OptionID: req.OptionID, VoterID: voterID}
	if err := s.repo.Insert(ctx, v); err != nil {
		return nil, err // ErrDuplicateVote or a real DB error; Redis untouched either way
	}

	results, err := s.counters.Increment(ctx, pollID.Hex(), req.OptionID)
	if err != nil {
		return nil, err
	}

	for _, opt := range p.Options {
		if _, ok := results[opt.ID]; !ok {
			results[opt.ID] = 0
		}
	}

	if err := s.counters.Publish(ctx, pollID.Hex(), results); err != nil {
		return nil, err
	}

	var total int64
	for _, v := range results {
		total += v
	}
	return &CastVoteResponse{Results: results, TotalVotes: total}, nil
}

func (s *Service) HasVoted(ctx context.Context, pollID primitive.ObjectID, voterID string) (bool, error) {
	return s.repo.HasVoted(ctx, pollID, voterID)
}

// CountByOptionAdapter exposes vote counting from MongoDB in the exact
// function shape poll.Service.RebuildActiveCounters expects, without
// poll needing to import vote directly (which would create an import
// cycle, since vote already imports poll).
func (s *Service) CountByOptionAdapter(ctx context.Context, pollID primitive.ObjectID) (map[string]int64, error) {
	return s.repo.CountByOption(ctx, pollID)
}
