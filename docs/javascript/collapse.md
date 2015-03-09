---
id: collapse
title: 折叠面板
titleEn: Collapse
prev: javascript/button.html
next: javascript/dropdown.html
source: js/ui.collapse.js
doc: docs/javascript/collapse.md
---

# Collapse
---

折叠效果组件，用于制作下滑菜单或手风琴效果。

## 使用演示

### 折叠面板

结合 [Panel](/css/panel) 组件实现手风琴效果。需结合以下辅助 class 使用：

* 要隐藏的内容添加 `.am-collapse`；
* 默认显示的内容添加 `.am-collapse.am-in`；

添加以上 class 以后，通过 Data API 来调用：

```html
<h4 data-am-collapse="{parent: '#accordion', target: '#do-not-say-1'}"></h4>
```

其中：

* `parent` 为容器 ID
* `target` 为要伸缩的容器 ID

如果触发元素为 `<a>` 元素，可以把 `target` 设置在 `href` 属性里。

```html
<a data-am-collapse="{parent: '#accordion'}" href="#do-not-say-1">
  ...
</a>
```

`````html
<div class="am-panel-group" id="accordion">
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-1'}">
        莫言 - 你不懂我，我不怪你 #1
      </h4>
    </div>
    <div id="do-not-say-1" class="am-panel-collapse am-collapse am-in">
      <div class="am-panel-bd">
        每个人都有一个死角， 自己走不出来，别人也闯不进去。 <br/>
        我把最深沉的秘密放在那里。 <br/>
        你不懂我，我不怪你。 <br/><br/>
        每个人都有一道伤口， 或深或浅，盖上布，以为不存在。 <br/>
        我把最殷红的鲜血涂在那里。 <br/>
        你不懂我，我不怪你。
      </div>
    </div>
  </div>
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-2'}">
        莫言 - 你不懂我，我不怪你 #2
      </h4>
    </div>
    <div id="do-not-say-2" class="am-panel-collapse am-collapse">
      <div class="am-panel-bd">
        每个人都有一场爱恋， 用心、用情、用力，感动也感伤。 <br/>
        我把最炙热的心情 藏在那里。 <br/>
        你不懂我，我不怪你。 <br/><br/>

        每个人都有 一行眼泪， 喝下的冰冷的水，酝酿成的热泪。 <br/>
        我把最心酸的委屈汇在那里。 <br/>
        你不懂我，我不怪你。 <br/><br/>

        每个人都有一段告白， 忐忑、不安，却饱含真心和勇气。 <br/>
        我把最抒情的语言用在那里。 <br/>
        你不懂我，我不怪你。
      </div>
    </div>
  </div>
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-3'}">
        莫言 - 你不懂我，我不怪你 #3
      </h4>
    </div>
    <div id="do-not-say-3" class="am-panel-collapse am-collapse">
      <div class="am-panel-bd">
        你永远也看不见我最爱你的时候， <br/>
        因为我只有在看不见你的时候，才最爱你。 <br/>
        同样，你永远也看不见我最寂寞的时候， <br/>
        因为我只有在你看不见我的时候，我才最寂寞。 <br/>
        <br/>
        也许，我太会隐藏自己的悲伤。 <br/>
        也许，我太会安慰自己的伤痕。 <br/>
        也许，你眼中的我，太会照顾自己， 所以，你从不考虑我的感受。 <br/>
        <br/>
        你以为，我可以很迅速的恢复过来，有些自私的以为。 <br/>

        从阴雨走到艳阳，我路过泥泞、路过风。 <br/>
        一路走来，你不曾懂我，我亦不曾怪你。 <br/>
      </div>
    </div>
  </div>
</div>
`````

```html
<div class="am-panel-group" id="accordion">
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-1'}">
        ...
      </h4>
    </div>
    <div id="do-not-say-1" class="am-panel-collapse am-collapse am-in">
      <div class="am-panel-bd">
        ...
      </div>
    </div>
  </div>
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-2'}">
        ...
      </h4>
    </div>
    <div id="do-not-say-2" class="am-panel-collapse am-collapse">
      <div class="am-panel-bd">
        ...
      </div>
    </div>
  </div>
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-3'}">...</h4>
    </div>
    <div id="do-not-say-3" class="am-panel-collapse am-collapse">
      <div class="am-panel-bd">
        ...
      </div>
    </div>
  </div>
</div>
```

