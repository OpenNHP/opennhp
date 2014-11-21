# Compatibility
---

Amaze UI 面向现代浏览器开发，对 IE 8/9 等浏览器只提供有限支持。

## 浏览器分级支持（<abbr title="Graded Browser Support">GBS</abbr>）

[<abbr title="Graded Browser Support">GBS</abbr>](https://yuilibrary.com/yui/docs/tutorials/gbs/) 是 YUI 团队提出的应对日益增长的浏览器兼容问题的思路，详情可以查看 [YUI 相关页面](https://yuilibrary.com/yui/docs/tutorials/gbs/)。

### Amaze UI GBS 描述

- __A 级：__最高支持级别，充分利用 H5 和 CSS3 等技术，提供最优的视觉和交互效果。
- __B 级：__有限支持，基本的样式和正常的交互，不考虑视觉、交互效果。
- __C 级：__核心支持，显示语义化的 HTML 标记渲染的内容，不考虑样式和行为。
- __X 级：__未知、零散的很少使用或已经停止开发的浏览器，可能不支持，也可能支持。

### Amaze UI GBS

按照国际惯例，Amaze UI 提供对主流浏览器（系统）最近两个稳定版本的全面支持。结合国内实际情况，一些浏览器的支持缩减为最新正式版，IE 则对更老版本做了有限支持。

Amaze UI 对浏览器做了一个粗略分级，**优先支持 A 级浏览器**。由于资源有限，无法列出所有的浏览器，**使用 `WebKit` 的浏览器只要不乱修改内核，理论上应该都支持**。

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
      <th scope="row">WinPhone(8+)</th>
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
    <th scope="row">Opera Next</th>
    <td class="`">L1</td>
    <td class="am-success">A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-success">A</td>
    <td class="am-disabled">N/A</td>
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
    <th scope="row">Android stock<sup>1</sup></th>
    <td>L1</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">N/A</td>
    <td class="am-disabled">X</td>
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


- `L` 代表 `last`，L2 - 最新的两个稳定版本；L1 - 最新稳定版本。
- `1` 安卓系统自动浏览器，由于部分厂商对浏览器做了修改，列为 X 级。

__参考链接__：

- [iOS Version Stats](http://david-smith.org/iosversionstats/)

### IE 8/9

<!--<div class="am-alert am-alert-warning">
  请注意，我们对 IE 8/9 的支持的极限为<strong>布局、JS 交互基本正常</strong>，不会提供更多的支持。
</div>-->

- IE 8/9 不支持 `transition`，基本看不到任何动画效果；
- **Web 组件部分不提供 IE 8/9 官方支持**。

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

## 关于 IE 6~7

国内市场对 IE 6~7 支持还有一定需求，但对于一个前端开发者，我们应该去推动这个行业向前发展，而不是一味迁就、妥协。 __Amaze UI 不会支持 IE 6~7__。

曾经，能够提供支持老版本 IE 是一个前端开发者的必备技能。随着移动互联网大潮来临，这个技能不再那么重要。

对于有多年前端开发经验的开发者，支持 IE 6~7 应该不在话下，但应该靠自己的经验和影响力，引领那些新入行的开发者关注、使用前沿的技术，而不是因循守旧。对于新入行的开发者，我们（可能只是我）建议直接忽略老的 IE 浏览器，在老的浏览器上浪费时间、精力对你的职业生涯毫无裨益。这些浏览器很快就会被放进博物馆，你现在学习的针对这些旧浏览器的知识将变得一文不值。当然，如果有人致力成为浏览器考古专家，那另当别论。

我们曾天真的期望国内的互联网巨头能像 Google 一样，放弃对老版本 IE 的支持，依靠他们庞大的用户量和影响力，引领用户使用体验更好的浏览器。显然，是我们想多了。商业利益何其重要，节操何其稀少，正如微软停止 XP 技术支持，有人迅速来插一脚一样。

作为这个行业从业者，我们应该尝试说服客户，说服主管，不要在旧浏览器上浪费人力物力。如果无法改变他们，可以改变自己，越来越多公司只针对现代浏览器开发（云适配便是其一）。把搞兼容老版浏览器的时间拿来享受生活，陪陪家人、陪陪妹子，多好。什么？你东西南北漂，家太远；也没有妹子？那你还不赶快去找妹子！天天搞古董浏览器，注定孤独到老。

我们无意批评老版 IE，只是说她已经完成她的使命，到了寿终正寝的时候了。

对于不愿放弃老版 IE 用户的，给两个解决思路：

- 给一个二维码，让他用手机扫描访问；
- 让他升级浏览器。说服他升级别的浏览器的可能不太容易，装一个 360 浏览器之类的应该还是可以的。
