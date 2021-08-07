package main

import (
	"net/http"
	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	fb "fancybasket"
)



func getProducts(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("[{ \"id\": 5 }]"))
}



func main() {
	db := fb.GetDB();
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/products", func (r chi.Router) {
		r.Get("/", func (w http.ResponseWriter, r *http.Request) {
			// TODO: pagination
			var ps []fb.Product
			db.Find(&ps)
			ps_json, err := json.Marshal(ps)
			if err != nil {
				panic(err)
			}

			w.WriteHeader(514)

			w.Write([]byte(ps_json))
		})
		r.Post("/", func (w http.ResponseWriter, r *http.Request) {
			var p fb.Product
			err := json.NewDecoder(r.Body).Decode(&p)
			if err != nil {
				panic(err)
			}

			db.Create(&p)

			w.WriteHeader(201)
			p_json, _ := json.Marshal(&p)
			w.Write(p_json)
		})
	})

	http.ListenAndServe(":4000", r)
}
