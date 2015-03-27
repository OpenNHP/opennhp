# Animation
---

CSS3 动画封装，浏览器需支持 CSS3 动画。

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th>Class</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.am-animation-fade</code></td>
    <td>淡入</td>
  </tr>
  <tr>
    <td><code>.am-animation-scale-up</code></td>
    <td>逐渐放大</td>
  </tr>
  <tr>
    <td><code>.am-animation-scale-down</code></td>
    <td>逐渐缩小</td>
  </tr>
  <tr>
    <td><code>.am-animation-slide-top</code></td>
    <td>顶部滑入</td>
  </tr>
  <tr>
    <td><code>.am-animation-slide-bottom</code></td>
    <td>底部滑入</td>
  </tr>
  <tr>
    <td><code>.am-animation-slide-left</code></td>
    <td>左侧滑入</td>
  </tr>
  <tr>
    <td><code>.am-animation-slide-right</code></td>
    <td>右侧滑入</td>
  </tr>
  <tr>
    <td><code>.am-animation-shake</code></td>
    <td>左右摇动</td>
  </tr>
  <tr>
    <td><code>.am-animation-spin</code></td>
    <td> 无限旋转</td>
  </tr>
  </tbody>
</table>


## 使用演示

点击按钮查看动画效果。

### 默认效果

`````html
<div class="am-g doc-animations">
  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-primary am-btn-block" data-doc-animation="fade">Fade</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-secondary am-btn-block" data-doc-animation="scale-up">Scale Up</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-success am-btn-block" data-doc-animation="scale-down">Scale Down</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-warning am-btn-block" data-doc-animation="slide-top">Slide Top</button>
  </div>
</div>

<div class="am-g doc-animations">
  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-primary am-btn-block" data-doc-animation="slide-bottom">Slide Bottom</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-secondary am-btn-block" data-doc-animation="slide-left">Slide Left</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-success am-btn-block" data-doc-animation="slide-right">Slide Right</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-warning am-btn-block" data-doc-animation="shake">Shake</button>
  </div>
</div>

<span class="am-icon-cog am-animation-spin"></span>
`````
```html
<div class="am-animation-fade">...</div>

<span class="am-icon-cog am-animation-spin"></span>
```

### 反向动画

