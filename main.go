package main

import (
	_ "github.com/joho/godotenv/autoload"

	"fmt"
	"os"
	"net/http"
	"encoding/json"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"

	fb "fancybasket/db"
)

type LoginAttempt struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func CreateToken(user_id uint) (string, error) {
	at_claims := jwt.MapClaims{}
	at_claims["authorized"] = true
	at_claims["user_id"] = user_id
	at_claims["exp"] = time.Now().Add(time.Minute * 60).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, at_claims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	return token, err
}

func main() {
	db := fb.GetDB();
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/tokens", func (r chi.Router) {
		r.Post("/", func (w http.ResponseWriter, r *http.Request) {
			var l LoginAttempt
			json.NewDecoder(r.Body).Decode(&l)

			var user fb.User
			db.Find(&user, "email = ?", l.Email)
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(l.Password))

			if err != nil {
				fmt.Println(err)
				w.WriteHeader(403)
				w.Write([]byte("get outta here"))
				return
			}

			token, err := CreateToken(user.ID)
			w.Write([]byte(token))
		})
	})

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
