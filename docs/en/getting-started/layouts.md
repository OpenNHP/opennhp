# Layout Samples
---

Amaze UI provides some layout samples for developers. 

---
**Explanation of Icon Display Exception:**

~~In order to avoid `Mixed Content` error, protocol is not added for URL of Icon Font, so icon can't be display correctly using disk path `file:///...`. Please view samples though HTTP service.~~

**Font path has been replaced with local path in the latest download package!**

- Use JetBrais Series(WebStorm etc) to open the sample folder, and then click on the preview button in editor;
- `cd` to the sample folder, `python -m SimpleHTTPServer`；
- or use other HTTP server.

---

## Sample

<div class="am-g">
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/landing.html">
      <img class="am-img-thumbnail"
           src="http://amazeui.org/i/examples/landingPage.png"
           alt="landingPage"/>
      <figcaption class="am-thumbnail-caption">Presenting Page</figcaption>
    </a>
  </div>
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/login.html">
      <img class="am-img-thumbnail"
           src="http://amazeui.org/i/examples/loginPage.png" alt="loginPage"/>
      <figcaption class="am-thumbnail-caption">Login/Register Page</figcaption>
    </a>
  </div>
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/blog.html">
      <img class="am-img-thumbnail"
           src="http://amazeui.org/i/examples/blogPage.png" alt="blogPage"/>
      <figcaption class="am-thumbnail-caption">Blog Page</figcaption>
    </a>
  </div>
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/sidebar.html">
      <img class="am-img-thumbnail"
           src="http://amazeui.org/i/examples/sidebarPage.png"
           alt="sidebarPage"/>
      <figcaption class="am-thumbnail-caption">Sidebar Page</figcaption>
    </a>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-6 am-u-md-3">
    <a class="am-thumbnail" href="/examples/admin-index.html">
      <img class="am-img-thumbnail" src="http://ww1.sinaimg.cn/large/005yyi5Jjw1elpr8svtoyj30j70demxe.jpg
" alt="adminPage"/>
      <figcaption class="am-thumbnail-caption">Administration Page</figcaption>
    </a>
  </div>
</div>

### Presenting Page

Components:

<ul>
  <li>CSS: <a class="am-badge am-badge-success" href="/css/grid">Grid</a> <a
    class="am-badge am-badge-success" href="/css/icon">Icon</a> <a
    class="am-badge am-badge-success" href="/css/button">Button</a> <a
    class="am-badge am-badge-success" href="/css/form">Form</a> <a
    class="am-badge am-badge-success" href="/css/article">Article</a> <a
    class="am-badge am-badge-success" href="/css/topbar">Topbar</a> <a
    class="am-badge am-badge-success" href="/css/utility">Utility</a></li>
  <li>JS plugins: <a class="am-badge am-badge-primary" href="/javascript/dropdown">Dropdown</a>
    <a class="am-badge am-badge-primary" href="/javascript/scrollspy">Scrollspy</a>
  </li>
</ul>

### Login/Register Page

Components：

<ul>
  <li>CSS: <a class="am-badge am-badge-success" href="/css/grid">Grid</a>
    <a class="am-badge am-badge-success" href="/css/icon">Icon</a>
    <a class="am-badge am-badge-success" href="/css/button">Button</a>
    <a class="am-badge am-badge-success" href="/css/button-group">Button-group</a>
    <a class="am-badge am-badge-success" href="/css/form">Form</a></li>
</ul>

### Blog Page

Components：

<ul>
  <li>CSS: <a class="am-badge am-badge-success" href="/css/grid">Grid</a> <a
    class="am-badge am-badge-success" href="/css/block-grid">Block Grid</a> <a
    class="am-badge am-badge-success" href="/css/icon">Icon</a> <a
    class="am-badge am-badge-success" href="/css/button">Button</a> <a
    class="am-badge am-badge-success" href="/css/panel">Panel</a> <a
    class="am-badge am-badge-success" href="/css/list">List</a> <a
    class="am-badge am-badge-success" href="/css/pagination">Pagination</a></li>
  <li>JS plugins: <a class="am-badge am-badge-primary" href="/css/dropdown">dropdown</a>
  </li>
</ul>

### Blog sidebar

Components：

<ul>
    <li>CSS: <a class="am-badge am-badge-success" href="/css/grid">Grid</a> <a class="am-badge am-badge-success" href="/css/article">Article</a> <a class="am-badge am-badge-success" href="/css/comment">Comment</a> <a class="am-badge am-badge-success" href="/css/button">Button</a> <a class="am-badge am-badge-success" href="/css/icon">Icon</a> <a class="am-badge am-badge-success" href="/css/list">List</a> <a class="am-badge am-badge-success" href="/css/utility">Utility</a></li>
    <li>JS: <a class="am-badge am-badge-primary" href="/javascript/offcanvas">Offcanvas</a></li>
</ul>

### Administration Page

- [index page](/examples/admin-index.html)
- [user page](/examples/admin-user.html)
- [gallery page](/examples/admin-gallery.html)
- [table page](/examples/admin-table.html)
- [form page](/examples/admin-form.html)
- [help page](/examples/admin-help.html)
- [log page](/examples/admin-log.html)
- [404 page](/examples/admin-404.html)

##Don't like responsive design? Disable it!

- Delete `meta` element about viewport in `head`;

```html
<!--<meta name="viewport"
    content="width=device-width, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no">-->
```

- Fix the width of container `.am-container` (You may build your own class);

```css
.am-container {
  width: 980px !important;
  max-width: none;
}
```

- Only use `.am-u-sm-*` classes when using grid, and remove all classes with other breakpoints.

Now, responsive has been disabled in layout.[Example](/examples/non-responsive.html).

But this is just a start, you may need to do some more adjustment on some components to achieve the best performance.
