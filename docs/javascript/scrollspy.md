# ScrollSpy
---

窗口滚动时为根据设置页面元素添加动画效果（仅在支持 CSS3 动画的浏览器上有效）。

## 使用方法

添加 `data-am-scrollspy` 属性，并设置相关属性。
    
<table class="am-table am-table-bd am-table-striped">
  <thead>
    <tr>
      <th>属性</th>
      <th>描述</th>
    </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>data-am-scrollspy="{animation:'fade'}"</code></td>
    <td>动画类名，参见 <a href="/css/animation">Amaze UI 动画库</a>，默认为 <code>fade</code></td>
  </tr>
  <tr>
    <td><code>data-am-scrollspy="{animation:'fade', delay: 300}"</code></td>
    <td>延迟动画执行时间（ms），默认为 <code>0</code></td>
  </tr>
  <tr>
    <td><code>data-am-scrollspy="{animation:'fade', repeat: false}"</code></td>
    <td>是否重复动画，默认为 <code>true</code></td>
  </tr>
  </tbody>
</table>

`````html
<div class="am-g">
  <div class="col-sm-12 col-md-4 col-lg-3">
    <div class="am-panel am-panel-primary" data-am-scrollspy="{animation: 'fade'}">
      <div class="am-panel-hd">Fade Animation</div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-12 col-md-4 col-lg-3 col-md-offset-4 col-lg-offset-3">
    <div class="am-panel am-panel-secondary" data-am-scrollspy="{animation: 'scale-up'}">
      <div class="am-panel-hd">Scale-up
      </div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-12 col-md-4 col-lg-3 col-md-offset-8 col-lg-offset-6">
    <div class="am-panel am-panel-success" data-am-scrollspy="{animation: 'scale-down'}">
      <div class="am-panel-hd">Scale-down
      </div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-12 col-md-4 col-lg-3 col-md-offset-8 col-lg-offset-9">
    <div class="am-panel am-panel-warning" data-am-scrollspy="{animation: 'slide-top'}">
      <div class="am-panel-hd">Slide Top
      </div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-12 col-md-4 col-lg-3 col-md-offset-8 col-lg-offset-6">
    <div class="am-panel am-panel-danger" data-am-scrollspy="{animation: 'slide-bottom'}">
      <div class="am-panel-hd">Slide Bottom
      </div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-12 col-md-4 col-lg-3 col-md-offset-4 col-lg-offset-3">
    <div class="am-panel am-panel-primary" data-am-scrollspy="{animation: 'slide-right'}">
      <div class="am-panel-hd">Slide Right
      </div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-12 col-md-4 col-lg-3">
    <div class="am-panel am-panel-secondary" data-am-scrollspy="{animation: 'slide-left'}">
      <div class="am-panel-hd">Slide Left</div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-12 col-md-4 col-lg-3">
    <div class="am-panel am-panel-success" data-am-scrollspy="{animation: 'fade'}">
      <div class="am-panel-hd">Fade</div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>

  <div class="col-sm-12 col-md-4 col-lg-3">
    <div class="am-panel am-panel-warning" data-am-scrollspy="{animation: 'fade', delay: 300}">
      <div class="am-panel-hd">Fade delay: 300</div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>

  <div class="col-sm-12 col-md-4 col-lg-3">
    <div class="am-panel am-panel-danger" data-am-scrollspy="{animation: 'fade', delay: 600}">
      <div class="am-panel-hd">Fade delay: 600
      </div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>

  <div class="col-sm-12 col-md-4 col-lg-3">
    <div class="am-panel am-panel-primary" data-am-scrollspy="{animation: 'fade', delay: 900}">
      <div class="am-panel-hd">Fade delay: 900
      </div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>
`````
```html
<div class="am-panel am-panel-default" data-am-scrollspy="{animation: 'fade'}">...</div>

<div class="am-panel am-panel-default" data-am-scrollspy="{animation: 'fade', delay: 300}">...</div>
```

## JS 调用

通过 `$().scrollspy(options)` 设置，参数同上。


`````html
<div class="am-panel am-panel-primary" id="my-scrollspy">
  <div class="am-panel-hd">ScrollSpy via JS
  </div>
  <div class="am-panel-bd">
    生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
  </div>
</div>
<script>
  seajs.use(['ui.scrollspy'], function() {
    $(function() {
      $('#my-scrollspy').scrollspy({
        animation: 'slide-left',
        delay: 500
      })
    });
  });
</script>
`````
```html
<div class="am-panel am-panel-primary" id="my-scrollspy">
  <div class="am-panel-hd">ScrollSpy via JS
  </div>
  <div class="am-panel-bd">
    生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
  </div>
</div>
<script>
$(function() {
  $('#my-scrollspy').scrollspy({
    animation: 'slide-left',
    delay: 500
  })
});
</script>
```