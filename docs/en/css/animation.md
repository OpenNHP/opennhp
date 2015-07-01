# Animation
---

CSS3 Animation, require support from browser.

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th>Class</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.am-animation-fade</code></td>
    <td>Fade</td>
  </tr>
  <tr>
    <td><code>.am-animation-scale-up</code></td>
    <td>Scale Up</td>
  </tr>
  <tr>
    <td><code>.am-animation-scale-down</code></td>
    <td>Scale Down</td>
  </tr>
  <tr>
    <td><code>.am-animation-slide-top</code></td>
    <td>Slide from Top</td>
  </tr>
  <tr>
    <td><code>.am-animation-slide-bottom</code></td>
    <td>Slide from Bottom</td>
  </tr>
  <tr>
    <td><code>.am-animation-slide-left</code></td>
    <td>Slide from Left</td>
  </tr>
  <tr>
    <td><code>.am-animation-slide-right</code></td>
    <td>Slide from Right</td>
  </tr>
  <tr>
    <td><code>.am-animation-shake</code></td>
    <td>Shake</td>
  </tr>
  <tr>
    <td><code>.am-animation-spin</code></td>
    <td>Infinitly spin</td>
  </tr>
  </tbody>
</table>


## Sample

Click on the buttons to see animations.

### Default effects

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

### Reverse Animations

Using `.am-animation-reverse` class, we can reverse the animation (by setting[`animation-direction`](https://developer.mozilla.org/zh-CN/docs/Web/CSS/animation-direction) to `reverse`).

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

### Delay Animation

Delay Animations using following classes.

- `.am-animation-delay-1`
- `.am-animation-delay-2`
- `.am-animation-delay-3`
- `.am-animation-delay-4`
- `.am-animation-delay-5`
- `.am-animation-delay-6`

Customized delay:

```css
.my-animation-delay {
  -webkit-animation-delay: 15s;
  animation-delay: 15s;
}
```

`````html
<button id="animation-start" type="button" class="am-btn am-btn-danger">Click Here to play animations</button>

<hr/>

<div id="animation-group">
<p><button type="button" class="am-btn am-btn-primary">No delay</button></p>

<p><button type="button" class="am-btn am-btn-primary am-animation-delay-1">Delay 1s</button></p>
<p><button type="button" class="am-btn am-btn-secondary am-animation-delay-2">Delay 2s</button></p>
<p><button type="button" class="am-btn am-btn-success am-animation-delay-3">Delay 3s</button></p>
<p><button type="button" class="am-btn am-btn-warning am-animation-delay-4">Delay 4s</button></p>
<p><button type="button" class="am-btn am-btn-danger am-animation-delay-5">Delay 5s</button></p>
<p><button type="button" class="am-btn am-btn-primary am-animation-delay-6">Delay 6s</button></p>
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
          console.log('[AMUI] - All animations end');
          dfds = [];
        });
      }
    });
  });
</script>
`````

```html
<button id="animation-start" type="button" class="am-btn am-btn-danger">Click Here to play animations</button>

<hr/>

<div id="animation-group">
<p><button type="button" class="am-btn am-btn-primary">No delay</button></p>

<p><button type="button" class="am-btn am-btn-primary am-animation-delay-1">Delay 1s</button></p>
<p><button type="button" class="am-btn am-btn-secondary am-animation-delay-2">Delay 2s</button></p>
<p><button type="button" class="am-btn am-btn-success am-animation-delay-3">Delay 3s</button></p>
<p><button type="button" class="am-btn am-btn-warning am-animation-delay-4">Delay 4s</button></p>
<p><button type="button" class="am-btn am-btn-danger am-animation-delay-5">Delay 5s</button></p>
<p><button type="button" class="am-btn am-btn-primary am-animation-delay-6">Delay 6s</button></p>
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
          console.log('[AMUI] - All animations end');
          dfds = [];
        });
      }
    });
  });
</script>
```


## Reference

- [CSS Animation](https://developer.mozilla.org/zh-CN/docs/Web/CSS/animation)
- [Animate.css](http://daneden.github.io/animate.css/)
