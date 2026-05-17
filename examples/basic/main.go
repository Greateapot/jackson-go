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

	users := []*User{
		alex,
		alex,
	}

	data, _ := jacksongo.Marshal(users)

	fmt.Println(string(data))

	var decoded []*User
	_ = jacksongo.Unmarshal(data, &decoded)

	fmt.Println(decoded[0] == decoded[1])
}

/* OUTPUT
[{"@id":1,"name":"Alex"},1]
true
*/
