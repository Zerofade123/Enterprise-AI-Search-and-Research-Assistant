package validation

import (
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/errors"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return errors.Wrap("validation.ValidateStruct", errors.CodeValidation, "validation failed", err)
	}
	return nil
}
