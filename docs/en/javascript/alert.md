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

Show a closable alert in page .

## Examples

### Default Style

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

### Close Button

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

### Modifiers

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

### More Contents

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

## Usage

### Use through Data API

Add the `data-am-alert` class to the container of Alert.

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

### Through JS

Enable/disable the close button:

```javascript
$('.am-alert').alert()
```

#### Methods

- `$().alert()`: Enable the close button in alert element.
- `$().alert('close')`: Close the alert.

#### Events

<table class="am-table am-table-bd am-table-striped">
  <thead>
    <tr>
      <th>Event Name</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>close.alert.amui</code></td>
      <td>Immediately triggered when <code>close</code> is clicked</td>
    </tr>
    <tr>
      <td><code>closed.alert.amui</code></td>
      <td>Triggered when the element is closed (after the CSS animation ends).</td>
    </tr>
  </tbody>
</table>

```js
$('#my-alert').on('closed.alert.amui', function() {
  alert('Alert is closed');
});
```

#### JS Examples

**Example 1:** Activate the close button of alert element

`````html
<div class="am-alert" id="your-alert">
  <button type="button" class="am-close">&times;</button>
  <p>You can't close this alert before close button is activated.</p>
</div>

<button type="button" class="am-btn am-btn-danger" id="doc-alert-btn-bind">Activate the close button</button>
<script>
  $(function() {
    $('#doc-alert-btn-bind').one('click', function() {
      $('#your-alert').alert();
      alert('Close button activated.');
    });
  });
</script>
`````
```html
<div class="am-alert" id="your-alert">
  <button type="button" class="am-close">&times;</button>
  <p>You can't close this alert before close button is activated.</p>
</div>

<button type="button" class="am-btn am-btn-danger" id="doc-alert-btn-bind">Activate the close button</button>
<script>
  $(function() {
    $('#doc-alert-btn-bind').one('click', function() {
      $('#your-alert').alert();
      alert('Close button activated.');
    });
  });
</script>
```

**Example 2**: Use JS to close the alert and listen to the customized event

`````html
<div class="am-alert" id="your-alert-1">
  <button type="button" class="am-close">&times;</button>
  <p>This is an alert!</p>
</div>

<button type="button" class="am-btn am-btn-warning" id="doc-alert-btn-close">Click to close the alert</button>
<script>
  $(function() {
    $('#doc-alert-btn-close').one('click', function() {
      $('#your-alert-1').alert('close');
    });

    $(document).on('close.alert.amui', '#your-alert-1', function(e) {
      alert('Alert starts to close！');
    });

    $(document).on('closed.alert.amui', '#your-alert-1', function(e) {
      alert('Alert has been closed！');
    });
  });
</script>
`````

```html
<div class="am-alert" id="your-alert-1">
  <button type="button" class="am-close">&times;</button>
  <p>This is an alert!</p>
</div>

<button type="button" class="am-btn am-btn-warning" id="doc-alert-btn-close">Click to close the alert</button>
<script>
  $(function() {
    $('#doc-alert-btn-close').one('click', function() {
      $('#your-alert-1').alert('close');
    });

    $(document).on('close.alert.amui', '#your-alert-1', function(e) {
      alert('Alert starts to close！');
    });

    $(document).on('closed.alert.amui', '#your-alert-1', function(e) {
      alert('Alert has been closed！');
    });
  });
</script>
```
