package envtag

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(prefix string, out interface{}) (err error) {
	return unmarshal(prefix, reflect.ValueOf(out).Elem())
}

func unmarshal(prefix string, rval reflect.Value) error {
	for i := 0; i < rval.NumField(); i++ {
		tag := strings.Split(rval.Type().Field(i).Tag.Get("ENV"), ",")

		omitEmpty := false
		if len(tag) > 1 {
			omitEmpty = tag[1] == "omitempty"
		}
		val := os.Getenv(fmt.Sprintf("%s%s", prefix, tag[0]))

		if rval.Field(i).Kind() == reflect.Struct {
			err := unmarshal(fmt.Sprintf("%s%s", prefix, tag[0]), rval.Field(i))
			if err != nil {
				return err
			}
			continue
		}

		if tag[0] == "" {
			continue
		}
		if val == "" && omitEmpty {
			continue
		}

		switch rval.Field(i).Kind() {
		case reflect.String:
			rval.Field(i).SetString(val)

		case reflect.Bool:
			if val == "" {
				val = "false"
			}
			bval, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			rval.Field(i).SetBool(bval)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if val == "" {
				val = "0"
			}
			ival, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			rval.Field(i).SetInt(ival)

		case reflect.Float32, reflect.Float64:
			if val == "" {
				val = "0"
			}
			fval, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return err
			}
			rval.Field(i).SetFloat(fval)

		case reflect.Slice, reflect.Array:
			vals := strings.Split(val, ",")
			typ := reflect.TypeOf(rval.Field(i).Interface()).Elem()
			nslice := reflect.MakeSlice(reflect.SliceOf(typ), len(vals), len(vals))
			for i, v := range vals {
				switch typ.Kind() {
				case reflect.String:
					nslice.Index(i).SetString(v)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					ival, err := strconv.ParseInt(v, 10, 64)
					if err != nil {
						return err
					}
					nslice.Index(i).SetInt(ival)
				case reflect.Bool:
					bval, err := strconv.ParseBool(v)
					if err != nil {
						return err
					}
					nslice.Index(i).SetBool(bval)
				case reflect.Float32, reflect.Float64:
					fval, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return err
					}
					nslice.Index(i).SetFloat(fval)
				}
			}

			rval.Field(i).Set(nslice)

		default:
			return fmt.Errorf("cannot parse ENV value into %s", rval.Field(i).Kind())
		}
	}

	return nil
}
