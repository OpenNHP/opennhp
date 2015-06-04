<h1><a href="http://amazeui.org/" title="Amaze UI 官网"><img style="float: left" width="240" src="https://raw.githubusercontent.com/allmobilize/amazeui/master/vendor/amazeui/amazeui-b.png" alt="Amaze UI Logo"/></a></h1>

[![Bower version](https://img.shields.io/bower/v/amazeui.svg?style=flat-square)](https://github.com/allmobilize/amazeui)
[![NPM version](https://img.shields.io/npm/v/amazeui.svg?style=flat-square)](https://www.npmjs.com/package/amazeui)
[![Build Status](https://img.shields.io/travis/allmobilize/amazeui.svg?style=flat-square)](https://travis-ci.org/allmobilize/amazeui)
[![Dependency Status](https://img.shields.io/david/allmobilize/amazeui.svg?style=flat-square)](https://david-dm.org/allmobilize/amazeui)
[![devDependency Status](https://img.shields.io/david/dev/allmobilize/amazeui.svg?style=flat-square)](https://david-dm.org/allmobilize/amazeui#info=devDependencies)

Amaze UI 是基于社区开源项目构建的一个跨屏前端框架。 __[README in English](https://github.com/allmobilize/amazeui/blob/master/README_EN.md)__

### [React 版](https://github.com/amazeui/amazeui-react) | [独立插件](https://github.com/amazeui)

## 功能简介

### 移动优先

以移动优先（Mobile first）为理念，从小屏逐步扩展到大屏，最终实现所有屏幕适配，适应移动互联潮流。

### 组件丰富，模块化

Amaze UI 含近 20 个 CSS 组件、10 个 JS 组件，更有 17 款包含近 60 个主题的 Web 组件，可快速构建界面出色、体验优秀的跨屏页面，大幅提升开发效率。

### 本地化支持

相比国外框架，Amaze UI 关注中文排版，根据用户代理调整字体，实现更好的中文排版效果；兼顾国内主流浏览器及 App 内置浏览器兼容支持。

### 轻量级，高性能

Amaze UI 面向 HTML5 开发，使用 CSS3 来做动画交互，平滑、高效，更适合移动设备，让 Web 应用更快速载入。

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
|-- dist        # 构建目录
|-- docs        # 文档
|-- fonts       # Icon font，目前使用了 http://staticfile.org/
|-- gulpfile.js # 构建配置文件
|-- js          # JS 文件
|-- less        # LESS 文件
|-- tools       # 相关工具
|-- vendor
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

欢迎给 Amaze UI [提交 Bug](https://github.com/allmobilize/amazeui/issues/new?title=Bug%3A%20&body=**%E9%97%AE%E9%A2%98%E6%8F%8F%E8%BF%B0**%0A%0A%EF%BC%88%E6%8F%8F%E8%BF%B0%E4%B8%80%E4%B8%8B%E9%97%AE%E9%A2%98%EF%BC%89%0A%0A**%E4%BA%A7%E7%94%9F%E7%8E%AF%E5%A2%83**%0A%0A-%20%E8%AE%BE%E5%A4%87%EF%BC%9A%EF%BC%88%E6%89%8B%E6%9C%BA%E3%80%81%E5%B9%B3%E6%9D%BF%E7%AD%89%E7%A7%BB%E5%8A%A8%E8%AE%BE%E5%A4%87%E6%97%B6%E5%A1%AB%E5%86%99%E6%AD%A4%E9%A1%B9%EF%BC%89%0A-%20%E6%93%8D%E4%BD%9C%E7%B3%BB%E7%BB%9F%E5%8F%8A%E7%89%88%E6%9C%AC%EF%BC%9A%0A-%20%E6%B5%8F%E8%A7%88%E5%99%A8%E5%8F%8A%E7%89%88%E6%9C%AC%EF%BC%9A%0A-%20%E6%BC%94%E7%A4%BA%E5%9C%B0%E5%9D%80%EF%BC%9A%0A%0A**%E5%A4%8D%E7%8E%B0%E6%AD%A5%E5%A5%8F**%0A%0A1.%20%0A2.%20%0A...)。

为了能最准确的传达所描述的问题， 建议你在反馈时附上演示，方便我们理解。

为了方便演示，我们使用 **[JSBin 搭建了在线一个调试工具](http://bin.amazeui.org)**，你可以把有问题的场景粘在里面，反馈给我们。

### 需求提交

用户可以通过 Issue 系统或者官网留言提交需求，符合 Amaze UI 理念的需求我们都会考虑。

## 贡献代码

欢迎大家加入捉虫队伍，同时大家也可以分享自己的开发的 Web 组件。Fork 本项目，然后提交 Pull Request 即可。

如果你没有相关经验，建议先阅读下面的文章：

- [Contributing to Open Source on GitHub](https://guides.github.com/activities/contributing-to-open-source/)
- [The Beginner’s Guide to Contributing to Open Source Projects](http://blog.newrelic.com/2014/05/05/open-source_gettingstarted/)
- [How to Start Contributing to Open Source](http://www.developer.com/open/how-to-start-contributing-to-open-source.html)

### 我们希望用户参与的项目

- 把使用 Amaze UI 的项目提交给我们，能获得一定的展示量，同时能给其他用户提供参考；
- SCSS 等其他 CSS 扩展语言的支持：通过编写自动转换脚本实现，人工迁移同步更新成本太高；
- 使用 Amaze UI 开发制作模板；
- Angular.js、Ember.js、Meteor 等框架的支持；
- 基于 Amaze UI 样式开发 JavaScript 插件，可以把 Bootstrap 的插件移植到 Amaze UI；
- 兼容性测试，我们的测试只能覆盖有限，需要更多用户参与，包括浏览器、WebView 等不同的用户代理；
- ...

### 开发文档

开发文档存放在 `docs/styleguide` 目录下，也可以通过 [Amaze UI 官网](http://amazeui.org/)查看：

- [Amaze UI HTML/CSS 编写规范](http://amazeui.org/getting-started/html-css)
- [Amaze UI JavaScript 编写规范](http://amazeui.org/getting-started/javascript)
- [Amaze UI Web 组件开发规范](http://amazeui.org/getting-started/widget)

## 参考、使用的项目

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
* [Hammer.js](https://github.com/hammerjs/hammer.js) ([MIT License](https://github.com/hammerjs/hammer.js/blob/master/LICENSE.md))
* [Flat UI](https://github.com/designmodo/Flat-UI) ([CC BY 3.0 and MIT License](https://github.com/designmodo/Flat-UI#copyright-and-license))
* [store.js](https://github.com/marcuswestin/store.js) ([MIT License](https://github.com/marcuswestin/store.js/blob/master/LICENSE))
* [bootstrap-datepicker.js](http://www.eyecon.ro/bootstrap-datepicker/) ([Apache License 2.0](http://www.eyecon.ro/bootstrap-datepicker/js/bootstrap-datepicker.js))
* [iScroll](http://iscrolljs.com/) ([MIT License](http://iscrolljs.com/#license))

可能会有部分项目遗漏，我们会不断整理更新。

### Developed with Open Source Licensed [WebStorm](http://www.jetbrains.com/webstorm/)

<a href="http://www.jetbrains.com/webstorm/" target="_blank">
<img src="http://ww1.sinaimg.cn/large/005yyi5Jjw1elpp6svs2eg30k004i3ye.gif" width="240" />
</a>
