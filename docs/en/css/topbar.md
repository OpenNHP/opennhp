# Topbar
---

Topbar is used on the top of a page, and usually contains logo and navigator.

It's hard do design a topbar which can be used in all websites. So in Amaze UI, we define this simple topbar with Logo, Navigator, Buttons and Searchbox.

## Usage

### Default Style (Light)

Add the `.am-topbar` class to container, and then add components as you need.
`````html
<header class="am-topbar">
  <h1 class="am-topbar-brand">
    <a href="#">Amaze UI</a>
  </h1>

  <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only" data-am-collapse="{target: '#doc-topbar-collapse'}"><span class="am-sr-only">Responsive button</span> <span class="am-icon-bars"></span></button>

  <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse">
    <ul class="am-nav am-nav-pills am-topbar-nav">
      <li class="am-active"><a href="#">Home</a></li>
      <li><a href="#">Project</a></li>
      <li class="am-dropdown" data-am-dropdown>
        <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
          Dropdown <span class="am-icon-caret-down"></span>
        </a>
        <ul class="am-dropdown-content">
          <li class="am-dropdown-header">Title</li>
          <li><a href="#">1. To Moon</a></li>
          <li class="am-active"><a href="#">2. To Mars</a></li>
          <li><a href="#">3. To Earth</a></li>
          <li class="am-disabled"><a href="#">4. To Sun</a></li>
          <li class="am-divider"></li>
          <li><a href="#">5. Go Back</a></li>
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
        <button class="am-btn am-btn-secondary am-topbar-btn am-btn-sm am-dropdown-toggle" data-am-dropdown-toggle>Other <span class="am-icon-caret-down"></span></button>
        <ul class="am-dropdown-content">
          <li><a href="#">Register</a></li>
          <li><a href="#">What you like</a></li>
        </ul>
      </div>
    </div>

    <div class="am-topbar-right">
      <button class="am-btn am-btn-primary am-topbar-btn am-btn-sm">Login</button>
    </div>
  </div>
</header>
`````

```html
<header class="am-topbar">
  <h1 class="am-topbar-brand">
    <a href="#">Amaze UI</a>
  </h1>

  <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only" data-am-collapse="{target: '#doc-topbar-collapse'}"><span class="am-sr-only">Responsive button</span> <span class="am-icon-bars"></span></button>

  <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse">
    <ul class="am-nav am-nav-pills am-topbar-nav">
      <li class="am-active"><a href="#">Home</a></li>
      <li><a href="#">Project</a></li>
      <li class="am-dropdown" data-am-dropdown>
        <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
          Dropdown <span class="am-icon-caret-down"></span>
        </a>
        <ul class="am-dropdown-content">
          <li class="am-dropdown-header">Title</li>
          <li><a href="#">1. To Moon</a></li>
          <li class="am-active"><a href="#">2. To Mars</a></li>
          <li><a href="#">3. To Earth</a></li>
          <li class="am-disabled"><a href="#">4. To Sun</a></li>
          <li class="am-divider"></li>
          <li><a href="#">5. Go Back</a></li>
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
        <button class="am-btn am-btn-secondary am-topbar-btn am-btn-sm am-dropdown-toggle" data-am-dropdown-toggle>Other <span class="am-icon-caret-down"></span></button>
        <ul class="am-dropdown-content">
          <li><a href="#">Register</a></li>
          <li><a href="#">What you like</a></li>
        </ul>
      </div>
    </div>

    <div class="am-topbar-right">
      <button class="am-btn am-btn-primary am-topbar-btn am-btn-sm">Login</button>
    </div>
  </div>
</header>
```

### Dark Style

Add the `.am-topbar-inverse` class to the container.

`````html
<header class="am-topbar am-topbar-inverse">
  <h1 class="am-topbar-brand">
    <a href="#">Amaze UI</a>
  </h1>

  <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only" data-am-collapse="{target: '#doc-topbar-collapse'}"><span class="am-sr-only">Responsive button</span> <span class="am-icon-bars"></span></button>

  <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse">
    <ul class="am-nav am-nav-pills am-topbar-nav">
      <li class="am-active"><a href="#">Home</a></li>
      <li><a href="#">Project</a></li>
      <li class="am-dropdown" data-am-dropdown>
        <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
          Dropdown <span class="am-icon-caret-down"></span>
        </a>
        <ul class="am-dropdown-content">
          <li class="am-dropdown-header">Title</li>
          <li><a href="#">1. To Moon</a></li>
          <li class="am-active"><a href="#">2. To Mars</a></li>
          <li><a href="#">3. To Earth</a></li>
          <li class="am-disabled"><a href="#">4. To Sun</a></li>
          <li class="am-divider"></li>
          <li><a href="#">5. Go Back</a></li>
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
        <button class="am-btn am-btn-secondary am-topbar-btn am-btn-sm am-dropdown-toggle" data-am-dropdown-toggle>Other <span class="am-icon-caret-down"></span></button>
        <ul class="am-dropdown-content">
          <li><a href="#">Register</a></li>
          <li><a href="#">What you like</a></li>
        </ul>
      </div>
    </div>

    <div class="am-topbar-right">
      <button class="am-btn am-btn-primary am-topbar-btn am-btn-sm">Login</button>
    </div>
  </div>
</header>
`````

