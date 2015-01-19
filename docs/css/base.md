# Base
---

Amaze UI 定义的一些基础样式。

## CSS 盒模型

曾几何时，IE 6 及更低版本的___非标准___盒模型被喷得体无完肤。IE 原来的盒模型真的不好么？最终，时间给了另外一个答案。
W3C 终认识到所谓标准盒模型在实际使用中的复杂性，于是在 CSS3 中增加 `box-sizing` 这一属性，允许用户自定义盒模型。

> __You tell me I'm wrong, Then you better prove you're right.__
>
> <small>King of Pop – Scream</small>

这就是 W3C 的证明。

扯远了，Amaze UI 将所有元素的盒模型设置为 `border-box`。这下好了，妈妈再也不用担心你没计算好 `padding`、`border` 而使布局破相了。咱来写样式，不是来学算术。

```css
 *,
 *:before,
 *:after {
   -moz-box-sizing: border-box;
   -webkit-box-sizing: border-box;
   box-sizing: border-box;
 }
```

![Box sizing](/i/docs/box-sizing.png)

参考链接：

- https://developer.mozilla.org/en-US/docs/Web/CSS/box-sizing
- http://www.paulirish.com/2012/box-sizing-border-box-ftw/
- [Box Sizing](http://css-tricks.com/box-sizing/)


## 字号及单位

Amaze UI 将浏览器的基准字号设置为 `62.5%`，也就是 `10px`，现在 `1rem = 10px` —— 为了计算方便。然后在 `body` 上应用了 `font-size: 1.6rem;`，将页面字号设置为 `16px`。

```css
html {
  font-size: 62.5%;
}

body {
  font-size: 1.6rem; /* =16px */
}
```

与 `em` 根据上下文变化不同，`rem` 只跟基准设置关联，只要修改基准字号，所有使用 `rem` 作为单位的设置都会相应改变。

当然，并非所有所有浏览器的默认字号都是 `16px`，所以在不同的浏览器上会有细微差异。

另外，一些需要根据字号做相应变化的场景也使用了 `em`，需要像素级别精确的场景也使用了 `px`。

__参考资源：__

- [FONT SIZING WITH REM](http://snook.ca/archives/html_and_css/font-size-with-rem)
- [Type study: Sizing the legible letter](http://blog.typekit.com/2011/11/09/type-study-sizing-the-legible-letter/)
- [The rem checker](https://offroadcode.com/prototypes/rem-calculator/)
- [Mixins for Rem Font Sizing](http://css-tricks.com/snippets/css/less-mixin-for-rem-font-sizing/)