### 折叠菜单

使用时注意目标元素外面应该有一个容器，以便动画执行时计算高度。

`````html
<button class="am-btn am-btn-primary" data-am-collapse="{target: '#collapse-nav'}">Menu <i class="am-icon-bars"></i></button>
<nav>
  <ul id="collapse-nav" class="am-nav am-collapse">
    <li><a href="">开始使用</a></li>
    <li><a href="">CSS 介绍</a></li>
    <li class="am-active"><a href="">JS 介绍</a></li>
    <li><a href="">功能定制</a></li>
  </ul>
</nav>
`````
```html
<button class="am-btn am-btn-primary" data-am-collapse="{target: '#collapse-nav'}">Menu <i class="am-icon-bars"></i></button>
<nav>
  <ul id="collapse-nav" class="am-nav am-collapse">
    <li><a href="">开始使用</a></li>
    <li><a href="">CSS 介绍</a></li>
    <li class="am-active"><a href="">JS 介绍</a></li>
    <li><a href="">功能定制</a></li>
  </ul>
</nav>
```

## 调用方式

### 通过 Data API

在元素上添加 `data-am-collapse` 并设置 `target` 的值为折叠元素 ID：

```html
<button data-am-collapse="{target: '#my-collapse'}"></button>
```

### 通过 JS

使用方法：

```js
$('#myCollapse').collapse()
```

#### 方法

- `$().collapse(options)` - 绑定元素展开/折叠操作

```javascript
$('#myCollapse').collapse({
  toggle: false
})
```

- `$().collapse('toggle')` - 切换面板状态
- `$().collapse('open')` - 展开面板
- `$().collapse('close')` - 关闭面板

#### 选项

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th style="width: 60px;">参数</th>
    <th style="width: 70px;">类型</th>
    <th style="width: 50px;">默认</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>parent</code></td>
    <td>选择符</td>
    <td><code>false</code></td>
    <td>如果设置了 <code>parent</code> 参数，且该容器下有多个可折叠的面板时，展开一个面板会关闭其它展开的面板。换言之，如果想让多个面板可以都处于展开状态，那不设置这个参数即可。</td>
  </tr>
  <tr>
    <td><code>toggle</code></td>
    <td>布尔值</td>
    <td><code>true</code></td>
    <td>交替执行展开/关闭操作</td>
  </tr>
  </tbody>
</table>


#### 自定义事件

自定义事件在**折叠的元素**上触发，以上面的折叠菜单为例，`#collapse-nav` 触发自定义事件：

<script>
$(function() {
  $('#collapse-nav').on('open.collapse.amui', function() {
    console.log('折叠菜单打开了！');
  }).on('close.collapse.amui', function() {
    console.log('折叠菜单关闭鸟！');
  });
});
</script>

```js
$('#collapse-nav').on('open.collapse.amui', function() {
  console.log('折叠菜单打开了！');
}).on('close.collapse.amui', function() {
  console.log('折叠菜单关闭鸟！');
});
```

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th>事件</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.collapse.amui</code></td>
    <td><code>open</code> 方法被调用时立即触发</td>
  </tr>
  <tr>
    <td><code>opened.collapse.amui</code></td>
    <td>元素完全展开后触发</td>
  </tr>
  <tr>
    <td><code>close.collapse.amui</code></td>
    <td><code>close</code> 方法被调用后立即触发
    </td>
  </tr>
  <tr>
    <td><code>closed.collapse.amui</code></td>
    <td>元素折叠完成后触发</td>
  </tr>
  </tbody>
</table>

## 注意事项

**不要在折叠内容的容器上设置垂直的 `margin`/`padding`/`border` 样式。**

jQuery 计算元素高度的方式有点奇葩，暂时只能通过上面的方式规避。
