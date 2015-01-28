---
id: sticky
title: 固定元素
titleEn: Sticky
prev: javascript/smooth-scroll.html
next: javascript/tabs.html
source: js/ui.sticky.js
doc: docs/javascript/sticky.md
---

# Sticky
---

当窗口滚动至元素上边距离时，将元素固定在窗口顶部。

## 基本演示

### 基本形式

在元素上添加 `data-am-sticky` 属性。

`````html
<div data-am-sticky>
  <button class="am-btn am-btn-primary am-btn-block">Stick to top</button>
</div>
`````
```html
<div data-am-sticky>
  <button class="am-btn am-btn-primary am-btn-block">Stick to top</button>
</div>
```

### 设置上边距

元素固定到窗口顶部后，默认上边距为 0，可以在设置上边距 `data-am-sticky="{top:100}"` 。

`````html
<div data-am-sticky="{top:80}">
  <button class="am-btn am-btn-primary">Stick 80px below the top</button>
</div>
`````
```html
<div data-am-sticky="{top:80}">
  <button class="am-btn am-btn-primary">Stick 80px below the top</button>
</div>
```

### 动画效果

使用 [CSS3 动画](http://amazeui.org/css/animation) 实现动画效果。

`````html
<div data-am-sticky="{animation: 'slide-top'}">
  <button class="am-btn am-btn-success am-btn-block">固定到顶部动画效果</button>
</div>
`````
```html
<div data-am-sticky="{animation: 'slide-top'}">
  <button class="am-btn am-btn-success am-btn-block">固定到顶部动画效果</button>
</div>
```

## 调用方式

### 通过 Data API

如上面的演示所示，在元素上添加 `data-am-sticky` 属性。

### 通过 JS

通过 `$.sticky(options)` 设置。

`````html
<div id="my-sticky">
  <button class="am-btn am-btn-danger">Stick via JavaScript & 150px below the top</button>
</div>
<script>
$(function() {
$('#my-sticky').sticky({
    top: 150,
    bottom: function() {
      return $('.amz-footer').height();
    }
  });
});
</script>
`````
```html
<div id="my-sticky">
  <button class="am-btn am-btn-danger">Stick via JavaScript</button>
</div>
<script>
$(function() {
  $('#my-sticky').sticky({
    top: 150
  })
});
</script>
```

### 选项

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th style="width: 60px;">参数</th>
    <th style="width: 70px;">类型</th>
    <th style="width: 50px;">默认</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>top</code></td>
    <td>数值</td>
    <td><code>0</code></td>
    <td>距离顶部位置</td>
  </tr>
  <tr>
    <td><code>animation</code></td>
    <td>字符串</td>
    <td><code>''</code></td>
    <td>动画名称</td>
  </tr>
  <tr>
    <td><code>bottom</code></td>
    <td>数值 <br/> 或返回数值的函数</td>
    <td><code>0</code></td>
    <td>距离底部小于该数值时不再往下滚动，避免覆盖下面的元素</td>
  </tr>
  </tbody>
</table>

## 注意事项

- Sticky 插件是监听窗口滚动事件，当滚动距离超过元素上边距时，添加 `.am-sticky` 类，将元素的 `position` 设置为 `fixed`，同时设置一个 `top` 值（默认为 0）。
- 插件初始化的时候会给在元素外面包裹 `.am-sticky-placeholder` 作为占位符避免页面抖动，有可能会影响使用使用子选择的样式。
- __已知问题__：如果设置了动画，窗口快速 `resize` 时，动画会执行多次。


<div style="height: 400px"></div>
