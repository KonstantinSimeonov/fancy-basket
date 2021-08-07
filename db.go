package fancybasket

import (
	"os"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"
)

type Product struct {
	gorm.Model
	Name string
	Price float64
}

type User struct {
	gorm.Model
	Name string
	Password string
	Email string
}

func GetDB() *gorm.DB {
	opts := fmt.Sprintf(
		"host=%s user=%s dbname=%s port=%s sslmode=disable password=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_PASSWORD"),
	)
	db, err := gorm.Open(
		"postgres",
		opts,
	)

	db.AutoMigrate(&Product{})
	db.AutoMigrate(&User{})
	if err != nil {
		panic(err)
	}

	return db
}

func CreateUser (db *gorm.DB) {
	return func (w http.ResponseWrite, r *http.Request) {
		user := &models.User{}

		pass, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

		user.Password = string(pass)
		createdUser = db.Create(user)

		json.NewEncodder(w).Encode(createdUser)
	}
}
