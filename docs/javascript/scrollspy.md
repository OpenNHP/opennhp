---
id: scrollspy
title: 滚动侦测
titleEn: ScrollSpy
prev: javascript/offcanvas.html
next: javascript/scrollspynav.html
source: js/ui.scrollspy.js
doc: docs/javascript/scrollspy.md
---

# ScrollSpy
---

窗口滚动时为根据设置页面元素添加动画效果（仅在支持 CSS3 动画的浏览器上有效）。

## 使用演示

下面的演示中包含了各种动画效果。

`````html
<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-primary" data-am-scrollspy="{animation: 'fade'}">
      <div class="am-panel-hd">Fade</div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-4 am-u-lg-offset-3">
    <div class="am-panel am-panel-secondary" data-am-scrollspy="{animation: 'scale-up'}">
      <div class="am-panel-hd">Scale-up</div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-8 am-u-lg-offset-6">
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
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-8 am-u-lg-offset-9">
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
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-8 am-u-lg-offset-6">
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
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-4 am-u-lg-offset-3">
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
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-secondary" data-am-scrollspy="{animation: 'slide-left'}">
      <div class="am-panel-hd">Slide Left</div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-success" data-am-scrollspy="{animation: 'fade'}">
      <div class="am-panel-hd">Fade</div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>

  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-warning" data-am-scrollspy="{animation: 'fade', delay: 300}">
      <div class="am-panel-hd">Fade delay: 300</div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>

  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-danger" data-am-scrollspy="{animation: 'fade', delay: 600}">
      <div class="am-panel-hd">Fade delay: 600
      </div>
      <div class="am-panel-bd">
        生命是一团欲望，欲望不满足便痛苦，满足便无聊。人生就在痛苦和无聊之间摇摆。——叔本华
      </div>
    </div>
  </div>

  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
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

## 调用方式

### 通过 Data API

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

### JS 调用

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
$(function() {
  $('#my-scrollspy').scrollspy({
    animation: 'slide-left',
    delay: 500
  })
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
#### 自定义事件

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>inview.scrollspy.amui</code></td>
    <td>元素进入窗口可视区域时触发</td>
  </tr>
  <tr>
    <td><code>outview.scrollspy.amui</code></td>
    <td>元素离开窗口可视区域时触发</td>
  </tr>
  </tbody>
</table>

<script>
  $(function() {
    $('#my-scrollspy').on('inview.scrollspy.amui', function() {
      console.log('进入视口');
    }).on('outview.scrollspy.amui', function() {
      console.log('离开视口');
    });
  });
</script>

```javascript
$(function() {
  $('#my-scrollspy').on('inview.scrollspy.amui', function() {
    console.log('进入视口');
  }).on('outview.scrollspy.amui', function() {
    console.log('离开视口');
  });
});
```

#### MutationObserver

通过 [Mutation Observer](https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver) 可以实现动态加载元素的动画效果。

`````html
<p><button class="am-btn am-btn-primary" id="doc-scrollspy-insert">插入</button></p>
<div id="doc-scrollspy-wrapper" data-am-observe>
<p>在下面插入一些元素试试：</p>
</div>
<script>
  $(function() {
    var i = 1;
    var $wrapper = $('#doc-scrollspy-wrapper');
    var appendPanel = function(index) {
      var panel = '<div class="am-panel am-panel-primary" ' +
        'data-am-scrollspy="{animation: \'scale-up\'}">' +
        '<div class="am-panel-bd">我是第' + index + '个插入的元素。</div></div>';
      $wrapper.append(panel);
    };

    $('#doc-scrollspy-insert').on('click', function() {
      appendPanel(i);
      i++;
    });
  });
</script>
`````
```html
<p><button class="am-btn am-btn-primary" id="doc-scrollspy-insert">插入</button></p>
<div id="doc-scrollspy-wrapper" data-am-observe>
  <p>在下面插入一些元素试试：</p>
</div>
```
```javascript
$(function() {
  var i = 1;
  var $wrapper = $('#doc-scrollspy-wrapper');
  var appendPanel = function(index) {
    var panel = '<div class="am-panel am-panel-primary" ' +
      'data-am-scrollspy="{animation: \'scale-up\'}">' +
      '<div class="am-panel-bd">我是第' + index + '个插入的元素。</div></div>';
    $wrapper.append(panel);
  };

  $('#doc-scrollspy-insert').on('click', function() {
    appendPanel(i);
    i++;
  });
});
```

