package main

import (
	"log"
	"ucp/global"
)

func main() {
	log.Println(global.Init("../../server.json"))
	buf := make([]byte, 1500)
	// for {
	n, err := global.Recv("s1", buf)
	log.Println(string(buf[:n]), err)
	// log.Println(global.Recv("s2"))
	log.Println(global.Send("s1", []byte("woyeshi")))
	// log.Println(global.Send("s2", []byte("metoo")))
	// }
	log.Println(global.Close())
}
