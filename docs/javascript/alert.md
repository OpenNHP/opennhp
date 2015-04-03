---
title: 警示信息插件
titleEn: Alert
prev: javascript.html
next: javascript/button.html
source: js/ui.alert.js
doc: docs/javascript/alert.md
---

# Alert
---

显示可关闭的页内警告信息。

## 使用演示

### 基本形式

`````html
<div class="am-alert">
  没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。
</div>
`````
```html
<div class="am-alert">
  没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。
</div>
```

### 关闭按钮

`````html
<div class="am-alert" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。</p>
</div>
`````
```html
<div class="am-alert" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。</p>
</div>
```

### 不同状态

`````html
<div class="am-alert am-alert-success" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。</p>
</div>
<div class="am-alert am-alert-warning" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。</p>
</div>
<div class="am-alert am-alert-danger" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。</p>
</div>
<div class="am-alert am-alert-secondary" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。</p>
</div>
`````
```html
<div class="am-alert am-alert-success" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。</p>
</div>

<div class="am-alert am-alert-warning" data-am-alert>
  ...
</div>

<div class="am-alert am-alert-danger" data-am-alert>
  ...
</div>

<div class="am-alert am-alert-secondary" data-am-alert>
  ...
</div>
```

### 多内容

`````html
<div class="am-alert" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <h3>共同渡过</h3>
  <p>《共同渡过》是张国荣1987年发行的专辑《Summer Romance》中的一首歌。</p>
  <ul>
    <li>若我可再活多一次都盼</li>
    <li>再可以在路途重逢着你</li>
    <li>共去写一生的句子</li>
    <li>若我可再活多一次千次</li>
    <li>我都盼面前仍是你</li>
    <li>我要他生都有今生的暖意</li>
  </ul>
</div>
`````
```html
<div class="am-alert" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <h3>共同渡过</h3>
  <p>《共同渡过》是张国荣1987年发行的专辑《Summer Romance》中的一首歌。</p>
  <ul>
    <li>若我可再活多一次都盼</li>
    <li>再可以在路途重逢着你</li>
    <li>共去写一生的句子</li>
    <li>若我可再活多一次千次</li>
    <li>我都盼面前仍是你</li>
    <li>我要他生都有今生的暖意</li>
  </ul>
</div>
```

## 调用方式

### 通过 Data API

在 Alert 最外层元素上添加 `data-am-alert`。

`````html
<div class="am-alert" id="my-alert" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>没什么可给你，但求凭这阙歌。谢谢你风雨里，都不退愿陪着我。</p>
</div>
<script>
$(function() {
  $(document).on('closed.alert.amui', function() {
    console.log('警告窗口已经关闭');
  });
});
</script>
`````
```html
<div class="am-alert" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  ...
</div>
```

### 通过 JS

开启关闭按钮交互功能：

```javascript
$('.am-alert').alert()
```

#### 方法

- `$().alert()` - 激活 Alert 元素关闭按钮的交互功能。
- `$().alert('close')`：直接关闭元素。

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
      <td><code>close.alert.amui</code></td>
      <td><code>close</code> 方法被调用时立即触发</td>
    </tr>
    <tr>
      <td><code>closed.alert.amui</code></td>
      <td>元素被关闭以后触发（CSS 动画执行完成）</td>
    </tr>
  </tbody>
</table>

```js
$('#my-alert').on('closed.alert.amui', function() {
  alert('警告窗口已经关闭');
});
```

#### JS 控制示例

**示例1：**激活 Alert 元素关闭按钮的交互功能

`````html
<div class="am-alert" id="your-alert">
  <button type="button" class="am-close">&times;</button>
  <p>在应用 JS 激活之前这个警告框是无法关闭的。不信点右边的 x 试试。</p>
</div>

<button type="button" class="am-btn am-btn-danger" id="doc-alert-btn-bind">点击激活上面 Alert 的关闭功能</button>
<script>
  $(function() {
    $('#doc-alert-btn-bind').one('click', function() {
      $('#your-alert').alert();
      alert('已激活，点击上的 X 试试！');
    });
  });
</script>
`````
```html
<div class="am-alert" id="your-alert">
  <button type="button" class="am-close">&times;</button>
  <p>在应用 JS 激活之前这个警告框是无法关闭的。不信点右边的 x 试试。</p>
</div>
<button type="button" class="am-btn am-btn-danger" id="doc-alert-btn-bind">点击激活上面 Alert 的关闭功能</button>
<script>
  $(function() {
    $('#doc-alert-btn-bind').one('click', function() {
      $('#your-alert').alert();
      alert('已激活，点击上的 X 试试！');
    });
  });
</script>
```

**示例2**：使用 JS 关闭警告框及监听自定义事件

`````html
<div class="am-alert" id="your-alert-1">
  <button type="button" class="am-close">&times;</button>
  <p>这是一个警告框！</p>
</div>

<button type="button" class="am-btn am-btn-warning" id="doc-alert-btn-close">点击关闭上面的 Alert</button>
<script>
  $(function() {
    $('#doc-alert-btn-close').one('click', function() {
      $('#your-alert-1').alert('close');
    });

    $(document).on('close.alert.amui', '#your-alert-1', function(e) {
      alert('警告框开始关闭！');
    });

    $(document).on('closed.alert.amui', '#your-alert-1', function(e) {
      alert('警告框关闭完成！');
    });
  });
</script>
`````

```html
<div class="am-alert" id="your-alert-1">
  <button type="button" class="am-close">&times;</button>
  <p>这是一个警告框！</p>
</div>

<button type="button" class="am-btn am-btn-warning" id="doc-alert-btn-close">点击关闭上面的 Alert</button>
<script>
  $(function() {
    $('#doc-alert-btn-close').one('click', function() {
      $('#your-alert-1').alert('close');
    });

    $(document).on('close.alert.amui', '#your-alert-1', function(e) {
      alert('警告框开始关闭！');
    });

    $(document).on('closed.alert.amui', '#your-alert-1', function(e) {
      alert('警告框关闭完成！');
    });
  });
</script>
```
