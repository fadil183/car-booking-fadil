package user

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type newEmployee struct {
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
	var employee newEmployee
	if err := c.Bind(&employee); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := r.Create(&employee).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "new user has been created",
	})
}
