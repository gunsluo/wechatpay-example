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
	"fmt"
	"log"

	"github.com/gunsluo/wechatpay-go/v3"
)

var (
	appId          string
	apiv3Secret    string
	serialNo       string
	mchId          string
	privateKeyPath string
	payClient      wechatpay.Client
)

func main() {
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
	req := wechatpay.RefundQueryRequest{
		OutRefundNo: "S20210219165335728986",
	}
	resp, err := req.Do(context.Background(), client)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(data))
}
