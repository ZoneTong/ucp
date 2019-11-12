package main

// //go:generate go run -tags generate gen.go
// go build -buildmode=c-archive -o dist/number.a
// https://books.studygolang.com/advanced-go-programming-book/ch2-cgo/ch2-06-static-shared-lib.html
// 静态库

import "C"

func main() {}

//export number_add_mod
func number_add_mod(a, b, mod C.int) C.int {
	return (a + b) % mod
}
