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
	"strings"

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

func CreateToken(user_id string) (string, error) {
	at_claims := jwt.MapClaims{}
	at_claims["authorized"] = true
	at_claims["user_id"] = user_id
	at_claims["exp"] = time.Now().Add(time.Minute * 60).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, at_claims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	return token, err
}

func GetUserIdFromRequest(r *http.Request) string {
	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		return ""
	}

	token, _ := jwt.Parse(tokenString, func (token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, _ := token.Claims.(jwt.MapClaims)
	return claims["user_id"].(string)
}

func GetUserFromRequest(r *http.Request, db *gorm.DB) (*fb.User, error) {
	user_id := GetUserIdFromRequest(r)
	if user_id == "" {
		return nil, nil
	}

	var u fb.User
	db.Find(&u, "id = ?", user_id)
	return &u, db.Error
}

func AllowRoles (db *gorm.DB, roles ...fb.Role) func (http.Handler) http.Handler {
	return func (next http.Handler) http.Handler {
		return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			u, _ := GetUserFromRequest(r, db)

			if u == nil {
				w.WriteHeader(403)
				return
			}

			has_role := false
			for _, v := range roles {
				if v == u.Role {
					has_role = true
					break
				}
			}

			if !has_role {
				w.WriteHeader(403)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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

		r.Route("/{user_id}/orders", func (r2 chi.Router) {
			r2.Post("/", func (w http.ResponseWriter, r *http.Request) {
				type PlaceOrder struct {
					Qty int
					ProductID string
				}

				jwt_user_id := GetUserIdFromRequest(r)
				param_user_id := chi.URLParam(r, "user_id")

				if jwt_user_id != param_user_id {
					w.WriteHeader(403)
					return
				}

				var po PlaceOrder
				json.NewDecoder(r.Body).Decode(&po)
				if po.Qty <= 0 || po.ProductID == "" {
					http.Error(w, "Bad order placement", 400)
					return
				}

				var product fb.Product
				res := db.First(&product, "id = (?)", po.ProductID)
				fmt.Println(product)

				if res.Error != nil {
					http.Error(w, "Product with id " + po.ProductID + " not found", 400)
					return
				}

				if product.Stock < uint64(po.Qty) {
					http.Error(w, "Insufficient stock", 400)
					return
				}

				result := db.Create(&fb.Order{
					ProductID: po.ProductID,
					UserID: jwt_user_id,
					Qty: po.Qty,
				})
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(&result)
			})

			r2.Get("/", func (w http.ResponseWriter, r *http.Request) {
				jwt_user_id := GetUserIdFromRequest(r)
				param_user_id := chi.URLParam(r, "user_id")

				if jwt_user_id != param_user_id {
					w.WriteHeader(403)
					return
				}

				var orders []fb.Order
				db.Where("user_id IN (?)", param_user_id).Find(&orders)

				json.NewEncoder(w).Encode(&orders)
			})
		})
	})

	r.Route("/products", func (r chi.Router) {
		r.With(Pagination).Get("/", func (w http.ResponseWriter, r *http.Request) {
			page := r.Context().Value("page").(int64)
			size := r.Context().Value("size").(int64)
			category := r.URL.Query()["category_ids"]
			var ps []fb.Product
			q := db.Offset(page * size).Limit(size).Order("created_at desc")
			if len(category) > 0 {
				q = q.Where("category_id IN (?)", strings.Split(category[0], ","))
			}
			q.Find(&ps)
			json.NewEncoder(w).Encode(&ps)
		})

		r.With(AllowRoles(db, fb.Admin)).Post("/", func (w http.ResponseWriter, r *http.Request) {
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
