package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoRecordFound     = errors.New("record not found")
	ErrCouldNotBeCreated = errors.New("could not be created in DB")
	ErrDBNotAvailable    = errors.New("DB connection is not available")
)

// DB is the interface for a simple table store.
type DB interface {
	ExecActionQuery(ctx context.Context, sql string, arguments ...interface{}) (rowsAffected int, err error)
	Insert(ctx context.Context, sql string, arguments ...interface{}) (lastInsertId int, err error)
	GetQueryInt(ctx context.Context, sql string, arguments ...interface{}) (result int, err error)
	GetQueryBool(ctx context.Context, sql string, arguments ...interface{}) (result bool, err error)
	GetQueryString(ctx context.Context, sql string, arguments ...interface{}) (result string, err error)
	GetVersion(ctx context.Context) (result string, err error)
	GetPGConn() (Conn *pgxpool.Pool, err error)
	HealthCheck(ctx context.Context) (alive bool, err error)
	DoesTableExist(ctx context.Context, schema, table string) (exist bool)
	Close()
}

func GetErrorF(errMsg string, err error) error {
	return fmt.Errorf("%s [%w]", errMsg, err) // %w allows errors.Is() to work!
}

// GetInstance with appropriate driver
func GetInstance(ctx context.Context, dbDriver, dbConnectionString string, maxConnectionCount int, log *slog.Logger) (DB, error) {
	var err error
	var db DB

	if dbDriver == "pgx" {
		db, err = newPgxConn(ctx, dbConnectionString, maxConnectionCount, log)
		if err != nil {
			return nil, fmt.Errorf("error opening postgresql database with pgx driver: %s", err)
		}
	} else {
		return nil, errors.New("unsupported DB driver type")
	}

	return db, nil
}
