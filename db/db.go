package db

import (
	"os"
	"fmt"
	"net/http"
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	Customer Role = "customer"
	Admin Role = "admin"
)

type Base struct {
	ID string `json:"id" gorm:"primary_key;type:uuid;not null;default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"createdAt" gorm:"not null"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"not null"`
	DeletedAt *time.Time `json:"deletedAt" sql:"index"`
}

type Category struct {
	Base
	Name string `json:"name" gorm:"not null;size:255;unique"`
}

type Product struct {
	Base
	Name string `json:"name" gorm:"not null;size:255"`
	Price float64 `json:"price" gorm:"not null"`
	CategoryID string `json:"categoryId"`
	Category Category `json:"-"`
	Stock uint64 `json:"stock" gorm:"not null,default:0"`
}

type User struct {
	Base
	Name string `json:"name" gorm:"not null;size:255"`
	Password string `json:"-"`
	Email string `json:"email" gorm:"not null;unique;size:255"`
	Role Role `json:"role" sql:"type:role" gorm:"not null;default:'customer'"`
}

type Order struct {
	Base
	ProductID string `json:"productId"`
	Product Product `json:"-"`
	UserID string `json:"userId"`
	User User `json:"-"`
	Qty int `json:"qty" gorm:"not null"`
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
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	`)
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
