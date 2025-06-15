package evon

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
)

const (
	envTag         = "env"
	evonTag        = "evon"
	sliceSeparator = ","
	tagOmitempty   = "omitempty"
)

var (
	ErrUnsupportedType = errors.New("")
)

type customMarshaller interface {
	MarshalEnv(prefix string) ([]*Node, error)
}

type customMarshallerFunc func(prefix string) ([]*Node, error)

func (c customMarshallerFunc) MarshalEnv(prefix string) ([]*Node, error) {
	return c(prefix)
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
		if ok {
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
		} else {
			marshaller = func(prefix string, ref reflect.Value) ([]*Node, error) {
				cm := &defaultSliceMarshaller{
					sliceRef: ref,
				}

				return cm.MarshalEnv(prefix)
			}
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
			cm = &defaultSliceMarshaller{
				sliceRef: ref,
			}
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
		cm = &defaultMapMarshaller{
			ref: ref,
		}
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

	n.InnerNodes = make([]*Node, 0, ref.NumField())

	switch ref.Type().PkgPath() {
	case "time":
		t := ref.Interface().(time.Time)
		n.Value = formatTime(t)
		return n, nil
	}
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Type().Field(i)
		t := toolbox.Coalesce(
			field.Tag.Get(evonTag),
			field.Tag.Get(envTag))
		tags := strings.Split(t, sliceSeparator)

		var skip, omitempty bool
		var tag string

		for _, t = range tags {
			if t == "-" {
				skip = true
				break
			}
			if t == tagOmitempty {
				omitempty = true
				continue
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

		if prefix != "" && tag != "" {
			tag = "_" + tag
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

type defaultSliceMarshaller struct {
	sliceRef reflect.Value
}

func (d *defaultSliceMarshaller) MarshalEnv(prefix string) ([]*Node, error) {
	nodes := make([]*Node, 0, d.sliceRef.Len())

	for idx := 0; idx < d.sliceRef.Len(); idx++ {
		v := d.sliceRef.Index(idx)
		localPref := prefix + fmt.Sprintf("_[%d]", idx)
		n, err := MarshalEnvWithPrefix(localPref, v.Interface())
		if err != nil {
			return nil, rerrors.Wrap(err, "error marshalling inside array", localPref)
		}

		nodes = append(nodes, n)
	}
	return nodes, nil
}

type defaultMapMarshaller struct {
	ref reflect.Value
}

func (d *defaultMapMarshaller) MarshalEnv(prefix string) ([]*Node, error) {
	keys := d.ref.MapKeys()
	out := make([]*Node, 0, len(keys))

	for _, key := range keys {
		value := d.ref.MapIndex(key).Interface()

		pref := prefix
		if pref != "" {
			pref += "_"
		}
		pref += nameToEvonName(fmt.Sprint(key.Interface()))

		n, err := MarshalEnvWithPrefix(pref, value)
		if err != nil {
			return nil, rerrors.Wrapf(err, "error marshalling inside map. Path %s", pref)
		}

		out = append(out, n)
	}

	return out, nil
}

func nameToEvonName(name string) string {
	return strings.Replace(name, "_", "-", -1)
}

func formatTime(t time.Time) string {
	if t.Nanosecond() != 0 {
		return t.Format(time.RFC3339Nano)
	}

	if t.Hour() != 0 ||
		t.Minute() != 0 ||
		t.Second() == 0 {
		return t.Format(time.DateTime)
	}

	return t.Format(time.DateOnly)
}
