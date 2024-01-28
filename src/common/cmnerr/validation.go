package cmnerr

import (
	"errors"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

func GetValidationErrors(err error) []string {
	errorItems := validator.ValidationErrors{}
	if ok := errors.As(err, &errorItems); !ok {
		return []string{FailedToCastReason}
	}

	var resErrors []string
	for _, field := range errorItems {
		errTxt := field.Error()
		if i := strings.Index(errTxt, "validation for"); i != -1 {
			errTxt = errTxt[i:]
		}
		resErrors = append(resErrors, "Field "+field.Field()+": "+errTxt)
	}

	return resErrors
}

const FailedToCastReason = "Failed to cast validation errors"
