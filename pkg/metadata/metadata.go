package metadata

import (
	"errors"
	"log"
)
import "github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"

const (
	defaultSchema       = "public"
	MetaTableName       = "go_metadata_db_schema"
	CountMetaSQL        = "SELECT COUNT(*) as num FROM go_metadata_db_schema;"
	countMetaServiceSQL = "SELECT COUNT(*) as num FROM go_metadata_db_schema WHERE service = $1;"
	selectMetaSQL       = "SELECT  id, service, schema, table_name, version FROM go_metadata_db_schema WHERE service = $1"
	insertMetaSQL       = "INSERT INTO go_metadata_db_schema (service, schema, table_name, version) VALUES ($1,$2,$3,$4)"
	CreateMetaTable     = `
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

func CreateMetadataTableIfNeeded(db database.DB, log *log.Logger) error {
	if db.DoesTableExist(defaultSchema, MetaTableName) {
		numberOfServicesSchema, errMetaTable := db.GetQueryInt(CountMetaSQL)
		if errMetaTable != nil {
			log.Printf("WARNING: problem counting the rows in metadata table : %v", errMetaTable)
		}
		if numberOfServicesSchema > 0 {
			log.Printf("INFO: 'database contains %d service in metadata'", numberOfServicesSchema)
		} else {
			log.Printf("WARNING: 'database contains %d service in metadata'", numberOfServicesSchema)
		}
		return nil
	} else {
		log.Printf("WARNING: database does not contain the metadata table, will try to create it  ! ")

		RowsAffected, err := db.ExecActionQuery(CreateMetaTable)
		if err != nil {
			log.Printf("ERROR: problem creating the metadata table : %v", err)
			return errors.New("unable to create the table «metadata» ")
		}
		log.Printf("SUCCESS: metadata table was created rows affected : %v", int(RowsAffected))
		return nil
	}
}
