# Amaze UI Change Log
---

### 2015.07 W2

**JS:**

- `New` Popover 增加 `theme` 选项，用于设置颜色和尺寸；
- `New` Popover 增加 `destroy` 方法，用于销毁实例。

## Amaze UI 2.4.2 (2015.07.06)

### 2015.07 W2

**Web 组件：**

- `Fixed` #659 Gotop 修复 ` $.AMUI is undefined` 错误。

## Amaze UI 2.4.1 (2015.06.30)

### 2015.06 W5

**JS:**

- `Improved` Tabs 增加标签切换及选项卡刷新、销毁 API。

**示例：**

- `Fixed` #647 修复后台管理模板在 iOS 部分浏览器上侧边栏按钮点击无效问题。

### 2015.06 W4

**JS:**

- `Update` Add to Homescreen 升级到 `v3.2.2`；
- `Update` NProgress 升级到 `v0.2.0`；
- `New` 添加二维码生产插件 jQuery API；
- `New` Popover 可以通过 `.popover('setContent', content)` 方法设置内容；
- `Fixed` #628 修复 Sticky 获取元素高度不准确问题。

**Web 组件：**

- `Improved` Gallery 展示透明图片时加载图标没有隐藏问题。

### 2015.06 W3

**CSS:**

- `New` Table 所有单元格居中对齐样式 `.am-table-centered`；
- `Improved` Table 全边框样式包含 `rowspan` 时后面几行第一列不显示边框问题。

### 2015.06 W2

**JS:**

- `Fixed` #579 Selected 初始化时触发 `change` 事件问题；
- `Fixed` Selected 设置 `btnWidth` 使用百分比无效问题。

### 2015.06 W1

**JS:**

- `New` #589 Modal 增加「确定」、「取消」时是否关闭的选项；
- `Fixed` 修复使用 `<select>` 选项为空时 Selected 报错问题。

## Amaze UI 2.4.0 (2015.06.01)

### 2015.05 W4

**CSS:**

- `Improved` Article 增加图片最大宽度限制样式。

**JS:**

- `Fixed` #587 修复 cookie 设置 `expires` 无效问题（Thx to @Stephenzxb）。

**模板**

- `New` 新增 `admin-index-fixed.html`：固定顶部和边栏的后台模板。


### 2015.05 W3

**CSS:**

- `Improved` #566 处理没有 `thead` 的 `table` 上边框重复问题（Thx to @Tobyee）；
- `Improved` #565 添加 `margin/padding` 为 `0` 的辅助类（Thx to @Tobyee）。

**JS:**

- `Changed` #567 引入 FastClick 后遇到了许多问题，权衡之后，做移除处理（Thx to @zhongjq）；
- `Improved` 处理 Slider 在部分浏览器中上下翻页箭头文字没有隐藏问题；
- `Improved` 优化 JS 构建流程，加快构建速度，输出 UMD 格式文件。

**Web 组件：**

- `Changed` Divider 增加点线和虚线样式。

### 2015.05 W2

**CSS:**

- `Improved` #556 调整 `select` 在 Firefox 下的样式（Thx to @photino）。

**JS:**

- `Changed` 使用 CSS 定位 Popover 箭头。

### 2015.05 W1

**Web 组件：**

- `Improved` 替换废弃的 `.on('ready')` 语法。

### 2015.04 W4

**CSS:**

- `Fixed` `:root` 基准字号由 `62.5%` 调整为 `10px`，解决某些浏览器的默认字号不是 `16px`；
- `Changed` Button/input 垂直 `padding` 由 `0.625em` 调整为 `0.5em`；
- `Fixed` #392 解决输入框组文本框与按钮在部分浏览器上的对齐问题。

**Example**:

- `Improved` #540 iScroll 示例添加 `click: true` 选项，解决 Android 浏览器上链接不可点击问题。

### 2015.04 W3

**CSS:**

- `Fixed` 修复 Input-group 设置颜色后在 `.am-form` 中无效问题；
- `Improved` Form 表单验证状态图标支持带有 `label` 的 from group；

**JS:**

- `Improved` #528 Validator `H5inputType` 优先使用用户定义的验证规则；
- `Improved` #520 添加设置触发验证程序事件选项 `activeKeyup`、`textareaMaxlenthKeyup`。

### 2015.04 W1

**CSS:**

- `Fixed` `.am-thumbnail` 添加 #502 中移除的样式，限制图片最大宽度。

**JS:**

- `Improved` Tabs 初始化 Hammer 增加判断，避免 DOM 结构不符报错问题（#513）。


## Amaze UI 2.3.0 (2015.04.08)

