package env

import (
	"reflect"
	"strings"
)

type CustomUnmarshaler interface {
	UnmarshalEnv(env *Node) error
}

type NodeMappingFunc func(v *Node) error

func Unmarshal(bytes []byte, dst any) {
	srcNodes := ParseToNodes(bytes)
	unmarshal("", srcNodes, dst)
}

func UnmarshalWithPrefix(prefix string, bytes []byte, dst any) {
	srcNodes := ParseToNodes(bytes)
	unmarshal(prefix, srcNodes, dst)
}

func NodeToStruct(prefix string, node *Node, dst any) {
	ns := NodeStorage{}
	for _, innerNode := range node.InnerNodes {
		ns.addNode(*innerNode)
	}
	unmarshal(prefix, ns, dst)
}

func unmarshal(prefix string, srcNodes NodeStorage, dst any) {
	dstValuesMapper := structToValueMapper(prefix, dst)

	for key, srcVal := range srcNodes {
		setDstValue, ok := dstValuesMapper[key]
		if ok {
			_ = setDstValue(srcVal)
		}
	}
}

func structToValueMapper(prefix string, dst any) map[string]NodeMappingFunc {
	valuesMapper := map[string]NodeMappingFunc{}
	dstReflectVal := reflect.ValueOf(dst)
	extractMappingForTarget(prefix, dstReflectVal, valuesMapper)

	return valuesMapper
}
func extractMappingForTarget(prefix string, target reflect.Value, valueMapping map[string]NodeMappingFunc) {
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
			tag := targetField.Tag.Get(envTag)
			if tag == "-" {
				continue
			}

			if tag == "" {
				tag = splitToSnake(targetField.Name)
			}

			field := target.Field(i)
			extractMappingForTarget(prefix+tag, field, valueMapping)
		}
		return

	case reflect.Slice:
		// TODO добавить проверку на базовый / не базовый типы
		if !target.CanAddr() {
			return
		}
		k := target.Addr()

		//if k.IsNil() {
		//	k.Set(reflect.New(k.Type().Elem()))
		//	k = k.Elem()
		//}
		val := k.Interface()
		customMarshaller, ok := val.(CustomUnmarshaler)
		if !ok {
			panic("Slices of non basic type require sliceMarshaller to be implemented")
		}
		valueMapFunc = customMarshaller.UnmarshalEnv

	default:
		valueMapFunc = getBasicTypeMappingFunc(kind, target)
	}

	if valueMapFunc != nil {
		envName := strings.ToUpper(prefix)
		valueMapping[envName] = valueMapFunc
	}
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
