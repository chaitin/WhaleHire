package validate

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator.New()}
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.applyDefaults(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "应用默认值失败: "+err.Error())
	}

	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (cv *CustomValidator) applyDefaults(i any) error {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return cv.applyDefaultsRecursive(v)
}

func (cv *CustomValidator) applyDefaultsRecursive(v reflect.Value) error {
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct {
			if err := cv.applyDefaultsRecursive(field); err != nil {
				return err
			}
			continue
		}

		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
			// 检查该结构体或其字段是否有default tag
			hasDefaultTag := false
			if field.IsNil() {
				// 检查结构体类型的字段是否有default tag
				structType := field.Type().Elem()
				for j := 0; j < structType.NumField(); j++ {
					if structType.Field(j).Tag.Get("default") != "" {
						hasDefaultTag = true
						break
					}
				}

				// 只有当有default tag时才创建新实例
				if hasDefaultTag {
					field.Set(reflect.New(field.Type().Elem()))
				} else {
					continue // 没有default tag，保持为nil
				}
			}

			if err := cv.applyDefaultsRecursive(field.Elem()); err != nil {
				return err
			}
			continue
		}

		defaultValue := fieldType.Tag.Get("default")
		if defaultValue == "" {
			continue
		}

		if !cv.isZeroValue(field) {
			continue
		}

		if err := cv.setFieldValue(field, defaultValue); err != nil {
			return err
		}
	}

	return nil
}

func (cv *CustomValidator) isZeroValue(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return field.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return field.Float() == 0
	case reflect.Bool:
		return !field.Bool()
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface:
		return field.IsNil()
	default:
		return false
	}
}

func (cv *CustomValidator) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(val)
		} else {
			return err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(val)
		} else {
			return err
		}
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(val)
		} else {
			return err
		}
	case reflect.Bool:
		if val, err := strconv.ParseBool(value); err == nil {
			field.SetBool(val)
		} else {
			return err
		}
	default:
		return nil
	}
	return nil
}

func (cv *CustomValidator) ValidateWithDefaults(i any) error {
	return cv.Validate(i)
}

func (cv *CustomValidator) SetDefault(i any, fieldName, defaultValue string) error {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return echo.NewHTTPError(http.StatusBadRequest, "字段不存在: "+fieldName)
	}

	if !field.CanSet() {
		return echo.NewHTTPError(http.StatusBadRequest, "字段不可设置: "+fieldName)
	}

	return cv.setFieldValue(field, defaultValue)
}
