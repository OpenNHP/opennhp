# Dropdown
---

## 基本使用

### 下拉菜单

`````html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-primary am-dropdown-toggle" data-am-dropdown-toggle>Dropdown <span class="am-icon-caret-down"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">Header</li>
    <li><a href="#">1. 一行代码，简单快捷</a></li>
    <li class="am-active"><a href="#">2. 网址不变且唯一</a></li>
    <li><a href="#">3. 内容实时同步更新</a></li>
    <li class="am-disabled"><a href="#">4. 云端跨平台适配</a></li>
    <li class="am-divider"></li>
    <li><a href="#">5. 专属的一键拨叫</a></li>
  </ul>
</div>
`````
```html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-default am-dropdown-toggle" data-am-dropdown-toggle>Dropdown <span class="am-icon-caret-down"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">Header</li>
    <li><a href="#">1. 一行代码，简单快捷</a></li>
    <li class="am-active"><a href="#">2. 网址不变且唯一</a></li>
    <li><a href="#">3. 内容实时同步更新</a></li>
    <li class="am-disabled"><a href="#">4. 云端跨平台适配</a></li>
    <li class="am-divider"></li>
    <li><a href="#">5. 专属的一键拨叫</a></li>
  </ul>
</div>
```

### 下拉内容

`````html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-success am-dropdown-toggle">Dropdown <span class="am-icon-caret-down"></span></button>
  <div class="am-dropdown-content">
    <h2>关于我们</h2>
    <p>
      AllMobilize Inc (美通云动科技有限公司) 由前微软美国总部IE浏览器核心研发团队成员及移动互联网行业专家在美国西雅图创立，旨在解决网页在不同移动设备屏幕上的适配问题。
    </p>

  </div>
</div>
<script>
$(function() {
  $('[data-am-dropdown]').on('open:dropdown:amui', function () {
    console.log('open event triggered');
  });
});
</script>
`````
```html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-success am-dropdown-toggle">Dropdown <span class="am-icon-caret-down"></span></button>
  <div class="am-dropdown-content">
    <h2>关于我们</h2>
    <p>
      AllMobilize Inc (美通云动科技有限公司) 由前微软美国总部IE浏览器核心研发团队成员及移动互联网行业专家在美国西雅图创立，旨在解决网页在不同移动设备屏幕上的适配问题。
    </p>
  </div>
</div>
```

### 选项

TBD.

## JS 交互

### 选项

TBD.

### 方法

`$().dropdown('toggle')`

### 自定义事件

下拉框的事件在 `[data-am-dropdown]` 元素上触发。

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open:dropdown:amui</code></td>
    <td>调用显示下拉框方法时立即触发</td>
  </tr>
  <tr>
    <td><code>opened:dropdown:amui</code></td>
    <td>下拉框显示完成时触发</td>
  </tr>
  <tr>
    <td><code>close:dropdown:amui</code></td>
    <td>调用隐藏方法时触发</td>
  </tr>
  <tr>
    <td><code>closed:dropdown:amui</code></td>
    <td>下拉框关闭完成时触发</td>
  </tr>
  </tbody>
</table>

```js
$(function() {
  $('[data-am-dropdown]').on('open:dropdown:amui', function () {
    console.log('open event triggered');
  });
});
```