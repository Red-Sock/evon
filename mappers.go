package evon

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func extractString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int:
		return strconv.FormatInt(v.Int(), 10)
	default:
		if !v.IsValid() {
			return ""
		}
		return fmt.Sprint(v.Interface())
	}
}
func mapString(target reflect.Value) NodeMappingFunc {
	return func(src *Node) error {
		target.SetString(extractString(reflect.ValueOf(src.Value)))
		return nil
	}
}

func extractInt(v reflect.Value) int64 {
	switch v.Kind() {
	case reflect.String:
		str := v.String()
		d, _ := strconv.ParseInt(str, 10, 64)
		return d
	case reflect.Int:
		return v.Int()
	case reflect.Uint:
		return int64(v.Uint())
	default:
		return 0
	}
}
func mapInt(target reflect.Value) NodeMappingFunc {
	return func(src *Node) error {
		target.SetInt(extractInt(reflect.ValueOf(src.Value)))
		return nil
	}
}

func extractUint(v reflect.Value) uint64 {
	kind := v.Kind()
	switch kind {
	case reflect.String:
		str := v.String()
		d, _ := strconv.ParseUint(str, 10, 64)
		return d
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	default:
		return 0
	}
}
func mapUint(target reflect.Value) NodeMappingFunc {
	return func(src *Node) error {
		target.SetUint(extractUint(reflect.ValueOf(src.Value)))
		return nil
	}
}

func extractDuration(v reflect.Value) int64 {
	k := v.Kind()
	switch k {
	case reflect.String:
		str := v.String()
		d, _ := time.ParseDuration(str)
		return int64(d)
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		return v.Int()
	default:
		return 0
	}
}
func mapDuration(target reflect.Value) NodeMappingFunc {
	return func(src *Node) error {
		target.SetInt(extractDuration(reflect.ValueOf(src.Value)))
		return nil
	}
}

func extractBool(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		b, _ := strconv.ParseBool(v.String())
		return b
	default:
		return false
	}
}
func mapBool(target reflect.Value) NodeMappingFunc {
	return func(src *Node) error {
		target.SetBool(extractBool(reflect.ValueOf(src.Value)))
		return nil
	}
}
