package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	err error
)

type serverHandler struct{}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//set route
	mux["/"] = Root
}

func main() {
	//server config
	server := http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      &serverHandler{},
		ReadTimeout:  time.Duration(10) * time.Second,
		WriteTimeout: time.Duration(10) * time.Second,
	}

	if err = server.ListenAndServe(); err != nil {
		log.Printf(fmt.Sprintf("%v", err))
	}
}

//server route handler
func (serverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}
	log.Printf(fmt.Sprintf("%v", r.URL.String()))
}

func Root(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Body)

}
