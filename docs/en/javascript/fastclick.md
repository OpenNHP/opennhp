---
id: fastclick
title: 移动端 300 毫秒延迟处理
titleEn: FastClick
prev: javascript/cookie.html
next: javascript/fullscreen.html
source: js/util.fastclick.js
doc: docs/javascript/fastclick.md
---

# FastClick
---

**Important Update in `v2.4`:**

**Because of some bug caused by FastClick. It is removed from Amaze UI in `v2.4`. Developers can get it from [FastClick Offical Website](https://github.com/ftlabs/fastclick).**

This component is developed by [FT Labs](http://labs.ft.com/) in order to solve the problem of `click` event's 300 ms delay on mobile devices.

## Why delay?

According to the [Google Developers Docs](https://developers.google.com/mobile/articles/fast_buttons)：

> ...mobile browsers will wait approximately 300ms from the time that you tap the button to fire the click event. The reason for this is that the browser is waiting to see if you are actually performing a double tap.

## Supported in

* Mobile Safari on iOS 3 and upwards
* Chrome on iOS 5 and upwards
* Chrome on Android (ICS)
* Opera Mobile 11.5 and upwards
* Android Browser since Android 2
* PlayBook OS 1 and upwards

## Don't Need FastClick

* In Desktop browsers;
* If `width=device-width` is set in [viewport meta tag](https://developer.mozilla.org/en-US/docs/Mobile/Viewport_meta_tag), which will disable the touch delay in Android Chrome 32+;

```html
<meta name="viewport" content="width=device-width, initial-scale=1">
```

* If `user-scalable=no` is set in [viewport meta tag](https://developer.mozilla.org/en-US/docs/Mobile/Viewport_meta_tag), which will disable the touch delay in Android Chrome 32+;
* In IE10, double click zoom can be disabled by using css attribute ` -ms-touch-action: none` ([Reference](http://blogs.msdn.com/b/askie/archive/2013/01/06/how-to-implement-the-ms-touch-action-none-property-to-disable-double-tap-zoom-on-touch-devices.aspx)).

## Usage

__FastClick is initialized in Amaze UI.__

Fask Click can be called using `$.AMUI.FastClick`.

jQuery / Zepto.js:

```js
$(function() {
  $.AMUI.FastClick.attach(document.body);
});
```

CommonJS:

```javascript
var attachFastClick = require('fastclick');
attachFastClick(document.body);
```

## FAQ

### Q: How to solve the problem of `contenteditable` getting disabled when using FastClick?

A: Add the `.needsclick` class to the element whose edit functionality needs to be activated ([Example](http://jsbin.com/wahilo/3/)).


`````html
<div contenteditable class="needsclick">
  <p>Try to edit this on touch screen</p>
  <p>Edit Me</p>
</div>
<hr>
<div contenteditable>
  Can't edit without <code>needsclick</code> class.
</div>
`````

```html
<div contenteditable class="needsclick">
  <p>Try to edit this on touch screen</p>
  <p>Edit Me</p>
</div>
<hr>
<div contenteditable>
  Can't edit without <code>needsclick</code> class.
</div>
```

## Licence

* @copyright The Financial Times Limited [All Rights Reserved]
* @license MIT License
