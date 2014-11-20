# 开始使用 Amaze UI
---

Amaze UI 是一个轻量级、 [**Mobile first**](http://cbrac.co/113eY5h) 的前端框架，
基于开源社区流行前端框架编写（[使用、参考的项目列表](https://github.com/allmobilize/amazeui#%E5%8F%82%E8%80%83%E4%BD%BF%E7%94%A8%E7%9A%84%E5%BC%80%E6%BA%90%E9%A1%B9%E7%9B%AE)）。


## 下载文件

<div class="am-g">
  <div class="am-u-md-8 am-u-md-centered">
    <a href="/download?ver=v2.0.0-dp1" class="am-btn am-btn-block am-btn-success am-btn-lg" onclick="window.ga && ga('send', 'pageview', '/download/AmazeUI.zip');
"><i class="am-icon-download"></i> Amaze UI v2.0.0-dp1</a>
  </div>
</div>

### 使用 CDN

Amaze UI CDN： 域名解析服务由 DNSPod 提供，CDN 存储由又拍云提供。

```html
http://cdn.amazeui.org/amazeui/2.0.0-dp1/css/amazeui.css
http://cdn.amazeui.org/amazeui/2.0.0-dp1/css/amazeui.min.css
http://cdn.amazeui.org/amazeui/2.0.0-dp1/js/amazeui.js
http://cdn.amazeui.org/amazeui/2.0.0-dp1/js/amazeui.min.js
http://cdn.amazeui.org/amazeui/2.0.0-dp1/js/amazeui.amd.js
http://cdn.amazeui.org/amazeui/2.0.0-dp1/js/amazeui.amd.min.js
http://cdn.amazeui.org/amazeui/2.0.0-dp1/js/amazeui.widgets.helper.js
http://cdn.amazeui.org/amazeui/2.0.0-dp1/js/amazeui.widgets.helper.min.js
```

### 使用 Bower

```html
bower install amazeui
```

### 获取源码

你可以从 GitHub 项目主页获取源代码。

<iframe src="http://ghbtns.com/github-btn.html?user=allmobilize&repo=amazeui&type=watch&count=true&size=large" allowtransparency="true" frameborder="0" scrolling="0" width="156px" height="30px"></iframe>

<iframe src="http://ghbtns.com/github-btn.html?user=allmobilize&repo=amazeui&type=fork&count=true&size=large" allowtransparency="true" frameborder="0" scrolling="0" width="156px" height="30px"></iframe>

## 版本说明

Amaze UI 遵循 [Semantic Versioning](http://semver.org/lang/zh-CN/) 规范，版本格式采用 `主版本号.次版本号.修订号` 的形式，版本号递增规则如下：

- 主版本号：做了不兼容的API 修改，如整体风格变化、大规模重构等；
- 次版本号：做了向下兼容的功能性新增；
- 修订号：做了向下兼容的问题修正、细节调整等。


## 目录结构

### 文件说明

- `amazeui.css` / `amazeui.js`：包含 Amaze UI 所有的样式和脚本。
- `amazeui.amd.js`：使用 AMD 规范打包的脚本。
- `amazeui.widgets.helper.js`：Amaze UI 必须的 Handlebars helper 以及 Web 组件模板 partials，**仅供使用 Handlebars 的用户使用。

以上每个文件都有对应的 minified 文件。

### 示例 HTML

- `index.html` - 空白 HTML 模板；
- `blog.html` - 博客页面模板（[预览](/examples/blog.html)）；
- `landing.html` - Landing Page 模板（[预览](/examples/landing.html)）；
- `login.html` - 登录界面模板（[预览](/examples/login.html)）；
- `sidebar.html` - 带边栏的文章模板（[预览](/examples/sidebar.html)）；
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
|       |-- amazeui.amd.js
|       |-- amazeui.amd.min.js        // 48 kB (gzipped)
|       |-- app.js
|       |-- handlebars.min.js
|       `-- zepto.min.js
|-- blog.html
|-- index.html
|-- landing.html
|-- login.html
|-- sidebar.html
`-- widget.html
```

## 参与讨论

有任何使用问题，请在评论中留言，也欢迎大家发表意见、建议。

__感谢大家对 Amaze UI 的关注和支持！__
