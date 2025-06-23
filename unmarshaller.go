package evon

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.redsock.ru/rerrors"
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

func UnmarshalWithNodes(srcNodes NodeStorage, dst any, opts ...unmarshalOpt) error {
	return unmarshal("", srcNodes, dst, opts...)
}

func UnmarshalWithNodesAndPrefix(prefix string, srcNodes NodeStorage, dst any) error {
	return unmarshal(prefix, srcNodes, dst)
}

func unmarshal(prefix string, srcNodes NodeStorage, dst any, opts ...unmarshalOpt) (err error) {
	dstRefVal := reflect.ValueOf(dst)

	var dstValuesMapper unmarshalMapper

	switch dstRefVal.Kind() {
	case reflect.Map:
		dstValuesMapper, err = newMapValueMapper(dst)
		if err != nil {
			return fmt.Errorf("error mapping to Golang's map: %w", err)
		}
	default:
		dstValuesMapper, err = newStructValueMapper(prefix, dstRefVal)
		if err != nil {
			return fmt.Errorf("%w:%s", err, "error getting struct value")
		}

	}

	unOpts := unmarshalOpts{
		keyName: func(s string) string { return s },
	}

	for _, opt := range opts {
		opt(&unOpts)
	}

	for key, srcVal := range srcNodes {
		keyPath := strings.Split(key, ObjectSplitter)
		for i := range keyPath {
			keyPath[i] = unOpts.keyName(keyPath[i])
		}

		err = dstValuesMapper.Map(keyPath, srcVal)
		if err != nil {
			return fmt.Errorf("%w:%s", err, "error setting value")
		}
	}

	dstValuesMapper.PostMapping()

	return nil
}

type unmarshalMapper interface {
	Map(keyPath []string, dst *Node) error
	PostMapping()
}

type structValueMapper struct {
	constructorsByPath map[string]NodeMappingFunc
}

func (s *structValueMapper) Map(keyPath []string, dst *Node) error {
	cbp, exists := s.constructorsByPath[strings.Join(keyPath, ObjectSplitter)]
	if exists {
		return cbp(dst)
	}

	return nil
}
func (s *structValueMapper) PostMapping() {

}
func newStructValueMapper(prefix string, dst reflect.Value) (unmarshalMapper, error) {
	valuesMapper := &structValueMapper{
		constructorsByPath: make(map[string]NodeMappingFunc),
	}

	err := extractMappingForTarget(prefix, dst, valuesMapper.constructorsByPath)
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
		if kind == reflect.Pointer {
			target = target.Elem()
			return extractMappingForTarget(prefix, target, valueMapping)
		}

		if prefix != "" {
			prefix += "_"
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
			cm = &defaultSliceUnmarshaller{
				ref: k,
			}
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

type mapValueMapper struct {
	m map[string]any
}

func (m mapValueMapper) Map(keyPath []string, dst *Node) error {
	if dst.Value == nil {
		return nil
	}

	node := m.m

	for _, pp := range keyPath[:len(keyPath)-1] {
		if pp == "" {
			continue
		}

		v, ok := node[pp]
		if !ok {
			newNode := map[string]any{}
			node[pp] = newNode
			node = newNode
		} else {
			newNode, _ := v.(map[string]any)
			if newNode == nil {
				newNode = map[string]any{}
				node[pp] = newNode
			}
			node = newNode
		}
	}

	node[keyPath[len(keyPath)-1]] = m.mapWithType(dst.Value)

	return nil
}

func (m mapValueMapper) PostMapping() {

	var fixSlice func(root map[string]any) []any

	fixSlice = func(root map[string]any) []any {
		sliced := make([]any, 0)

		for key, v := range root {
			innerMap, ok := v.(map[string]any)
			if ok {
				sl := fixSlice(innerMap)
				if len(sl) > 0 {
					root[key] = sl
				}
			}

			sliceIndexStart := strings.Index(key, "[")
			if sliceIndexStart == -1 {
				continue
			}

			sliceIndexEnd := strings.Index(key, "]")
			if sliceIndexStart == -1 {
				continue
			}

			indexStr := key[sliceIndexStart+1 : sliceIndexEnd]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				continue
			}

			if len(sliced) == 0 {
				sliced = make([]any, len(root))
			}

			sliced[index] = v
		}

		return sliced
	}

	_ = fixSlice(m.m)
}

func (m mapValueMapper) mapWithType(val any) any {
	s, ok := val.(string)
	if !ok {
		return val
	}

	t := tryParseTime(s)
	if t != nil {
		return customTime(*t)
	}

	if s == "true" {
		return true
	}

	if s == "false" {
		return false
	}

	if strings.Contains(s, ".") {
		fl, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return fl
		}
	}

	integer, err := strconv.Atoi(s)
	if err == nil {
		return integer
	}

	return val
}

func newMapValueMapper(dst any) (unmarshalMapper, error) {
	mapDst, ok := dst.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to unmarshal to map NOT map[string]any. Got: %T", dst)
	}
	m := mapValueMapper{
		m: mapDst,
	}

	return m, nil
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

type defaultSliceUnmarshaller struct {
	ref reflect.Value
}

func (d *defaultSliceUnmarshaller) UnmarshalEnv(rootSlice *Node) error {
	typpedSlice := d.ref.Elem()
	elemType := typpedSlice.Type().Elem()

	for idx, e := range rootSlice.InnerNodes {
		newElem := reflect.New(elemType).Elem()
		ns := NodeStorage{}

		e.RemovePrefix(e.Name)
		ns.AddNode(e)

		ne := newElem.Addr().Interface()
		err := unmarshal("", ns, ne)
		if err != nil {
			return rerrors.Wrapf(err,
				"error unmarshalling struct inside array. Path: %s_[%d]", rootSlice.Name, idx)
		}

		typpedSlice.Set(reflect.Append(typpedSlice, newElem))
	}
	return nil
}

var timeFormats = []string{
	time.DateOnly,
	time.DateTime,
	time.RFC3339Nano,
}

func tryParseTime(s string) *time.Time {
	for _, format := range timeFormats {
		if t, err := time.Parse(format, s); err == nil {
			return &t
		}
	}
	return nil
}
