# Animation
---

CSS3 动画封装，浏览器需支持 CSS3 动画。

<table class="am-table am-table-bd am-table-striped">
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

点击按钮查看动画效果。

## 基本效果

`````html
<div class="am-g doc-animations">
  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-primary am-btn-block" data-doc-animation="fade">Fade</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-secondary am-btn-block" data-doc-animation="scale-up">Scale Up</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-success am-btn-block" data-doc-animation="scale-down">Scale Down</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-warning am-btn-block" data-doc-animation="slide-top">Slide Top</button>
  </div>
</div>

<div class="am-g doc-animations">
  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-primary am-btn-block" data-doc-animation="slide-bottom">Slide Bottom</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-secondary am-btn-block" data-doc-animation="slide-left">Slide Left</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-success am-btn-block" data-doc-animation="slide-right">Slide Right</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-warning am-btn-block" data-doc-animation="shake">Shake</button>
  </div>
</div>

<span class="am-icon-cog am-animation-spin"></span>
`````
```html
<div class="am-animation-fade">...</div>

<span class="am-icon-cog am-animation-spin"></span>
```

## 反向动画

添加 `.am-animation-reverse` class，让动画反向运行（通过把 [`animation-direction`](https://developer.mozilla.org/zh-CN/docs/Web/CSS/animation-direction) 设置为 `reverse` 实现）。

`````html
<div class="am-g doc-animations">
  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-primary am-btn-block am-animation-reverse" data-doc-animation="fade">Fade</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-secondary am-btn-block am-animation-reverse" data-doc-animation="scale-up">Scale Up</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-success am-btn-block am-animation-reverse" data-doc-animation="scale-down">Scale Down</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-warning am-btn-block am-animation-reverse" data-doc-animation="slide-top">Slide Top</button>
  </div>
</div>

<div class="am-g doc-animations">
  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-primary am-btn-block am-animation-reverse" data-doc-animation="slide-bottom">Slide Bottom</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-secondary am-btn-block am-animation-reverse" data-doc-animation="slide-left">Slide Left</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-success am-btn-block am-animation-reverse" data-doc-animation="slide-right">Slide Right</button>
  </div>

  <div class="col-sm-6 col-md-3">
    <button class="am-btn am-btn-warning am-btn-block am-animation-reverse" data-doc-animation="shake">Shake</button>
  </div>
</div>

<span class="am-icon-cog am-animation-spin am-animation-reverse"></span>
<script>
  Zepto(function($) {
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

## 参考资源

- [CSS 动画](https://developer.mozilla.org/zh-CN/docs/Web/CSS/animation)
- [Animate.css](http://daneden.github.io/animate.css/)