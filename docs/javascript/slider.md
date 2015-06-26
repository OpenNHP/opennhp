---
id: slider
title: 图片轮播
titleEn: Slider
prev: javascript/nprogress.html
next: javascript/offcanvas.html
source: js/ui.slider.js
doc: docs/javascript/slider.md
---

# Slider
---

图片轮播模块，支持触控事件，源自 [FlexSlider](https://github.com/woothemes/FlexSlider)。本插件算不上轻量，但功能相对完善，基本能满足各种不同的需求。

**类似插件**：

- [Swiper](https://github.com/amazeui/swiper)
- [Slick](https://github.com/amazeui/slick)

[Web 组件中的图片轮播](/widgets/slider?_ver=2.x) 调用此插件，只是样式上做了一些扩展。

**演示图标版权归[微软 Bing](http://www.bing.com) 所有。**

## 使用演示

### 基本形式

`````html
<div class="am-slider am-slider-default" data-am-flexslider id="demo-slider-0">
  <ul class="am-slides">
    <li><img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-3.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-4.jpg" /></li>
  </ul>
</div>

<hr />
<div class="am-btn-toolbar">
  <button type="button" class="am-btn am-btn-primary js-demo-slider-btn" data-action="add">添加 slide</button>
  <button type="button" class="am-btn am-btn-danger js-demo-slider-btn" data-action="remove">移除 slide</button>
</div>
<script>
  $(function() {
    var $slider = $('#demo-slider-0');
    var counter = 0;
    var getSlide = function() {
      counter++;
      return '<li><img src="http://s.amazeui.org/media/i/demos/bing-' +
        (Math.floor(Math.random() * 4) + 1) + '.jpg" />' +
        '<div class="am-slider-desc">动态插入的 slide ' + counter + '</div></li>';
    };

    $('.js-demo-slider-btn').on('click', function() {
      var action = this.getAttribute('data-action');
      if (action === 'add') {
        $slider.flexslider('addSlide', getSlide());
      } else {
        var count = $slider.flexslider('count');
        count > 1 && $slider.flexslider('removeSlide', $slider.flexslider('count') - 1);
      }
    });

  });
</script>
`````
```html
<div class="am-slider am-slider-default" data-am-flexslider id="demo-slider-0">
  <ul class="am-slides">
    <li><img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
  </ul>
</div>

<div class="am-btn-toolbar">
  <button type="button" class="am-btn am-btn-primary js-demo-slider-btn" data-action="add">添加 slide</button>
  <button type="button" class="am-btn am-btn-danger js-demo-slider-btn" data-action="remove">移除 slide</button>
</div>
<script>
  $(function() {
    var $slider = $('#demo-slider-0');
    var counter = 0;
    var getSlide = function() {
      counter++;
      return '<li><img src="http://s.amazeui.org/media/i/demos/bing-' +
        (Math.floor(Math.random() * 4) + 1) + '.jpg" />' +
        '<div class="am-slider-desc">动态插入的 slide ' + counter + '</div></li>';
    };

    $('.js-demo-slider-btn').on('click', function() {
      var action = this.getAttribute('data-action');
      if (action === 'add') {
        $slider.flexslider('addSlide', getSlide());
      } else {
        var count = $slider.flexslider('count');
        count > 1 && $slider.flexslider('removeSlide', $slider.flexslider('count') - 1);
      }
    });

  });
</script>
```

### 缩略图模式

使用该模式时需要在 `li` 上使用 `data-thumb` 指定缩略图地址。

`````html
<div
  class="am-slider am-slider-default"
  data-am-flexslider="{controlNav: 'thumbnails', directionNav: false, slideshow: false}">
  <ul class="am-slides">
    <li data-thumb="http://s.amazeui.org/media/i/demos/pure-1.jpg?imageView2/0/w/360"><img
      src="http://s.amazeui.org/media/i/demos/pure-1.jpg" /></li>
    <li data-thumb="http://s.amazeui.org/media/i/demos/pure-2.jpg?imageView2/0/w/360"><img
      src="http://s.amazeui.org/media/i/demos/pure-2.jpg" /></li>
    <li data-thumb="http://s.amazeui.org/media/i/demos/pure-3.jpg?imageView2/0/w/360"><img
      src="http://s.amazeui.org/media/i/demos/pure-3.jpg" /></li>
    <li data-thumb="http://s.amazeui.org/media/i/demos/pure-4.jpg?imageView2/0/w/360"><img
      src="http://s.amazeui.org/media/i/demos/pure-4.jpg" /></li>
  </ul>
</div>
`````
```html
<div
  class="am-slider am-slider-default"
  data-am-flexslider="{controlNav: 'thumbnails', directionNav: false, slideshow: false}">
  <ul class="am-slides">
    <li data-thumb="http://s.amazeui.org/media/i/demos/pure-4.jpg?imageView2/0/w/360">
      <img src="http://s.amazeui.org/media/i/demos/pure-4.jpg" /></li>
  </ul>
</div>
```

轮播图数量不等于 `4` 时需要手动设定缩略图的宽度

```css
.am-control-thumbs li {
  width: 100%/n; /* n 为轮播图数量 */
}
```

### 多图滚动

设置 `itemWidth` 选项实现多图滚动。`itemMargin` 仅用作计算滚动宽度用，要设置 `<li>` 之间的距离仍然需要 CSS，
  默认设置了 5px 的宽度，加上 `.am-slider-carousel` 后生效。

```css
.am-slider-carousel li {
  margin-right: 5px;
}
```

`````html
<div
  class="am-slider am-slider-default am-slider-carousel"
  data-am-flexslider="{itemWidth: 200, itemMargin: 5, slideshow: false}">
  <ul class="am-slides">
    <li><img src="http://s.amazeui.org/media/i/demos/pure-1.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/pure-2.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/pure-3.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/pure-4.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/pure-1.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/pure-2.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/pure-3.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/pure-4.jpg?imageView2/0/w/640" /></li>
  </ul>
</div>
`````
```html
<div
  class="am-slider am-slider-default am-slider-carousel"
  data-am-flexslider="{itemWidth: 200, itemMargin: 5, slideshow: false}">
  <ul class="am-slides">
    <li><img src="..." /></li>
  </ul>
</div>
```

## 使用方式

### HTML 结构

以下结构__必须__，`<li>` 中的内容可以自由组合。

```html
<div class="am-slider am-slider-default">
  <ul class="am-slides">
    <li>...</li>
  </ul>
</div>
```

### 通过 Data API 初始化

在容器上添加 `data-am-flexslider` 自动初始化。

```html
<div class="am-slider am-slider-default" data-am-flexslider>
  <ul class="am-slides">
    <li>...</li>
  </ul>
</div>
```

### 通过 JS 初始化

```javascript
$('.am-slider').flexslider();
```

#### 自定参数

在初始化函数中传入所需参数即可。

```javascript
$(function() {
  $('.am-slider').flexslider({
    // options
  });
});
```

#### 方法

- `$('#slider').flexslider('play');` - 播放
- `$('#slider').flexslider('pause');` - 暂停
- `$('#slider').flexslider('stop');` - 停止
- `$('#slider').flexslider('next');` - 上一个
- `$('#slider').flexslider('prev');` - 下一个
- `$('#slider').flexslider(3);` - 切换到第 4 个 slide

**`v2.4.1` 新增方法：**

- `$('#slider').flexslider('addSlide', obj[, position]);` - 添加幻灯片，`obj` 为 jQuery 对象或者 HTML，`position` 为插入的位置（可选）；
- `$('#slider').flexslider('removeSlide', obj);` - 移除指定幻灯片，`obj` 可以是数字或 jQuery 对象或 DOM 对象或选择符。

**注意**: `addSlide` 方法在缩略图模式下无法正常工作。

#### slider 实例

通过 slider 实例可以调用更多方法。**`v2.4.1` 开始，可以直接通过 jQuery 方式调用以下方法。**

```js
var slider = $('#slider').data('flexslider');

slider.addSlide(obj, position); // 插入的元素（选择符或 jQuery 对象）以及位置
slider.removeSlide(obj); // 移除 slide
```

#### 暂停以后重新自动播放

`pauseOnAction` 设置为 `true` 以后，用户操作 slider 就会暂停播放。

有的开发者希望过一段时间以后恢复播放，可以通过在 Data API 中设置 `playAfterPaused` 参数来实现。

`````html
<div class="am-slider am-slider-default" data-am-flexslider="{playAfterPaused: 8000}">
  <ul class="am-slides">
    <li><img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-3.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-4.jpg" /></li>
  </ul>
</div>
`````
```html
<div class="am-slider am-slider-default" data-am-flexslider="{playAfterPaused: 8000}">
  <ul class="am-slides">
    <li><img src="..." /></li>
  </ul>
</div>
```

**注意：这个参数只有通过 Data API 初始化的 Slider 才有效。** 使用 JS 初始化的请添加通过以下回调函数实现。

```js
$('#your-slider').flexslider({
  playAfterPaused: 8000,
  before: function(slider) {
    if (slider._pausedTimer) {
      window.clearTimeout(slider._pausedTimer);
      slider._pausedTimer = null;
    }
  },
  after: function(slider) {
    var pauseTime = slider.vars.playAfterPaused;
    if (pauseTime && !isNaN(pauseTime) && !slider.playing) {
      if (!slider.manualPause && !slider.manualPlay && !slider.stopped) {
        slider._pausedTimer = window.setTimeout(function() {
          slider.play();
        }, pauseTime);
      }
    }
  }
  // 设置其他参数
});
```

### 默认参数及说明

```javascript
{
  animation: "slide",             // String: ["fade"|"slide"]，动画效果
  easing: "swing",                // String: 滚动动画计时函数
  direction: "horizontal",        // String: 滚动方向 ["horizontal"|"vertical"]
  reverse: false,                 // Boolean: 翻转 slide 运动方向
  animationLoop: true,            // Boolean: 是否循环播放
  smoothHeight: false,            // Boolean: 当 slide 图片比例不一样时
                                  // "true": 父类自动适应图片高度
                                  // "false": 不自动适应，父类高度为图片的最高高度，默认为false

  startAt: 0,                     // Integer: 开始播放的 slide，从 0 开始计数
  slideshow: true,                // Boolean: 是否自动播放
  slideshowSpeed: 5000,           // Integer: ms 滚动间隔时间
  animationSpeed: 600,            // Integer: ms 动画滚动速度
  initDelay: 0,                   // Integer: ms 首次执行动画的延迟
  randomize: false,               // Boolean: 是否随机 slide 顺序

  // Usability features
  pauseOnAction: true,            // Boolean: 用户操作时停止自动播放
  pauseOnHover: false,            // Boolean: 悬停时暂停自动播放
  useCSS: true,                   // Boolean: 是否使用 css3 transition
  touch: true,                    // Boolean: 允许触摸屏触摸滑动滑块
  video: false,                   // Boolean: 使用视频的 slider，防止 CSS3 3D 变换毛刺

  // Primary Controls
  controlNav: true,               // Boolean: 是否创建控制点
  directionNav: true,             // Boolean: 是否创建上/下一个按钮（previous/next）
  prevText: "Previous",           // String: 上一个按钮文字
  nextText: "Next",               // String: 下一个按钮文字

  // Secondary Navigation
  keyboard: true,                 // Boolean: 是否开启键盘左（←）右（→）控制
  multipleKeyboard: false,        // Boolean: 是否允许键盘控制多个 slide
  mousewheel: true,               // Boolean: 是否开启鼠标滚轮控制
  pausePlay: false,               // Boolean: 是否创建暂停与播放按钮
  pauseText: 'Pause',             // String: 暂停按钮文字
  playText: 'Play',               // String: 播放按钮文字

  // Special properties
  controlsContainer: "",          // jQuery Object/Selector
  manualControls: "",             // jQuery Object/Selector 自定义控制 slider 的元素，
                                  // 如 "#tabs-nav li img"，导航数量和 slide 数量一样
  sync: "",                       // Selector: 关联 slide 与 slide 之间的操作。
  asNavFor: "",                   // Selector: Internal property exposed for turning the slider into a thumbnail navigation for another slider

  // Carousel Options
  itemWidth: 0,                   // Integer: slide 宽度，多个同时滚动时设置
  itemMargin: 0,                  // Integer: slide 间距
  minItems: 1,                    // Integer: 最少显示 slide 数, 与 `itemWidth` 相关
  maxItems: 0,                    // Integer: 最多显示 slide 数, 与 `itemWidth` 相关
  move: 0,                        // Integer: 一次滚动移动的 slide 数量，0 - 滚动可见的 slide

  // Callback API
  start: function(){},            // Callback: function(slider) - 初始化完成的回调
  before: function(){},           // Callback: function(slider) - 每次滚动开始前的回调
  after: function(){},            // Callback: function(slider) - 每次滚动完成后的回调
  end: function(){},              // Callback: function(slider) - 执行到最后一个 slide 的回调
  added: function(){},            // Callback: function(slider) - slide 添加时触发
  removed: function(){}           // Callback: function(slider) - slide 被移除时触发

  // Amaze UI 扩展参数
  playAfterPaused: null           // Integer: ms 用户停止操作多长时间以后重新开始自动播放
}
```
