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

After trying different touch libraries, we finally decide to use [Hammer.js 2.0](https://github.com/hammerjs/hammer.js) in Amaze UI. More detail about Hammer in its [Official Docs](http://hammerjs.github.io/getting-started/).

## Usage

### `$().hammer()`

Hammer.js can be used like plugins.

```javascript
$(element).hammer(options).on('pan', myPanHandler);
```

Hammer instance is stored in `$(element).data('hammer')`.

### Hammer Instance

Use `Hammer` instance through `$.AMUI.Hammer`.

```js
var Hammer = $.AMUI.Hammer;
var hammertime = new Hammer(myElement, myOptions);

hammertime.on('pan', function(e) {
  console.log(e);
});
```

**Hammer.js is a little complicated. We will add more usage instructions later.**
