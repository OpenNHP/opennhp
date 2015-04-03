---
id: tabs
title: 选项卡
titleEn: Tabs
prev: javascript/sticky.html
next: javascript/datepicker.html
source: js/ui.tabs.js
doc: docs/javascript/tabs.md
---

# Tabs
---

选项卡插件，基于 CSS3 `transition` 实现水平平滑滚动；CSS3 动画实现回弹效果（触控操作时）。

## 使用演示

### 内容高度一致

`````html
<div class="am-tabs" data-am-tabs="{noSwipe: 1}">
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="javascript: void(0)">流浪</a></li>
    <li><a href="javascript: void(0)">流浪</a></li>
    <li><a href="javascript: void(0)">再流浪</a></li>
  </ul>

  <div class="am-tabs-bd">
    <div class="am-tab-panel am-active">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？
    </div>
    <div class="am-tab-panel">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？
    </div>
    <div class="am-tab-panel">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？
    </div>
  </div>
</div>
`````

```html
<div class="am-tabs" data-am-tabs>
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="javascript: void(0)">流浪</a></li>
    <li><a href="javascript: void(0)">流浪</a></li>
    <li><a href="javascript: void(0)">再流浪</a></li>
  </ul>

  <div class="am-tabs-bd">
    <div class="am-tab-panel am-active">
      ...
    </div>
    <div class="am-tab-panel">
      ...
    </div>
    <div class="am-tab-panel">
      ...
    </div>
  </div>
</div>
```

### 自适应内容高度

`````html
<div class="am-tabs" data-am-tabs>
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="#tab1">恣意</a></li>
    <li><a href="#tab2">等候</a></li>
    <li><a href="#tab3">流浪</a></li>
  </ul>

  <div class="am-tabs-bd">
    <div class="am-tab-panel am-fade am-in am-active" id="tab1">
      置身人群中<br>你只需要被淹没 享受 沉默<br>退到人群后<br>你只需给予双手 微笑 等候
    </div>
    <div class="am-tab-panel am-fade" id="tab2">
      走在忠孝东路<br>徘徊在茫然中<br>在我的人生旅途<br>选择了多少错误<br>我在睡梦中惊醒<br>感叹悔言无尽<br>恨我不能说服自己<br>接受一切教训<br>让生命去等候<br>等候下一个漂流<br>让生命去等候<br>等候下一个伤口
    </div>
    <div class="am-tab-panel am-fade" id="tab3">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？
    </div>
  </div>
</div>
`````
```html
<div class="am-tabs" data-am-tabs>
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="#tab1">恣意</a></li>
    <li><a href="#tab2">等候</a></li>
    <li><a href="#tab3">流浪</a></li>
  </ul>

  <div class="am-tabs-bd">
    <div class="am-tab-panel am-fade am-in am-active" id="tab1">
      置身人群中<br>你只需要被淹没 享受 沉默<br>退到人群后<br>你只需给予双手 微笑 等候
    </div>
    <div class="am-tab-panel am-fade" id="tab2">
      走在忠孝东路<br>徘徊在茫然中<br>在我的人生旅途<br>选择了多少错误<br>我在睡梦中惊醒<br>感叹悔言无尽<br>恨我不能说服自己<br>接受一切教训<br>让生命去等候<br>等候下一个漂流<br>让生命去等候<br>等候下一个伤口
    </div>
    <div class="am-tab-panel am-fade" id="tab3">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？
    </div>
  </div>
</div>
```

### 禁用触控操作

部分用户反应在过长的 Tabs 中滚动页面时会意外触发 Tab 切换事件，用户可以选择禁用触控操作。

`````html
<div class="am-tabs" data-am-tabs="{noSwipe: 1}">
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="#tab2-1">恣意</a></li>
    <li><a href="#tab2-2">等候</a></li>
    <li><a href="#tab2-3">流浪</a></li>
  </ul>

  <div class="am-tabs-bd">
    <div class="am-tab-panel am-fade am-in am-active" id="tab2-1">
      置身人群中<br>你只需要被淹没 享受 沉默<br>退到人群后<br>你只需给予双手 微笑 等候
    </div>
    <div class="am-tab-panel am-fade" id="tab2-2">
      走在忠孝东路<br>徘徊在茫然中<br>在我的人生旅途<br>选择了多少错误<br>我在睡梦中惊醒<br>感叹悔言无尽<br>恨我不能说服自己<br>接受一切教训<br>让生命去等候<br>等候下一个漂流<br>让生命去等候<br>等候下一个伤口
    </div>
    <div class="am-tab-panel am-fade" id="tab2-3">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？
    </div>
  </div>
</div>
`````
```html
<div class="am-tabs" data-am-tabs="{noSwipe: 1}">
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="#tab2-1">恣意</a></li>
    <li><a href="#tab2-2">等候</a></li>
    <li><a href="#tab2-3">流浪</a></li>
  </ul>

  <div class="am-tabs-bd">
    ...
  </div>
</div>
```

