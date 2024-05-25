package commands

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"project/internal/pkg/repository/postgresql"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")

type Scheme struct {
	Index       int
	Description string
	Query       string
}

var scheme = []Scheme{
	{
		Index:       1,
		Description: "Create table: users.",
		Query: `
				CREATE TABLE IF NOT EXISTS users (
                                     id serial primary key,
                                     login text not null,
                                     password text not null,
                                     full_name text,
                                     avatar text,
                                     role text not null,
                                     status bool default false,
                                     phone text,
                   					 created_at timestamp default now(),
                   					 created_by int references users(id),
                   					 updated_at timestamp,
                   					 updated_by int references users(id),
                   					 deleted_at timestamp,
                   					 deleted_by int references users(id)
				);
			`,
	},
	{
		Index:       2,
		Description: "Create area with login:admin, password: 1",
		Query: `
				INSERT INTO users(login, role, password)
				SElECT 'Admin','ADMIN', '$2a$10$NKtnMwDPFSQLG6uOi4Zqheru5Ygbj9TWFHjpl478rRSaO5cJ9QuH2' WHERE NOT EXISTS (SELECT login FROM users WHERE login = 'Admin');
			`,
	},
}

// Migrate creates the scheme in the database.
func Migrate(db *postgresql.Database) {
	for _, s := range scheme {
		if _, err := db.Query(s.Query); err != nil {
			log.Fatalln("migrate error", err)
		}
	}
}

func MigrateUP(db *postgresql.Database) {
	var (
		version int
		dirty   bool
		er      *string
	)
	err := db.QueryRow("SELECT version,dirty,error FROM schema_migrations").Scan(&version, &dirty, &er)
	if err != nil {
		if err.Error() == `ERROR: relation "schema_migrations" does not exist (SQLSTATE=42P01)` {
			if _, err = db.Exec(`
										CREATE TABLE IF NOT EXISTS schema_migrations (version int not null,dirty bool not null ,error text);
										DELETE FROM schema_migrations;
										INSERT INTO schema_migrations (version, dirty) values (0,false);
								`); err != nil {
				log.Fatalln("migrate schema_migrations create error", err)
			}
			version = 0
			dirty = false
		} else {
			log.Fatalln("migrate schema_migrations scan: ", err)
		}

	}

	if dirty {
		for _, v := range scheme {
			if v.Index == version {
				if _, err = db.Exec(v.Query); err != nil {
					if _, err = db.Exec(fmt.Sprintf(`UPDATE schema_migrations SET error = '%s'`, err.Error())); err != nil {
						log.Fatalln("migrate error", err)
					}
					log.Fatalln(fmt.Sprintf("migrate error version: %d", version), err)
				}
				if _, err = db.Exec(fmt.Sprintf(`UPDATE schema_migrations SET dirty = false, error = null`)); err != nil {
					log.Fatalln("migrate error", err)
				}
			}
		}
	}

	for _, s := range scheme {
		if s.Index > version {
			if _, err = db.Exec(s.Query); err != nil {
				if _, err = db.Exec(fmt.Sprintf(`UPDATE schema_migrations SET error = '%s', version = %d, dirty = true`, err.Error(), s.Index)); err != nil {
					log.Fatalln("migrate error", err)
				}
				log.Fatalln(fmt.Sprintf("migrate error version: %d", s.Index), err)
			}
			if _, err = db.Exec(fmt.Sprintf(`UPDATE schema_migrations SET version = %d`, s.Index)); err != nil {
				log.Fatalln("migrate error", err)
			}
		}
	}
}
