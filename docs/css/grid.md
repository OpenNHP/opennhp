# Grid
---

Amaze UI 使用了 12 列的响应式网格系统。使用时需在外围容器上添加 `.am-g` class，在列上添加 `.col-[sm|md|lg]-[1-12]` class，然后根据不同的屏幕需求设置不同的宽度。

响应式断点如下：

<table class="am-table am-table-bd am-table-striped">
  <thead>
    <tr>
      <th style="width: 100px">Class</th>
      <th>区间</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>col-sm-*</code></td>
      <td><code>0 - 640px</code></td>
    </tr>
    <tr>
      <td><code>col-md-*</code></td>
      <td><code>641px - 1024px</code></td>
    </tr>
    <tr>
      <td><code>col-lg-*</code></td>
      <td><code>1025px + </code></td>
    </tr>
  </tbody>
</table>

Amaze UI 以 **移动优先** 的理念开发， __如果不设置大屏的网格，应用到小屏幕的样式会继承到更大的屏幕上__。

___注意：为了方便查看效果，演示中的网格加了红色边框，实际中没有。___

## 流式布局

### `.am-g` 未限定宽度

`.am-g` 的宽度被设置为 `100%`，会随着窗口自动缩放。

可以在网格容器上添加 `.am-g-fixed` class，将最大宽度限制为 `1000px`。

```css
.am-g {
  margin: 0 auto;
  width: 100%;
}

.am-g-fixed {
  max-width: 1000px;
}
```
### 容器 `.am-container`

Amaze UI 曾尝试直接使用 `.am-g` 作为容器，但在有些场景不太适用，因此又加了一个容器 `.am-container` 辅助 class。

虽然现在的显示器越来越大，但出于用户体验的考虑，容器的最大宽度限定为 `1000px`。

```css
.am-container {
  margin-left: auto;
  margin-right: auto;
  width: 100%;
  max-width: 1000px;
}
```

<!--
TODO: 使用网站本身做演示（js 添加移除 类名）
-->

## 基本使用

根据需求设置不同视口宽度下的布局。调整浏览器窗口以查看响应式效果。

`````html
<div class="am-g doc-am-g">
  <div class="col-sm-2 col-lg-4">
    <span class="am-show-md-down">2</span>
    <span class="am-show-lg-only">4</span>
  </div>
  <div class="col-sm-4 col-lg-4">4</div>
  <div class="col-sm-6 col-lg-4">
    <span class="am-show-md-down">6</span>
    <span class="am-show-lg-only">4</span>
  </div>
</div>

<div class="am-g doc-am-g">
  <div class="col-lg-3">
    <span class="am-show-md-down">full</span>
    <span class="am-show-lg-only">3</span>
  </div>
  <div class="col-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-show-lg-only">6</span>
  </div>
  <div class="col-lg-3">
    <span class="am-show-md-down">full</span>
    <span class="am-show-lg-only">3</span>
  </div>
</div>

<div class="am-g doc-am-g">
  <div class="col-sm-6 col-lg-2">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">2</span>
  </div>
  <div class="col-sm-6 col-lg-8">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">8</span>
  </div>
  <div class="col-sm-12 col-lg-2">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">2</span>
  </div>
</div>
<div class="am-g doc-am-g">
  <div class="col-sm-3">3</div>
  <div class="col-sm-9">9</div>
</div>
<div class="am-g doc-am-g">
  <div class="col-lg-4">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">4</span>
  </div>
  <div class="col-lg-8">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">8</span>
  </div>
</div>
<div class="am-g doc-am-g">
  <div class="col-sm-6 col-lg-5">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">5</span>
  </div>
  <div class="col-sm-6 col-lg-7">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">7</span>
  </div>
</div>
<div class="am-g doc-am-g">
  <div class="col-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">6</span>
  </div>
  <div class="col-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">6</span>
  </div>
</div>
`````

```html
<div class="am-g">
  <div class="col-sm-2 col-lg-4">
    <span class="am-show-md-down">2</span>
    <span class="am-show-lg-only">4</span>
  </div>
  <div class="col-sm-4 col-lg-4">4</div>
  <div class="col-sm-6 col-lg-4">
    <span class="am-show-md-down">6</span>
    <span class="am-show-lg-only">4</span>
  </div>
</div>

<div class="am-g">
  <div class="col-lg-3">
    <span class="am-show-md-down">full</span>
    <span class="am-show-lg-only">3</span>
  </div>
  <div class="col-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-show-lg-only">6</span>
  </div>
  <div class="col-lg-3">
    <span class="am-show-md-down">full</span>
    <span class="am-show-lg-only">3</span>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-6 col-lg-2">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">2</span>
  </div>
  <div class="col-sm-6 col-lg-8">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">8</span>
  </div>
  <div class="col-sm-12 col-lg-2">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">2</span>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-3">3</div>
  <div class="col-sm-9">9</div>
</div>

<div class="am-g">
  <div class="col-lg-4">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">4</span>
  </div>
  <div class="col-lg-8">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">8</span>
  </div>
</div>

<div class="am-g">
  <div class="col-sm-6 col-lg-5">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">5</span>
  </div>
    <div class="col-sm-6 col-lg-7">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">7</span>
    </div>
</div>

<div class="am-g">
  <div class="col-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">6</span>
  </div>
  <div class="col-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">6</span>
  </div>
</div>
```

