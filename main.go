package main

import (
	"auth-filter/auth" // 패키지 경로
	"fmt"
	"net/http"
	"time"

	"log" // 로그 패키지 추가

	"github.com/gorilla/handlers" // CORS 패키지 추가
)

// CustomResponseWriter 구조체
type CustomResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader 메서드 오버라이드
func (crw *CustomResponseWriter) WriteHeader(code int) {
	crw.StatusCode = code
	crw.ResponseWriter.WriteHeader(code)
}

// 로그 미들웨어
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // 요청 시작 시간 기록

		crw := &CustomResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK} // CustomResponseWriter 사용

		// 요청 처리
		next.ServeHTTP(crw, r)

		// 로그 출력
		log.Printf(
			"[%s] %s %s | Status: %d | Duration: %v | Headers: %v",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			crw.StatusCode, // 상태 코드
			time.Since(start),
			r.Header, // 요청 헤더
		)
	})
}

func main() {
	// 핸들러 설정
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/register", auth.RegisterHandler)

	// 기본 핸들러로 모든 요청 처리
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 로그인 요청이 아니면 JWT 인증 미들웨어 적용
		if r.URL.Path != "/login" {
			// JWT 인증 미들웨어 적용
			auth.JwtAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 여기서 URL에 따라 다른 처리 가능
				fmt.Fprintf(w, "Welcome to %s", r.URL.Path)
			})).ServeHTTP(w, r)
		} else {
			// 로그인 요청 처리
			auth.LoginHandler(w, r)
		}
	})

	// CORS 설정
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // 모든 출처 허용
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(http.DefaultServeMux)

	// 로그 미들웨어 추가
	loggedRouter := loggingMiddleware(corsHandler)

	fmt.Println("Server is running on port 5000")
	http.ListenAndServe(":5000", loggedRouter) // 로그 미들웨어와 CORS 핸들러 적용
}
