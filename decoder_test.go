package jacksongo

import (
	"errors"
	"testing"
)

type nestedBad struct {
	User *testUser `json:"user"`
}

func TestUnmarshalNonPointer(t *testing.T) {
	var v []testUser

	err := Unmarshal([]byte(`[]`), v)

	if !errors.Is(err, ErrNonPointerPassed) {
		t.Fatalf("expected ErrNonPointerPassed, got %v", err)
	}
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	var v []testUser

	err := Unmarshal([]byte(`{invalid`), &v)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalNil(t *testing.T) {
	var v *testUser

	err := Unmarshal([]byte(`null`), &v)
	if err != nil {
		t.Fatal(err)
	}

	if v != nil {
		t.Fatal("expected nil")
	}
}

func TestUnmarshalPrimitive(t *testing.T) {
	var v int

	err := Unmarshal([]byte(`42`), &v)
	if err != nil {
		t.Fatal(err)
	}

	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestUnmarshalString(t *testing.T) {
	var v string

	err := Unmarshal([]byte(`"alex"`), &v)
	if err != nil {
		t.Fatal(err)
	}

	if v != "alex" {
		t.Fatal("string mismatch")
	}
}

func TestUnmarshalStruct(t *testing.T) {
	var u testUser

	err := Unmarshal([]byte(`{"name":"Alex","age":20}`), &u)
	if err != nil {
		t.Fatal(err)
	}

	if u.Name != "Alex" || u.Age != 20 {
		t.Fatal("struct mismatch")
	}
}

func TestUnmarshalJSONTags(t *testing.T) {
	var u taggedUser

	err := Unmarshal([]byte(`{
		"name":"Alex",
		"Alias":"A",
		"Hidden":"secret"
	}`), &u)

	if err != nil {
		t.Fatal(err)
	}

	if u.Name != "Alex" {
		t.Fatal("tagged field mismatch")
	}

	if u.Hidden != "" {
		t.Fatal("ignored field should stay zero")
	}
}

func TestUnmarshalPointerIdentity(t *testing.T) {
	var users []*testUser

	err := Unmarshal([]byte(`
	[
		{"@id":1,"name":"Alex","age":20},
		1
	]
	`), &users)

	if err != nil {
		t.Fatal(err)
	}

	if users[0] != users[1] {
		t.Fatal("identity not preserved")
	}
}

func TestUnmarshalUnknownReference(t *testing.T) {
	var users []*testUser

	err := Unmarshal([]byte(`[999]`), &users)

	if err != nil || len(users) != 1 || users[0] != nil {
		t.Fatal("expected nil")
	}
}

func TestUnmarshalExpectedPointerObject(t *testing.T) {
	var u *testUser

	err := Unmarshal([]byte(`"abc"`), &u)

	if !errors.Is(err, ErrNoPointerObject) {
		t.Fatalf("expected ErrNoPointerObject, got %v", err)
	}
}

func TestUnmarshalDistinctPointers(t *testing.T) {
	var users []*testUser

	err := Unmarshal([]byte(`
	[
		{"@id":1,"name":"Alex"},
		{"@id":2,"name":"Alex"}
	]
	`), &users)

	if err != nil {
		t.Fatal(err)
	}

	if users[0] == users[1] {
		t.Fatal("distinct objects collapsed")
	}
}

func TestUnmarshalCyclicRefs(t *testing.T) {
	var u *friendUser

	err := Unmarshal([]byte(`
	{
		"@id":1,
		"name":"Alex",
		"friend":{
			"@id":2,
			"name":"John",
			"friend":1
		}
	}
	`), &u)

	if err != nil {
		t.Fatal(err)
	}

	if u.Friend.Friend != u {
		t.Fatal("cycle not restored")
	}
}

func TestUnmarshalSlice(t *testing.T) {
	var v []int

	err := Unmarshal([]byte(`[1,2,3]`), &v)
	if err != nil {
		t.Fatal(err)
	}

	if len(v) != 3 || v[2] != 3 {
		t.Fatal("slice mismatch")
	}
}

func TestUnmarshalMap(t *testing.T) {
	var m map[string]int

	err := Unmarshal([]byte(`{"a":1,"b":2}`), &m)
	if err != nil {
		t.Fatal(err)
	}

	if m["a"] != 1 || m["b"] != 2 {
		t.Fatal("map mismatch")
	}
}

func TestUnmarshalMapWithPointers(t *testing.T) {
	var m map[string]*testUser

	err := Unmarshal([]byte(`
	{
		"x":{"@id":1,"name":"Alex"},
		"y":1
	}
	`), &m)

	if err != nil {
		t.Fatal(err)
	}

	if m["x"] != m["y"] {
		t.Fatal("map identity broken")
	}
}

func TestUnmarshalNestedSlicePointers(t *testing.T) {
	var v [][]*testUser

	err := Unmarshal([]byte(`
	[
		[
			{"@id":1,"name":"Alex"},
			1
		]
	]
	`), &v)

	if err != nil {
		t.Fatal(err)
	}

	if v[0][0] != v[0][1] {
		t.Fatal("nested identity broken")
	}
}

func TestUnmarshalInterface(t *testing.T) {
	var v interfaceHolder

	err := Unmarshal([]byte(`{"value":"abc"}`), &v)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalMissingFields(t *testing.T) {
	var u testUser

	err := Unmarshal([]byte(`{"name":"Alex"}`), &u)
	if err != nil {
		t.Fatal(err)
	}

	if u.Age != 0 {
		t.Fatal("missing field should stay zero")
	}
}

func TestUnmarshalSkipsUnexportedFields(t *testing.T) {
	var v privateField

	err := Unmarshal([]byte(`{
		"public":"ok",
		"private":"hidden"
	}`), &v)

	if err != nil {
		t.Fatal(err)
	}

	if v.Public != "ok" {
		t.Fatal("public not decoded")
	}

	if v.private != "" {
		t.Fatal("private field should stay zero")
	}
}

func TestRoundTripIdentity(t *testing.T) {
	alex := &friendUser{Name: "Alex"}
	john := &friendUser{Name: "John"}

	alex.Friend = john
	john.Friend = alex

	orig := []*friendUser{alex, alex, john}

	b, err := Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}

	var decoded []*friendUser
	err = Unmarshal(b, &decoded)
	if err != nil {
		t.Fatal(err)
	}

	if decoded[0] != decoded[1] {
		t.Fatal("shared ref broken")
	}

	if decoded[0].Friend != decoded[2] {
		t.Fatal("friend ref broken")
	}

	if decoded[2].Friend != decoded[0] {
		t.Fatal("cycle broken")
	}
}

func TestUnmarshalMapNestedDecodeError(t *testing.T) {
	var m map[string]*testUser

	err := Unmarshal([]byte(`{
		"x":"bad"
	}`), &m)

	if !errors.Is(err, ErrNoPointerObject) {
		t.Fatalf("expected ErrNoPointerObject, got %v", err)
	}
}

func TestUnmarshalNonConvertibleReturnsZero(t *testing.T) {
	var v int

	err := Unmarshal([]byte(`"abc"`), &v)
	if err != nil {
		t.Fatal(err)
	}

	if v != 0 {
		t.Fatalf("expected zero value, got %d", v)
	}
}

func TestUnmarshalStructFieldDecodeError(t *testing.T) {
	var v nestedBad

	err := Unmarshal([]byte(`{
		"user":"bad"
	}`), &v)

	if !errors.Is(err, ErrNoPointerObject) {
		t.Fatalf("expected ErrNoPointerObject, got %v", err)
	}
}

type failingCodec struct{}

func (failingCodec) EncodeRef(id int) any {
	return id
}

func (failingCodec) DecodeRef(raw any) (int, bool, error) {
	return 0, false, errors.New("decode ref failed")
}

func TestUnmarshalReferenceCodecError(t *testing.T) {
	var users []*testUser

	err := UnmarshalWithOptions(
		[]byte(`[1]`),
		&users,
		DecoderOptions{
			ReferenceCodec: failingCodec{},
		},
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "decode ref failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshalWithObjectReferenceCodec(t *testing.T) {
	var users []*testUser

	err := UnmarshalWithOptions(
		[]byte(`[
			{"@id":1,"name":"Alex"},
			{"@ref":1}
		]`),
		&users,
		DecoderOptions{
			ReferenceCodec: ObjectReferenceCodec{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if users[0] != users[1] {
		t.Fatal("identity not preserved")
	}
}

func TestUnmarshalWithCustomReferenceCodec(t *testing.T) {
	var users []*testUser

	err := UnmarshalWithOptions(
		[]byte(`[
			{"@id":1,"name":"Alex"},
			"$ref:1"
		]`),
		&users,
		DecoderOptions{
			ReferenceCodec: stringRefCodec{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if users[0] != users[1] {
		t.Fatal("custom codec identity broken")
	}
}

func TestUnmarshalCustomReferenceCodecInvalid(t *testing.T) {
	var users []*testUser

	err := UnmarshalWithOptions(
		[]byte(`["$ref:abc"]`),
		&users,
		DecoderOptions{
			ReferenceCodec: stringRefCodec{},
		},
	)

	if err == nil {
		t.Fatal("expected error")
	}
}
