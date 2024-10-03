package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthRequest struct {
	Email    string `json:"Email"`
	Password string `json:"password"`
}

// JWT 생성 함수
func GenerateJWT(Email string, password string) (string, error) {
	// API 서버에 인증 요청
	authRequest := AuthRequest{Email: Email, Password: password}
	requestBody, err := json.Marshal(authRequest)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed: %s", resp.Status)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["token"]
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	return token, nil
}
