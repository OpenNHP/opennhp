# WechatPay
---

WeChat JS SDK Example. **99% of a WechatPay project is on backend. If you want to use WechatPay in your project, please pay more attention to the backend.**

**Please return the WeChat JS SDK and the authentication configuration from the backend**. See[WeChat official Docs](http://mp.weixin.qq.com/wiki/7/aaa137b55fb2e0456bf8dd9148dd613f.html) for more imformation.

The WechayPay requrest making through JS SDK is only avaliable in WeChat. Please scan the QR code in Wechat to see the example.

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
    // Order imformation and signature
    "order": {
      "timestamp": "", // Signature timestamp.
      "nonceStr": "", // Signature random string.
      "package": "", // The prepay_id returned by the unified order API
      "signType": "", // Encryption method
      "paySign": "" // Signature
    },
    "title": "" // Texts on button
  }
}
```

## Something Important

### JS SDK Authentication

All the pages need JS-SDK should have configuration injected. Otherwise, JS-SDK will not be avaliable. See more details in [WechatPay API](http://mp.weixin.qq.com/wiki/7/aaa137b55fb2e0456bf8dd9148dd613f.html). Signature and other configurations can be made in server-side, and the details can be found in [Wechat Offical Docs](http://mp.weixin.qq.com/wiki/7/aaa137b55fb2e0456bf8dd9148dd613f.html#JSSDK.E4.BD.BF.E7.94.A8.E6.AD.A5.E9.AA.A4).

```javascript
  wx.config({
    debug: true,
    appId: '', // Required. ID of offical account.
    timestamp: '', // Signature timestamp.
    nonceStr: '', // Signature random train.
    signature: '' // Signature
    jsApiList: [
      'chooseWXPay'
    ] // Required. The required JS API list.
  });
```

### Create order with WechatPay Unified Order API

See more details in [Official Docs](http://pay.weixin.qq.com/wiki/doc/api/index.php?chapter=9_1)

### Generate Signature

See more details in [Wechat signature algorithm](http://pay.weixin.qq.com/wiki/doc/api/index.php?chapter=4_3)

### Send pay request

```
wx.chooseWXPay({
  timestamp: 0, // Time stamp. Though "timestamp" in jssdk are all in lowercase, the "S" in the "timeStamp" used to create signature is capitalized.
  nonceStr: '', // The random string. At most 32 bits.
  package: '', // The prepay_id returned by the unified order API.
  signType: 'MD5', // Encryption method. Default method is 'SHA1'. The latest version require to use 'MD5'.
  paySign: '', // Signature.
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