### 2015.03 W5

**CSS:**

- `Fixed` #502 修复 `img` 样式 `max-width: 100%` 导致部分地图不显示的问题，新增 `.am-img-responsive`。

**JS:**

- `Fixed` #470 表单验证插件验证 `<input type="submit">` 问题；使用 `checkbox`/`radio` 时必须设置 `name` 属性；
- `Improved` #476 尝试修复下拉选框选中状态设置问题；
- `Improved` Selected 增加对 JS 动态操作 `<select>` 支持（插入选项、选中选项、禁用选项等）；
- `Improved` Selected 增加 `disabled` 选项支持；
- `Fixed` #500 修复 Validator 包含异步验证程序时表单提交问题；Validator 包含异步验证时 `.isFormValid()` 返回 Promise，避免返回布尔值不准确的问题；
- `New` #500 Validator 增加返回错误提示信息接口；
- `Fixed` #486 修复 DatePicker 禁用日期后仍能通过月份、年份选择问题（Thx to @meowl）。

**Web 组件：**

- `New` Accordion 增加始终显示（禁用）某些面板选项。

### 2015.03 W4

**CSS:**

- `Improved` 调整 Button `:focus`、`:active` 样式，对二者加以区分；
- `New` Icon 添加图标固定宽度 class `.am-icon-fw` 以及新动画 `.am-icon-pulse`；
- `Improved` 根据网友反馈调整 Icon 字号相关属性设置，以期解决部分设备图标不显示问题（有待进一步手机用户反馈）；
- `Improved` #471 Form 增加 `.am-form-group-sm`、`.am-form-group-lg` class，解决表单组调整大小后对齐问题；
- `New` Form 增加 `.am-form-file`，用于设置 `<input type="file">` 的样式；
- `New` Table 添加 `.am-table-compact` 紧凑样式。
- `New` Animation 添加 `.am-animation-delay-{1,6}` 动画延迟 class。

**JS:**

- `Fixed` #473 Sticky 外边距计算问题修复；
- `Fixed` #457 Button JS 交互复选框在触控设备中无法正确获取值由 FastClick 引起，需在 `input` 上添加 `.needsclick`；
- `Improved` #445 日期选择插件调整为一个页面只能同时打开一个日期选择面板；
- `Improved` #446 Tabs 支持嵌套使用（需禁用触控功能）；
- `Improved` Modal Prompt 打开时自动聚焦输入框，Loading 禁用通过遮罩层关闭；
- `Improved` uCheck 对 `.am-checkbox-inline`/`.am-radio-inline` 支持。

### 2015.03 W3

**JS:**

- `Improved` 确认 `contenteditable` 在触摸屏上失效由 FastClick 导致，解决方法为在需要编辑的元素上添加 `.needsclick` class。

### 2015.03 W2

**JS:**

- `Fixed` #457 修复 PureView 插件动态增加图片以后，调用微图片查看器时数量错误问题，Thx to @woxinqidai；

## Amaze UI 2.2.1

- 修复文件错误！


## Amaze UI 2.2.0

**CSS:**

- `New` 更新 Font Awesome 至 `4.3`，新增 40 个图标以及 `woff2` 格式字体；
- `New` 增加表格元素隐藏/显示辅助 class，参考 #414。

**JS:**

- `New` 增加 Geolocation 封装。

## Amaze UI 2.1.0

### 2015.01 W2

**CSS:**

- `Fixed` 修复网格系统 `.am-u-xx-push-0` 误写成 `.am-u-xx-push-0`；

**JS:**

- `New` Selected 插件增加 `dropUp` 选项，实现上拉效果；

### 2015.01 W1

**JS:**

- `New` 表单验证插件增加等值验证、`validate` 选项供用户自定义验证程序（实现 Ajax 验证等）、`submit` 选项设置表单提交处理程序；
- `New` Datepicker 增加选定日期后自动关闭选项 `autoClose`，样式细节调整；
- `Improved` 阻止 Dropdown toggle 默认事件；
- `Fixed` 修复 Selected 插件点击选择时提交表单问题；
- `Fixed` 修复 PureView 插件标题不显示问题；
- `New` PureView 插件增加 `toggleToolbar` 选项 - 长按是否隐藏/显示工具栏，默认为 `true`；
- `Improved` #379 处理 Selected `$.fn.selected` 与 [jQuery Form](https://github.com/malsup/form) 冲突问题， Thx @Ahoo-Wang；
- `Improved` #367 尝试修复 Modal 遮罩层关闭回调函数被触发多次问题；
- `Update` 升级 FullScreen 至 `2.0.0`，移除 `.onchange()`、`.onerror()` 方法。

