# FastClick
---

处理移动端 `click` 事件 300 毫秒延迟， 由 [FT Labs](http://labs.ft.com/) 开发，Github 项目地址：https://github.com/ftlabs/fastclick 。

## 为什么存在延迟？

根据 [Google 开发者文档](https://developers.google.com/mobile/articles/fast_buttons)：

> ...mobile browsers will wait approximately 300ms from the time that you tap the button to fire the click event. The reason for this is that the browser is waiting to see if you are actually performing a double tap.

从点击屏幕上的元素到触发元素的 `click` 事件，移动浏览器会有大约 300 毫秒的等待时间。为什么这么设计呢？
因为它想看看你是不是要进行双击（double tap）操作。

## 兼容性

* Mobile Safari on iOS 3 and upwards
* Chrome on iOS 5 and upwards
* Chrome on Android (ICS)
* Opera Mobile 11.5 and upwards
* Android Browser since Android 2
* PlayBook OS 1 and upwards

## 不应用 FastClick 的场景

* 桌面浏览器；
* 如果 [viewport meta 标签](https://developer.mozilla.org/en-US/docs/Mobile/Viewport_meta_tag) 中设置了 `width=device-width`， Android 上的 Chrome 32+ 会禁用 300ms 延时；

```html
<meta name="viewport" content="width=device-width, initial-scale=1">
```

* viewport meta 标签如果设置了 `user-scalable=no`，Android 上的 Chrome（所有版本）都会禁用 300ms 延迟。
* IE10 中，可以使用 css 属性 ` -ms-touch-action: none` 禁止元素双击缩放（[参考文章](http://blogs.msdn.com/b/askie/archive/2013/01/06/how-to-implement-the-ms-touch-action-none-property-to-disable-double-tap-zoom-on-touch-devices.aspx)）。

## 使用方法

__TODO: 修改使用接口__

```js
window.addEventListener('load', function() {
  FastClick.attach(document.body);
}, false);
```

Zepto.js:

```js    
$(function() {
  FastClick.attach(document.body);
});
```

```javascript
var attachFastClick = require('fastclick');
attachFastClick(document.body);
```

<!--
AMD
var FastClick = require('fastclick');
FastClick.attach(document.body);
-->
    

## Licence

* @copyright The Financial Times Limited [All Rights Reserved]
* @license MIT License