package config

import (
	"os"
	"testing"
)

func TestGetAdminUserFromFromEnv(t *testing.T) {
	type args struct {
		defaultAdminUser string
	}

	tests := []struct {
		name          string
		args          args
		envAdminUser  string
		want          string
		wantErr       bool
		wantErrPrefix string
	}{
		{
			name: "should return the default values when env variable is not set",
			args: args{
				defaultAdminUser: "goAdmin",
			},
			envAdminUser:  "",
			want:          "goAdmin",
			wantErr:       false,
			wantErrPrefix: "",
		},
		{
			name: "should return the env variable is set to valid values",
			args: args{
				defaultAdminUser: "goAdmin",
			},
			envAdminUser:  "admin",
			want:          "admin",
			wantErr:       false,
			wantErrPrefix: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.envAdminUser) > 0 {
				err := os.Setenv("ADMIN_USER", tt.envAdminUser)
				if err != nil {
					t.Errorf("Unable to set env variable ADMIN_USER")
					return
				}
			} else {
				// we do not want that an external setting of ADMIN_USER breaks this test
				err := os.Unsetenv("ADMIN_USER")
				if err != nil {
					t.Errorf("Unable to unset env variable ADMIN_USER")
					return
				}
			}
			if got := GetAdminUserFromFromEnv(tt.args.defaultAdminUser); got != tt.want {
				t.Errorf("GetAdminUserFromFromEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAdminPasswordFromFromEnv(t *testing.T) {

	tests := []struct {
		name             string
		envAdminPassword string
		want             string
		wantErr          bool
		wantErrPrefix    string
	}{
		{
			name:             "should return Error when env variable is not set",
			envAdminPassword: "",
			want:             "",
			wantErr:          true,
			wantErrPrefix:    "",
		},
		{
			name:             "should return the env variable is set to valid values",
			envAdminPassword: "theOneYouWillMayBeChoose",
			want:             "theOneYouWillMayBeChoose",
			wantErr:          false,
			wantErrPrefix:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.envAdminPassword) > 0 {
				err := os.Setenv("ADMIN_PASSWORD", tt.envAdminPassword)
				if err != nil {
					t.Errorf("Unable to set env variable ADMIN_PASSWORD")
					return
				}
			} else {
				// we do not want that an external setting of ADMIN_USER breaks this test
				err := os.Unsetenv("ADMIN_PASSWORD")
				if err != nil {
					t.Errorf("Unable to unset env variable ADMIN_PASSWORD")
					return
				}
			}
			got, err := GetAdminPasswordFromFromEnv()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("GetAdminUserFromFromEnv() returned un unwanted error: %s got= %v, want %v", err, got, tt.want)
				}
			} else {
				if tt.wantErr {
					t.Errorf("GetAdminUserFromFromEnv() di not return the expected error: %s got= %v, want %v", err, got, tt.want)
				}
			}
			if got != tt.want {
				t.Errorf("GetAdminUserFromFromEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
