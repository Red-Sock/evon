package evon

import (
	"strings"
)

type unmarshalOpts struct {
	keyName func(string) string
}

type unmarshalOpt func(o *unmarshalOpts)

func WithSnakeUnmarshal() func(o *unmarshalOpts) {
	return func(o *unmarshalOpts) {
		o.keyName = func(s string) string {
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, "-", "_")
			return s
		}
	}
}
