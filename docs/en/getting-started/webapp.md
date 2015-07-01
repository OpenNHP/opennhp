# About Web App
---

## Touch Icon

We always need to set an Touch Icon when we are developing a website. Because of the difference of systems, there is no common standard for adding Touch Icon. IOS is the system that support Touch Icon earliest and has clear standard, so other systems partly support IOS's standard.

### Solution

This is a way to support both iOS 4.2+ and Android 2.1+:

```html
<link rel="apple-touch-icon-precomposed" href="/path/to/icon72x72@2x.png">
```
- `rel="apple-touch-icon-precomposed"`：No extra effects. Works on Android 1.5 - 2.1.
- Icon size：144px * 144px, works on iPhone, iPad and most of Android devices.

Read following contents for more details.

### iOS

```html
<!-- Add to homescreen for Safari on iOS. Hide address bar and status bar if enter from homescreen. Full screen (content="yes") -->
<meta name="apple-mobile-web-app-capable" content="yes">

<!-- Color of the system headbar(content = black、white 和 black-translucent) -->
<meta name="apple-mobile-web-app-status-bar-style" content="black">
<meta name="apple-mobile-web-app-status-bar-style" content="white">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">

<!-- Setting title -->
<meta name="apple-mobile-web-app-title" content="Web Starter Kit">

<!-- Setting Touch Icon, PNG is recommended -->
<link rel="apple-touch-icon" href="touch-icon-iphone.png">
<link rel="apple-touch-icon" sizes="76x76" href="touch-icon-ipad.png">
<link rel="apple-touch-icon" sizes="120x120" href="touch-icon-iphone-retina.png">
<link rel="apple-touch-icon" sizes="152x152" href="touch-icon-ipad-retina.png">

<!--
 Using rel="apple-touch-icon" will add round corner and highlight to icon;
 Using rel="apple-touch-icon-precomposed" will use the original icon.
-->
```
Size：

<table class="am-table am-table-bd am-table-striped" style="text-align: center;">
  <thead style="text-align: center;">
  <tr>
    <th style="width:100px; text-align: center">Device</th>
    <th>iPhone 5 and iPod touch (High Resolution)</th>
    <th>iPhone and iPod touch (High Resolution)</th>
    <th>iPad and iPad mini (High Resolution)</th>
    <th>iPad 2 and iPad mini (Normal Resolution)</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td>Size (px)</td>
    <td>120 x 120</td>
    <td>120 x 120</td>
    <td>152 x 152</td>
    <td>76 x 76</td>
  </tr>
  </tbody>
</table>

__Reference: __

- [iOS Human Interface Guidelines - Icon and Image Sizes
](https://developer.apple.com/library/ios/documentation/UserExperience/Conceptual/MobileHIG/IconMatrix.html)
- [Everything you always wanted to know about touch icons](http://mathiasbynens.be/notes/touch-icons)
- [Configuring Web Applications](https://developer.apple.com/library/ios/documentation/AppleApplications/Reference/SafariWebContent/ConfiguringWebApplications/ConfiguringWebApplications.html)

### Android Chrome

Only Chrome 31 support Add to Homescreen in Android.

```html
<!-- Add to homescreen for Chrome on Android -->
<meta name="mobile-web-app-capable" content="yes">
<link rel="icon" sizes="196x196" href="images/touch/chrome-touch-icon-196x196.png">
```

__Reference: __

- [Chrome Add to Homescreen](https://developer.chrome.com/multidevice/android/installtohomescreen)

### Win 8

```html
<!-- Tile icon for Win8 (144x144 + tile color) -->
<!-- win 8 Tile icon -->
<meta name="msapplication-TileImage" content="images/touch/ms-touch-icon-144x144-precomposed.png">
<!-- win 8 Tile color -->
<meta name="msapplication-TileColor" content="#3372DF">
```

__Reference: __

- [Pinned Sites][pinedsites]
- [MSDN - Pinned site metadata reference][msdn-pin]

[pinedsites]:http://msdn.microsoft.com/en-us/library/ie/hh772707(v=vs.85).aspx
[msdn-pin]:http://msdn.microsoft.com/zh-cn/library/ie/dn255024(v=vs.85).aspx
