package main

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var (
	mux              = make(map[string]func(http.ResponseWriter, *http.Request))
	muxEMQContent    = make(map[string]int)
	muxResponseEmoji [10]string
	err              error
)

type serverHandler struct{}

type RobotResponse struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	}
	EMQCtt        int
	ResponseEmoji string
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

	//set flag
	muxEMQContent["开灯"] = 1
	muxEMQContent["关灯"] = 2
	muxEMQContent["开启监控"] = 3
	muxEMQContent["关闭监控"] = 4
	muxEMQContent["拍照"] = 5

	//set Emoji for Response
	muxResponseEmoji[0] = "[捧脸]"
	muxResponseEmoji[0] = "[凄凉]"
	muxResponseEmoji[0] = "[发呆]"
	muxResponseEmoji[0] = "[灵感]"
	muxResponseEmoji[0] = "[迷惑]"
	muxResponseEmoji[0] = "[天使]"
	muxResponseEmoji[0] = "[无聊]"
	muxResponseEmoji[0] = "[亲亲]"
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
		RR RobotResponse
	)
	_ = RR.initializeBody(r.Body)

	RR.response(w)
	RR.pipLine()
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

func (R *RobotResponse) response(w http.ResponseWriter) {
	var b []byte
	log.Println("R.MsgType", R.MsgType)
	log.Println("R.Text.Content", R.Text.Content)

	content := "RespBerry HTTPServer" + R.ResponseEmoji
	var d = DingText{
		"text",
		Text{
			content,
		},
	}
	if b, err = json.Marshal(d); err == nil {
		log.Printf("[DingAlert] Send TO DingTalk %v ", string(b))
	}

	_, err = io.WriteString(w, string(b))
}

func (R *RobotResponse) pipLine() {

	var ctt = R.Text.Content

	//judge if exist
	if _, ok := muxEMQContent[ctt]; ok {
		R.EMQCtt = muxEMQContent[ctt]
		log.Println("[pipLine] Get ", ctt)
	} else {
		log.Println("[pipLine] No This Key")
	}

	//Take a random number
	randomEmoji := rand.Intn(len(muxResponseEmoji))

	//Take a random Emoji
	R.ResponseEmoji = muxResponseEmoji[randomEmoji]

}

func emqX() {

}
