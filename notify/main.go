package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gunsluo/wechatpay-go/v3"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

var (
	address        string
	appId          string
	apiv3Secret    string
	serialNo       string
	mchId          string
	privateKeyPath string
	payClient      wechatpay.Client
	tpl            *template.Template
)

func init() {
	tpl = template.New("")
	data, err := ioutil.ReadFile("templates/notify.html")
	if err != nil {
		log.Fatal(err)
	}
	tpl, err = tpl.Parse(string(data))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.StringVar(&address, "a", "127.0.0.1:8080", "listen port")
	flag.StringVar(&appId, "appid", "your appid", "appid")
	flag.StringVar(&apiv3Secret, "apiv3-secret", "your apiv3Secret", "apiv3Secret")
	flag.StringVar(&serialNo, "serial-no", "your serial_no", "serialNo")
	flag.StringVar(&mchId, "mchid", "your mchId", "mchId")
	flag.StringVar(&privateKeyPath, "private-key-path", "apiclient_key.pem", "private key path")
	flag.Parse()

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("unable to listen on %s, %v", address, err)
	}

	// create a client of wechat pay
	client, err := wechatpay.NewClient(
		wechatpay.Config{
			AppId:       appId,
			MchId:       mchId,
			Apiv3Secret: apiv3Secret,
			Cert: wechatpay.CertSuite{
				SerialNo:       serialNo,
				PrivateKeyPath: privateKeyPath,
			},
		})
	if err != nil {
		log.Fatalf("unable to create client, %v", err)
	}
	payClient = client

	m := http.NewServeMux()
	m.HandleFunc("/notify", notify)

	httpServer := &http.Server{
		Handler:      m,
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
	}

	log.Printf("start up example server, listen on %s", l.Addr().String())
	err = httpServer.Serve(l)
	if err != nil {
		log.Fatalf("unable to start up http server, %v", err)
	}
}

func notify(w http.ResponseWriter, r *http.Request) {
	notification := &wechatpay.PayNotification{}

	trans, err := notification.ParseHttpRequest(payClient, r)
	if err != nil {
		answer := &wechatpay.PayNotificationAnswer{Code: "Failed", Message: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(answer.Bytes())
		return
	}

	buffer, err := json.Marshal(trans)
	if err != nil {
		answer := &wechatpay.PayNotificationAnswer{Code: "Failed", Message: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(answer.Bytes())
		return
	}
	fmt.Println("notify: ", string(buffer))

	answer := &wechatpay.PayNotificationAnswer{Code: "SUCCESS"}
	w.WriteHeader(http.StatusOK)
	w.Write(answer.Bytes())
}