## Amaze UI 2.1.0-beta1

### 2014.12 W4

**JS:**

- `Improved` #294 Pureview 插件优化：动态插入图片处理，用户查看时才加载大图，修复部分浏览器上未隐藏完整问题；
- `Improved` 修复 Modal 退出时的动画效果；
- `Improved` #332 修复 Dropdown 嵌套在 Header 里时 `line-height`/`display` 问题；

### 2014.12 W3

**JS:**

- `New` #26 添加 radio / checkbox 样式覆写插件；
- `New` #283 添加表单验证插件；
- `Improved` Popover 增加大小、颜色样式；
- `Changed` iScroll 接口由 `$.AMUI.IScroll` 改为 `$.AMUI.iScroll`；


### 2014.12 W2

**CSS:**

- `Improved` #310 调整 Table 圆角边框实现方式，解决在 IE 下显示问题（THX to [@bug](http://t.qq.com/sky14way), @infinnie）

**JS:**

- `Improved` 注册 Handlebars helper 和 partials 时添加 `this`, via #311 (THX to @msdlisper)；
- `New` #4 Modal 插件增加点击遮罩层是否关闭选项；
- `New` #24 Modal 插件增加宽高设置选项；
- `Improved` Modal 插件 Prompt 支持多个输入框；
- `Changed` Modal 插件 `onConfirm`/`onCancel` 参数变为对象；
- `Changed` #274 Modal 插件 `onConfirm`/`onCancel` 提供参数钩子；
- `Changed` Modal 插件 `relatedElement` 参数变为 `relatedTarget`；

**Web 组件：**

- `Changed` Footer 移除`云适配驱动`；
- `Improved` 处理 Gotop `Fixed` 样式与 Navbar 共存问题。

## Amaze UI 2.0.0

### 2014.12 W1

**CSS**:

- `Improved` #277 Icon Font HTTPS 支持；
- `Changed` #284 移除 `::selection` 样式，避免日文输入文字不显示问题。

**JS**:

- `Fixed` #165 部分浏览器中 Dropdown 收回去时闪烁问题。



## Amaze UI 2.0.0-beta1

### CSS

#### 基础变化

- 移除所有标准属性的浏览器前缀，构建的时候通过 AutoPrefixer 添加；
- 部分组件中开始使用 Flexbox，兼容 IE 9 的样式独立到 `legacy.ie.less`。

#### 网格

- `Changed` #112 重命名网格 class，`{break-point}-col-{n}` -> `.am-u-{break-point}-{n}`;

#### 等分网格

- `Changed` #112 重命名等分网格 class，`{break-point}-block-grid-{n}` -> `.am-avg-{break-point}-{n}`;

#### 图片

- `Changed` 移除 `.am-img-bdrs`，使用 `.am-radius` 替代；
- `Changed` 移除 `.am-img-circle`，使用 `.am-circle` 替代。

#### 表格 Table

- `Changed` `.am-table-bd` 重命名为 `.am-table-bordered`；
- `Changed` `.am-table-bdrs` 重命名为 `.am-table-radius`；
- `New` 单元格状态更多颜色。

#### 评论列表 Comment

- `IMPOROVED` 头像大小调整为响应式，在小屏幕上头像变小；
- `New` 增加评论操作（编辑、删除等）样式。

#### 缩略图 Thumbnail

- `New` 增加缩略图列表样式 `.am-thumbnails`。

### JS

- `Changed` 使用 jQuery 替代 Zepto.js；
- `REMOVED` 不再使用 Sea.js；
- `Changed` 所有模块接口通过 `$.AMUI` 暴露，部分插件添加到 `$.fn`；
- `Changed` 自定义事件命名由 `:` 分隔符改为 `.` 分隔符；
- `Improved` 调整部分代码以兼容 jQuery 及 IE 9；
- `New` #70 Slider 增加初始化接口及用户操作间隔一段时间以后恢复自动播放选项。

### Web 组件

- `Improved` Menu、Header 按钮尺寸增大，更容易点击；
- `REMOVED` 删除搜狐畅言组件。

### 2014.11 W2

__CSS__：

- `Changed` #112 重命名等分网格 class，`{break-point}-block-grid-{n}` -> `.am-avg-{break-point}-{n}`;

## Amaze UI v1.0.0 （2014.10.28）

### 2014.10 W4

__JS插件__：

- `Fixed` #164 Alert 插件 closed:alert:amui 自定义事件未触发；
- `Fixed` #163 Pureview 插件在 Firefox 桌面版异常。

### 2014.10 W3

__CSS__：

