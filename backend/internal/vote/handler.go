package vote

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"livepoll/internal/poll"
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

// Cast handles POST /api/polls/:id/vote. Public — no account required —
// but every validation step (poll exists/open/not expired, option
// valid, not a duplicate) is enforced here on the server; the frontend's
// own checks are UX only.
func (h *Handler) Cast(c *gin.Context) {
	pollID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid poll id")
		return
	}

	var req CastVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "Request body is malformed")
		return
	}

	res, err := h.svc.Cast(c.Request.Context(), pollID, req)
	if err != nil {
		switch {
		case errors.Is(err, poll.ErrPollNotFound):
			apiError(c, http.StatusNotFound, "POLL_NOT_FOUND", "Poll not found")
		case errors.Is(err, ErrPollClosed):
			apiError(c, http.StatusConflict, "POLL_CLOSED", "This poll is closed")
		case errors.Is(err, ErrPollExpired):
			apiError(c, http.StatusConflict, "POLL_EXPIRED", "This poll has ended")
		case errors.Is(err, ErrInvalidOption):
			apiError(c, http.StatusUnprocessableEntity, "INVALID_OPTION", "That option does not exist on this poll")
		case errors.Is(err, ErrInvalidVoterID):
			apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "voterId is required")
		case errors.Is(err, ErrDuplicateVote):
			results, total, _ := h.svc.GetResults(c.Request.Context(), pollID)
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "ALREADY_VOTED",
					"message": "You've already voted in this poll",
				},
				"results":    results,
				"totalVotes": total,
			})
		default:
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not record vote")
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

// Status handles GET /api/polls/:id/vote-status?voterId=... — lets the
// frontend know on page load whether this browser has already voted,
// without needing to attempt (and fail) a vote first.
func (h *Handler) Status(c *gin.Context) {
	pollID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid poll id")
		return
	}
	voterID := c.Query("voterId")
	if voterID == "" {
		c.JSON(http.StatusOK, gin.H{"hasVoted": false})
		return
	}

	voted, err := h.svc.HasVoted(c.Request.Context(), pollID, voterID)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not check vote status")
		return
	}
	c.JSON(http.StatusOK, gin.H{"hasVoted": voted})
}
