package auth

import "testing"

func TestValidateSignup(t *testing.T) {
	cases := []struct {
		name    string
		req     SignupRequest
		wantErr error
	}{
		{"valid", SignupRequest{Email: "User@Example.com", Password: "password1", ConfirmPassword: "password1"}, nil},
		{"missing email", SignupRequest{Password: "password1", ConfirmPassword: "password1"}, ErrEmailRequired},
		{"bad email", SignupRequest{Email: "not-an-email", Password: "password1", ConfirmPassword: "password1"}, ErrEmailInvalid},
		{"short password", SignupRequest{Email: "a@b.com", Password: "short", ConfirmPassword: "short"}, ErrPasswordTooShort},
		{"mismatched", SignupRequest{Email: "a@b.com", Password: "password1", ConfirmPassword: "password2"}, ErrPasswordMismatch},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateSignup(tc.req)
			if err != tc.wantErr {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestNormalizeEmailLowercasesAndTrims(t *testing.T) {
	req, err := validateSignup(SignupRequest{
		Email: "  Foo@BAR.com  ", Password: "password1", ConfirmPassword: "password1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Email != "foo@bar.com" {
		t.Fatalf("got %q, want foo@bar.com", req.Email)
	}
}
