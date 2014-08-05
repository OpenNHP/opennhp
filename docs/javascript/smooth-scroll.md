# Smooth Scroll
---

基于 Zepto 的平滑滚动插件，源自 [Zepto 作者](https://gist.github.com/madrobby/8507960#file-scrolltotop-annotated-js)。


## 使用演示

`````html
<button id="doc-scroll-to-btm" class="am-btn am-btn-primary">滚动到底部</button>
<button data-am-smooth-scroll class="am-btn am-btn-success">滚动到顶部</button>
<script>
  $('#doc-scroll-to-btm').on('click', function() {
    var $w = $(window);
    $w.smoothScroll($(document).height() - $w.height());
  });
</script>
`````

## 使用方法

### 通过 Data API
    
在元素上添加 `data-am-smooth-scroll` 属性。

```html
<button data-am-smooth-scroll class="am-btn am-btn-success">滚动到顶部</button>
```
    
如果要指定滚动的位置，可以给这个属性设一个值。
    
`````html
<button data-am-smooth-scroll="189" class="am-btn am-btn-secondary">滚动到滚动条距离顶部 189px 的位置</button>
`````
```html
<button data-am-smooth-scroll="189" class="am-btn am-btn-secondary">...</button>
```
### 通过 Javascript

#### 方法

```javascript
$(window).smoothScroll([position])
```

`position` 为可选参数，默认为 `0`，即滚动到顶部。

```javascript
// 滚动到底部
$('#my-button').on('click', function() {
  var $w = $(window);
  $w.smoothScroll($(document).height() - $w.height());
});
```
