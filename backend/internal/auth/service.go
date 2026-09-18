package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

const tokenTTL = 7 * 24 * time.Hour

type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

type Service struct {
	repo      *Repository
	jwtSecret []byte
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: []byte(jwtSecret)}
}

func (s *Service) Signup(ctx context.Context, req SignupRequest) (*User, string, error) {
	req, err := validateSignup(req)
	if err != nil {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &User{Email: req.Email, PasswordHash: string(hash)}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.issueToken(user.ID)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*User, string, error) {
	req, err := validateLogin(req)
	if err != nil {
		// Same generic error for validation failures as for bad
		// credentials — we never want to hint at which field was wrong.
		return nil, "", ErrInvalidCredentials
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Run a dummy bcrypt comparison so that the response time for
			// "no such user" and "wrong password" is indistinguishable,
			// which avoids leaking account existence via timing.
			_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$........................................"), []byte(req.Password))
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.issueToken(user.ID)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *Service) issueToken(userID primitive.ObjectID) (string, error) {
	claims := Claims{
		UserID: userID.Hex(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ParseToken validates a JWT and returns the embedded user ID. Used by
// the auth middleware on every protected request.
func (s *Service) ParseToken(tokenStr string) (primitive.ObjectID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return primitive.NilObjectID, errors.New("invalid or expired token")
	}
	return primitive.ObjectIDFromHex(claims.UserID)
}

func (s *Service) GetUser(ctx context.Context, id primitive.ObjectID) (*User, error) {
	return s.repo.FindByID(ctx, id)
}
