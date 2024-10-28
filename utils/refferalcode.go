package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

func GenerateReferralCode() string {
	//Generate a six digit refferal code
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(bytes)
}
