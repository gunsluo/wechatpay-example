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
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gunsluo/wechatpay-go/v3"
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
	data, err := ioutil.ReadFile("templates/close.html")
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
	outTradeNo := r.URL.Query().Get("out_trade_no")
	w.Header().Set("Content-Type", "text/html")
	if outTradeNo == "" {
		tpl.Execute(w, nil)
	} else {
		req := wechatpay.CloseRequest{
			OutTradeNo: outTradeNo,
		}
		w.Header().Set("Content-Type", "text/html")
		err := req.Do(context.Background(), payClient)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		tpl.Execute(w, "ok")
	}
}
