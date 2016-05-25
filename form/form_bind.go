package form

import (
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func inStringSlice(ele string, list []string) bool {
	for _, field := range list {
		if ele == field {
			return true
		}
	}
	return false
}

func Bind(kwargs url.Values, form interface{}, fields ...string) {
	ptrFormValue := reflect.ValueOf(form)
	if ptrFormValue.Kind() != reflect.Ptr {
		panic("you should bind to a pointer to a struct")
	}

	formValue := ptrFormValue.Elem()
	if formValue.Kind() != reflect.Struct {
		panic("you should bind to a pointer to a struct")
	}

	formType := formValue.Type()
	for i := 0; i < formType.NumField(); i += 1 {
		fieldValue := formValue.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		field := formType.Field(i)

		if len(fields) != 0 && !inStringSlice(field.Name, fields) {
			continue
		}

		form_name := field.Tag.Get("form")

		if form_name == "-" {
			continue
		}

		if form_name == "" {
			// 将驼峰转化成下划线
			re := regexp.MustCompile("[A-Z]")
			form_name = re.ReplaceAllString(field.Name, "_$0")
			form_name = strings.TrimPrefix(form_name, "_")
			form_name = strings.ToLower(form_name)
		}

		value := kwargs.Get(form_name)
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				fieldValue.SetInt(intValue)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
			if uintValue, err := strconv.ParseUint(value, 10, 64); err == nil {
				fieldValue.SetUint(uintValue)
			}
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(value); err == nil {
				fieldValue.SetBool(boolValue)
			}
		case reflect.Float32, reflect.Float64:
			if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
				fieldValue.SetFloat(floatValue)
			}

		}

	}

}
