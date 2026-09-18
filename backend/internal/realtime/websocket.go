package realtime

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 // clients never need to send us more than a ping
)

// upgrader is configured to only accept requests from the configured
// frontend origin, mirroring the CORS policy on the REST API.
func newUpgrader(allowedOrigin string) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "" || origin == allowedOrigin
		},
	}
}

// Handler exposes GET /api/polls/:id/ws. Each connection:
//  1. Fetches the current snapshot and sends it immediately, so the
//     client never renders stale/zero results while waiting for the
//     next vote.
//  2. Joins the poll's hub room (creating the shared Redis subscription
//     if it's the first watcher).
//  3. Pumps hub messages out to the socket, and runs a ping/pong
//     heartbeat with read/write deadlines so dead connections are
//     cleaned up instead of leaking.
type Handler struct {
	hub      *Hub
	counters *Counters
	upgrader websocket.Upgrader
}

func NewHandler(hub *Hub, counters *Counters, allowedOrigin string) *Handler {
	return &Handler{hub: hub, counters: counters, upgrader: newUpgrader(allowedOrigin)}
}

func (h *Handler) Serve(c *gin.Context) {
	pollID := c.Param("id")

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return // Upgrade already wrote an HTTP error response.
	}

	cl := h.hub.Join(pollID)
	defer h.hub.Leave(cl)

	// Send an initial snapshot from Redis so the client is never stale
	// between connecting and the first live event.
	if results, err := h.counters.GetAll(context.Background(), pollID); err == nil {
		if payload, err := SnapshotMessage(results); err == nil {
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			_ = conn.WriteMessage(websocket.TextMessage, payload)
		}
	}

	done := make(chan struct{})
	go h.readPump(conn, done)
	h.writePump(conn, cl, done)
}

// readPump only exists to process pong frames (resetting the read
// deadline) and detect client disconnects; LivePoll's protocol is
// server-to-client only, so any inbound text message is discarded.
func (h *Handler) readPump(conn *websocket.Conn, done chan struct{}) {
	defer close(done)
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

// writePump owns all writes to the connection (gorilla/websocket
// requires a single writer goroutine per connection) and sends periodic
// pings to detect dead connections that never sent a close frame.
func (h *Handler) writePump(conn *websocket.Conn, cl *client, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg, ok := <-cl.send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}
