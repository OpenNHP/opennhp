# Advanced Develop Based on Amaze UI
---

Please leave us messages in [comments](#ds-thread), if you have any problems. We will continously improve and update our Docs.

## Files Structure

```
Amaze UI
|-- HISTORY.md
|-- LICENSE
|-- README.md
|-- package.json
|-- dist        # deployment folder
|-- docs        # docs
|-- fonts       # Icon font, currently using http://staticfile.org/
|-- gulpfile.js # build configuration file
|-- js          # JS files
|-- less        # LESS files
|-- tools       # tools
|-- vendor      # third party files
`-- widget      # Web widgets
```

## LESS

`less` folder contains all the LESS sourse files. Those begin with `ui.` are style files for JS plugins.

The following two files are important in development:

- `mixins.less` - LESS common functions
- `variables.less` - variable list

```
less
|-- amazeui.less          // Import all Amaze UI style files
|-- amazeui.widgets.less  // Built by gulp. Web widgets and styles they rely on.
|-- amui.less             // CSS part styles
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
├── themes       // Currently not used.
├── thumbnail.less
├── topbar.less
├── ui.add2home.less
├── ui.alert.less
├── ui.component.less
├── ui.dropdown.less
├── ui.flexslider.less
├── ui.modal.less
├── ui.offcanvas.less
├── ui.popover.less
├── ui.progress.less
├── ui.pureview.less
├── ui.share.less
├── ui.tabs.less
├── utility.less
└── variables.less

```

## JavaScript

There are three kinds of JS files：

- `ui.` prefix means this is a part of UI enhancements in [JS plugins](/javascript);
- `util.` prefix means this is a part of utilities in [JS plugins](/javascript).

```
js
├── amazeui.legacy.js // 针对 IE 8 打包的 JS
├── core.js
├── ui.add2home.js
├── ui.alert.js
├── ui.button.js
├── ui.collapse.js
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
├── ui.share.js
├── ui.smooth-scroll.js
├── ui.sticky.js
├── ui.tabs.js
├── util.cookie.js
├── util.fastclick.js
├── util.fullscreen.js
├── util.hammer.js
└── util.qrcode.js
```
