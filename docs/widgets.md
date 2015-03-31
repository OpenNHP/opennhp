---
id: widgets
title: Web 组件
titleEn: Web Components
permalink: widgets.html
next: widgets/accordion.html
---

# Web 组件
---

Amaze UI Web 组件把一些常见的网页组件拆分成不同的部分，进行类似 Web Components 的封装（[Web 组件简介](/getting-started/widget-dev)、[开发规范](/getting-started/widget)）。

- **文档中提及的「云适配 WebIDE」为云适配开发的网站适配工具，[详情请点击访问](http://platform.yunshipei.com/)。**

<div class="am-alert am-alert-danger"><strong>注意：</strong>Web 组件不考虑 IE 8/9 支持。</div>

## Web 组件演示

Amaze UI 目前封装了 10 余个主要面向移动端的 Web 组件（[独立窗口中浏览 Demo](/widgets/m)）。

<iframe src="/widgets/m" frameborder="0" id="doc-widget-frame" frameborder="0"></iframe>

## 使用场景

### 使用模板

通过模板（hbs）将数据和 HTML 分离，这是 Web 组件的价值之一。用户可以在不同的开发环境中使用 Web 组件。

#### 纯浏览器环境

Amaze UI 提供的开发模板中，包含一个 `widget.html` 文件，里面展示了 Widget 在纯浏览器环境中的使用。

__要点如下：__

1. 引入 Handlebars 模板 `handlebars.min.js`；
2. 引入 Amaze UI Widget helper `amui.widget.helper.js`；
3. 根据需求编写模板 `<script type="text/x-handlebars-template" id="amz-tpl">{{>slider slider}}</script>`；
4. 传入数据，编译模板并插入页面中。

```javascript
$(function() {
  var $tpl = $('#amz-tpl');
  var source = $tpl.text();
  var template = Handlebars.compile(source);
  var data = {};
  var html = template(data);

  $tpl.before(html);
});
```

#### Node.js 环境

可以结合 [Express.js](http://github.com/visionmedia/express)、[hbs](https://github.com/donpark/hbs) 使用。

用户了可以直接使用打包好的模块 [Amaze UI Widget hbs helper](https://www.npmjs.org/package/amui-hbs-helper)，[example](https://github.com/Minwe/amui-hbs-helper/tree/master/example) 里有完整的使用示例。

当然，你也可以自由调用：

__首先__，把 Web 组件的模板注册为 `partial`。

```javascript
// ...
var hbs = require('hbs');

app.set('view engine', 'hbs');

hbs.registerPartials(widgetDir + '/slider/src');
```

__然后__， 在页面模板中调用 `partial`，其中 `data` 为组件对应的数据。

```javascript
{{>slider data}}
```

#### PHP

参考以下链接：

- https://github.com/zordius/lightncandy
- https://github.com/XaminProject/handlebars.php
- http://www.sitepoint.com/sharing-templates-between-php-and-javascript/

#### Java

- [Handlebars.java](https://github.com/jknack/handlebars.java)
- [When Handlebars.js met Handlebars.java](http://jknack.github.io/handlebars.java/meeting.html)

#### 必须的 helper

无论你在上面那种环境中使用 Web 组件，都必须注册下面的 helper（我们提供的 `amui.widget.helper.js` 和 Node.js 模块中已经注册）:

```javascript
(function(hbs) {
  hbs.registerHelper("ifCond", function(v1, operator, v2, options) {
    switch (operator) {
      case "==":
        return (v1 == v2) ? options.fn(this) : options.inverse(this);

      case "!=":
        return (v1 != v2) ? options.fn(this) : options.inverse(this);

      case "===":
        return (v1 === v2) ? options.fn(this) : options.inverse(this);

      case "!==":
        return (v1 !== v2) ? options.fn(this) : options.inverse(this);

      case "&&":
        return (v1 && v2) ? options.fn(this) : options.inverse(this);

      case "||":
        return (v1 || v2) ? options.fn(this) : options.inverse(this);

      case "<":
        return (v1 < v2) ? options.fn(this) : options.inverse(this);

      case "<=":
        return (v1 <= v2) ? options.fn(this) : options.inverse(this);

      case ">":
        return (v1 > v2) ? options.fn(this) : options.inverse(this);

      case ">=":
        return (v1 >= v2) ? options.fn(this) : options.inverse(this);

      default:
        return eval("" + v1 + operator + v2) ? options.fn(this) : options.inverse(this);
    }
})(Handlebars);
```

### 不使用模板

不使用模板似乎就失去了 Web 组件的核心价值，但有用户可能真想这么用。

这就回归到最原始的网页书写方式了：

1. 引入 Amaze UI 的 CSS 和 JS 文件；
2. 按照 Web 组件的模板组织 HTML（可以点击左侧菜单进入组件演示页面拷贝示例代码，当然，你也可以使用其他模板引擎，只要渲染出来的结构符合跟 Web 组件相同就行）。
