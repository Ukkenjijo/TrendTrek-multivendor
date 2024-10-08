package utils

import (
	
	"regexp"

	"github.com/go-playground/validator/v10"
)

func ValidateStruct(data interface{}) error {
	validate := validator.New()
	validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		return len(password) >= 8 &&
			regexp.MustCompile(`[a-zA-Z]`).MatchString(password) &&
			regexp.MustCompile(`[0-9]`).MatchString(password)
	})
	err := validate.Struct(data)
	

	if err != nil {
		
		
		return err
	}

	return nil
}
