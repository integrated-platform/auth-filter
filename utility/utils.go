package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv .env 파일을 로드하는 함수
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf(".env 파일을 로드할 수 없습니다: %v", err)
	}
	return nil
}

// GetAPIURL API 서버 주소를 생성하는 함수
func GetAPIURL(endpoint string) string {
	host := os.Getenv("AUTH_SERVER_HOST")
	port := os.Getenv("AUTH_SERVER_PORT")
	apiURL := fmt.Sprintf("%s:%s%s", host, port, endpoint)

	// 로그 출력
	log.Printf("API URL: %s", apiURL)
	log.Printf("Using AUTH_SERVER_HOST: %s", host)
	log.Printf("Using AUTH_SERVER_PORT: %s", port)

	return apiURL
}
