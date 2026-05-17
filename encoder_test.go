package jacksongo

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

type testUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type taggedUser struct {
	Name   string `json:"name"`
	Hidden string `json:"-"`
	Alias  string `json:",omitempty"`
}

type friendUser struct {
	Name   string      `json:"name"`
	Friend *friendUser `json:"friend"`
}

type interfaceHolder struct {
	Value any `json:"value"`
}

type badStruct struct {
	Ch chan int `json:"ch"`
}

type badPtr struct {
	Inner *func()
}

type emptyTag struct {
	Field string `json:",omitempty"`
}

type privateField struct {
	Public  string `json:"public"`
	private string
}

type nilInterfaceHolder struct {
	Value any `json:"value"`
}

func decodeJSON(t *testing.T, b []byte) any {
	t.Helper()

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	return v
}

func TestMarshalNil(t *testing.T) {
	b, err := Marshal(nil)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "null" {
		t.Fatalf("expected null, got %s", string(b))
	}
}

func TestMarshalPrimitive(t *testing.T) {
	b, err := Marshal(42)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "42" {
		t.Fatalf("expected 42, got %s", string(b))
	}
}

func TestMarshalStruct(t *testing.T) {
	u := testUser{
		Name: "Alex",
		Age:  20,
	}

	b, err := Marshal(u)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).(map[string]any)

	if got["name"] != "Alex" {
		t.Fatal("name mismatch")
	}

	if got["age"] != float64(20) {
		t.Fatal("age mismatch")
	}
}

func TestMarshalPointerIdentity(t *testing.T) {
	u := &testUser{Name: "Alex"}

	data := []*testUser{u, u}

	b, err := Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).([]any)

	first := got[0].(map[string]any)

	if first["@id"] != float64(1) {
		t.Fatal("missing @id")
	}

	if got[1] != float64(1) {
		t.Fatal("expected reference")
	}
}

func TestMarshalDistinctPointers(t *testing.T) {
	a := &testUser{Name: "Alex"}
	bb := &testUser{Name: "Alex"}

	data := []*testUser{a, bb}

	b, err := Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).([]any)

	first := got[0].(map[string]any)
	second := got[1].(map[string]any)

	if first["@id"] == second["@id"] {
		t.Fatal("distinct pointers got same id")
	}
}

func TestMarshalCyclicRefs(t *testing.T) {
	a := &friendUser{Name: "Alex"}
	b := &friendUser{Name: "John"}

	a.Friend = b
	b.Friend = a

	res, err := Marshal(a)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, res).(map[string]any)

	if got["@id"] != float64(1) {
		t.Fatal("root id mismatch")
	}

	friend := got["friend"].(map[string]any)

	if friend["@id"] != float64(2) {
		t.Fatal("friend id mismatch")
	}

	if friend["friend"] != float64(1) {
		t.Fatal("cycle not preserved")
	}
}

func TestMarshalJSONTags(t *testing.T) {
	u := taggedUser{
		Name:   "Alex",
		Hidden: "secret",
		Alias:  "A",
	}

	b, err := Marshal(u)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).(map[string]any)

	if _, ok := got["Hidden"]; ok {
		t.Fatal("ignored field serialized")
	}

	if _, ok := got["name"]; !ok {
		t.Fatal("json tag ignored")
	}

	if _, ok := got["Alias"]; !ok {
		t.Fatal("empty json tag name broken")
	}
}

func TestMarshalSlice(t *testing.T) {
	b, err := Marshal([]int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "[1,2,3]" {
		t.Fatalf("unexpected slice: %s", string(b))
	}
}

func TestMarshalArray(t *testing.T) {
	b, err := Marshal([2]string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != `["a","b"]` {
		t.Fatalf("unexpected array: %s", string(b))
	}
}

func TestMarshalMap(t *testing.T) {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}

	b, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).(map[string]any)

	if got["a"] != float64(1) || got["b"] != float64(2) {
		t.Fatal("map mismatch")
	}
}

func TestMarshalMapWithPointers(t *testing.T) {
	u := &testUser{Name: "Alex"}

	m := map[string]*testUser{
		"x": u,
		"y": u,
	}

	b, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).(map[string]any)

	obj := got["x"].(map[string]any)

	if obj["@id"] != float64(1) {
		t.Fatal("missing id")
	}

	if got["y"] != float64(1) {
		t.Fatal("ref mismatch")
	}
}

