package metadata

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

func TestService_CreateMetadataTableOrFail_TableExists(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockDB := database.NewMockDB()

	// Table exists with services registered
	mockDB.DoesTableExistResult = true
	mockDB.GetQueryIntResult = 5 // 5 services registered

	service := &Service{
		Log: logger,
		Db:  mockDB,
	}

	// Should not panic
	service.CreateMetadataTableOrFail(context.Background())

	// Verify DoesTableExist was called
	if len(mockDB.DoesTableExistCalls) != 1 {
		t.Errorf("expected 1 DoesTableExist call, got %d", len(mockDB.DoesTableExistCalls))
	}

	// Verify it checked for metadata table
	if mockDB.DoesTableExistCalls[0].Table != MetaTableName {
		t.Errorf("expected table %q, got %q", MetaTableName, mockDB.DoesTableExistCalls[0].Table)
	}
}

func TestService_CreateMetadataTableOrFail_TableNotExists(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockDB := database.NewMockDB()

	// Table does not exist
	mockDB.DoesTableExistResult = false
	mockDB.ExecActionQueryResult = 0 // CREATE TABLE returns 0 rows

	service := &Service{
		Log: logger,
		Db:  mockDB,
	}

	// Should not panic when table is created successfully
	service.CreateMetadataTableOrFail(context.Background())

	// Verify ExecActionQuery was called to create table
	if len(mockDB.ExecActionQueryCalls) != 1 {
		t.Errorf("expected 1 ExecActionQuery call, got %d", len(mockDB.ExecActionQueryCalls))
	}
}

func TestService_GetServiceVersionOrFail_ServiceExists(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockDB := database.NewMockDB()

	// Service exists with version
	mockDB.GetQueryIntResult = 1 // count > 0
	mockDB.GetQueryStringResult = "1.0.0"

	service := &Service{
		Log: logger,
		Db:  mockDB,
	}

	found, version := service.GetServiceVersionOrFail(context.Background(), "test-service")

	if !found {
		t.Error("expected found to be true")
	}
	if version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", version)
	}

	// Verify GetQueryInt was called for count
	if len(mockDB.GetQueryIntCalls) != 1 {
		t.Errorf("expected 1 GetQueryInt call, got %d", len(mockDB.GetQueryIntCalls))
	}

	// Verify GetQueryString was called for version
	if len(mockDB.GetQueryStringCalls) != 1 {
		t.Errorf("expected 1 GetQueryString call, got %d", len(mockDB.GetQueryStringCalls))
	}
}

func TestService_GetServiceVersionOrFail_ServiceNotExists(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockDB := database.NewMockDB()

	// Service does not exist
	mockDB.GetQueryIntResult = 0 // count == 0

	service := &Service{
		Log: logger,
		Db:  mockDB,
	}

	found, version := service.GetServiceVersionOrFail(context.Background(), "nonexistent-service")

	if found {
		t.Error("expected found to be false")
	}
	if version != "" {
		t.Errorf("expected empty version, got %q", version)
	}
}

func TestService_SetServiceVersionOrFail_InsertNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockDB := database.NewMockDB()

	// Service does not exist (count = 0)
	mockDB.GetQueryIntResult = 0
	mockDB.ExecActionQueryResult = 1

	service := &Service{
		Log: logger,
		Db:  mockDB,
	}

	// Should not panic
	service.SetServiceVersionOrFail(context.Background(), "new-service", "1.0.0")

	// Verify INSERT was called
	if len(mockDB.ExecActionQueryCalls) != 1 {
		t.Errorf("expected 1 ExecActionQuery call, got %d", len(mockDB.ExecActionQueryCalls))
	}
}

func TestService_SetServiceVersionOrFail_UpdateExisting(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockDB := database.NewMockDB()

	// Service exists with different version
	mockDB.GetQueryIntResult = 1
	mockDB.GetQueryStringResult = "0.9.0" // old version
	mockDB.ExecActionQueryResult = 1

	service := &Service{
		Log: logger,
		Db:  mockDB,
	}

	// Should not panic
	service.SetServiceVersionOrFail(context.Background(), "existing-service", "1.0.0")

	// Verify UPDATE was called
	if len(mockDB.ExecActionQueryCalls) != 1 {
		t.Errorf("expected 1 ExecActionQuery call for UPDATE, got %d", len(mockDB.ExecActionQueryCalls))
	}
}

func TestService_SetServiceVersionOrFail_NoChangeNeeded(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockDB := database.NewMockDB()

	// Service exists with same version
	mockDB.GetQueryIntResult = 1
	mockDB.GetQueryStringResult = "1.0.0" // same version

	service := &Service{
		Log: logger,
		Db:  mockDB,
	}

	// Should not panic and should not call ExecActionQuery
	service.SetServiceVersionOrFail(context.Background(), "existing-service", "1.0.0")

	// Verify no UPDATE/INSERT was called
	if len(mockDB.ExecActionQueryCalls) != 0 {
		t.Errorf("expected 0 ExecActionQuery calls, got %d", len(mockDB.ExecActionQueryCalls))
	}
}
