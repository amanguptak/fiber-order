package helpers

import "github.com/go-playground/validator/v10"

func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		errors["error"] = "Invalid input"
		return errors
	}
	for _, fieldError := range validationErrors {
		field := fieldError.Field() // this name should be same as it is can not change
		tag := fieldError.Tag()

		switch tag {
		case "required":
			errors[field] = field + " is required"
		case "min":
			errors[field] = field + " must be at least " + fieldError.Param() + " characters"
		case "max":
			errors[field] = field + " must be at most " + fieldError.Param() + " characters"
		case "email":
			errors[field] = field + " must be a valid email"
		default:
			errors[field] = field + " is invalid"
		}
	}

	return errors
}