- `Improved` 分页组件添加 `.am-disabled` 后移除所有事件（`pointer-events: none;`）；

__Web 组件__：

- `Fixed` #155 修复 Accordion 一个页面中多个共存问题。


## Amaze UI v1.0.0-rc1

### 2014.10 W3

__JS插件__：

- `Improved` #131 window `resize` 时关闭 OffCanvas；
- `New` Tabs 增加 tab 文字无法选择的说明。

__Web 组件__：

- `New` #144 Footer 增加关闭添加到桌面图标接口。

### 2014.10 W2

- 重新梳理代码规范，根据代码规范整理代码；
- 添加 JS 代码质量检测设置；


__JS插件__：

- `Improved` 更新 Add2Home 至 `3.0.8`；
- `Improved` 更新 iScroll 至 `5.1.3`；
- `Fixed` 修复 Share 插件中严格模式报错；

### 2014.10 W1

__CSS__：

- 增加 OSX 平台 Firefox(25+) 字体平滑样式；

__JS插件__：

- `New` Modal 增加宽度响应，视口大于 640 时，modal 的窗口的宽度显示为 540px；
- `Improved` #79 OffCanvas 调整；
- `Improved` #119 PureView loading 图标 z-index 调整，兼容渣浏览器；
- `Improved` #120 PureView 图片容器 flexbox 样式调整，兼容渣浏览器。

__Web 组件__：

- `New` Header 增加固定顶部样式。
- `Fixed` 修复 Figure、Gallery、Paragraph 图片查看变形， via #120 #119


### 2014.09 W5

- `New` 文档标题添加链接按钮；
- `Fixed` #103 #17 Modal 关闭打开/关闭时间间隔过短操作失效问题；
- `New` FastClick 升级到 `1.0.3`；
- `Improved` 调整 FastClick 加载模式，解决在 Chrome 模拟器中链接不能点击的问题；

## Amaze UI v1.0.0-beta2

### 2014.09 W4

#### Build 文件

