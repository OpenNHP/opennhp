---
id: fullscreen
title: 全屏切换
titleEn: Fullscreen
prev: javascript/cookie.html
next: javascript/hammer.html
source: js/util.fullscreen.js
doc: docs/javascript/fullscreen.md
---

# Fullscreen
---

JavaScript [Fullscreen API](https://developer.mozilla.org/en/DOM/Using_full-screen_mode) 跨浏览器兼容封装（[Fullscreen
API 兼容性列表](http://caniuse.com/fullscreen)），免去苦逼写各种浏览器前缀的麻烦。源自 [screenfull.js](https://github.com/sindresorhus/screenfull.js)。

## 方法

以下方法通过 `$.AMUI.fullscreen` 接口调用。

### `.request()`

使元素全屏。接受一个 DOM 元素作为参数，默认为 `html`。

`<iframe>` 中需要添加 `allowfullscreen` 属性 (+ `webkitallowfullscreen` and `moz
  allowfullscreen`)。

```html
<iframe id="frame1" allowfullscreen mozallowfullscreen webkitallowfullscreen src="iframeTest.html"></iframe>
```

注意全屏 API 仅能通过用户事件（如 click、touch、key） 初始化。

### `.exit()`

退出全屏模式。

### `.toggle()`

全屏模式切换。

### 事件监听

#### 全屏状态变化

```js
var fullscreen = $.AMUI.fullscreen;
if (fullscreen.enabled) {
  document.addEventListener(fullscreen.raw.fullscreenchange, function() {
    console.log('Am I fullscreen? ' + (fullscreen.isFullscreen ? 'Yes' : 'No'));
  });
}
```

#### 监听全屏错误

```js
var fullscreen = $.AMUI.fullscreen;
if (fullscreen.enabled) {
  document.addEventListener(fullscreen.raw.fullscreenerror, function(e) {
    console.error('Failed to enable fullscreen', e);
  });
}
```

## 属性

### `.isFullscreen`

布尔值，是否处于全屏模式。

### `.element`

返回当前处于全屏模式的元素，没有则返回 `null`。

### `.enabled`

是否允许全屏模式。`<iframe>` 中的页面需要添加 `allowfullscreen` 属性 (+ `webkitallowfullscreen` and `moz
    allowfullscreen`)。

### `.raw`

返回包含原始方法名称的对象，对象 `key` 包括： `requestFullscreen`, `exitFullscreen`, `fullscreenElement`, `fullscreenEnabled`, `fullscreenchange`,`fullscreenerror`

```js
$(document).on($.AMUI.fullscreen.raw.fullscreenchange, function () {
	console.log('Fullscreen change');
});
```

## 使用示例


### 全屏整个页面

`````html
<button class="am-btn am-btn-primary" id="doc-fs-request">全屏窗口</button>
<button class="am-btn am-btn-secondary" id="doc-fs-exit">退出全屏</button>
<button class="am-btn am-btn-warning" id="doc-fs-toggle">全屏交替</button>
`````

```js
$('#demo-full-page').on('click', function () {
	if ($.AMUI.fullscreen.enabled) {
    $.AMUI.fullscreen.request();
	} else {
		// Ignore or do something else
	}
});
```

### 全屏显示元素

`````html
<style>
  #demo-full-img:-webkit-full-screen {
    width: 100%;
    height: auto;
    display: block;
  }
</style>
<div>
  <img id="demo-full-img" src="http://s.amazeui.org/media/i/demos/bing-2.jpg"
       width="340"
       height="142"
       alt=""/>
  <br/>
  点击图片全屏显示
  <br/>
  <span id="doc-fs-img"></span>
</div>
`````
```html
<div>
  <img id="demo-full-img" src="http://s.amazeui.org/media/i/demos/bing-2.jpg"
       width="340"
       height="142"
       alt=""/>
  <br/>
  点击图片全屏显示
  <br/>
  <span id="doc-fs-img"></span>
</div>
```

```js
var fullscreen = $.AMUI.fullscreen;

$('#demo-full-img').on('click', function () {
  if (fullscreen.enabled) {
    fullscreen.request(this);
  }
}).on(fullscreen.raw.fullscreenchange, function () {
  // 监听图片全屏状态
  var text = '图片状态：' + (fullscreen.isFullscreen ? '全屏' : '非全屏');
  $('#doc-fs-img').html(text);
});
```

### 监听全屏状态改变

`````html
<p>打开控制台，点击上面的演示看看</p>
<script>
$(function() {
  var fullscreen = $.AMUI.fullscreen;

  // demo1
  $('#doc-fs-request').on('click', function () {
    fullscreen.enabled && fullscreen.request();
  });

  $('#doc-fs-exit').on('click', function () {
    fullscreen.enabled && fullscreen.exit();
  });

  $('#doc-fs-toggle').on('click', function () {
    fullscreen.enabled && fullscreen.toggle();
  });

  // demo2
  $('#demo-full-img').on('click', function () {
    if (fullscreen.enabled) {
      fullscreen.request(this);
    }
  }).on(fullscreen.raw.fullscreenchange, function () {
    // 监听图片全屏状态
    var text = '图片状态：<strong>' + (fullscreen.isFullscreen ? '全屏' : '非全屏')
      + '</strong>';
    $('#doc-fs-img').html(text);
  });

  // demo3
  if (fullscreen.enabled) {
    $(document).on(fullscreen.raw.fullscreenchange, function () {
      console.log('Am I fullscreen? ' + (fullscreen.isFullscreen ? 'Yes' : 'No'));
    });
  }
});
</script>
`````

```js
if (fullscreen.enabled) {
	$(document).on($.AMUI.fullscreen.raw.fullscreenchange, function () {
		console.log('Am I fullscreen? ' + ($.AMUI.fullscreen.isFullscreen ? 'Yes' : 'No'));
	});
}
```

## 参考资源

- [Using the Fullscreen API in web browsers](http://hacks.mozilla.org/2012/01/using-the-fullscreen-api-in-web-browsers/)
- [MDN - Fullscreen API](https://developer.mozilla.org/en/DOM/Using_full-screen_mode)
- [W3C Fullscreen spec](http://dvcs.w3.org/hg/fullscreen/raw-file/tip/Overview.html)
- [Building an amazing fullscreen mobile experience](http://www.html5rocks.com/en/mobile/fullscreen/)
- [Can I use Full Screen API?](http://caniuse.com/fullscreen)

## License

MIT © [Sindre Sorhus](http://sindresorhus.com)
