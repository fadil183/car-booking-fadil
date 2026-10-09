package user

import (
	"car-booking-fadil/app/handler/encryption"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type newUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GormRepository struct {
	*gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{
		db.Table("users"),
	}
}

// AddNewUser adds a user from an API request to the database.
func (r *GormRepository) AddNewUser(c echo.Context) error {
	var user newUser
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if user.Email == "" || user.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	user.Password = encryption.HashPassword(user.Password)

	if err := r.Create(&user).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "new user has been created",
	})
}
