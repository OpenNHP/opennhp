# Nav
---

导航样式组件，在 `<ul>` 链接列表中添加 `.am-nav` class。

## 基本样式

`<ul>` 添加 `.am-nav` class 以后就是一个基本的垂直导航。默认样式中并没有限定导航的宽度，可以结合网格使用。

`````html
<ul class="am-nav">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">开始使用</a></li>
  <li><a href="#">按需定制</a></li>
</ul>
`````
```html
<ul class="am-nav">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">开始使用</a></li>
  <li><a href="#">按需定制</a></li>
</ul>
```

## 水平导航

在 `.am-nav` 的基础上再添加 `.am-nav-pills`，形成一个水平导航。

`````html
<ul class="am-nav am-nav-pills">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">开始使用</a></li>
  <li><a href="#">按需定制</a></li>
</ul>
`````
```html
<ul class="am-nav am-nav-pills">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">开始使用</a></li>
  <li><a href="#">按需定制</a></li>
</ul>
```

## 标签式导航

在 `.am-nav` 的基础上添加 `.am-nav-tabs`，形成一个标签式的导航。激活的标签在 `<li>` 上添加 `.am-active`。

`````html
<ul class="am-nav am-nav-tabs">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">开始使用</a></li>
  <li><a href="#">按需定制</a></li>
</ul>
`````
```html
<ul class="am-nav am-nav-tabs">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">开始使用</a></li>
  <li><a href="#">按需定制</a></li>
</ul>
```

## 宽度自适应

在水平导航或标签式导航上添加 `.am-nav-justify` 让 `<li>` 平均分配宽度（通过`display: table-cell` 实现）。

平均分配只在 `media-up` (> 640px) 有效，<= 640px 时菜单会垂直堆叠（缩小浏览器窗口效果可以查看效果）。

`````html
<ul class="am-nav am-nav-pills am-nav-justify">
    <li class="am-active"><a href="#">首页</a></li>
    <li><a href="#">开始使用</a></li>
    <li><a href="#">按需定制</a></li>
    <li><a href="#">加入我们</a></li>
</ul>

<ul class="am-nav am-nav-tabs am-nav-justify">
    <li class="am-active"><a href="#">首页</a></li>
    <li><a href="#">开始使用</a></li>
    <li><a href="#">加入我们</a></li>
</ul>
`````

```html
<ul class="am-nav am-nav-pills am-nav-justify">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">开始使用</a></li>
  <li><a href="#">按需定制</a></li>
  <li><a href="#">加入我们</a></li>
</ul>

<ul class="am-nav am-nav-tabs am-nav-justify">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">开始使用</a></li>
  <li><a href="#">加入我们</a></li>
</ul>
```

## 导航状态

导航状态 class 添加在 `<li>` 上。

- `.am-disabled` - 禁用
- `.am-active` - 激活

`````html
<ul class="am-nav am-nav-pills">
    <li class="am-active"><a href="#">首页</a></li>
    <li><a href="#">关于我们</a></li>
    <li class="am-disabled"><a href="#">禁用链接</a></li>
</ul>
`````

```html
<ul class="am-nav am-nav-pills">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">关于我们</a></li>
  <li class="am-disabled"><a href="#">禁用链接</a></li>
</ul>
```


## 导航标题及分隔线

导航标题及分隔线目前仅适用于垂直菜单。

- `.am-nav-header` 导航标题，直接放在 `<li>` 中。
- `.am-nav-divider` 导航分隔线，添加到空的 `<li>` 上。

`````html
<ul class="am-nav">
  <li><a href="#"><span class="am-icon-home"></span>首页</a></li>
  <li class="am-nav-header">开始使用</li>
  <li><a href="#">关于我们</a></li>
  <li><a href="#">联系我们</a></li>
  <li class="am-nav-divider"></li>
  <li><a href="#">响应式</a></li>
  <li><a href="#">移动优先</a></li>
</ul>
`````

```html
<ul class="am-nav">
  <li><a href="#">首页</a></li>
  <li class="am-nav-header">开始使用</li>
  <li><a href="#">关于我们</a></li>
  <li><a href="#">联系我们</a></li>
  <li class="am-nav-divider"></li>
  <li><a href="#">响应式</a></li>
  <li><a href="#">移动优先</a></li>
</ul>
```


## 下拉菜单

需结合 JS [Dropdown](/##) 组件使用。

`````html
<ul class="am-nav am-nav-pills">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">项目</a></li>
  <li class="am-dropdown" data-am-dropdown>
    <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
      菜单 <span class="am-icon-caret-down"></span>
    </a>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">Header</li>
      <li><a href="#">1. 一行代码，简单快捷</a></li>
      <li class="am-active"><a href="#">2. 网址不变且唯一</a></li>
      <li><a href="#">3. 内容实时同步更新</a></li>
      <li class="am-disabled"><a href="#">4. 云端跨平台适配</a></li>
      <li class="am-divider"></li>
      <li><a href="#">5. 专属的一键拨叫</a></li>
    </ul>
  </li>
</ul>
`````

```html
<ul class="am-nav am-nav-pills">
    <li class="am-active"><a href="#">首页</a></li>
    <li><a href="#">项目</a></li>
    <li class="am-dropdown" data-am-dropdown>
        <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
            菜单 <span class="am-icon-caret-down"></span>
        </a>
        <ul class="am-dropdown-content">
            <li class="am-dropdown-header">Header</li>
            <li><a href="#">1. 一行代码，简单快捷</a></li>
            <li class="am-active"><a href="#">2. 网址不变且唯一</a></li>
            <li><a href="#">3. 内容实时同步更新</a></li>
            <li class="am-disabled"><a href="#">4. 云端跨平台适配</a></li>
            <li class="am-divider"></li>
            <li><a href="#">5. 专属的一键拨叫</a></li>
          </ul>
    </li>
</ul>
```


### Tab 式


`````html
<ul class="am-nav am-nav-tabs">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">项目</a></li>
  <li class="am-dropdown" data-am-dropdown>
    <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
      菜单 <span class="am-icon-caret-down"></span>
    </a>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">Header</li>
      <li><a href="#">1. 一行代码，简单快捷</a></li>
      <li class="am-active"><a href="#">2. 网址不变且唯一</a></li>
      <li><a href="#">3. 内容实时同步更新</a></li>
      <li class="am-disabled"><a href="#">4. 云端跨平台适配</a></li>
      <li class="am-divider"></li>
      <li><a href="#">5. 专属的一键拨叫</a></li>
    </ul>
  </li>
</ul>
`````

```html
<ul class="am-nav am-nav-tabs">
  <li class="am-active"><a href="#">首页</a></li>
  <li><a href="#">项目</a></li>
  <li class="am-dropdown" data-am-dropdown>
    <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
      菜单 <span class="am-icon-caret-down"></span>
    </a>
    <ul class="am-dropdown-content">
      ...
    </ul>
  </li>
</ul>
```
