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

	data := map[string]*User{
		"first":  alex,
		"second": alex,
	}

	b, _ := jacksongo.Marshal(data)

	fmt.Println(string(b))

	var decoded map[string]*User
	_ = jacksongo.Unmarshal(b, &decoded)

	fmt.Println(decoded["first"] == decoded["second"])
}

/* OUTPUT
{"first":{"@id":1,"name":"Alex"},"second":1}
true
*/
