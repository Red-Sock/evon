package env

import (
	"bytes"
	"fmt"

	errors "github.com/Red-Sock/trace-errors"
	"gopkg.in/yaml.v3"
)

type EnvVal struct {
	Name  string
	Value interface{}
}

func ExtractVariables(prefix string, m map[string]interface{}) (vals []EnvVal) {
	for k, v := range m {
		if prefix != "" {
			k = prefix + "_" + k
		}
		if newMap, ok := v.(map[string]interface{}); ok {
			vals = append(vals, ExtractVariables(k, newMap)...)
		} else {
			vals = append(vals, EnvVal{
				Name:  k,
				Value: v,
			})
		}
	}

	return vals
}

func ExtractFromAny(prefix string, in any) ([]EnvVal, error) {
	m, err := anyToMap(in)
	if err != nil {
		return nil, errors.Wrap(err, "error mapping in value")
	}

	return ExtractVariables(prefix, m), nil
}

func anyToMap(in any) (map[string]interface{}, error) {
	bytes, err := yaml.Marshal(in)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling to yaml")
	}

	m := make(map[string]interface{})

	err = yaml.Unmarshal(bytes, m)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling to map")
	}

	return m, nil
}

func ToFile(vals []EnvVal) []byte {
	b := &bytes.Buffer{}
	for _, val := range vals {
		b.Write([]byte(val.Name))
		b.WriteByte('=')
		b.Write([]byte(fmt.Sprint(val.Value)))
		b.WriteByte('\n')
	}

	return b.Bytes()
}
