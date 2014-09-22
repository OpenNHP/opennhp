<h1><a href="http://amazeui.org/" title="Amaze UI 官网"><img style="float: left" width="240" src="https://raw.githubusercontent.com/allmobilize/amazeui/master/vendor/amazeui/amazeui-b.png" alt="Amaze UI Logo"/></a></h1>   


Amaze UI 是基于社区开源项目构建的一个跨屏前端框架。 __[README in English](https://github.com/allmobilize/amazeui/blob/master/README_EN.md)__

## 功能简介

### 为移动而生

Amaze UI 遵循 Mobile first 理念，从小屏逐步扩展到大屏，最终实现跨屏适配，适应移动互联潮流。

### 组件丰富，模块化

Amaze UI 含近 20 个 CSS 组件、10 个 JS 组件，更有 17 款包含近 60 个主题的 Web 组件，可快速构建界面出色、体验优秀的跨屏页面，大幅度提升你的开发效率。

### 本地化支持

相比国外的前端框架，Amaze UI 专注解决中文排版优化问题，根据操作系统调整字体，实现最佳中文排版效果；针对国内主流浏览器及 App 内置浏览器提供更好的兼容性支持，为你节省大量兼容性调试时间。

### 轻量级，高性能

Amaze UI 非常注重性能，基于轻量的 Zepto.js 开发，并使用 CSS3 来做动画交互，平滑、高效，更适合移动设备，让你的 Web 应用可以高速载入。

## 下载及使用

用户可以从 [Amaze UI 官网](http://amazeui.org/getting-started) 下载打包好的模板。

所有文档存放在 `docs/` 目录下，为方便查看演示效果，建议通过 [Amaze UI 官网](http://amazeui.org/)查看文档。


## 开发及构建

用户可以在 Amaze UI 的基础上进行二次开发。

### 目录结构

```
amazeui
|-- HISTORY.md
|-- LICENSE
|-- README.md
|-- package.json
|-- dist        # 部署目录
|-- docs        # 文档
|-- fonts       # Icon font，目前使用了 http://staticfile.org/
|-- gulpfile.js # 构建配置文件
|-- js          # JS 文件
|-- less        # LESS 文件
|-- tools       # 相关工具
|-- vendor      # 第三方文件
|   |-- amazeui.partials.js
|   |-- handlebars
|   |-- json.format.js
|   |-- seajs
|   `-- zepto
`-- widget      # Web 组件
```

### 构建工具

Amaze UI 使用 [gulp.js](http://gulpjs.com/) 构建项目。

首先全局安装 Gulp：

```
npm install -g gulp
```

Clone 项目文件:

```
git clone https://github.com/allmobilize/amazeui.git
```

然后进入目录安装依赖：

```
npm install
```

接下来，执行 `gulp`：

```
gulp
```

## Bug 反馈及需求提交

### Bug 反馈

欢迎给 Amaze UI [提交 Bug](https://github.com/allmobilize/amazeui/issues)。

为了能最准确的传达所描述的问题， 建议你在反馈时附上演示，方便我们理解。

下面的几个链接是我们在几个在线调试工具上建的页面， 已经引入了 Amaze UI 样式和脚本，你可以选择你喜欢的工具【Fork】一份， 把要有问题的场景粘在里面，反馈给我们。

- [JSBin](http://jsbin.com/kijiqu/1/edit?html,output)
- [JSFiddle](http://jsfiddle.net/hegfirose/W22fV/)
- [CodePen](http://codepen.io/minwe/pen/AEeup)

### 需求提交

用户可以通过 Issue 系统或者官网留言提交需求，符合 Amaze UI 理念的需求我们都会考虑。


## 贡献代码

欢迎大家加入捉虫队伍，同时大家也可以分享自己的开发的 Web 组件。Fork 本项目，然后提交 Pull Request 即可。

如果你没有相关经验，建议先阅读下面的文章：

- [Contributing to Open Source on GitHub](https://guides.github.com/activities/contributing-to-open-source/)
- [The Beginner’s Guide to Contributing to Open Source Projects](http://blog.newrelic.com/2014/05/05/open-source_gettingstarted/)
- [How to Start Contributing to Open Source](http://www.developer.com/open/how-to-start-contributing-to-open-source.html)

__开发文档__

开发文档存放在 `docs/rules` 目录下，也可以通过 [Amaze UI 官网](http://amazeui.org/)查看：

- [Amaze UI HTML/CSS 编写规范](http://amazeui.org/getting-started/html-css)
- [Amaze UI JavaScript 编写规范](http://amazeui.org/getting-started/javascript)
- [Amaze UI Web 组件开发规范](http://amazeui.org/getting-started/widget)

## 参考、使用的开源项目

* [Zepto.js](https://github.com/madrobby/zepto) ([MIT
License](https://github.com/madrobby/zepto/blob/master/MIT-LICENSE))
* [Sea.js](https://github.com/seajs/seajs) ([MIT License](https://github.com/seajs/seajs/blob/master/LICENSE.md))
* [Handlebars.js](https://github.com/wycats/handlebars.js) ([MIT
License](https://github.com/wycats/handlebars.js/blob/master/LICENSE))
* [normalize.css](https://github.com/necolas/normalize.css) ([MIT
License](https://github.com/necolas/normalize.css/blob/master/LICENSE.md))
* [FontAwesome](https://github.com/FortAwesome/Font-Awesome/) ([CC BY 3.0 License](http://creativecommons.org/licenses/by/3.0/))
* [Bootstrap](https://github.com/twbs/bootstrap) ([MIT License](https://github.com/twbs/bootstrap/blob/master/LICENSE))
* [UIkit](https://github.com/uikit/uikit) ([MIT License](https://github.com/uikit/uikit/blob/master/LICENSE.md))
* [Foundation](https://github.com/zurb/foundation) ([MIT
License](https://github.com/zurb/foundation/blob/master/LICENSE))
* [Framework7](https://github.com/nolimits4web/Framework7) ([MIT
License](https://github.com/nolimits4web/Framework7/blob/master/LICENSE))
* [Alice](https://github.com/aliceui/aliceui.org/) ([MIT
License](https://github.com/aliceui/aliceui.org/blob/master/LICENSE))
* [Arale](https://github.com/aralejs/aralejs.org/) ([MIT
License](https://github.com/aralejs/aralejs.org/blob/master/LICENSE))
* [Pure](https://github.com/yui/pure) ([BSD License](https://github.com/yui/pure/blob/master/LICENSE.md))
* [Semantic UI](https://github.com/Semantic-Org/Semantic-UI) ([MIT
License](https://github.com/Semantic-Org/Semantic-UI/blob/master/LICENSE.md))
* [FastClick](https://github.com/ftlabs/fastclick) ([MIT
License](https://github.com/ftlabs/fastclick/blob/master/LICENSE))
* [screenfull.js](https://github.com/sindresorhus/screenfull.js) ([MIT
License](https://github.com/sindresorhus/screenfull.js/blob/gh-pages/license))
* [FlexSlider](https://github.com/woothemes/FlexSlider) ([GPL 2.0](http://www.gnu.org/licenses/gpl-2.0.html))

可能会有部分项目遗漏，我们会不断整理更新。
