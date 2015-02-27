// Package db provides a ridiculously simple database/sql wrapper that supports
// automatic schema migrations. It handles connection initialization, schema
// version checks and migration using your provided schema DDL statements.
//
// Example usage:
//
//     func init() {
//       db.Schema = []string{
//         `CREATE TABLE hotels (
//             id serial primary key,
//             name varchar
//          );`,
//         `CREATE TABLE rooms (
//             id serial primary key,
//             number integer,
//             hotel int references hotels(id)
//          );`,
//       }
//     }
//
//     func main() {
//       db.MustInit("dbname=hotels")
//       db.DB.Exec(`INSERT INTO hotels (name) VALUES ($1)`, "Hello World Hotel")
//     }
//
// Then after future development occurs, schema migration statements are
// appended to the init() function's db.Schema setup. For example, to add an
// address to the hotels table above:
//
//              number integer,
//              hotel int references hotels(id)
//           );`,
//     +    `ALTER TABLE hotels ADD COLUMN address varchar;`,
//        }
//      }
//
//      func main() {
//        db.MustInit("dbname=hotels")
//     -  db.DB.Exec(`INSERT INTO hotels (name) VALUES ($1)`, "Hello World Hotel")
//     +  db.DB.Exec(`INSERT INTO hotels (name, address) VALUES ($1, $2)`,
//     +      "Hello World Hotel", "123 Main St")
//      }
//
package db

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/lib/pq"
)

var preSchema = []string{
	`CREATE TABLE IF NOT EXISTS _db_meta (
      key VARCHAR PRIMARY KEY,
      val VARCHAR NOT NULL
    );`,
	`INSERT INTO _db_meta (key,val) VALUES ('version', '2');`,
}

var (
	// ErrNotFound is provided here so there's no need to import database/sql
	// just to check for empty results.
	ErrNotFound = sql.ErrNoRows

	// Schema contains the DDL statements to create your entire database from
	// scratch. When developing your project, ONLY append DDL statements to this
	// array so that existing installs can be migrated properly. This also allows
	// your version control to track database schema updates cleanly.
	Schema = []string{}

	// DB contains a sql.DB instance (which is opened, tested, and migrated
	// automatically by calling Init).
	DB *sql.DB

	dbConnString = flag.String("db", "", "postgresql database connection string")
)

// Init the database using the optional connection string (if provided), if the
// `flag` package is being used, a String "db" parameter is parsed, and if set
// will override any parameters here.
func Init(connStrs ...string) error {
	connStr := ""
	if len(connStrs) > 0 {
		connStr = connStrs[0]
	}
	if *dbConnString != "" {
		connStr = *dbConnString
	}
	// err ignored because postgres driver will never throw one here
	DB, _ = sql.Open("postgres", connStr)
	// dummy query, forces connection attempt so we get an error if invalid
	row := DB.QueryRow("SELECT 0")
	i := 0
	err := row.Scan(&i)
	if err != nil {
		return err
	}

	// join the metaschema and user schemas into one slice
	dbSchema := append(preSchema, Schema...)

	// check current schema version
	err = DB.QueryRow("SELECT val FROM _db_meta WHERE key='version'").Scan(&i)
	if err != nil {
		// either meta table doesn't exist, or version key isn't set.
		// either way, need to run an update.
		i = 0
	}
	if i != len(dbSchema) {
		if i == 0 {
			log.Println("Creating database")
		} else {
			log.Printf("Performing database migration from %d -> %d\n", i, len(dbSchema))
		}
	}
	for ; i < len(dbSchema); i++ {
		// use a transaction to ensure the schema+version update is atomic
		tx, err := DB.Begin()
		if err != nil {
			return err
		}
		log.Println(dbSchema[i])
		_, err = tx.Exec(dbSchema[i])
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec("UPDATE _db_meta SET val=$1 WHERE key='version'", i+1)
		if err != nil {
			tx.Rollback()
			return err
		}
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

// MustInit invokes Init but panics on any errors.
func MustInit(connStrs ...string) {
	err := Init(connStrs...)
	if err != nil {
		panic(err)
	}
}
