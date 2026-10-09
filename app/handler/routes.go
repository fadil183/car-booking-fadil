package routes

import (
	userRepository "car-booking-fadil/app/repository"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *echo.Echo {
	e := echo.New()
	users := userRepository.NewGormRepository(db)

	e.POST("/users", users.AddNewUser)

	return e
}
