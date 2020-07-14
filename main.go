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
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	}
}

//钉钉消息提示数据结构
//text文本提醒
type DingText struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
	//At      At     `json:"at"`
}

type Text struct {
	Content string `json:"content"`
}

type At struct {
	AtMobiles [1]string `json:"atMobiles"`
	IsAtAll   string    `json:"isAtAll"`
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

	log.Println("Start RespBerry HTTPServer")
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
	R.response(w)
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
		log.Println(R)
		return
	}

	return
}

func (R *RobotResponse) response(w http.ResponseWriter) {
	var b []byte
	log.Println("R.MsgType", R.MsgType)
	log.Println("R.Text.Content", R.Text.Content)

	content := "RespBerry HTTPServer"
	var d = DingText{
		"text",
		Text{
			content,
		},
	}
	if b, err = json.Marshal(d); err == nil {
		log.Printf("[DingAlert] Send TO DingTalk %v ", string(b))
	}

	//// 忽略证书校验
	//	//tr := &http.Transport{
	//	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//	//}

	_, err = io.WriteString(w, string(b))
}
