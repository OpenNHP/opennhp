# iScroll Lite
---

Amaze UI 默认打包了 iScroll Lite（当前版本为 5.1.3），具体使用请[参考官网](http://iscrolljs.com/)。

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
var wrapper = document.getElementById('wrapper');
var IScroll = $.AMUI.iScroll;
var myScroll = new IScroll(wrapper);
```

**目前仅是曝露接口给有经验的用户使用，更多功能实现后续添加。**
