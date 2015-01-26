# WechatPay
---

微信 JS SDK 应用示例。**微信支付 99% 是后端的工作，如果项目中需要微信支付，请把重心放在后端。**

**微信 JS SDK 脚本及权限验证配置请从后端返回**。相关支持请联系微信官方或参考[官方文档](http://mp.weixin.qq.com/wiki/7/aaa137b55fb2e0456bf8dd9148dd613f.html)。

通过 JS SDK 发起微信支付请求只能在微信中才能执行，请扫描二维码在微信内查看效果。

<div id="doc-wechat-pay-qr"></div>

---

## API

```javascript
{
  "id": "",
  "className": "",
  "theme": "",
  "options": {},
  "content": {
    // 订单信息及支付签名
    "order": {
      "timestamp": "", // 支付签名时间戳
      "nonceStr": "", // 支付签名随机串
      "package": "", // 统一支付接口返回的 prepay_id 参数值
      "signType": "", // 加密类型
      "paySign": "" // 支付签名
    },
    "title": "" // 按钮文字
  }
}
```

## 微信 JS SDK 调用支付接口要点

### JS SDK 权限验证

所有需要使用JS-SDK的页面必须先注入配置信息，否则将无法调用（同一个url仅需调用一次，对于变化 URL 的 SPA 的 Web App 可在每次 URL 变化时进行调用），具体文档在[微信支付 API](http://mp.weixin.qq.com/wiki/7/aaa137b55fb2e0456bf8dd9148dd613f.html) ，签名和其他配置参数在服务端实现逻辑。

详细步骤见[微信官网文档](http://mp.weixin.qq.com/wiki/7/aaa137b55fb2e0456bf8dd9148dd613f.html#JSSDK.E4.BD.BF.E7.94.A8.E6.AD.A5.E9.AA.A4)。

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

### 使用微信支付统一下单接口生成订单信息

详见[微信支付统一下单接口文档](http://pay.weixin.qq.com/wiki/doc/api/index.php?chapter=9_1)

### 生成支付签名

详见[微信支付签名算法](http://pay.weixin.qq.com/wiki/doc/api/index.php?chapter=4_3)

### 发起支付请求

```
wx.chooseWXPay({
  timestamp: 0, // 支付签名时间戳，注意微信jssdk中的所有使用timestamp字段均为小写。但最新版的支付后台生成签名使用的timeStamp字段名需大写其中的S字符
  nonceStr: '', // 支付签名随机串，不长于 32 位
  package: '', // 统一支付接口返回的prepay_id参数值，提交格式如：prepay_id=***）
  signType: 'MD5', // 签名方式，默认为'SHA1'，使用新版支付需传入'MD5'
  paySign: '', // 支付签名
});
```

<script>
  (function($) {
    var QRCode = $.AMUI.qrcode;

    $(function() {
      var qrnode = new QRCode({
        render: 'canvas',
        correctLevel: 0,
        text: 'http://amazeui.org/widgets/wechatpay',
        width: 200,
        height: 200,
        background: '#fff',
        foreground: '#000'
      });

      $('#doc-wechat-pay-qr').html(qrnode);
    });
  })(window.jQuery);
</script>
