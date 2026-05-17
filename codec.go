package jacksongo

import "errors"

var (
	ErrInvalidReference = errors.New("invalid reference")
)

type ReferenceCodec interface {
	EncodeRef(id int) any
	DecodeRef(raw any) (id int, ok bool, err error)
}

type ScalarReferenceCodec struct{}

func (ScalarReferenceCodec) EncodeRef(id int) any {
	return id
}

func (ScalarReferenceCodec) DecodeRef(raw any) (int, bool, error) {
	v, ok := raw.(float64)
	if !ok {
		return 0, false, nil
	}

	return int(v), true, nil
}

type ObjectReferenceCodec struct{}

func (ObjectReferenceCodec) EncodeRef(id int) any {
	return map[string]any{
		"@ref": id,
	}
}

func (ObjectReferenceCodec) DecodeRef(raw any) (int, bool, error) {
	m, ok := raw.(map[string]any)
	if !ok {
		return 0, false, nil
	}

	refRaw, exists := m["@ref"]
	if !exists {
		return 0, false, nil
	}

	v, ok := refRaw.(float64)
	if !ok {
		return 0, true, ErrInvalidReference
	}

	return int(v), true, nil
}
