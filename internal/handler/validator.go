package handler

import (
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
)

func bindAndValidate(c *fiber.Ctx, out interface{}) error {
	if err := c.BodyParser(out); err != nil {
		return apperrors.New(apperrors.ErrBadRequest.Code, "invalid JSON body", apperrors.ErrBadRequest.Status)
	}
	if err := validateStruct(out); err != nil {
		return err
	}
	return nil
}

func validateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return validateValue(v, "")
}

func validateValue(value reflect.Value, name string) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		t := value.Type()
		for i := 0; i < value.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			fieldValue := value.Field(i)
			fieldName := jsonFieldName(field)
			tag := field.Tag.Get("validate")
			if tag == "" {
				if err := validateValue(fieldValue, fieldName); err != nil {
					return err
				}
				continue
			}
			if err := applyRules(tag, fieldName, fieldValue); err != nil {
				return err
			}
		}
	default:
		return nil
	}
	return nil
}

func applyRules(tag, name string, value reflect.Value) error {
	rules := strings.Split(tag, ",")
	for _, rule := range rules {
		if rule == "dive" {
			if value.Kind() != reflect.Slice {
				continue
			}
			for i := 0; i < value.Len(); i++ {
				item := value.Index(i)
				itemName := name
				if name != "" {
					itemName = name + "[" + itoa(i) + "]"
				}
				if err := validateValue(item, itemName); err != nil {
					return err
				}
			}
			continue
		}
		if err := applyRule(rule, name, value); err != nil {
			return err
		}
	}
	return nil
}

func jsonFieldName(field reflect.StructField) string {
	name := field.Tag.Get("json")
	if idx := strings.Index(name, ","); idx >= 0 {
		name = name[:idx]
	}
	if name == "" || name == "-" {
		return field.Name
	}
	return name
}

func applyRule(rule, name string, value reflect.Value) error {
	switch {
	case rule == "required":
		if isEmpty(value) {
			return apperrors.New(apperrors.ErrValidation.Code, name+" is required", apperrors.ErrValidation.Status)
		}
	case rule == "email":
		if value.Kind() == reflect.String && value.String() != "" && !strings.Contains(value.String(), "@") {
			return apperrors.New(apperrors.ErrValidation.Code, name+" must be a valid email", apperrors.ErrValidation.Status)
		}
	case strings.HasPrefix(rule, "min="):
		min := parseIntSuffix(rule, "min=")
		if value.Kind() == reflect.String && len(value.String()) < min {
			return apperrors.New(apperrors.ErrValidation.Code, name+" must be at least "+itoa(min)+" characters", apperrors.ErrValidation.Status)
		}
		if value.Kind() == reflect.Slice && value.Len() < min {
			return apperrors.New(apperrors.ErrValidation.Code, name+" must contain at least "+itoa(min)+" items", apperrors.ErrValidation.Status)
		}
		if (value.Kind() == reflect.Int || value.Kind() == reflect.Int64) && value.Int() < int64(min) {
			return apperrors.New(apperrors.ErrValidation.Code, name+" must be at least "+itoa(min), apperrors.ErrValidation.Status)
		}
	}
	return nil
}

func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Int, reflect.Int64:
		return v.Int() == 0
	case reflect.Slice:
		return v.Len() == 0
	default:
		return v.IsZero()
	}
}

func parseIntSuffix(rule, prefix string) int {
	n := 0
	for _, ch := range strings.TrimPrefix(rule, prefix) {
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int(ch-'0')
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
