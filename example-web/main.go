package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gunsluo/wechatpay-go/v3"
	"github.com/skip2/go-qrcode"
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
	m.HandleFunc("/pay", payment)
	m.HandleFunc("/qr", qrCode)
	m.HandleFunc("/notify", notify)
	m.HandleFunc("/query", query)

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

func payment(w http.ResponseWriter, r *http.Request) {
	amount := 0.01
	tradeNo := NewTradeNo()

	req := &wechatpay.PayRequest{
		AppId:       appId,
		MchId:       mchId,
		Description: "for testing",
		OutTradeNo:  tradeNo,
		TimeExpire:  time.Now().Add(10 * time.Minute),
		Attach:      "cipher code",
		NotifyUrl:   "http://ip.clearcode.cn/notify",
		Amount: wechatpay.PayAmount{
			Total:    int(amount * 100),
			Currency: "CNY",
		},
		TradeType: wechatpay.Native,
	}

	resp, err := req.Do(r.Context(), payClient)
	if err != nil {
		e := &wechatpay.Error{}
		if errors.As(err, &e) {
			fmt.Println("status", e.Status, "code:", e.Code, "message:", e.Message)
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed payment: " + err.Error()))
		return
	}
	codeUrl := resp.CodeUrl

	payTpl.Execute(w, map[string]interface{}{
		"amount":  amount,
		"tradeNo": tradeNo,
		"codeUrl": codeUrl,
		"method":  r.URL.Query().Get("method"),
	})
}

func qrCode(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		w.Write(nil)
		return
	}

	img, _ := qrcode.Encode(url, qrcode.High, 180)
	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
}

func query(w http.ResponseWriter, r *http.Request) {
	outTradeNo := r.URL.Query().Get("outTradeNo")
	transactionId := r.URL.Query().Get("transactionId")
	if outTradeNo == "" && transactionId == "" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fmt.Sprintf(queryHtml, "")))
	} else {
		req := wechatpay.QueryRequest{
			MchId:         mchId,
			OutTradeNo:    outTradeNo,
			TransactionId: transactionId,
		}
		w.Header().Set("Content-Type", "text/html")

		resp, err := req.Do(context.Background(), payClient)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(fmt.Sprintf(queryHtml, fmt.Sprintf("<br /> detail=====> <br />%+v", resp))))
	}
}

// NewTradeNo new a trade no
func NewTradeNo() string {
	now := time.Now()
	ms := now.Format(".000000")
	return "S" + now.Format("20060102150405") + ms[1:]
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHtml))
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

var indexHtml = `
<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <title>payment</title>
</head>

<body>
    <a href="/pay">send transaction</a>
    <a href="/query">query transaction</a>
</body>

</html>
`

var queryHtml = `<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <title>payment</title>
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

var payHtml = `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>PAY</title>
</head>
<body>
<h3>tradeNo: {{ .tradeNo }}</h3>
<img src="/qr?url={{ .codeUrl }}" alt="">
<p>{{ .amount }}</p>
<p id="status">PAYING</p>

<a href="/pay">send again</a>
<script>
    var tradeNo = "{{ .tradeNo }}";
    var interval = setInterval(function (){
        fetch("/status?tradeno=" + tradeNo)
			.then(response => response.json())
			.then(function(data){
				if(data.status == 2) {
					document.getElementById("status").innerText = "PAYED"; 
					clearInterval(interval);
				} 
			})
			
    },60000)
</script>
</body>
</html>`

var payTpl, _ = template.New("").Parse(payHtml)
