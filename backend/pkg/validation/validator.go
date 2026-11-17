package validation

import (
	"log/slog"
	"reflect"
	"strings"

	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	idtranslations "github.com/go-playground/validator/v10/translations/id"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var validate = newValidator()

func Struct(value any) []FieldError {
	if err := validate.Struct(value); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			slog.Error("struct validation failed", "error", err)
			return []FieldError{{Message: "Data tidak valid."}}
		}
		result := make([]FieldError, 0, len(validationErrors))
		for _, fieldErr := range validationErrors {
			result = append(result, FieldError{
				Field:   fieldErr.Field(),
				Message: fieldErr.Translate(translator),
			})
		}
		return result
	}
	return nil
}

var translator ut.Translator

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			return field.Name
		}
		return name
	})
	universalTranslator := ut.New(id.New(), id.New())
	var found bool
	translator, found = universalTranslator.GetTranslator("id")
	if !found {
		panic("Indonesian validation translator unavailable")
	}
	if err := idtranslations.RegisterDefaultTranslations(v, translator); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("notblank", func(field validator.FieldLevel) bool {
		return strings.TrimSpace(field.Field().String()) != ""
	}); err != nil {
		panic(err)
	}
	if err := v.RegisterTranslation("notblank", translator, func(translator ut.Translator) error {
		return translator.Add("notblank", "{0} tidak boleh kosong.", false)
	}, func(translator ut.Translator, fieldErr validator.FieldError) string {
		message, _ := translator.T("notblank", fieldErr.Field())
		return message
	}); err != nil {
		panic(err)
	}
	return v
}
