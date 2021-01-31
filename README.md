# wechatpay-example

This is a demo for [wechatpay-go](https://github.com/gunsluo/wechatpay-go)

## Run Demo

```
go run main.go -a :8080 -mchid yourmchid -appid yourappid -apiv3-secret yourapiv3seret -serial-no yourserialno -private-key-path yourapiprivatekeypath
```

then open browser:
[http://localhost:8080](http://localhost:8080)

## Notice

* Please change NotifyURL, or you will not get wechat async notify
