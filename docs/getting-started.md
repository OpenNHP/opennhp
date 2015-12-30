---
id: getting-started
title: 开始使用
titleEn: Getting Started
permalink: getting-started.html
next: getting-started/layouts.html
---

# 开始使用 Amaze UI
---

Amaze UI 是一个轻量级（所有 CSS 和 JS gzip 后 100 kB 左右）、 [**Mobile first**](http://cbrac.co/113eY5h) 的前端框架，
基于开源社区流行前端框架编写（[使用、参考的项目列表](https://github.com/amazeui/amazeui#参考使用的项目)）。

## 获取 Amaze UI

### 下载文件

<div class="am-g">
  <div class="am-u-md-8 am-u-md-centered">
    <a id="doc-dl-btn" href="http://amazeui.org/download?ver=__VERSION__" class="am-btn am-btn-block am-btn-success am-btn-lg" onclick="window.ga && ga('send', 'pageview', '/download/AmazeUI.zip');
"><i class="am-icon-download"></i> Amaze UI v__VERSION__</a>
  </div>
</div>

- [**更新日志**](https://github.com/amazeui/amazeui/blob/master/CHANGELOG.md)

**离线文档：**

<div class="am-g">
  <div class="am-u-sm-6"><a href="http://amazeui.org/download?ver=docs" class="am-btn am-btn-block am-btn-primary">HTML 版离线文档</a></div>
  <div class="am-u-sm-6"><a href="http://amazeui.org/download?ver=dash" class="am-btn am-btn-block am-btn-warning">Dash Docsets</a></div>
</div>

项目地址：

- [Amaze UI Docs](https://github.com/amazeui/docs)
- [Amaze UI Dash Docsets Generator](https://github.com/amazeui/docs-generator)

**代码片段：**

<div class="am-g">
  <div class="am-u-sm-6"><a href="http://amazeui.org/download?ver=jetbrains" class="am-btn am-btn-block am-btn-secondary">JetBrains 系列编辑器</a></div>
  <div class="am-u-sm-6"><a href="http://amazeui.org/download?ver=sublime" class="am-btn am-btn-block am-btn-danger">Sublime</a></div>
</div>

详见 [Amaze UI Snippets](https://github.com/amazeui/snippets)。

**Starter Kit：**

Gulp、NPM 构建的前端开发工作流，点击[访问项目主页](https://github.com/amazeui/starter-kit)。

### 使用 CDN

#### 官方 CDN

域名解析服务由 DNSPod 提供，CDN 存储由七牛提供。

```html
http://cdn.amazeui.org/amazeui/__VERSION__/css/amazeui.css
http://cdn.amazeui.org/amazeui/__VERSION__/css/amazeui.min.css
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.min.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.ie8polyfill.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.ie8polyfill.min.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.widgets.helper.js
http://cdn.amazeui.org/amazeui/__VERSION__/js/amazeui.widgets.helper.min.js
```

#### cdnjs

面向国外的用户可以使用 cdnjs 提供的 CDN 服务（支持 HTTPS）。

- https://cdnjs.com/libraries/amazeui

### 使用 Bower

```html
bower install amazeui
```

### 移植的插件（使用示例）

- [DateTimePicker - 日期时间选择](https://github.com/amazeui/datetimepicker)
- [Echo.js - 图片懒加载](https://github.com/amazeui/echo)
- [Lazyload - 图片懒加载](https://github.com/amazeui/lazyload)
- [Chosen - 下拉选框增强](https://github.com/amazeui/chosen)
- [Masonry - 瀑布流](https://github.com/amazeui/masonry)
- [Switch - 开关切换插件](https://github.com/amazeui/switch)
- [Tags Input - 标签输入框](https://github.com/amazeui/tagsinput)
- [Video.js Amaze UI 皮肤](https://github.com/amazeui/videojs)
- [jQuery DataTables - 表格分页、排序等](https://github.com/amazeui/datatables)
- [Tree - 树形菜单插件](https://github.com/amazeui/tree)
- [Swiper - 图片轮播插件](https://github.com/amazeui/swiper)
- [Slick - 图片轮播插件](https://github.com/amazeui/slick)

### 获取源码

你可以从 GitHub 项目主页获取源代码。

<iframe src="http://ghbtns.com/github-btn.html?user=amazeui&repo=amazeui&type=watch&count=true&size=large" allowtransparency="true" frameborder="0" scrolling="0" width="156px" height="30px"></iframe>

<iframe src="http://ghbtns.com/github-btn.html?user=amazeui&repo=amazeui&type=fork&count=true&size=large" allowtransparency="true" frameborder="0" scrolling="0" width="156px" height="30px"></iframe>

## 文件及版本说明

### 文件说明

- `amazeui.css` / `amazeui.js`：包含 Amaze UI 所有的 CSS、JS。
- `amazeui.flat.css`：圆角版 Amaze UI CSS，演示参见 `1.x`。
- `amazeui.ie8polyfill.js`：IE8 polyfill。
- `amazeui.widgets.helper.js`： **供使用 Handlebars 的用户使用，其他用户请忽略**，内含 Web 组件必须的 Handlebars helper 及 Web 组件模板 partials。

以上每个文件都有对应的 minified 文件。

### 版本号说明

Amaze UI 遵循 [Semantic Versioning](http://semver.org/lang/zh-CN/) 规范，版本格式采用 `主版本号.次版本号.修订号` 的形式，版本号递增规则如下：

- 主版本号：做了不兼容的API 修改，如整体风格变化、大规模重构等；
- 次版本号：做了向下兼容的功能性新增；
- 修订号：做了向下兼容的问题修正、细节调整等。

- [**1.x 到 2.x 变更记录暨升级指南**](https://github.com/amazeui/amazeui/wiki/Migration-form-1.x-to-2.x)


## 下载包目录结构

下载包中包含 Amaze UI 的 CSS、JS 文件，以及示例文件：

- `index.html` - 空白 HTML 模板；
- `blog.html` - 博客页面模板（[预览](/examples/blog.html)）；
- `landing.html` - Landing Page 模板（[预览](/examples/landing.html)）；
- `login.html` - 登录界面模板（[预览](/examples/login.html)）；
- `sidebar.html` - 带边栏的文章模板（[预览](/examples/sidebar.html)）；
- `admin-*.html` - 简单的管理后台界面（[预览](/examples/admin-index.html)）
- 在 `app.css` 中编写 CSS；
- 在 `app.js` 中编写 JavaScript；
- 图片资源可以放在 `i` 目录下。

```
AmazeUI
|-- assets
|   |-- css
|   |   |-- amazeui.css             // Amaze UI 所有样式文件
|   |   |-- amazeui.min.css           // 约 42 kB (gzipped)
|   |   `-- app.css
|   |-- i
|   |   |-- app-icon72x72@2x.png
|   |   |-- favicon.png
|   |   `-- startup-640x1096.png
|   `-- js
|       |-- amazeui.js
|       |-- amazeui.min.js                // 约 56 kB (gzipped)
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

## 创建一个页面

1. 新建一个 HTML 文档，将下面的代码粘贴到文档中；
2. 查看 CSS 组件及 JS 插件，拷贝符合的演示代码，粘贴到 `<body>` 区域，并按需调整；
3. 一个简单的页面完成。

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

<!--在这里编写你的代码-->

<!--[if (gte IE 9)|!(IE)]><!-->
<script src="assets/js/jquery.min.js"></script>
<!--<![endif]-->
<!--[if lte IE 8 ]>
<script src="http://libs.baidu.com/jquery/1.11.3/jquery.min.js"></script>
<script src="http://cdn.staticfile.org/modernizr/2.8.3/modernizr.js"></script>
<script src="assets/js/amazeui.ie8polyfill.min.js"></script>
<![endif]-->
<script src="assets/js/amazeui.min.js"></script>
</body>
</html>
```

## 参与讨论

有任何使用问题，请在评论中留言，也欢迎大家发表意见、建议。

__感谢大家对 Amaze UI 的关注和支持！__
