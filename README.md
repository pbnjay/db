# db
A ridiculously simple Go database/sql wrapper that supports automatic schema
migrations. It is mainly intented to reduce boilerplate database setup for
simple projects, NOT for multi-database setups, complex migrations, dependency
management, etc.

[![GoDoc](https://godoc.org/github.com/pbnjay/db?status.svg)](http://godoc.org/github.com/pbnjay/db)

DDL statements are appended to a .go file containing your schema creation
statements in an array. This array should be though of as append-only once DDL
statements are committed to version control.

Although the interface remains simple, this package forces you to think about
your database migrations during development, in addition to keeping code and
database DDL in sync in the same source repository.

## Example Usage

To use this package, you'd simply define your schema in an imported init()
function like this:

```go
func init() {
  db.Schema = []string{
    `CREATE TABLE hotels (
        id serial primary key,
        name varchar
     );`,
    `CREATE TABLE rooms (
        id serial primary key,
        number integer,
        hotel int references hotels(id)
     );`,
  }
}

func main() {
  db.MustInit("dbname=hotels")
  db.DB.Exec(`INSERT INTO hotels (name) VALUES ($1)`, "Hello World Hotel")
}
```

Then after future development occurs, schema migration statements are
appended to the init() function's db.Schema array. For example, to add an
address to the hotels table above:

```diff
         number integer,
         hotel int references hotels(id)
      );`,
+    `ALTER TABLE hotels ADD COLUMN address varchar;`,
   }
 }

 func main() {
   db.MustInit("dbname=hotels")
-  db.DB.Exec(`INSERT INTO hotels (name) VALUES ($1)`, "Hello World Hotel")
+  db.DB.Exec(`INSERT INTO hotels (name, address) VALUES ($1, $2)`,
+      "Hello World Hotel", "123 Main St")
 }
```

## TODOs

 - Create a git hook command to verify immutable db.Schema updates. Issue #1
 - Create (optional) reverse migration DDL and interface methods. Issue #2

## License

MIT
