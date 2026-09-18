package realtime

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// These tests exercise the real Redis atomic-increment and Pub/Sub path
// described in the spec (HINCRBY + publish). They require a reachable
// Redis and are skipped otherwise, so `go test ./...` still passes in
// environments without infrastructure running (e.g. plain CI without
// docker-compose up). Set TEST_REDIS_URL to run them, e.g.:
//
//	TEST_REDIS_URL=redis://localhost:6379/15 go test ./internal/realtime/...
func testRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set; skipping Redis integration test")
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("bad TEST_REDIS_URL: %v", err)
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not reachable at %s: %v", url, err)
	}
	return client
}

func TestConcurrentIncrementsAreAtomic(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	ctx := context.Background()

	c := NewCounters(client)
	pollID := "test-concurrent-poll"
	defer client.Del(ctx, "poll:"+pollID+":results")

	if err := c.InitCounters(ctx, pollID, []string{"a"}); err != nil {
		t.Fatalf("init: %v", err)
	}

	const n = 100
	done := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			_, err := c.Increment(ctx, pollID, "a")
			done <- err
		}()
	}
	for i := 0; i < n; i++ {
		if err := <-done; err != nil {
			t.Fatalf("increment: %v", err)
		}
	}

	results, err := c.GetAll(ctx, pollID)
	if err != nil {
		t.Fatalf("getall: %v", err)
	}
	// This is exactly the property a naive read-modify-write would
	// violate under concurrency; HINCRBY guarantees it holds exactly.
	if results["a"] != n {
		t.Fatalf("got %d concurrent votes counted, want %d — HINCRBY should be atomic", results["a"], n)
	}
}

func TestPublishDeliversToSubscriber(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()
	ctx := context.Background()

	c := NewCounters(client)
	pollID := "test-pubsub-poll"

	sub := c.Subscribe(ctx, pollID)
	defer sub.Close()

	// Give the subscription a moment to actually register with Redis
	// before we publish, otherwise the message can be missed.
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	if err := c.Publish(ctx, pollID, map[string]int64{"a": 1}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case msg := <-sub.Channel():
		if msg.Payload == "" {
			t.Fatal("expected non-empty published payload")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for published vote-update event")
	}
}
