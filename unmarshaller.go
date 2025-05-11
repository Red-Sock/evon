package evon

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrCustomMarshallerRequired = errors.New("slices of non basic type and maps require customMarshaller to be implemented")
)

type CustomUnmarshaler interface {
	UnmarshalEnv(env *Node) error
}

type NodeMappingFunc func(v *Node) error

func Unmarshal(bytes []byte, dst any) error {
	srcNodes := ParseToNodes(bytes)
	return unmarshal("", srcNodes, dst)
}

func UnmarshalWithPrefix(prefix string, bytes []byte, dst any) error {
	srcNodes := ParseToNodes(bytes)
	return unmarshal(prefix, srcNodes, dst)
}

func NodeToStruct(prefix string, node *Node, dst any) error {
	ns := NodeStorage{}
	for _, innerNode := range node.InnerNodes {
		ns.AddNode(innerNode)
	}
	return unmarshal(prefix, ns, dst)
}

func UnmarshalWithNodes(srcNodes NodeStorage, dst any) error {
	return unmarshal("", srcNodes, dst)
}

func UnmarshalWithNodesAndPrefix(prefix string, srcNodes NodeStorage, dst any) error {
	return unmarshal(prefix, srcNodes, dst)
}

func unmarshal(prefix string, srcNodes NodeStorage, dst any) error {
	dstValuesMapper, err := structToValueMapper(prefix, dst)
	if err != nil {
		return fmt.Errorf("%w:%s", err, "error getting struct value")
	}

	for key, srcVal := range srcNodes {
		setDstValue, ok := dstValuesMapper[key]
		if ok {
			err = setDstValue(srcVal)
			if err != nil {
				return fmt.Errorf("%w:%s", err, "error setting value")
			}
		}
	}

	return nil
}

func structToValueMapper(prefix string, dst any) (map[string]NodeMappingFunc, error) {
	valuesMapper := map[string]NodeMappingFunc{}
	dstReflectVal := reflect.ValueOf(dst)
	err := extractMappingForTarget(prefix, dstReflectVal, valuesMapper)
	if err != nil {
		return nil, fmt.Errorf("%w:%s", err, "error extracting mapping for target")
	}

	return valuesMapper, nil
}
func extractMappingForTarget(prefix string, target reflect.Value, valueMapping map[string]NodeMappingFunc) error {
	kind := target.Kind()

	var valueMapFunc NodeMappingFunc
	switch kind {
	case reflect.Pointer, reflect.Struct:
		if prefix != "" {
			prefix += "_"
		}
		if kind == reflect.Pointer {
			target = target.Elem()
		}

		for i := 0; i < target.NumField(); i++ {
			targetField := target.Type().Field(i)
			tags := strings.Split(targetField.Tag.Get(envTag), ",")
			var tag string
			for _, t := range tags {
				if t == tagOmitempty {
					continue
				}
				tag = t
				break
			}

			if tag == "-" {
				continue
			}

			if tag == "" {
				tag = splitToKebab(targetField.Name)
			}

			field := target.Field(i)
			err := extractMappingForTarget(prefix+tag, field, valueMapping)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
		}
		return nil

	case reflect.Slice, reflect.Map:
		// TODO добавить проверку на базовый / не базовый типы
		if !target.CanAddr() {
			return nil
		}
		k := target.Addr()

		val := k.Interface()
		cm, ok := val.(CustomUnmarshaler)
		if !ok {
			return ErrCustomMarshallerRequired
		}
		valueMapFunc = cm.UnmarshalEnv

	default:
		valueMapFunc = getBasicTypeMappingFunc(kind, target)
	}

	if valueMapFunc != nil {
		envName := strings.ToUpper(prefix)
		valueMapping[envName] = valueMapFunc
	}

	return nil
}

func getBasicTypeMappingFunc(kind reflect.Kind, target reflect.Value) NodeMappingFunc {
	switch kind {
	case reflect.String:
		return mapString(target)
	case reflect.Bool:
		return mapBool(target)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		tp := target.Type().Name()
		if tp == "Duration" {
			return mapDuration(target)
		} else {
			return mapInt(target)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return mapUint(target)
	default:
		return nil
	}
}
