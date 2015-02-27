// Simple example of package usage. For an example database update and
// migration on this file, see the commit on this file at:
// https://github.com/pbnjay/db/commit/7b414cc742aedc13d064fcddd09cb5d16ae05406
//
// To see the migration in action, build the prior version of this file and
// the latest:
//
//     $ cd $GOPATH/src/github.com/pbnjay/db/example
//     $ git checkout 4546a34
//     $ go build -o hotels_v1 hotels.go
//     $ git checkout master
//     $ go build -o hotels_latest hotels.go
//
//     $ createdb hotels
//     $ ./hotels_v1 -db "dbname=hotels" HotelPlace 101
//     2015/02/27 11:55:59 Creating database
//     2015/02/27 11:55:59 CREATE TABLE ...
//     ...
//     $ ./hotels_latest -db "dbname=hotels" HotelPlace '123 Main St' 202
//     2015/02/27 11:57:04 Performing database migration from 4 -> 5
//     2015/02/27 11:57:04 ALTER TABLE hotels ADD COLUMN address varchar;
//
package main

import (
	"flag"
	"fmt"

	"github.com/pbnjay/db"
)

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
		`ALTER TABLE hotels ADD COLUMN address varchar;`,
	}
}

func main() {
	flag.Parse()
	db.MustInit()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("USAGE: ./hotels 'Hotel Name' 'Hotel Address' 101 102 103 ...")
		return
	}

	var hotelID int
	err := db.DB.QueryRow(`SELECT id FROM hotels WHERE name=$1`, args[0]).Scan(&hotelID)
	if err == nil {
		db.DB.Exec("UPDATE hotels SET address=$2 WHERE id=$1", hotelID, args[1])
	}
	if err == db.ErrNotFound {
		fmt.Printf("'%s' doesn't exist, creating...\n", args[0])
		err = db.DB.QueryRow(`INSERT INTO hotels (name,address) VALUES ($1,$2) RETURNING id`,
			args[0], args[1]).Scan(&hotelID)
	}
	if err != nil {
		fmt.Println("error creating hotel: ", err)
		return
	}

	for _, roomNum := range args[2:] {
		// no error checking for brevity of example
		db.DB.Exec(`INSERT INTO rooms (number, hotel) VALUES ($1,$2)`, roomNum, hotelID)
	}
}
