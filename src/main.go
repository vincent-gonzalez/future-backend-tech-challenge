package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func indexHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return
	}

	switch req.Method {
	case http.MethodGet:
		res.Write([]byte("<h1>Hello World 2!</h1>"))
	}
}

func main() {
	fmt.Println("Hello World!")
	//mux := http.NewServeMux()
	//mux.HandleFunc("/", indexHandler)
	router := httprouter.New()
	router.GET("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8081", router))
}
