# Slider
---

图片轮播模块，源自 [FlexSlider](https://github.com/woothemes/FlexSlider)。

## 基本使用

`````html
<div class="am-slider am-slider-default">
  <ul class="am-slides">
    <li><img src="http://cn.bing.com/az/hprichv/LondonTrainStation_GettyRR_139321755_ZH-CN742316019.jpg" /></li>
    <li><img src="http://s.cn.bing.net/az/hprichbg/rb/CardinalsBerries_ZH-CN10679090179_1366x768.jpg"></li>
    <li><img src="http://s.cn.bing.net/az/hprichbg/rb/QingdaoJiaozhou_ZH-CN10690497202_1366x768.jpg"></li>
    <li><img src="http://s.cn.bing.net/az/hprichbg/rb/FennecFox_ZH-CN13720911949_1366x768.jpg"></li>
  </ul>
</div>
<script>
$(function() {
  $('.am-slider').flexslider();
});
</script>
`````

### HTML 结构

以下结构 __必须__，`<li>` 中的内容可以自由组合。

```html
<div class="am-slider am-slider-default">
  <ul class="am-slides">
    <li>...</li>
  </ul>
</div>
```

### 初始化 JS

```javascript
Zepto(function($) {
  $('.am-slider').flexslider();
});
```

## 自定参数

在初始化函数中传入所需参数即可。

```javascript
$(function() {
  $('.am-slider').flexslider({
    // options
  });
});
```

## 所有参数

```javascript
{
  namespace: "am-",               //{NEW} String: class的前缀字符串。
  selector: ".am-slides > li",    //{NEW} Selector: Must match a simple pattern. '{container} > {slide}' -- Ignore pattern at your own peril
  animation: "slide",             //String: "fade" or "slide", AmazeUI 默认使用"slide"滑动，可以选择 "fade": 淡入淡出。
  easing: "swing",                //{NEW} String: Determines the easing method used in jQuery transitions. jQuery easing plugin is supported!
  direction: "horizontal",        *//String: "horizontal" or "vertical" 选择slide滚动形式，"horizontal": 为水平滚动，"vertical": 上下滚动。
  reverse: false,                 *//{NEW} Boolean: true or false 翻转slide子项运动方向，"false": 默认为正常运动方向, "true": 为相反方向。
  animationLoop: true,            *官方bug//Boolean: true or false Should the animation loop? If false, directionNav will received "disable" classes at either end
  smoothHeight: false,            *-//{NEW} Boolean: true or false 当slide图片比例不一样时，"true": 父类自动适应图片高度，"false": 不自动适应，父类高度为图片的最高高度，默认为false。
  startAt: 0,                     //Integer: 开始显示的slide顺序，0为第一张slide。
  slideshow: true,                //Boolean: true or false 是否自动播放，默认为true。
  slideshowSpeed: 5000,           //Integer: 单位为 ms 自动播放的间隔时间。
  animationSpeed: 600,            //Integer: 单位为 ms 动画运动时间，配合参数slideshowSpeed可以做出一直滚动效果。
  initDelay: 0,                   //{NEW} Integer: 单位为 ms 首次执行动画的延迟时间，默认为0
  randomize: false,               *官方bug//Boolean: true or false 是否随机slide顺序，默认为false。

  // Usability features
  pauseOnAction: true,            *//Boolean: true or false 作用于控制点暂停自动播放程序。
  pauseOnHover: false,            //Boolean: true or false 悬停在slide上时，暂停自动播放程序。
  useCSS: true,                   //{NEW} Boolean: true or false 开启使用css3移动。
  touch: true,                    //{NEW} Boolean: true or false 允许触摸屏触摸滑动滑块。
  video: false,                   //{NEW} Boolean: true or false 如果使用视频的滑块，可以防止CSS3的3D变换避免毛刺。

  // Primary Controls
  controlNav: true,               //Boolean: true or false 是否创建控制点，方便控制滑块。
  directionNav: true,             //Boolean: true or false 是否创建上一个和下一个控制点（previous/next）control。
  prevText: "Previous",           //String: 设置上一个控制点文本，默认为"previous"
  nextText: "Next",               //String: 设置下一个控制点文本，默认为"next"

  // Secondary Navigation
  keyboard: true,                 //Boolean: true or false 允许开启键盘左（←）右（→）控制滑块滑动。
  multipleKeyboard: false,        //{NEW} Boolean: true or false 允许键盘控制多个slide，默认为 false 不允许控制多个slide。
  mousewheel: true,               //{UPDATED} Boolean: 是否开启鼠标滚轮控制 slide 滑动。
  pausePlay: false,               //Boolean: true or false 是否创建暂停与开启自动播放的控件。
  pauseText: 'Pause',             //String: 设置暂停控件的文本。默认为"pause"
  playText: 'Play',               //String: 设置播放控件的文本。默认为"play"

  // Special properties
  controlsContainer: "",          //{UPDATED} Zepto Object/Selector 控制容器，声明的容器的导航元素比原容器大。默认的容器是flexslider元。如果给定的元素属性是不存在忽视。
  manualControls: "",             //{UPDATED} Zepto Object/Selector 声明自定义导航控件。例如".flex-control-nav li" or "#tabs-nav li img"，使用自定义导航控件，需要把导航数量和滑块数量相等。
  sync: "",                       //{NEW} Selector: 关联slide与slide之间的操作。
  asNavFor: "",                   *//{NEW} Selector: Internal property exposed for turning the slider into a thumbnail navigation for another slider

  // Carousel Options
  itemWidth: 0,                   //{NEW} Integer: 滑块的宽度，盒模型包含horizontal borders and padding
  itemMargin: 0,                  *官方bug//{NEW} Integer: 两个滑块之间间隔距离。
  minItems: 1,                    //{NEW} Integer: 最少显示滑块的可见数, 与参数itemWidth相关。
  maxItems: 0,                    //{NEW} Integer: 最多显示滑块的可见数, 与参数itemWidth相关。
  move: 0,                        *//{NEW} Integer: Number of carousel items that should move on animation. If 0, slider will move all visible items.

  // Callback API
  start: function(){},            //Callback: function(slider) - slide 初始化完成时的回调。
  before: function(){},           //Callback: function(slider) - 每次动画完成前的回调。
  after: function(){},            //Callback: function(slider) - 每次动画完成后的回调。
  end: function(){},              //Callback: function(slider) - slide 动画执行到最后一个元素时的回调，与start相反。
  added: function(){},            *//{NEW} Callback: function(slider) - Fires after a slide is added
  removed: function(){}           *//{NEW} Callback: function(slider) - Fires after a slide is removed
}
```
