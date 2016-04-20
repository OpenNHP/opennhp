---
id: javascript
title: JavaScript 规范
titleEn: JavaScript Guide
permalink: javascript.html
prev: html-css.html
next: widget.html
---

# Amaze UI JavaScript 规范
---

## 基本编码规范

* [AllMobilize JavaScript Style Guide](/getting-started/javascript-guide)
* [CMD 模块定义规范](https://github.com/seajs/seajs/issues/242)

## 代码质量控制工具

Amaze UI 使用 ~~[JSHint](http://jshint.com/) 和 [JSCS](https://github.com/jscs-dev/node-jscs)~~[ESLint](http://eslint.org/)控制代码质量。

详细设置参见 ~~[.jshintrc](https://github.com/amazeui/amazeui/blob/master/.jshintrc)、[.jscsrc](https://github.com/amazeui/amazeui/blob/master/.jscsrc)~~[.eslintrc](https://github.com/amazeui/amazeui/blob/master/.eslintrc)。

> 2016.04.20 替换为 ESLint，参见 [Welcoming JSCS to ESLint](http://eslint.org/blog/2016/04/welcoming-jscs-to-eslint)

（部分直接使用第三方库的代码未通过质量控制工具检测。）

## jQuery / Zepto.js 使用规范

为提高代码执行效率，为二者兼容提供可能，在使用 jQuery / Zepto.js 时做以下约定：

- 存放 jQuery / Zepto 对象的变量以 `$` 开头；
- 禁止使用 `slideUp/Down()` `fadeIn/fadeOut()` 等方法；
- 尽量不使用 `animate()` 方法；
- 使用和 Zepto.js 兼容的基本选择符，不使用效率较低且与 Zepto.js 不兼容的选择符。


__问题：__

- **自定义事件命名空间：** Zepto.js 不支持 `.` 语法，只能使用 `:` 语法。
- http://zeptojs.com/#event
- http://api.jquery.com/event.namespace/

## 代码格式

- **缩进 2 个空格**；
- **使用多 `var` 模式声明变量**：更容易维护，比如要删除第一个变量或者最后一个变量时，多 `var` 模式直接删除即可，单 `var`
还要去修改别的行（**`for` 循环例外**）；

***Valid***

```javascript
var x = 1;
var y = 2;

for (var i = 0, j = arr.length; i < j; i++) {
}
```

***Invalid***

```javascript
var x = 1,
    y = 2;
```

## 命名规范

### 基本原则

1. 文件和目录名只能包含 `[a-z\d\-]`，并以英文字母开头
2. 首选合适的英文单词
3. Data API 命名使用小写、用连字符连接、添加 `am` 命名空间，如 `data-am-trigger`
4. 事件名使用小写字母，包含模块名及 `amui` 命名空间名，使用 `:` 连接（Zepto 不支持使用 `.` 链接的自定义事件），如 `.trigger('open:modal:amui')`
5. 符合规范
   - 常量全大写 `UPPERCASE_WORD`
   - 变量驼峰 `camelName`
   - 类名驼峰，并且首字母要大写 `CamelName`


### HTML Data API

- 基本: `data-am-{组件名}`，如 `data-am-modal`、`data-am-navbar-qrcode`
- 传参: `data-am-modal="{key1: 'val1', key2: false}"`，core.js 中增加一个专门解析参数的函数

### JavaScript

- 自定义事件命名：`{自定义事件名}:{组件名}:{后缀}`，Zepto 不支持 `.` 分隔的自定义事件语法，官方示例中使用 `:` 分隔，如 `closed:modal:amui`。Zepto 中没有 [event.namespace](http://api.jquery.com/event.namespace/)，这样的命名方式只用于清晰区分不同模块的自定义事件。另外，按照 Zepto 官方示例，应该写成 `amui:modal:closed`，为跟 jQuery 的写法统一，颠倒顺序书写。
- 默认绑定事件：事件名（内置事件，非自定义事件）采用 `{事件名}.{组件名}.{命名空间}`，如 `$(document).on('click.modal.amui',...`。
     * 取消所有默认绑定事件： `$(document).off('.amui',...`
     * 取消特定组件的默认绑定事件： `$(document).off('.modal.amui',...`

## 接口命名规范

通过接口规范，统一模块对外接口的命名，形成一致的编写习惯。

__规则：__

* **可读性强，见名晓义。**
* 尽量不与 jQuery 社区已有的习惯冲突。
* 尽量写全。不用缩写，除非是下面列表中约定的。（变量以表达清楚为目标，uglify 会完成压缩体积工作）

<table class="am-table am-table-bd am-table-striped">
  <thead>
    <tr>
      <th>常用词</th>
      <th>说明</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>options</td>
      <td>表示选项，与 jQuery 社区保持一致，不要用 config, opts 等</td>
    </tr>
    <tr>
      <td>active</td>
      <td>表示当前，不要用 current 等</td>
    </tr>
    <tr>
      <td>index</td>
      <td>表示索引，不要用 idx 等</td>
    </tr>
    <tr>
      <td>trigger</td>
      <td>触点元素</td>
    </tr>
    <tr>
      <td>triggerType</td>
      <td>触发类型、方式</td>
    </tr>
    <tr>
      <td>context</td>
      <td>表示传入的 this 对象</td>
    </tr>
    <tr>
      <td>object</td>
      <td>推荐写全，不推荐简写为 o, obj 等</td>
    </tr>
    <tr>
      <td>element</td>
      <td>推荐写全，不推荐简写为 el, elem 等</td>
    </tr>
    <tr>
      <td>length</td>
      <td>不要写成 len, l</td>
    </tr>
    <tr>
      <td>prev</td>
      <td>previous 的缩写</td>
    </tr>
    <tr>
      <td>next</td>
      <td>next 下一个</td>
    </tr>
    <tr>
      <td>constructor</td>
      <td>不能写成 ctor</td>
    </tr>
    <tr>
      <td> easing</td>
      <td>示动画平滑函数</td>
    </tr>
    <tr>
      <td>min</td>
      <td>minimize 的缩写</td>
    </tr>
    <tr>
      <td>max</td>
      <td>maximize 的缩写</td>
    </tr>
    <tr>
      <td>DOM</td>
      <td>不要写成 dom, Dom</td>
    </tr>
    <tr>
      <td>.hbs</td>
      <td>使用 hbs 后缀表示模版</td>
    </tr>
    <tr>
      <td>btn</td>
      <td>button 的缩写</td>
    </tr>
    <tr>
      <td>link</td>
      <td>超链接</td>
    </tr>
    <tr>
      <td>title</td>
      <td>主要文本</td>
    </tr>
    <tr>
      <td>img</td>
      <td>图片路径（img标签src属性）</td>
    </tr>
    <tr>
      <td>dataset</td>
      <td>html5 data-xxx 数据接口</td>
    </tr>
    <tr>
      <td>theme</td>
      <td>主题</td>
    </tr>
    <tr>
      <td>className</td>
      <td>类名</td>
    </tr>
    <tr>
      <td>classNameSpace</td>
      <td>class 命名空间</td>
    </tr>
  </tbody>
</table>


## 注释规范

### 总原则

* **As short as possible（如无必要，勿增注释）**：尽量提高代码本身的清晰性、可读性。
* **As long as necessary（如有必要，尽量详尽）**：合理的注释、空行排版等，可以让代码更易阅读、更具美感。

总之，注释的目的是： **提高代码的可读性，从而提高代码的可维护性。**


### 什么时候需要添加注释

- 某段代码的写法，需要注释说明 why 时：

```js
// Using loop is more efficient than `rest = slice.call(arguments, 1)`.
for (i = 1, len = arguments.length; i < len; i++) {
    rest[i - 1] = arguments[i];
}
```

- 添加上注释，能让代码结构更清晰时：

```js
init: function(selector, context, rootjQuery) {
    var match, elem, ret, doc;

    // Handle $(""), $(null), or $(undefined)
    if ( !selector ) {
        ...
    }

    // Handle $(DOMElement)
    if ( selector.nodeType ) {
        ...
    }

    // The body element only exists once, optimize finding it
    if ( typeof selector === "string" ) {
        ...
     }
}
```

- 有借鉴、使用第三方代码，需要说明时：

```js
// Inspired by https://github.com/jquery/jquery/blob/master/src/core.js
function ready() {
    ...
}
```

- 文件最后空一行，可以保证在 combo 合并后，源码的层次清晰。

###  注释书写规范

1. 源码中的注释，推荐用英文。
2. 含有中文时，标点符号用中文全角。
3. 中英文夹杂时， __英文与中文之间要用一个空格分开__。
4. 注释标识符与注释内容要用一个空格分开：`// 注释` 与 `/* 注释 */`。

## 文档规范

### README.md

每个组件必须有 README.md 文件，用来描述组件的基本情况。

```
# 模块名称

-----

该模块的概要介绍。

------

## 使用说明

如何使用该模块，可以根据组件的具体特征，合理组织。

## API

需要提供 API 说明，属性、方法、事件等。

## 使用示例
```

### HISTORY.md

记录组件的变更，最好和 `issues` 进行关联。请阅读[历史记录书写规范](/getting-started/history)。


## 参考链接

Amaze UI 的编码规范参考了社区里一些先行者的做法，在此致谢！

-  [注释规范](https://github.com/aralejs/aralejs.org/wiki/JavaScript-%E6%B3%A8%E9%87%8A%E8%A7%84%E8%8C%83)
-  [编码风格](https://github.com/aralejs/aralejs.org/wiki/JavaScript-%E7%BC%96%E7%A0%81%E9%A3%8E%E6%A0%BC)
-  [编码与文档的讨论](https://github.com/aralejs/aralejs.org/issues/36)
-  [常用词命名统一表](https://github.com/aralejs/aralejs.org/wiki/%E5%B8%B8%E7%94%A8%E8%AF%8D%E5%91%BD%E5%90%8D%E7%BB%9F%E4%B8%80%E8%A1%A8)
