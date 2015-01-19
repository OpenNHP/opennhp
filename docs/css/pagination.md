# Pagination
---

分页组件，`<ul>` / `<ol>` 添加 `.am-pagination` class， 包含一系列 `<li>`。

在 `<li>` 上添加状态 class：

- `.am-disabled` - 禁用（不可用）
- `.am-active` - 激活

## 基本样式

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

## 对齐方式

默认为左对齐。

### 居中对齐

在默认样式的基础上添加 `.am-pagination-centered` class。

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

### 右对齐

在默认样式的基础上添加 `.am-pagination-right` class。

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

## 左右分布

添加 `.am-pagination-prev` 及 `.am-pagination-next` 到 `<li>`，创建一个仅包含 `上一页` 和 `下一页` 的分页组件。

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

## 结合图标使用

将文字部分用图标替换即可。

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

__注意：__ `<li>` 里面的非链接文字需要使用 `<span>` 包裹。

__Tips:__ 使用 MongoDB 的同学可以试试 [mongoose-paginater](https://www.npmjs.org/package/mongoose-paginater)。
