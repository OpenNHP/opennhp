# Close
---

关闭按钮样式，可以结合其他不同组件使用。对 `<a>` 或者 `<button>` 添加 `.am-close` class。


## 基本样式

在元素上添加 `.am-close` class。

`````html
<a href="#" class="am-close">&times;</a>
<br />
<button type="button" class="am-close">&times;</button>
`````

```html
<a href="#" class="am-close">&times;</a>

<button type="button" class="am-close">&times;</button>
```


## 带边框样式

添加 `.am-close-alt` class。

### 使用 x

`````html
<a href="" class="am-close am-close-alt">&times;</a>
`````

```html
<a href="" class="am-close am-close-alt">&times;</a>
```

### 使用 Icon Font

`````html
  <a href="" class="am-close am-close-alt am-icon-times"></a>
`````

```html
<a href="" class="am-close am-close-alt am-icon-times"></a>
```

### 垂直居中思密达

<div class="am-alert am-alert-warning">
这个问题有点头疼，不同字体对齐方式有差异，很难做到所有字体都垂直居中。现在增加了使用 Icon Font 样式（貌似还行，看得我眼睛都眨巴了），大家如果有好的解决方案也可以提供给我们。
</div>

## hover 旋转

添加 `.am-close-spin` class（需支持 [CSS3 transform](https://developer.mozilla.org/zh-CN/docs/Web/CSS/transform#.E6.97.8B.E8.BD.AC)）。

`````html
<a href="" class="am-close am-close-alt am-close-spin">&times;</a>
<a href="" class="am-close am-close-alt am-close-spin am-icon-times"></a>
`````

```html
<a href="" class="am-close am-close-alt am-close-spin">&times;</a>
<a href="" class="am-close am-close-alt am-close-spin am-icon-times"></a>
```
