package global

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var globalConfig GlobalConfig

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
	return globalConfig.Endpoints[tag].Send(p)
}

func Recv(tag string, p []byte) (int, error) {
	return globalConfig.Endpoints[tag].Recv(p)
}

func Close() error {
	return globalConfig.Close()
}
