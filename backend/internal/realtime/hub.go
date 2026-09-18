package realtime

import (
	"context"
	"encoding/json"
	"log"
	"sync"
)

// client is one connected WebSocket browser tab. send is buffered so a
// slow reader doesn't block the broadcaster; if it ever fills up we drop
// the client rather than let one bad connection back up the whole hub.
type client struct {
	pollID string
	send   chan []byte
}

// pollRoom holds every client currently watching one poll, plus the
// single shared Redis subscription for that poll's update channel. We
// intentionally keep exactly ONE Redis subscription per poll no matter
// how many browsers are watching it — never one Redis connection per
// WebSocket client — per the spec's reliability requirements.
type pollRoom struct {
	clients map[*client]bool
	cancel  context.CancelFunc
}

// Hub fans out Redis Pub/Sub events to WebSocket clients, grouped by
// poll ID. It is the component described in section 17: React <->
// WebSocket <-> Go Hub <-> Redis Pub/Sub.
type Hub struct {
	mu       sync.Mutex
	rooms    map[string]*pollRoom
	counters *Counters
}

func NewHub(counters *Counters) *Hub {
	return &Hub{
		rooms:    make(map[string]*pollRoom),
		counters: counters,
	}
}

// Join registers a new client for a poll, starting the poll's shared
// Redis subscription if this is the first client watching it.
func (h *Hub) Join(pollID string) *client {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[pollID]
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		room = &pollRoom{clients: make(map[*client]bool), cancel: cancel}
		h.rooms[pollID] = room
		go h.subscribeRoom(ctx, pollID, room)
	}

	c := &client{pollID: pollID, send: make(chan []byte, 16)}
	room.clients[c] = true
	return c
}

// Leave removes a client, and — if it was the last one watching that
// poll — tears down the poll's Redis subscription so we don't leak
// goroutines or connections for polls nobody is watching anymore.
func (h *Hub) Leave(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[c.pollID]
	if !ok {
		return
	}
	delete(room.clients, c)
	close(c.send)

	if len(room.clients) == 0 {
		room.cancel()
		delete(h.rooms, c.pollID)
	}
}

// subscribeRoom runs for the lifetime of a poll room: it reads messages
// from the poll's Redis Pub/Sub channel and fans each one out to every
// currently-joined client for that poll.
func (h *Hub) subscribeRoom(ctx context.Context, pollID string, room *pollRoom) {
	sub := h.counters.Subscribe(ctx, pollID)
	defer sub.Close()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			h.broadcast(room, []byte(msg.Payload))
		}
	}
}

func (h *Hub) broadcast(room *pollRoom, payload []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for c := range room.clients {
		select {
		case c.send <- payload:
		default:
			// Client is too slow to keep up; drop it rather than block
			// every other client in the room.
			log.Println("realtime: dropping slow client")
			delete(room.clients, c)
			close(c.send)
		}
	}
}

// GetActiveWatchers returns the current number of connected WebSocket clients for a poll.
func (h *Hub) GetActiveWatchers(pollID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[pollID]
	if !ok {
		return 0
	}
	return len(room.clients)
}

// SnapshotMessage builds a one-off "current state" message in the same
// envelope as a live update, used to seed a client the instant it
// connects (and after it reconnects) so it never shows stale zeros
// while waiting for the next vote.
func SnapshotMessage(results map[string]int64, activeWatchers ...int) ([]byte, error) {
	var total int64
	for _, v := range results {
		total += v
	}
	watchers := 1
	if len(activeWatchers) > 0 {
		watchers = activeWatchers[0]
	}
	return json.Marshal(map[string]interface{}{
		"type":           "poll.results.snapshot",
		"results":        results,
		"totalVotes":     total,
		"activeWatchers": watchers,
	})
}

