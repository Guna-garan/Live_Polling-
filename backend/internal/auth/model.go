package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the persisted representation of an account. PasswordHash is
// never marshaled to JSON so it can never leak through an API response,
// even by accident.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type SignupRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MeResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Token string `json:"token,omitempty"`
}
