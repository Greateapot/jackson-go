package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Greateapot/jackson-go"
)

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

type User struct {
	Name string `json:"name"`
}

func main() {
	alex := &User{Name: "Alex"}

	codec := StringRefCodec{}

	data, _ := jacksongo.MarshalWithOptions(
		[]*User{alex, alex},
		jacksongo.EncoderOptions{
			ReferenceCodec: codec,
		},
	)

	fmt.Println(string(data))
}

/* OUTPUT
[{"@id":1,"name":"Alex"},"$ref:1"]
*/
