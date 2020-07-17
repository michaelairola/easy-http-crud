package easyHttpCrud

import (
	"fmt"
	"log"
	"net/http"
)

func Serve(model interface{}, opts map[string]interface{}) {
	err := Init(model)
	if err != nil {
		fmt.Println(err)
		return
	}
	router := CreateRouter(model, opts)
	log.Printf("HTTP listening on port %v\n", Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", Port), router))
}