```html
<header class="am-topbar am-topbar-inverse">
  ...
</header>
```

### Logo Replacement

The most interesting part of front-end is having a fantastic castle, but using your imagination to create your own little house with all those simple or even boring components.

By changing styles for `.am-text-ir` in [Utility Class](/css/utility), we can easily change the Logo.

`````html
<header class="am-topbar am-topbar-inverse">
  <h1 class="am-topbar-brand">
    <a href="#" class="am-text-ir">Amaze UI</a>
  </h1>

  <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only" data-am-collapse="{target: '#doc-topbar-collapse'}"><span class="am-sr-only">Responsive button</span> <span class="am-icon-bars"></span></button>

  <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse">
    <ul class="am-nav am-nav-pills am-topbar-nav">
      <li class="am-active"><a href="#">Home</a></li>
      <li><a href="#">Project</a></li>
      <li class="am-dropdown" data-am-dropdown>
        <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
          Dropdown <span class="am-icon-caret-down"></span>
        </a>
        <ul class="am-dropdown-content">
          <li class="am-dropdown-header">Title</li>
          <li><a href="#">1. To Moon</a></li>
          <li class="am-active"><a href="#">2. To Mars</a></li>
          <li><a href="#">3. To Earth</a></li>
          <li class="am-disabled"><a href="#">4. To Sun</a></li>
          <li class="am-divider"></li>
          <li><a href="#">5. Go Back</a></li>
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

## Fix to Top/Bottom

### Fix to Top

To fix top bar to the top, add the `.am-topbar-fixed-top` class to `.am-topbar`.

The `body` element in the page with top-fixed topbar will be given a `.am-with-topbar-fixed-top` class. You may modify this class as you wish. Default style is `padding-top: 51px`.

```html
<header class="am-topbar am-topbar-inverse am-topbar-fixed-top">
  <div class="am-container">
    <h1 class="am-topbar-brand">
      <a href="#" class="am-text-ir">Amaze UI</a>
    </h1>

    <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only" data-am-collapse="{target: '#doc-topbar-collapse'}"><span class="am-sr-only">Responsive button</span> <span class="am-icon-bars"></span></button>

    <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse">
      <ul class="am-nav am-nav-pills am-topbar-nav">
        <li class="am-active"><a href="#">Home</a></li>
        <li><a href="#">Project</a></li>
        <li class="am-dropdown" data-am-dropdown>
          <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
            Dropdown <span class="am-icon-caret-down"></span>
          </a>
          <ul class="am-dropdown-content">
            <li class="am-dropdown-header">Title</li>
            <li><a href="#">1. To Moon</a></li>
            <li class="am-active"><a href="#">2. To Mars</a></li>
            <li><a href="#">3. To Earth</a></li>
            <li class="am-disabled"><a href="#">4. To Sun</a></li>
            <li class="am-divider"></li>
            <li><a href="#">5. Go Back</a></li>
          </ul>
        </li>
      </ul>
    </div>
  </div>
</header>
```

### Fix to Bottom

Add the `.am-topbar-fixed-bottom` class to fix topbar to bottom.

The `body` element in the page with top-fixed topbar will be given a `.am-with-topbar-fixed-bottom` class. You may modify this class as you wish. Default style is `padding-bottom: 51px`.

```html
<header class="am-topbar am-topbar-inverse am-topbar-fixed-bottom">
  <div class="am-container">
    <h1 class="am-topbar-brand">
      <a href="#" class="am-text-ir">Amaze UI</a>
    </h1>

    <button class="am-topbar-btn am-topbar-toggle am-btn am-btn-sm am-btn-success am-show-sm-only" data-am-collapse="{target: '#doc-topbar-collapse'}"><span class="am-sr-only">Responsive button</span> <span class="am-icon-bars"></span></button>

    <div class="am-collapse am-topbar-collapse" id="doc-topbar-collapse">
      <ul class="am-nav am-nav-pills am-topbar-nav">
        <li class="am-active"><a href="#">Home</a></li>
        <li><a href="#">Project</a></li>
        <li class="am-dropdown" data-am-dropdown>
          <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
            Dropdown <span class="am-icon-caret-down"></span>
          </a>
          <ul class="am-dropdown-content">
            <li class="am-dropdown-header">Title</li>
            <li><a href="#">1. To Moon</a></li>
            <li class="am-active"><a href="#">2. To Mars</a></li>
            <li><a href="#">3. To Earth</a></li>
            <li class="am-disabled"><a href="#">4. To Sun</a></li>
            <li class="am-divider"></li>
            <li><a href="#">5. Go Back</a></li>
          </ul>
        </li>
      </ul>
    </div>
  </div>
</header>
```



