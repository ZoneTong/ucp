package main

import (
	"log"
	"ucp/global"
)

func main() {
	log.Println(global.Init("../../client.json"))
	log.Println(global.Send("c1", []byte("woaini")))
	// log.Println(global.Send("c1", []byte("ilove you")))
	// log.Println(global.Recv("c1"))
	log.Println(global.Recv("c1"))
	log.Println(global.Close())
}
