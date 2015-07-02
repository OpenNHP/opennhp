# Pagination
---

Add the `.am-pagination` class to `<ul>` / `<ol>`, which contains a series of `<li>`.

Add status classes to `<li>`:

- `.am-disabled` - Disabled
- `.am-active` - Activated

## Default Styles

`````html
<ul class="am-pagination">
  <li class="am-disabled"><a href="#">&laquo;</a></li>
  <li class="am-active"><a href="#">1</a></li>
  <li><a href="#">2</a></li>
  <li><a href="#">3</a></li>
  <li><a href="#">4</a></li>
  <li><a href="#">5</a></li>
  <li><a href="#">&raquo;</a></li>
</ul>


<ul class="am-pagination">
  <li><a href="">&laquo; Prev</a></li>
  <li><a href="">Next &raquo;</a></li>
</ul>
`````

```html
<ul class="am-pagination">
  <li class="am-disabled"><a href="#">&laquo;</a></li>
  <li class="am-active"><a href="#">1</a></li>
  <li><a href="#">2</a></li>
  <li><a href="#">3</a></li>
  <li><a href="#">4</a></li>
  <li><a href="#">5</a></li>
  <li><a href="#">&raquo;</a></li>
</ul>

<hr />

<ul class="am-pagination">
  <li><a href="">&laquo; Prev</a></li>
  <li><a href="">Next &raquo;</a></li>
</ul>
```

## Alignment Modifier

Default to be align to left.

### Center Align

Add the `.am-pagination-centered` class to the default styles.

`````html
<ul class="am-pagination am-pagination-centered">
  <li class="am-disabled"><a href="#">&laquo;</a></li>
  <li class="am-active"><a href="#">1</a></li>
  <li><a href="#">2</a></li>
  <li><a href="#">3</a></li>
  <li><a href="#">4</a></li>
  <li><a href="#">5</a></li>
  <li><a href="#">&raquo;</a></li>
</ul>
`````
```html
<ul class="am-pagination am-pagination-centered">
  <li class="am-disabled"><a href="#">&laquo;</a></li>
  <li class="am-active"><a href="#">1</a></li>
  <li><a href="#">2</a></li>
  <li><a href="#">3</a></li>
  <li><a href="#">4</a></li>
  <li><a href="#">5</a></li>
  <li><a href="#">&raquo;</a></li>
</ul>
```

### Right Align

Add the `.am-pagination-right` class to default style.

`````html
<ul class="am-pagination am-pagination-right">
  <li class="am-disabled"><a href="#">&laquo;</a></li>
  <li class="am-active"><a href="#">1</a></li>
  <li><a href="#">2</a></li>
  <li><a href="#">3</a></li>
  <li><a href="#">4</a></li>
  <li><a href="#">5</a></li>
  <li><a href="#">&raquo;</a></li>
</ul>
`````
```html
<ul class="am-pagination am-pagination-right">
  <li class="am-disabled"><a href="#">&laquo;</a></li>
  <li class="am-active"><a href="#">1</a></li>
  <li><a href="#">2</a></li>
  <li><a href="#">3</a></li>
  <li><a href="#">4</a></li>
  <li><a href="#">5</a></li>
  <li><a href="#">&raquo;</a></li>
</ul>
```

## Previous and next

Add the `.am-pagination-prev` or `.am-pagination-next` to a `<li>` element to align previous and next buttons left or right.

`````html
<ul class="am-pagination">
  <li class="am-pagination-prev"><a href="">&laquo; Prev</a></li>
  <li class="am-pagination-next"><a href="">Next &raquo;</a></li>
</ul>
`````

```html
<ul class="am-pagination">
  <li class="am-pagination-prev"><a href="">&laquo; Prev</a></li>
  <li class="am-pagination-next"><a href="">Next &raquo;</a></li>
</ul>
```

## With Icons

Simply replace the texts with icons.

`````html
<ul class="am-pagination">
  <li><a href=""><span class="am-icon-angle-double-left"></span></a></li>
  <li><span>...</span></li>
  <li><a href=""><span class="am-icon-angle-double-right"></span></a></li>
</ul>
`````
```html
<ul class="am-pagination">
  <li><a href=""><span class="am-icon-angle-double-left"></span></a></li>
  <li><span>...</span></li>
  <li><a href=""><span class="am-icon-angle-double-right"></span></a></li>
</ul>
```

__Attention：__ The non-link text in `<li>` should be put inside a `<span>` element.

__Tips:__ Developers using MongoDB should try on [mongoose-paginater](https://www.npmjs.org/package/mongoose-paginater)。
