# Image
---

定义图片样式。

## 基础样式

基础样式定义在 `base` 中。

```css
img {
  box-sizing: border-box;
  max-width: 100%;
  height: auto;
  vertical-align: middle;
  border: 0;
}
```

## 宽度自适应

如上面的代码所示， `base` 里已经设置了 `max-width: 100%`，图片会自动适应到容器的宽度（但不会超过图片原始宽度），不需要添加额外的 class（[演示](http://jsbin.com/ciduf/1)）。如果要让图片始终和容器一样宽，需要设置 `width: 100%`。

`````html
<img src="http://amui.qiniudn.com/bw-2014-06-19.jpg" alt=""/>
`````
```html
<img src="http://amui.qiniudn.com/bw-2014-06-19.jpg" alt=""/>
```

## 增强样式

为`<img>`元素设置不同的 class，增强其样式。

- `.am-img-bdrs`     圆角
- `.am-img-circle`      圆形，一般用于正方形的图片(你要觉得椭圆好看，用在长方形上也可以)
- `.am-img-thumbnail`   边框

`````html
<img class="am-img-bdrs" alt="140*140" src="http://amui.qiniudn.com/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />
<img class="am-img-circle" alt="140*140" src="http://amui.qiniudn.com/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80"  width="140" height="140" />
<img class="am-img-thumbnail" alt="140*140" src="http://amui.qiniudn.com/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />
`````

```html
<img src="..." alt="..." class="am-img-bdrs">
<img src="..." alt="..." class="am-img-circle">
<img src="..." alt="..." class="am-img-thumbnail">
```

<!--
## 响应式图片

通过添加 `.am-img-responsive` class 让图片按比例缩放。
    
`````html
<img class="am-img-responsive" alt="Responsive image" src="http://www.bing.com/az/hprichbg/rb/AdelaideFrog_EN-US12171255358_1366x768.jpg" />
`````

```html
<img src="..." class="am-img-responsive" alt="Responsive image">
```-->