- 文件名与项目名称统一；
- 增加 `basic`、`widgets` 两个版本供用户选择；
- 上传到 [http://staticfile.org/](http://staticfile.org/)，搜索 amazeui 查看相关文件链接。

#### 官网

- 优化左侧菜单固定效果，菜单高度超过窗口高度时增加滚动条。

#### CSS

- 重新样式梳理 `z-index`；

__Button__：

- `New` 增加圆角和直角样式；
- `Improved` 按钮样式细节调整。

__Button Group__：

- `Changed` 使用 `flexbox` 实现按钮等分。

__Comment__：

- 尖角样式细节调整。

__Form__：

- `Changed` 调整 `select` 样式。

__Input Group__：

- `New` 增加不同颜色样式；
- `Improved` 处理不同尺寸垂直对齐问题。


#### JS 插件

__Button__：

- `New` 增加 loading 文字设置接口；
- `New` 增加 reset 文字设置接口；
- `New` 增加 spinner 设置接口。

__Modal__：

- `Improved` #2 解决打开/关闭时窗口抖动问题；
- `Fixed` #1 解决 Firefox transitionend 事件处理问题（Firefox bug，参见 http://codepen.io/harryfino/full/jphis ）。

__PureView__：

- `New` 增加获取大图选项，可以通过 `a` 的 `href` 或者 `img` 的 `data-rel` 获取大图地址；
- `New` 是否显示分享按钮选项；
- `New` 增加在微信中打开时调用微信图片查看器选项；
- `New` 增加图片 Loading 图标；
- `Improved` 解决打开/关闭窗口抖动问题；
- `Improved` 解决图片比例失调问题。

__Sticky__：

- `Changed` #90 底部边距计算逻辑有问题，暂时取消。

__Tabs__：

- `Improved` #96 改进触控事件处理逻辑，避免标签中有 DOM 元素时触控失效问题。


#### Web 组件

- 重新样式梳理 `z-index`；

__Figure、Paragraph__：

- 移除遗留的无用的样式。

__Figure__：

- `New` `data-rel` 接口，可以设置大图路径；

__Gallery__：

- `New` `data-rel` 接口，可以设置大图路径；
- `Improved` PureView 调用逻辑增强，判断是否设置了 PureView 的选项。


### 2014.09 W3

#### CSS

__Badge__：

- `New` 增加圆角和直角样式。

__Pagination__：

- `New` 增加居中、右对齐样式。

__Topbar__：

- `New` 增加顶部/底部固定样式。

#### JS 插件

__Dropdown__:

- `Improved` #78 增加上拉效果，调整尖角样式；
- `New` #77 增加 `boundary` 选项，判断边界；
- `New` 增加 `justify` 选项，可以设置下拉内容宽度参照对象。

__Smooth Scroll__:

- `Improved` #35 兼容性增强；
- `New` 增加 `speed` 选项。

__Sticky__：

- `Improved` 动态获取元素的宽高，支持响应式；
- `Improved` 解决 #55 中的问题；
- `New` 使用占位符包裹元素，避免窗口抖动问题；
- `New` 增加 `media` 选项，可以设置 Media Query；
- `New` 增加 `bottom` 选项。

__ScrollSpyNav__:

- `Fixed` 链接点击失效问题。

__Popover__：

- `Improved` 每次打开时重新计算弹出层的位置，减少位置偏移问题；
- `Improved` 增加 `focus` 触发；
- `Improved` 样式调整。

__Progress__：

- `Changed` 样式主色调整为绿色。

#### Web 组件

__Gotop__：

- `Improved` #6 兼容 Firefox；
- `New` `Fixed` 主题根据滚动条位置自动显隐。

__Menu__：

- `Improved` #73 调整触发按钮高度。

__Tabs__：

- `Improved` #72 使用 `flexbox` 实现均分，提高微信 webview 兼容性。


### 2014.09 W2

__CSS__：

- `Improved` Close 增加使用 Icon Font 的样式；
- `Improved` Comment 样式调整：采纳网友建议（感谢 [@麦田一根葱](http://www.yuxiaoxi.com/)、[@老黄](http://amazeui.org/css/comment)），增加内容左右对齐；更多高亮颜色；
- `New` Utility 增加 `.am-angle` 尖角样式。

__JS 插件__：

- `Improved` 使用 [hammer.js](https://github.com/hammerjs/hammer.js)，移除 Zepto.js touch 扩展；
- `Improved` Tabs、PureView 使用 hammer.js；
- `Improved` AddToHomeScreen 升级至 `3.0.7`；
- `Changed` 删除 TouchGallery 插件；
- `Fixed` Tabs 修复标签里包含其他元素点击失效问题；
- `Improved` Tabs 没有设置或者多个标签设置了激活状态时默认激活第一个。

__Web 组件__：

- `Improved` Gallery 使用 PureView 插件；
- `Improved` Figure 使用 PureView 插件；
- `Improved` Paragraph 使用 PureView 插件；
- `Changed` 移除 Navbar `package.json` 中图标位置设置选项；

### 2014.09 W1

__JS 插件__：

- `New` 新增移动端图片浏览插件（还在优化完善）。

__Navbar 组件__：

- `Improved` 使用 Flexbox 实现工具栏平均分布；
- `Fixed` 修复二维码 URL。


### 2014.08 W4

__Navbar 组件__：

- `Improved` 重写交互；
- `Improved` 使用 Amaze UI Share 插件，移除百度分享；
- `Changed` 删除图标位置选项，只提供图标在上方的样式；
- `Improved` 主题细节调整。

__JS Tabs 插件__：

- `New` 新增 Tabs 插件，使用 CSS3 实现平滑滚动及回弹效果。

__Tabs 组件__：

- `Improved` 使用新提取的 `Tabs` 插件，解决高度自适应问题；
- `Changed` 调整数据接口，删除标签宽度选项；
- `Improved` 调整主题。

__Paragraph 组件__：

- `Improved` 使用 `iScroll-lite.js` 提高组件中 `table` 拖拽的体验。

__布局示例__：

- 调整路径，下载包中 Amaze UI 相关的资源使用本地文件。

### 2014.08 W3

#### 官网

- 首页增加更新订阅；
- 文档增加目录。

#### Web 组件

- 增加 Web 组件本地预览服务器；
- 公用的 Demo 数据提取到 `package.json` 的 `demoContent` 下面；
- 组件细节调整。


### 2014.08 W2

#### Web 组件

__Titlebar__:

- `New` 更新数据结构；
- `New` 精简主题；
- `New` 页头功能转移到 Header 组件，本组件专注于页内标题栏。

__Pagination__:

- `New` 删除 `options.select` 选项，根据主题判断要生产的 HTML 结构；
- `New` 合并三个类似的主题为 `select`；
- `New` 主题颜色调整，细节优化。

__Accordion__:

- `Improved` 使用 `Collapse` 插件实现交互，移除原来单独写的代码；
- `Improved` 主题细节调整。

__Header__:

- `New` 新增 header 组件，专注于页头功能实现。

__Gotop__

- `New` 新的数据接口；
- `New` 精简主题；
- `New` 根据滚动条位置自动悬浮。


其他组件细节亦有调整。
