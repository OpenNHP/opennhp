---
id: dropdown
title: 下拉组件
titleEn: Dropdown
prev: javascript/collapse.html
next: javascript/modal.html
source: js/ui.dropdown.js
doc: docs/javascript/dropdown.md
---

# Dropdown
---

## 使用演示

### 下拉列表

`````html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-primary am-dropdown-toggle" data-am-dropdown-toggle>下拉列表 <span class="am-icon-caret-down"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">标题</li>
    <li><a href="#">快乐的方式不只一种</a></li>
    <li class="am-active"><a href="#">最荣幸是</a></li>
    <li><a href="#">谁都是造物者的光荣</a></li>
    <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
    <li class="am-divider"></li>
    <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
  </ul>
</div>
`````
```html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-primary am-dropdown-toggle" data-am-dropdown-toggle>下拉列表 <span class="am-icon-caret-down"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">标题</li>
    <li><a href="#">快乐的方式不只一种</a></li>
    <li class="am-active"><a href="#">最荣幸是</a></li>
    <li><a href="#">谁都是造物者的光荣</a></li>
    <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
    <li class="am-divider"></li>
    <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
  </ul>
</div>
```

### 上拉列表

在 `.am-dropdown` 上添加 `.am-dropdown-up` class，在上面弹出内容。

`````html
<div class="am-dropdown am-dropdown-up" data-am-dropdown>
  <button class="am-btn am-btn-danger am-dropdown-toggle" data-am-dropdown-toggle>上拉列表 <span class="am-icon-caret-up"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">标题</li>
    <li><a href="#">快乐的方式不只一种</a></li>
    <li class="am-active"><a href="#">最荣幸是</a></li>
    <li><a href="#">谁都是造物者的光荣</a></li>
    <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
    <li class="am-divider"></li>
    <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
  </ul>
</div>
`````
```html
<div class="am-dropdown am-dropdown-up" data-am-dropdown>
  <button class="am-btn am-btn-danger am-dropdown-toggle" data-am-dropdown-toggle>上拉列表 <span class="am-icon-caret-up"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">标题</li>
    <li><a href="#">快乐的方式不只一种</a></li>
    <li class="am-active"><a href="#">最荣幸是</a></li>
    <li><a href="#">谁都是造物者的光荣</a></li>
    <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
    <li class="am-divider"></li>
    <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
  </ul>
</div>
```

### 下拉内容

`````html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-success am-dropdown-toggle">下拉内容 <span class="am-icon-caret-down"></span></button>
  <div class="am-dropdown-content">
    <h2>I am what I am</h2>
    <p>
      多么高兴 在琉璃屋中快乐生活
      对世界说 甚么是光明和磊落
      我就是我 是颜色不一样的烟火
    </p>
  </div>
</div>
`````
```html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-success am-dropdown-toggle">Dropdown <span class="am-icon-caret-down"></span></button>
  <div class="am-dropdown-content">
    <h2>I am what I am</h2>
    <p>
      多么高兴 在琉璃屋中快乐生活
      对世界说 甚么是光明和磊落
      我就是我 是颜色不一样的烟火
    </p>
  </div>
</div>
```

### 宽度适应

下拉内容 `.am-dropdown-content` 为绝对定位，宽度会根据内容缩放（最小为 `160px`）。

在 `data-am-dropdown` 里指定要适应到的元素，下拉内容的宽度会设置为该元素的宽度。当然可以直接在 CSS 里设置下拉内容的宽度。

`````html
<div id="doc-dropdown-justify">
  <div class="am-dropdown" data-am-dropdown="{justify: '#doc-dropdown-justify'}">
    <button class="am-btn am-btn-success am-dropdown-toggle">宽度适应下拉 <span class="am-icon-caret-down"></span></button>
    <div class="am-dropdown-content">
      <h2>I am what I am</h2>
      <p>
        多么高兴 在琉璃屋中快乐生活
        对世界说 甚么是光明和磊落
        我就是我 是颜色不一样的烟火
      </p>
    </div>
  </div>
</div>
`````
```html
<div id="doc-dropdown-justify">
  <div class="am-dropdown" data-am-dropdown="{justify: '#doc-dropdown-justify'}">
    <button class="am-btn am-btn-success am-dropdown-toggle">宽度适应下拉 <span class="am-icon-caret-down"></span></button>
    <div class="am-dropdown-content">
      <h2>I am what I am</h2>
      <p>
        多么高兴 在琉璃屋中快乐生活
        对世界说 甚么是光明和磊落
        我就是我 是颜色不一样的烟火
      </p>
    </div>
  </div>
</div>
```
<!--
### 与 Header 嵌套使用

`````html
<header data-am-widget="header" class="am-header am-header-default">
  <h1 class="am-header-title">
    <a>Title</a>
  </h1>

  <div class="am-header-right am-header-nav am-dropdown" data-am-dropdown>
    <a class="am-dropdown-toggle" href="javascript: void(0)" data-am-dropdown-toggle>
      <i class="am-header-icon am-icon-bars"></i>
    </a>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">标题</li>
      <li><a href="#">快乐的方式不只一种</a></li>
      <li class="am-active"><a href="#">最荣幸是</a></li>
      <li><a href="#">谁都是造物者的光荣</a></li>
      <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
      <li class="am-divider"></li>
      <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
    </ul>
  </div>
</header>
`````-->

