package config

import (
	"os"
	"strings"
	"testing"
)

func TestGetJwtDurationFromEnv(t *testing.T) {
	type args struct {
		defaultJwtDuration int
	}
	tests := []struct {
		name          string
		args          args
		envVariable   string
		want          int
		wantErr       bool
		wantErrPrefix string
	}{
		{
			name: "should return the default values when env variables are not set",
			args: args{
				defaultJwtDuration: 600,
			},
			envVariable:   "",
			want:          600,
			wantErr:       false,
			wantErrPrefix: "",
		},
		{
			name: "should return JWT_DURATION_MINUTES when env variables is set to valid values",
			args: args{
				defaultJwtDuration: 600,
			},
			envVariable:   "60",
			want:          60,
			wantErr:       false,
			wantErrPrefix: "",
		},
		{
			name: "should return an empty string and report an error when PORT is not a number",
			args: args{
				defaultJwtDuration: 600,
			},
			envVariable:   "aBigOne",
			want:          0,
			wantErr:       true,
			wantErrPrefix: "ERROR: CONFIG ENV JWT_DURATION_MINUTES should contain a valid integer.",
		},
		{
			name: "should return an empty string and report an error when PORT is < 1",
			args: args{
				defaultJwtDuration: 600,
			},
			envVariable:   "0",
			want:          0,
			wantErr:       true,
			wantErrPrefix: "ERROR: CONFIG ENV JWT_DURATION_MINUTES should contain an integer between 1 and 14400",
		},
		{
			name: "should return an empty string and report an error when PORT is > 65535",
			args: args{
				defaultJwtDuration: 600,
			},
			envVariable:   "70000",
			want:          0,
			wantErr:       true,
			wantErrPrefix: "ERROR: CONFIG ENV JWT_DURATION_MINUTES should contain an integer between 1 and 14400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.envVariable) > 0 {
				err := os.Setenv("JWT_DURATION_MINUTES", tt.envVariable)
				if err != nil {
					t.Errorf("Unable to set env variable PORT")
					return
				}
			} else {
				// we do not want that an external setting of PORT breaks this test
				err := os.Unsetenv("JWT_DURATION_MINUTES")
				if err != nil {
					t.Errorf("Unable to unset env variable PORT")
					return
				}
			}
			got, err := GetJwtDurationFromEnv(tt.args.defaultJwtDuration)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPortFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				// check that error contains the ERROR keyword
				if strings.HasPrefix(err.Error(), "ERROR:") == false {
					t.Errorf("GetPortFromEnv() error = %v, wantErrPrefix %v", err, tt.wantErrPrefix)
				}
			}
			if got != tt.want {
				t.Errorf("GetPortFromEnv() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetJwtSecretFromEnv(t *testing.T) {
	tests := []struct {
		name          string
		envVariable   string
		want          string
		wantErr       bool
		wantErrPrefix string
	}{
		{
			name:          "should return the correct env variable values",
			envVariable:   "this is not a secret",
			want:          "this is not a secret",
			wantErr:       false,
			wantErrPrefix: "",
		},
		{
			name:          "should return the correct env variable values",
			envVariable:   "",
			want:          "",
			wantErr:       true,
			wantErrPrefix: "ERROR: CONFIG ENV JWT_SECRET should contain your JWT secret.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.envVariable) > 0 {
				err := os.Setenv("JWT_SECRET", tt.envVariable)
				if err != nil {
					t.Errorf("Unable to set env variable JWT_SECRET")
					return
				}
			} else {
				// we do not want that an external setting of ADMIN_USER breaks this test
				err := os.Unsetenv("JWT_SECRET")
				if err != nil {
					t.Errorf("Unable to unset env variable JWT_SECRET")
					return
				}
			}
			got, err := GetJwtSecretFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetJwtSecretFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetJwtSecretFromEnv() got = %v, want %v", got, tt.want)
			}
		})
	}
}
