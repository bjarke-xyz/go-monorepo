package validation

// TODO: fix bad package name

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func NewValidator() *validator.Validate {
	validate := validator.New()

	_ = validate.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		field := fl.Field().String()
		if _, err := uuid.Parse(field); err != nil {
			return true
		}
		return false
	})
	return validate
}

// ValidatorErrors func for show validation errors for each invalid fields.
func ValidatorErrors(err error) map[string]string {
	// Define fields map.
	fields := map[string]string{}

	// Make error message for each invalid field.
	// for _, err := range err.(validator.ValidationErrors) {
	// 	fields[err.Field()] = err.???
	// }
	// TODO: couldnt make this work for some reason
	// https://github.com/koddr/tutorial-go-fiber-rest-api/blob/05be4c0db3/pkg/utils/validator.go

	return fields
}
