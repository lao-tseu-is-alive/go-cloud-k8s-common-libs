package metadata

import (
	"errors"
	"log"
	"strings"
)
import "github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"

type Service struct {
	Log *log.Logger
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
)

// CreateMetadataTableIfNeeded create the metadata table if needed in the db
func (s *Service) CreateMetadataTableIfNeeded() error {
	if s.Db.DoesTableExist(defaultSchema, MetaTableName) {
		numberOfServicesSchema, errMetaTable := s.Db.GetQueryInt(CountMetaSQL)
		if errMetaTable != nil {
			s.Log.Printf("WARNING: problem counting the rows in metadata table : %v\n", errMetaTable)
		}
		if numberOfServicesSchema > 0 {
			s.Log.Printf("INFO: database contains %d service(s) in metadata\n", numberOfServicesSchema)
		} else {
			s.Log.Print("WARNING: database does not contain any  registered service in metadata table")
		}
		return nil
	} else {
		s.Log.Printf("WARNING: database does not contain the metadata table, will try to create it...\n")

		RowsAffected, err := s.Db.ExecActionQuery(CreateMetaTable)
		if err != nil {
			s.Log.Printf("ERROR: problem creating the metadata table : %v\n", err)
			return errors.New("unable to create the table «metadata»")
		}
		s.Log.Printf("SUCCESS: metadata table was created, rows affected : %v\n", int(RowsAffected))
		return nil
	}
}

// GetServiceVersion allows to retrieve the version of the microservice registered (if any) in the metadata table of the db
func (s *Service) GetServiceVersion(serviceName string) (found bool, version string, err error) {
	s.Log.Printf("trace : entering GetServiceVersion(%s)", serviceName)
	count, err := s.Db.GetQueryInt(countMetaServiceSQL, serviceName)
	if err != nil {
		s.Log.Printf("error: GetServiceVersion(%s) could not be retrieved from DB. failed db.Query err: %v", serviceName, err)
		return false, "", err
	}
	if count > 0 {
		s.Log.Printf("info : GetServiceVersion(%s) service does exist\n", serviceName)
		version, err := s.Db.GetQueryString(getVersionServiceSQL, serviceName)
		if err != nil {
			s.Log.Printf("error: GetServiceVersion(%s) version could not be retrieved from DB. failed db.Query err: %v", serviceName, err)
			return true, "", err
		}
		return true, version, nil
	} else {
		s.Log.Printf("info : GetServiceVersion(%s) service does not exist\n", serviceName)
		return false, "", nil
	}
}

// SetServiceVersion allows to insert/update the version of the microservice registered (if any) in the metadata table of the db
func (s *Service) SetServiceVersion(serviceName, version string) error {
	s.Log.Printf("trace : entering SetServiceVersion(%s, %s)", serviceName, version)
	count, err := s.Db.GetQueryInt(countMetaServiceSQL, serviceName)
	if err != nil {
		s.Log.Printf("error: SetServiceVersion(%s) could not be retrieved from DB. failed db.Query err: %v", serviceName, err)
		return err
	}
	if count > 0 {
		s.Log.Printf("info : GetServiceVersion(%s) service does exist\n", serviceName)
		versionInDB, err := s.Db.GetQueryString(getVersionServiceSQL, serviceName)
		if err != nil {
			s.Log.Printf("error: SetServiceVersion(%s) version could not be retrieved from DB. failed db.Query err: %v", serviceName, err)
			return err
		}
		if versionInDB == version {
			s.Log.Printf("info : SetServiceVersion(%s) service does already exist with this version, nothing to do\n", serviceName)
			return nil
		} else {
			rowsAffected, err := s.Db.ExecActionQuery(updateVersionServiceSQL, serviceName, version)
			if err != nil {
				s.Log.Printf("error: SetServiceVersion(%s) version could not be updated in DB. failed db.Query err: %v", serviceName, err)
				return err
			}
			s.Log.Printf("info : SetServiceVersion(%s) service updated to version: %s (%d rows)\n", serviceName, version, rowsAffected)
			return nil
		}
	} else {
		s.Log.Printf("info : SetServiceVersion(%s) service does not exist will insert it\n", serviceName)
		rowsAffected, err := s.Db.ExecActionQuery(insertMetaSQL, serviceName, defaultSchema, strings.ToLower(serviceName), version)
		if err != nil {
			s.Log.Printf("error: SetServiceVersion(%s) version could not be inserted in DB. failed db.Query err: %v", serviceName, err)
			return err
		}
		s.Log.Printf("info : SetServiceVersion(%s) service inserted with version: %s (%d rows)\n", serviceName, version, rowsAffected)
		return nil
	}
}
