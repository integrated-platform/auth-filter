package auth

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// 사용자 정보 구조체
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 사용자 검증 함수 (API 서버로 요청)
func isValidUser(username, password string) bool {
	// API 서버 URL (예시)
	apiURL := "http://api.example.com/validate-user"

	// 요청 데이터 준비
	userData := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(userData)

	// API 요청
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error sending request to API:", err)
		return false
	}
	defer resp.Body.Close()

	// 응답 상태코드 체크
	return resp.StatusCode == http.StatusOK
}

// 로그인 핸들러
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 사용자 인증 로직
	if isValidUser(user.Username, user.Password) {
		tokenString, err := GenerateJWT(user.Username)
		if err != nil {
			http.Error(w, "Could not generate token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

// JWT 인증 미들웨어
func JwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		// 서버에 토큰 검증 요청
		resp, err := http.Post("http://api-server-url/validate-token", "application/json", bytes.NewBuffer([]byte(`{"token":"`+tokenString+`"}`)))
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// 토큰이 유효하면 다음 핸들러로 진행
		next.ServeHTTP(w, r)
	})
}