## 调用方式

### 通过 Data API

在选项卡容器 `.am-tabs` 上添加 `data-am-tabs` 属性。上面的演示即通过此种方式调用。

```html
<div class="am-tabs" data-am-tabs>
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="#tab1">恣意</a></li>
    <li><a href="#tab2">等候</a></li>
    <li><a href="#tab3">流浪</a></li>
  </ul>

  <div class="am-tabs-bd">
    <div class="am-tab-panel am-active" id="tab1">...</div>
    <div class="am-tab-panel" id="tab2">...</div>
    <div class="am-tab-panel" id="tab3">...</div>
  </div>
</div>
```

### 通过 JS

#### 方法

通过 `$().tabs(options)` 开启选项卡的交互功能。

`````html
<div class="am-tabs" id="doc-my-tabs">
  <ul class="am-tabs-nav am-nav am-nav-tabs am-nav-justify">
    <li class="am-active"><a href="">彩虹</a></li>
    <li><a href="">画面</a></li>
    <li><a href="">窗外</a></li>
  </ul>
  <div class="am-tabs-bd">
    <div class="am-tab-panel am-active">外面是个下雨天，不由得就会想念</div>
    <div class="am-tab-panel">像重复的广告片，总是闪烁的画面，那么地熟悉</div>
    <div class="am-tab-panel">像窗外的雨，绿油油的树叶，自由地在说笑</div>
  </div>
</div>
<script>
  $(function() {
    var $myTabs = $('#doc-my-tabs');

    $myTabs.tabs();

    $myTabs.find('a').on('opened.tabs.amui', function(e) {
      console.log('[%s] 选项卡打开了', $(this).text());
    })
  })
</script>
`````
```html
<div class="am-tabs" id="doc-my-tabs">
  <ul class="am-tabs-nav am-nav am-nav-tabs am-nav-justify">
    <li class="am-active"><a href="">彩虹</a></li>
    <li><a href="">画面</a></li>
    <li><a href="">窗外</a></li>
  </ul>
  <div class="am-tabs-bd">
    <div class="am-tab-panel am-active">...</div>
    <div class="am-tab-panel">...</div>
    <div class="am-tab-panel">...</div>
  </div>
</div>
<script>
  $(function() {
    $('#doc-my-tabs').tabs();
  })
</script>
```

#### 选项

- `options.noSwipe` 是否禁用触控事件。

```js
$('#someTabs').tabs({noSwipe: 1});
```
#### 自定义事件

自定义事件触发在标签上。

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.tabs.amui</code></td>
    <td>打开一个选项卡时立即触发</td>
  </tr>
  <tr>
    <td><code>opened.tabs.amui</code></td>
    <td>选项卡打开完成时触发（CSS 动画执行完成）</td>
  </tr>
  </tbody>
</table>

```javascript
$('#doc-my-tabs').find('a').on('opened.tabs.amui', function(e) {
  console.log('[%s] 选项卡打开了', $(this).text());
})
```

打开控制台操作[上面的选项卡](#doc-my-tabs)查看事件监听输出的文字。

## FAQ

### Tab 内容不能选择，如何处理？

这个问题由 [Hammer.js](http://hammerjs.github.io/tips/) 引起。

> Hammer is setting a property to improve the UX of the panning on desktop.

可以使用下面的样式覆盖掉 Hammer.js 的样式：

```css
.am-tabs-bd {
  -moz-user-select: text !important;
  -webkit-user-select: text !important;
  -ms-user-select: text !important;
  user-select: text !important;
}
```

也可以选择禁用触控事件。

<!--
## TODO

- Ajax 载入选项卡内容支持
-->
