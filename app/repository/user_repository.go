package user

import (
	"car-booking-fadil/app/handler/encryption"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GormRepository struct {
	*gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	// Session makes each query start from a clean statement instead of accumulating conditions.
	return &GormRepository{
		DB: db.Table("users").Session(&gorm.Session{}),
	}
}

// AddNewUser adds a user from an API request to the database.
func (r *GormRepository) SaveUser(c echo.Context) error {
	var user User
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	user.Email = strings.TrimSpace(user.Email)
	if user.Email == "" || strings.TrimSpace(user.Password) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	hashed, err := encryption.HashPassword(user.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process password")
	}
	user.Password = hashed

	if err := r.Create(&user).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "new user has been created",
	})
}

func (r *GormRepository) Login(c echo.Context) error {
	var req User
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	var stored struct {
		ID       uint   `gorm:"column:userid"`
		Email    string `gorm:"column:email"`
		Password string `gorm:"column:password"`
	}
	err := r.Where("email = ?", req.Email).Take(&stored).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process login")
	}

	// Same message for unknown email and wrong password to avoid user enumeration.
	if err != nil || !encryption.VerifyPassword(req.Password, stored.Password) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "login success",
	})
}

type topUpRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Amount   float64 `json:"amount"`
}

// TopUp adds amount to the user's deposit_amount after verifying email and password.
func (r *GormRepository) TopUp(c echo.Context) error {
	var req topUpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}
	if req.Amount <= 0 || math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) {
		return echo.NewHTTPError(http.StatusBadRequest, "amount must be greater than 0")
	}

	userID, err := r.authenticate(req.Email, req.Password)
	if errors.Is(err, errInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process top up")
	}

	var balance float64
	err = r.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw("UPDATE users SET deposit_amount = deposit_amount + ? WHERE userid = ? RETURNING deposit_amount",
			req.Amount, userID).Scan(&balance).Error; err != nil {
			return err
		}
		return tx.Exec(`INSERT INTO history (userid, type, amount, balance_after, description)
			VALUES (?, 'TOPUP', ?, ?, 'Deposit top up')`, userID, req.Amount, balance).Error
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to top up deposit")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":        "top up success",
		"deposit_amount": balance,
	})
}

type carResponse struct {
	ID           uint    `gorm:"column:carid" json:"car_id"`
	Name         string  `gorm:"column:name" json:"name"`
	Transmission string  `gorm:"column:transmission" json:"transmission"`
	Availability int     `gorm:"column:availability" json:"availability"`
	RentalCost   float64 `gorm:"column:rental_cost" json:"rental_cost"`
	Category     string  `gorm:"column:category" json:"category"`
}

// GetAllRentCarAvailable returns all cars whose availability is greater than 0.
func (r *GormRepository) GetAllRentCarAvailable(c echo.Context) error {
	cars := []carResponse{}
	if err := r.Table("cars").Where("availability > ?", 0).Order("carid").Find(&cars).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get available cars")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "available cars",
		"total":   len(cars),
		"cars":    cars,
	})
}

var (
	errCarNotFound     = errors.New("car not found")
	errCarUnavailable  = errors.New("car not available")
	errInsufficientBal = errors.New("insufficient deposit")
)

// authenticate verifies the email and password posted by the user and returns the user ID.
func (r *GormRepository) authenticate(email, password string) (uint, error) {
	var stored struct {
		ID       uint   `gorm:"column:userid"`
		Password string `gorm:"column:password"`
	}
	err := r.Where("email = ?", email).Take(&stored).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	if err != nil || !encryption.VerifyPassword(password, stored.Password) {
		return 0, errInvalidCredentials
	}
	return stored.ID, nil
}

var errInvalidCredentials = errors.New("invalid email or password")

type bookCarRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	CarID     uint   `json:"car_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// BookCar books a car for [start_date, end_date), charging rental_cost per day from the user's deposit.
func (r *GormRepository) BookCar(c echo.Context) error {
	var req bookCarRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}
	if req.CarID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "car_id is required")
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "start_date must use format YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must use format YYYY-MM-DD")
	}
	today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	if start.Before(today) {
		return echo.NewHTTPError(http.StatusBadRequest, "start_date cannot be in the past")
	}
	days := int(end.Sub(start).Hours() / 24)
	if days < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must be after start_date")
	}

	userID, err := r.authenticate(req.Email, req.Password)
	if errors.Is(err, errInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process booking")
	}

	var bookID uint
	var totalCost, balanceAfter float64
	err = r.Transaction(func(tx *gorm.DB) error {
		var car struct {
			RentalCost float64 `gorm:"column:rental_cost"`
		}
		res := tx.Raw("SELECT rental_cost FROM cars WHERE carid = ?", req.CarID).Scan(&car)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errCarNotFound
		}
		totalCost = car.RentalCost * float64(days)

		res = tx.Exec("UPDATE cars SET availability = availability - 1 WHERE carid = ? AND availability > 0", req.CarID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errCarUnavailable
		}

		res = tx.Raw(`UPDATE users SET deposit_amount = deposit_amount - ?
			WHERE userid = ? AND deposit_amount >= ? RETURNING deposit_amount`,
			totalCost, userID, totalCost).Scan(&balanceAfter)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errInsufficientBal
		}

		if err := tx.Raw(`INSERT INTO books (carid, userid, booking_period)
			VALUES (?, ?, daterange(?::date, ?::date, '[)')) RETURNING bookid`,
			req.CarID, userID, req.StartDate, req.EndDate).Scan(&bookID).Error; err != nil {
			return err
		}

		return tx.Exec(`INSERT INTO history (userid, bookid, carid, type, amount, balance_after, description)
			VALUES (?, ?, ?, 'BOOKING', ?, ?, ?)`,
			userID, bookID, req.CarID, totalCost, balanceAfter,
			fmt.Sprintf("Car booking %s to %s (%d days)", req.StartDate, req.EndDate, days)).Error
	})
	switch {
	case errors.Is(err, errCarNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "car not found")
	case errors.Is(err, errCarUnavailable):
		return echo.NewHTTPError(http.StatusConflict, "car is not available")
	case errors.Is(err, errInsufficientBal):
		return echo.NewHTTPError(http.StatusPaymentRequired, "insufficient deposit, please top up")
	case err != nil:
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to book car")
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message":    "booking success",
		"book_id":    bookID,
		"car_id":     req.CarID,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"days":       days,
		"total_cost": totalCost,
	})
}

type historyEntry struct {
	HistoryID    uint      `gorm:"column:historyid" json:"history_id"`
	Type         string    `gorm:"column:type" json:"type"`
	BookID       *uint     `gorm:"column:bookid" json:"book_id"`
	CarID        *uint     `gorm:"column:carid" json:"car_id"`
	CarName      *string   `gorm:"column:car_name" json:"car_name"`
	Amount       float64   `gorm:"column:amount" json:"amount"`
	BalanceAfter float64   `gorm:"column:balance_after" json:"balance_after"`
	Description  *string   `gorm:"column:description" json:"description"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

// History returns the top up and booking history of the user identified by the posted email and password.
func (r *GormRepository) History(c echo.Context) error {
	var req User
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	userID, err := r.authenticate(req.Email, req.Password)
	if errors.Is(err, errInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process history")
	}

	history := []historyEntry{}
	err = r.Raw(`SELECT h.historyid, h.type, h.bookid, h.carid, c.name AS car_name,
			h.amount, h.balance_after, h.description, h.created_at
		FROM history h
		LEFT JOIN cars c ON c.carid = h.carid
		WHERE h.userid = ?
		ORDER BY h.created_at DESC, h.historyid DESC`, userID).Scan(&history).Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get history")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "user history",
		"email":   req.Email,
		"total":   len(history),
		"history": history,
	})
}
