# Topbar 导航条
---

常用于网页顶部放置 Logo、导航等信息（这个名称起得不太好，出现了 `top` 这个表象的词）。

由于需求的差异性，很难整理出一个适合不同网站的导航条组件，Amaze UI 现在整理的版本适合导航条相对简单（Logo + 链接 + 按钮 + 搜索框）的页面，后续还会进一步优化。

## 基本样式

### 默认样式（浅色）

在容器上添加 `.am-topbar` class，然后按照示例组织所需内容。

`````html
<header class="am-topbar">
  <h1 class="am-topbar-brand">
    <a href="#">Amaze UI</a>
  </h1>

  <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only" data-am-collapse="{target: '#doc-topbar-collapse'}"><span class="am-sr-only">导航切换</span> <span class="am-icon-bars"></span></button>

  <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse">
    <ul class="am-nav am-nav-pills am-topbar-nav">
      <li class="am-active"><a href="#">首页</a></li>
      <li><a href="#">项目</a></li>
      <li class="am-dropdown" data-am-dropdown>
        <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
          下拉 <span class="am-icon-caret-down"></span>
        </a>
        <ul class="am-dropdown-content">
          <li class="am-dropdown-header">标题</li>
          <li><a href="#">1. 去月球</a></li>
          <li class="am-active"><a href="#">2. 去火星</a></li>
          <li><a href="#">3. 还是回地球</a></li>
          <li class="am-disabled"><a href="#">4. 下地狱</a></li>
          <li class="am-divider"></li>
          <li><a href="#">5. 桥头一回首</a></li>
        </ul>
      </li>
    </ul>

    <form class="am-topbar-form am-topbar-left am-form-inline" role="search">
      <div class="am-form-group">
        <input type="text" class="am-form-field am-input-sm" placeholder="搜索">
      </div>
    </form>

    <div class="am-topbar-right">
      <div class="am-dropdown" data-am-dropdown="{boundary: '.am-topbar'}">
        <button class="am-btn am-btn-secondary am-topbar-btn am-btn-sm am-dropdown-toggle" data-am-dropdown-toggle>其他 <span class="am-icon-caret-down"></span></button>
        <ul class="am-dropdown-content">
          <li><a href="#">注册</a></li>
          <li><a href="#">随便看看</a></li>
        </ul>
      </div>
    </div>

    <div class="am-topbar-right">
      <button class="am-btn am-btn-primary am-topbar-btn am-btn-sm">登录</button>
    </div>
  </div>
</header>
`````

```html
<header class="am-topbar">
  <h1 class="am-topbar-brand">
    <a href="#">Amaze UI</a>
  </h1>

  <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only" data-am-collapse="{target: '#doc-topbar-collapse'}"><span class="am-sr-only">导航切换</span> <span class="am-icon-bars"></span></button>

  <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse">
    <ul class="am-nav am-nav-pills am-topbar-nav">
      <li class="am-active"><a href="#">首页</a></li>
      <li><a href="#">项目</a></li>
      <li class="am-dropdown" data-am-dropdown>
        <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
          下拉 <span class="am-icon-caret-down"></span>
        </a>
        <ul class="am-dropdown-content">
          <li class="am-dropdown-header">标题</li>
          <li><a href="#">1. 去月球</a></li>
          <li class="am-active"><a href="#">2. 去火星</a></li>
          <li><a href="#">3. 还是回地球</a></li>
          <li class="am-disabled"><a href="#">4. 下地狱</a></li>
          <li class="am-divider"></li>
          <li><a href="#">5. 桥头一回首</a></li>
        </ul>
      </li>
    </ul>

    <form class="am-topbar-form am-topbar-left am-form-inline" role="search">
      <div class="am-form-group">
        <input type="text" class="am-form-field am-input-sm" placeholder="搜索">
      </div>
    </form>

    <div class="am-topbar-right">
      <div class="am-dropdown" data-am-dropdown="{boundary: '.am-topbar'}">
        <button class="am-btn am-btn-secondary am-topbar-btn am-btn-sm am-dropdown-toggle" data-am-dropdown-toggle>其他 <span class="am-icon-caret-down"></span></button>
        <ul class="am-dropdown-content">
          <li><a href="#">注册</a></li>
          <li><a href="#">随便看看</a></li>
        </ul>
      </div>
    </div>

    <div class="am-topbar-right">
      <button class="am-btn am-btn-primary am-topbar-btn am-btn-sm">登录</button>
    </div>
  </div>
</header>
```

### 深色样式

在容器上添加 `.am-topbar-inverse`，调整为背景色为主色调的样式，内部结构同上。

`````html
<header class="am-topbar am-topbar-inverse">
  <h1 class="am-topbar-brand">
    <a href="#">Amaze UI</a>
  </h1>

  <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-secondary am-show-sm-only" data-am-collapse="{target: '#doc-topbar-collapse-2'}"><span class="am-sr-only">导航切换</span> <span class="am-icon-bars"></span></button>

  <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse-2">
    <ul class="am-nav am-nav-pills am-topbar-nav">
      <li class="am-active"><a href="#">首页</a></li>
      <li><a href="#">项目</a></li>
      <li class="am-dropdown" data-am-dropdown>
        <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
          下拉 <span class="am-icon-caret-down"></span>
        </a>
        <ul class="am-dropdown-content">
          <li class="am-dropdown-header">标题</li>
          <li><a href="#">1. 去月球</a></li>
          <li class="am-active"><a href="#">2. 去火星</a></li>
          <li><a href="#">3. 还是回地球</a></li>
          <li class="am-disabled"><a href="#">4. 下地狱</a></li>
          <li class="am-divider"></li>
          <li><a href="#">5. 桥头一回首</a></li>
        </ul>
      </li>
    </ul>

    <form class="am-topbar-form am-topbar-left am-form-inline" role="search">
      <div class="am-form-group">
        <input type="text" class="am-form-field am-input-sm" placeholder="搜索">
      </div>
    </form>

    <div class="am-topbar-right">
      <div class="am-dropdown" data-am-dropdown="{boundary: '.am-topbar'}">
        <button class="am-btn am-btn-secondary am-topbar-btn am-btn-sm am-dropdown-toggle" data-am-dropdown-toggle>其他 <span class="am-icon-caret-down"></span></button>
        <ul class="am-dropdown-content">
          <li><a href="#">注册</a></li>
          <li><a href="#">随便看看</a></li>
        </ul>
      </div>
    </div>

    <div class="am-topbar-right">
      <button class="am-btn am-btn-primary am-topbar-btn am-btn-sm">登录</button>
    </div>
  </div>
</header>
`````

