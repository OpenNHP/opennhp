---
id: qrcode
title: 二维码生成插件
titleEn: QRCode
prev: javascript/geolocation.html
source: js/util.qrcode.js
doc: docs/javascript/qrcode.md
---

# QRCode
---

二维码生成工具（[via](https://github.com/aralejs/qrcode)，JS 相关的二维码生成工具大多基于 [QR Code Generator 项目](https://github.com/kazuhikoarase/qrcode-generator)）。

## 演示

`````html
<div class="am-input-group">
  <input type="text" class="am-form-field" id="doc-qr-text">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button" id="doc-gen-qr">生成</button>
      </span>
</div>
<div id="doc-qrcode" class="am-text-center"></div>

<script>
  $(function() {
    var $input = $('#doc-qr-text');
    $qr = $('#doc-qrcode');

    function makeCode(text) {
      $qr.empty().qrcode(text);
    }

    $input.val(location.href);
    makeCode(location.href);

    $('#doc-gen-qr').on('click', function() {
      makeCode($input.val());
    });

    $input.on('focusout', function() {
      makeCode($input.val());
    });
  });
</script>
`````
```html
<div class="am-input-group">
  <input type="text" class="am-form-field" id="doc-qr-text">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button" id="doc-gen-qr">生成</button>
      </span>
</div>
<div id="doc-qrcode" class="am-text-center"></div>

<script>
  $(function() {
    var $input = $('#doc-qr-text');
    $qr = $('#doc-qrcode');

    function makeCode(text) {
      $qr.empty().qrcode(text);
    }

    $input.val(location.href);
    makeCode(location.href);

    $('#doc-gen-qr').on('click', function() {
      makeCode($input.val());
    });

    $input.on('focusout', function() {
      makeCode($input.val());
    });
  });
</script>
```

## API

- `$(element).qrcode(options)`: 根据指定 `options` 生成二维码并插入到 `$(element)` 中 <span
  class="am-badge am-badge-danger">v2.4.1</span>；
- `v2.4.1` 以前的版本可以通过 `$.AMUI.qrcode` 调用构造函数。

  ```js
  var QRCode = $.AMUI.qrcode;
  $(element).html(new QRCode({text: 'xxx'}));
  ```

如果直接传递字符串，按照默认参数生成该字符串的二维码。

默认参数及说明：

```js
{
  text: "", // 要生产二维码的文字
  render: "", // 渲染方式，默认的选择顺序为 `canvas` -> `svg` -> `table`
  width: 256, // 二维码宽度 `px`
  height: 256, // 二维码高度 `px`
  correctLevel: 3, // 纠错级别，可取 0、1、2、3，数字越大说明所需纠错级别越大
  background: "#ffffff", // 背景色
  foreground: "#000000" // 前景色
}
```
