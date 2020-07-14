package main

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	err error
)

type serverHandler struct{}

type RobotResponse struct {
	MsgType string
	Text    struct {
		Context string
	}
}

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
	var (
		R RobotResponse
	)
	log.Println(r.Body)
	_ = R.initializeBody(r.Body)
	log.Println("R.MsgType", R.MsgType)
	log.Println("R.Text", R.Text)

}

//InitializeBody : config initialize
func (R *RobotResponse) initializeBody(rBody io.Reader) (err error) {
	var (
		body    []byte
		jsonObj interface{}
	)

	// if the body exist
	if body, err = ioutil.ReadAll(rBody); err != nil {
		err = fmt.Errorf("Read Body ERR: %v ", err)
		return
	}

	// if the body can be turn to interface
	if err = json.Unmarshal(body, &jsonObj); err != nil {
		err = fmt.Errorf("Unmarshal Body ERR: %v", err)
		return
	}

	//turn map to struck
	if err = mapstructure.Decode(jsonObj, &R); err != nil {
		return
	}

	return
}
