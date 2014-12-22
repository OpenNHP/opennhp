# 开始使用 Amaze UI
---

Amaze UI 是一个轻量级（所有 CSS 和 JS gzip 后 90 kB 左右）、 [**Mobile first**](http://cbrac.co/113eY5h) 的前端框架，
基于开源社区流行前端框架编写（[使用、参考的项目列表](https://github.com/allmobilize/amazeui#%E5%8F%82%E8%80%83%E4%BD%BF%E7%94%A8%E7%9A%84%E5%BC%80%E6%BA%90%E9%A1%B9%E7%9B%AE)）。

## 获取 Amaze UI

### 下载文件

<div class="am-g">
  <div class="am-u-md-8 am-u-md-centered">
    <a id="doc-dl-btn" href="http://amazeui.org/download?ver=2.0.0" class="am-btn am-btn-block am-btn-success am-btn-lg" onclick="window.ga && ga('send', 'pageview', '/download/AmazeUI.zip');
"><i class="am-icon-download"></i> Amaze UI v2.0.0 正式版</a>
  </div>
</div>


### 使用 CDN

Amaze UI CDN： 域名解析服务由 DNSPod 提供，CDN 存储由又拍云提供。

```html
http://cdn.amazeui.org/amazeui/2.0.0/css/amazeui.css
http://cdn.amazeui.org/amazeui/2.0.0/css/amazeui.min.css
http://cdn.amazeui.org/amazeui/2.0.0/js/amazeui.js
http://cdn.amazeui.org/amazeui/2.0.0/js/amazeui.min.js
http://cdn.amazeui.org/amazeui/2.0.0/js/amazeui.legacy.js
http://cdn.amazeui.org/amazeui/2.0.0/js/amazeui.legacy.min.js
http://cdn.amazeui.org/amazeui/2.0.0/js/amazeui.widgets.helper.js
http://cdn.amazeui.org/amazeui/2.0.0/js/amazeui.widgets.helper.min.js
```

### 使用 Bower

```html
bower install amazeui
```

### 获取源码

你可以从 GitHub 项目主页获取源代码。

<iframe src="http://ghbtns.com/github-btn.html?user=allmobilize&repo=amazeui&type=watch&count=true&size=large" allowtransparency="true" frameborder="0" scrolling="0" width="156px" height="30px"></iframe>

<iframe src="http://ghbtns.com/github-btn.html?user=allmobilize&repo=amazeui&type=fork&count=true&size=large" allowtransparency="true" frameborder="0" scrolling="0" width="156px" height="30px"></iframe>

## 文件及版本说明

### 文件说明

- `amazeui.css` / `amazeui.js`：包含 Amaze UI 所有的 CSS、JS。
- `amazeui.flat.css`：圆角版 Amaze UI CSS，演示参见 `1.x`。
- `amazeui.legacy.js`：为 IE 8 打包的 JS。
- `amazeui.widgets.helper.js`： Web 组件必须的 Handlebars helper 及 Web 组件模板 partials，**仅供 Handlebars 用户使用**。

以上每个文件都有对应的 minified 文件。

### 版本号说明

Amaze UI 遵循 [Semantic Versioning](http://semver.org/lang/zh-CN/) 规范，版本格式采用 `主版本号.次版本号.修订号` 的形式，版本号递增规则如下：

- 主版本号：做了不兼容的API 修改，如整体风格变化、大规模重构等；
- 次版本号：做了向下兼容的功能性新增；
- 修订号：做了向下兼容的问题修正、细节调整等。

- [**1.x 到 2.x 变更记录暨升级指南**](https://github.com/allmobilize/amazeui/wiki/Migration-form-1.x-to-2.x)


## 下载包目录结构

下载包种包含 Amaze UI 的 CSS、JS 文件，以及示例文件：

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
|   |   |-- amazeui.min.css           // 约 40 kB (gzipped)
|   |   `-- app.css
|   |-- i
|   |   |-- app-icon72x72@2x.png
|   |   |-- favicon.png
|   |   `-- startup-640x1096.png
|   `-- js
|       |-- amazeui.js
|       |-- amazeui.min.js                // 约 48 kB (gzipped)
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
        content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
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
<script src="http://libs.baidu.com/jquery/1.11.1/jquery.min.js"></script>
<![endif]-->
</body>
</html>
```

## 参与讨论

有任何使用问题，请在评论中留言，也欢迎大家发表意见、建议。

__感谢大家对 Amaze UI 的关注和支持！__

<script>
  function pyRegisterCvt(){
    var w=window,d=document,e=encodeURIComponent;
    var b=location.href,c=d.referrer,f,g=d.cookie,h=g.match(/(^|;)\s*ipycookie=([^;]*)/),i=g.match(/(^|;)\s*ipysession=([^;]*)/);
    if (w.parent!=w){f=b;b=c;c=f;};u='//stats.ipinyou.com/cvt?a='+e('9J.H5.EtdEOxc6fV-EiorJDjFOTP')+'&c='+e(h?h[2]:'')+'&s='+e(i?i[2].match(/jump\%3D(\d+)/)[1]:'')+'&u='+e(b)+'&r='+e(c)+'&rd='+(new Date()).getTime()+'&e=';
    (new Image()).src=u;
  }

  $(function() {
    $('#doc-dl-btn').on('click', pyRegisterCvt);
  })
</script>
