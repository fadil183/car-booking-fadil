package main

import (
	"car-booking-fadil/config"
	"log"
)

func main() {
	db := config.NewDB()

	log.Println(db)

}
