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
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var globalConfig GlobalConfig

func main() {
	log.Println(Init("config.json"))
	log.Println(Send("c1", []byte("woaini")))
	log.Println(Send("c1", []byte("ilove you")))
	log.Println(Recv("c1"))
	log.Println(Close())
}

func Init(config string) (err error) {
	bsjson, err := ioutil.ReadFile(config)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bsjson, &globalConfig)
	if err != nil {
		return
	}

	err = globalConfig.Start()
	return
}

func Send(tag string, p []byte) (int, error) {
	return globalConfig.Clients[tag].Send(p)
}

func Recv(tag string) ([]byte, error) {
	return globalConfig.Clients[tag].Recv()
}

func Close() error {
	return globalConfig.Close()
}

//export multipleInit
func multipleInit(config *C.char) *C.char {
	sjson := C.GoString(config)
	err := Init(sjson)
	var serr string
	if err != nil {
		serr = err.Error()
	}
	return C.CString(serr)
}

//export multipleSend
func multipleSend(ptag, pdata *C.char, length C.int) (C.int, *C.char) {
	tag := C.GoString(ptag)
	n, err := Send(tag, C.GoBytes(unsafe.Pointer(pdata), length))
	var serr string
	if err != nil {
		serr = err.Error()
	}
	return C.int(n), C.CString(serr)
}

//export multipleRecv
func multipleRecv(ptag *C.char) (n C.int, pdata, cerr *C.char) {
	tag := C.GoString(ptag)
	data, err := Recv(tag)
	var serr string
	if err != nil {
		serr = err.Error()
	}
	return C.int(len(data)), (*C.char)(C.CBytes(data)), C.CString(serr)
}

//export multipleClose
func multipleClose() *C.char {
	err := Close()
	var serr string
	if err != nil {
		serr = err.Error()
	}
	return C.CString(serr)
}
