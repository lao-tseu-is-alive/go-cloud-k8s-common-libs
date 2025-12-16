package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const getPGVersion = "SELECT version();"
const getTableExists = "SELECT EXISTS(SELECT FROM information_schema.tables WHERE  table_schema = $1 AND table_name = $2) as exists;"

type PgxDB struct {
	Conn *pgxpool.Pool
	log  *slog.Logger
}

func newPgxConn(ctx context.Context, dbConnectionString string, maxConnectionsInPool int, log *slog.Logger) (DB, error) {
	var psql PgxDB
	var parsedConfig *pgx.ConnConfig
	var err error
	parsedConfig, err = pgx.ParseConfig(dbConnectionString)
	if err != nil {
		return nil, fmt.Errorf("error doing pgx.ParseConfig(%s). err: %s", dbConnectionString, err)
	}

	dbHost := parsedConfig.Host
	dbPort := parsedConfig.Port
	dbUser := parsedConfig.User
	dbPass := parsedConfig.Password
	dbName := parsedConfig.Database

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s pool_max_conns=%d", dbHost, dbPort, dbUser, dbPass, dbName, maxConnectionsInPool)

	connPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Error("FAILED to connect to database", "dbName", dbName, "dbUser", dbUser)
		return nil, fmt.Errorf("error connecting to database. err : %s", err)
	} else {
		log.Info("SUCCESS connecting to database", "dbName", dbName, "dbUser", dbUser)
		// let's first check that we can really make a query by querying the postgres version
		var version string
		errPing := connPool.QueryRow(ctx, getPGVersion).Scan(&version)
		if errPing != nil {
			log.Error("got db error retrieving postgres version", "error", errPing)
			// Return the error, don't kill the process
			return nil, fmt.Errorf("failed to verify db connection: %w", errPing)
		}

		log.Info("Postgres version", "version", version)
	}

	psql.Conn = connPool
	psql.log = log
	return &psql, err
}

// ExecActionQuery is a postgres helper function for an action query, returning the numbers of rows affected
func (db *PgxDB) ExecActionQuery(ctx context.Context, sql string, arguments ...interface{}) (rowsAffected int, err error) {
	commandTag, err := db.Conn.Exec(ctx, sql, arguments...)
	if err != nil {
		db.log.Error("ExecActionQuery unexpectedly failed", "sql", sql, "args", arguments, "error", err)
		return 0, err
	}
	return int(commandTag.RowsAffected()), err
}

func (db *PgxDB) Insert(ctx context.Context, sql string, arguments ...interface{}) (lastInsertId int, err error) {
	err = db.Conn.QueryRow(ctx, sql, arguments...).Scan(&lastInsertId) //let dev add "RETURNING id" if they need it
	if err != nil {
		db.log.Error("Insert unexpectedly failed", "sql", sql, "args", arguments, "error", err)
		return 0, err
	}
	return lastInsertId, err
}

// GetQueryInt is a postgres helper function for a query expecting an integer result
func (db *PgxDB) GetQueryInt(ctx context.Context, sql string, arguments ...interface{}) (result int, err error) {
	err = db.Conn.QueryRow(ctx, sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Error("GetQueryInt queryRow unexpectedly failed", "sql", sql, "args", arguments, "error", err)
		return 0, err
	}
	return result, err
}

// GetQueryBool is a postgres helper function for a query expecting an integer result
func (db *PgxDB) GetQueryBool(ctx context.Context, sql string, arguments ...interface{}) (result bool, err error) {
	err = db.Conn.QueryRow(ctx, sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Error("GetQueryBool queryRow unexpectedly failed", "sql", sql, "args", arguments, "error", err)
		return false, err
	}
	return result, err
}

func (db *PgxDB) GetQueryString(ctx context.Context, sql string, arguments ...interface{}) (result string, err error) {
	var mayBeResultIsNull *string
	err = db.Conn.QueryRow(ctx, sql, arguments...).Scan(&mayBeResultIsNull)
	if err != nil {
		db.log.Error("GetQueryString queryRow unexpectedly failed", "sql", sql, "args", arguments, "error", err)
		return "", err
	}
	if mayBeResultIsNull == nil {
		db.log.Error("GetQueryString queryRow returned no results", "sql", sql, "args", arguments)
		return "", ErrNoRecordFound
	}
	result = *mayBeResultIsNull
	return result, err
}

func (db *PgxDB) GetVersion(ctx context.Context) (result string, err error) {
	var mayBeResultIsNull *string
	err = db.Conn.QueryRow(ctx, getPGVersion).Scan(&mayBeResultIsNull)
	if err != nil {
		db.log.Error("GetVersion queryRow unexpectedly failed", "error", err)
		return "", err
	}
	if mayBeResultIsNull == nil {
		db.log.Error("GetVersion queryRow returned no results")
		return "", ErrNoRecordFound
	}
	result = *mayBeResultIsNull
	return result, err
}

func (db *PgxDB) GetPGConn() (Conn *pgxpool.Pool, err error) {
	if db.Conn != nil {
		return db.Conn, nil
	} else {
		return nil, ErrDBNotAvailable
	}
}

func (db *PgxDB) HealthCheck(ctx context.Context) (alive bool, err error) {
	dbVersion, err := db.GetVersion(ctx)
	if err != nil || len(dbVersion) < 2 {
		return false, err
	}
	return true, nil
}

func (db *PgxDB) DoesTableExist(ctx context.Context, schema, table string) (exist bool) {
	tableExists, err := db.GetQueryBool(ctx, getTableExists, schema, table)
	if err != nil {
		db.log.Error("DoesTableExist GetQueryBool returned error", "error", err)
		return false
	}
	return tableExists
}

// Close is a postgres helper function to close the connection to the database
func (db *PgxDB) Close() {
	db.Conn.Close()
	return
}
