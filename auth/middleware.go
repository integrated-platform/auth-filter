package auth

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	utils "auth-filter/utility" // utils 패키지 import
)

// 사용자 정보 구조체
type User struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// 사용자 검증 함수 (API 서버로 요청)
func isValidUser(email, password string) bool {
	// .env 파일 로드
	if err := utils.LoadEnv(); err != nil {
		log.Println("환경 변수 로드 실패:", err)
		return false
	}

	// API 서버 URL
	apiURL := utils.GetAPIURL("/users/login")

	// 요청 데이터 준비
	loginRequest := map[string]string{
		"email":    email, // 여기에 이메일 대신 사용자 이름으로 사용
		"password": password,
	}
	jsonData, err := json.Marshal(loginRequest)
	if err != nil {
		log.Println("JSON 변환 실패:", err)
		return false
	}

	// API 요청
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("API 서버로 요청 실패:", err)
		return false
	}
	defer resp.Body.Close()

	// 응답 상태코드 체크
	return resp.StatusCode == http.StatusOK // 로그인 성공 여부
}

// 회원가입 핸들러
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "잘못된 요청입니다", http.StatusBadRequest)
		return
	}

	// .env 파일 로드
	if err := utils.LoadEnv(); err != nil {
		http.Error(w, "환경 변수를 로드할 수 없습니다", http.StatusInternalServerError)
		return
	}

	// API 서버 URL
	apiURL := utils.GetAPIURL("/users")

	// 요청 데이터 준비
	registrationRequest := map[string]string{
		"email":    user.Email,
		"password": user.Password,
		"username": user.Username,
	}
	jsonData, err := json.Marshal(registrationRequest)
	if err != nil {
		http.Error(w, "데이터 처리 중 오류가 발생했습니다", http.StatusInternalServerError)
		return
	}

	// API 서버에 사용자 추가 요청
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "사용자를 추가하는 데 실패했습니다", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 사용자 등록 성공 시 응답
	w.WriteHeader(http.StatusCreated) // 201 Created
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "사용자 등록이 완료되었습니다."})
}

// 로그인 핸들러
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "잘못된 요청입니다", http.StatusBadRequest)
		return
	}

	// 사용자 인증 로직
	if isValidUser(user.Email, user.Password) {
		tokenString, err := GenerateJWT(user.Email, user.Password)
		if err != nil {
			http.Error(w, "토큰 생성에 실패했습니다", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	} else {
		http.Error(w, "잘못된 자격 증명입니다", http.StatusUnauthorized)
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

		// .env 파일 로드
		if err := utils.LoadEnv(); err != nil {
			http.Error(w, "환경 변수를 로드할 수 없습니다", http.StatusInternalServerError)
			return
		}

		// 서버에 토큰 검증 요청
		apiURL := utils.GetAPIURL("/validate-token")
		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer([]byte(`{"token":"`+tokenString+`"}`)))
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, "잘못된 토큰입니다", http.StatusUnauthorized)
			return
		}

		// 토큰이 유효하면 다음 핸들러로 진행
		next.ServeHTTP(w, r)
	})
}
