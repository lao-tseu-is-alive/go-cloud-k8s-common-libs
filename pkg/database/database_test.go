package database

import (
	"errors"
	"testing"
)

func TestGetErrorF(t *testing.T) {
	tests := []struct {
		name        string
		errMsg      string
		err         error
		wantContain string
	}{
		{
			name:        "with error message",
			errMsg:      "database operation failed",
			err:         errors.New("connection refused"),
			wantContain: "database operation failed",
		},
		{
			name:        "error is wrapped",
			errMsg:      "query failed",
			err:         ErrNoRecordFound,
			wantContain: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorF(tt.errMsg, tt.err)
			if result == nil {
				t.Fatal("GetErrorF returned nil")
			}
			if !containsString(result.Error(), tt.wantContain) {
				t.Errorf("GetErrorF() = %q, should contain %q", result.Error(), tt.wantContain)
			}
			// Test that errors.Is works for wrapped errors
			if tt.err == ErrNoRecordFound && !errors.Is(result, ErrNoRecordFound) {
				t.Error("errors.Is should work with wrapped error")
			}
		})
	}
}

func TestGetInstance_InvalidDriver(t *testing.T) {
	// Test with invalid driver - should return error without DB connection attempt
	_, err := GetInstance(nil, "invalid-driver", "connection-string", 10, nil)
	if err == nil {
		t.Error("GetInstance should return error for invalid driver")
	}
	if !containsString(err.Error(), "unsupported") {
		t.Errorf("error should mention 'unsupported', got: %q", err.Error())
	}
}

func TestMockDB_Implementation(t *testing.T) {
	// Verify MockDB implements DB interface
	var _ DB = (*MockDB)(nil)

	mock := NewMockDB()

	t.Run("default values", func(t *testing.T) {
		if !mock.HealthCheckResult {
			t.Error("HealthCheckResult should default to true")
		}
		if mock.GetVersionResult == "" {
			t.Error("GetVersionResult should have default value")
		}
	})

	t.Run("ExecActionQuery tracking", func(t *testing.T) {
		mock.Reset()
		_, _ = mock.ExecActionQuery(nil, "UPDATE test SET x = 1", "arg1", "arg2")
		if len(mock.ExecActionQueryCalls) != 1 {
			t.Errorf("expected 1 call, got %d", len(mock.ExecActionQueryCalls))
		}
		if mock.ExecActionQueryCalls[0].SQL != "UPDATE test SET x = 1" {
			t.Errorf("SQL not tracked correctly")
		}
	})

	t.Run("Insert tracking", func(t *testing.T) {
		mock.Reset()
		result, _ := mock.Insert(nil, "INSERT INTO test VALUES ($1)", "value")
		if result != mock.InsertResult {
			t.Errorf("Insert() = %d, want %d", result, mock.InsertResult)
		}
		if len(mock.InsertCalls) != 1 {
			t.Error("Insert call not tracked")
		}
	})

	t.Run("GetQueryInt", func(t *testing.T) {
		mock.Reset()
		mock.GetQueryIntResult = 42
		result, _ := mock.GetQueryInt(nil, "SELECT count(*)")
		if result != 42 {
			t.Errorf("GetQueryInt() = %d, want 42", result)
		}
	})

	t.Run("GetQueryBool", func(t *testing.T) {
		mock.Reset()
		mock.GetQueryBoolResult = false
		result, _ := mock.GetQueryBool(nil, "SELECT exists")
		if result != false {
			t.Error("GetQueryBool should return false")
		}
	})

	t.Run("GetQueryString", func(t *testing.T) {
		mock.Reset()
		mock.GetQueryStringResult = "test-value"
		result, _ := mock.GetQueryString(nil, "SELECT name")
		if result != "test-value" {
			t.Errorf("GetQueryString() = %q, want 'test-value'", result)
		}
	})

	t.Run("GetVersion", func(t *testing.T) {
		mock.Reset()
		result, _ := mock.GetVersion(nil)
		if result != mock.GetVersionResult {
			t.Errorf("GetVersion() = %q, want %q", result, mock.GetVersionResult)
		}
		if mock.GetVersionCalls != 1 {
			t.Error("GetVersion call not tracked")
		}
	})

	t.Run("HealthCheck", func(t *testing.T) {
		mock.Reset()
		result, _ := mock.HealthCheck(nil)
		if result != true {
			t.Error("HealthCheck should return true by default")
		}
	})

	t.Run("DoesTableExist", func(t *testing.T) {
		mock.Reset()
		result := mock.DoesTableExist(nil, "public", "users")
		if result != mock.DoesTableExistResult {
			t.Error("DoesTableExist should return configured result")
		}
		if len(mock.DoesTableExistCalls) != 1 {
			t.Error("DoesTableExist call not tracked")
		}
	})

	t.Run("Close", func(t *testing.T) {
		mock.Reset()
		mock.Close()
		if mock.CloseCalls != 1 {
			t.Error("Close call not tracked")
		}
	})

	t.Run("SetError", func(t *testing.T) {
		mock.Reset()
		mock.SetError(ErrMockDBConnection)
		_, err := mock.GetVersion(nil)
		if err != ErrMockDBConnection {
			t.Error("SetError should set error for all methods")
		}
	})
}

func TestErrorVariables(t *testing.T) {
	if ErrNoRecordFound == nil {
		t.Error("ErrNoRecordFound should not be nil")
	}
	if ErrCouldNotBeCreated == nil {
		t.Error("ErrCouldNotBeCreated should not be nil")
	}
	if ErrDBNotAvailable == nil {
		t.Error("ErrDBNotAvailable should not be nil")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
