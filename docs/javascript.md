---
id: javaScript
title: JS 插件
titleEn: JavaScript
permalink: javaScript.html
next: javascript/alert.html
---

# JavaScript
---

## 基本使用

### 基于 jQuery

从 2.0 开始，Amaze UI JavaScript 组件转向基于 [jQuery](http://jquery.com/) 开发，使用时确保在 Amaze UI 的脚本之前引入了 jQuery 最新正式版。

### 组件调用

组件的调用方式普通 jQuery 插件一样，具体细节请查看各个组件的文档。

### jQuery 和 Zepto.js 的一些差异

jQuery 和 Zepto.js 表面看起来差不多，其实一些细节上差异很大，同时支持 jQuery 和 Zepto.js 是一件吃力不讨好的事情，这应该也是 [Foundation 5 放弃支持 Zepto](http://zurb.com/article/1293/why-we-dropped-zepto) 的一个原因。（[下面列举的差异 Demo](http://jsbin.com/noxuvi/1/edit?html,css,js,console)）

#### `width()`/`height()`

- Zepto.js: 由盒模型（`box-sizing`）决定
- jQuery: 忽略盒模型，始终返回内容区域的宽/高（不包含 `padding`、`border`）

jQuery [官方的说明](http://api.jquery.com/width/#width)：

> Note that `.width()` will always return the content width, regardless of the value of the CSS `box-sizing` property. As of jQuery 1.8, this may require retrieving the CSS width plus `box-sizing` property and then subtracting any potential border and padding on each element when the element has `box-sizing: border-box`. To avoid this penalty, use `.css("width")` rather than `.width()`.

解决方式就是在 jQuery 中使用 `.css('width')`，而不是 `.width()`。

**这点上 jQuery 的处理方式是值得商榷的**，比如下面的例子，`$('.box').css('height')` 仍然返回 `20px`，**这不是扯蛋么**：

```html
<style>
  .box {
    box-sizing: border-box;
    padding: 10px;
    height: 0;
  }
</style>

<div class="box"></div>
```

##### 边框三角形宽高的获取

假设用下面的 HTML 和 CSS 画了一个小三角形：

```html
<div class="caret"></div>
```

```css
.caret {
  width: 0;
  height: 0;
  border-width: 0 20px 20px;
  border-color: transparent transparent blue;
  border-style: none dotted solid;
}
```

- jQuery 使用 `.width()` 和 `.css('width')` 都返回 `0`，高度也一样；
- Zepto 使用 `.width()` 返回 `40`，使用 `.css('width')` 返回 `0px`。

所以，这种场景，**jQuery 使用 `.outerWidth()`/`.outerHeight()`；Zepto 使用 `.width()`/`.height()`**。

#### `offset()`

- Zepto.js: 返回 `top`、`left`、`width`、`height`
- jQuery: 返回 `width`、`height`

#### `$(htmlString, attributes)`

- [jQuery 文档](http://api.jquery.com/jQuery/#jQuery-html-attributes)
- [Zepto 文档](http://zeptojs.com/#$())

##### DOM 操作区别

```js
$(function() {
  var $list = $('<ul><li>jQuery 插入</li></ul>', {
    id: 'insert-by-jquery'
  });
  $list.appendTo($('body'));
});
```
jQuery 操作 `ul` 上的 `id` 不会被添加；Zepto 可以在 `ul` 上添加 `id`。

##### 事件触发区别

```js
$script = $('<script />', {
  src: 'http://cdn.amazeui.org/amazeui/1.0.1/js/amazeui.min.js',
  id: 'ui-jquery'
});

$script.appendTo($('body'));

$script.on('load', function() {
  console.log('jQ script loaded');
});
```

使用 jQuery 时 `load` 事件的处理函数**不会**执行；使用 Zepto 时 `load` 事件的处理函数**会**执行。

**其他参考链接：**

- [jQuery 市场份额](http://w3techs.com/technologies/overview/javascript_library/all)
- [jQuery vs Zepto 性能比较](http://jsperf.com/zepto-vs-jquery-2013/82)

## 高级使用

### 默认初始化事件接口

Amaze UI 通过特定的 HTML 来绑定事件，多数 JS 组件通过 HTML 标记就可以实现调用。

默认的初始化事件都在 `xx.amui.data-api` 命名空间下，用户可以自行关闭。

关闭所有默认事件：

```javascript
$(document).off('.data-api');
```

关闭特定组件的默认事件：

```javascript
$(document).off('.modal.amui.data-api');
```

### 自定义事件

一些组件提供了自定义事件，命名的方式为 `{事件名称}.{组件名称}.amui`，用户可以查看组件文档了解、使用这些事件。

```javascript
$('#myAlert').on('close.alert.amui', function() {
  // do something
});
```

### MutationObserver

双向数据绑定很酷？[Mutation Observer](http://www.w3.org/TR/dom/#mutation-observers) 才是（或即将成为）幕后的英雄。

Amaze UI 2.1 中实验性地引入了 `MutationObserver`，**请谨慎使用**。

#### `data-am-observe`

在元素上添加 `data-am-observe` 属性以后，动态插入该元素的 Amaze UI JS 插件会自动初始化（[演示](/javascript/scrollspy#mutationobserver?_ver=2.x)），
  支持的插件包括 Button、Dropdown、Slider、Popover、ScrollSpy、Tabs。

#### `$().DOMObserve(options, callback)`

- `options`: 监视的属性（[具体参见](https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver#MutationObserverInit)），默认为 `{childList: true, subtree: true}`；
- `callback(mutations, observer)`: DOM 发生变化时的处理函数，第一个参数为存储 [MutationRecord](https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver#MutationRecord) 对象的数组，第二个参数为 MutationObserver 实例本身。

`````html
<p id="js-do-actions">
  <button class="am-btn am-btn-primary" data-insert>插入 p 元素</button>
  <button class="am-btn am-btn-secondary" data-addClass>添加 Class</button>
  <button class="am-btn am-btn-warning" data-remove>移除 p 元素</button>
</p>
<div id="js-do-demo">
  <p>DOM 变化监视演示，打开控制台查看 log</p>

</div>
<script>
  $(function() {
    var $wrapper = $('#js-do-demo');
    $wrapper.DOMObserve({
      childList: true,
      attributes: true,
      subtree: true
    }, function(mutations, observer) {
      console.log(observer.constructor === window.MutationObserver);
      console.log('#js-do-demo 的 DOM 发生变化鸟：' + mutations[0].type);
    });

    $('#js-do-actions').find('button').on('click', function(e) {
      var $t = $(e.target);
      if ($t.is('[data-insert]')) {
        $wrapper.append('<p>插入了一个 p</p>');
      } else if($t.is('[data-remove]')) {
        $wrapper.find('p').last().remove();
      } else {
        $wrapper.addClass('am-text-danger');
      }
    });
  })
</script>
`````
```html
<p id="js-do-actions">
  <button class="am-btn am-btn-primary" data-insert>插入 p 元素</button>
  <button class="am-btn am-btn-secondary" data-addClass>添加 Class</button>
  <button class="am-btn am-btn-warning" data-remove>移除 p 元素</button>
</p>
<div id="js-do-demo">
  <p>DOM 变化监视演示，打开控制台查看 log</p>
</div>
<script>
  $(function() {
    var $wrapper = $('#js-do-demo');
    $wrapper.DOMObserve({
      childList: true,
      attributes: true,
      subtree: true
    }, function(mutations, observer) {
      console.log(observer.constructor === window.MutationObserver);
      console.log('#js-do-demo 的 DOM 发生变化鸟：' + mutations[0].type);
    });

    $('#js-do-actions').find('button').on('click', function(e) {
      var $t = $(e.target);
      if ($t.is('[data-insert]')) {
        $wrapper.append('<p>插入了一个 p</p>');
      } else if($t.is('[data-remove]')) {
        $wrapper.find('p').last().remove();
      } else {
        $wrapper.addClass('am-text-danger');
      }
    });
  })
</script>
```

**参考链接：**

- [MDN - MutationObserver](https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver)；
- [CIU - Mutation Observer 浏览器支持](http://caniuse.com/#feat=mutationobserver)
- [Polyfill - MutationObserver.js](https://github.com/webcomponents/webcomponentsjs/blob/master/src/MutationObserver/MutationObserver.js)

### 模块化开发

关于前端模块化，Amaze UI 1.0 的时候曾做过一个[关于前端 JS 模块化的调查](/javascript?_ver=1.x)，截止 2014.11.13 共 1869 个投票：

- CMD - Sea.js  23.86%  (446 votes)
- AMD - RequireJS  24.51%  (458 votes)
- CommonJS - Browserify  9.58%  (179 votes)
- 其他加载工具（或者自行开发的）  8.19%  (153 votes)
- 什么是 JS 模块化？可以吃吗？  34%  (633 votes)

<div class="am-progress">
  <div class="am-progress-bar" style="width: 23.8%" data-am-popover="{content: 'CMD - Sea.js  23.86%  (446 votes)', trigger: 'hover focus'}">CMD</div>
  <div class="am-progress-bar am-progress-bar-secondary" data-am-popover="{content: 'AMD - RequireJS  24.51%  (458 votes)', trigger: 'hover focus'}" style="width: 24.5%" >AMD</div>
  <div class="am-progress-bar am-progress-bar-success" style="width: 9.5%" data-am-popover="{content: 'CommonJS - Browserify  9.58%  (179 votes)', trigger: 'hover focus'}">CJS</div>
  <div class="am-progress-bar am-progress-bar-warning" style="width: 8.2%" data-am-popover="{content: '其他加载工具（或者自行开发的）  8.19%  (153 votes)', trigger: 'hover focus'}">other</div>
  <div class="am-progress-bar am-progress-bar-danger" style="width: 34%" data-am-popover="{content: '什么是 JS 模块化？可以吃吗？  34%  (633 votes)', trigger: 'hover focus'}">unknown</div>
</div>

显然，**模块化是必然趋势**，[ES6](http://wiki.ecmascript.org/doku.php?id=harmony:modules) 将原生支持模块化。

Amaze UI 2.0 按照 [CommonJS](http://wiki.commonjs.org/wiki/CommonJS) 规范来组织模块（前端也像 Node.js 一样编写模块）。最终如何打包，用户可以自行选择。

- [Browserify](http://browserify.org/)：结合 NPM，实现前端模块管理。很多前端模块都已经发布到 NPM，可以抛弃 Bower 这类功能很单一的工具了；
- [Duo](http://duojs.org/)：除管理本地模块以外，还可以从 GitHub 上直接获取开源项目，支持 Javascript、HTML、CSS；
- [gulp-amd-bundler](https://www.npmjs.org/package/gulp-amd-bundler)：把按照 CJS 编写的模块打包成 AMD 模块;
- [Webpack](https://github.com/webpack/webpack)。

[SPM](http://spmjs.io/) 貌似不支持直接通过源码提取依赖，使用 Sea.js 的用户可能需要自行修改打包工具。

__建议阅读的文章：__

* [前端模块化开发那点历史](https://github.com/seajs/seajs/issues/588)
* [Writing Modular JavaScript With AMD, CommonJS & ES Harmony](http://addyosmani.com/writing-modular-js/)
