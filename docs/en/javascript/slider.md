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

Create a list of items to use as a carousel slider. Support touch event. Via [FlexSlider](https://github.com/woothemes/FlexSlider)。

This plugin can hardly be counted as a light-weight plugin, but it's fully functional and can fit the developers' need in most situation.

[Slider in Web Widget](/widgets/slider?_ver=2.x) uses this plugin, and adds some style extensions to it.

**All rights of pictures used in following samples belongs to [Microsoft Bing](http://www.bing.com).**

## Examples

### Default Style

`````html
<div class="am-slider am-slider-default" data-am-flexslider>
  <ul class="am-slides">
    <li><img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-3.jpg" /></li>
    <li><img src="http://s.amazeui.org/media/i/demos/bing-4.jpg" /></li>
  </ul>
</div>
`````
```html
<div class="am-slider am-slider-default" data-am-flexslider>
  <ul class="am-slides">
    <li><img src="..." /></li>
  </ul>
</div>
```

### Thumbnail Style

`````html
<div class="am-slider am-slider-default"
     data-am-flexslider="{controlNav: 'thumbnails', directionNav: false}">
  <ul class="am-slides">
    <li data-thumb="http://amui.qiniudn.com/pure-1.jpg?imageView2/0/w/360"><img
      src="http://amui.qiniudn.com/pure-1.jpg" /></li>
    <li data-thumb="http://amui.qiniudn.com/pure-2.jpg?imageView2/0/w/360"><img
      src="http://amui.qiniudn.com/pure-2.jpg" /></li>
    <li data-thumb="http://amui.qiniudn.com/pure-3.jpg?imageView2/0/w/360"><img
      src="http://amui.qiniudn.com/pure-3.jpg" /></li>
    <li data-thumb="http://amui.qiniudn.com/pure-4.jpg?imageView2/0/w/360"><img
      src="http://amui.qiniudn.com/pure-4.jpg" /></li>

  </ul>
</div>
`````
```html
<div class="am-slider am-slider-default"
     data-am-flexslider="{controlNav: 'thumbnails', directionNav: false}">
  <ul class="am-slides">
    <li><img src="..." /></li>
  </ul>
</div>
```

### Carousel Style


Create a carousel style slider by setting `itemWidth`. `itemMargin` is only used to calculate scrolling distance. Use CSS to set the margin between two `<li>`s. Default margin is 5px. 
Carousel style will be enabled after adding `.am-slider-carousel` class to slider.s

```css
.am-slider-carousel li {
  margin-right: 5px;
}
```

`````html
<div class="am-slider am-slider-default am-slider-carousel"
     data-am-flexslider="{itemWidth: 200, itemMargin: 5}">
  <ul class="am-slides">
    <li><img src="http://amui.qiniudn.com/pure-1.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://amui.qiniudn.com/pure-2.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://amui.qiniudn.com/pure-3.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://amui.qiniudn.com/pure-4.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://amui.qiniudn.com/pure-1.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://amui.qiniudn.com/pure-2.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://amui.qiniudn.com/pure-3.jpg?imageView2/0/w/640" /></li>
    <li><img src="http://amui.qiniudn.com/pure-4.jpg?imageView2/0/w/640" /></li>
  </ul>
</div>
`````
```html
<div class="am-slider am-slider-default am-slider-carousel"
     data-am-flexslider="{itemWidth: 200, itemMargin: 5}">
  <ul class="am-slides">
    <li><img src="..." /></li>
  </ul>
</div>
```

## Usage

### HTML

HTML structure of slider __must__ be as below.
But you may use anything you want inside `<li>`.

```html
<div class="am-slider am-slider-default">
  <ul class="am-slides">
    <li>...</li>
  </ul>
</div>
```

### Initializing with Data API

Add `.data-am-flexslider` to container to initialize.

```html
<div class="am-slider am-slider-default">
  <ul class="am-slides">
    <li>...</li>
  </ul>
</div>
```

### Initializing with JS

```javascript
$('.am-slider').flexslider();
```

#### Options

Pass the value of options as parameter when initializing.

```javascript
$(function() {
  $('.am-slider').flexslider({
    // options
  });
});
```

#### Methods

- `$('#slider').flexslider('play');` - Play
- `$('#slider').flexslider('pause');` - Pause
- `$('#slider').flexslider('stop');` - Stop
- `$('#slider').flexslider('next');` - Next
- `$('#slider').flexslider('prev');` - Prev
- `$('#slider').flexslider(3);` // - Scroll to the fourth slide

#### Slider Instance

Use slider instance to call more methods.

```js
var slider = $('#slider').data('flexslider');

slider.addSlide(obj, position); // Inserted element (selector or jQuery object) and its position.
slider.removeSlide(obj); Remove slide
```

#### Autometically Resume to Play

The slider will be paused after setting `pauseOnAction` to `true`.

Some developers want the slider to resume to play after several seconds. This can be achieved by setting `playAfterPaused` in Data API.

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

**Attension: Only in sliders initialized using Data API is this option avaliable.** For those sliders initialzed with JS, please using following callback function.

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
  // Other options
});
```

### Options

```javascript
{
  animation: "slide",             // String: ["fade"|"slide"]. Animation
  easing: "swing",                // String: Easing function of Animation
  direction: "horizontal",        // String: Scrolling direction ["horizontal"|"vertical"]
  reverse: false,                 // Boolean: Reverse the scrolling direction
  animationLoop: true,            // Boolean: Looping scrolling
  smoothHeight: false,            // Boolean: When width/height ratios of images in slide are different
                                  // "true": Parent container adapt to image
                                  // "false": No adaption, The height of parent container is same as the largest height.(Default)

  startAt: 0,                     // Integer: The initail slide. The first image is 0.
  slideshow: true,                // Boolean: Auto scroll if true.
  slideshowSpeed: 5000,           // Integer: Scrolling interval in ms.
  animationSpeed: 600,            // Integer: Scrolling speed in ms.
  initDelay: 0,                   // Integer: Delay before the first scroll in ms
  randomize: false,               // Boolean: Random order

  // Usability features
  pauseOnAction: true,            // Boolean: Pause auto scroll when use is making actions.
  pauseOnHover: false,            // Boolean: Pause auto scroll when hover.
  useCSS: true,                   // Boolean: Use css3 transition or not
  touch: true,                    // Boolean: Allow touch
  video: false,                   // Boolean: Use video for slider, to avoid glitches in CSS3 3D transforation.

  // Primary Controls
  controlNav: true,               // Boolean: Create control point or not
  directionNav: true,             // Boolean: Create Prev/Next button or not
  prevText: "Previous",           // String: Text for Prev button
  nextText: "Next",               // String: Text for Next button

  // Secondary Navigation
  keyboard: true,                 // Boolean: Allow slider to be controlled by keyboard.
  multipleKeyboard: false,        // Boolean: Allow keyboard to control multiple slide
  mousewheel: true,               // Boolean: Allow slider to be controlled by mouse wheel.
  pausePlay: false,               // Boolean: Create pause/play button or not
  pauseText: 'Pause',             // String: Text for pause button
  playText: 'Play',               // String: Text for play button

  // Special properties
  controlsContainer: "",          // jQuery Object/Selector
  manualControls: "",             // jQuery Object/Selector: Customized slider element,
                                  // such as "#tabs-nav li img", the number of navs is same as the number of slides
  sync: "",                       // Selector: sync between different slides.
  asNavFor: "",                   // Selector: Internal property exposed for turning the slider into a thumbnail navigation for another slider

  // Carousel Options
  itemWidth: 0,                   // Integer: slide width
  itemMargin: 0,                  // Integer: slide margin
  minItems: 1,                    // Integer: Minimum slide item number. Related to `itemWidth`
  maxItems: 0,                    // Integer: Maximum slide item number. Related to `itemWidth`
  move: 0,                        // Integer: The number of slide moved at once. 0 - Move all visible slides.

  // Callback API
  start: function(){},            // Callback: function(slider) - After initialzation.
  before: function(){},           // Callback: function(slider) - Before scrolling.
  after: function(){},            // Callback: function(slider) - After scrolling.
  end: function(){},              // Callback: function(slider) - When scrolling to the end.
  added: function(){},            // Callback: function(slider) - When slide is added.
  removed: function(){}           // Callback: function(slider) - When slide is removed.

  // Amaze UI extension
  playAfterPaused: null           // Integer: How long does the slider wait before continue to play (in ms).
}
```
