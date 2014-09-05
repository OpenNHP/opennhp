# Fullscreen
---

JavaScript [Fullscreen API](https://developer.mozilla.org/en/DOM/Using_full-screen_mode) 跨浏览器兼容封装（[点击查看 Fullscreen
API 兼容性列表](http://caniuse.com/fullscreen)），免去苦逼写各种浏览器前缀的麻烦。

## 方法

一下方法通过 `$.AMUI.fullscreen` 接口调用。

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

### `.onchange()`

<del>全屏模式发生改变时获得通知。</del>

建议使用下面的事件监听：

```js
$(document).on($.AMUI.fullscreen.raw.fullscreenchange, function () {});
```

#### .onerror()

<del>全屏模式发生错误时获得通知。</del>

建议使用下面的事件监听：

```js
$(document).on($.AMUI.fullscreen.raw.fullscreenerror, function () {});
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

返回包含原始方法名称的对象，对象 `key` 包括： `requestFullscreen`, `exitFullscreen`, `fullscreenElement`, `fullscreenEnabled`, `
    fullscreenchange`,
    `fullscreenerror`

```js
$(document).on($.AMUI.fullscreen.raw.fullscreenchange, function () {
	console.log('Fullscreen change');
});
```

## 使用示例


### 全屏整个页面

`````html
<button class="am-btn am-btn-primary" id="demo-full-page">Fullscreen the page</button>
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
  <div>
    <img id="demo-full-img" src="http://s.cn.bing.net/az/hprichbg/rb/WorkingFarmer_ZH-CN9182210796_1366x768.jpg"
         width="340"
         height="142"
         alt=""/>
    <br/>
    点击图片全屏显示
  </div>
`````

```js
$('#demo-full-img').on('click', function () {
	if ($.AMUI.fullscreen.enabled) {
    $.AMUI.fullscreen.request(this);
	}
});
```

### 监听全屏状态改变

`````html
    <p>打开控制台，点击上面的演示看看</p>
<script>
$(function() {
  var fullscreen = $.AMUI.fullscreen;

  // demo1
  $('#demo-full-page').on('click', function () {
    if (fullscreen.enabled) {
      fullscreen.request();
    } else {
      // Ignore or do something else
    }
  });

  // demo2
  $('#demo-full-img').on('click', function () {
    if (fullscreen.enabled) {
      fullscreen.request(this);
    }
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
