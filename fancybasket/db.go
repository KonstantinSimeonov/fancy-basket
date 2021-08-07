package basketdb

import (
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type Product struct {
	gorm.Model
	Name string
	Price float64
}

func GetDB() gorm.Session {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=fancybasket dbname=fancybasket sslmode=disable password=shoppingfart")

	db.AutoMigrate(&Product{})
	if err != nil {
		panic(err)
	}

	return db
}
