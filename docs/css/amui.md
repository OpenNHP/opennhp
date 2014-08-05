# AM UI 1.0 Alpha

-----------------

AM UI 是一个轻量级、 [**Mobile first**](http://cbrac.co/113eY5h) 的移动前端框架，
基于主流前端框架（Bootstrap、Foundation、UIKit）编写。

## CSS class 命名规范

AM UI CSS class 命名规范遵循关注分离、松耦合的原则，同时注重易于辨识、理解，参考
[BEM 命名法](http://bem.info/method/definitions/) 的基础上，采用更优雅的书写方式。

下面的代码直观展示了 AM UI CSS class 命名规范。

```css
.am-post           {} /* Block */
.am-post-title     {} /* Element */
.am-post-meta      {} /* Element */
.am-post-sticky    {} /* Generic Modifier - status */
.am-post-active	   {} /* Generic Modifier - status */
.am-post-title-highlight {}  /* Element Modifier */
```

参考文章：[Decoupling HTML From CSS](http://coding.smashingmagazine.com/2012/04/20/decoupling-html-from-css/)

同时为了避免和别的框架命名冲突，AM UI 以 `am` 为命名空间。