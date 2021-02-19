package main

import (
	"context"
	"flag"
	"fmt"
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
)

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

	m.HandleFunc("/", index)
	m.HandleFunc("/tradebill", tradebill)
	m.HandleFunc("/fundflowbill", fundflowbill)

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

func tradebill(w http.ResponseWriter, r *http.Request) {
	billDate := r.URL.Query().Get("billDate")
	if billDate == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(fmt.Sprintf(tradeBillHtml, "")))
	} else {
		req := wechatpay.TradeBillRequest{
			BillDate: billDate,
			BillType: wechatpay.AllBill,
			TarType:  wechatpay.GZIP,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		ctx := context.Background()
		data, err := req.Download(ctx, payClient)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte(fmt.Sprintf(tradeBillHtml, fmt.Sprintf("<br /> trade bill: %s<br />", string(data)))))
	}
}

func fundflowbill(w http.ResponseWriter, r *http.Request) {
	billDate := r.URL.Query().Get("billDate")
	if billDate == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(fmt.Sprintf(tradeBillHtml, "")))
	} else {
		req := wechatpay.FundFlowBillRequest{
			BillDate:    billDate,
			AccountType: wechatpay.BasicAccount,
			TarType:     wechatpay.GZIP,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		ctx := context.Background()
		data, err := req.Download(ctx, payClient)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		//resp, err := wechatpay.UnmarshalFundFlowBillRespone(req.AccountType, data)
		//if err != nil {
		//	w.Write([]byte(err.Error()))
		//	return
		//}

		w.Write([]byte(fmt.Sprintf(tradeBillHtml, fmt.Sprintf("<br /> fundflow bill: %s<br />", string(data)))))
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHtml))
}

var indexHtml = `
<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <title>bill</title>
</head>

<body>
    <a href="/tradebill">download tradebill</a>
    <a href="/fundflowbill">download fundflowbill</a>
</body>

</html>
`

var queryHtml = `<!DOCTYPE html>
<html>

<head>
    <meta charset="UTF-8">
    <title>bill</title>
</head>

<body>
<form action="" method="get">
<input name="transactionId" placeholder="transactionId">
<br />
  OR
<br />
<input name="outTradeNo" placeholder="outTradeNo">
<input name="submit" type="submit" value="submit">

%s

</form>
</body>

</html>`

var tradeBillHtml = `<!DOCTYPE html>
<html>

<head>
    <meta charset="UTF-8">
    <title>trade bill</title>
</head>

<body>
<form action="" method="get">
<input name="billDate" placeholder="billDate">
<br />
<input name="submit" type="submit" value="submit">

%s

</form>
</body>

</html>`
