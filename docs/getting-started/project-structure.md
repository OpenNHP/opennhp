# 基于 Amaze UI 二次开发
---

使用中有何问题，请直接在[评论](#ds-thread)中留言，我们会不断补充完善文档。

## 项目结构

```
Amaze UI
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

## LESS

`less` 文件中存放了所有 LESS 源文件，其中以 `ui.` 开头的为 JS 插件相关的样式。

下面两个文件为开发中可能需要关注的：

- `mixins.less` - LESS 常用函数封装
- `variables.less` - 所有变量列表

```
less
|-- amui.all.less
|-- amui.less
|-- animation.less
|-- article.less
|-- badge.less
|-- base.less
|-- block-grid.less
|-- breadcrumb.less
|-- button-group.less
|-- button.less
|-- close.less
|-- code.less
|-- color-schemes
|-- comment.less
|-- form.less
|-- grid.less
|-- icon.less
|-- image.less
|-- input-group.less
|-- list.less
|-- mixins.less
|-- nav.less
|-- pagination.less
|-- panel.less
|-- print.less
|-- progress.less
|-- table.less
|-- themes
|-- thumbnail.less
|-- topbar.less
|-- ui.accordion.less
|-- ui.add2home.less
|-- ui.alert.less
|-- ui.component.less
|-- ui.dropdown.less
|-- ui.modal.less
|-- ui.offcanvas.less
|-- ui.popover.less
|-- ui.progress.less
|-- utility.less
|-- variables.less
`-- zepto.touchgallery.less
```

## JavaScript

JS 文件分三类：

- `ui.` 开头的为 [JS 插件](/javascript)中的【UI 增强】部分；
- `util.` 开头的为 [JS 插件](/javascript)中的 【使用工具】部分；
- `zepto.` 开头的为 Zepto 的一些扩展及插件（可能会被替换为 jQuery）。

```
js
|-- core.js         # 一些基础方法
|-- nav.js          # 遗留组件，将会被重构
|-- ui.accordion.js # 遗留组件，将会被移除
|-- ui.add2home.js  # iOS 添加到桌面
|-- ui.alert.js
|-- ui.button.js
|-- ui.collapse.js
|-- ui.dimmer.js
|-- ui.dropdown.js
|-- ui.modal.js
|-- ui.offcanvas.js
|-- ui.popover.js
|-- ui.progress.js
|-- ui.scrollspy.js
|-- ui.scrollspynav.js
|-- ui.smooth-scroll.js
|-- ui.sticky.js
|-- util.cookie.js
|-- util.fastclick.js
|-- util.fullscreen.js
|-- util.qrcode.js
|-- zepto.extend.data.js
|-- zepto.extend.fx.js
|-- zepto.extend.selector.js
|-- zepto.extend.touch.js
|-- zepto.flexslider.js
|-- zepto.outerdemension.js
|-- zepto.pinchzoom.js
`-- zepto.touchgallery.js
```