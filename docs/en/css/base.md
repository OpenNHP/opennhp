# Base
---

Some basic styles defined in Amaze UI.

## CSS Box Model

IE's non-standard box model used to make web developers suffer so much pain. However, is IE's box really that terrible? Eventually, the answer comes to be different from what we thought. W3C finally realize the standard box model is too complicated to use, so they add `box-sizing` in CSS3, which allow users to customize the box model.

> __You tell me I'm wrong, Then you better prove you're right.__
>
> <small>King of Pop – Scream</small>

This is the proof given by W3C.

In this way, Amaze UI set box model of all elements to be `border-box`. Now we will never need to worry about calculating `padding` and `border`.

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

Reference: 

- https://developer.mozilla.org/en-US/docs/Web/CSS/box-sizing
- http://www.paulirish.com/2012/box-sizing-border-box-ftw/
- [Box Sizing](http://css-tricks.com/box-sizing/)


## Font Size and Units

Amaze UI set the base font size to be `62.5%`, which is `10px`, so now we have `1rem = 10px`. Then we use `font-size: 1.6rem;` for `body` to set the font size to be `16px`.

```css
html {
  font-size: 62.5%;
}

body {
  font-size: 1.6rem; /* =16px */
}
```

Unlike `em` which changes according to font size in parent element, `rem` is only related to base font size. If the base font size is changed, all setting using `rem` as unit will change.

Of course, not all browsers use `16px` as their default font size, so the font size can be slightly different between different browsers.

What's more, sometimes in different scenario font size need to be adjusted accordingly, so we used `em` for them, and we used `px` when size is required to be precise in pixel level.

__Reference: __

- [FONT SIZING WITH REM](http://snook.ca/archives/html_and_css/font-size-with-rem)
- [Type study: Sizing the legible letter](http://blog.typekit.com/2011/11/09/type-study-sizing-the-legible-letter/)
- [The rem checker](https://offroadcode.com/prototypes/rem-calculator/)
- [Mixins for Rem Font Sizing](http://css-tricks.com/snippets/css/less-mixin-for-rem-font-sizing/)
