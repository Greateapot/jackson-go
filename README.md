# jackson-go

Jackson-style object identity JSON serialization for Go.

`jackson-go` extends Go's standard JSON encoding model with **object identity preservation**, allowing repeated pointers and cyclic object graphs to be serialized without duplication.

This enables behavior similar to Java's Jackson object references.

## Features

- Object identity preservation
- Cyclic reference support
- Jackson-style `@id` references
- Pluggable reference codecs
- Full `json` struct tag support
- Compatible API design (`Marshal`, `Unmarshal`)
- Support for:
  - pointers
  - structs
  - slices
  - arrays
  - maps with string keys
  - interfaces

## Installation

```bash
go get github.com/Greateapot/jackson-go
```

## Basic Usage

```go
package main

import (
	"fmt"

	"github.com/Greateapot/jackson-go"
)

type User struct {
	Name string `json:"name"`
}

func main() {
	alex := &User{Name: "Alex"}

	users := []*User{alex, alex}

	data, _ := jacksongo.Marshal(users)

	fmt.Println(string(data))
}
```

Output:

```json
[{"@id":1,"name":"Alex"},1]
```


## Decoding

```go
var users []*User
_ = jacksongo.Unmarshal(data, &users)

fmt.Println(users[0] == users[1])
```

Output:

```json
true
```


## Cyclic References

```go
type User struct {
	Name   string `json:"name"`
	Friend *User  `json:"friend"`
}
```

```go
alex := &User{Name: "Alex"}
john := &User{Name: "John"}

alex.Friend = john
john.Friend = alex
```

## Reference Codecs

Reference representation is customizable.

### Default Scalar References

```json
[
  {"@id":1,"name":"Alex"},
  1
]
```


### Object References

```go
opts := jacksongo.EncoderOptions{
	ReferenceCodec: jacksongo.ObjectReferenceCodec{},
}
```

Produces:

```json
[
  {"@id":1,"name":"Alex"},
  {"@ref":1}
]
```


## Custom Reference Codec

Implement:

```go
type ReferenceCodec interface {
	EncodeRef(id int) any
	DecodeRef(raw any) (id int, ok bool, err error)
}
```

Example:

```go
type StringRefCodec struct{}

func (StringRefCodec) EncodeRef(id int) any {
	return "$ref:" + strconv.Itoa(id)
}

func (StringRefCodec) DecodeRef(raw any) (int, bool, error) {
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
```

Produces:

```json
[
  {"@id":1,"name":"Alex"},
  "$ref:1"
]
```

## Supported Types

### Supported

* pointers
* structs
* slices
* arrays
* maps with string keys
* interfaces
* cyclic object graphs

### Unsupported

* maps with non-string keys
* channels
* functions
* unsafe pointers


## Examples

See:

* `examples/basic`
* `examples/cyclic`
* `examples/object_refs`
* `examples/custom_codec`
* `examples/maps`


## Compatibility

Designed for interoperability with Java Jackson object identity patterns.


## License

MIT
