# Alert
---

显示可关闭的页内警告信息。

## 使用演示

### 基本形式

`````html
<div class="am-alert">
  云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。
</div>
`````
```html
<div class="am-alert">
  云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。
</div>
```

### 关闭按钮

`````html
<div class="am-alert" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
</div>
`````
```html
<div class="am-alert">
  <button type="button" class="am-close">&times;</button>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
</div>
```

### 不同状态

`````html
<div class="am-alert am-alert-success" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
</div>
<div class="am-alert am-alert-warning" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
</div>
<div class="am-alert am-alert-danger" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
</div>
<div class="am-alert am-alert-secondary" data-am-alert>
  <button type="button" class="am-close" data-alert->&times;</button>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
</div>
`````
```html
<div class="am-alert am-alert-success">
  <button type="button" class="am-close">&times;</button>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
</div>

<div class="am-alert am-alert-warning">
  ...
</div>

<div class="am-alert am-alert-danger">
  ...
</div>

<div class="am-alert am-alert-secondary">
  ...
</div>
```

### 多内容

`````html
<div class="am-alert" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <h3>云适配相对其他手机适配服务有哪些优点</h3>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
  <ul>
  	<li>1. 一行代码，简单快捷;</li>
  	<li>2. 网址不变且唯一</li>
  	<li>3. 内容实时同步更新</li>
  	<li>4. 云端跨平台适配</li>
  	<li>5. 专属的一键拨叫 在线咨询 地图导航 二维码 一键分享实用功能。</li>
  </ul>
</div>
<script>
  seajs.use(['ui.alert']);
</script>
`````
```html
<div class="am-alert">
  <button type="button" class="am-close">&times;</button>
  <h3>云适配相对其他手机适配服务有哪些优点</h3>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
  <ul>
  	<li>1. 一行代码，简单快捷;</li>
  	<li>2. 网址不变且唯一</li>
  	<li>3. 内容实时同步更新</li>
  	<li>4. 云端跨平台适配</li>
  	<li>5. 专属的一键拨叫 在线咨询 地图导航 二维码 一键分享实用功能。</li>
  </ul>
</div>
```

## 调用方式

### 通过 Data API

在 Alert 最外层元素上添加 `data-am-alert`。

`````html
<div class="am-alert" id="my-alert" data-am-alert>
  <button type="button" class="am-close">&times;</button>
  <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
</div>
<script>
(function($){
  $(document).on('closed:alert:amui', function() {
    console.log('警告窗口已经关闭');
  });
})(window.Zepto);
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
