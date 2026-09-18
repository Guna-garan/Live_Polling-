package poll

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"livepoll/internal/realtime"
)

var ErrForbidden = errors.New("you do not own this poll")

type VoteCounterFn func(context.Context, primitive.ObjectID) (map[string]int64, error)

type Service struct {
	repo        *Repository
	counters    *realtime.Counters
	voteCounter VoteCounterFn
}

func NewService(repo *Repository, counters *realtime.Counters) *Service {
	return &Service{repo: repo, counters: counters}
}

func (s *Service) SetVoteCounter(fn VoteCounterFn) {
	s.voteCounter = fn
}

func (s *Service) Create(ctx context.Context, creatorID primitive.ObjectID, req CreatePollRequest) (*Poll, error) {
	n, err := validateCreate(req)
	if err != nil {
		return nil, err
	}

	options := make([]Option, len(n.Options))
	optionIDs := make([]string, len(n.Options))
	for i, text := range n.Options {
		id := uuid.NewString()
		options[i] = Option{ID: id, Text: text}
		optionIDs[i] = id
	}

	p := &Poll{
		CreatorID:   creatorID,
		Question:    n.Question,
		Description: n.Description,
		Options:     options,
		Status:      StatusActive,
		ExpiresAt:   n.ExpiresAt,
		MaxVotes:    n.MaxVotes,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	if err := s.counters.InitCounters(ctx, p.ID.Hex(), optionIDs); err != nil {
		return nil, err
	}

	return p, nil
}

// Get fetches a poll, lazily closing it first if it has expired, and
// attaches live results read from Redis (rebuilding them from MongoDB
// first if Redis has no data for this poll — e.g. after a Redis
// restart).
func (s *Service) Get(ctx context.Context, id primitive.ObjectID) (*PollWithResults, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if p.Status == StatusActive && p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
		_ = s.repo.MarkExpired(ctx, id)
		p.Status = StatusClosed
	}

	results, err := s.resultsFor(ctx, p)
	if err != nil {
		return nil, err
	}

	var total int64
	for _, v := range results {
		total += v
	}

	return &PollWithResults{Poll: *p, Results: results, TotalVotes: total}, nil
}

// resultsFor returns live counts for a poll, rebuilding Redis from
// MongoDB votes if the hash is missing/empty (Redis restart recovery).
func (s *Service) resultsFor(ctx context.Context, p *Poll) (map[string]int64, error) {
	results, err := s.counters.GetAll(ctx, p.ID.Hex())
	if err != nil {
		return nil, err
	}

	var total int64
	for _, v := range results {
		total += v
	}

	if (len(results) == 0 || total == 0) && len(p.Options) > 0 {
		if s.voteCounter != nil {
			if counts, err := s.voteCounter(ctx, p.ID); err == nil && len(counts) > 0 {
				var mongoTotal int64
				for _, v := range counts {
					mongoTotal += v
				}
				if mongoTotal > 0 {
					optionIDs := make([]string, len(p.Options))
					for i, o := range p.Options {
						optionIDs[i] = o.ID
						if _, ok := counts[o.ID]; !ok {
							counts[o.ID] = 0
						}
					}
					_ = s.counters.Rebuild(ctx, p.ID.Hex(), counts, optionIDs)
					return counts, nil
				}
			}
		}
		if len(results) == 0 {
			results = make(map[string]int64, len(p.Options))
			for _, o := range p.Options {
				results[o.ID] = 0
			}
			optionIDs := make([]string, len(p.Options))
			for i, o := range p.Options {
				optionIDs[i] = o.ID
			}
			_ = s.counters.InitCounters(ctx, p.ID.Hex(), optionIDs)
		}
	}

	for _, o := range p.Options {
		if _, ok := results[o.ID]; !ok {
			results[o.ID] = 0
		}
	}
	return results, nil
}

func (s *Service) ListByCreator(ctx context.Context, creatorID primitive.ObjectID) ([]PollWithResults, error) {
	polls, err := s.repo.FindByCreator(ctx, creatorID)
	if err != nil {
		return nil, err
	}

	out := make([]PollWithResults, 0, len(polls))
	for _, p := range polls {
		results, err := s.resultsFor(ctx, &p)
		if err != nil {
			return nil, err
		}
		var total int64
		for _, v := range results {
			total += v
		}
		out = append(out, PollWithResults{Poll: p, Results: results, TotalVotes: total})
	}
	return out, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id, requesterID primitive.ObjectID, status string) (*Poll, error) {
	if status != StatusActive && status != StatusClosed {
		return nil, errors.New("invalid status")
	}
	p, err := s.repo.UpdateStatus(ctx, id, requesterID, status)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID primitive.ObjectID) error {
	return s.repo.Delete(ctx, id, requesterID)
}

// RebuildActiveCounters recomputes every active poll's Redis hash from
// MongoDB vote counts. Called once at startup (section 36) so a Redis
// restart never silently loses live counters.
func (s *Service) RebuildActiveCounters(ctx context.Context, countVotes func(context.Context, primitive.ObjectID) (map[string]int64, error)) error {
	polls, err := s.repo.FindActive(ctx)
	if err != nil {
		return err
	}
	for _, p := range polls {
		counts, err := countVotes(ctx, p.ID)
		if err != nil {
			return err
		}
		optionIDs := make([]string, len(p.Options))
		for i, o := range p.Options {
			optionIDs[i] = o.ID
		}
		if err := s.counters.Rebuild(ctx, p.ID.Hex(), counts, optionIDs); err != nil {
			return err
		}
	}
	return nil
}
