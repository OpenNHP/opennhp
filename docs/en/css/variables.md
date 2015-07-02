# Variables
---

#### Colors
`````html
<div class="varibles">
  <span class="gray-darker">@gray-darker</span>
  <span class="gray-dark">@gray-dark</span>
  <span class="gray">@gray</span>
  <span class="gray-light">@gray-light</span>
  <span class="gray-lighter">@gray-lighter</span>
  <span class="global-primary">@global-primary</span>
  <span class="global-success">@global-success</span>
  <span class="global-warning">@global-warning</span>
  <span class="global-danger">@global-danger</span>
  <span class="global-info">@global-info</span>
</div>
`````

#### Basic variables (snippet)
```css
@body_bg : #fff;
@text-color : @gray-dark;
@link-color : @global-primary;
@link-hover-color : darken(@link-color, 15%);
@global-contrast-color : #fff;
@global-background : #fff;
@global-border : #ddd;
@component-active-color : #fff;
@component-active-bg : @global-primary;
```