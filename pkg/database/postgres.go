package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/metadata"
	"log"
)

const getPGVersion = "SELECT version();"

var (
	ErrNoRecordFound     = errors.New("record not found")
	ErrCouldNotBeCreated = errors.New("could not be created in DB")
)

type PgxDB struct {
	Conn *pgxpool.Pool
	log  *log.Logger
}

func GetPgxConn(dbConnectionString string, maxConnectionsInPool int, log *log.Logger) (*PgxDB, error) {
	var psql PgxDB
	var parsedConfig *pgx.ConnConfig
	var err error
	parsedConfig, err = pgx.ParseConfig(dbConnectionString)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error doing pgx.ParseConfig(%s). err: %s", dbConnectionString, err))
	}

	dbHost := parsedConfig.Host
	dbPort := parsedConfig.Port
	dbUser := parsedConfig.User
	dbPass := parsedConfig.Password
	dbName := parsedConfig.Database

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s pool_max_conns=%d", dbHost, dbPort, dbUser, dbPass, dbName, maxConnectionsInPool)

	connPool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Printf("ERROR: FAILED to connect to database %s with user %s\n", dbName, dbUser)
		return nil, errors.New(fmt.Sprintf("error connecting to database. err : %s", err))
	} else {
		log.Printf("INFO: 'SUCCESS connecting to database %s with user %s'\n", dbName, dbUser)
		// let's first check that we can really make a query by querying the postgres version
		var version string
		errPing := connPool.QueryRow(context.Background(), getPGVersion).Scan(&version)
		if errPing != nil {
			log.Printf("Hoho something very weird is occurring here... this db connection is invalid ! ")
			log.Fatalf("FATAL DB ERROR: testing db connection with : [%s] error: %s", getPGVersion, errPing)
			return nil, errPing
		}
		var numberOfServicesSchema int
		errMetaTable := connPool.QueryRow(context.Background(), metadata.CountMetaSQL).Scan(&numberOfServicesSchema)
		if errMetaTable != nil {
			log.Printf("WARNING: problem counting the rows in metadata table : %v", errMetaTable)
			log.Printf("WARNING: database does not contain the metadata table, will try to create it  ! ")
			commandTag, err := connPool.Exec(context.Background(), metadata.CreateMetaTable)
			if err != nil {
				log.Printf("ERROR: problem creating the metadata table : %v", err)
				return nil, errors.New("unable to create the table «metadata» ")
			}
			log.Printf("SUCCESS: metadata table was created rows affected : %v", int(commandTag.RowsAffected()))
		}

		log.Printf("INFO: 'Postgres version: [%s]'", version)
		if numberOfServicesSchema > 0 {
			log.Printf("INFO: 'database contains %d service in metadata'", numberOfServicesSchema)
		} else {
			log.Printf("WARNING: 'database contains %d service in metadata'", numberOfServicesSchema)
		}
	}

	psql.Conn = connPool
	psql.log = log
	return &psql, err
}

// GetQueryInt is a postgres helper function for a query expecting an integer result
func (db *PgxDB) GetQueryInt(sql string, arguments ...interface{}) (result int, err error) {
	err = db.Conn.QueryRow(context.Background(), sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Printf("error : GetQueryInt(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return 0, err
	}
	return result, err
}

func (db *PgxDB) GetQueryString(sql string, arguments ...interface{}) (result string, err error) {
	var mayBeResultIsNull *string
	err = db.Conn.QueryRow(context.Background(), sql, arguments...).Scan(&mayBeResultIsNull)
	if err != nil {
		db.log.Printf("error : GetQueryString(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return "", err
	}
	if mayBeResultIsNull == nil {
		db.log.Printf("error : GetQueryString() queryRow returned no results with sql: %v ; parameters:(%v)\n", sql, arguments)
		return "", ErrNoRecordFound
	}
	result = *mayBeResultIsNull
	return result, err
}

// GetQueryBool is a postgres helper function for a query expecting an integer result
func (db *PgxDB) GetQueryBool(sql string, arguments ...interface{}) (result bool, err error) {
	err = db.Conn.QueryRow(context.Background(), sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Printf("error : GetQueryBool(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return false, err
	}
	return result, err
}

// ExecActionQuery is a postgres helper function for an action query, returning the numbers of rows affected
func (db *PgxDB) ExecActionQuery(sql string, arguments ...interface{}) (rowsAffected int, err error) {
	commandTag, err := db.Conn.Exec(context.Background(), sql, arguments...)
	if err != nil {
		db.log.Printf("ExecActionQuery unexpectedly failed with sql: %v . Args(%+v), error : %v", sql, arguments, err)
		return 0, err
	}
	return int(commandTag.RowsAffected()), err
}

// Close is a postgres helper function to close the connection to the database
func (db *PgxDB) Close() {
	db.Conn.Close()
	return
}
