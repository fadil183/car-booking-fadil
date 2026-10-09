package car

import "gorm.io/gorm"

type car struct {
	gorm.Model
	Name            string
	stock_available string
	rental_cost     string
	category        string
}
