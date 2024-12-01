package helpers

import (
	"crypto/rand"
	"log"
	"math/big"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return
	}
}

func generateRandomID(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomByte, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[randomByte.Int64()]
	}
	return string(b), nil
}

func GenerateRoomID() (string, error) {
	blocks := 3
	blockSize := 4
	var roomID string
	for i := 0; i < blocks; i++ {
		block, err := generateRandomID(blockSize)
		if err != nil {
			return "", err
		}
		if i > 0 {
			roomID += "-"
		}
		roomID += block
	}
	return roomID, nil
}
