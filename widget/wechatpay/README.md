# WechatPay
---

微信支付只能在微信内置浏览器中才能发起支付请求，请在微信内查看该组件效果。

## API

```javascript
{
  "id": "",
  "className": "",
  "theme": "",
  "options": {},
  "content": {
    "pay": {
      "timestamp": "", // 支付签名时间戳
      "nonceStr": "", // 支付签名随机串
      "package": "", // 统一支付接口返回的 prepay_id 参数值
      "signType": "", // 加密类型
      "paySign": "" // 支付签名
    },
    "btn": "" // 按钮内容
  }
}
```

## Wechat config 接口注入权限验证配置

所有需要使用JS-SDK的页面必须先注入配置信息，否则将无法调用（同一个url仅需调用一次，对于变化url的SPA的web app可在每次url变化时进行调用），具体文档在[微信支付 API](http://mp.weixin.qq.com/wiki/7/aaa137b55fb2e0456bf8dd9148dd613f.html) ，签名和其他配置参数在服务端实现逻辑。

```javascript
  wx.config({
    debug: true,
    appId: '', // 必填，公众号的唯一标识
    timestamp: '', // 必填，生成签名的时间戳
    nonceStr: '', // 必填，生成签名的随机串
    signature: '', // 必填，微信签名
    jsApiList: [
      'chooseWXPay'
    ] // 必填，需要使用的JS接口列表
  });
```