## goCloudK8sExampleServer
here is a simple golang server illustrating usage of this library

### db migrations integration
we use [go-migrate](https://github.com/golang-migrate/migrate) to give an example of embedding db-migrations sql scripts inside a server deployment.

You can check the [postgres tutorial](https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md) for go-migrate to gent more information.

inside this directory on your shell just use something like : 

    migrate create -ext sql -dir db/migrations -seq create_demo_type_table

will create those two files in the db/migrations directory:

    000001_create_demo_table.down.sql
    000001_create_demo_table.up.sql

then you can enter the corresponding up and down sql code.

whatever sql migrations are present on the db/migrations directory   
will be used by the code near line 190
```golang
        // example of go-migrate db migration with embed files in go program
	// https://github.com/golang-migrate/migrate
	d, err := iofs.New(sqlMigrations, defaultSqlDbMigrationsPath)
	if err != nil {
		l.Fatalf("💥💥 error doing iofs.New for db migrations  error: %v\n", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, strings.Replace(dbDsn, "postgres", "pgx", 1))
	if err != nil {
		l.Fatalf("💥💥 error doing migrate.NewWithSourceInstance(iofs, dbURL:%s)  error: %v\n", dbDsn, err)
	}
	err = m.Up()
	if err != nil {
		l.Fatalf("💥💥 error doing migrate.Up error: %v\n", err)
	}
```