package poll

import (
	"testing"
	"time"
)

func TestValidateCreate(t *testing.T) {
	cases := []struct {
		name    string
		req     CreatePollRequest
		wantErr error
	}{
		{
			"valid",
			CreatePollRequest{Question: "Which language?", Options: []string{"Go", "JS"}},
			nil,
		},
		{
			"question too short",
			CreatePollRequest{Question: "Hi?", Options: []string{"Go", "JS"}},
			ErrQuestionTooShort,
		},
		{
			"too few options",
			CreatePollRequest{Question: "Which language?", Options: []string{"Go"}},
			ErrTooFewOptions,
		},
		{
			"too many options",
			CreatePollRequest{Question: "Which language?", Options: make([]string, 11)},
			ErrTooManyOptions,
		},
		{
			"duplicate options",
			CreatePollRequest{Question: "Which language?", Options: []string{"Go", "go"}},
			ErrDuplicateOption,
		},
		{
			"empty option",
			CreatePollRequest{Question: "Which language?", Options: []string{"Go", "   "}},
			ErrOptionEmpty,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateCreate(tc.req)
			if err != tc.wantErr {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestPollIsOpen(t *testing.T) {
	now := time.Now()
	p := Poll{Status: StatusActive}
	if !p.IsOpen(now) {
		t.Fatal("expected active, non-expiring poll to be open")
	}

	p.Status = StatusClosed
	if p.IsOpen(now) {
		t.Fatal("expected closed poll to be not open")
	}

	expired := now.Add(-time.Minute)
	p = Poll{Status: StatusActive, ExpiresAt: &expired}
	if p.IsOpen(now) {
		t.Fatal("expected expired poll to be not open")
	}
}
