<h1><a href="http://amazeui.org/" title="Amaze UI 官网"><img style="float: left" width="240" src="https://raw.githubusercontent.com/allmobilize/amazeui/master/vendor/amazeui/amazeui-b.png" alt="Amaze UI Logo"/></a></h1>

### My 2 cents. Yahia

[![Bower version](https://img.shields.io/bower/v/amazeui.svg?style=flat-square)](https://github.com/amazeui/amazeui)
[![NPM version](https://img.shields.io/npm/v/amazeui.svg?style=flat-square)](https://www.npmjs.com/package/amazeui)
[![Build Status](https://img.shields.io/travis/amazeui/amazeui.svg?style=flat-square)](https://travis-ci.org/amazeui/amazeui)
[![Dependency Status](https://img.shields.io/david/amazeui/amazeui.svg?style=flat-square)](https://david-dm.org/amazeui/amazeui)
[![devDependency Status](https://img.shields.io/david/dev/amazeui/amazeui.svg?style=flat-square)](https://david-dm.org/amazeui/amazeui#info=devDependencies)

Amaze UI 是基于社区开源项目构建的一个跨屏前端框架。

### [Docs in English](http://amazeui.github.io/docs/en/)

**[README in English](README.en.md)**

### [React 版](https://github.com/amazeui/amazeui-react) | [独立插件](https://github.com/amazeui) | [Amaze UI Touch](https://github.com/amazeui/amazeui-touch)

## 功能简介

### 移动优先

以移动优先（Mobile first）为理念，从小屏逐步扩展到大屏，最终实现所有屏幕适配，适应移动互联潮流。

### 组件丰富，模块化

Amaze UI 含近 20 个 CSS 组件、20 余 JS 组件，更有多个包含不同主题的 Web 组件，可快速构建界面出色、体验优秀的跨屏页面，大幅提升开发效率。

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

首先全局安装 gulp：

```
npm install -g gulp
```

克隆项目文件:

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

## [Bug 反馈及需求提交](CONTRIBUTING.md)

## 参考、使用的项目

- [Zepto.js](https://github.com/madrobby/zepto) ([MIT
  License](https://github.com/madrobby/zepto/blob/master/MIT-LICENSE))
- [Sea.js](https://github.com/seajs/seajs) ([MIT License](https://github.com/seajs/seajs/blob/master/LICENSE.md))
- [Handlebars.js](https://github.com/wycats/handlebars.js) ([MIT
  License](https://github.com/wycats/handlebars.js/blob/master/LICENSE))
- [normalize.css](https://github.com/necolas/normalize.css) ([MIT
  License](https://github.com/necolas/normalize.css/blob/master/LICENSE.md))
- [FontAwesome](https://github.com/FortAwesome/Font-Awesome/) ([CC BY 3.0 License](http://creativecommons.org/licenses/by/3.0/))
- [Bootstrap](https://github.com/twbs/bootstrap) ([MIT License](https://github.com/twbs/bootstrap/blob/master/LICENSE))
- [UIkit](https://github.com/uikit/uikit) ([MIT License](https://github.com/uikit/uikit/blob/master/LICENSE.md))
- [Foundation](https://github.com/zurb/foundation) ([MIT
  License](https://github.com/zurb/foundation/blob/master/LICENSE))
- [Framework7](https://github.com/nolimits4web/Framework7) ([MIT
  License](https://github.com/nolimits4web/Framework7/blob/master/LICENSE))
- [Alice](https://github.com/aliceui/aliceui.org/) ([MIT
  License](https://github.com/aliceui/aliceui.org/blob/master/LICENSE))
- [Arale](https://github.com/aralejs/aralejs.org/) ([MIT
  License](https://github.com/aralejs/aralejs.org/blob/master/LICENSE))
- [Pure](https://github.com/yui/pure) ([BSD License](https://github.com/yui/pure/blob/master/LICENSE.md))
- [Semantic UI](https://github.com/Semantic-Org/Semantic-UI) ([MIT
  License](https://github.com/Semantic-Org/Semantic-UI/blob/master/LICENSE.md))
- [FastClick](https://github.com/ftlabs/fastclick) ([MIT
  License](https://github.com/ftlabs/fastclick/blob/master/LICENSE))
- [screenfull.js](https://github.com/sindresorhus/screenfull.js) ([MIT
  License](https://github.com/sindresorhus/screenfull.js/blob/gh-pages/license))
- [FlexSlider](https://github.com/woothemes/FlexSlider) ([GPL 2.0](http://www.gnu.org/licenses/gpl-2.0.html))
- [Hammer.js](https://github.com/hammerjs/hammer.js) ([MIT License](https://github.com/hammerjs/hammer.js/blob/master/LICENSE.md))
- [Flat UI](https://github.com/designmodo/Flat-UI) ([CC BY 3.0 and MIT License](https://github.com/designmodo/Flat-UI#copyright-and-license))
- [store.js](https://github.com/marcuswestin/store.js) ([MIT License](https://github.com/marcuswestin/store.js/blob/master/LICENSE))
- [bootstrap-datepicker.js](http://www.eyecon.ro/bootstrap-datepicker/) ([Apache License 2.0](http://www.eyecon.ro/bootstrap-datepicker/js/bootstrap-datepicker.js))
- [iScroll](http://iscrolljs.com/) ([MIT License](http://iscrolljs.com/#license))

可能会有部分项目遗漏，我们会不断整理更新。

### Developed with Open Source Licensed [WebStorm](http://www.jetbrains.com/webstorm/)

<a href="http://www.jetbrains.com/webstorm/" target="_blank">
<img src="http://ww1.sinaimg.cn/large/005yyi5Jjw1elpp6svs2eg30k004i3ye.gif" width="240" />
</a>
