# Compatibility
---

Amaze UI 面向现代浏览器开发，对 IE 8/9 等浏览器只提供有限的支持。

**注意：**

- **请不要使用 `IETester` 这种不靠谱的工具测试**；
- 按照微软[官方的说法](https://www.modern.ie/en-us/f12)，IE 开发者工具中的浏览器模式也不一定靠谱；
- 微软官方提供了[各种 IE 测试虚拟机](https://www.modern.ie/zh-cn/virtualization-tools#downloads)。

## 分级浏览器支持（<abbr title="Graded Browser Support">GBS</abbr>）

[<abbr title="Graded Browser Support">GBS</abbr>](https://github.com/yui/yui3/wiki/Graded-Browser-Support) 是 YUI 团队提出的应对日益增长的浏览器兼容问题的思路，详情可以查看 [YUI 相关页面](https://github.com/yui/yui3/wiki/Graded-Browser-Support)。

### Amaze UI GBS 描述

- __A 级：__最高支持级别，充分利用 H5 和 CSS3 等技术，提供最优的视觉和交互效果。
- __B 级：__有限支持，基本的样式和正常的交互，不考虑视觉、交互效果。
- __C 级：__核心支持，显示语义化的 HTML 标记渲染的内容，不考虑样式和行为。
- __X 级：__未知、零散的很少使用或已经停止开发的浏览器，可能不支持，也可能支持。

### Amaze UI GBS

按照国际惯例，Amaze UI 提供对主流浏览器（系统）最近两个稳定版本的全面支持。结合国内实际情况，一些浏览器的支持缩减为最新正式版，IE 则对更老版本做了有限支持。

Amaze UI 对浏览器做了一个粗略分级，**优先支持 A 级浏览器**。

由于资源有限，无法列出所有的浏览器，**使用 `WebKit` 的浏览器只要不乱修改内核，理论上应该都支持**。

关于浏览器功能支持的更多细节请参考 [Can I use](http://caniuse.com/)（UC 浏览器的数据已经被收录，不知是喜是忧）。

<table class="am-table am-table-bordered am-table-striped">
  <thead>
    <tr>
      <th scope="row">OS/Browser</th>
      <th scope="row">Ver</th>
      <th scope="row">Windows</th>
      <th scope="row">iOS(7.1.2+)</th>
      <th scope="row">OS X (10.9+)</th>
      <th scope="row">Android (4.1+)</th>
      <th scope="row">WP(8+)</th>
    </tr>
  </thead>
  <tbody>
  <tr>
    <th scope="row">Chrome</th>
    <td>L2</td>
    <td class="am-success">A</td>
    <td class="am-success">A</td>
    <td class="am-success">A</td>
    <td class="am-success">A</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <th scope="row" rowspan="4">IE</th>
    <td>10+</td>
    <td class="am-success">A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-success">A-</td>
  </tr>
  <tr>
    <td>9</td>
    <td class="am-warning">B</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <td>8</td>
    <td class="am-danger">C+</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <td>lte7</td>
    <td class="am-danger">C</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <th scope="row">Firefox</th>
    <td>L2</td>
    <td class="am-success">A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-success">A</td>
    <td class="am-primary">X</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <th scope="row">Safari</th>
    <td>L2</td>
    <td class="am-primary">X</td>
    <td class="am-success">A</td>
    <td class="am-success">A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <th scope="row">Opera</th>
    <td>L1</td>
    <td class="am-primary">X</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-primary">X</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <th scope="row">Opera Coast</th>
    <td>L1</td>
    <td class="am-disabled">N/A</td>
    <td class="am-success">A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <th scope="row">Opera Mini</th>
    <td>L1</td>
    <td class="am-disabled">N/A</td>
    <td class="am-primary">X</td>
    <td class="am-disabled">N/A</td>
    <td class="am-primary">X</td>
    <td class="am-primary">X</td>
  </tr>
  <tr>
    <th scope="row" class="am-text-nowrap">Stock<sup>1</sup></th>
    <td>L1</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-primary">X</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <th scope="row">UC 浏览器</th>
    <td>L1</td>
    <td class="am-primary">X</td>
    <td class="am-success">A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-success">A</td>
    <td class="am-success">A-</td>
  </tr>
  <tr>
    <th scope="row" rowspan="2">360浏览器</th>
    <td>L1-极速</td>
    <td class="am-success">A-</td>
    <td class="am-primary">X</td>
    <td class="am-success">N/A</td>
    <td class="am-primary">X</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <td>L1-IE8</td>
    <td class="am-danger">C+</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-primary">X</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <th scope="row" rowspan="2">搜狗浏览器</th>
    <td class="am-text-nowrap">L1-极速</td>
    <td class="am-success">A-</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <td>L1-IE8</td>
    <td class="am-danger">C+</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
  </tr>
  <tr>
    <th scope="row">FF Mobile</th>
    <td>L1</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td>X</td>
    <td class="am-disabled">N/A</td>
  </tr>
  </tbody>
</table>

__注释：__

- `L` 代表 `last`，`L2` - 最新的两个稳定版本；`L1` - 最新稳定版本。
- `1` 安卓系统自动浏览器，由于部分厂商对浏览器做了修改，列为 X 级。

__参考链接__：

- [iOS Version Stats](http://david-smith.org/iosversionstats/)

### IE 8/9

- IE 8/9 不支持 `transition`，看不到任何动画效果；
- IE 9 对 ES5 支持相对较好，引入完整的 `amazeui.js` 不会报错，IE 8 则不然；
- **Web 组件部分不提供 IE 8/9 官方支持**。

**有限支持 IE 8/9 的 JS 插件**：

- 警告框（Alert）；
- 按钮交互（Button）；
- 折叠面板（Collpase）；
- 下拉组件（Dropdown）；
- 模态窗口（Modal）；
- 弹出框（Popover）；
- 图片轮播（Slider）；
- 侧边栏（OffCanvas）；
- 滚动侦测（ScrollSpyNav）；
- 固定元素（Sticky）；
- 选项卡（Tabs）；

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th scope="col" class="col-xs-4">功能</th>
    <th scope="col" class="col-xs-4">IE 8</th>
    <th scope="col" class="col-xs-4">IE 9</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <th scope="row"><code>border-radius</code></th>
    <td class="am-danger"><span class="am-icon-close"></span>NO</td>
    <td class="am-success"><span class="am-icon-check"></span>YES</td>
  </tr>
  <tr>
    <th scope="row"><code>box-shadow</code></th>
    <td class="am-danger"><span class="am-icon-remove"></span>NO</td>
    <td class="am-success"><span class="am-icon-check"></span>YES</td>
  </tr>
  <tr>
    <th scope="row"><code>transform</code></th>
    <td class="am-danger"><span class="am-icon-remove"></span>NO</td>
    <td class="am-success"><span class="am-icon-check"></span>YES（<code>-ms</code> 前缀）</td>
  </tr>
  <tr>
    <th scope="row"><code>Flex Box</code></th>
    <td colspan="2" class="am-danger"><span class="am-icon-remove"></span>NO</td>
  </tr>
  <tr>
    <th scope="row"><code>transition</code></th>
    <td colspan="2" class="am-danger"><span class="am-icon-remove"></span>NO</td>
  </tr>
  <tr>
    <th scope="row"><code>placeholder</code></th>
    <td colspan="2" class="am-danger"><span class="am-icon-remove"></span>NO</td>
  </tr>
  </tbody>
</table>

### IE 8

<div class="am-alert am-alert-warning">
  请注意，我们对 IE 8 的支持的仅限为<strong>布局、部分 JS 插件交互基本正常</strong>，IE 8 相关 Issue 将列为 `P4`(最低优先级，可能不会处理)。
</div>

需要**支持 IE 8 的用户请使用条件注释引入~~`amazeui.legacy.js`~~** `amazeui.ie8polyfill.min.js`。

```html
﻿<!--[if (gte IE 9)|!(IE)]><!-->
<script src="http://libs.baidu.com/jquery/2.1.4/jquery.min.js"></script>
<!--<![endif]-->

<!--[if lt IE 9]>
<script src="http://libs.baidu.com/jquery/1.11.3/jquery.min.js"></script>
<script src="http://cdn.staticfile.org/modernizr/2.8.3/modernizr.js"></script>
<script src="assets/js/amazeui.ie8polyfill.min.js"></script>
<![endif]-->
<script src="assets/js/amazeui.js"></script>
```

`amazeui.ie8polyfill.js` 中包含以下 polyfill：

- [Console-polyfill](https://github.com/paulmillr/console-polyfill)
- [es5-shim](https://github.com/es-shims/es5-shim)
- [es5-sham](https://github.com/es-shims/es5-shim)
- [EventListener Polyfill](https://gist.github.com/jonathantneal/3748027)
- [REM-unit-polyfill](https://github.com/chuckcarpenter/REM-unit-polyfill)
- [Respond.js](https://github.com/scottjehl/Respond)

#### HTML5 新元素

以下任意引入一个即可，Amaze UI 官网引的是 Modernizr。

- [Modernizr](https://github.com/Modernizr/Modernizr)
- [HTML5 Shiv](https://github.com/aFarkas/html5shiv)

#### Media Query

- [Respond.js](https://github.com/scottjehl/Respond)

#### rem

- [REM unit polyfill](https://github.com/chuckcarpenter/REM-unit-polyfill)

#### `box-sizing`

> IE 8 ignores `box-sizing: border-box` if min/max-width/height is used.

#### 伪元素

IE 8 只支持 CSS 2.1 规范中的单冒号语法（`:before`/`:after`），不支持 CSS3 的双冒号语法(`::before`/`::after`)。

#### 字体图标

参见 Bootstrap 中的 [issue 及里面提供的解决方法](https://github.com/twbs/bootstrap/issues/13863)。

## 关于 IE 6/7

**Amaze UI 不会支持 IE 6~7**。
