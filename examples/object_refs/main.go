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

	data, _ := jacksongo.MarshalWithOptions(
		[]*User{alex, alex},
		jacksongo.EncoderOptions{
			ReferenceCodec: jacksongo.ObjectReferenceCodec{},
		},
	)

	fmt.Println(string(data))

	var decoded []*User

	_ = jacksongo.UnmarshalWithOptions(
		data,
		&decoded,
		jacksongo.DecoderOptions{
			ReferenceCodec: jacksongo.ObjectReferenceCodec{},
		},
	)

	fmt.Println(decoded[0] == decoded[1])
}

/* OUTPUT
[{"@id":1,"name":"Alex"},{"@ref":1}]
true
*/
