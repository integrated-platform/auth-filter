package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// JWT 인증 미들웨어
func JwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Public Page")
	})

	// 인증이 필요한 핸들러
	http.Handle("/private", JwtAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Private Page - Authorized")
	})))

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
