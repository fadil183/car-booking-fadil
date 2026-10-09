package user

import "gorm.io/gorm"

type User struct{
	gorm.Model
	Email string `validate:"required"`
	Password string `validate:"required"`
	Deposit string
	
}
