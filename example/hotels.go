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
	}
}

func main() {
	flag.Parse()
	db.MustInit()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("USAGE: ./hotels 'Hotel Name' 101 102 103 ...")
		return
	}

	var hotelID int
	err := db.DB.QueryRow(`SELECT id FROM hotels WHERE name=$1`, args[0]).Scan(&hotelID)
	if err == db.ErrNotFound {
		fmt.Printf("'%s' doesn't exist, creating...\n", args[0])
		err = db.DB.QueryRow(`INSERT INTO hotels (name) VALUES ($1) RETURNING id`,
			args[0]).Scan(&hotelID)
	}
	if err != nil {
		fmt.Println("error creating hotel: ", err)
		return
	}

	for _, roomNum := range args[1:] {
		// no error checking for brevity of example
		db.DB.Exec(`INSERT INTO rooms (number, hotel) VALUES ($1,$2)`, roomNum, hotelID)
	}
}
