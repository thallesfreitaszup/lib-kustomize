package main

import (
	"encoding/json"
	"fmt"
	"lib-kustomize/kustomize"
)

func main() {
	//ticker := time.NewTicker(500 * time.Millisecond)
	//done := make(chan bool)
	//go func() {
	//	for {
	//		select {
	//		case <-done:
	//			fmt.Println("Finished")
	//			return
	//		case t := <-ticker.C:
	//			fmt.Println("Ticked at ", t)
	//		}
	//	}
	//}()
	//
	//time.Sleep(3 * time.Second)
	//ticker.Stop()
	//done <- true
	k := kustomize.New()
	manifests, err := k.Render("github.com/thallesfreitaszup/kustomize-demo", "overlays/dev")
	if err != nil {
		panic(err)
	}
	bytes, err := json.Marshal(manifests)
	fmt.Println(string(bytes))
}
