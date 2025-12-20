package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validatorInstance = newValidator()

func newValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	_ = validate.RegisterValidation("notblank", notBlank)
	return validate
}

func notBlank(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return strings.TrimSpace(field.String()) != ""
	case reflect.Ptr:
		if field.IsNil() {
			return true
		}
		elem := field.Elem()
		if elem.Kind() == reflect.String {
			return strings.TrimSpace(elem.String()) != ""
		}
	}

	return true
}

func ParseAndValidate(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	return ValidateStruct(dst)
}

func ValidateStruct(value any) error {
	if err := validatorInstance.Struct(value); err != nil {
		return mapValidationError(err)
	}
	return nil
}

func RequireParam(value, label string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if err := validatorInstance.Var(trimmed, "required,notblank"); err != nil {
		return "", fiber.NewError(fiber.StatusBadRequest, label+" is required")
	}
	return trimmed, nil
}

func mapValidationError(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(*validator.InvalidValidationError); ok {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	fieldError := validationErrors[0]
	field := fieldError.Field()

	switch fieldError.Tag() {
	case "required":
		return fiber.NewError(fiber.StatusBadRequest, field+" is required")
	case "notblank":
		return fiber.NewError(fiber.StatusBadRequest, field+" cannot be empty")
	case "email":
		return fiber.NewError(fiber.StatusBadRequest, field+" must be a valid email")
	case "min":
		return fiber.NewError(fiber.StatusBadRequest, field+" is too short")
	case "required_without_all":
		return fiber.NewError(fiber.StatusBadRequest, "no fields to update")
	default:
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}
}
