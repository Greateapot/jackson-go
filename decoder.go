package jacksongo

import (
	"encoding/json"
	"errors"
	// "fmt"
	"reflect"
)

var (
	ErrNonPointerPassed = errors.New("non-pointer passed")
	ErrNoPointerObject  = errors.New("expected object for pointer")
)

type DecoderOptions struct {
	ReferenceCodec ReferenceCodec
}

type Decoder struct {
	objects map[int]reflect.Value
	refs    ReferenceCodec
}

func Unmarshal(data []byte, v any) error {
	return UnmarshalWithOptions(data, v, DecoderOptions{})
}

func UnmarshalWithOptions(data []byte, v any, opts DecoderOptions) error {
	if reflect.TypeOf(v).Kind() != reflect.Pointer {
		return ErrNonPointerPassed
	}

	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	codec := opts.ReferenceCodec
	if codec == nil {
		codec = ScalarReferenceCodec{}
	}

	d := &Decoder{
		objects: make(map[int]reflect.Value),
		refs:    codec,
	}

	result, err := d.decode(raw, reflect.TypeOf(v).Elem())
	if err != nil {
		return err
	}

	reflect.ValueOf(v).Elem().Set(result)
	return nil
}

func (d *Decoder) decode(raw any, t reflect.Type) (reflect.Value, error) {
	if raw == nil {
		return reflect.Zero(t), nil
	}

	switch t.Kind() {
	case reflect.Pointer:
		elem := t.Elem()

		if !isReferenceType(elem) {
			val, err := d.decode(raw, elem)
			if err != nil {
				return reflect.Value{}, err
			}

			ptr := reflect.New(elem)
			ptr.Elem().Set(val)

			return ptr, nil
		}

		if id, ok, err := d.refs.DecodeRef(raw); err != nil {
			return reflect.Value{}, err
		} else if ok {
			obj, exists := d.objects[id]
			if !exists {
				return reflect.Zero(t), nil
				// return reflect.Value{}, fmt.Errorf("unknown reference id %d", id)
			}

			return obj, nil
		}

		rawMap, ok := raw.(map[string]any)
		if !ok {
			return reflect.Value{}, ErrNoPointerObject
		}

		obj := reflect.New(t.Elem())

		if idRaw, ok := rawMap["@id"]; ok {
			id := int(idRaw.(float64))
			d.objects[id] = obj
		}

		err := d.fillStruct(rawMap, obj.Elem())
		return obj, err

	case reflect.Struct:
		rawMap := raw.(map[string]any)

		obj := reflect.New(t).Elem()
		err := d.fillStruct(rawMap, obj)
		return obj, err

	case reflect.Map:
		rawMap := raw.(map[string]any)

		result := reflect.MakeMap(t)

		for k, v := range rawMap {
			decoded, err := d.decode(v, t.Elem())
			if err != nil {
				return reflect.Value{}, err
			}

			result.SetMapIndex(
				reflect.ValueOf(k),
				decoded,
			)
		}

		return result, nil

	case reflect.Slice:
		rawArr := raw.([]any)

		slice := reflect.MakeSlice(t, len(rawArr), len(rawArr))

		for i, item := range rawArr {
			val, err := d.decode(item, t.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			slice.Index(i).Set(val)
		}

		return slice, nil

	default:
		val := reflect.ValueOf(raw)

		if val.Type().ConvertibleTo(t) {
			return val.Convert(t), nil
		}

		return reflect.Zero(t), nil
	}
}

func (d *Decoder) fillStruct(raw map[string]any, v reflect.Value) error {
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

		rawField, exists := raw[name]
		if !exists {
			continue
		}

		decoded, err := d.decode(rawField, field.Type)
		if err != nil {
			return err
		}

		v.Field(i).Set(decoded)
	}

	return nil
}

func isReferenceType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}
