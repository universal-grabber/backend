package helper

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"unicode"
)

// MapIt maps get request into struct
func ParseRequestQuery(req *http.Request, d interface{}) error {
	dType := reflect.TypeOf(d)

	if err := shouldBeStruct(dType); err != nil {
		return err
	}

	// Data Holder Value
	dhVal := reflect.ValueOf(d)

	// Loop over all the fields present in struct (Title, Body, JSON)
	for i := 0; i < dType.Elem().NumField(); i++ {

		// Give me ith field. Elem() is used to dereference the pointer
		field := dType.Elem().Field(i)

		// Get the value from field tag i.e in case of Title it is "title"
		key := field.Tag.Get("mapper")

		if len(key) == 0 {
			key = lcFirst(field.Name)
		}

		// Get the type of field
		kind := field.Type.Kind()

		// Get the value from query params with given key
		val := req.URL.Query().Get(key)

		if len(val) == 0 {
			continue
		}

		//  Get reference of field value provided to input `d`
		result := dhVal.Elem().Field(i)

		// we only check for string for now so,
		if kind == reflect.String {
			// set the value to string field
			// for other kinds we need to use different functions like; SetInt, Set etc
			result.SetString(val)
		} else if kind == reflect.Int {
			val, err := strconv.Atoi(val)

			if err != nil {
				return err
			}

			result.SetInt(int64(val))
		} else if kind == reflect.Bool {

			result.SetBool(val == "true")
		} else {
			return errors.New("only supports string")
		}

	}
	return nil
}

func shouldBeStruct(d reflect.Type) error {
	td := d.Elem()
	if td.Kind() != reflect.Struct {
		errStr := fmt.Sprintf("Input should be %v, found %v", reflect.Struct, td.Kind())
		return errors.New(errStr)
	}
	return nil
}

func lcFirst(str string) string {
	for _, v := range str {
		u := string(unicode.ToLower(v))
		return u + str[len(u):]
	}
	return ""
}
