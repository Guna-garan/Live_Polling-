package poll

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrQuestionRequired  = errors.New("question is required")
	ErrQuestionTooShort  = errors.New("question must be at least 5 characters")
	ErrQuestionTooLong   = errors.New("question must be at most 200 characters")
	ErrTooFewOptions     = errors.New("a poll needs at least 2 options")
	ErrTooManyOptions    = errors.New("a poll may have at most 10 options")
	ErrOptionEmpty       = errors.New("option text cannot be empty")
	ErrOptionTooLong     = errors.New("option text must be at most 100 characters")
	ErrDuplicateOption   = errors.New("duplicate options are not allowed")
	ErrInvalidExpiration = errors.New("expiresAt must be a valid RFC3339 timestamp in the future")
)

// normalized is the cleaned-up, validated form of a create-poll request:
// trimmed strings, a parsed expiration, ready to persist.
type normalized struct {
	Question    string
	Description string
	Options     []string
	ExpiresAt   *time.Time
	MaxVotes    *int64
}

func validateCreate(req CreatePollRequest) (normalized, error) {
	out := normalized{
		Question:    strings.TrimSpace(req.Question),
		Description: strings.TrimSpace(req.Description),
		MaxVotes:    req.MaxVotes,
	}

	if out.Question == "" {
		return out, ErrQuestionRequired
	}
	if len(out.Question) < 5 {
		return out, ErrQuestionTooShort
	}
	if len(out.Question) > 200 {
		return out, ErrQuestionTooLong
	}

	if len(req.Options) < 2 {
		return out, ErrTooFewOptions
	}
	if len(req.Options) > 10 {
		return out, ErrTooManyOptions
	}

	seen := make(map[string]bool, len(req.Options))
	for _, raw := range req.Options {
		text := strings.TrimSpace(raw)
		if text == "" {
			return out, ErrOptionEmpty
		}
		if len(text) > 100 {
			return out, ErrOptionTooLong
		}
		key := strings.ToLower(text)
		if seen[key] {
			return out, ErrDuplicateOption
		}
		seen[key] = true
		out.Options = append(out.Options, text)
	}

	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil || !t.After(time.Now()) {
			return out, ErrInvalidExpiration
		}
		out.ExpiresAt = &t
	}

	return out, nil
}
