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
  $(document).on('closed:alert:amui', function() {
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
      <td><code>close:alert:amui</code></td>
      <td><code>close</code> 方法被调用时立即触发</td>
    </tr>
    <tr>
      <td><code>closed:alert:amui</code></td>
      <td>元素被关闭以后触发（CSS 动画执行完成）</td>
    </tr>
  </tbody>
</table>

```js
$('#my-alert').on('closed:alert:amui', function() {
  alert('警告窗口已经关闭');
});
```
