package db

import (
	"os"
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	Customer Role = "customer"
	Admin Role = "admin"
)

type Category struct {
	gorm.Model
	Name string `gorm:"not null;size:255;unique"`
}

type Product struct {
	gorm.Model
	Name string `gorm:"not null;size:255"`
	Price float64 `gorm:"not null"`
	CategoryID string
	Category Category
	Stock uint64 `gorm:"not null,default:0"`
}

type User struct {
	gorm.Model
	Name string `gorm:"not null;size:255"`
	Password string
	Email string `gorm:"not null;unique;size:255"`
	Role Role `sql:"type:role" gorm:"not null;default:'Customer'"`
}

type Order struct {
	gorm.Model
	ProductID string
	Product Product
	UserID string
	User User
	Qty int `gorm:"not null"`
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

	fmt.Println("Running migrations...")
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role') THEN
				create type role AS ENUM ('admin', 'customer');
			END IF;
		END
		$$;
	`)
	db.AutoMigrate(
		Product{},
		User{},
		Category{},
		Order{},
	)
	if err != nil {
		panic(err)
	}

	return db
}

func CreateUser (db *gorm.DB) func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			panic(err)
		}


		pass, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

		user.Password = string(pass)
		createdUser := db.Create(&user)
		w.WriteHeader(201)

		json.NewEncoder(w).Encode(createdUser)
	}
}
