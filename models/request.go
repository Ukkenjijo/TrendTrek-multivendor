package models

type EmailSignupRequest struct {
	Name            string `validate:"required" json:"name"`
	Email           string `validate:"required,email" json:"email"`
	PhoneNumber     string `validate:"required" json:"phone_number"`
	Password        string `validate:"required,password" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}

type StoreSignupRequest struct {
	// User (Seller) Fields
	Name        string `json:"name" validate:"required"`              // Seller's name
	Email       string `json:"email" validate:"required,email"`       // Seller's email
	PhoneNumber string `json:"phone_number" validate:"required"`      // Seller's phone number
	Password    string `json:"password" validate:"required,password"` // Seller's password

	// Store Fields
	StoreName   string `json:"store_name" validate:"required"`  // Store name
	Description string `json:"description" validate:"required"` // Store description
	Address     string `json:"address" validate:"required"`     // Store address
	City        string `json:"city" validate:"required"`        // Store city
	State       string `json:"state" validate:"required"`       // Store state
	Country     string `json:"country" validate:"required"`     // Store country
}

type EmailLoginRequest struct {
	Email    string `form:"email" validate:"required,email" json:"email"`
	Password string `form:"password" validate:"required" json:"password"`
}
type ResetPasswordRequest struct {
	Email    string `json:"email" validate:"required,email"`
	OTP      string `json:"otp"`
	Password string `json:"password" validate:"required,password"`
}

type OrderRequest struct {
	AddressID uint `json:"address_id" validate:"required"`
	PaymentMode string `json:"payment_mode" validate:"required"`
}

type StatusRequest struct {
	Status string `json:"status" validate:"required"`
}