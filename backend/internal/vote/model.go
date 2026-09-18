package vote

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID    primitive.ObjectID `bson:"pollId" json:"pollId"`
	OptionID  string             `bson:"optionId" json:"optionId"`
	VoterID   string             `bson:"voterId" json:"voterId"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

type CastVoteRequest struct {
	OptionID string `json:"optionId"`
	VoterID  string `json:"voterId"`
}

type CastVoteResponse struct {
	Results    map[string]int64 `json:"results"`
	TotalVotes int64            `json:"totalVotes"`
}
