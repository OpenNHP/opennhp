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

下面的例子同时演示了动态插入、删除选项卡的实现方式。

`````html
<style>
  .am-tabs-nav li {
    position: relative;
    z-index: 1;
  }

  .am-tabs-nav .am-icon-close {
    position: absolute;
    top: 0;
    right: 10px;
    color: #888;
    cursor: pointer;
    z-index: 100;
  }

  .am-tabs-nav .am-icon-close:hover {
    color: #333;
  }

  .am-tabs-nav .am-icon-close ~ a {
    padding-right: 25px!important;
  }
</style>
<div class="am-tabs" data-am-tabs="{noSwipe: 1}" id="doc-tab-demo-1">
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="javascript: void(0)">流浪</a></li>
    <li><a href="javascript: void(0)">流浪</a></li>
    <li><a href="javascript: void(0)">再流浪</a></li>
  </ul>

  <div class="am-tabs-bd">
    <div class="am-tab-panel am-active">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？ A
    </div>
    <div class="am-tab-panel">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？ B
    </div>
    <div class="am-tab-panel">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？ C
    </div>
  </div>
</div>
<br />
<button type="button" class="am-btn am-btn-primary js-append-tab">插入 Tab</button>
<script>
  $(function() {
    var tabCounter = 0;
    var $tab = $('#doc-tab-demo-1');
    var $nav = $tab.find('.am-tabs-nav');
    var $bd = $tab.find('.am-tabs-bd');

    function addTab() {
      var nav = '<li><span class="am-icon-close"></span>' +
        '<a href="javascript: void(0)">标签 ' + tabCounter + '</a></li>';
      var content = '<div class="am-tab-panel">动态插入的标签内容' + tabCounter + '</div>';

      $nav.append(nav);
      $bd.append(content);
      tabCounter++;
      $tab.tabs('refresh');
    }

    // 动态添加标签页
    $('.js-append-tab').on('click', function() {
      addTab();
    });

    // 移除标签页
    $nav.on('click', '.am-icon-close', function() {
      var $item = $(this).closest('li');
      var index = $nav.children('li').index($item);

      $item.remove();
      $bd.find('.am-tab-panel').eq(index).remove();

      $tab.tabs('open', index > 0 ? index - 1 : index + 1);
      $tab.tabs('refresh');
    });
  });
</script>
`````

```html
<div class="am-tabs" data-am-tabs="{noSwipe: 1}" id="doc-tab-demo-1">
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
<br />
<button type="button" class="am-btn am-btn-primary js-append-tab">插入 Tab</button>
<script>
  $(function() {
    var tabCounter = 0;
    var $tab = $('#doc-tab-demo-1');
    var $nav = $tab.find('.am-tabs-nav');
    var $bd = $tab.find('.am-tabs-bd');

    function addTab() {
      var nav = '<li><span class="am-icon-close"></span>' +
        '<a href="javascript: void(0)">标签 ' + tabCounter + '</a></li>';
      var content = '<div class="am-tab-panel">动态插入的标签内容' + tabCounter + '</div>';

      $nav.append(nav);
      $bd.append(content);
      tabCounter++;
      $tab.tabs('refresh');
    }

    // 动态添加标签页
    $('.js-append-tab').on('click', function() {
      addTab();
    });

    // 移除标签页
    $nav.on('click', '.am-icon-close', function() {
      var $item = $(this).closest('li');
      var index = $nav.children('li').index($item);

      $item.remove();
      $bd.find('.am-tab-panel').eq(index).remove();

      $tab.tabs('open', index > 0 ? index - 1 : index + 1);
      $tab.tabs('refresh');
    });
  });
</script>
```
```css
  .am-tabs-nav li {
    position: relative;
    z-index: 1;
  }

  .am-tabs-nav .am-icon-close {
    position: absolute;
    top: 0;
    right: 10px;
    color: #888;
    cursor: pointer;
    z-index: 100;
  }

  .am-tabs-nav .am-icon-close:hover {
    color: #333;
  }

  .am-tabs-nav .am-icon-close ~ a {
    padding-right: 25px!important;
  }
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

### Tab 内容溢出容器问题

为了实现动画效果，标签内容容器 `.am-tabs-bd` 上添加了 `overflow: hidden`，在某些场景会有一些问题（[#833](https://github.com/amazeui/amazeui/issues/833)、[#901](https://github.com/amazeui/amazeui/issues/901)），目前的处理方式是在 `.am-tabs-bd` 上添加 `.am-tabs-bd-ofv`（动画效果会被取消）。

`````html
<div class="am-tabs" data-am-tabs>
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="#tab-4-1">恣意</a></li>
    <li><a href="#tab-4-2">等候</a></li>
    <li><a href="#tab-4-3">流浪</a></li>
  </ul>
  <div class="am-tabs-bd am-tabs-bd-ofv">
    <div class="am-tab-panel am-active" id="tab-4-1">
       置身人群中<br>你只需要被淹没 享受 沉默<br>退到人群后<br>你只需给予双手 微笑 等候
       <br />
        <div class="am-dropdown" data-am-dropdown>
          <button class="am-btn am-btn-primary am-dropdown-toggle" data-am-dropdown-toggle>下拉列表 <span class="am-icon-caret-down"></span></button>
          <ul class="am-dropdown-content">
            <li class="am-dropdown-header">标题</li>
            <li><a href="#">快乐的方式不只一种</a></li>
            <li class="am-active"><a href="#">最荣幸是</a></li>
            <li><a href="#">谁都是造物者的光荣</a></li>
            <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
            <li class="am-divider"></li>
            <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
          </ul>
        </div>
    </div>
    <div class="am-tab-panel" id="tab-4-2">
      走在忠孝东路<br>徘徊在茫然中<br>在我的人生旅途<br>选择了多少错误<br>我在睡梦中惊醒<br>感叹悔言无尽<br>恨我不能说服自己<br>接受一切教训<br>让生命去等候<br>等候下一个漂流<br>让生命去等候<br>等候下一个伤口
    </div>
    <div class="am-tab-panel" id="tab-4-3">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？
    </div>
  </div>
