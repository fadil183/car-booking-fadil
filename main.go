package main

import (
	"car-booking-fadil/app/handler"
	"car-booking-fadil/config"
)

func main() {
	db := config.NewDB()

	e := routes.NewRouter(db)
	e.Logger.Fatal(e.Start(":8080"))
}
