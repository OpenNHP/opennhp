---
id: getting-started
title: 开始使用
titleEn: Getting Started
permalink: getting-started.html
next: getting-started/layouts.html
---

# Getting Started With Amaze UI
---

Amaze UI is a lightweight (Only 100kB after gzip all CSS and JS files), [**Mobile first**](http://cbrac.co/113eY5h) front-end framework.
and is built based on those popular open-source front-end frameworks ([List of projects we used or refered to](https://github.com/allmobilize/amazeui#%E5%8F%82%E8%80%83%E4%BD%BF%E7%94%A8%E7%9A%84%E5%BC%80%E6%BA%90%E9%A1%B9%E7%9B%AE)), .

## Get Amaze UI

### Download

<div class="am-g">
  <div class="am-u-md-8 am-u-md-centered">
    <a id="doc-dl-btn" href="http://amazeui.org/download?ver=__VERSION__" class="am-btn am-btn-block am-btn-success am-btn-lg" onclick="window.ga && ga('send', 'pageview', '/download/AmazeUI.zip');
"><i class="am-icon-download"></i> Amaze UI v__VERSION__</a>
  </div>
</div>

**Offline Docs：**

<div class="am-g">
  <div class="am-u-sm-6"><a href="http://amazeui.org/download?ver=docs" class="am-btn am-btn-block am-btn-primary">HTML Docs</a></div>
  <div class="am-u-sm-6"><a href="http://amazeui.org/download?ver=dash" class="am-btn am-btn-block am-btn-warning">Dash Docsets</a></div>
</div>

Repos:

- [Amaze UI Docs](https://github.com/amazeui/docs)
- [Amaze UI Dash Docsets Generator](https://github.com/amazeui/docs-generator)

**Code Snippets：**

<div class="am-g">
  <div class="am-u-sm-6"><a href="http://amazeui.org/download?ver=jetbrains" class="am-btn am-btn-block am-btn-secondary">JetBrains Series Editor</a></div>
  <div class="am-u-sm-6"><a href="http://amazeui.org/download?ver=sublime" class="am-btn am-btn-block am-btn-danger">Sublime</a></div>
</div>

More Detail: [Amaze UI Snippets](https://github.com/amazeui/snippets)。

**Starter Kit：**

Front-end workflow built based on Gulp and NPM. Click [Here](https://github.com/amazeui/starter-kit) to visit the repo.

### Amaze UI CDN

Amaze UI CDN： DNS is provided by DNSPod. CDN is provided by Qiniu。

```html
http://cdn.amazeui.org/amazeui/__VERSION__/css/amazeui.css
http://cdn.amazeui.org/amazeui/__VERSION__/css/amazeui.min.css
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.min.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.legacy.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.legacy.min.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.widgets.helper.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.widgets.helper.min.js
```

### Install with Bower

```html
bower install amazeui
```

### Plugins and Examples

- [DateTimePicker - Select date](https://github.com/amazeui/datetimepicker)
- [Echo.js - Lazy-loading image](https://github.com/amazeui/echo)
- [Lazyload - Lazy-loading image](https://github.com/amazeui/lazyload)
- [Chosen - Enhance select boxes](https://github.com/amazeui/chosen)
- [Masonry - Cascading grid layout](https://github.com/amazeui/masonry)
- [Switch - Turn checkboxes and radio buttons in toggle switches](https://github.com/amazeui/switch)
- [Tags Input - Tag input field plugin](https://github.com/amazeui/tagsinput)
- [Amaze UI theme for Video.js](https://github.com/amazeui/videojs)
- [jQuery DataTables - Table plug-in](https://github.com/amazeui/datatables)

### Source Codes

You can get the source codes from our repo on github.

<iframe src="http://ghbtns.com/github-btn.html?user=amazeui&repo=amazeui&type=watch&count=true&size=large" allowtransparency="true" frameborder="0" scrolling="0" width="156px" height="30px"></iframe>

<iframe src="http://ghbtns.com/github-btn.html?user=amazeui&repo=amazeui&type=fork&count=true&size=large" allowtransparency="true" frameborder="0" scrolling="0" width="156px" height="30px"></iframe>

## Files and Versions

### Files

- `amazeui.css` / `amazeui.js`：Include all CSS and JS files of Amaze UI.
- `amazeui.flat.css`：Flat version of Amaze UI CSS. Check `1.x` for more information.
- `amazeui.legacy.js`：JS for IE 8.
- `amazeui.widgets.helper.js`： **For developers using Handlebars. If you are not using Handlebars, please dismiss this file.** Includes Handlebars helper and template partials for Web widgets.

Every file above has a related minified file.

### Versions

Amaze UI follows [Semantic Versioning](http://semver.org/lang/zh-CN/). Version number follows MAJOR.MINOR.PATCH format. The increment rule is:

- Increase MAJOR version when you make incompatible API changes,
- Increase MINOR version when you add functionality in a backwards-compatible manner, and
- Increase PATCH version when you make backwards-compatible bug fixes.

- [**Migration form from 1.x to 2.x**](https://github.com/allmobilize/amazeui/wiki/Migration-form-1.x-to-2.x)


## What's included

Within the package downloaded you'll find the following directories and files, including CSS, JS files and examples of Amaze UI.

- `index.html` - Blank HTML template；
- `blog.html` - Blog template（[preview](/examples/blog.html)）；
- `landing.html` - Landing Page template（[preview](/examples/landing.html)）；
- `login.html` - Login Page template（[preview](/examples/login.html)）；
- `sidebar.html` - Article template with sidebar[preview](/examples/sidebar.html)）；
- `admin-*.html` - simple administration template（[preview](/examples/admin-index.html)）
- Write your CSS in `app.css`；
- Write your JavaScript in `app.js`；
- Save images in `i`。

```
AmazeUI
|-- assets
|   |-- css
|   |   |-- amazeui.css             // All Styles for Amaze UI
|   |   |-- amazeui.min.css           // Around 42 kB (gzipped)
|   |   `-- app.css
|   |-- i
|   |   |-- app-icon72x72@2x.png
|   |   |-- favicon.png
|   |   `-- startup-640x1096.png
|   `-- js
|       |-- amazeui.js
|       |-- amazeui.min.js                // Around 56 kB (gzipped)
|       |-- amazeui.widgets.helper.js
|       |-- amazeui.widgets.helper.min.js
|       |-- app.js
|       `-- handlebars.min.js
|-- blog.html
|-- index.html
|-- landing.html
|-- login.html
|-- sidebar.html
`-- widget.html
```

## Build your first Page with Amaze UI

1. Open your favorite editor and create a new HTML file.
2. Copy the following codes into HTML file.
3. Check CSS and JS components, and add all components you want to `<body>`.

```html
<!doctype html>
<html class="no-js">
<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="description" content="">
  <meta name="keywords" content="">
  <meta name="viewport"
        content="width=device-width, initial-scale=1">
  <title>Hello Amaze UI</title>

  <!-- Set render engine for 360 browser -->
  <meta name="renderer" content="webkit">

  <!-- No Baidu Siteapp-->
  <meta http-equiv="Cache-Control" content="no-siteapp"/>

  <link rel="icon" type="image/png" href="assets/i/favicon.png">

  <!-- Add to homescreen for Chrome on Android -->
  <meta name="mobile-web-app-capable" content="yes">
  <link rel="icon" sizes="192x192" href="assets/i/app-icon72x72@2x.png">

  <!-- Add to homescreen for Safari on iOS -->
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="black">
  <meta name="apple-mobile-web-app-title" content="Amaze UI"/>
  <link rel="apple-touch-icon-precomposed" href="assets/i/app-icon72x72@2x.png">

  <!-- Tile icon for Win8 (144x144 + tile color) -->
  <meta name="msapplication-TileImage" content="assets/i/app-icon72x72@2x.png">
  <meta name="msapplication-TileColor" content="#0e90d2">

  <link rel="stylesheet" href="assets/css/amazeui.min.css">
  <link rel="stylesheet" href="assets/css/app.css">
</head>
<body>
<p>
  Hello Amaze UI.
</p>

<!--Add Your Codes Here-->

<!--[if (gte IE 9)|!(IE)]><!-->
<script src="assets/js/jquery.min.js"></script>
<!--<![endif]-->
<!--[if lte IE 8 ]>
<script src="http://libs.baidu.com/jquery/1.11.1/jquery.min.js"></script>
<![endif]-->
</body>
</html>
```

## Join the Discussion

If you have any problem about using Amaze UI, please leave us a message in comments. Any comments or suggestions will be appreciated.

__Thank you so much for your attention and support to Amaze UI!__
