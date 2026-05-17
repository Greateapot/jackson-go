package jacksongo

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

var (
	ErrUnsupportedType   = errors.New("unsupported type")
	ErrUnsupportedMapKey = errors.New("map key must be string")
)

type EncoderOptions struct {
	ReferenceCodec ReferenceCodec
}

type Encoder struct {
	seen   map[uintptr]int
	nextID int
	refs   ReferenceCodec
}

func Marshal(v any) ([]byte, error) {
	return MarshalWithOptions(v, EncoderOptions{})
}

func MarshalWithOptions(v any, opts EncoderOptions) ([]byte, error) {
	codec := opts.ReferenceCodec
	if codec == nil {
		codec = ScalarReferenceCodec{}
	}

	e := &Encoder{
		seen:   make(map[uintptr]int),
		nextID: 1,
		refs:   codec,
	}

	normalized, err := e.encode(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}

	return json.Marshal(normalized)
}

func (e *Encoder) encode(v reflect.Value) (any, error) {
	if !v.IsValid() {
		return nil, nil
	}

	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return nil, nil
		}

		ptr := v.Pointer()

		if id, ok := e.seen[ptr]; ok {
			return e.refs.EncodeRef(id), nil
		}

		id := e.nextID
		e.nextID++
		e.seen[ptr] = id

		obj, err := e.encode(v.Elem())
		if err != nil {
			return nil, err
		}

		if m, ok := obj.(map[string]any); ok {
			m["@id"] = id
		}

		return obj, nil

	case reflect.Struct:
		result := make(map[string]any)
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)

			if !field.IsExported() {
				continue
			}

			name, ok := jsonFieldName(field)
			if !ok {
				continue
			}

			encoded, err := e.encode(v.Field(i))
			if err != nil {
				return nil, err
			}

			result[name] = encoded
		}

		return result, nil

	case reflect.Map:
		if v.IsNil() {
			return nil, nil
		}

		if v.Type().Key().Kind() != reflect.String {
			return nil, ErrUnsupportedMapKey
		}

		result := make(map[string]any)

		iter := v.MapRange()
		for iter.Next() {
			val, err := e.encode(iter.Value())
			if err != nil {
				return nil, err
			}

			result[iter.Key().String()] = val
		}

		return result, nil

	case reflect.Slice, reflect.Array:
		arr := make([]any, v.Len())

		for i := 0; i < v.Len(); i++ {
			x, err := e.encode(v.Index(i))
			if err != nil {
				return nil, err
			}
			arr[i] = x
		}

		return arr, nil

	case reflect.Interface:
		if v.IsNil() {
			return nil, nil
		}
		return e.encode(v.Elem())

	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return nil, ErrUnsupportedType

	default:
		return v.Interface(), nil
	}
}

func jsonFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")

	if tag == "-" {
		return "", false
	}

	if tag == "" {
		return field.Name, true
	}

	parts := strings.Split(tag, ",")
	name := parts[0]

	if name == "" {
		name = field.Name
	}

	return name, true
}
