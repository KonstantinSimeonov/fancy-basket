package main

import (
	_ "github.com/joho/godotenv/autoload"

	"fmt"
	"os"
	"net/http"
	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	fb "fancybasket/db"
)

func main() {
	db := fb.GetDB();
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/users", func (r chi.Router) {
		r.Post("/", func (w http.ResponseWriter, r *http.Request) {
			fb.CreateUser(db)(w, r)
		})
	})

	r.Route("/products", func (r chi.Router) {
		r.Get("/", func (w http.ResponseWriter, r *http.Request) {
			// TODO: pagination
			var ps []fb.Product
			db.Find(&ps)
			json.NewEncoder(w).Encode(&ps)
		})

		r.Post("/", func (w http.ResponseWriter, r *http.Request) {
			var p fb.Product
			
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				w.WriteHeader(422)
				fmt.Println(err)
				return
			}

			createdProduct := db.Create(&p)

			w.WriteHeader(201)
			json.NewEncoder(w).Encode(createdProduct)
		})
	})

	port := os.Getenv("API_PORT")
	fmt.Printf("Starting up server on port %s\n", port)
	http.ListenAndServe(":" + port, r)
}
