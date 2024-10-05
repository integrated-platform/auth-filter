package utils

import (
	"fmt"
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
	return fmt.Sprintf("http://%s:%s%s", host, port, endpoint)
}
