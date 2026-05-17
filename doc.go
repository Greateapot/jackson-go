// Package jacksongo provides JSON serialization with object identity
// preservation compatible with Jackson-style reference encoding.
//
// Repeated pointers are serialized as references instead of duplicated
// objects, preserving object graphs and cyclic references.
//
// Example:
//
//	alice := &User{Name: "Alice"}
//	data := []*User{alice, alice}
//
//	b, _ := jacksongo.Marshal(data)
//	// [{"@id":1,"name":"Alice"},1]
package jacksongo