func TestMarshalInterface(t *testing.T) {
	v := interfaceHolder{
		Value: &testUser{Name: "Alex"},
	}

	_, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMarshalNilPointer(t *testing.T) {
	var u *testUser

	b, err := Marshal(u)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "null" {
		t.Fatal("expected null")
	}
}

func TestMarshalNilMap(t *testing.T) {
	var m map[string]int

	b, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "null" {
		t.Fatal("expected null")
	}
}

func TestMarshalUnsupportedMapKey(t *testing.T) {
	_, err := Marshal(map[int]string{
		1: "a",
	})

	if err != ErrUnsupportedMapKey {
		t.Fatalf("expected ErrUnsupportedMapKey, got %v", err)
	}
}

func TestMarshalUnsupportedFunc(t *testing.T) {
	_, err := Marshal(func() {})

	if err != ErrUnsupportedType {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestMarshalUnsupportedChan(t *testing.T) {
	ch := make(chan int)

	_, err := Marshal(ch)

	if err != ErrUnsupportedType {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestMarshalNestedUnsupportedStructField(t *testing.T) {
	_, err := Marshal(badStruct{
		Ch: make(chan int),
	})

	if err != ErrUnsupportedType {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestMarshalNestedUnsupportedSlice(t *testing.T) {
	_, err := Marshal([]any{
		make(chan int),
	})

	if err != ErrUnsupportedType {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestMarshalNestedUnsupportedMapValue(t *testing.T) {
	_, err := Marshal(map[string]any{
		"x": make(chan int),
	})

	if err != ErrUnsupportedType {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestMarshalNestedUnsupportedPointer(t *testing.T) {
	fn := func() {}

	_, err := Marshal(&badPtr{
		Inner: &fn,
	})

	if err != ErrUnsupportedType {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestMarshalNilInterface(t *testing.T) {
	var x any

	b, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "null" {
		t.Fatalf("expected null, got %s", b)
	}
}

func TestMarshalEmptyJSONTag(t *testing.T) {
	b, err := Marshal(emptyTag{Field: "x"})
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).(map[string]any)

	if got["Field"] != "x" {
		t.Fatal("field name fallback failed")
	}
}

func TestMarshalSkipsUnexportedField(t *testing.T) {
	v := privateField{
		Public:  "ok",
		private: "hidden",
	}

	b, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).(map[string]any)

	if got["public"] != "ok" {
		t.Fatal("public field missing")
	}

	if _, exists := got["private"]; exists {
		t.Fatal("unexported field should not be serialized")
	}
}

func TestMarshalTypedNilInterface(t *testing.T) {
	var p *testUser = nil

	var v any = p

	b, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "null" {
		t.Fatalf("expected null, got %s", string(b))
	}
}

func TestMarshalNilInterfaceField(t *testing.T) {
	v := nilInterfaceHolder{
		Value: nil,
	}

	b, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).(map[string]any)

	if got["value"] != nil {
		t.Fatal("expected nil interface field")
	}
}

func TestMarshalWithObjectReferenceCodec(t *testing.T) {
	alex := &testUser{Name: "Alex"}

	b, err := MarshalWithOptions(
		[]*testUser{alex, alex},
		EncoderOptions{
			ReferenceCodec: ObjectReferenceCodec{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).([]any)

	ref := got[1].(map[string]any)

	if ref["@ref"] != float64(1) {
		t.Fatalf("expected @ref=1, got %v", ref["@ref"])
	}
}

func TestRoundTripObjectReferenceCodec(t *testing.T) {
	alex := &friendUser{Name: "Alex"}
	john := &friendUser{Name: "John"}

	alex.Friend = john
	john.Friend = alex

	b, err := MarshalWithOptions(
		[]*friendUser{alex, alex, john},
		EncoderOptions{
			ReferenceCodec: ObjectReferenceCodec{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var decoded []*friendUser

	err = UnmarshalWithOptions(
		b,
		&decoded,
		DecoderOptions{
			ReferenceCodec: ObjectReferenceCodec{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if decoded[0] != decoded[1] {
		t.Fatal("shared ref broken")
	}

	if decoded[2].Friend != decoded[0] {
		t.Fatal("cycle broken")
	}
}

type stringRefCodec struct{}

func (stringRefCodec) EncodeRef(id int) any {
	return "$ref:" + strconv.Itoa(id)
}

func (stringRefCodec) DecodeRef(raw any) (int, bool, error) {
	s, ok := raw.(string)
	if !ok {
		return 0, false, nil
	}

	if !strings.HasPrefix(s, "$ref:") {
		return 0, false, nil
	}

	id, err := strconv.Atoi(strings.TrimPrefix(s, "$ref:"))
	if err != nil {
		return 0, true, err
	}

	return id, true, nil
}

func TestMarshalWithCustomReferenceCodec(t *testing.T) {
	alex := &testUser{Name: "Alex"}

	b, err := MarshalWithOptions(
		[]*testUser{alex, alex},
		EncoderOptions{
			ReferenceCodec: stringRefCodec{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	got := decodeJSON(t, b).([]any)

	if got[1] != "$ref:1" {
		t.Fatalf("unexpected ref: %v", got[1])
	}
}
