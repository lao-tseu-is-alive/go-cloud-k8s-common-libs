package metadata

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

type Service struct {
	Log *slog.Logger
	Db  database.DB
}

const (
	defaultSchema           = "public"
	MetaTableName           = "go_metadata_db_schema"
	CountMetaSQL            = "SELECT COUNT(*) as num FROM go_metadata_db_schema;"
	countMetaServiceSQL     = "SELECT COUNT(*) as num FROM go_metadata_db_schema WHERE service = $1;"
	getVersionServiceSQL    = "SELECT  version FROM go_metadata_db_schema WHERE service = $1"
	updateVersionServiceSQL = "UPDATE public.go_metadata_db_schema SET version = $2 WHERE service = $1"
	selectMetaSQL           = "SELECT  id, service, schema, table_name, version FROM go_metadata_db_schema WHERE service = $1"
	insertMetaSQL           = "INSERT INTO go_metadata_db_schema (service, schema, table_name, version) VALUES ($1,$2,$3,$4)"
	CreateMetaTable         = `
CREATE TABLE IF NOT EXISTS go_metadata_db_schema
(
    id          serial    CONSTRAINT go_metadata_db_schema_pk   primary key,
    service     text                             not null,
    schema      text      default 'public'::text not null,
    table_name  text                             not null,
    version     text                             not null,
    create_time timestamp default now()          not null,
    CONSTRAINT go_metadata_db_schema_unique_service_schema_table
        unique (service, schema, table_name)
);
comment on table go_metadata_db_schema is 'to track version of schema of different micro services';
`
	unableToCountService = "💥💥 unable to count for service %s in the table «metadata». error:%v"
	unableToGetVersion   = "💥💥 unable to retrieve version for service %s in the table «metadata». error:%v"
	unableToSetVersion   = "💥💥 unable to set version for service %s in the table «metadata». error:%v"
)

// CreateMetadataTableOrFail create the metadata table if it is not already present in the db
func (s *Service) CreateMetadataTableOrFail(ctx context.Context) {
	if s.Db.DoesTableExist(ctx, defaultSchema, MetaTableName) {
		numberOfServicesSchema, errMetaTable := s.Db.GetQueryInt(ctx, CountMetaSQL)
		if errMetaTable != nil {
			s.Log.Warn("problem counting the rows in metadata table", "error", errMetaTable)
		}
		if numberOfServicesSchema > 0 {
			s.Log.Info("database contains service(s) in metadata", "count", numberOfServicesSchema)
		} else {
			s.Log.Warn("database does not contain any registered service in metadata table")
		}
		return
	} else {
		s.Log.Warn("database does not contain the metadata table, will try to create it...")

		RowsAffected, err := s.Db.ExecActionQuery(ctx, CreateMetaTable)
		if err != nil {
			s.Log.Error("problem creating the metadata table", "error", err)
			panic(fmt.Errorf("💥💥 unable to create the table «metadata». error:%v", err))
		}
		s.Log.Info("metadata table was created", "rowsAffected", int(RowsAffected))
		return
	}
}

// GetServiceVersionOrFail allows to retrieve the version of the microservice registered (if any) in the metadata table of the db
func (s *Service) GetServiceVersionOrFail(ctx context.Context, serviceName string) (found bool, version string) {
	s.Log.Debug("entering GetServiceVersion", "serviceName", serviceName)
	count, err := s.Db.GetQueryInt(ctx, countMetaServiceSQL, serviceName)
	if err != nil {
		s.Log.Error("GetServiceVersion could not be retrieved from DB", "serviceName", serviceName, "error", err)
		panic(fmt.Errorf(unableToCountService, serviceName, err))
	}
	if count > 0 {
		s.Log.Info("GetServiceVersion service does exist", "serviceName", serviceName)
		version, err := s.Db.GetQueryString(ctx, getVersionServiceSQL, serviceName)
		if err != nil {
			s.Log.Error("GetServiceVersion version could not be retrieved from DB", "serviceName", serviceName, "error", err)
			panic(fmt.Errorf(unableToGetVersion, serviceName, err))
		}
		return true, version
	} else {
		s.Log.Info("GetServiceVersion service does not exist", "serviceName", serviceName)
		return false, ""
	}
}

// SetServiceVersionOrFail allows to insert/update the version of the microservice registered (if any) in the metadata table of the db
func (s *Service) SetServiceVersionOrFail(ctx context.Context, serviceName, version string) {
	s.Log.Debug("entering SetServiceVersion", "serviceName", serviceName, "version", version)
	count, err := s.Db.GetQueryInt(ctx, countMetaServiceSQL, serviceName)
	if err != nil {
		s.Log.Error("SetServiceVersion could not be retrieved from DB", "serviceName", serviceName, "error", err)
		panic(fmt.Errorf(unableToCountService, serviceName, err))
	}
	if count > 0 {
		s.Log.Info("GetServiceVersion service does exist", "serviceName", serviceName)
		versionInDB, err := s.Db.GetQueryString(ctx, getVersionServiceSQL, serviceName)
		if err != nil {
			s.Log.Error("SetServiceVersion version could not be retrieved from DB", "serviceName", serviceName, "error", err)

			panic(fmt.Errorf(unableToGetVersion, serviceName, err))
		}
		if versionInDB == version {
			s.Log.Info("SetServiceVersion service does already exist with this version, nothing to do", "serviceName", serviceName)
			return
		} else {
			rowsAffected, err := s.Db.ExecActionQuery(ctx, updateVersionServiceSQL, serviceName, version)
			if err != nil {
				s.Log.Error("SetServiceVersion version could not be updated in DB", "serviceName", serviceName, "error", err)
				panic(fmt.Errorf(unableToSetVersion, serviceName, err))
			}
			s.Log.Info("SetServiceVersion service updated", "serviceName", serviceName, "version", version, "rowsAffected", rowsAffected)
			return
		}
	} else {
		s.Log.Info("SetServiceVersion service does not exist will insert it", "serviceName", serviceName)
		rowsAffected, err := s.Db.ExecActionQuery(ctx, insertMetaSQL, serviceName, defaultSchema, strings.ToLower(serviceName), version)
		if err != nil {
			s.Log.Error("SetServiceVersion version could not be inserted in DB", "serviceName", serviceName, "error", err)
			panic(fmt.Errorf(unableToSetVersion, serviceName, err))
		}
		s.Log.Info("SetServiceVersion service inserted", "serviceName", serviceName, "version", version, "rowsAffected", rowsAffected)
		return
	}
}
