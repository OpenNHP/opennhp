# 基于 Amaze UI 进阶开发
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
|-- fonts       # Icon font，使用 Font Awesome
|-- gulpfile.js # 构建配置文件
|-- js          # JS 文件
|-- less        # Less 文件
|-- tools       # 相关工具
|-- vendor      # 第三方文件
`-- widget      # Web 组件
```

## Less

`less` 文件中存放了所有 LESS 源文件，其中以 `ui.` 开头的为 JS 插件相关的样式。

下面两个文件为开发中可能需要关注的：

- `mixins.less` - LESS 常用函数封装
- `variables.less` - 所有变量列表

```
less
|-- amazeui.less          // Amaze UI 所有样式文件
|-- amui.less             // CSS、JavaScript 样式，不包含 Web 组件
├── amazeui.less
├── amui.less
├── animation.less
├── article.less
├── badge.less
├── base.less
├── block-grid.less
├── breadcrumb.less
├── button-group.less
├── button.less
├── close.less
├── code.less
├── comment.less
├── form.less
├── grid.less
├── icon.less
├── image.less
├── input-group.less
├── legacy.ie.less
├── list.less
├── mixins.less
├── nav.less
├── pagination.less
├── panel.less
├── print.less
├── progress.less
├── table.less
├── thumbnail.less
├── topbar.less
├── ui.add2home.less
├── ui.alert.less
├── ui.component.less
├── ui.datepicker.less
├── ui.dropdown.less
├── ui.flexslider.less
├── ui.modal.less
├── ui.offcanvas.less
├── ui.popover.less
├── ui.progress.less
├── ui.pureview.less
├── ui.selected.less
├── ui.share.less
├── ui.tabs.less
├── ui.ucheck.less
├── utility.less
└── variables.less

```

## JavaScript

JS 文件分三类：

- `ui.` 开头的为 [JS 插件](/javascript)中的【UI 增强】部分；
- `util.` 开头的为 [JS 插件](/javascript)中的 【实用工具】部分。

```
js
├── core.js
├── ui.add2home.js
├── ui.alert.js
├── ui.button.js
├── ui.collapse.js
├── ui.datepicker.js
├── ui.dimmer.js
├── ui.dropdown.js
├── ui.flexslider.js
├── ui.iscroll-lite.js
├── ui.modal.js
├── ui.offcanvas.js
├── ui.pinchzoom.js
├── ui.popover.js
├── ui.progress.js
├── ui.pureview.js
├── ui.scrollspy.js
├── ui.scrollspynav.js
├── ui.selected.js
├── ui.share.js
├── ui.smooth-scroll.js
├── ui.sticky.js
├── ui.tabs.js
├── ui.ucheck.js
├── ui.validator.js
├── util.cookie.js
├── util.fullscreen.js
├── util.geolocation.js
├── util.hammer.js
├── util.qrcode.js
└── util.store.js
```