</div>
`````
```html
<div class="am-tabs" data-am-tabs>
  <ul class="am-tabs-nav am-nav am-nav-tabs">
    <li class="am-active"><a href="#tab-4-1">恣意</a></li>
    <li><a href="#tab-4-2">等候</a></li>
    <li><a href="#tab-4-3">流浪</a></li>
  </ul>
  <div class="am-tabs-bd am-tabs-bd-ofv">
    <div class="am-tab-panel am-active" id="tab-4-1">
       置身人群中<br>你只需要被淹没 享受 沉默<br>退到人群后<br>你只需给予双手 微笑 等候
       <br />
         <div class="am-dropdown" data-am-dropdown>
           <button class="am-btn am-btn-primary am-dropdown-toggle" data-am-dropdown-toggle>下拉列表 <span class="am-icon-caret-down"></span></button>
           <ul class="am-dropdown-content">
             <li class="am-dropdown-header">标题</li>
             <li><a href="#">快乐的方式不只一种</a></li>
             <li class="am-active"><a href="#">最荣幸是</a></li>
             <li><a href="#">谁都是造物者的光荣</a></li>
             <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
             <li class="am-divider"></li>
             <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
           </ul>
         </div>
    </div>
    <div class="am-tab-panel" id="tab-4-2">
      走在忠孝东路<br>徘徊在茫然中<br>在我的人生旅途<br>选择了多少错误<br>我在睡梦中惊醒<br>感叹悔言无尽<br>恨我不能说服自己<br>接受一切教训<br>让生命去等候<br>等候下一个漂流<br>让生命去等候<br>等候下一个伤口
    </div>
    <div class="am-tab-panel" id="tab-4-3">
      我就这样告别山下的家，我实在不愿轻易让眼泪留下。我以为我并不差不会害怕，我就这样自己照顾自己长大。我不想因为现实把头低下，我以为我并不差能学会虚假。怎样才能够看穿面具里的谎话？别让我的真心散的像沙。如果有一天我变得更复杂，还能不能唱出歌声里的那幅画？
    </div>
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

#### 选项

- `options.noSwipe` 是否禁用触控事件。

```js
$('#someTabs').tabs({noSwipe: 1});
```

#### 方法

- `$().tabs(options)` - 初始化选项卡；
- `$().tabs('open', index)` - 切换到指定的标签页，`index` 可以是数值或 jQuery 对象（选择符），如 `$('.am-tabs-nav a').eq(2)`；
- `$().tabs('refresh')` - 刷新选项卡，动态添加、移除标签页后需手动刷新；
- `$().tabs('destroy')` - 销毁选项卡。

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
