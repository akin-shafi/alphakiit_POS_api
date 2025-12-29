package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationErrors(err error) map[string][]string {
	errors := make(map[string][]string)

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		errors["error"] = []string{"Invalid request"}
		return errors
	}

	for _, fieldErr := range validationErrors {
		field := strings.ToLower(fieldErr.Field())

		var message string

		switch fieldErr.Tag() {
		case "required":
			message = "This field is required"
		case "email":
			message = "Must be a valid email address"
		case "min":
			message = fmt.Sprintf("Must be at least %s characters", fieldErr.Param())
		case "len":
			message = fmt.Sprintf("Must be exactly %s characters", fieldErr.Param())
		default:
			message = "Invalid value"
		}

		errors[field] = append(errors[field], message)
	}

	return errors
}
