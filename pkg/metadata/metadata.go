package metadata

const (
	defaultServiceSchema = "public"
	CountMetaSQL         = "SELECT COUNT(*) as num FROM go_metadata_db_schema;"
	countMetaServiceSQL  = "SELECT COUNT(*) as num FROM go_metadata_db_schema WHERE service = $1;"
	selectMetaSQL        = "SELECT  id, service, schema, table_name, version FROM go_metadata_db_schema WHERE service = $1"
	insertMetaSQL        = "INSERT INTO go_metadata_db_schema (service, schema, table_name, version) VALUES ($1,$2,$3,$4)"
	CreateMetaTable      = `
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
