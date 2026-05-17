package main

import (
	"fmt"

	"github.com/Greateapot/jackson-go"
)

type User struct {
	Name   string `json:"name"`
	Friend *User  `json:"friend"`
}

func main() {
	alex := &User{Name: "Alex"}
	john := &User{Name: "John"}

	alex.Friend = john
	john.Friend = alex

	data, _ := jacksongo.Marshal(alex)

	fmt.Println(string(data))

	var decoded *User
	_ = jacksongo.Unmarshal(data, &decoded)

	fmt.Println(decoded.Friend.Friend == decoded)
}

/* OUTPUT
{"@id":1,"friend":{"@id":2,"friend":1,"name":"John"},"name":"Alex"}
true
*/
