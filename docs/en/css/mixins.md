# Mixins
---

This part contains some useful Less mixins, including some packaged CSS3 attributes. Browser prefixes are considered in the mixins, so Amaze UI can function well in most modern browsers. We recommend using these mixin in your project.

## `clearfix()`

See [A new micro clearfix hack](http://nicolasgallagher.com/micro-clearfix-hack/) for detail.

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

Set color for placeholder

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

Display text in single line, and cut off the overflow text. The value of `display` should be `block` or `inline-block`.

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
<p class="doc-to">
One Web, Any Device.</p>
`````

```html
<div class="doc-to">
  One Web, Any Device.
</div>
```

```css
.doc-to {
  .text-overflow();
  width: 100px;
}
```

## `text-hide()`

CSS image replacement。Consider the accessibility, SEO and the display effect.

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
<h1 class="doc-ir"><a href="#">Allmobilize</a></h1>
`````

```html
<h1 class="doc-ir"><a href="#">Allmobilize</a></h1>
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

## Partial Round Cornor

- `border-top-radius()` Top Round Cordor
- `border-right-radius()` Right Round Cordor
- `border-bottom-radius()` Bottom Round Cordor
- `border-left-radius()` Left Round Cordor

`border-radius` has been fully supported in modern browsers, so you kan just remove those prefixes.

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

[CSS3 Shadow](https://developer.mozilla.org/zh-CN/docs/CSS/box-shadow)。

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

Some CSS transform functions. Use when you need.

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

 [`box-sizing`](https://developer.mozilla.org/zh-CN/docs/CSS/box-sizing) can be used to change the [CSS box model](https://developer.mozilla.org/en-US/docs/CSS/Box_model), and therefore change the way of calculating width and hight of elements.


- `content-box`
    Default. Standard box model. Width and height is defined as width and height of content, not including border, padding and margin.
- `padding-box`
     Width and height include padding, but not border and margin.
- `border-box`
     Width and height include padding and border. This is the so called Quirks mode used in IE. 

### Source Code

```css
.box-sizing(@boxmodel) {
  -webkit-box-sizing: @boxmodel;
  -moz-box-sizing: @boxmodel;
  box-sizing: @boxmodel;
}
```


## CSS3 Gradient
- `.horizontal(@start-color: #555; @end-color: #333; @start-percent: 0%; @end-percent: 100%)` - From left to right
- `.vertical(@start-color: #555; @end-color: #333; @start-percent: 0%; @end-percent: 100%)` - From top to bottom
- `.directional(@start-color: #555; @end-color: #333; @deg: 45deg)` - Specific direction
- `.horizontal-3c(@start-color: #00b3ee; @mid-color: #7a43b6; @color-stop: 50%; @end-color: #c3325f)` - Vertical gradient with three colors
- `.vertical-3c(@start-color: #00b3ee; @mid-color: #7a43b6; @color-stop: 50%; @end-color: #c3325f)` - Horizontal gradient with three colors
- `.radial(@inner-color: #555; @outer-color: #333)` - Radial gradient
- `.striped(@color: rgba(255,255,255,.15); @angle: 45deg)` - Striped gradient

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

**DON'T USE THESE COLORS FOR GRADIENT IN YOUR WEBPAGE**


## CSS Triangle

- `.caret-up(@size: 6px; @color:#222)` - Up
- `.caret-down(@size: 6px; @color:#222)` - Down
- `.caret-left(@size: 6px; @color:#222)` - Left
- `.caret-right(@size: 6px; @color:#222)` - Right

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

CSS Arrow。
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
