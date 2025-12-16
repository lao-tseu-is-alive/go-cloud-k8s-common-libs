package f5

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

type PGX struct {
	Conn *pgxpool.Pool
	dbi  database.DB
	log  *slog.Logger
}

func (db *PGX) Get(ctx context.Context, login string) (*User, error) {
	db.log.Debug("entering Get", "login", login)
	if !db.Exist(ctx, login) {
		msg := fmt.Sprintf(UserDoesNotExist, login)
		db.log.Warn(msg)
		return nil, errors.New(msg)
	}
	res := &User{}
	err := pgxscan.Get(ctx, db.Conn, res, getUser, login)
	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "function", "Get", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info(FunctionNReturnedNoResults, "function", "Get")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

func (db *PGX) Exist(ctx context.Context, login string) bool {
	db.log.Debug("entering Exist", "login", login)
	count, err := db.dbi.GetQueryInt(ctx, existUser, login)
	if err != nil {
		db.log.Error("Exist could not be retrieved from DB", "login", login, "error", err)
		return false
	}
	if count > 0 {
		db.log.Info("Exist id does exist", "login", login, "count", count)
		return true
	} else {
		db.log.Info("Exist id does not exist", "login", login, "count", count)
		return false
	}
}

// NewPgxDB will instantiate a new storage of type postgres and ensure schema exist
func NewPgxDB(db database.DB, log *slog.Logger) (Storage, error) {
	var psql PGX
	pgConn, err := db.GetPGConn()
	if err != nil {
		return nil, err
	}
	psql.Conn = pgConn
	psql.dbi = db
	psql.log = log
	var postgresVersion string
	errVersionPostgres := pgConn.QueryRow(context.Background(), getPostgresVersion).Scan(&postgresVersion)
	if errVersionPostgres != nil {
		log.Error("Unable to retrieve the postgres version", "error", err)
		return nil, errVersionPostgres
	}
	log.Info("connected to postgres database", "version", postgresVersion)
	//check if table for F5 authentication is present with corresponding fields
	var numEmploye int
	errNumEmploye := pgConn.QueryRow(context.Background(), countUsers).Scan(&numEmploye)
	if errNumEmploye != nil {
		log.Error("Unable to count number of rows in table employe")
		return nil, errNumEmploye
	}
	log.Info("found rows in table employe", "count", numEmploye)
	//check if all fields are present
	res := &User{}
	err = pgxscan.Get(context.Background(), pgConn, res, checkUserFields)
	if err != nil {
		log.Error(SelectFailedInNWithErrorE, "function", "Get", "error", err)
		return nil, err
	}
	log.Info("found all fields in table employe", "id", res.Id)
	return &psql, err
}
