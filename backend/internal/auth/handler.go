package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CookieName = "livepoll_token"

type Handler struct {
	svc        *Service
	cookieOpts CookieOptions
}

// CookieOptions controls how the auth cookie is written. Secure should
// be true in production (HTTPS/WSS); it's kept false for plain-HTTP
// local development.
type CookieOptions struct {
	Secure bool
	Domain string
}

func NewHandler(svc *Service, opts CookieOptions) *Handler {
	return &Handler{svc: svc, cookieOpts: opts}
}

func (h *Handler) setCookie(c *gin.Context, token string, maxAgeSeconds int) {
	if h.cookieOpts.Secure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie(CookieName, token, maxAgeSeconds, "/", h.cookieOpts.Domain, h.cookieOpts.Secure, true)
}

func apiError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func (h *Handler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "Request body is malformed")
		return
	}

	user, token, err := h.svc.Signup(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			apiError(c, http.StatusConflict, "USER_EXISTS", "An account with this email already exists")
		case errors.Is(err, ErrEmailRequired), errors.Is(err, ErrEmailInvalid),
			errors.Is(err, ErrPasswordRequired), errors.Is(err, ErrPasswordTooShort),
			errors.Is(err, ErrPasswordMismatch):
			apiError(c, http.StatusUnprocessableEntity, "INVALID_REQUEST", err.Error())
		default:
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong")
		}
		return
	}

	h.setCookie(c, token, int((7 * 24 * time.Hour).Seconds()))
	c.JSON(http.StatusCreated, MeResponse{ID: user.ID.Hex(), Email: user.Email, Token: token})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiError(c, http.StatusBadRequest, "INVALID_REQUEST", "Request body is malformed")
		return
	}

	user, token, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		apiError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	h.setCookie(c, token, int((7 * 24 * time.Hour).Seconds()))
	c.JSON(http.StatusOK, MeResponse{ID: user.ID.Hex(), Email: user.Email, Token: token})
}

func (h *Handler) Logout(c *gin.Context) {
	h.setCookie(c, "", -1)
	c.Status(http.StatusNoContent)
}

// Me returns the currently authenticated user. Requires the auth
// middleware to have already run and set "userID" in the context.
func (h *Handler) Me(c *gin.Context) {
	userID := c.MustGet("userID").(primitive.ObjectID)

	user, err := h.svc.GetUser(c.Request.Context(), userID)
	if err != nil {
		apiError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	c.JSON(http.StatusOK, MeResponse{ID: user.ID.Hex(), Email: user.Email})
}
