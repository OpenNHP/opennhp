# Image
---

定义图片样式。

## 基础样式

基础样式定义在 `base` 中。

```css
img {
  box-sizing: border-box;
  /* v2.3 开始移除以下代码，详见 #502 */
  /* max-width: 100%;
  height: auto;*/
  vertical-align: middle;
  border: 0;
}
```

## 响应式图片

~~如上面的代码所示， `base` 里已经设置了 `max-width: 100%`，图片会自动适应到容器的宽度（但不会超过图片原始宽度），不需要添加额外的 class（[演示](http://jsbin.com/ciduf/1)）。如果要让图片始终和容器一样宽，需要设置 `width: 100%`。~~

`v2.3` 为解决 [#502](https://github.com/allmobilize/amazeui/issues/502)，基础样式中取消了图片最大宽度设置，新增了 `.am-img-responsive` class。

`````html
<img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg" class="am-img-responsive" alt=""/>
`````
```html
<img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg" class="am-img-responsive" alt=""/>
```

## 增强样式

### 圆角样式

为`<img>`元素设置不同的 class，增强其样式。

- `.am-radius`     圆角
- `.am-round`      椭圆
- `.am-circle`     圆形，一般用于正方形的图片(你要觉得椭圆好看，用在长方形上也可以)

`````html
<p><img class="am-radius" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />
<img class="am-round" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/600/q/80" width="200" height="120"/>
<img class="am-circle" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140"/></p>
`````
```html
<p>
  <img class="am-radius" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />

  <img class="am-round" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/600/q/80" width="200" height="120"/>

  <img class="am-circle" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140"/>
</p>
```


### 边框

- `.am-img-thumbnail`   边框

`````html
<img class="am-img-thumbnail" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />

<img class="am-img-thumbnail am-radius" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />

<img class="am-img-thumbnail am-circle" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />
`````

```html
<img src="..." alt="..." class="am-img-thumbnail">
<img src="..." alt="..." class="am-img-thumbnail am-radius">
<img src="..." alt="..." class="am-img-thumbnail am-circle">
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
