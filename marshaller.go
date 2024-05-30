package env

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

const (
	envTag         = "env"
	sliceSeparator = ","
)

type sliceMarshaller interface {
	MarshalEnv(prefix string) []EnvVal
}

func MarshalEnv(in any) []EnvVal {
	return marshal("", reflect.ValueOf(in))
}

func MarshalEnvWithPrefix(prefix string, in any) []EnvVal {
	return marshal(prefix, reflect.ValueOf(in))
}

func marshal(prefix string, ref reflect.Value) []EnvVal {
	prefix = strings.ToUpper(prefix)

	res := make([]EnvVal, 0)
	switch ref.Kind() {
	case reflect.Slice:
		res = append(res, marshalSlice(prefix, ref)...)
	case reflect.Struct:
		res = append(res, marshalStruct(prefix, ref)...)
	case reflect.Ptr:
		if ref.IsNil() {
			return nil
		}

		res = append(res, marshalStruct(prefix, ref.Elem())...)
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		res = append(res, EnvVal{
			Name:  prefix,
			Value: ref.Interface(),
		})
	default:
		return nil
	}

	return res
}

func marshalSlice(prefix string, ref reflect.Value) []EnvVal {
	if ref.Len() == 0 {
		return nil
	}

	tp := ref.Index(0).Kind()

	var marshaller func(prefix string, ref reflect.Value) []EnvVal
	switch {
	case tp == reflect.Struct:

	case tp == reflect.Interface:
		var val any
		if ref.CanAddr() {
			val = ref.Addr().Interface()
		} else {
			val = ref.Interface()
		}
		customMarshaller, ok := val.(sliceMarshaller)
		if !ok {
			panic("Slices of non basic type require sliceMarshaller to be implemented")
		}

		marshaller = func(prefix string, ref reflect.Value) []EnvVal {
			return customMarshaller.MarshalEnv(prefix)
		}

	case tp < reflect.Complex64:
		marshaller = marshallSliceOfBasicType
	default:
		panic("unsupported type " + tp.String())
	}

	return marshaller(prefix, ref)
}

func marshallSliceOfBasicType(prefix string, ref reflect.Value) []EnvVal {
	out := make([]EnvVal, 1)
	outStr := make([]string, 0, ref.Len())
	for i := 0; i < ref.Len(); i++ {
		elem := fmt.Sprint(ref.Index(i).Interface())
		outStr = append(outStr, elem)
	}

	out[0].Name = prefix
	out[0].Value = strings.Join(outStr, sliceSeparator)
	return out
}

func marshalStruct(prefix string, ref reflect.Value) []EnvVal {
	if prefix != "" {
		prefix += "_"
	}
	res := make([]EnvVal, 0, ref.NumField())

	for i := 0; i < ref.NumField(); i++ {
		tag := ref.Type().Field(i).Tag.Get(envTag)
		if tag == "-" {
			continue
		}

		if tag == "" {
			tag = splitToSnake(ref.Type().Field(i).Name)
		}
		tag = prefix + tag
		value := ref.Field(i)
		res = append(res, marshal(tag, value)...)
	}

	return res
}

func splitToSnake(in string) string {
	inR := []rune(in)
	out := make([]rune, 0, len(inR)+2)
	for idx, r := range inR {
		if unicode.IsUpper(r) && idx != 0 {
			out = append(out, '_')
		}

		out = append(out, r)
	}

	return string(out)
}
