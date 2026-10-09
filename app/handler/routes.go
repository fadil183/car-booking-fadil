package routes

import (
	userRepository "car-booking-fadil/app/repository"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *echo.Echo {
	e := echo.New()
	users := userRepository.NewGormRepository(db)

	e.POST("/register", users.SaveUser)
	e.POST("/login", users.Login)

	e.POST("/topup", users.TopUp)
	e.POST("/rent", users.GetAllRentCarAvailable)
	e.POST("/books", users.BookCar)
	e.POST("/history", users.History)

	return e
}
