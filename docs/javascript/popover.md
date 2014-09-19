# Popover
---

气泡式弹出层组件，本组件无需创建 HTML 结构。

插件根据元素在窗口中的位置自动判断弹出层显示位置，目前没有设置位置的选项。

## 使用演示

### 点击显示

`````html
  <button class="am-btn am-btn-primary" data-am-popover="{content: '鄙是点击显示的 Popover'}">点击显示 Popover</button>
`````
```html
<button class="am-btn am-btn-primary" data-am-popover="{content: '鄙是点击显示的 Popover'}">点击显示 Popover</button>
```

### Hover/Focus 显示

Tooltip 效果。

`````html
<button class="am-btn am-btn-success" data-am-popover="{content: '鄙是 Hover/Focus 显示的 Popover', trigger: 'hover focus'}">Hover/Focus 显示 Popover</button>
`````
```html
<button class="am-btn am-btn-success"
        data-am-popover="{content: '鄙是点击 Hover 显示的 Popover', trigger: 'hover focus'}">
  Hover 显示 Popover
</button>
```

## 使用方式

### 通过 Data API

在元素上添加 `data-am-popover` 属性并设置相关参数。上面的演示都是通过 Data API 实现的。

```html
<button data-am-popover="{content: '想显示啥', trigger: 'hover'}">Popover
</button>
```

### 通过 JS

通过 `$().popover(options)` 方式添加元素 Popover 交互。

`````html
<button class="am-btn am-btn-danger" id="my-popover">Popover via JS</button>
<script>
$(function() {
  $('#my-popover').popover({
    content: 'Popover via JavaScript'
  })
})
</script>
`````
```html
<button class="am-btn am-btn-danger" id="my-popover">Popover via JS</button>
```
```javascript
$(function() {
  $('#my-popover').popover({
    content: 'Popover via JavaScript'
  })
});
```

#### 参数说明

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>参数</th>
    <th>类型</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>content</code></td>
    <td><code>string</code></td>
    <td>Popover 显示的内容</td>
  </tr>
  <tr>
    <td><code>trigger</code></td>
    <td><code>string</code></td>
    <td>交互方式，<code>click|hover|focus</code>，默认为 <code>click</code></td>
  </tr>
  </tbody>
</table>

#### 方法

- `.popover(options)` - 激活元素的 Popover 交互功能，`options` 为对象
- `.popover('toggle')` - 交替 Popover 状态
- `.popover('open')` - 显示 Popover
- `.popover('close')` - 关闭 Popover

#### 自定义事件

事件定义在触发 Popover 交互的元素上。

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open:popover:amui</code></td>
    <td><code>open</code> 方法被调用是立即触发</td>
  </tr>
  <tr>
    <td><code>close:popover:amui</code></td>
    <td><code>close</code> 方法被调用是立即触发</td>
  </tr>
  </tbody>
</table>
