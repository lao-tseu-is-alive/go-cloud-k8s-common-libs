package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MockDB is a mock implementation of the DB interface for testing
type MockDB struct {
	// Configurable return values
	ExecActionQueryResult int
	ExecActionQueryError  error
	InsertResult          int
	InsertError           error
	GetQueryIntResult     int
	GetQueryIntError      error
	GetQueryBoolResult    bool
	GetQueryBoolError     error
	GetQueryStringResult  string
	GetQueryStringError   error
	GetVersionResult      string
	GetVersionError       error
	HealthCheckResult     bool
	HealthCheckError      error
	DoesTableExistResult  bool
	GetPGConnError        error

	// Call tracking
	ExecActionQueryCalls []MockCall
	InsertCalls          []MockCall
	GetQueryIntCalls     []MockCall
	GetQueryBoolCalls    []MockCall
	GetQueryStringCalls  []MockCall
	GetVersionCalls      int
	HealthCheckCalls     int
	DoesTableExistCalls  []MockTableExistCall
	CloseCalls           int
}

// MockCall tracks a call with SQL and arguments
type MockCall struct {
	SQL       string
	Arguments []interface{}
}

// MockTableExistCall tracks DoesTableExist calls
type MockTableExistCall struct {
	Schema string
	Table  string
}

// NewMockDB creates a new MockDB with sensible defaults
func NewMockDB() *MockDB {
	return &MockDB{
		ExecActionQueryResult: 1,
		InsertResult:          1,
		GetQueryIntResult:     0,
		GetQueryBoolResult:    true,
		GetQueryStringResult:  "",
		GetVersionResult:      "PostgreSQL 14.0",
		HealthCheckResult:     true,
		DoesTableExistResult:  true,
	}
}

func (m *MockDB) ExecActionQuery(ctx context.Context, sql string, arguments ...interface{}) (rowsAffected int, err error) {
	m.ExecActionQueryCalls = append(m.ExecActionQueryCalls, MockCall{SQL: sql, Arguments: arguments})
	return m.ExecActionQueryResult, m.ExecActionQueryError
}

func (m *MockDB) Insert(ctx context.Context, sql string, arguments ...interface{}) (lastInsertId int, err error) {
	m.InsertCalls = append(m.InsertCalls, MockCall{SQL: sql, Arguments: arguments})
	return m.InsertResult, m.InsertError
}

func (m *MockDB) GetQueryInt(ctx context.Context, sql string, arguments ...interface{}) (result int, err error) {
	m.GetQueryIntCalls = append(m.GetQueryIntCalls, MockCall{SQL: sql, Arguments: arguments})
	return m.GetQueryIntResult, m.GetQueryIntError
}

func (m *MockDB) GetQueryBool(ctx context.Context, sql string, arguments ...interface{}) (result bool, err error) {
	m.GetQueryBoolCalls = append(m.GetQueryBoolCalls, MockCall{SQL: sql, Arguments: arguments})
	return m.GetQueryBoolResult, m.GetQueryBoolError
}

func (m *MockDB) GetQueryString(ctx context.Context, sql string, arguments ...interface{}) (result string, err error) {
	m.GetQueryStringCalls = append(m.GetQueryStringCalls, MockCall{SQL: sql, Arguments: arguments})
	return m.GetQueryStringResult, m.GetQueryStringError
}

func (m *MockDB) GetVersion(ctx context.Context) (result string, err error) {
	m.GetVersionCalls++
	return m.GetVersionResult, m.GetVersionError
}

func (m *MockDB) GetPGConn() (Conn *pgxpool.Pool, err error) {
	return nil, m.GetPGConnError
}

func (m *MockDB) HealthCheck(ctx context.Context) (alive bool, err error) {
	m.HealthCheckCalls++
	return m.HealthCheckResult, m.HealthCheckError
}

func (m *MockDB) DoesTableExist(ctx context.Context, schema, table string) (exist bool) {
	m.DoesTableExistCalls = append(m.DoesTableExistCalls, MockTableExistCall{Schema: schema, Table: table})
	return m.DoesTableExistResult
}

func (m *MockDB) Close() {
	m.CloseCalls++
}

// Reset clears all call tracking
func (m *MockDB) Reset() {
	m.ExecActionQueryCalls = nil
	m.InsertCalls = nil
	m.GetQueryIntCalls = nil
	m.GetQueryBoolCalls = nil
	m.GetQueryStringCalls = nil
	m.GetVersionCalls = 0
	m.HealthCheckCalls = 0
	m.DoesTableExistCalls = nil
	m.CloseCalls = 0
}

// SetError sets error for all methods - useful for simulating DB failure
func (m *MockDB) SetError(err error) {
	m.ExecActionQueryError = err
	m.InsertError = err
	m.GetQueryIntError = err
	m.GetQueryBoolError = err
	m.GetQueryStringError = err
	m.GetVersionError = err
	m.HealthCheckError = err
	m.GetPGConnError = err
}

// ErrMockDBConnection is a sample error for testing
var ErrMockDBConnection = errors.New("mock database connection error")
