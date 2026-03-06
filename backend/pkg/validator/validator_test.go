package validator

import (
	"strings"
	"testing"
)

type registerRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type createPostRequest struct {
	Content    string `json:"content"    validate:"required,min=1,max=280"`
	Visibility string `json:"visibility" validate:"omitempty,oneof=public friends private"`
}

type updateProfileRequest struct {
	DisplayName     string `json:"displayName"     validate:"omitempty,max=50"`
	Bio             string `json:"bio"             validate:"omitempty,max=160"`
	Username        string `json:"username"        validate:"omitempty,min=3,max=30,alphanum"`
	ProfileImageURL string `json:"profileImageUrl" validate:"omitempty,url"`
	HeaderImageURL  string `json:"headerImageUrl"  validate:"omitempty,url"`
}

func TestValidate_RegisterRequest(t *testing.T) {
	tests := []struct {
		name        string
		input       registerRequest
		wantErr     bool
		wantField   string
		wantMessage string
	}{
		{
			name: "valid request",
			input: registerRequest{
				Email:    "user@example.com",
				Username: "testuser",
				Password: "securepass123",
			},
			wantErr: false,
		},
		{
			name: "empty email",
			input: registerRequest{
				Email:    "",
				Username: "testuser",
				Password: "securepass123",
			},
			wantErr:     true,
			wantField:   "email",
			wantMessage: "this field is required",
		},
		{
			name: "invalid email format",
			input: registerRequest{
				Email:    "not-an-email",
				Username: "testuser",
				Password: "securepass123",
			},
			wantErr:     true,
			wantField:   "email",
			wantMessage: "must be a valid email address",
		},
		{
			name: "short password",
			input: registerRequest{
				Email:    "user@example.com",
				Username: "testuser",
				Password: "short",
			},
			wantErr:     true,
			wantField:   "password",
			wantMessage: "must be at least 8 characters",
		},
		{
			name: "short username",
			input: registerRequest{
				Email:    "user@example.com",
				Username: "ab",
				Password: "securepass123",
			},
			wantErr:     true,
			wantField:   "username",
			wantMessage: "must be at least 3 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Validate(tt.input)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Fatal("expected validation errors but got none")
				}
				found := false
				for _, e := range errs {
					if e.Field == tt.wantField && e.Message == tt.wantMessage {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error field=%q message=%q, got %+v", tt.wantField, tt.wantMessage, errs)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("expected no errors, got %+v", errs)
				}
			}
		})
	}
}

func TestValidate_CreatePostRequest(t *testing.T) {
	tests := []struct {
		name        string
		input       createPostRequest
		wantErr     bool
		wantField   string
		wantMessage string
	}{
		{
			name: "valid request",
			input: createPostRequest{
				Content:    "Hello, world!",
				Visibility: "public",
			},
			wantErr: false,
		},
		{
			name: "valid request without visibility",
			input: createPostRequest{
				Content: "Hello, world!",
			},
			wantErr: false,
		},
		{
			name: "empty content",
			input: createPostRequest{
				Content: "",
			},
			wantErr:     true,
			wantField:   "content",
			wantMessage: "this field is required",
		},
		{
			name: "content exceeds 280 characters",
			input: createPostRequest{
				Content: strings.Repeat("a", 281),
			},
			wantErr:     true,
			wantField:   "content",
			wantMessage: "must be at most 280 characters",
		},
		{
			name: "invalid visibility value",
			input: createPostRequest{
				Content:    "Hello",
				Visibility: "everyone",
			},
			wantErr:     true,
			wantField:   "visibility",
			wantMessage: "must be one of: public friends private",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Validate(tt.input)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Fatal("expected validation errors but got none")
				}
				found := false
				for _, e := range errs {
					if e.Field == tt.wantField && e.Message == tt.wantMessage {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error field=%q message=%q, got %+v", tt.wantField, tt.wantMessage, errs)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("expected no errors, got %+v", errs)
				}
			}
		})
	}
}

func TestValidate_UpdateProfileRequest(t *testing.T) {
	tests := []struct {
		name        string
		input       updateProfileRequest
		wantErr     bool
		wantField   string
		wantMessage string
	}{
		{
			name: "valid request with all fields",
			input: updateProfileRequest{
				DisplayName:     "Test User",
				Bio:             "Hello, I am a test user.",
				Username:        "testuser",
				ProfileImageURL: "https://example.com/avatar.png",
				HeaderImageURL:  "https://example.com/header.png",
			},
			wantErr: false,
		},
		{
			name:    "valid empty request (all optional)",
			input:   updateProfileRequest{},
			wantErr: false,
		},
		{
			name: "bio exceeds 160 characters",
			input: updateProfileRequest{
				Bio: strings.Repeat("a", 161),
			},
			wantErr:     true,
			wantField:   "bio",
			wantMessage: "must be at most 160 characters",
		},
		{
			name: "invalid profile image URL",
			input: updateProfileRequest{
				ProfileImageURL: "not-a-url",
			},
			wantErr:     true,
			wantField:   "profile_image_u_r_l",
			wantMessage: "must be a valid URL",
		},
		{
			name: "invalid header image URL",
			input: updateProfileRequest{
				HeaderImageURL: "not-a-url",
			},
			wantErr:     true,
			wantField:   "header_image_u_r_l",
			wantMessage: "must be a valid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Validate(tt.input)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Fatal("expected validation errors but got none")
				}
				found := false
				for _, e := range errs {
					if e.Field == tt.wantField && e.Message == tt.wantMessage {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error field=%q message=%q, got %+v", tt.wantField, tt.wantMessage, errs)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("expected no errors, got %+v", errs)
				}
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Email", want: "email"},
		{input: "Username", want: "username"},
		{input: "DisplayName", want: "display_name"},
		{input: "ProfileImageURL", want: "profile_image_u_r_l"},
		{input: "already_snake", want: "already_snake"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
