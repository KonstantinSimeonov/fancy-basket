package main

import (
	_ "github.com/joho/godotenv/autoload"

	"fmt"
	"os"
	"net/http"
	"encoding/json"
	"time"
	"strconv"
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"github.com/jinzhu/gorm"

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

func GetUserFromRequest(r *http.Request, db *gorm.DB) (*fb.User, error) {
	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		return nil, nil
	}

	token, err := jwt.Parse(tokenString, func (token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, _ := token.Claims.(jwt.MapClaims)
	var u fb.User
	db.Find(&u, "id = ?", claims["user_id"])
	fmt.Println(claims["user_id"])
	return &u, err
}

func Clamp (low, high, val int64) int64 {
	if val < low {
		return low
	}

	if val > high {
		return high
	}

	return val
}

func Pagination (next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		page, _ := strconv.ParseInt(query.Get("page"), 10, 64)
		if page < 0 {
			page = 0
		}
		size, _ := strconv.ParseInt(query.Get("size"), 10, 64)
		size = Clamp(20, 100, size)
		if size <= 0 {
			size = 50
		}

		ctx := context.WithValue(r.Context(), "page", page)
		ctx = context.WithValue(ctx, "size", size)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
		r.With(Pagination).Get("/", func (w http.ResponseWriter, r *http.Request) {
			// TODO: pagination
			page := r.Context().Value("page").(int64)
			size := r.Context().Value("size").(int64)
			var ps []fb.Product
			db.Offset(page * size).Limit(size).Order("created_at desc").Find(&ps)
			json.NewEncoder(w).Encode(&ps)
		})

		r.Post("/", func (w http.ResponseWriter, r *http.Request) {
			var p fb.Product

			u, _ := GetUserFromRequest(r, db)

			if u.Role != fb.Admin {
				w.WriteHeader(403)
				return
			}
			
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
