package main

import (
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/mitchellh/mapstructure"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	Title           = "[惊喜] RespBerry HTTP Server [惊喜]"
	EMQxAds         = "emqx.ranzhendong.com.cn:1883"
	EMQxTopic       = "respberry"
	EMQxSetClientID = "respberry"
)

var (
	f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("TOPIC: %s\n", msg.Topic())
		log.Printf("MSG: %s\n", msg.Payload())
	}
	mux              = make(map[string]func(http.ResponseWriter, *http.Request, *mqtt.Client))
	muxResponseEMQ   = make(map[string]int)
	muxResponseEmoji [10]string
	err              error
)

type serverHandler struct {
	Connect mqtt.Client
}

type RobotResponse struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	}
	EMQ struct {
		Connect mqtt.Client
		Ext     bool
		CttKey  string
		CttVal  int
	}
	ResponseEmoji   string
	ResponseContent string
}

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

	//set flag for emqX
	muxResponseEMQ["HELPME"] = 0
	muxResponseEMQ["开灯"] = 1
	muxResponseEMQ["关灯"] = 2
	muxResponseEMQ["开启监控"] = 3
	muxResponseEMQ["关闭监控"] = 4
	muxResponseEMQ["拍照"] = 5

	//set Emoji for Response
	muxResponseEmoji[0] = "[捧脸]"
	muxResponseEmoji[1] = "[凄凉]"
	muxResponseEmoji[2] = "[发呆]"
	muxResponseEmoji[3] = "[灵感]"
	muxResponseEmoji[4] = "[迷惑]"
	muxResponseEmoji[5] = "[天使]"
	muxResponseEmoji[6] = "[无聊]"
	muxResponseEmoji[7] = "[亲亲]"
}

func main() {

	//server config
	server := http.Server{
		Addr: "0.0.0.0:8080",
		Handler: &serverHandler{
			//创建emqX连接
			emqXConnect(),
		},
		ReadTimeout:  time.Duration(10) * time.Second,
		WriteTimeout: time.Duration(10) * time.Second,
	}

	log.Println("Start RespBerry HTTPServer")
	if err = server.ListenAndServe(); err != nil {
		log.Printf(fmt.Sprintf("%v", err))
	}
}

//server route handler
func (s serverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r, &s.Connect)
		return
	}
	log.Printf(fmt.Sprintf("%v", r.URL.String()))
}

func Root(w http.ResponseWriter, r *http.Request, c *mqtt.Client) {
	var (
		RR RobotResponse
	)

	RR.EMQ.Connect = *c

	_ = RR.initializeBody(r.Body)

	RR.pipLine()

	RR.responseContent()

	RR.dingTalk(w)
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

func (R *RobotResponse) responseContent() {
	log.Println("R.MsgType", R.MsgType)
	log.Println("R.Text.Content", R.Text.Content)

	if R.EMQ.Ext && R.EMQ.CttKey == "HELPME" {
		R.ResponseContent = Title + "\n" + R.ResponseEmoji +
			"提供如下功能" +
			"\n" + "1.开灯" +
			"\n" + "2.关灯" +
			"\n" + "3.开启监控" +
			"\n" + "4.关闭监控" +
			"\n" + "5.拍照" +
			"\n" + "HELPME"
	} else if R.EMQ.Ext {
		R.ResponseContent = Title + "\n" +
			"【" + R.EMQ.CttKey + "】选项已经生效啦" + R.ResponseEmoji
		R.emqXPublish()
	} else {
		R.ResponseContent = Title + "\n" +
			"没有【" + R.EMQ.CttKey + "】选项啦" + R.ResponseEmoji
	}
}

func (R *RobotResponse) pipLine() {

	var (
		ctt = R.Text.Content
	)

	//judge if exist
	for k, v := range muxResponseEMQ {
		reg := regexp.MustCompile(k)
		if MEId := reg.FindString(ctt); MEId != "" {
			R.EMQ.Ext = true
			R.EMQ.CttKey = k
			R.EMQ.CttVal = v
			log.Println("[pipLine] Get ", ctt)
		}
	}

	if R.EMQ.CttKey == "" {
		R.EMQ.CttKey = R.Text.Content
	}

	//Take a random number
	randomEmoji := rand.Intn(len(muxResponseEmoji) - 1)
	log.Println(randomEmoji)

	//Take a random Emoji
	R.ResponseEmoji = muxResponseEmoji[randomEmoji]
	log.Println(R)

}

func emqXConnect() mqtt.Client {
	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	//mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + EMQxAds).SetClientID(EMQxSetClientID)

	opts.SetKeepAlive(60 * time.Second)
	// Set the message callback handler
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return c
}

func (R *RobotResponse) emqXPublish() {
	// Publish message
	c := R.EMQ.Connect

	// The payload must be a string that needs to be converted using the Strconv.itoa method
	token := c.Publish(EMQxTopic, 0, false, strconv.Itoa(R.EMQ.CttVal))
	token.Wait()
}

func (R *RobotResponse) dingTalk(w http.ResponseWriter) {
	var b []byte

	var d = DingText{
		"text",
		Text{
			R.ResponseContent,
		},
	}
	if b, err = json.Marshal(d); err == nil {
		log.Printf("[DingAlert] Send TO DingTalk %v ", string(b))
	}

	_, err = io.WriteString(w, string(b))
}
