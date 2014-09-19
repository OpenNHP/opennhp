# Amaze UI Change Log
---

## 2014.09 W3

### CSS

__Badge__：

- `NEW` 增加圆角和直角样式。

__Pagination__：

- `NEW` 增加居中、右对齐样式。

__Topbar__：

- `NEW` 增加顶部/底部固定样式。

### JS 插件

__Dropdown__:

- `IMPROVED` #78 增加上拉效果，调整尖角样式；
- `NEW` #77 增加 `boundary` 选项，判断边界；
- `NEW` 增加 `justify` 选项，可以设置下拉内容宽度参照对象。

__Smooth Scroll__:

- `IMPROVED` #35 兼容性增强；
- `NEW` 增加 `speed` 选项。

__Sticky__：

- `IMPROVED` 动态获取元素的宽高，支持响应式；
- `IMPROVED` 解决 #55 中的问题；
- `NEW` 使用占位符包裹元素，避免窗口抖动问题；
- `NEW` 增加 `media` 选项，可以设置 Media Query；
- `NEW` 增加 `bottom` 选项。

__ScrollSpyNav__:

- `FIXED` 链接点击失效问题。

__Popover__：

- `IMPROVED` 每次打开时重新计算弹出层的位置，减少位置偏移问题；
- `IMPROVED` 增加 `focus` 触发；
- `IMPROVED` 样式调整。

__Progress__：

- `CHANGED` 样式主色调整为绿色。

### Web 组件

__Gotop__：

- `IMPROVED` #6 兼容 Firefox；
- `NEW` `fixed` 主题根据滚动条位置自动显隐。

__Menu__：

- `IMPROVED` #73 调整触发按钮高度。

__Tabs__：

- `IMPROVED` #72 使用 `flexbox` 实现均分，提高微信 webview 兼容性。


## 2014.09 W2

__CSS__：

- `IMPROVED` Close 增加使用 Icon Font 的样式；
- `IMPROVED` Comment 样式调整：采纳网友建议（感谢 [@麦田一根葱](http://www.yuxiaoxi.com/)、[@老黄](http://amazeui.org/css/comment)），增加内容左右对齐；更多高亮颜色；
- `NEW` Utility 增加 `.am-angle` 尖角样式。

__JS 插件__：

- `IMPROVED` 使用 [hammer.js](https://github.com/hammerjs/hammer.js)，移除 Zepto.js touch 扩展；
- `IMPROVED` Tabs、PureView 使用 hammer.js；
- `IMPROVED` AddToHomeScreen 升级至 `3.0.7`；
- `CHANGED` 删除 TouchGallery 插件；
- `FIXED` Tabs 修复标签里包含其他元素点击失效问题；
- `IMPROVED` Tabs 没有设置或者多个标签设置了激活状态时默认激活第一个。

__Web 组件__：

- `IMPROVED` Gallery 使用 PureView 插件；
- `IMPROVED` Figure 使用 PureView 插件；
- `IMPROVED` Paragraph 使用 PureView 插件；
- `CHANGED` 移除 Navbar `package.json` 中图标位置设置选项；

## 2014.09 W1

__JS 插件__：

- `NEW` 新增移动端图片浏览插件（还在优化完善）。

__Navbar 组件__：

- `IMPROVED` 使用 Flexbox 实现工具栏平均分布；
- `FIXED` 修复二维码 URL。


## 2014.08 W4

__Navbar 组件__：

- `IMPROVED` 重写交互；
- `IMPROVED` 使用 Amaze UI Share 插件，移除百度分享；
- `CHANGED` 删除图标位置选项，只提供图标在上方的样式；
- `IMPROVED` 主题细节调整。

__JS Tabs 插件__：

- `NEW` 新增 Tabs 插件，使用 CSS3 实现平滑滚动及回弹效果。

__Tabs 组件__：

- `IMPROVED` 使用新提取的 `Tabs` 插件，解决高度自适应问题；
- `CHANGED` 调整数据接口，删除标签宽度选项；
- `IMPROVED` 调整主题。

__Paragraph 组件__：

- `IMPROVED` 使用 `iScroll-lite.js` 提高组件中 `table` 拖拽的体验。

__布局示例__：

- 调整路径，下载包中 Amaze UI 相关的资源使用本地文件。

## 2014.08 W3

### 官网

- 首页增加更新订阅；
- 文档增加目录。

### Web 组件

- 增加 Web 组件本地预览服务器；
- 公用的 Demo 数据提取到 `package.json` 的 `demoContent` 下面；
- 组件细节调整。


## 2014.08 W2

### Web 组件

__Titlebar__:

- `NEW` 更新数据结构；
- `NEW` 精简主题；
- `NEW` 页头功能转移到 Header 组件，本组件专注于页内标题栏。

__Pagination__:

- `NEW` 删除 `options.select` 选项，根据主题判断要生产的 HTML 结构；
- `NEW` 合并三个类似的主题为 `select`；
- `NEW` 主题颜色调整，细节优化。

__Accordion__:

- `IMPROVED` 使用 `Collapse` 插件实现交互，移除原来单独写的代码；
- `IMPROVED` 主题细节调整。

__Header__:

- `NEW` 新增 header 组件，专注于页头功能实现。

__Gotop__

- `NEW` 新的数据接口；
- `NEW` 精简主题；
- `NEW` 根据滚动条位置自动悬浮。


其他组件细节亦有调整。
