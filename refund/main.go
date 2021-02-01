// Copyright The Wechat Pay Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/gunsluo/wechatpay-go/v3"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
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
	data, err := ioutil.ReadFile("templates/refund.html")
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

	client, err := wechatpay.NewClient(
		wechatpay.Config{
			AppId:       appId,
			MchId:       mchId,
			Apiv3Secret: apiv3Secret,
			Cert: wechatpay.CertSuite{
				SerialNo:       serialNo,
				PrivateKeyPath: privateKeyPath,
			},
		},
	)

	if err != nil {
		log.Fatalf("unable to create client, %v", err)
	}

	payClient = client

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("unable to listen on %s, %v", address, err)
	}
	m := http.NewServeMux()
	m.HandleFunc("/", index)
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

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	outTradeNo := r.URL.Query().Get("out_trade_no")
	if outTradeNo == "" {
		err := tpl.Execute(w, nil)
		if err != nil {
			serveError(w, err)
			return
		}
		return
	}

	refund := r.URL.Query().Get("refund")
	total := r.URL.Query().Get("total")
	transactionId := r.URL.Query().Get("transaction_id")

	ir, err := strconv.ParseInt(refund, 10, 64)
	if err != nil {
		serveError(w, err)
		return
	}
	it, err := strconv.ParseInt(total, 10, 64)
	if err != nil {
		serveError(w, err)
		return
	}
	var refundRequest = wechatpay.RefundRequest{
		TransactionId: transactionId,
		OutTradeNo:    outTradeNo,
		OutRefundNo:   NewTradeNo(),
		Reason:        "test refund",
		NotifyUrl:     "http://ip.clearcode.cn/notify",
		Amount: wechatpay.RefundAmount{
			Refund:   int(ir),
			Total:    int(it),
			Currency: "CNY",
		},
	}

	resp, err := refundRequest.Do(context.Background(), payClient)
	if err != nil {
		serveError(w, err)
		return
	}
	data, _ := json.MarshalIndent(resp, "", "\t")
	tpl.Execute(w, template.HTML(data))
}

// NewTradeNo new a trade no
func NewTradeNo() string {
	var now = time.Now()
	ms := now.Format(".000000")
	return "S" + now.Format("20060102150405") + ms[1:]
}

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), 200)
}