添加 `.am-animation-reverse` class，让动画反向运行（通过把 [`animation-direction`](https://developer.mozilla.org/zh-CN/docs/Web/CSS/animation-direction) 设置为 `reverse` 实现）。

`````html
<div class="am-g doc-animations">
  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-primary am-btn-block am-animation-reverse" data-doc-animation="fade">Fade</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-secondary am-btn-block am-animation-reverse" data-doc-animation="scale-up">Scale Up</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-success am-btn-block am-animation-reverse" data-doc-animation="scale-down">Scale Down</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-warning am-btn-block am-animation-reverse" data-doc-animation="slide-top">Slide Top</button>
  </div>
</div>

<div class="am-g doc-animations">
  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-primary am-btn-block am-animation-reverse" data-doc-animation="slide-bottom">Slide Bottom</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-secondary am-btn-block am-animation-reverse" data-doc-animation="slide-left">Slide Left</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-success am-btn-block am-animation-reverse" data-doc-animation="slide-right">Slide Right</button>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <button class="am-btn am-btn-warning am-btn-block am-animation-reverse" data-doc-animation="shake">Shake</button>
  </div>
</div>

<span class="am-icon-cog am-animation-spin am-animation-reverse"></span>
<script>
$(function() {
  $('.doc-animations').on('click', '[data-doc-animation]', function() {

    var $clicked = $(this),
      animation = 'am-animation-' + $clicked.data('docAnimation');

    if($clicked.data('animation-idle')) {
      clearTimeout($clicked.data('animation-idle'));
    }

    $clicked.removeClass(animation);

    setTimeout(function(){
      $clicked.addClass(animation);
      $clicked.data('animation-idle', setTimeout(function(){
        $clicked.removeClass(animation);
        $clicked.data('animation-idle', false);
      }, 500));
    }, 50);
  });
});
</script>
`````
```html
<div class="am-animation-fade am-animation-reverse">...</div>

<span class="am-icon-cog am-animation-spin am-animation-reverse"></span>
```

### 动画延迟执行

添加以下 class 可以使动画延迟 `1-6s` 执行。

- `.am-animation-delay-1`
- `.am-animation-delay-2`
- `.am-animation-delay-3`
- `.am-animation-delay-4`
- `.am-animation-delay-5`
- `.am-animation-delay-6`

自定义延时：

```css
.my-animation-delay {
  -webkit-animation-delay: 15s;
  animation-delay: 15s;
}
```

`````html
<button id="animation-start" type="button" class="am-btn am-btn-danger">点击开始执行动画</button>

<hr/>

<div id="animation-group">
<p><button type="button" class="am-btn am-btn-primary">没延迟的动画</button></p>

<p><button type="button" class="am-btn am-btn-primary am-animation-delay-1">延迟 1s 执行</button></p>
<p><button type="button" class="am-btn am-btn-secondary am-animation-delay-2">延迟 2s 执行</button></p>
<p><button type="button" class="am-btn am-btn-success am-animation-delay-3">延迟 3s 执行</button></p>
<p><button type="button" class="am-btn am-btn-warning am-animation-delay-4">延迟 4s 执行</button></p>
<p><button type="button" class="am-btn am-btn-danger am-animation-delay-5">延迟 5s 执行</button></p>
<p><button type="button" class="am-btn am-btn-primary am-animation-delay-6">延迟 6s 执行</button></p>
</div>

<script>
  $(function() {
    var $btns = $('#animation-group').find('.am-btn');
    var dfds = [];
    var animating = false;
    var animation = 'am-animation-scale-up';

    $('#animation-start').on('click', function() {
      if (!animating) {
        animating = true;
        $btns.each(function() {
          var dfd = new $.Deferred();
          dfds.push(dfd);
          var $this = $(this);
          if ($.AMUI.support.animation) {
            $this.addClass(animation).one($.AMUI.support.animation.end, function() {
              $this.removeClass(animation);
              dfd.resolve();
            });
          }
        });

        $.when.apply(null, dfds).done(function() {
          animating = false;
          console.log('[AMUI] - 所有动画执行完成');
          dfds = [];
        });
      }
    });
  });
</script>
`````

```html
<button id="animation-start" type="button" class="am-btn am-btn-danger">点击开始执行动画</button>

<hr/>

<div id="animation-group">
<p><button type="button" class="am-btn am-btn-primary">没延迟的动画</button></p>

<p><button type="button" class="am-btn am-btn-primary am-animation-delay-1">延迟 1s 执行</button></p>
<p><button type="button" class="am-btn am-btn-secondary am-animation-delay-2">延迟 2s 执行</button></p>
<p><button type="button" class="am-btn am-btn-success am-animation-delay-3">延迟 3s 执行</button></p>
<p><button type="button" class="am-btn am-btn-warning am-animation-delay-4">延迟 4s 执行</button></p>
<p><button type="button" class="am-btn am-btn-danger am-animation-delay-5">延迟 5s 执行</button></p>
<p><button type="button" class="am-btn am-btn-primary am-animation-delay-6">延迟 6s 执行</button></p>
</div>

<script>
  $(function() {
    var $btns = $('#animation-group').find('.am-btn');
    var dfds = [];
    var animating = false;
    var animation = 'am-animation-scale-up';

    $('#animation-start').on('click', function() {
      if (!animating) {
        animating = true;
        $btns.each(function() {
          var dfd = new $.Deferred();
          dfds.push(dfd);
          var $this = $(this);
          if ($.AMUI.support.animation) {
            $this.addClass(animation).one($.AMUI.support.animation.end, function() {
              $this.removeClass(animation);
              dfd.resolve();
            });
          }
        });

        $.when.apply(null, dfds).done(function() {
          animating = false;
          console.log('[AMUI] - 所有动画执行完成');
          dfds = [];
        });
      }
    });
  });
</script>
```


## 参考资源

- [CSS 动画](https://developer.mozilla.org/zh-CN/docs/Web/CSS/animation)
- [Animate.css](http://daneden.github.io/animate.css/)
