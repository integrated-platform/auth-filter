package main

import (
	"auth-filter/auth" // 패키지 경로
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/register", auth.RegisterHandler)
	// 기본 핸들러로 모든 요청 처리
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 로그인 요청이 아니면 JWT 인증 미들웨어 적용
		if r.URL.Path != "/login" {
			auth.JwtAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 여기서 URL에 따라 다른 처리 가능
				fmt.Fprintf(w, "Welcome to %s", r.URL.Path)
			})).ServeHTTP(w, r)
		} else {
			// 로그인 요청 처리
			auth.LoginHandler(w, r)
		}
	})
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":5000", nil)
}
