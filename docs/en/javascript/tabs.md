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

Tabs plugin is based on CSS3 `transition` and CSS3 animate.

## Example

### Same height

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

### Adaptive Height

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

### Disable Touch

Some users said tab switch event can be accidently triggered when scrolling a tab that is too long, so we allow user to disable touch events.

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

## Usage

### Using Data API

Add `data-am-tabs` to tab container `.am-tabs`.

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

### Using JS

#### Methods

Initialize tabs by using `$().tabs(options)`.

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

#### Options

- `options.noSwipe` Whether disable touch event.

```js
$('#someTabs').tabs({noSwipe: 1});
```
#### Events

The events are triggered on tabs.

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Event</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.tabs.amui</code></td>
    <td>Fired immediately when a tab is opened.</td>
  </tr>
  <tr>
    <td><code>opened.tabs.amui</code></td>
    <td>Fired after a tab is fully opened (after CSS animate ends).</td>
  </tr>
  </tbody>
</table>

```javascript
$('#doc-my-tabs').find('a').on('opened.tabs.amui', function(e) {
  console.log('[%s] Tab is opened', $(this).text());
})
```

Open the console and operate on [the tabs above](#doc-my-tabs) to check the output of event listener.

## FAQ

### I can't select the content of Tab.

This problem is caused by [Hammer.js](http://hammerjs.github.io/tips/).

> Hammer is setting a property to improve the UX of the panning on desktop.

Use following style to replace the style in Hammer.js:

```css
.am-tabs-bd {
  -moz-user-select: text !important;
  -webkit-user-select: text !important;
  -ms-user-select: text !important;
  user-select: text !important;
}
```

Or disable the touch events.

<!--
## TODO

- Ajax 载入选项卡内容支持
-->
