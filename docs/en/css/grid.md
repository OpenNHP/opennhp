# Grid
---

Amaze UI use 12 divisions responsive grid system. To use grid, add the `.am-g` class to container and add `.am-u-[sm|md|lg]-[1-12]` to columns. Then setting width for different screens. (`u` represents `unit`. we consulted naming in [Pure CSS](http://purecss.io/grids/)).

Responsive Breakpoints is listed below:

<table class="am-table am-table-bordered am-table-striped">
  <thead>
    <tr>
      <th style="width: 100px">Class</th>
      <th>Interval</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>am-u-sm-*</code></td>
      <td><code>0 - 640px</code></td>
    </tr>
    <tr>
      <td><code>am-u-md-*</code></td>
      <td><code>641px - 1024px</code></td>
    </tr>
    <tr>
      <td><code>am-u-lg-*</code></td>
      <td><code>1025px + </code></td>
    </tr>
  </tbody>
</table>

Amaze UI follows the **Mobile first** concept. __If grid for larger screen is not defined, setting will be inherited from smaller screens__.

<div class="am-alert am-alert-warning">
  <strong>Attention: </strong>The red border in examples is added for effect, which won't exist in your own pages.
  <code>.doc-am-g</code> Class only for Demo. Don't use it in your page.
</div>

## Usage

### Basic

In `<table>`, tables are divided in to rows with `<tr>` tag, and rows are divided into columns with `<td>` tag. Grids work in the similar way:

- **Row** - `.am-g`: Rows in the Grid;
- **Column** - `.am-u-xx-n`: Columns in the grid, where `xx` is the size of viewport, and `n` is the number of divisions the column takes. For example, replacing `n` by `3` means this column takes 3 divisions, namely `1/4` of the row.

**Example 1：A Basic Grid**

In the following example, there are two columns in the row. The first column takes `1/3` of the row, and the second one takes `2/3`. We only use `.am-u-sm-n` here, so it will be applied to all screens.

`````html
<div class="am-g doc-am-g">
  <div class="am-u-sm-4">4</div>
  <div class="am-u-sm-8">8</div>
</div>
`````
```html
<div class="am-g">
  <div class="am-u-sm-4">4</div>
  <div class="am-u-sm-8">8</div>
</div>
```

**Example 2：Responsive Width**

In this example, row will be responsively divided in different ratio on different screens. It's divided evenly in `sm` interval, divided by `1:2` in `md` interval, and divided by `1:3` in `lg` interval.

`````html
<div class="am-g doc-am-g">
  <div class="am-u-sm-6 am-u-md-4 am-u-lg-3">sm-6 md-4 lg-3</div>
  <div class="am-u-sm-6 am-u-md-8 am-u-lg-9">sm-6 md-8 lg-9</div>
</div>
`````
```html
<div class="am-g doc-am-g">
  <div class="am-u-sm-6 am-u-md-4 am-u-lg-3">sm-6 md-4 lg-3</div>
  <div class="am-u-sm-6 am-u-md-8 am-u-lg-9">sm-6 md-8 lg-9</div>
</div>
```

### Max Width of Rows.

In Amaze UI, width of row is `100%`, and max width is not set, so it will scale according to the window.

You may want to give a max width to your rows. Then you can use `.am-g-fixed` class to set the max width to `1000px`. You may also set your own max width.

CSS in source codes of Amaze UI:

```css
.am-g {
  margin: 0 auto;
  width: 100%;
}

.am-g-fixed {
  max-width: 1000px;
}
```

**Example 3：Max Width of Grid** （[View](http://jsbin.com/mamole/)）

Please view this example in a window with width larger than `1000px`.

```html
<h2>Grid without Max Width</h2>
<div class="am-g">
  <div class="am-u-sm-4">4</div>
  <div class="am-u-sm-8">8</div>
</div>

<h2>Grid with Max Width</h2>
<div class="am-g am-g-fixed">
  <div class="am-u-sm-4">4</div>
  <div class="am-u-sm-8">8</div>
</div>
```


Unlike some other frameworks, grid in Amaze UI is **not required to be put inside a container**.

### Row with full width

Some rows may need to be full width, we can make a full-width row with following HTML:

```html
<div class="am-g am-g-fixed">
  <div class="am-u-sm-12">全宽限制最大宽度的行</div>
</div>
```

This is fine, but a little bit redundant. It time to use **container `.am-container`**: 

`.am-container` has the same `padding` with grid columns, so there will be no problem using them together.

**Example 4: Container**

Grid with blue border is `.am-container`.

`````html
<div class="am-container">
  I'm in the .am-container.
</div>

<div class="am-g am-g-fixed doc-am-g">
  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>
</div>

<div class="am-container">
  <div class="am-g doc-am-g">
    <div class="am-u-sm-6">.am-u-sm-6</div>
    <div class="am-u-sm-6">.am-u-sm-6</div>
  </div>
</div>
`````

```html
<!--Only Container-->
<div class="am-container">
  I'm in the .am-container.
</div>

<!--Grid with Fixed Width-->
<div class="am-g am-g-fixed">
  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>
</div>

<!--If you don't hate nesting, you can also do this-->
<div class="am-container">
  <div class="am-g">
    <div class="am-u-sm-6">.am-u-sm-6</div>
    <div class="am-u-sm-6">.am-u-sm-6</div>
  </div>
</div>
```

Styles of `.am-container`：

```css
.am-container {
  -webkit-box-sizing: border-box;
  -moz-box-sizing: border-box;
  box-sizing: border-box;
  margin-left: auto;
  margin-right: auto;
  padding-left: 1rem;
  padding-right: 1rem;
  width: 100%;
  max-width: 1000px;
}

@media only screen and (min-width:641px) {
  .am-container {
    padding-left: 1.5rem;
    padding-right: 1.5rem;
  }
}
```

### Multiple Rows

```html
<div class="am-g am-g-fixed">
  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>
</div>

<div class="am-g am-g-fixed">
  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>
</div>

<div class="am-g am-g-fixed">
  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>
</div>
```

Or

```html
<div class="am-g am-g-fixed">
  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>

  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>

  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>
</div>
```


Both styles are fine, but we recommand the first style because it's easier to control the style of each row.

### Grid with less than 12 divisions

We divided the grid with non-integer percentage, so the sum of width of all columns can possibly not be 100%, which results in the small gap on the right side of the grid, so we let the last column to flow right.

If the total division is smaller than `12`, the `.am-u-end` class should be added to the last column.

**Example 5：Grid with less than 12 divisions**

`````html
<div class="am-g doc-am-g">
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-3">3</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-3 am-u-end">3</div>
</div>
`````
```html
<!-- 未添加 .am-u-end 的情形 -->
<div class="am-g">
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-3">3</div>
</div>

<!-- 添加 .am-u-end 后 -->
<div class="am-g">
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-3 am-u-end">3</div>
</div>
```
<!--
TODO: 使用网站本身做演示（js 添加移除 类名）
-->

**The basic usage of grid ends here**

## Advanced Usage

### Responsive Utility Class

Amaze UI has some responsive utility classes, find more detail in [Utillity Class](/css/utility#%E5%93%8D%E5%BA%94%E5%BC%8F%E8%BE%85%E5%8A%A9?_ver=2.x)。

**Example 6: Use Responsive Utility classes to help show/hide the elements**

Adjust browser window to check the responsive effect.

`````html
<div class="am-g doc-am-g">
  <div class="am-u-lg-3">
    <span class="am-show-md-down">sm-full</span>
    <span class="am-show-lg-only">lg-3</span>
  </div>
  <div class="am-u-lg-6">
    <span class="am-show-md-down">sm-full</span>
    <span class="am-show-lg-only">lg-6</span>
  </div>
  <div class="am-u-lg-3">
    <span class="am-show-md-down">sm-full</span>
    <span class="am-show-lg-only">lg-3</span>
  </div>
</div>

<div class="am-g doc-am-g">
  <div class="am-u-sm-6 am-u-lg-2">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">2</span>
  </div>
  <div class="am-u-sm-6 am-u-lg-8">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">8</span>
  </div>
  <div class="am-u-sm-12 am-u-lg-2">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">2</span>
  </div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-9">9</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-lg-4">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">4</span>
  </div>
  <div class="am-u-lg-8">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">8</span>
  </div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-6 am-u-lg-5">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">5</span>
  </div>
  <div class="am-u-sm-6 am-u-lg-7">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">7</span>
  </div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">6</span>
  </div>
  <div class="am-u-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">6</span>
  </div>
</div>
`````

```html
<div class="am-g">
  <div class="am-u-sm-2 am-u-lg-4">
    <span class="am-show-md-down">sm-2</span>
    <span class="am-show-lg-only">lg-4</span>
  </div>
  <div class="am-u-sm-4 am-u-lg-4">sm4 lg4</div>
  <div class="am-u-sm-6 am-u-lg-4">
    <span class="am-show-md-down">sm-6</span>
    <span class="am-show-lg-only">lg-4</span>
  </div>
</div>

<div class="am-g">
  <div class="am-u-lg-3">
    <span class="am-show-md-down">sm-full</span>
    <span class="am-show-lg-only">lg-3</span>
  </div>
  <div class="am-u-lg-6">
    <span class="am-show-md-down">sm-full</span>
    <span class="am-show-lg-only">lg-6</span>
  </div>
  <div class="am-u-lg-3">
    <span class="am-show-md-down">sm-full</span>
    <span class="am-show-lg-only">lg-3</span>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-6 am-u-lg-2">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">2</span>
  </div>
  <div class="am-u-sm-6 am-u-lg-8">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">8</span>
  </div>
  <div class="am-u-sm-12 am-u-lg-2">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">2</span>
  </div>
</div>
<div class="am-g">
  <div class="am-u-sm-3">3</div>
  <div class="am-u-sm-9">9</div>
</div>
<div class="am-g">
  <div class="am-u-lg-4">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">4</span>
  </div>
  <div class="am-u-lg-8">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">8</span>
  </div>
</div>
<div class="am-g">
  <div class="am-u-sm-6 am-u-lg-5">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">5</span>
  </div>
  <div class="am-u-sm-6 am-u-lg-7">
    <span class="am-show-md-down">6</span>
    <span class="am-hide-md-down">7</span>
  </div>
</div>
<div class="am-g">
  <div class="am-u-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">6</span>
  </div>
  <div class="am-u-lg-6">
    <span class="am-show-md-down">full</span>
    <span class="am-hide-md-down">6</span>
  </div>
</div>
```

### Column Offsetting

Add `am-u-sm-offset-*`, `am-u-md-offset-*` or `am-u-lg-offset-*` classes to set the offsetting.

**Example 7: Column Offsetting**

`````html
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-11">11</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-10 am-u-sm-offset-1">10, offset 1</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-9 am-u-sm-offset-2">9, offset 2</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-8 am-u-sm-offset-3">8, offset 3</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-7 am-u-sm-offset-4">7, offset 4</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-6 am-u-sm-offset-5">6, offset 5</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-5 am-u-sm-offset-6">5, offset 6</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-4 am-u-sm-offset-7">4, offset 7</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-3 am-u-sm-offset-8">3, offset 8</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-2 am-u-sm-offset-9">2, offset 9</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-1 am-u-sm-offset-10">1, offset 10</div>
</div>
`````
```html

<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-11">11</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-10 am-u-sm-offset-1">10, offset 1</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-9 am-u-sm-offset-2">9, offset 2</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-8 am-u-sm-offset-3">8, offset 3</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-7 am-u-sm-offset-4">7, offset 4</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-6 am-u-sm-offset-5">6, offset 5</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-5 am-u-sm-offset-6">5, offset 6</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-4 am-u-sm-offset-7">4, offset 7</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-3 am-u-sm-offset-8">3, offset 8</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-2 am-u-sm-offset-9">2, offset 9</div>
</div>
<div class="am-g">
  <div class="am-u-sm-1">1</div>
  <div class="am-u-sm-1 am-u-sm-offset-10">1, offset 10</div>
</div>
```

### Center Rows

Add the `.am-u-xx-centered` class to center a row: 

- Use `.am-u-sm-centered` to center the row on all devices;
- Use `.am-u-xx-uncentered` to uncenter the row in some interval.

**Example 8：Center**

`````html
<div class="am-g doc-am-g">
  <div class="am-u-sm-3 am-u-sm-centered">3 centered</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-6 am-u-lg-centered">6 centered</div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-9 am-u-sm-centered am-u-lg-uncentered">9 md-down-centered </div>
</div>
<div class="am-g doc-am-g">
  <div class="am-u-sm-11 am-u-sm-centered">11 centered</div>
</div>
`````

```html
<!-- .am-u-sm-centered 始终居中 -->
<div class="am-g">
  <div class="am-u-sm-3 am-u-sm-centered">3 centered</div>
</div>

<!-- .am-u-lg-centered 大于 1024 时居中 -->
<div class="am-g">
  <div class="am-u-sm-6 am-u-lg-centered">6 centered</div>
</div>

<!-- 大于 1024 时不居中 -->
<div class="am-g">
  <div class="am-u-sm-9 am-u-sm-centered am-u-lg-uncentered">9 md-down-centered </div>
</div>

<!-- 始终居中 -->
<div class="am-g">
  <div class="am-u-sm-11 am-u-sm-centered">11 centered</div>
</div>
```

### Order

In consider of the SEO, we sometimes need the HTML to be different from its display. For instance, in a layout with main content and sidebar, the sidebar should follows the main content, while we need the sidebar to be on the left when display. We can use `.am-u-xx-push-*`/`.am-u-xx-pull-*` to do this.

**Example 9：Difference Between HTML and Display**

Adjust the browser to check effect.

`````html
<div class="am-g doc-am-g">
  <div class="am-u-md-8 am-u-md-push-4 am-u-lg-reset-order">8 main</div>
  <div class="am-u-md-4 am-u-md-pull-8 am-u-lg-reset-order">4 sidebar</div>
</div>
`````
```html
<!--
 Using push/pull to push  main to the right and pull the sidebar to the left.
 reset the order in large interval
 -->
<div class="am-g">
  <div class="am-u-md-8 am-u-md-push-4 am-u-lg-reset-order">8 main</div>
  <div class="am-u-md-4 am-u-md-pull-8 am-u-lg-reset-order">4 sidebar</div>
</div>
```

### Remove the Padding

Some users told us that the padding is too large for them, so we made the `.am-g-collapse` class to remove the padding in the columns.

**Example 10: Columns without Padding**

Add the `.am-g-collapse` class to `.am-g` to remove `padding`.

`````html
<div class="am-g am-g-collapse doc-am-g">
  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>
</div>
`````
```html
<div class="am-g am-g-collapse">
  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>
</div>
```
