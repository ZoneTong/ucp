package main

import (
	"log"
	"ucp/global"
)

func main() {
	log.Println(global.Init("../../server.json"))
	for {
		buf, err := global.Recv("s1")
		log.Println(string(buf), err)
		// log.Println(global.Recv("s2"))
		log.Println(global.Send("s1", []byte("woyeshi")))
		// log.Println(global.Send("s2", []byte("metoo")))
	}
	log.Println(global.Close())
}
