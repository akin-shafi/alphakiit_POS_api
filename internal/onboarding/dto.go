package onboarding

import "github.com/go-playground/validator/v10"

type OnboardingRequest struct {
	Business struct {
		Name     string `json:"name" validate:"required,min=2"`
		Type     string `json:"type" validate:"required"`
		Address  string `json:"address" validate:"required"`
		City     string `json:"city" validate:"required"`
		Currency string `json:"currency" validate:"required"`
	} `json:"business" validate:"required"`

	User struct {
		FirstName string `json:"first_name" validate:"required,min=2"`
		LastName  string `json:"last_name" validate:"required,min=2"`
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=4"`
	} `json:"user" validate:"required"`

	ReferralToken string   `json:"referral_token,omitempty"`
	BasePlanType  string   `json:"base_plan_type,omitempty"` // TRIAL, SERVICE_MONTHLY (as trial) etc
	BundleCode    string   `json:"bundle_code,omitempty"`
	Modules       []string `json:"modules,omitempty"` // Selected optional modules
	UseSampleData bool     `json:"use_sample_data"`   // Whether to seed sample data
	SkipTrial     bool     `json:"skip_trial"`        // To skip 14-day trial and pay now
}

var validate = validator.New()

func (r *OnboardingRequest) Validate() error {
	return validate.Struct(r)
}