## 列边距

添加 `col-sm-offset-*`、`col-md-offset-*`、`col-lg-offset-*` 设置列的左边距。

`````html
<div class="am-g doc-am-g">
  <div class="col-sm-1">1</div>
  <div class="col-sm-11">11</div>
</div>
<div class="am-g doc-am-g">
  <div class="col-sm-1">1</div>
  <div class="col-sm-10 col-sm-offset-1">10, offset 1</div>
</div>
<div class="am-g doc-am-g">
  <div class="col-sm-1">1</div>
  <div class="col-sm-9 col-sm-offset-2">9, offset 2</div>
</div>
<div class="am-g doc-am-g">
  <div class="col-sm-1">1</div>
  <div class="col-sm-8 col-sm-offset-3">8, offset 3</div>
</div>
`````
```html
<div class="am-g">
  <div class="col-sm-1">1</div>
  <div class="col-sm-11">11</div>
</div>
<div class="am-g">
  <div class="col-sm-1">1</div>
  <div class="col-sm-10 col-sm-offset-1">10, offset 1</div>
</div>
<div class="am-g">
  <div class="col-sm-1">1</div>
  <div class="col-sm-9 col-sm-offset-2">9, offset 2</div>
</div>
<div class="am-g">
  <div class="col-sm-1">1</div>
  <div class="col-sm-8 col-sm-offset-3">8, offset 3</div>
</div>
```

## 不足 12 列的网格

网格拆分时使用了非整数的百分比（100/12 * 占的网格数），浏览器在计算的时候会有一些差异，最终所有列的宽度和可能为达到 100%。因此，最后一列会被向右浮动，以填满行的左右两边（如下面示例中的第一行）。

实际使用中，如果行的网格数不足 `12`，需要在最后一列上添加 `.col-end`。

`````html
<div class="am-g doc-am-g">
  <div class="col-sm-3">3</div>
  <div class="col-sm-3">3</div>
  <div class="col-sm-3">3</div>
</div>
<div class="am-g doc-am-g">
  <div class="col-sm-3">3</div>
  <div class="col-sm-3">3</div>
  <div class="col-sm-3 col-end">3</div>
</div>
`````
```html
<!-- 未添加 .col-end 的情形 -->
<div class="am-g">
  <div class="col-sm-3">3</div>
  <div class="col-sm-3">3</div>
  <div class="col-sm-3">3</div>
</div>

<!-- 添加 .col-end 后 -->
<div class="am-g">
  <div class="col-sm-3">3</div>
  <div class="col-sm-3">3</div>
  <div class="col-sm-3 col-end">3</div>
</div>
```

## 居中的列

添加 `.col-xx-centered` 实现列居中：

- 如果始终的设为居中，只需要添加 `.col-sm-centered`（移动优先，继承）；
- 某些区间不居中添加， `.col-xx-uncentered`。

`````html
<div class="am-g doc-am-g">
  <div class="col-sm-3 col-sm-centered">3 centered</div>
</div>
<div class="am-g doc-am-g">
  <div class="col-sm-6 col-lg-centered">6 centered</div>
</div>
<div class="am-g doc-am-g">
  <div class="col-sm-9 col-sm-centered col-lg-uncentered">9 centered</div>
</div>
<div class="am-g doc-am-g">
  <div class="col-sm-11 col-sm-centered columns">11 centered</div>
</div>
`````

```html
<!-- .col-sm-centered 始终居中 -->
<div class="am-g">
  <div class="col-sm-3 col-sm-centered">3 centered</div>
</div>

<!-- .col-lg-centered 大于 1024 时居中 -->
<div class="am-g">
  <div class="col-sm-6 col-lg-centered">6 centered</div>
</div>

<!-- 大于 1024 时不居中 -->
<div class="am-g">
  <div class="col-sm-9 col-sm-centered col-lg-uncentered">9 centered</div>
</div>

<!-- 始终居中 -->
<div class="am-g">
  <div class="col-sm-11 col-sm-centered">11 centered</div>
</div>
```

## 列排序

出于 SEO 考虑，有时会有一些结构和表现不一致的情况，比如一个主要内容 + 侧边栏的布局，结构中主要内容在前、侧边栏在后，但表现中需要把边栏放在左边，主要内容放在右边，通过 `.col-xx-push-*` / `.col-xx-pull-*` 来实现。

`````html
<div class="am-g doc-am-g">
  <div class="col-md-8 col-md-push-4 col-lg-reset-order">8 main</div>
  <div class="col-md-4 col-md-pull-8 col-lg-reset-order">4 sidebar</div>
</div>
`````
```html
<!--
 结构中 main 在前， sidebar 在后
 通过 push/pull，在 medium 区间将 sidebar 显示到左侧，main 显示到右侧
 large 区间 reset 回结构排序
 -->
<div class="am-g">
  <div class="col-md-8 col-md-push-4 col-lg-reset-order">8 main</div>
  <div class="col-md-4 col-md-pull-8 col-lg-reset-order">4 sidebar</div>
</div>
```