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

JavaScript [Fullscreen API](https://developer.mozilla.org/en/DOM/Using_full-screen_mode) is supported by most of the popular browsers ([List of Browsers Support Fullscreen
API](http://caniuse.com/fullscreen)). Source: [screenfull.js](https://github.com/sindresorhus/screenfull.js)。

## Methods

Called using `$.AMUI.fullscreen`.

### `.request()`

Make an element displayed on full screen mode. Use an DOM element as parameter. Default parameter is `html`.

If full screen element is a `<iframe>` element, add `allowfullscreen` attribute to it first (+ `webkitallowfullscreen` and `moz llowfullscreen`).

```html
<iframe id="frame1" allowfullscreen mozallowfullscreen webkitallowfullscreen src="iframeTest.html"></iframe>
```

Be aware of that full screen API can only be initialized through user event, such as click, touch, key etc.

### `.exit()`

Exit from the full screen mode.

### `.toggle()`

Swich between the normal mode and full screen mode.

### Events

#### Full screen status

```js
var fullscreen = $.AMUI.fullscreen;
if (fullscreen.enabled) {
  document.addEventListener(fullscreen.raw.fullscreenchange, function() {
    console.log('Am I fullscreen? ' + (fullscreen.isFullscreen ? 'Yes' : 'No'));
  });
}
```

#### Full screen error

```js
var fullscreen = $.AMUI.fullscreen;
if (fullscreen.enabled) {
  document.addEventListener(fullscreen.raw.fullscreenerror, function(e) {
    console.error('Failed to enable fullscreen', e);
  });
}
```

## Options

### `.isFullscreen`

Boolean. Return true if is in full screen mode.

### `.element`

Return the full screen element. Return `null` if there is no full screen element.

### `.enabled`

Boolean. If full screen mode is enabled. If full screen element is a `<iframe>` element, add `allowfullscreen` attribute to it first (+ `webkitallowfullscreen` and `moz llowfullscreen`).

### `.raw`

Return an object with original method name. The `key` object includes: `requestFullscreen`, `exitFullscreen`, `fullscreenElement`, `fullscreenEnabled`, `fullscreenchange` and `fullscreenerror`.

```js
$(document).on($.AMUI.fullscreen.raw.fullscreenchange, function () {
	console.log('Fullscreen change');
});
```

## Usage


### Whole Page

`````html
<button class="am-btn am-btn-primary" id="doc-fs-request">Fullscreen</button>
<button class="am-btn am-btn-secondary" id="doc-fs-exit">Exit</button>
<button class="am-btn am-btn-warning" id="doc-fs-toggle">Troggle</button>
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

### Single Element

`````html
<style>
  #demo-full-img:-webkit-full-screen {
    width: 100%;
    height: auto;
    display: block;
  }
</style>
<div>
  <img id="demo-full-img" src="http://s.cn.bing.net/az/hprichbg/rb/FortBourtange_ZH-CN9788197909_1920x1080.jpg"
       width="340"
       height="142"
       alt=""/>
  <br/>
  Click on the image to view it in full screen mode.
  <br/>
  <span id="doc-fs-img"></span>
</div>
`````
```html
<div>
  <img id="demo-full-img" src="http://s.cn.bing.net/az/hprichbg/rb/WorkingFarmer_ZH-CN9182210796_1366x768.jpg"
       width="340"
       height="142"
       alt=""/>
  <br/>
  Click on the image to view it in full screen mode.
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
  var text = 'Image status: ' + (fullscreen.isFullscreen ? 'Fullscreen' : 'Non-fullscreen');
  $('#doc-fs-img').html(text);
});
```

### Listen to the change of fullscreen status

`````html
<p>Open the console and try the example above.</p>
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
    // Listen to the fullscreen status
    var text = 'Image status: <strong>' + (fullscreen.isFullscreen ? 'Fullscreen' : 'Non-fullscreen')
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

## Reference:

- [Using the Fullscreen API in web browsers](http://hacks.mozilla.org/2012/01/using-the-fullscreen-api-in-web-browsers/)
- [MDN - Fullscreen API](https://developer.mozilla.org/en/DOM/Using_full-screen_mode)
- [W3C Fullscreen spec](http://dvcs.w3.org/hg/fullscreen/raw-file/tip/Overview.html)
- [Building an amazing fullscreen mobile experience](http://www.html5rocks.com/en/mobile/fullscreen/)
- [Can I use Full Screen API?](http://caniuse.com/fullscreen)

## License

MIT © [Sindre Sorhus](http://sindresorhus.com)