## 调用方式

### 通过 Data API

如上面的演示所示，根据示例组织好 HTML 结构，然后在 `.am-dropdown` 上添加 `data-am-dropdown` 属性，相关选项可以设置在该属性的值里。

### 通过 JS

按照示例组织好 HTML 结构（不加 `data-am-dropdown` 属性），然后通过 JS 来调用。

`````html
<div id="doc-dropdown-justify-js" style="width: 400px">
  <div class="am-dropdown" id="doc-dropdown-js">
    <button class="am-btn am-btn-danger am-dropdown-toggle">通过 JS 调用 <span class="am-icon-caret-down"></span></button>
    <div class="am-dropdown-content">
      <h2>I am what I am</h2>
      <p>
        多么高兴 在琉璃屋中快乐生活
        对世界说 甚么是光明和磊落
        我就是我 是颜色不一样的烟火
      </p>
    </div>
  </div>
</div>
<script>
  $(function() {
    $('#doc-dropdown-js').dropdown({justify: '#doc-dropdown-justify-js'});
  });
</script>
`````
```html
<div id="doc-dropdown-justify-js" style="width: 400px">
  <div class="am-dropdown" id="doc-dropdown-js">
    <button class="am-btn am-btn-danger am-dropdown-toggle">通过 JS 调用 <span class="am-icon-caret-down"></span></button>
    <div class="am-dropdown-content">...</div>
  </div>
</div>
<script>
  $(function() {
    $('#doc-dropdown-js').dropdown({justify: '#doc-dropdown-justify-js'});
  });
</script>
```

#### 方法

- `$(element).dropdown(options)` 激活下拉功能；
- `$(element).dropdown('toggle')` 下拉状态交替；
- `$(element).dropdown('close')` 隐藏下拉菜单；
- `$(element).dropdown('open')` 显示下拉菜单。

`````html
<button class="am-btn am-btn-secondary" id="doc-dropdown-toggle">调用 Toggle</button>
<button class="am-btn am-btn-success" id="doc-dropdown-open">调用 Open</button>
<button class="am-btn am-btn-warning" id="doc-dropdown-close">调用 Close</button>
<script>
  $(function() {
    var $dropdown = $('#doc-dropdown-js'),
        data = $dropdown.data('amui.dropdown');

    function scrollToDropdown() {
      $(window).smoothScroll({position: $dropdown.offset().top});
    }

    $('#doc-dropdown-toggle').on('click', function(e) {
      scrollToDropdown();
      $dropdown.dropdown('toggle');
      return false;
    });

    $('#doc-dropdown-open').on('click', function(e) {
      scrollToDropdown();
      data.active ? alert('已经打开了，施主又何必再纠缠呢！') : $dropdown.dropdown('open');
      return false;
    });

    $('#doc-dropdown-close').on('click', function(e) {
      scrollToDropdown();
      data.active ? $dropdown.dropdown('close') : alert('没有开哪有关，没有失哪有得！');
      return false;
    });

    $dropdown.on('open.dropdown.amui', function (e) {
      console.log('open event triggered');
    });
  });
</script>
`````
```html
<button class="am-btn am-btn-secondary" id="doc-dropdown-toggle">调用 Toggle</button>
<button class="am-btn am-btn-success" id="doc-dropdown-open">调用 Open</button>
<button class="am-btn am-btn-warning" id="doc-dropdown-close">调用 Close</button>
<script>
  $(function() {
    var $dropdown = $('#doc-dropdown-js'),
        data = $dropdown.data('amui.dropdown');
    $('#doc-dropdown-toggle').on('click', function(e) {
      $dropdown.dropdown('toggle');
      return false;
    });

    $('#doc-dropdown-open').on('click', function(e) {
      data.active ? alert('已经打开了，施主又何必再纠缠呢！') : $dropdown.dropdown('open');
      return false;
    });

    $('#doc-dropdown-close').on('click', function(e) {
      data.active ? $dropdown.dropdown('close') : alert('没有开哪有关，没有失哪有得！');
      return false;
    });

    $dropdown.on('open.dropdown.amui', function (e) {
      console.log('open event triggered');
    });
  });
</script>
```

#### 参数说明

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
    <td><code>boundary</code></td>
    <td>选择器</td>
    <td><code>window</code></td>
    <td>下拉内容边界，避免下拉内容超过边界被遮盖截断</td>
  </tr>
  <tr>
    <td><code>justify</code></td>
    <td>选择器</td>
    <td><code>undefined</code></td>
    <td>下拉内容适应宽度的元素</td>
  </tr>
  </tbody>
</table>

#### 自定义事件

下拉框的自定义事件在 `.am-dropdown` 上触发。

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.dropdown.amui</code></td>
    <td>调用显示下拉框方法时立即触发</td>
  </tr>
  <tr>
    <td><code>opened.dropdown.amui</code></td>
    <td>下拉框显示完成时触发</td>
  </tr>
  <tr>
    <td><code>close.dropdown.amui</code></td>
    <td>调用隐藏方法时触发</td>
  </tr>
  <tr>
    <td><code>closed.dropdown.amui</code></td>
    <td>下拉框关闭完成时触发</td>
  </tr>
  </tbody>
</table>

```js
$(function() {
  $dropdown.on('open.dropdown.amui', function (e) {
    console.log('open event triggered');
  });
});
```
