package vote

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"livepoll/internal/poll"
	"livepoll/internal/realtime"
)

// setupTestService spins up service instances against real MongoDB and
// Redis, using dedicated collections/keys per test poll ID so tests
// don't collide. Requires TEST_MONGO_URI and TEST_REDIS_URL; skipped
// otherwise (see internal/realtime/redis_integration_test.go for the
// rationale). Run with:
//
//	TEST_MONGO_URI=mongodb://localhost:27017 TEST_REDIS_URL=redis://localhost:6379/15 \
//	  go test ./internal/vote/...
func setupTestService(t *testing.T) (*Service, *poll.Repository, func()) {
	t.Helper()
	mongoURI := os.Getenv("TEST_MONGO_URI")
	redisURL := os.Getenv("TEST_REDIS_URL")
	if mongoURI == "" || redisURL == "" {
		t.Skip("TEST_MONGO_URI / TEST_REDIS_URL not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("mongo connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("mongo not reachable: %v", err)
	}
	db := client.Database("livepoll_test")

	rOpt, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("bad redis url: %v", err)
	}
	rdb := redis.NewClient(rOpt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not reachable: %v", err)
	}

	pollRepo := poll.NewRepository(db)
	voteRepo := NewRepository(db)
	counters := realtime.NewCounters(rdb)
	svc := NewService(voteRepo, pollRepo, counters)

	cleanup := func() {
		_ = db.Drop(context.Background())
		rdb.Close()
		client.Disconnect(context.Background())
	}
	return svc, pollRepo, cleanup
}

func TestCastVote_ValidAndDuplicate(t *testing.T) {
	svc, pollRepo, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	p := &poll.Poll{Question: "Test?", Status: poll.StatusActive,
		Options: []poll.Option{{ID: "opt-a", Text: "A"}, {ID: "opt-b", Text: "B"}}}
	if err := pollRepo.Create(ctx, p); err != nil {
		t.Fatalf("create poll: %v", err)
	}

	if _, err := svc.Cast(ctx, p.ID, CastVoteRequest{OptionID: "opt-a", VoterID: "voter-1"}); err != nil {
		t.Fatalf("first vote should succeed: %v", err)
	}

	if _, err := svc.Cast(ctx, p.ID, CastVoteRequest{OptionID: "opt-a", VoterID: "voter-1"}); err != ErrDuplicateVote {
		t.Fatalf("expected ErrDuplicateVote, got %v", err)
	}

	if _, err := svc.Cast(ctx, p.ID, CastVoteRequest{OptionID: "does-not-exist", VoterID: "voter-2"}); err != ErrInvalidOption {
		t.Fatalf("expected ErrInvalidOption, got %v", err)
	}
}

func TestCastVote_ClosedPollRejected(t *testing.T) {
	svc, pollRepo, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	p := &poll.Poll{Question: "Test?", Status: poll.StatusClosed,
		Options: []poll.Option{{ID: "opt-a", Text: "A"}}}
	if err := pollRepo.Create(ctx, p); err != nil {
		t.Fatalf("create poll: %v", err)
	}

	if _, err := svc.Cast(ctx, p.ID, CastVoteRequest{OptionID: "opt-a", VoterID: "voter-1"}); err != ErrPollClosed {
		t.Fatalf("expected ErrPollClosed, got %v", err)
	}
}

func TestCastVote_ExpiredPollRejected(t *testing.T) {
	svc, pollRepo, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	past := time.Now().Add(-time.Minute)
	p := &poll.Poll{Question: "Test?", Status: poll.StatusActive, ExpiresAt: &past,
		Options: []poll.Option{{ID: "opt-a", Text: "A"}}}
	if err := pollRepo.Create(ctx, p); err != nil {
		t.Fatalf("create poll: %v", err)
	}

	if _, err := svc.Cast(ctx, p.ID, CastVoteRequest{OptionID: "opt-a", VoterID: "voter-1"}); err != ErrPollExpired {
		t.Fatalf("expected ErrPollExpired, got %v", err)
	}
}

// TestConcurrentVotingFromSameVoter is the "concurrent voting" case
// called out explicitly in the spec's test list (section 40): many
// simultaneous requests from the same voterId must result in exactly
// one accepted vote, enforced by MongoDB's unique index rather than a
// racy application-level check.
func TestConcurrentVotingFromSameVoter(t *testing.T) {
	svc, pollRepo, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	p := &poll.Poll{Question: "Test?", Status: poll.StatusActive,
		Options: []poll.Option{{ID: "opt-a", Text: "A"}}}
	if err := pollRepo.Create(ctx, p); err != nil {
		t.Fatalf("create poll: %v", err)
	}

	const n = 20
	var wg sync.WaitGroup
	successes := make(chan bool, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Cast(ctx, p.ID, CastVoteRequest{OptionID: "opt-a", VoterID: "same-voter"})
			successes <- err == nil
		}()
	}
	wg.Wait()
	close(successes)

	successCount := 0
	for ok := range successes {
		if ok {
			successCount++
		}
	}
	if successCount != 1 {
		t.Fatalf("expected exactly 1 accepted vote out of %d concurrent attempts, got %d", n, successCount)
	}
}
