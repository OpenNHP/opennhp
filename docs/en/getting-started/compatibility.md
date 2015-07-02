# Compatibility
---

Amaze UI is developed for modern browsers, and only provide limited support for browsers like ID 8/9.

**Attention: **

- **Please don't use unreliable test tools like `IETester`**；
- According to the [official opinion](https://www.modern.ie/en-us/f12) from Microsoft, ther browser mode in IE Developer Tools is also not reliable;
- Microsoft provids many[IE testing virtual machines](https://www.modern.ie/zh-cn/virtualization-tools#downloads).

## Graded Browser Support(<abbr title="Graded Browser Support">GBS</abbr>)

[<abbr title="Graded Browser Support">GBS</abbr>](https://github.com/yui/yui3/wiki/Graded-Browser-Support) is a solution raised by YUI team to deal with growing browser compatibility problem. See [YUI Page](https://github.com/yui/yui3/wiki/Graded-Browser-Support) for more details.

### Amaze UI GBS Description

- __A-grade: __A-grade support is the highest support level. By taking full advantage of the powerful capabilities of modern web standards, the A-grade experience provides advanced functionality and visual fidelity.
- __B-grade: __B-grade support is the limited support level. It provide basic styles and normal functionality. Advanced functionality and visual fidelity is not under consideration.
- __C-grade: __C-grade support is the core support level. Delivered via nothing more than semantic HTML, the content and experience is highly accessible, unenhanced by decoration or advanced functionality, and forward and backward compatible. Layers of style and behavior are omitted.
- __X 级：__X-grade provides support for unknown, fringe or rare browsers as well as browsers on which development has ceased. Browsers receiving X-grade support are assumed to be capable. X-grade browsers are all browsers not designated as any other grade.

### Amaze UI GBS

Like many other frameworks, Amaze UI fully supports the latest two stable version of popular browsers. Because of practical reasons, for some browsers, Amaze UI only support their latest official version. We also provided limited support for older version of IE.

Support for browsers is listed as below, **A-grade browsers have higher priority**。

In limit of space, We are not going to list all browsers. **Theoretically, browsers using `WebKit` are all supported as long as the core is not modified too much**.

More details about browsers support can be found [here](http://caniuse.com/)(data for UC browser has been collected).

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
    <th scope="row">UC Browser</th>
    <td>L1</td>
    <td class="am-primary">X</td>
    <td class="am-success">A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-success">A</td>
    <td class="am-success">A-</td>
  </tr>
  <tr>
    <th scope="row" rowspan="2">360 Browser</th>
    <td>L1-Speed Mode</td>
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
    <th scope="row" rowspan="2">Sogou Browser</th>
    <td class="am-text-nowrap">L1-Speed Mode</td>
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

__Comments：__

- `L` represents `last`，`L2` - Latest two stable versions；`L1` - Latest stable version.
- `1` Android's build-in browser. Because of the modification in different branches of Android, we list it as X-grade.

__Reference: __

- [iOS Version Stats](http://david-smith.org/iosversionstats/)

### IE 8/9

- IE 8/9 don't support `transition`, so there will be no animate effects;
- IE 9 has better support for ES5, so importing complete `amazeui.js` won't cause error. Error will be caused in IE8;
- **Official support is not provided for Web widgets in IE8/9**.

**JS plugins that limited support IE8/9**：

- Alert;
- Button;
- Collpase;
- Dropdown;
- Modal;
- Popover;
- Slider;
- OffCanvas;
- ScrollSpyNav;
- Sticky;
- Tabs - **Only IE 9**；

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
    <td class="am-success"><span class="am-icon-check"></span>YES（<code>-ms</code> prefix）</td>
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
  Attention: Our support for IE 8 is only in <strong>Layout and partial JS plugins</strong>, Issues related to IE 8 will be listed as `P4`(Lowest priority).
</div>

Developers need **support for IE 8 please import `amazeui.legacy.js` using conditial comments**, and import corresponding polyfill.

```html
﻿<!--[if (gte IE 9)|!(IE)]><!-->
<script src="http://libs.baidu.com/jquery/2.1.1/jquery.min.js"></script>
<script src="assets/js/amazeui.js"></script>
<!--<![endif]-->

<!--[if lt IE 9]>
<script src="http://libs.baidu.com/jquery/1.11.1/jquery.min.js"></script>
<script src="http://cdn.staticfile.org/modernizr/2.8.3/modernizr.js"></script>
<script src="assets/js/polyfill/rem.min.js"></script>
<script src="assets/js/polyfill/respond.min.js"></script>
<script src="assets/js/amazeui.legacy.min.js"></script>
<![endif]-->
```

#### HTML5 New Elements

Importing any one of these is fine, we used Modernizr for our official website.

- [Modernizr](https://github.com/Modernizr/Modernizr)
- [HTML5 Shiv](https://github.com/aFarkas/html5shiv)

#### Media Query

- [Respond.js](https://github.com/scottjehl/Respond)

#### rem

- [REM unit polyfill](https://github.com/chuckcarpenter/REM-unit-polyfill)

#### `box-sizing`

> IE 8 ignores `box-sizing: border-box` if min/max-width/height is used.

#### Pseudo-elements

IE 8 only support single colon in CSS 2.1(`:before`/`:after`), and don't support double colon in CSS3(`::before`/`::after`).

#### Font ICON

See [issue and solutions](https://github.com/twbs/bootstrap/issues/13863) in Boostrap.

## About IE 6/7

**Amaze UI won't support IE 6~7**。
