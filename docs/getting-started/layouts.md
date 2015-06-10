# 布局示例
---

Amaze UI 目前提供几个常见的布局示例，供开发者参考，后续会增加更多示例。

---
**关于图标显示异常的说明：**

~~为避免 `Mixed Content` 错误，Icon Font 的 URL 没有添加协议，直接使用磁盘路径 `file:///...` 打开时示例页面时无法正常显示图标，请放在 HTTP 服务中查看。~~

**最新的下载包中已经将字体路径替换为本地路径！**

- 使用 JetBrais 系（WebStorm 等）打开示例文件夹，然后在编辑器里点预览按钮；
- `cd` 到示例目录，`python -m SimpleHTTPServer`；
- 或者使用其他 HTTP 服务器。

---

## 示例说明

<div class="am-g">
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/landing.html">
      <img class="am-img-thumbnail"
           src="http://amazeui.org/i/examples/landingPage.png"
           alt="landingPage"/>
      <figcaption class="am-thumbnail-caption">展示页面</figcaption>
    </a>
  </div>
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/login.html">
      <img class="am-img-thumbnail"
           src="http://amazeui.org/i/examples/loginPage.png" alt="loginPage"/>
      <figcaption class="am-thumbnail-caption">登录注册页面</figcaption>
    </a>
  </div>
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/blog.html">
      <img class="am-img-thumbnail"
           src="http://amazeui.org/i/examples/blogPage.png" alt="blogPage"/>
      <figcaption class="am-thumbnail-caption">博客页面</figcaption>
    </a>
  </div>
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/sidebar.html">
      <img class="am-img-thumbnail"
           src="http://amazeui.org/i/examples/sidebarPage.png"
           alt="sidebarPage"/>
      <figcaption class="am-thumbnail-caption">侧边栏页面</figcaption>
    </a>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/admin-index.html">
      <img class="am-img-thumbnail" src="http://ww1.sinaimg.cn/large/005yyi5Jjw1elpr8svtoyj30j70demxe.jpg
" alt="adminPage"/>
      <figcaption class="am-thumbnail-caption">管理后台模板</figcaption>
    </a>
  </div>
</div>

### 展示页面

使用组件说明：

<ul>
  <li>CSS 部分：<a class="am-badge am-badge-success" href="/css/grid">网格</a> <a
    class="am-badge am-badge-success" href="/css/icon">图标</a> <a
    class="am-badge am-badge-success" href="/css/button">按钮</a> <a
    class="am-badge am-badge-success" href="/css/form">表单</a> <a
    class="am-badge am-badge-success" href="/css/article">文章页</a> <a
    class="am-badge am-badge-success" href="/css/topbar">导航条</a> <a
    class="am-badge am-badge-success" href="/css/utility">辅助类</a></li>
  <li>JS 插件部分：<a class="am-badge am-badge-primary" href="/javascript/dropdown">下拉组件</a>
    <a class="am-badge am-badge-primary" href="/javascript/scrollspy">滚动侦测</a>
  </li>
</ul>

### 登录页面

使用组件说明：

<ul>
  <li>CSS 部分：<a class="am-badge am-badge-success" href="/css/grid">网格</a>
    <a class="am-badge am-badge-success" href="/css/icon">图标</a>
    <a class="am-badge am-badge-success" href="/css/button">按钮</a>
    <a class="am-badge am-badge-success" href="/css/button-group">按钮组</a>
    <a class="am-badge am-badge-success" href="/css/form">表单</a></li>
</ul>

### 博客页面

使用组件说明：

<ul>
  <li>CSS 部分：<a class="am-badge am-badge-success" href="/css/grid">网格</a> <a
    class="am-badge am-badge-success" href="/css/block-grid">等宽布局</a> <a
    class="am-badge am-badge-success" href="/css/icon">图标</a> <a
    class="am-badge am-badge-success" href="/css/button">按钮</a> <a
    class="am-badge am-badge-success" href="/css/panel">面板</a> <a
    class="am-badge am-badge-success" href="/css/list">列表</a> <a
    class="am-badge am-badge-success" href="/css/pagination">分页</a></li>
  <li>JS 插件部分：<a class="am-badge am-badge-primary" href="/css/dropdown">下拉组件</a>
  </li>
</ul>

### 博客侧栏页面

使用组件说明：

<ul>
    <li>CSS 部分：<a class="am-badge am-badge-success" href="/css/grid">网格</a> <a class="am-badge am-badge-success" href="/css/article">文章页</a> <a class="am-badge am-badge-success" href="/css/comment">评论列表</a> <a class="am-badge am-badge-success" href="/css/button">按钮</a> <a class="am-badge am-badge-success" href="/css/icon">图标</a> <a class="am-badge am-badge-success" href="/css/list">列表</a> <a class="am-badge am-badge-success" href="/css/utility">辅助类</a></li>
    <li>JS 插件部分：<a class="am-badge am-badge-primary" href="/javascript/offcanvas">侧边栏组件</a></li>
</ul>

### 管理后台模板

- [index 页面](/examples/admin-index.html)
- [user 页面](/examples/admin-user.html)
- [gallery 页面](/examples/admin-gallery.html)
- [table 页面](/examples/admin-table.html)
- [form 页面](/examples/admin-form.html)
- [help 页面](/examples/admin-help.html)
- [log 页面](/examples/admin-log.html)
- [404 页面](/examples/admin-404.html)

## 禁用响应式

不喜欢响应式？可以尝试禁用：

- 删除 `head` 里的视口设置 `meta` 标签；

```html
<!--<meta name="viewport"
    content="width=device-width, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no">-->
```

- 固定容器 `.am-container` 宽度（可以自己添加一个 class，不一定要使用内置的）：

```css
.am-container {
  width: 980px !important;
  max-width: none;
}
```

- 使用网格系统时，只添加 `.am-u-sm-*` class，移除其他断点的 class。

至此，布局层的响应式被禁用了（[参考示例](/examples/non-responsive.html)）。
