package poll

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusActive = "active"
	StatusClosed = "closed"
)

type Option struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

type Poll struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatorID   primitive.ObjectID `bson:"creatorId" json:"creatorId"`
	Question    string             `bson:"question" json:"question"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Options     []Option           `bson:"options" json:"options"`
	Status      string             `bson:"status" json:"status"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
	ExpiresAt   *time.Time         `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
	MaxVotes    *int64             `bson:"maxVotes,omitempty" json:"maxVotes,omitempty"`
}

// IsOpen reports whether the poll currently accepts votes, taking both
// its manual status and its expiration time into account. This is the
// single source of truth the backend uses everywhere a "can this poll
// still receive votes?" decision is made.
func (p *Poll) IsOpen(now time.Time) bool {
	if p.Status != StatusActive {
		return false
	}
	if p.ExpiresAt != nil && now.After(*p.ExpiresAt) {
		return false
	}
	return true
}

func (p *Poll) HasOption(optionID string) bool {
	for _, o := range p.Options {
		if o.ID == optionID {
			return true
		}
	}
	return false
}

type CreatePollRequest struct {
	Question    string   `json:"question"`
	Description string   `json:"description"`
	Options     []string `json:"options"`
	ExpiresAt   *string  `json:"expiresAt"` // RFC3339, or omitted for no expiration
	MaxVotes    *int64   `json:"maxVotes"`
}

type UpdatePollRequest struct {
	Status *string `json:"status"`
}

// PollWithResults is what GET /api/polls/:id and the list endpoints
// return: the poll plus its live vote counts, so the frontend never has
// to make two round trips to render a poll.
type PollWithResults struct {
	Poll
	Results    map[string]int64 `json:"results"`
	TotalVotes int64            `json:"totalVotes"`
}
