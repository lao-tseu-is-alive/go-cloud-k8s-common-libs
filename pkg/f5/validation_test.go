package f5

import (
	"testing"
)

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		wantErr bool
	}{
		{
			name:    "valid login simple",
			login:   "testuser",
			wantErr: false,
		},
		{
			name:    "valid login with underscore",
			login:   "test_user",
			wantErr: false,
		},
		{
			name:    "valid login with dot",
			login:   "test.user",
			wantErr: false,
		},
		{
			name:    "valid login with dash",
			login:   "test-user",
			wantErr: false,
		},
		{
			name:    "valid login with numbers",
			login:   "user123",
			wantErr: false,
		},
		{
			name:    "valid login minimum length",
			login:   "usr",
			wantErr: false,
		},
		{
			name:    "too short login",
			login:   "ab",
			wantErr: true,
		},
		{
			name:    "empty login",
			login:   "",
			wantErr: true,
		},
		{
			name:    "too long login",
			login:   "thisloginiswaytoolongandexceedsthefiftycharacterlimitsetbyvalidation",
			wantErr: true,
		},
		{
			name:    "invalid characters - space",
			login:   "test user",
			wantErr: true,
		},
		{
			name:    "invalid characters - special",
			login:   "user@domain",
			wantErr: true,
		},
		{
			name:    "invalid characters - unicode",
			login:   "usér",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogin(tt.login)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLogin(%q) error = %v, wantErr %v", tt.login, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "valid SHA-256 hash lowercase",
			hash:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr: false,
		},
		{
			name:    "valid SHA-256 hash uppercase",
			hash:    "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
			wantErr: false,
		},
		{
			name:    "valid SHA-256 hash mixed case",
			hash:    "E3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr: false,
		},
		{
			name:    "invalid - too short",
			hash:    "e3b0c44298fc1c149afbf4c8996fb924",
			wantErr: true,
		},
		{
			name:    "invalid - too long",
			hash:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b8551234",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			hash:    "",
			wantErr: true,
		},
		{
			name:    "invalid - non-hex characters",
			hash:    "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr: true,
		},
		{
			name:    "invalid - contains spaces",
			hash:    "e3b0c44298fc1c14 afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordHash(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordHash(%q) error = %v, wantErr %v", tt.hash, err, tt.wantErr)
			}
		})
	}
}
