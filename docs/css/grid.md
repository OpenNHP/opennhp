# Grid
---

Amaze UI 使用了 `12` 列的响应式网格系统。使用时需在外围容器上添加 `.am-g` class，在列上添加 `.am-u-[sm|md|lg]-[1-12]` class，然后根据不同的屏幕需求设置不同的宽度（`u` 可以理解为 `unit`， 比较贴合网格数字的含义，参考了 [Pure CSS](http://purecss.io/grids/) 的网格命名。）。

响应式断点如下：

<table class="am-table am-table-bordered am-table-striped">
  <thead>
    <tr>
      <th style="width: 100px">Class</th>
      <th>区间</th>
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

Amaze UI 以 **移动优先** 的理念开发， __如果不设置大屏的网格，应用到小屏幕的样式会继承到更大的屏幕上__。

<div class="am-alert am-alert-warning">
  <strong>注意：</strong>为了方便查看效果，演示中的网格加了红色边框，实际中没有。
  <code>.doc-am-g</code> 为辅助 Demo 显示添加的 class，实际使用时不需要。
</div>

## 基本使用

### 基本概念

在 `<table>` 中，行用 `<tr>` 划分，列用 `<td>` 划分，行和列组合在一起形成行，网格中也类似：

- **行** - `.am-g`: 网格中的行，用于包裹列，清除列的浮动；
- **列** - `.am-u-xx-n`: 网格中的列，`xx` 为视口区间，`n` 为该列所占的份数，如 `n` 为 `3` 时表示这一列占整行宽度的 `3/12`，即 `1/4`。

**示例 1：一个基本的网格**

下面的示例中，行包含两列，第一列占 `4` 份，第二列占 `8` 份，我们只设置了 `.am-u-sm-n` 这个 class，意味着无论视口多大，都保持这个比例的划分。

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

**示例 2：不同区间不同的划分比例**

下面的示例中，`sm` 区间两列是等分的，`md` 区间为 `1:2` 划分，`lg` 区间为 `1:3`。

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

### 限制行的宽度

Amaze UI 中， 行 `.am-g` 的宽度被设置为 `100%`， 未限定最大宽度，会随着窗口自动缩放。

可以在行上添加 `.am-g-fixed` class，将最大宽度限制为 `1000px`（虽然显示器分辨率越来越高，但基于用户体验考虑，仍然选择这个值），也可以根据自己的需求设置一个最大宽度限制。

源代码中的相关 CSS 为：

```css
.am-g {
  margin: 0 auto;
  width: 100%;
}

.am-g-fixed {
  max-width: 1000px;
}
```

**示例 3：限制宽度的网格** （[查看演示](http://jsbin.com/mamole/)）

请在宽度大于 `1000px` 中的窗口中查看。

```html
<h2>没有限制宽度的网格</h2>
<div class="am-g">
  <div class="am-u-sm-4">4</div>
  <div class="am-u-sm-8">8</div>
</div>

<h2>限制宽度的网格</h2>
<div class="am-g am-g-fixed">
  <div class="am-u-sm-4">4</div>
  <div class="am-u-sm-8">8</div>
</div>
```

通过 `.am-g` + `.am-g-fixed` 限制行的宽度，**网格并不需要容器**，这可能和某些框架不太一样。

### 全宽的行

有时某些可能是全宽的，按照上面的逻辑，应该使用下面的代码来实现：

```html
<div class="am-g am-g-fixed">
  <div class="am-u-sm-12">全宽限制最大宽度的行</div>
</div>
```

是的，这样用也没问题，不过不觉得有点冗余么？这时候**容器 `.am-container`**就派上用场了：

`.am-container` 和网格列设置了相同的左右 `padding`，可以和网格内容对齐。

**示例 4：容器**

蓝色边框的是 `.am-container`。

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
<!--没有使用网格的内容-->
<div class="am-container">
  I'm in the .am-container.
</div>

<!--网格行限制宽度-->
<div class="am-g am-g-fixed">
  <div class="am-u-sm-6">.am-u-sm-6</div>
  <div class="am-u-sm-6">.am-u-sm-6</div>
</div>

<!--当然，如果你觉得嵌套一层也无所谓的话，这样用也可以-->
<div class="am-container">
  <div class="am-g">
    <div class="am-u-sm-6">.am-u-sm-6</div>
    <div class="am-u-sm-6">.am-u-sm-6</div>
  </div>
</div>
```

`.am-container` 的样式为：

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

### 多行网格

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

或者

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

上面两种写法的都没有问题，但出于对每行样式控制的方便程度考虑，推荐第一种。

### 不足 12 份的网格

网格拆分时使用了非整数的百分比（100/12 * n），浏览器在计算的时候会有一些差异，最终所有列的宽度和可能没有达到 100%，导致网格右侧会有很小的空隙。因此，向右浮动最后一列，以填满行的右边（如下面示例中的第一行）。

实际使用中，如果行的网格数不足 `12`，需要在最后一列上添加 `.am-u-end`。

**示例 5：不足 12 份的网格**

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

**至此，网格的基本使用就介绍完了。**如果已经满足需求，那施主就请回吧，该摇一摇，或撸啊撸，请自便。

如非要往下看，出现头昏眼花等各种不适，后果自负，请自备郝思嘉速效救心丸……

## 进阶使用

### 响应式辅助 Class

Amaze UI 中内置了一些辅助响应式的 class，详见[辅助类-辅助 Class](/css/utility#%E5%93%8D%E5%BA%94%E5%BC%8F%E8%BE%85%E5%8A%A9?_ver=2.x)。

**示例 6：响应式辅助类控制元素显隐**

调整浏览器窗口以查看响应式效果。

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

### 列边距

添加 `am-u-sm-offset-*`、`am-u-md-offset-*`、`am-u-lg-offset-*` 设置列的左边距。

**示例 7：列边距啪啪啪**

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

### 居中的列

添加 `.am-u-xx-centered` 实现列居中：

- 如果始终的设为居中，只需要添加 `.am-u-sm-centered`（移动优先，继承）；
- 某些区间不居中添加， `.am-u-xx-uncentered`。

**示例 8：居中， To be or not to be**

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

### 列排序

出于 SEO 考虑，有时会有一些结构和表现不一致的情况，比如一个主要内容 + 边栏的布局，结构中主要内容在前、边栏在后，但表现中需要把边栏放在左边，主要内容放在右边，可以通过 `.am-u-xx-push-*` / `.am-u-xx-pull-*` 来实现。

**示例 9：结构与表现表里不一**

改变浏览器窗口宽度查看效果。

`````html
<div class="am-g doc-am-g">
  <div class="am-u-md-8 am-u-md-push-4 am-u-lg-reset-order">8 main</div>
  <div class="am-u-md-4 am-u-md-pull-8 am-u-lg-reset-order">4 sidebar</div>
</div>
`````
```html
<!--
 结构中 main 在前， sidebar 在后
 通过 push/pull，在 medium 区间将 sidebar 显示到左侧，main 显示到右侧
 large 区间 reset 回结构排序
 -->
<div class="am-g">
  <div class="am-u-md-8 am-u-md-push-4 am-u-lg-reset-order">8 main</div>
  <div class="am-u-md-4 am-u-md-pull-8 am-u-lg-reset-order">4 sidebar</div>
</div>
```

### 移除列内边距

有同学提出列默认的内边距在某些场景太大，这时 `.am-g-collapse` 就派上用场了。

**示例 10: 没有内边距的列**

在行 `.am-g` 上添加 `.am-g-collapse`，移除列的内边距（`padding`）。

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
