package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"os"
	"time"
)

func GenerateOTP() (string, error) {
	log.Println("otp")
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("OTP generation failed: %w", err)

	}

	return fmt.Sprintf("%06d", n.Int64()), err
}




var otpStore = make(map[string]string)

// SendOTPEmail sends OTP to the user's email
func SendOTPEmail(email string, otp string) error {
	log.Println("sending email")
	subject := "Your OTP for Signup Verification"
	body := fmt.Sprintf("Your OTP is %s. It will expire in 5 minutes.", otp)
	return SendEmail(email, subject, body)
}

func StoreOTP(email string, otp string, expiration time.Duration) {
	otpStore[email] = otp
	go func() {
		time.Sleep(expiration)
		delete(otpStore, email) // OTP expires after `expiration`
	}()
}

// VerifyOTP checks if the provided OTP matches the stored one
func VerifyOTP(email string, providedOTP string) bool {
	storedOTP, exists := otpStore[email]
	if !exists || storedOTP != providedOTP {
		return false
	}
	delete(otpStore, email) // Delete OTP after successful verification
	return true
}

// SendEmail sends an email with the given subject and body
func SendEmail(to string, subject string, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")
	log.Println(`sendmail`)	

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{to}, msg)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}
	return nil
}
