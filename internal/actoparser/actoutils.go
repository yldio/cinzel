package actoparser

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func GetHclTag(i any, name string) (string, error) {
	st := reflect.TypeOf(i)
	field, ok := st.FieldByName(name)
	if !ok {
		return "", errors.New("")
	}

	attr := field.Tag.Get("hcl")

	if attr == "" {
		return "", errors.New("")
	}

	return attr, nil
}
