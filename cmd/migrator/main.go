package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate"

	_ "github.com/golang-migrate/migrate/source/file" 
	_ "github.com/golang-migrate/migrate/database/postgres"
)

func main() {
	var user, password, host, port, dbName, sslmode, migrationsPath string

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&user, "db-user", "postgres", "database user")
	flag.StringVar(&password, "db-password", "", "database password")
	flag.StringVar(&host, "db-host", "localhost", "database host")
	flag.StringVar(&port, "db-port", "5432", "database port")
	flag.StringVar(&dbName, "db-name", "wallet_db", "database name")
	flag.StringVar(&sslmode, "sslmode", "disable", "SSL mode")
	flag.Parse()

	
	if migrationsPath == "" {
		panic("migrations-path is required")
	}



	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbName, sslmode),
	)

	if err != nil {
		panic(err)
	}
	
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied successfull")
}
