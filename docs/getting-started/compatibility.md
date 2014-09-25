# Compatibility
---

> 尊敬的用户你好！我们很遗憾的通知你，你正在查看的 Amaze UI 目前只支持 IE10+ 及其他 H5 浏览器。古董浏览器支持时间请等候通知。

在机场听到类似的广播时，有人习以为常，有人愤怒，有人沮丧……

听到 Amaze UI 的浏览器支持情况时，你什么感受呢？至少应该兴奋一下：太好了，以后再也不用考虑那些蹩脚的古董浏览器了。当你嘴角带着微笑沉醉其中时，你项目经理跑过来跟你说：xx 页面在 xx 浏览器下有问题……

原来是黄粱一梦，每个前端开发者都做过的梦。

Amaze UI 由移动端发展而来，由于 Zepto 的原因，目前只支持 IE10+ 及其他 H5 浏览器，后续的版本中会考虑增加 `IE8-9` 的支持，Zepto 可能会被替换为 jQuery。

## 浏览器分级支持（<abbr title="Graded Browser Support">GBS</abbr>）

[<abbr title="Graded Browser Support">GBS</abbr>](https://yuilibrary.com/yui/docs/tutorials/gbs/) 是 YUI 团队提出的应对日益增长的浏览器兼容问题的思路，详情可以查看 [YUI 相关页面](https://yuilibrary.com/yui/docs/tutorials/gbs/)。

### Amaze UI GBS 描述

- __A 级：__最高支持级别，充分利用 H5 和 CSS3 等技术，提供最优的视觉和交互效果。
- __B 级：__有限支持，基本的样式和正常的交互，不考虑视觉、交互效果。
- __C 级：__核心支持，显示语义化的 HTML 标记渲染的内容，不考虑样式和行为。
- __X 级：__未知、零散的很少使用或已经停止开发的浏览器，可能不支持，也可能支持。

### Amaze UI GBS

Amaze UI 对浏览器做了一个粗略分级，这个列表会进一步细化、完善。

<div class="am-alert am-alert-warning">
  由于 Amaze UI 1.x 中使用了 <a href="http://zeptojs.com/#browsers">Zepto.js</a>，只能提供 A 级浏览器支持，B 级浏览器支持将在 2.x 中提供。
</div>

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th style="width: 64px">级别</th>
    <th>浏览器</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td rowspan="8">A 级</td>
    <td>Internet Explorer 10+</td>
  </tr>
  <tr>
    <td>Chrome†</td>
  </tr>
  <tr>
    <td>Firefox†</td>
  </tr>
  <tr>
    <td>Opera Next †</td>
  </tr>
  <tr>
    <td>360 浏览器† 极速模式</td>
  </tr>
  <tr>
    <td>搜狗浏览器† 极速模式</td>
  </tr>
  <tr>
    <td>Safari: iOS 6.†, iOS 7.†, Safari 6.1(OS X 10.8), Safari 7.†(OS X 10.9)
    </td>
  </tr>
  <tr>
    <td>Android 4.†: UC†, Chrome†, 自带浏览器</td>
  </tr>
  <tr>
    <td rowspan="5">B 级</td>
    <td>Internet Explorer 9 (<strong>IE9 对 H5 的支持非常有限，归为 B 级</strong>) <br/>
      Internet Explorer 8</td>
  </tr>
  <tr>
    <td>360 浏览器† IE8 内核</td>
  </tr>
  <tr>
    <td>搜狗浏览器† IE8 内核</td>
  </tr>
  <tr>
    <td>Safari: iOS 5.†
    </td>
  </tr>
  <tr>
    <td>Android 2.3.†: 自带浏览器</td>
  </tr>
  <tr>
    <td>C 级</td>
    <td>Internet Explorer lte 7</td>
  </tr>
  </tbody>
</table>

__注释：__

- `†` 表示最新正式版

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
