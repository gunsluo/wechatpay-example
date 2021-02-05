# wechatpay-example

This is a demo for [wechatpay-go](https://github.com/gunsluo/wechatpay-go)

| Directory                                 | Description                                                      | 
| -----------------------------------------:| -----------------------------------------------------------------|
| [pay/native/main.go](./pay/native/main.go) | wechat native pay demo |
| [pay/jsapi/main.go](./pay/jsapi/main.go) | wechat jsapi pay demo |
| [query/main.go](./query/main.go) | query transaction |
| [refund/main.go](./refund/main.go) | refund transaction |
| [refund_query/main.go](./refund_query/main.go) | query refund transaction |
| [close/main.go](./close/main.go) | close transaction |

## Run Demo

```
cd $demoDirectory
export WECHAT_MCHID=yourmchid
export WECHAT_APPID=yourappid
export WECHAT_APIV3_SECRET=yourapiv3seret
export WECHAT_SERIAL_NO=yourserialno
export WECHAT_PRIVATE_KEY_PATH=/path/to/yourapiprivatekeypath

go run main.go -a :8080 -mchid ${WECHAT_MCHID} \
    -appid ${WECHAT_APPID} \
    -apiv3-secret ${WECHAT_APIV3_SECRET} \
    -serial-no ${WECHAT_SERIAL_NO} \
    -private-key-path ${WECHAT_PRIVATE_KEY_PATH}
```
