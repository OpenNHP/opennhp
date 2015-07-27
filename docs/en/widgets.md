---
id: widgets
title: Web 组件
titleEn: Web Widgets
permalink: widgets.html
next: widgets/accordion.html
---

# Web Widgets
---

We packed some frequently used web components into Amaze UI Web Widgets. ([Introduction](/getting-started/widget-dev), [Developing Web Widgets](/getting-started/widget)）。

- **The 「Allmobilize WebIDE」 refers to the website developing tool developed by Allmobilize Inc. See more details [here](http://platform.yunshipei.com/).**

<div class="am-alert am-alert-danger"><strong>Attention: </strong>Web widget in Amaze UI currently don't support IE 8/9.</div>

## Demo

Amaze UI has over 10 web widgets for mobile website development ([View demos in new window](__M_PAGE__/widgets/m)).

<iframe src="__M_PAGE__/widgets/m" frameborder="0" id="doc-widget-frame" frameborder="0"></iframe>

## Usage

### Using Templates

Using templates(hbs) to seperate the data and HTML is one of the most valuable feature. Users can use web widget in different development environment.

#### Pure browser environment

There is a `widget.html` file in the development templates. It helps to understand how to use widgets in pure browser environment.

__Procedure:__

1. Import Handlebars template `handlebars.min.js`;
2. Import Amaze UI Widget helper `amui.widget.helper.js`;
3. Modify the template according to your need `<script type="text/x-handlebars-template" id="amz-tpl">{{>slider slider}}</script>`；
4. Pass the data, compile the template and insert it into your page.

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

#### Node.js

Widget can also work with [Express.js](http://github.com/visionmedia/express) and [hbs](https://github.com/donpark/hbs).

You may directly use our node module, [Amaze UI Widget hbs helper](https://www.npmjs.org/package/amui-hbs-helper). Here is an [example](https://github.com/Minwe/amui-hbs-helper/tree/master/example).

Of course, you may call it in your way.

__First__, register the web widget template to be `partial`.

```javascript
// ...
var hbs = require('hbs');

app.set('view engine', 'hbs');

hbs.registerPartials(widgetDir + '/slider/src');
```

__Then__, Call `partial` in the page template, where `data` is the data of the widget.

```javascript
{{>slider data}}
```

#### PHP

See following links:

- https://github.com/zordius/lightncandy
- https://github.com/XaminProject/handlebars.php
- http://www.sitepoint.com/sharing-templates-between-php-and-javascript/

#### Java

- [Handlebars.java](https://github.com/jknack/handlebars.java)
- [When Handlebars.js met Handlebars.java](http://jknack.github.io/handlebars.java/meeting.html)

#### Helper(Required)

Now metter which environment above you are using, the following helpers must be registered. (We have already registered these in `amui.widget.helper.js` and Node.js module)

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

### No template

There seems to be no reason to use web widget without template, but just in case some developers want to do that, we will come back to the old fashion:

1. Import CSS and JS files of Amaze UI;
2. Write HTML according to the template of web widgets. ( Sample codes can be found in documents of each widgets)
