# 开始使用 Amaze UI
---

Amaze UI 是一个轻量级、 [**Mobile first**](http://cbrac.co/113eY5h) 的前端框架，
基于开源社区流行前端框架编写（[使用、参考的项目列表](https://github.com/allmobilize/amazeui#%E5%8F%82%E8%80%83%E4%BD%BF%E7%94%A8%E7%9A%84%E5%BC%80%E6%BA%90%E9%A1%B9%E7%9B%AE)）。


## 下载文件

<div class="am-g">
  <div class="am-u-md-8 am-u-md-centered">
    <a href="/download?ver=v1.0.0" class="am-btn am-btn-block am-btn-success am-btn-lg" onclick="window.ga && ga('send', 'pageview', '/download/AmazeUI.zip');
">下载 Amaze UI v1.0.0</a>
  </div>
</div>

### 使用 CDN

#### Staticfile.org

下面的链接由 [Staticfile.org](http://staticfile.org) 提供 CDN 服务。

```html
http://cdn.staticfile.org/amazeui/1.0.0-rc1/css/amazeui.basic.css
http://cdn.staticfile.org/amazeui/1.0.0-rc1/css/amazeui.basic.min.css
http://cdn.staticfile.org/amazeui/1.0.0-rc1/css/amazeui.css
http://cdn.staticfile.org/amazeui/1.0.0-rc1/css/amazeui.min.css
http://cdn.staticfile.org/amazeui/1.0.0-rc1/css/amazeui.widgets.css
http://cdn.staticfile.org/amazeui/1.0.0-rc1/css/amazeui.widgets.min.css
http://cdn.staticfile.org/amazeui/1.0.0-rc1/js/amazeui.basic.js
http://cdn.staticfile.org/amazeui/1.0.0-rc1/js/amazeui.basic.min.js
http://cdn.staticfile.org/amazeui/1.0.0-rc1/js/amazeui.js
http://cdn.staticfile.org/amazeui/1.0.0-rc1/js/amazeui.min.js
http://cdn.staticfile.org/amazeui/1.0.0-rc1/js/amazeui.widgets.helper.js
http://cdn.staticfile.org/amazeui/1.0.0-rc1/js/amazeui.widgets.helper.min.js
http://cdn.staticfile.org/amazeui/1.0.0-rc1/js/amazeui.widgets.js
http://cdn.staticfile.org/amazeui/1.0.0-rc1/js/amazeui.widgets.min.js
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

- `amazeui.css` / `amazeui.js`：包含 Amaze UI 所有的样式和脚本；
- `amazeui.basic.css` / `amazeui.basic.js`：包含 Amaze UI CSS 部分、JS 插件部分的样式和脚本，供不使用 Web 组件的用户选择；
- `amazeui.widgets.css` / `amazeui.widgets.js`：包含 Web 组件及其依赖的基础插件的样式和脚本，供只使用 Web 组件的用户选择。

以上每个文件都有对应的 minified 文件。

```
{basic} = [CSS] + [JS插件]
```
```
{widgets} = [Web组件] + [Web 组件依赖的 CSS] + [Web组件依赖的JS插件]
```

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
|   |   |-- amazeui.basic.css       // Amaze UI CSS/JS插件样式
|   |   |-- amazeui.basic.min.css     // 29.9 kB gzipped
|   |   |-- amazeui.css             // Amaze UI 所有样式文件
|   |   |-- amazeui.min.css           // 38.9 kB gzipped
|   |   |-- amazeui.widgets.css     // Web 组件主题及依赖的样式
|   |   |-- amazeui.widgets.min.css   // 30.4 kB gzipped
|   |   `-- app.css
|   |-- i
|   |   |-- app-icon72x72@2x.png
|   |   |-- favicon.png
|   |   `-- startup-640x1096.png
|   `-- js
|       |-- amazeui.basic.js
|       |-- amazeui.basic.min.js          // 49 kB gzipped
|       |-- amazeui.js
|       |-- amazeui.min.js                // 52.4 kb gzipped
|       |-- amazeui.widgets.helper.js
|       |-- amazeui.widgets.helper.min.js
|       |-- amazeui.widgets.js
|       |-- amazeui.widgets.min.js        // 47 kB gzipped
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
