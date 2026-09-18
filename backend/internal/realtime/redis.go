package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Counters wraps Redis and is the ONLY place in the codebase that talks
// to Redis for vote counting and pub/sub. Keeping it isolated here (per
// the architecture in section 30 of the spec) means every atomic
// increment and every published event goes through one well-tested path.
type Counters struct {
	rdb *redis.Client
}

func NewCounters(rdb *redis.Client) *Counters {
	return &Counters{rdb: rdb}
}

func resultsKey(pollID string) string {
	return fmt.Sprintf("poll:%s:results", pollID)
}

func updatesChannel(pollID string) string {
	return fmt.Sprintf("poll:%s:updates", pollID)
}

// InitCounters creates the results hash for a brand-new poll with every
// option starting at zero. HSETNX-style behavior (via HSET only if
// absent) means calling this twice is harmless.
func (c *Counters) InitCounters(ctx context.Context, pollID string, optionIDs []string) error {
	key := resultsKey(pollID)
	pipe := c.rdb.Pipeline()
	for _, id := range optionIDs {
		pipe.HSetNX(ctx, key, id, 0)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Increment atomically increments a single option's vote count using
// HINCRBY — the whole point of routing votes through Redis instead of a
// naive read-modify-write, since HINCRBY is atomic even under heavy
// concurrent voting.
func (c *Counters) Increment(ctx context.Context, pollID, optionID string) (map[string]int64, error) {
	if _, err := c.rdb.HIncrBy(ctx, resultsKey(pollID), optionID, 1).Result(); err != nil {
		return nil, err
	}
	return c.GetAll(ctx, pollID)
}

// GetAll returns the current counts for every option in a poll using
// HGETALL.
func (c *Counters) GetAll(ctx context.Context, pollID string) (map[string]int64, error) {
	raw, err := c.rdb.HGetAll(ctx, resultsKey(pollID)).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(raw))
	for k, v := range raw {
		var n int64
		fmt.Sscanf(v, "%d", &n)
		out[k] = n
	}
	return out, nil
}

// Rebuild overwrites the Redis hash for a poll with counts computed from
// MongoDB. Used at startup (and lazily on first access) so that if Redis
// ever restarts and loses its data, live counters recover exactly from
// the persistent source of truth.
func (c *Counters) Rebuild(ctx context.Context, pollID string, counts map[string]int64, optionIDs []string) error {
	key := resultsKey(pollID)
	pipe := c.rdb.TxPipeline()
	pipe.Del(ctx, key)
	for _, id := range optionIDs {
		pipe.HSet(ctx, key, id, counts[id])
	}
	_, err := pipe.Exec(ctx)
	return err
}

// VoteUpdateEvent is the message published to a poll's update channel
// and broadcast verbatim (after being read back by the subscriber) to
// every connected WebSocket client.
type VoteUpdateEvent struct {
	Type       string           `json:"type"`
	PollID     string           `json:"pollId"`
	Results    map[string]int64 `json:"results"`
	TotalVotes int64            `json:"totalVotes"`
	Timestamp  time.Time        `json:"timestamp"`
}

// Publish sends a vote-update event on the poll's Pub/Sub channel. This
// is only ever called AFTER the vote has been durably persisted in
// MongoDB and the Redis counter has been incremented — never before,
// so a published event always corresponds to a real, accepted vote.
func (c *Counters) Publish(ctx context.Context, pollID string, results map[string]int64) error {
	var total int64
	for _, v := range results {
		total += v
	}
	evt := VoteUpdateEvent{
		Type:       "poll.results.updated",
		PollID:     pollID,
		Results:    results,
		TotalVotes: total,
		Timestamp:  time.Now().UTC(),
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return c.rdb.Publish(ctx, updatesChannel(pollID), payload).Err()
}

// Subscribe returns a Redis Pub/Sub subscription for a single poll's
// update channel. The caller (the realtime hub) owns its lifecycle and
// must Close() it when the last WebSocket client for that poll leaves.
func (c *Counters) Subscribe(ctx context.Context, pollID string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, updatesChannel(pollID))
}
