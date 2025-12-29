package onboarding

import "github.com/go-playground/validator/v10"

type OnboardingRequest struct {
	Business struct {
		Name    string `json:"name" validate:"required,min=2"`
		Type    string `json:"type" validate:"required"`
		Address string `json:"address" validate:"required"`
		City    string `json:"city" validate:"required"`
	} `json:"business" validate:"required"`

	User struct {
		FirstName string `json:"first_name" validate:"required,min=2"`
		LastName  string `json:"last_name" validate:"required,min=2"`
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,len=4|len=6"`
	} `json:"user" validate:"required"`
}

var validate = validator.New()

func (r *OnboardingRequest) Validate() error {
	return validate.Struct(r)
}
