---
id: iscroll-lite
title: iScroll 元素滚动
titleEn: iScroll
prev: javascript/hammer.html
next: javascript/store.html
source: js/util.iscroll-lite.js
doc: docs/javascript/iscroll-lite.md
---

# iScroll Lite
---

Amaze UI 默认打包了 iScroll `lite`（当前版本为 5.1.3），具体使用请[参考官网](http://iscrolljs.com/)。

> iscroll-lite.js, it is a stripped down version of the main script. It doesn't support snap, scrollbars, mouse wheel, key bindings. But if all you need is scrolling (especially on mobile) iScroll lite is the smallest, fastest solution.


通过 `$.AMUI.iScroll` 访问 `IScroll` 对象。

```html
<div id="wrapper">
  <ul>
    <li>...</li>
    <li>...</li>
    ...
  </ul>
</div>
```

```js
var IScroll = $.AMUI.iScroll;
var myScroll = new IScroll('#wrapper');
```

iScroll 主要用来实现平滑的垂直、水平的拖动操作及控制。

网上很多 iScroll 的例子都是基于 iScroll 4 实现的，`5.x` API 变动很大，许多 `4.x` 的例子都不能正常工作，需要做调整。

- [iSroll 实现的固定顶部](/widgets/m?_ver=2.x)
- [iScroll 5 实现的下拉刷新和上拉加载更多](/examples/iscroll.html)
- [iPhone Scrollbars with iScroll (适用于 iScroll 4)](http://davidwalsh.name/iphone-scrollbars)

更多例子请参见[项目主页](https://github.com/cubiq/iscroll/tree/master/demos)，**请注意，Amaze UI 打包的是 `lite` 版，有些 Demo 并不适用。**

## 常见问题

### 使用 iScroll 以后链接无法点击？

初始化 iScroll 的时候加上 `click: true` 参数即可(感谢[FB总司令](http://weibo.com/songzibin))。

```js
var myScroll = new IScroll('#wrapper', {
  click: true
});
```
