# Sticky
---

当窗口滚动至元素上边距离时，将元素固定在窗口顶部。

## 基本使用

在元素上添加 `data-am-sticky` 属性。

`````html
<div data-am-sticky="{animation: 'slide-top'}">
  <button class="am-btn am-btn-primary am-btn-block">Stick to top</button>
</div>
`````
```html
<div data-am-sticky>
  <button class="am-btn am-btn-primary am-btn-block">Stick to top</button>
</div>
```

## 设置上边距

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

## JS 调用

`````html
<div id="my-sticky">
  <button class="am-btn am-btn-danger">Stick via JavaScript & 150px below the top</button>
</div>
<script>
  seajs.use(['ui.sticky'], function() {
    $(function() {
      $('#my-sticky').sticky({
        top: 150
      })
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

## 注意事项

Sticky 插件的功能只是监听窗口滚动事件，当滚动距离超过元素上边距时，添加 `.am-sticky` 类，将元素的 `position` 设置为 `fixed`，同时设置一个 `top` 值（默认为 0）。实际使用中，还需要根据需求设定元素的 `width`、`left`、`right` 等样式。


`````html
<div style="height: 400px"></div>
`````