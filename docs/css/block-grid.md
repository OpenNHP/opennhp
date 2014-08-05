[//]: DONE

# Block-Grid
---

Block Grid 用来创建等分的内容网格，用于内容的排列。

响应式断点为：

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th style="width: 160px">Class</th>
    <th>区间</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>sm-block-grid-*</code></td>
    <td><code>0 - 640px</code></td>
  </tr>
  <tr>
    <td><code>md-block-grid-*</code></td>
    <td><code>641px - 1024px</code></td>
  </tr>
  <tr>
    <td><code>lg-block-grid-*</code></td>
    <td><code>1025px + </code></td>
  </tr>
  </tbody>
</table>

与布局网格不同的是， __这里的数字表示几等分__，而不是占12等分中几列，比如 `.sm-block-grid-2` 会将子元素 `<li>` 的宽度设置为 `50%`。

考虑到通用性（菜单、图片）等，`<li>` 没有设置 `padding`，使用时需根据需求自行设置。

另外需要注意的 Block Grid __只能用于 `<ul>` / `<ol>` 结构__。

下面的演示中，我们添加了以下自定义样式：

```css
// block-grid
.doc-block-grid {
  margin-left: -5px;
  margin-right: -5px;
  > li {
    padding: 0 5px 10px 5px;
    img {
      border: 1px solid #CCC;
      padding: 2px;
      background: #FFF;
    }
  }
}
```


## 基本使用

只添加 `sm-block-grid-*`，应用于所有屏幕尺寸。

`````html
<ul class="sm-block-grid-4 doc-block-grid">
  <li><img src="http://cn.bing.com/az/hprichv/LondonTrainStation_GettyRR_139321755_ZH-CN742316019.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/CardinalsBerries_ZH-CN10679090179_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/QingdaoJiaozhou_ZH-CN10690497202_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/FennecFox_ZH-CN13720911949_1366x768.jpg" /></li>
</ul>
`````

```html
<ul class="sm-block-grid-4">
  <li><img src="http://cn.bing.com/az/hprichv/LondonTrainStation_GettyRR_139321755_ZH-CN742316019.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/CardinalsBerries_ZH-CN10679090179_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/QingdaoJiaozhou_ZH-CN10690497202_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/FennecFox_ZH-CN13720911949_1366x768.jpg" /></li>
</ul>
```


## 响应式

按需增加更多响应式 class，缩放窗口可以查看响应效果。

`````html
<ul class="sm-block-grid-2 md-block-grid-3 lg-block-grid-4 doc-block-grid">
  <li><img src="http://cn.bing.com/az/hprichv/LondonTrainStation_GettyRR_139321755_ZH-CN742316019.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/CardinalsBerries_ZH-CN10679090179_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/QingdaoJiaozhou_ZH-CN10690497202_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/FennecFox_ZH-CN13720911949_1366x768.jpg" /></li>
  <li><img src="http://cn.bing.com/az/hprichv/LondonTrainStation_GettyRR_139321755_ZH-CN742316019.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/CardinalsBerries_ZH-CN10679090179_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/QingdaoJiaozhou_ZH-CN10690497202_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/FennecFox_ZH-CN13720911949_1366x768.jpg" /></li>
</ul>
`````

```html
<ul class="sm-block-grid-2 md-block-grid-3 lg-block-grid-4">
  <li><img src="http://cn.bing.com/az/hprichv/LondonTrainStation_GettyRR_139321755_ZH-CN742316019.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/CardinalsBerries_ZH-CN10679090179_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/QingdaoJiaozhou_ZH-CN10690497202_1366x768.jpg" /></li>
  <li><img src="http://s.cn.bing.net/az/hprichbg/rb/FennecFox_ZH-CN13720911949_1366x768.jpg" /></li>
  <li><img src="..." /></li>
  <li><img src="..." /></li>
  <li><img src="..." /></li>
  <li><img src="..." /></li>
</ul>
```
