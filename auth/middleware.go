package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	utils "auth-filter/utility" // utils 패키지 import
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

// 사용자 정보 구조체
type User struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
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
	if err != nil {
		http.Error(w, "API 서버에 요청을 보내는 데 실패했습니다", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// API 응답 처리
	var apiResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		http.Error(w, "응답 데이터를 처리하는 중 오류가 발생했습니다", http.StatusInternalServerError)
		return
	}

	// Spring API 응답에서 성공 여부와 메시지를 가져옴
	success, _ := apiResponse["success"].(bool)
	message, _ := apiResponse["message"].(string)

	// 클라이언트에 성공 여부와 메시지 전달
	w.Header().Set("Content-Type", "application/json")
	if !success {
		// 실패한 경우, 예: 이메일 중복
		w.WriteHeader(http.StatusUnprocessableEntity) // 422
		json.NewEncoder(w).Encode(map[string]string{
			"message": message, // 예: "이미 존재하는 이메일입니다."
		})
	} else {
		// 성공한 경우
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(map[string]string{
			"message": "사용자 등록이 완료되었습니다.", // 성공 메시지
		})
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		// 잘못된 요청일 경우 JSON으로 응답
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "잘못된 요청입니다",
		})
		return
	}

	// .env 파일 로드
	if err := utils.LoadEnv(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "서버 환경 변수를 로드하지 못했습니다.",
		})
		return
	}

	// API 서버 URL
	apiURL := utils.GetAPIURL("/users/login")

	// 요청 데이터 준비
	loginRequest := map[string]string{
		"email":    user.Email,
		"password": user.Password,
	}
	jsonData, err := json.Marshal(loginRequest)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "요청 데이터를 처리하는 중 오류가 발생했습니다.",
		})
		return
	}

	// API 요청
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "서버에 요청을 보내는 중 오류가 발생했습니다.",
		})
		return
	}
	defer resp.Body.Close()

	// 응답 처리
	if resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "로그인 실패: 잘못된 자격 증명입니다.",
		})
		return
	}

	// JWT 생성
	tokenString, err := GenerateJWT(user.Email, user.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "토큰 생성에 실패했습니다.",
		})
		return
	}

	// 성공 응답
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

// ForwardRequest 함수는 요청을 API 서버로 전달합니다.
func ForwardRequest(w http.ResponseWriter, r *http.Request) {
	// API 서버 URL
	apiURL := utils.GetAPIURL(r.URL.Path)

	// 요청 데이터 복사
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "잘못된 요청입니다: 요청 본문을 읽을 수 없습니다", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()                              // 요청 본문을 닫음
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Body를 다시 읽을 수 있도록 설정

	// 새 요청 생성
	req, err := http.NewRequest(r.Method, apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		http.Error(w, "요청 생성 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 헤더 복사
	req.Header = r.Header

	// API 서버로 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "API 서버로 요청을 전송하는 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// API 응답 처리
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) // API 응답을 클라이언트로 전달
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
