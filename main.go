package main

// //go:generate go run -tags generate gen.go
// go build -buildmode=c-archive -o dist/number.a
// https://books.studygolang.com/advanced-go-programming-book/ch2-cgo/ch2-06-static-shared-lib.html
// 静态库

import (
	"C"
	"unsafe"
)
import (
	"encoding/json"
	"io/ioutil"
)

var globalConfig GlobalConfig

func main() {}

//export multipleInit
func multipleInit(config *C.char) *C.char {
	sjson := C.GoString(config)
	bsjson, err := ioutil.ReadFile(sjson)
	if err != nil {
		return C.CString(err.Error())
	}
	json.Unmarshal(bsjson, &globalConfig)
	err = globalConfig.Start()
	var serr string
	if err != nil {
		serr = err.Error()
	}
	return C.CString(serr)
}

//export multipleSend
func multipleSend(ptag, pheader, pdata *C.char, length C.int) (C.int, *C.char) {
	tag := C.GoString(ptag)
	n, err := globalConfig.Clients[tag].Send(C.GoBytes(unsafe.Pointer(pdata), length))
	var serr string
	if err != nil {
		serr = err.Error()
	}
	return C.int(n), C.CString(serr)
}

//export multipleRecv
func multipleRecv(ptag, pheader *C.char) (n C.int, pdata, cerr *C.char) {
	tag := C.GoString(ptag)
	data, err := globalConfig.Clients[tag].Recv()
	var serr string
	if err != nil {
		serr = err.Error()
	}
	return C.int(len(data)), (*C.char)(C.CBytes(data)), C.CString(serr)
}

//export multipleClose
func multipleClose(config *C.char) *C.char {
	err := globalConfig.Close()
	var serr string
	if err != nil {
		serr = err.Error()
	}
	return C.CString(serr)
}
