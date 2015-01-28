---
id: hammer
title: 触控事件
titleEn: Hammer
prev: javascript/fullscreen.html
next: javascript/iscroll-lite.html
source: js/util.hammer.js
doc: docs/javascript/hammer.md
---

# Hammer.js
---

几经尝试，Amaze UI 最终选择了 [Hammer.js 2.0](https://github.com/hammerjs/hammer.js)
作为触控操作库，更多细节参见[官方文档](http://hammerjs.github.io/getting-started/)。

## 调用方式

### `$().hammer()`

可以通过插件的形式调用 Hammer.js。

```javascript
$(element).hammer(options).on('pan', myPanHandler);
```

Hammer 实例存储在 `$(element).data('hammer')` 上。

### Hammer 对象

可以通过 `$.AMUI.Hammer` 访问 `Hammer` 对象。

```js
var Hammer = $.AMUI.Hammer;
var hammertime = new Hammer(myElement, myOptions);

hammertime.on('pan', function(e) {
  console.log(e);
});
```

**Hammer.js 略显复杂，后续会增加更多使用说明。**
