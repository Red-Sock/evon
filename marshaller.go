package evon

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"go.redsock.ru/toolbox"
)

const (
	envTag         = "env"
	evonTag        = "evon"
	sliceSeparator = ","
	tagOmitempty   = "omitempty"
)

var (
	ErrSliceRequireMarshaller = errors.New("slices of non basic type require customMarshaller to be implemented")
	ErrUnsupportedType        = errors.New("")
)

type customMarshaller interface {
	MarshalEnv(prefix string) ([]*Node, error)
}

var StdMarshaller marshaller

type marshaller struct {
}

func MarshalEnv(in any) (*Node, error) {
	return StdMarshaller.marshal("", reflect.ValueOf(in))
}

func MarshalEnvWithPrefix(prefix string, in any) (*Node, error) {
	return StdMarshaller.marshal(prefix, reflect.ValueOf(in))
}

func Marshal(nodes []*Node) []byte {
	b := bytes.NewBuffer(nil)
	for _, node := range nodes {
		if node.Value != nil && len(node.InnerNodes) == 0 {
			b.WriteString(node.Name)
			b.WriteByte('=')
			b.WriteString(fmt.Sprint(node.Value))
			b.WriteByte('\n')
		}
		b.Write(Marshal(node.InnerNodes))
	}
	return b.Bytes()
}

func (m marshaller) marshal(prefix string, ref reflect.Value) (n *Node, err error) {
	prefix = strings.ToUpper(prefix)

	switch ref.Kind() {
	case reflect.Slice:
		n, err = marshalSlice(prefix, ref)
	case reflect.Struct:
		n, err = marshalStruct(prefix, ref)
	case reflect.Ptr:
		if ref.IsNil() {
			return nil, nil
		}

		n, err = m.marshal(prefix, ref.Elem())

	case reflect.Map:
		n, err = marshalMap(prefix, ref)
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n = &Node{
			Name:  prefix,
			Value: ref.Interface(),
		}
	default:
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error marshalling node: %w", err)
	}

	return n, nil
}

func marshalSlice(prefix string, ref reflect.Value) (*Node, error) {
	if ref.Len() == 0 {
		return nil, nil
	}

	n := &Node{
		Name: prefix,
	}

	tp := ref.Index(0).Kind()

	var marshaller func(prefix string, ref reflect.Value) ([]*Node, error)
	switch {
	case tp == reflect.Struct:
		val := ref.Index(0).Interface()
		_, ok := val.(customMarshaller)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, tp.String())
		}
		marshaller = func(prefix string, ref reflect.Value) ([]*Node, error) {
			out := make([]*Node, 0, ref.Len())
			for i := range ref.Len() {
				val = ref.Index(i).Interface()
				m, ok := val.(customMarshaller)
				if !ok {
					return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, tp.String())
				}

				node, err := m.MarshalEnv(prefix)
				if err != nil {
					return nil, fmt.Errorf("error marshalling slice element: %w", err)
				}

				out = append(out, node...)
			}

			return out, nil
		}

	case tp == reflect.Interface, tp == reflect.Ptr:
		var val any
		if ref.CanAddr() {
			val = ref.Addr().Interface()
		} else {
			val = ref.Interface()
		}
		cm, ok := val.(customMarshaller)
		if !ok {
			return nil, ErrSliceRequireMarshaller
		}

		marshaller = func(prefix string, ref reflect.Value) ([]*Node, error) {
			return cm.MarshalEnv(prefix)
		}

	case tp < reflect.Complex64, tp == reflect.String:
		node, err := marshallSliceOfBasicType(prefix, ref)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return node, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, tp.String())
	}

	var err error
	n.InnerNodes, err = marshaller(prefix, ref)
	if err != nil {
		return nil, fmt.Errorf("%w:%s", err, "error marshalling")
	}
	return n, nil
}
func marshalMap(prefix string, ref reflect.Value) (*Node, error) {
	val := ref.Interface()

	cm, ok := val.(customMarshaller)
	if !ok {
		return nil, ErrSliceRequireMarshaller
	}

	innerNodes, err := cm.MarshalEnv(prefix)
	if err != nil {
		return nil, fmt.Errorf("%w:%s", err, "error marshalling env")
	}
	return &Node{
		Name:       prefix,
		InnerNodes: innerNodes,
	}, nil
}
func marshallSliceOfBasicType(prefix string, ref reflect.Value) (*Node, error) {
	out := &Node{}

	outStr := make([]string, 0, ref.Len())
	for i := 0; i < ref.Len(); i++ {
		elem := fmt.Sprint(ref.Index(i).Interface())
		outStr = append(outStr, elem)
	}

	out.Name = prefix
	out.Value = strings.Join(outStr, sliceSeparator)
	return out, nil
}

func marshalStruct(prefix string, ref reflect.Value) (*Node, error) {
	n := &Node{
		Name: prefix,
	}

	if prefix != "" {
		prefix += "_"
	}

	n.InnerNodes = make([]*Node, 0, ref.NumField())

	for i := 0; i < ref.NumField(); i++ {
		field := ref.Type().Field(i)
		t := toolbox.Coalesce(
			field.Tag.Get(evonTag),
			field.Tag.Get(envTag))
		tags := strings.Split(t, sliceSeparator)

		var skip, omitempty bool
		tag := tags[0]
		for _, t = range tags {
			if t == "-" {
				skip = true
				break
			}
			if t == tagOmitempty {
				omitempty = true
			}
			tag = t
		}

		if skip {
			continue
		}

		value := ref.Field(i)
		if value.IsZero() && omitempty {
			continue
		}

		if tag == "" {
			tag = splitToKebab(field.Name)
		}
		tag = prefix + tag

		node, err := StdMarshaller.marshal(tag, value)
		if err != nil {
			return nil, err
		}
		if node != nil {
			n.InnerNodes = append(n.InnerNodes, node)
		}
	}

	return n, nil
}

func splitToKebab(in string) string {
	inR := []rune(in)
	out := make([]rune, 0, len(inR)+2)
	for idx, r := range inR {
		if unicode.IsUpper(r) && idx != 0 {
			out = append(out, '-')
		}

		out = append(out, r)
	}

	return string(out)
}
