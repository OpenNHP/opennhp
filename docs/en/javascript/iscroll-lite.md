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

Amaze UI use iScroll `lite`(version 5.1.3) by default. More details in [Official Website](http://iscrolljs.com/)。

> iscroll-lite.js, it is a stripped down version of the main script. It doesn't support snap, scrollbars, mouse wheel, key bindings. But if all you need is scrolling (especially on mobile) iScroll lite is the smallest, fastest solution.


Use `$.AMUI.iScroll` to construct `IScroll` instance.

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

iScroll is used to smooth the vertical and horizontal drag operations and controls.

Most of iScroll examples can be found on internet is based on iScroll 4. However, iScroll API changed a lot in `5.x` version, so many `4.x` examples don't work with `5.x` iScroll.

- [Fix Top using iSroll](/widgets/m?_ver=2.x)
- [Use iScroll 5 to implement a drag down refreshing and drag up loading](/examples/iscroll.html)
- [iPhone Scrollbars with iScroll (iScroll 4)](http://davidwalsh.name/iphone-scrollbars)

More examples in [Repo Homepage](https://github.com/cubiq/iscroll/tree/master/demos)，**Attension: iScroll in Amaze UI is `lite` version. Some Demo may not work with Amaze UI.**

## FAQ

### Why I can't click on links after using iScroll?

Add `click: true` when initializing iScroll, and the problem will be solved (credit to [FB总司令](http://weibo.com/songzibin))。

```js
var myScroll = new IScroll('#wrapper', {
  click: true
});
```
