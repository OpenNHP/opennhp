---
id: javaScript
title: JS 插件
titleEn: JavaScript
permalink: javaScript.html
next: javascript/alert.html
---

# JavaScript
---

## Basic Usage

### Based on jQuery

Amaze UI JavaScript Components require [jQuery](http://jquery.com/) since version 2.0. Please make sure the latest version of jQuery is loaded before Amaze UI script.

### Using Components

Using Amaze UI JS Components is just like using other jQuery plugins. Details can be found in documents of each component.

### Differences between jQuery and Zepto.js

jQuery and Zepto.js seems to be similar, but actually they are different in detail. Supporting jQuery and Zepto.js is not that worthy to do. Maybe that's why [Foundation 5 dropped Zepto](http://zurb.com/article/1293/why-we-dropped-zepto). ([Difference Demo listed below](http://jsbin.com/noxuvi/1/edit?html,css,js,console)）

#### `width()`/`height()`

- Zepto.js: Decided by box model(`box-sizing`).
- jQuery: Dismiss box model, always return width/height of content(not include `padding` and `border`)

jQuery [Official Explanation](http://api.jquery.com/width/#width)：

> Note that `.width()` will always return the content width, regardless of the value of the CSS `box-sizing` property. As of jQuery 1.8, this may require retrieving the CSS width plus `box-sizing` property and then subtracting any potential border and padding on each element when the element has `box-sizing: border-box`. To avoid this penalty, use `.css("width")` rather than `.width()`.

Solution is using `.css('width')` instead of `.width()` in jQuery.

**Well, what jQuery does here is questionable**, For example, `$('.box').css('height')` will return `20px`, **Whaaaaaat the f**k**?

```html
<style>
  .box {
    box-sizing: border-box;
    padding: 10px;
    height: 0;
  }
</style>

<div class="box"></div>
```

##### Getting height and width of a triangle

Suppose we draw a small lovely triangle using following HTML and CSS:

```html
<div class="caret"></div>
```

```css
.caret {
  width: 0;
  height: 0;
  border-width: 0 20px 20px;
  border-color: transparent transparent blue;
  border-style: none dotted solid;
}
```

- In jQuery, using `.width()` and `.css('width')` will return `0`, same to height.
- In Zepto, using `.width()` returns `40`, but using `.css('width')` returns `0px`。

So, in this case, **use `.outerWidth()`/`.outerHeight()` in jQuery; use `.width()`/`.height()` in Zepto**。

#### `offset()`

- Zepto.js: Return `top`, `left`, `width`, `height`
- jQuery: Return `width`, `height`

#### `$(htmlString, attributes)`

- [jQuery Document](http://api.jquery.com/jQuery/#jQuery-html-attributes)
- [Zepto Document](http://zeptojs.com/#$())

##### Difference in DOM Manipulation

```js
$(function() {
  var $list = $('<ul><li>jQuery Insertion</li></ul>', {
    id: 'insert-by-jquery'
  });
  $list.appendTo($('body'));
});
```
In jQuery, `id` won't be added to `ul`; In Zepto, it will。

##### Difference in event trigger

```js
$script = $('<script />', {
  src: 'http://cdn.amazeui.org/amazeui/1.0.1/js/amazeui.min.js',
  id: 'ui-jquery'
});

$script.appendTo($('body'));

$script.on('load', function() {
  console.log('jQ script loaded');
});
```

In jQuery, the process function of `load` event won't be executed; In Zepto, It will。

**其他参考链接：**

- [Market Share of jQuery](http://w3techs.com/technologies/overview/javascript_library/all)
- [jQuery vs Zepto in Performance](http://jsperf.com/zepto-vs-jquery-2013/82)

## Advanced Usage

### Default Initialize Events Interface

Amaze UI use specific HTML to bind events. Most of the JS components can be used simply by using HTML tags.

Default initialize events are all in namespace `xx.amui.data-api`. Developers can turn them off as their will.

Turn off all default events:

```javascript
$(document).off('.data-api');
```

Turn off default events for specific component:

```javascript
$(document).off('.modal.amui.data-api');
```

### Customized Events

Some components provide customized events, which is named by `{事件名称}.{组件名称}.amui`. Read documents of components for more details.

```javascript
$('#myAlert').on('close.alert.amui', function() {
  // do something
});
```

### MutationObserver

Two-way data binding is cool? [Mutation Observer](http://www.w3.org/TR/dom/#mutation-observers) is (or will be) the unsung hero.

We experimentally include `MutationObserver` in Amaze UI 2.1 , **Be careful using it**.

#### `data-am-observe`

After adding `data-am-observe` to an element, Amaze UI JS plugins that dynamiclly inserted will be automatically initiallized([Example](/javascript/scrollspy#mutationobserver?_ver=2.x)), 
  Supported plugins: Button、Dropdown、Slider、Popover、ScrollSpy、Tabs。

#### `$().DOMObserve(options, callback)`

- `options`: Attributes of observing([Details](https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver#MutationObserverInit)). Default: `{childList: true, subtree: true}`；
- `callback(mutations, observer)`: Processing function when DOM changes. The first parameter is the array that stores [MutationRecord](https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver#MutationRecord) object; the second parameter is the MutationObserver instance itself。

`````html
<p id="js-do-actions">
  <button class="am-btn am-btn-primary" data-insert>Insert p element</button>
  <button class="am-btn am-btn-secondary" data-addClass>Add Class</button>
  <button class="am-btn am-btn-warning" data-remove>Remove p element</button>
</p>
<div id="js-do-demo">
  <p>Example of DOM change observing. See console for log.</p>

</div>
<script>
  $(function() {
    var $wrapper = $('#js-do-demo');
    $wrapper.DOMObserve({
      childList: true,
      attributes: true,
      subtree: true
    }, function(mutations, observer) {
      console.log(observer.constructor === window.MutationObserver);
      console.log('DOM of #js-do-demo is changing：' + mutations[0].type);
    });

    $('#js-do-actions').find('button').on('click', function(e) {
      var $t = $(e.target);
      if ($t.is('[data-insert]')) {
        $wrapper.append('<p>a p is inserted</p>');
      } else if($t.is('[data-remove]')) {
        $wrapper.find('p').last().remove();
      } else {
        $wrapper.addClass('am-text-danger');
      }
    });
  })
</script>
`````
```html
<p id="js-do-actions">
  <button class="am-btn am-btn-primary" data-insert>Insert p element</button>
  <button class="am-btn am-btn-secondary" data-addClass>Add Class</button>
  <button class="am-btn am-btn-warning" data-remove>Remove p element</button>
</p>
<div id="js-do-demo">
  <p>Example of DOM change observing. See console for log.</p>
</div>
<script>
  $(function() {
    var $wrapper = $('#js-do-demo');
    $wrapper.DOMObserve({
      childList: true,
      attributes: true,
      subtree: true
    }, function(mutations, observer) {
      console.log(observer.constructor === window.MutationObserver);
      console.log('DOM of #js-do-demo is changing: ' + mutations[0].type);
    });

    $('#js-do-actions').find('button').on('click', function(e) {
      var $t = $(e.target);
      if ($t.is('[data-insert]')) {
        $wrapper.append('<p>a p is inserted</p>');
      } else if($t.is('[data-remove]')) {
        $wrapper.find('p').last().remove();
      } else {
        $wrapper.addClass('am-text-danger');
      }
    });
  })
</script>
```

**Reference：**

- [MDN - MutationObserver](https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver)；
- [CIU - Mutation Observer Browser Support](http://caniuse.com/#feat=mutationobserver)
- [Polyfill - MutationObserver.js](https://github.com/webcomponents/webcomponentsjs/blob/master/src/MutationObserver/MutationObserver.js)

### Modular Development

We did an survey about [front-end JS modularization](/javascript?_ver=1.x), when Amaze UI was in version 1.0 . Until Nov 13th, 2014, we recieved 1869 votes：

- CMD - Sea.js  23.86%  (446 votes)
- AMD - RequireJS  24.51%  (458 votes)
- CommonJS - Browserify  9.58%  (179 votes)
- Other(or develop a new one)  8.19%  (153 votes)
- What is JS Modularization? Some sort of dessert?  34%  (633 votes)

<div class="am-progress">
  <div class="am-progress-bar" style="width: 23.8%" data-am-popover="{content: 'CMD - Sea.js  23.86%  (446 votes)', trigger: 'hover focus'}">CMD</div>
  <div class="am-progress-bar am-progress-bar-secondary" data-am-popover="{content: 'AMD - RequireJS  24.51%  (458 votes)', trigger: 'hover focus'}" style="width: 24.5%" >AMD</div>
  <div class="am-progress-bar am-progress-bar-success" style="width: 9.5%" data-am-popover="{content: 'CommonJS - Browserify  9.58%  (179 votes)', trigger: 'hover focus'}">CJS</div>
  <div class="am-progress-bar am-progress-bar-warning" style="width: 8.2%" data-am-popover="{content: 'Other(or develop a new one)  8.19%  (153 votes)', trigger: 'hover focus'}">other</div>
  <div class="am-progress-bar am-progress-bar-danger" style="width: 34%" data-am-popover="{content: 'What is JS Modularize? Some sort of dessert?  34%  (633 votes)', trigger: 'hover focus'}">unknown</div>
</div>

Obviously, **Modularization is the future**. [ES6](http://wiki.ecmascript.org/doku.php?id=harmony:modules) will have native support for modularization.

Amaze UI 2.0 follows [CommonJS](http://wiki.commonjs.org/wiki/CommonJS) standard to organize modules(write modules like node.js in front-end). Developers can choose package tools they like.

- [Browserify](http://browserify.org/)：Work with NPM, Browserify can manage front-end modules. Many front-end modules have been published to NPM, so we can drop some tools like Bower;
- [Duo](http://duojs.org/)：Besides managing local modules, Duo can help us fetch open source projects directly from GitHub. Support Javascript, HTML and CSS;
- [gulp-amd-bundler](https://www.npmjs.org/package/gulp-amd-bundler)：Package modules coded in CJS to AMD modules;
- [Webpack](https://github.com/webpack/webpack)。

[SPM](http://spmjs.io/) Can't extract dependancies from source codes. If you are are using Sea.js, you probably need to modify your package tool by yourself.

__Further Reading:__

* [History of Front-end Modular Development](https://github.com/seajs/seajs/issues/588)
* [Writing Modular JavaScript With AMD, CommonJS & ES Harmony](http://addyosmani.com/writing-modular-js/)
