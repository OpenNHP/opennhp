---
id: smooth-scroll
title: 平滑滚动
titleEn: Smooth Scroll
prev: javascript/scrollspynav.html
next: javascript/sticky.html
source: js/ui.smooth-scroll.js
doc: docs/javascript/smooth-scroll.md
---

# Smooth Scroll
---

平滑滚动插件，源自 [Zepto 作者](https://gist.github.com/madrobby/8507960#file-scrolltotop-annotated-js)。

<div class="am-alert am-alert-danger">本插件不支持 IE 9 及以下版本！如果有相关需求请找别的插件代替。</div>

如果要支持旧版 IE，可以使用下面的代码实现：

```js
$('html, body').animate({scrollTop: 0}, '500');
```

## 使用演示

### 滚动到顶部

`````html
<button data-am-smooth-scroll class="am-btn am-btn-success">滚动到顶部</button>
`````

```html
<button data-am-smooth-scroll class="am-btn am-btn-success">滚动到顶部</button>
```

### 滚动到底部

`````html
<button id="doc-scroll-to-btm" class="am-btn am-btn-primary">滚动到底部</button>
<script>
  $('#doc-scroll-to-btm').on('click', function() {
    var $w = $(window);
    $w.smoothScroll({position: $(document).height() - $w.height()});
  });
</script>
`````

```html
<button id="doc-scroll-to-btm" class="am-btn am-btn-primary">滚动到底部</button>
<script>
  $('#doc-scroll-to-btm').on('click', function() {
    var $w = $(window);
    $w.smoothScroll({position: $(document).height() - $w.height()});
  });
</script>
```

### 定义选项

`````html
<button data-am-smooth-scroll="{position: 57, speed: 5000}" class="am-btn am-btn-danger">慢悠悠地滚到距离顶部 57px 的位置</button>
`````

```html
<button data-am-smooth-scroll="{position: 57, speed: 5000}" class="am-btn am-btn-danger">慢悠悠地滚到距离顶部 57px 的位置</button>
```

## 使用方法

### 通过 Data API

在元素上添加 `data-am-smooth-scroll` 属性。

```html
<button data-am-smooth-scroll class="am-btn am-btn-success">滚动到顶部</button>
```

如果要指定滚动的位置，可以给这个属性设一个值。

`````html
<button data-am-smooth-scroll="{position: 189}" class="am-btn am-btn-secondary">滚动到滚动条距离顶部 189px 的位置</button>
`````
```html
<button data-am-smooth-scroll="{position: 189}" class="am-btn am-btn-secondary">...</button>
```
### 通过 Javascript

#### 方法

为了保证不同浏览器兼容，请在 `$(window)` 上调用 `$().smoothScroll()` 方法。

```javascript
$(window).smoothScroll([options])
```

```javascript
// 滚动到底部
$('#my-button').on('click', function() {
  var $w = $(window);
  $w.smoothScroll({position: $(document).height() - $w.height()});
});
```

#### 选项说明

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th style="width: 60px;">参数</th>
    <th style="width: 70px;">类型</th>
    <th style="width: 110px;">默认</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>position</code></td>
    <td>数值</td>
    <td><code>0</code></td>
    <td>要滚动到位置，默认为 `0`，即滚动到顶部</td>
  </tr>
  <tr>
    <td><code>speed</code></td>
    <td>数值</td>
    <td><code>750 ~ 1500</code></td>
    <td>滚动速度，单位 `ms`，默认为 `750 - 1500`，根据距离判断</td>
  </tr>
  </tbody>
</table>
