package main

import (
	"log"
	"ucp/global"
)

func main() {
	log.Println(global.Init("../../client.json"))
	log.Println(global.Send("c1", []byte("woaini")))
	// log.Println(global.Send("c1", []byte("ilove you")))
	buf := make([]byte, 1500)
	n, err := global.Recv("c1", buf)
	log.Println(string(buf[:n]), err)
	log.Println(global.Close())
}
