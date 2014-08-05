# Mixins
---

Mixins 整理了一些常用的 Less 函数，包括 CSS3 各种属性封装。Mixins 封装的时候即考虑了主流浏览器的兼容性，
同时避免一些无用的浏览器前缀属性，建议在实际项目中使用相关 mixin。

## `clearfix()`

详见 [A new micro clearfix hack](http://nicolasgallagher.com/micro-clearfix-hack/)

### Source Code

```css
.clearfix() {
  &:before,
  &:after {
    content: " "; /* 1 */
    display: table; /* 2 */
  }
  &:after {
    clear: both;
  }
}
```

### Example

`````html
<div class="am-cf doc-cf">
  <div class="am-fl">float:left</div>
  <div class="am-fr">float:right</div>
</div>
`````

```html
<div class="am-cf">
  <div class="am-fl">float:left</div>
  <div class="am-fr">float:right</div>
</div>
```

## `placeholder(@color)`

设置表单占位符颜色

### Source Code

```css
.placeholder(@color: @input-color-placeholder) {
  &:-moz-placeholder            { color: @color; } // Firefox 4-18
  &::-moz-placeholder           { color: @color; } // Firefox 19+
  &:-ms-input-placeholder       { color: @color; } // Internet Explorer 10+
  &::-webkit-input-placeholder  { color: @color; } // Safari and Chrome
}
```

### Example

`````html
<input type="text" class="doc-placeholder" placeholder="Hello, AM UI."/>
`````

```html
<input type="text" class="doc-placeholder" placeholder="Hello, AM UI."/>
```

```css
.doc-placeholder {
  .placeholder(red);
}
```

## `text-overflow()`

单行显示文字，超过容器宽度时自动截断。 `display` 的值应为 `block` 或者 `inline-block`。

### Source Code

```css
.text-overflow(@display: block;){
  display: @display;
  word-wrap: normal; /*for IE*/
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
}
```

### Example

`````html
<p class="doc-to">云适配是一项全球独一无二的网站跨平台适配技术，是指在原网站中插入一行代码</p>
`````

```html
<div class="doc-to">云适配是一项全球独一无二的网站跨平台适配技术，是指在原网站中插入一行代码
</div>
```

```css
.doc-to {
  .text-overflow();
  width: 100px;
}
```

## `text-hide()`

CSS image replacement。考虑可访性和SEO，同时兼顾显示效果。

### Source Code

```css
.text-hide() {
  font: ~"0/0" a;
  color: transparent;
  text-shadow: none;
  background-color: transparent;
  border: 0;
}
```

### Example

`````html
<h1 class="doc-ir"><a href="#">云适配</a></h1>
`````

```html
<h1 class="doc-ir"><a href="#">云适配</a></h1>
```

```css
.doc-ir {
  a {
    display: block;
    .text-hide;
    background:#000 url(http://www.yunshipei.com/static/img/logo.png) no-repeat 0 0;
    height: 38px;
  }
}
```

## 局部边框圆角

- `border-top-radius()` 上边框圆角
- `border-right-radius()` 右边看圆角
- `border-bottom-radius()` 下边框圆角
- `border-left-radius()` 左边框圆角

`border-radius` 已经得到主流现代浏览器完整的支持，把你那没有的前缀都去掉吧。

### Source Code

```css
.border-top-radius(@radius) {
  border-top-right-radius: @radius;
  border-top-left-radius: @radius;
}

.border-right-radius(@radius) {
  border-bottom-right-radius: @radius;
  border-top-right-radius: @radius;
}

.border-bottom-radius(@radius) {
  border-bottom-right-radius: @radius;
  border-bottom-left-radius: @radius;
}

.border-left-radius(@radius) {
  border-bottom-left-radius: @radius;
  border-top-left-radius: @radius;
}
```

### Example

`````html
<div class="doc-box doc-br-top">border-top-radius</div>
<div class="doc-box doc-br-right">border-right-radius</div>
<div class="doc-box doc-br-bottom">border-bottom-radius</div>
<div class="doc-box doc-br-left">border-left-radius</div>
`````

## `box-shadow()`

[CSS3 阴影](https://developer.mozilla.org/zh-CN/docs/CSS/box-shadow)。

### Source Code
```css
.box-shadow(@shadow) {
  -webkit-box-shadow: @shadow; // iOS <4.3 & Android <4.1 & bb7.0
  box-shadow: @shadow;
}
```

### Example

`````html
<div class="doc-box doc-box-shadow"></div>
`````


## `transition()`

[CSS3 transition](https://developer.mozilla.org/zh-CN/docs/CSS/transition).

### Source Code

```css
.transition(@transition) {
  -webkit-transition: @transition;
  transition: @transition;
}
```


## [CSS3 transform](https://developer.mozilla.org/zh-CN/docs/CSS/transform)

封装了几种 CSS transform函数，按需使用。

```css
.rotate(@degrees) {
  -webkit-transform: rotate(@degrees);
  -ms-transform: rotate(@degrees); // IE9+
  transform: rotate(@degrees);
}
.scale(@ratio) {
  -webkit-transform: scale(@ratio);
  -ms-transform: scale(@ratio); // IE9+
  transform: scale(@ratio);
}
.translate(@x; @y) {
  -webkit-transform: translate(@x, @y);
  -ms-transform: translate(@x, @y); // IE9+
  transform: translate(@x, @y);
}
.skew(@x; @y) {
  -webkit-transform: skew(@x, @y);
  -ms-transform: skewX(@x) skewY(@y); // See https://github.com/twbs/bootstrap/issues/4885; IE9+
  transform: skew(@x, @y);
}
.translate3d(@x; @y; @z) {
  -webkit-transform: translate3d(@x, @y, @z);
  transform: translate3d(@x, @y, @z);
}

.rotateX(@degrees) {
  -webkit-transform: rotateX(@degrees);
  -ms-transform: rotateX(@degrees); // IE9+
  transform: rotateX(@degrees);
}
.rotateY(@degrees) {
  -webkit-transform: rotateY(@degrees);
  -ms-transform: rotateY(@degrees); // IE9+
  transform: rotateY(@degrees);
}
.perspective(@perspective) {
  -webkit-perspective: @perspective;
  -moz-perspective: @perspective;
  perspective: @perspective;
}
.perspective-origin(@perspective) {
  -webkit-perspective-origin: @perspective;
  -moz-perspective-origin: @perspective;
  perspective-origin: @perspective;
}
.transform-origin(@origin) {
  -webkit-transform-origin: @origin;
  -moz-transform-origin: @origin;
  transform-origin: @origin;
}
```


## [CSS3 Animation](https://developer.mozilla.org/en-US/docs/Web/CSS/animation?redirectlocale=en-US&redirectslug=CSS%2Fanimation)

```css
.animation(@animation) {
  -webkit-animation: @animation;
  animation: @animation;
}
```

## `box-sizing()`

 [`box-sizing`](https://developer.mozilla.org/zh-CN/docs/CSS/box-sizing) 用来改变 [CSS 盒模型](https://developer.mozilla.org/en-US/docs/CSS/Box_model) ，从而改变元素高宽的计算方式。


- `content-box`
    默认值，标准盒模型。 width 与 height 是内容区的宽与高， 不包括边框，内边距，外边距。
- `padding-box`
     width 与 height 包括内边距，不包括边距与外边距。
- `border-box`
     width 与 height 包括内边距与边框，不包括外边距。这是IE 怪异模式（Quirks mode）使用的 盒模型 。 

### Source Code

```css
.box-sizing(@boxmodel) {
  -webkit-box-sizing: @boxmodel;
  -moz-box-sizing: @boxmodel;
  box-sizing: @boxmodel;
}
```


## CSS3 渐变
- `.horizontal(@start-color: #555; @end-color: #333; @start-percent: 0%; @end-percent: 100%)` - 从左至右的水平渐变
- `.vertical(@start-color: #555; @end-color: #333; @start-percent: 0%; @end-percent: 100%)` - 从上到下垂直渐变
- `.directional(@start-color: #555; @end-color: #333; @deg: 45deg)` - 定向渐变
- `.horizontal-3c(@start-color: #00b3ee; @mid-color: #7a43b6; @color-stop: 50%; @end-color: #c3325f)` - 水平三色渐变
- `.vertical-3c(@start-color: #00b3ee; @mid-color: #7a43b6; @color-stop: 50%; @end-color: #c3325f)` - 垂直三色渐变
- `.radial(@inner-color: #555; @outer-color: #333)` - 放射状渐变
- `.striped(@color: rgba(255,255,255,.15); @angle: 45deg)` - 条纹渐变

### Example

`````html
<div class="doc-gradient doc-gradient-1"></div>
<div class="doc-gradient doc-gradient-2"></div>
<div class="doc-gradient doc-gradient-3"></div>
<div class="doc-gradient doc-gradient-4"></div>
<div class="doc-gradient doc-gradient-5"></div>
<div class="doc-gradient doc-gradient-6"></div>
<div class="doc-gradient doc-gradient-7"></div>
`````

```html
<div class="doc-gradient doc-gradient-1"></div>
<div class="doc-gradient doc-gradient-2"></div>
<div class="doc-gradient doc-gradient-3"></div>
<div class="doc-gradient doc-gradient-4"></div>
<div class="doc-gradient doc-gradient-5"></div>
<div class="doc-gradient doc-gradient-6"></div>
<div class="doc-gradient doc-gradient-7"></div>
```

```css
.doc-gradient-1 {
  #gradient > .horizontal(red, black);
}
.doc-gradient-2 {
  #gradient > .vertical(red, black);
}
.doc-gradient-3 {
  #gradient > .directional(red, black);
}
.doc-gradient-4 {
  #gradient > .horizontal-3c();
}
.doc-gradient-5 {
  #gradient > .vertical-3c();
}
.doc-gradient-6 {
    #gradient > .radial(red, black);
}
.doc-gradient-7 {
  #gradient > .striped(blue);
}
```

示例中为了突出效果，设置的颜色比较奇葩，如果不想被设计掐死，最好别写出这样的颜色。

渐变 mixins 的用法如上面，`#gradient > .striped` 这样子（别告诉我单词太长记不住，框架不解决智商问题）。


## CSS 三角形

- `.caret-up(@size: 6px; @color:#222)` - 普通三角形
- `.caret-down(@size: 6px; @color:#222)` - 二逼（倒立）三角形
- `.caret-left(@size: 6px; @color:#222)` - 左撇三角形
- `.caret-right(@size: 6px; @color:#222)` - 右拐三角形

### Source Code

```css
.caret-down(@size: 6px; @color:#222) {
  display: inline-block;
  width: 0;
  height: 0;
  vertical-align: middle;
  border-top:   @size solid @color;
  border-right: @size solid transparent;
  border-left:  @size solid transparent;
  // Firefox fix for https://github.com/twbs/bootstrap/issues/9538. Once fixed,
  // we can just straight up remove this.
  border-bottom: 0 dotted;
  .rotate(360deg);
}
```

### Example

`````html
<span class="doc-caret doc-caret-u"></span>
<span class="doc-caret doc-caret-d"></span>
<span class="doc-caret doc-caret-l"></span>
<span class="doc-caret doc-caret-r"></span>
`````

```html
<span class="doc-caret doc-caret-u"></span>
<span class="doc-caret doc-caret-d"></span>
<span class="doc-caret doc-caret-l"></span>
<span class="doc-caret doc-caret-r"></span>
```

```css
.doc-caret-u {
   .caret-up(10px,red);
 }

.doc-caret-d {
  .caret-down(15px,purple);
}

.doc-caret-l {
  .caret-left(20px, blue);
}

.doc-caret-r {
  .caret-right(25px, green);
}
```

## CSS Arrow

CSS 箭头。
- 45deg (default) - right
- 135deg - down
- -45deg - up
- -135deg -left

### Source Code

```css
.arrow(@color:#DDD; @width:6px; @border-width: 2px; @deg: 45deg) {
  display: inline-block;
  width: @width;
  height: @width;
  border: @color solid;
  border-width: @border-width @border-width 0 0;
  .rotate(@deg);
}
```

### Example

`````html
<span class="doc-arrow doc-arrow-r"></span>
<span class="doc-arrow doc-arrow-d"></span>
<span class="doc-arrow doc-arrow-u"></span>
<span class="doc-arrow doc-arrow-l"></span>
`````

```html
<span class="doc-arrow doc-arrow-r"></span>
<span class="doc-arrow doc-arrow-d"></span>
<span class="doc-arrow doc-arrow-u"></span>
<span class="doc-arrow doc-arrow-l"></span>
```

```css
.doc-arrow {
  .arrow(purple, 10px, 3px)
}

.doc-arrow-d {
  .arrow(red, 10px, 3px, 135deg)
}

.doc-arrow-u {
  .arrow(green, 10px, 3px, -45deg)
}

.doc-arrow-l {
  .arrow(green, 10px, 3px, -135deg)
}
```
