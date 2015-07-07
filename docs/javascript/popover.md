---
id: popover
title: 弹出框
titleEn: Popover
prev: javascript/modal.html
next: javascript/nprogress.html
source: js/ui.popover.js
doc: docs/javascript/popover.md
---

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

### 颜色/尺寸

通过 `theme` 选项可以设置 Popover 的颜色和尺寸。

`````html
<!--蓝色的 Popover-->
<button
  class="am-btn am-btn-primary"
  data-am-popover="{theme: 'primary', content: '点击显示的 Primary'}">
  Primary
</button>

<!--红色、sm Popover-->
<button
  class="am-btn am-btn-secondary"
  data-am-popover="{theme: 'danger sm', content: '点击显示的 Danger & Small'}">
  Danger
</button>

<!--警示、lg Popover-->
<button
  class="am-btn am-btn-warning"
  data-am-popover="{theme: 'warning lg', content: '点击显示的 Danger & Small'}">
  Warning
</button>
`````
```html
<!--蓝色的 Popover-->
<button
  class="am-btn am-btn-primary"
  data-am-popover="{theme: 'primary', content: '点击显示的 Primary'}">
  Primary
</button>

<!--红色、sm Popover-->
<button
  class="am-btn am-btn-secondary"
  data-am-popover="{theme: 'danger sm', content: '点击显示的 Danger & Small'}">
  Danger
</button>

<!--警示、lg Popover-->
<button
  class="am-btn am-btn-warning"
  data-am-popover="{theme: 'warning lg', content: '点击显示的 Danger & Small'}">
  Warning
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

| 参数 | 类型 | 描述 |
| --- |  --- | --- |
| `content` | `string` | Popover 显示的内容
| `trigger` | `string` | 交互方式，`click` / `hover` / `focus`，默认为 `click`|
|`theme`|`string`| Popover 样式，颜色：`primary` / `secondary` / `success` / `warning` / `danger`；尺寸： `sm` / `lg`。同时设置颜色和尺寸使用一个半角空格 ` ` 分隔。|



#### 方法

- `.popover(options)` - 激活元素的 Popover 交互功能，`options` 为对象
- `.popover('toggle')` - 交替 Popover 状态
- `.popover('open')` - 显示 Popover
- `.popover('close')` - 关闭 Popover
- `.popover('setContent', content)` - 设置弹出层内容 <span class="am-badge am-badge-danger">v2.4.1+</span>

#### 自定义事件

事件定义在触发 Popover 交互的元素上。

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.popover.amui</code></td>
    <td><code>open</code> 方法被调用是立即触发</td>
  </tr>
  <tr>
    <td><code>close.popover.amui</code></td>
    <td><code>close</code> 方法被调用是立即触发</td>
  </tr>
  </tbody>
</table>