```html
<header class="am-topbar am-topbar-inverse">
  ...
</header>
```

### Logo 图片替换

前端的有趣之处，不是给你一个美轮美奂的城堡，而是使用简单甚至枯燥的组件，通过发挥自己的想象，创造出属于自己的那个小屋。

结合[辅助 Class](/css/utility) `.am-text-ir` 使用，需要手动设置替换的背景图片。

`````html
<header class="am-topbar am-topbar-inverse">
  <h1 class="am-topbar-brand">
    <a href="#" class="am-text-ir">Amaze UI</a>
  </h1>

  <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only"
          data-am-collapse="{target: '#doc-topbar-collapse-3'}"><span class="am-sr-only">导航切换</span> <span
      class="am-icon-bars"></span></button>

  <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse-3">
    <ul class="am-nav am-nav-pills am-topbar-nav">
      <li class="am-active"><a href="#">首页</a></li>
      <li><a href="#">项目</a></li>
      <li class="am-dropdown" data-am-dropdown>
        <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
          下拉 <span class="am-icon-caret-down"></span>
        </a>
        <ul class="am-dropdown-content">
          <li><a href="#">带我去月球</a></li>
          <li><a href="#">还是回地球</a></li>
          <li class="am-disabled"><a href="#">臣妾做不到</a></li>
        </ul>
      </li>
    </ul>
  </div>
</header>
`````

```html
<header class="am-topbar am-topbar-inverse">
  <h1 class="am-topbar-brand">
    <a href="#" class="am-text-ir">Amaze UI</a>
  </h1>
  ...
</header>
```

```css
.am-topbar .am-text-ir {
  display: block;
  margin-right: 10px;
  height: 50px;
  width: 125px;
  background: url(http://a.static.amazeui.org/assets/i/ui/logo.png) no-repeat left center;
  -webkit-background-size: 125px 24px;
  background-size: 125px 24px;
}
```

## 顶部、底部固定

### 顶部固定

在 `.am-topbar` 上添加  `.am-topbar-fixed-top` class，实现顶部固定。

包含顶部固定导航条的页面会在 `body` 上添加 `.am-with-topbar-fixed-top`，用户可以在这个 class 下设置样式，默认设置了 `padding-top: 51px`。

```html
<header class="am-topbar am-topbar-inverse am-topbar-fixed-top">
  <div class="am-container">
    <h1 class="am-topbar-brand">
      <a href="#" class="am-text-ir">Amaze UI</a>
    </h1>

    <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only"
            data-am-collapse="{target: '#doc-topbar-collapse-4'}"><span class="am-sr-only">导航切换</span> <span
        class="am-icon-bars"></span></button>

    <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse-4">
      <ul class="am-nav am-nav-pills am-topbar-nav">
        <li class="am-active"><a href="#">首页</a></li>
        <li><a href="#">项目</a></li>
        <li class="am-dropdown" data-am-dropdown>
          <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
            下拉 <span class="am-icon-caret-down"></span>
          </a>
          <ul class="am-dropdown-content">
            <li><a href="#">带我去月球</a></li>
            <li><a href="#">还是回地球</a></li>
            <li class="am-disabled"><a href="#">臣妾做不到</a></li>
          </ul>
        </li>
      </ul>
    </div>
  </div>
</header>
```

### 底部固定

在 `.am-topbar` 上添加  `.am-topbar-fixed-bottom` class，实现底部固定。

包含底部固定导航条的页面会在 `body` 上添加 `.am-with-topbar-fixed-bottom`，用户可以在这个 class 下设置样式，默认设置了 `padding-bottom: 51px`。

```html
<header class="am-topbar am-topbar-inverse am-topbar-fixed-bottom">
  <div class="am-container">
    <h1 class="am-topbar-brand">
      <a href="#" class="am-text-ir">Amaze UI</a>
    </h1>

    <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only"
            data-am-collapse="{target: '#doc-topbar-collapse-5'}"><span class="am-sr-only">导航切换</span> <span
        class="am-icon-bars"></span></button>

    <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse-5">
      <ul class="am-nav am-nav-pills am-topbar-nav">
        <li class="am-active"><a href="#">首页</a></li>
        <li><a href="#">项目</a></li>
        <li class="am-dropdown am-dropdown-up" data-am-dropdown>
          <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
            上拉 <span class="am-icon-caret-up"></span>
          </a>
          <ul class="am-dropdown-content">
            <li><a href="#">带我去月球</a></li>
            <li><a href="#">还是回地球</a></li>
            <li class="am-disabled"><a href="#">臣妾做不到</a></li>
          </ul>
        </li>
      </ul>
    </div>
  </div>
</header>
```



