package poll

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func apiError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

// Create handles POST /api/polls. Requires auth; the creator ID always
// comes from the authenticated session, never from the request body.
func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(primitive.ObjectID)

	var req CreatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "Request body is malformed")
		return
	}

	p, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		apiError(c, http.StatusUnprocessableEntity, "INVALID_REQUEST", err.Error())
		return
	}

	c.JSON(http.StatusCreated, p)
}

// List handles GET /api/polls — the authenticated creator's own polls
// (the dashboard). It is intentionally not a public "browse all polls"
// endpoint.
func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet("userID").(primitive.ObjectID)

	polls, err := h.svc.ListByCreator(c.Request.Context(), userID)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not load polls")
		return
	}
	c.JSON(http.StatusOK, polls)
}

// Get handles GET /api/polls/:id — public, no auth required, so anyone
// with the link can view an active poll's live results.
func (h *Handler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid poll id")
		return
	}

	p, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrPollNotFound) {
			apiError(c, http.StatusNotFound, "POLL_NOT_FOUND", "Poll not found")
			return
		}
		apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not load poll")
		return
	}
	c.JSON(http.StatusOK, p)
}

// Update handles PATCH /api/polls/:id (currently: open/close a poll).
// Ownership is verified server-side via UpdateStatus's creatorID filter
// — a mismatched creator simply finds no document to update.
func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(primitive.ObjectID)

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid poll id")
		return
	}

	var req UpdatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Status == nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "status is required")
		return
	}

	p, err := h.svc.UpdateStatus(c.Request.Context(), id, userID, *req.Status)
	if err != nil {
		if errors.Is(err, ErrPollNotFound) {
			apiError(c, http.StatusForbidden, "FORBIDDEN", "You do not own this poll, or it does not exist")
			return
		}
		apiError(c, http.StatusUnprocessableEntity, "INVALID_REQUEST", err.Error())
		return
	}
	c.JSON(http.StatusOK, p)
}

// Delete handles DELETE /api/polls/:id.
func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(primitive.ObjectID)

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid poll id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, ErrPollNotFound) {
			apiError(c, http.StatusForbidden, "FORBIDDEN", "You do not own this poll, or it does not exist")
			return
		}
		apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not delete poll")
		return
	}
	c.Status(http.StatusNoContent)
}
