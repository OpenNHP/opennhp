---
id: nprogress
title: 加载进度条
titleEn: NProgress
prev: javascript/popover.html
next: javascript/slider.html
source: js/ui.progress.js
doc: docs/javascript/nprogress.md
---

# Progress
---

类似 Google、Youtube、Medium 的进度条，源自 [NProgress](http://ricostacruz.com/nprogress)。

## 基本使用
----------

调用 `start()` 、 `done()` 控制进度条。

```javascript
$.AMUI.progress.start();
$.AMUI.progress.done();
```

`````html
<button id="np-s" class="am-btn am-btn-primary">$.AMUI.progress.start();</button> <button id="np-d" class="am-btn am-btn-success">$.AMUI.progress.done();</button>
`````
```html
<button id="np-s" class="am-btn am-btn-primary">$.AMUI.progress.start();</button>
<button id="np-d" class="am-btn am-btn-success">$.AMUI.progress.done();</button>
```
```js
$(function(){
  var progress = $.AMUI.progress;

  $('#np-s').on('click', function() {
    progress.start();
  });

  $('#np-d').on('click', function() {
    progress.done();
  });
});
```

如果使用 [Turbolinks] 1.3.0+ 或者类似的库，可以在自定义事件回调中调用 Progress。

~~~ js
$(document).on('page:fetch',   function() { $.AMUI.progress.start(); });
$(document).on('page:change',  function() { $.AMUI.progress.done(); });
$(document).on('page:restore', function() { $.AMUI.progress.remove(); });
~~~

使用场景
----------

 * 在 Ajax 应用中添加进度条，绑定到 Zepto（jQuery） `ajaxStart` 和
 `ajaxStop` 事件中。

 * 没有使用 Turbolinks/Pjax 也可以添加高大上的进度条，绑定到
 `$(document).ready` 和 `$(window).load` 即可。

高级使用
--------------

**设置百分比**: 调用 `.set(n)` 可以设置进度百分比, *n* 的取值区间为 `0..1`。

~~~ js
$.AMUI.progress.set(0.0);     // Sorta same as .start()
$.AMUI.progress.set(0.4);
$.AMUI.progress.set(1.0);     // Sorta same as .done()
~~~

**增加进度**: 调用 `.inc()`，进度将会增加一个随机的数量，但不会到达 100%。

~~~ js
$.AMUI.progress.inc();
~~~

也可以给 `.inc()` 传递一个数值参数:

~~~ js
$.AMUI.progress.inc(0.2);    // This will get the current status value and adds 0.2 until status is 0.994
~~~

`.inc()` 方法会获取当前进度值并增加 0.2，但最多只到达 0.994。

**强制结束**: 传递 `true` 给 `done()`，强制显示进度条（默认情况是如果没有 *.start()*，*.done()* 执行任何操作）。

~~~ js
$.AMUI.progress.done(true);
~~~

**获取进度值**: 使用 `.status` 属性。

`````html
<button id="np-set" class="am-btn am-btn-primary">$.AMUI.progress.set(0.4);</button>
<button id="np-inc" class="am-btn am-btn-warning">$.AMUI.progress.inc();</button>
<button id="np-fd" class="am-btn am-btn-success">$.AMUI.progress.done(true);</button>
<button id="np-status" class="am-btn am-btn-danger">$.AMUI.progress.status;</button>
`````
```html
<button id="np-set" class="am-btn am-btn-primary">$.AMUI.progress.set(0.4);</button>
<button id="np-inc" class="am-btn am-btn-warning">$.AMUI.progress.inc();</button>
<button id="np-fd" class="am-btn am-btn-success">$.AMUI.progress.done(true);</button>
<button id="np-status" class="am-btn am-btn-danger">$.AMUI.progress.status;</button>
```
```js
$(function(){
  var progress = $.AMUI.progress;

  $('#np-set').on('click', function() {
    progress.set(0.4);
  });

  $('#np-inc').on('click', function() {
    progress.inc();
  });

  $('#np-fd').on('click', function() {
    progress.done(true);
  });

  $('#np-status').on('click', function() {
    $(this).text('Status: ' + progress.status);
  });
});
```

参数设置
-------

### 默认参数

```js
  {
    minimum: 0.08,
    easing: 'ease',
    positionUsing: '',
    speed: 200,
    trickle: true,
    trickleRate: 0.02,
    trickleSpeed: 800,
    showSpinner: true,
    barSelector: '[role="nprogress-bar"]',
    spinnerSelector: '[role="nprogress-spinner"]',
    parent: 'body',
    template: '<div class="nprogress-bar" role="nprogress-bar">' +
        '<div class="nprogress-peg"></div></div>' +
        '<div class="nprogress-spinner" role="nprogress-spinner">' +
        '<div class="nprogress-spinner-icon"></div></div>'
  }
```

### 参数解释
`minimum`: 设置最小百分比。

~~~ js
$.AMUI.progress.configure({ minimum: 0.1 });
~~~

`template`: 设置模板，注意相应更改 `barSelector`、`spinnerSelector`。

~~~ js
$.AMUI.progress.configure({
  template: "<div class='....'>...</div>"
});
~~~

`ease`、`speed`: 设置动画缓动函数和速度（`ms`）

~~~ js
$.AMUI.progress.configure({ ease: 'ease', speed: 500 });
~~~

`trickle`、`trickleRate`、`trickleSpeed`:

~~~ js
$.AMUI.progress.configure({ trickle: false });
~~~

~~~ js
$.AMUI.progress.configure({ trickleRate: 0.02, trickleSpeed: 800 });
~~~

`showSpinner`:

~~~ js
NProgress.configure({ showSpinner: false });
~~~

`parent`:

设置插入进度条的父容器，默认为 `body`。

```js
$.AMUI.progress.configure({ parent: '#container' });
```

自定义
-----

通过更改 css 改变进度条样式，对应的样式为 `less` 目录下面的 `ui.progress.less`。


参考链接
-------

 * [New UI Pattern: Website Loading
 Bars](http://www.usabilitypost.com/2013/08/19/new-ui-pattern-website-loading-bars/)


[Turbolinks]: https://github.com/rails/turbolinks
[nprogress.js]: http://ricostacruz.com/nprogress/nprogress.js
[nprogress.css]: http://ricostacruz.com/nprogress/nprogress.css


<script>
$(function(){
  var progress = $.AMUI.progress;

  $('#np-s').on('click', function() {
    progress.start();
  });

  $('#np-d').on('click', function() {
    progress.done();
  });

  $('#np-set').on('click', function() {
    progress.set(0.4);
  });

  $('#np-inc').on('click', function() {
    progress.inc();
  });

  $('#np-fd').on('click', function() {
    progress.done(true);
  });

  $('#np-status').on('click', function() {
    $(this).text('Status: ' + progress.status);
  });
});
</script>


