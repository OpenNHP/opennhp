---
id: css
title: CSS
titleEn: CSS
permalink: css.html
next: css/normalize.html
---

# CSS
---

## Overview

CSS of Amaze UI consists of 4 parts.

<div class="am-g">
  <div class="am-u-md-6">
    <div class="am-panel am-panel-default">
      <div class="am-panel-hd">Basic（default）Styles</div>
      <div class="am-panel-bd">
        Use normalize.css to unify the difference of browsers, and include some basic styles for elements.
      </div>
    </div>
  </div>
  <div class="am-u-md-6">
    <div class="am-panel am-panel-default">
      <div class="am-panel-hd">Layout Styles</div>
      <div class="am-panel-bd">
        Include Grid, AVG Grid for layout and some helper classes.
      </div>
    </div>
  </div>

  <div class="am-u-md-6">
    <div class="am-panel am-panel-default">
      <div class="am-panel-hd">Element Styles</div>
      <div class="am-panel-bd">
        Define more styles for <code>code</code>, <code>form</code>, <code>table</code> and other HTML elements.
      </div>
    </div>
  </div>

  <div class="am-u-md-6">
    <div class="am-panel am-panel-default">
      <div class="am-panel-hd">Page Components</div>
      <div class="am-panel-bd">
        Define some useful components combined by multiple elements, such as Pagination and Breadcrumb.
      </div>
    </div>
  </div>
</div>

### Browser Prefix

We removed all standardized browser prefixes in Amaze UI 2.0. Prefixes can be added by using [AutoPrefixer](https://github.com/postcss/autoprefixer) when building project.

Current browser support setting of AutoPrefixer is:

```javascript
[
  'ie >= 8',
  'ie_mob >= 10',
  'ff >= 30',
  'chrome >= 34',
  'safari >= 7',
  'opera >= 23',
  'ios >= 7',
  'android >= 2.3',
  'bb >= 10'
]
```
## Responsive Breakpoints

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Size</th>
    <th class="am-text-nowrap">class abbreviation</th>
    <th>Breakpoints</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
    <tr>
      <td>small</td>
      <td><code>sm</code></td>
      <td>0 - 640px</td>
      <td>For most of phones' portrait and landscape modes.(<a href="http://viewportsizes.com/?filter=Galaxy%20Note%20N7" target="_blank">View port of Galaxy Note 2 is 360 * 640</a>)</td>
    </tr>
    <tr>
      <td>medium</td>
      <td><code>md</code></td>
      <td>641px - 1024px</td>
      <td>For tablets' portrait and landscape modes</td>
    </tr>
    <tr>
      <td>large</td>
      <td><code>lg</code></td>
      <td>1025px + </td>
      <td>For desktops' portrait and landscape modes</td>
    </tr>
  </tbody>
</table>

We currently don't consider screens larger than 1025px. Although screens are getting larger and larger, it could be tiring for users to look around if web page is too wide. 

We have already defined some Media Query variables in "variables.less". If you are using LESS, you may directly use these variables.


```css
@screen:        ~"only screen";

@landscape:     ~"@{screen} and (orientation: landscape)";
@portrait:      ~"@{screen} and (orientation: portrait)";

@small-up:      ~"@{screen}";
@small-only:    ~"@{screen} and (max-width: @{breakpoint-small-max})";

@medium-up:     ~"@{screen} and (min-width:@{breakpoint-medium-min})";
@medium-only:   ~"@{screen} and (min-width:@{breakpoint-medium-min}) and (max-width:@{breakpoint-medium-max})";

@large-up:      ~"@{screen} and (min-width:@{breakpoint-large-min})";
@large-only:    ~"@{screen} and (min-width:@{breakpoint-large-min}) and (max-width:@{breakpoint-large-max})";
```


## Using HTML5

Amaze UI is developed in HTML5, and it's not ensured to work correctly in other doctypes, so please make sure the first line of your HTML is `<!doctype html>`.

HTML head sample:

```html
<!doctype html>
<html class="no-js">
<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="description" content="">
  <meta name="keywords" content="">
  <meta name="viewport"
        content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
  <title>Amaze UI Examples</title>

  <!-- Set render engine for 360 browser -->
  <meta name="renderer" content="webkit">

  <!-- No Baidu Siteapp-->
  <meta http-equiv="Cache-Control" content="no-siteapp"/>

  <link rel="icon" type="image/png" href="{{assets}}i/favicon.png">

  <!-- Add to homescreen for Chrome on Android -->
  <meta name="mobile-web-app-capable" content="yes">
  <link rel="icon" sizes="192x192" href="{{assets}}i/app-icon72x72@2x.png">

  <!-- Add to homescreen for Safari on iOS -->
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="black">
  <meta name="apple-mobile-web-app-title" content="Amaze UI"/>
  <link rel="apple-touch-icon-precomposed" href="{{assets}}i/app-icon72x72@2x.png">

  <!-- Tile icon for Win8 (144x144 + tile color) -->
  <meta name="msapplication-TileImage" content="{{assets}}i/app-icon72x72@2x.png">
  <meta name="msapplication-TileColor" content="#0e90d2">

  <!-- SEO: If your mobile URL is different from the desktop URL, add a canonical link to the desktop page https://developers.google.com/webmasters/smartphone-sites/feature-phones -->
  <!--
  <link rel="canonical" href="http://www.example.com/">
  -->

  <link rel="stylesheet" href="{{assets}}css/amazeui.min.css">
  <link rel="stylesheet" href="{{assets}}css/app.css">
</head>
<body>
...

<!--[if (gte IE 9)|!(IE)]><!-->
<script src="{{assets}}js/jquery.min.js"></script>
<!--<![endif]-->
<!--[if lte IE 8 ]>
<script src="http://libs.baidu.com/jquery/1.11.1/jquery.min.js"></script>
<![endif]-->
</body>
</html>
```

**Reference: **

- [HTML5 Boilerplate](https://github.com/h5bp/html5-boilerplate)
- [Google Web Starter Kit](https://developers.google.com/web/starter-kit/)

## Mobile First

Amaze UI is developed based on the idea of Mobile First, so you should first set `viewport` attribute  in `meta`.

`width=device-width, initial-scale=1` is necessery，and we believe a good design will never require users to scale the view manuelly, so we add `maximum-scale=1, user-scalable=no` to fix the scale.

```html
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
```

## Written For Developers in China

China always has it's own "specialty", even in front-end development. This chapter is only for those who want their page to display correctly in China. If you don't care, just skip it.

### Render Engine

There are a lot of `X cores` browsers in China. Most of the time they play the roles of Troublemakers, while sometimes they may be impressive.

```html
<meta name="renderer" content="webkit">
```

This meta element sets browser to use `webkit` to render the page, but it only works on [360 browser 6.5+](http://se.360.cn/v6/help/meta.html). In this case, I really hope everyone uses 360 browser.

### Anti-Ads

If you don't want to find extra advertisements to be sticked on your web pages, add [setting as below](http://m.baidu.com/pub/help.php?pn=22&ssid=0&from=844b&bd_page_type=1).

```html
<meta http-equiv="Cache-Control" content="no-siteapp" />
```

## Class Naming

### Seperation of Concerns

The naming of Amaze UI CSS classes follows the concepts of seperation of concerns, loose coupling and easily understandable. After consulting [BEM](http://bem.info/method/definitions/), we developed an elegant naming convention.

The following codes directly show the naming convention of Amaze UI CSS classes.

```css
.am-post           {} /* Block */
.am-post-title     {} /* Element */
.am-post-meta      {} /* Element */
.am-post-sticky    {} /* Generic Modifier - status */
.am-post-active	   {} /* Generic Modifier - status */
.am-post-title-highlight {}  /* Element Modifier */
```

#### Example of Seperation of Concerns

```html
<article class="am-post">
  <h2 class="am-post-title"></h2>
  <p class="am-post-meta"></p>
  <main class="am-post-content">
    ...
  </main>
</article>
```

We can simply use following styles to control elements in above codes.

```css
.am-post > h2 {
    ...
}

.am-post > p {
  ...
}
```

It seems to have no problem, and the two selectors won't affect elements in `<main>`, But if we want to use other tags, we will have to modify CSS selectors at the same time. In this way, adding class to corresponding element is a better solution.

```html
<article class="am-post">
  <h1 class="am-post-title"></h1>
  <div class="am-post-meta"></div>
  <main class="am-post-content">
    ...
  </main>
</article>
```

#### anti-patterns of Seperation of Concerns

```html
<ul class="am-nav">
  <li class="am-nav-item"></li>
  <li class="am-nav-item"></li>
  <li class="am-nav-item"></li>
</ul>
```

This is a snippet of a navigation. We add `.am-nav-item` to every `<li>`. This looks like following Seperation of Concerns, while actually it is an anti-pattern, because there must be `<ul>` in `<li>`, and it will be unnecessery to give it an class if there is no other more complicated elements.

Therefore, __Seperation of Concerns is not simply give classes to elements__, we will also need to consider the different situation.


Further Reading：

- [Decoupling HTML From CSS](http://coding.smashingmagazine.com/2012/04/20/decoupling-html-from-css/)

### Classitis

Reading HTML source code full of classes always make developers crazy.

However, divide codes into different classes is always necessery in order to reuse codes and reduce redundant. We can only try to find a balance, and avoid writing unnecessary classes.

### Selectors

While LESS allow us to easily write nested selectors, we suggest developers to avoid using unnecessary nest.

At the mean time, we also believe using too much qualifying selectors is not a good idea.


```css
.ui.form .fields.error .field .ui.selection.dropdown .menu .item:hover {
    ...
}
```

Look at this code snippet from [Semantic UI](http://semantic-ui.com/). A whole line of selectors is so BADASS, but just imagine reading the whole CSS file, just imagine...

__DON'T use more than three levels of nesting and more than two qualifying selectors selectors when you don't have to__.

## WARNING

Everyone is free to use Amaze UI, but there are still something you can't do.

There are two classes represent the state:

- `.am-active` - activited
- `.am-disabled` - disabled

__Don't use them alone or write styles for them__!!!

```css
/* Use them as nesting */
.am-nav .am-active {
  ...
}

/* Use them as qualifying */
.am-btn.am-active {
  ...
}

/* Do NOT use them alone */
.am-active {
  color: red;
}

/* Of course, you may take a try if you just want to have some fun. */
```

## Disable Responsive

Don't like responsive design? Disable it!

- Delete `meta` element about viewport in `head`;

```html
<!--<meta name="viewport"
    content="width=device-width, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no">-->
```

- Fix the width of container `.am-container` (You may build your own class);

```css
.am-container {
  width: 980px !important;
  max-width: none;
}
```

- Only use `.am-u-sm-*` classes when using grid, and remove all classes with other breakpoints.

Now, responsive has been disabled in layout.[Example](/examples/non-responsive.html).

But this is just a start, you may need to do some more adjustment on some components to achieve the best performance.

## About Namespace

<div class="am-alert am-alert-danger">
  It seems like someone is unconfortable with <code>.am</code>, we will have some explanation here.
</div>

Probably you still don't know what is namespace. Just like `yui` in [YUI](https://github.com/yui/yui3/blob/master/build/app-transitions-css/app-transitions-css.css), `pure` in [Pure](http://purecss.io/), `am` is the namespace of Amaze UI.

As long as namespaces make class names much longer, why do we need it?

### Objective: Defending!

Most of CSS styles is based on Classes, we don't want:

- When __Multiple Framework Works together__, HTMLs designed to use our CSS finally use styles from other framework;
- __HTMLs from third parties__ have same classes with us and apply styles from Amaze UI by mistake;
- __When coding their own styles, developers accidentally overwrite styles in Amaze UI;__
- When__multiple developers coorperate together__, styles influence each other because of conflicting naming;__
- __Third party services__(such as share buttens, comment components) can insert styles to pages, which will possibly use structures we provid;
- ......

Amaze UI is used in our company as well as by other developers, and namespace can help us solving these problems effectively. It has nothing to do with advertising our brand.

### So?

<dl>
  <dt>Will namespace be deleted</dt>
  <dd>Normally? No! 100 years later? Probably.</dd>
  <dt>I HATE namespace SOOOOO MUCH. What can I do?</dt>
  <dd>We will try to provide customized namespace in the future. You can also try to remove when compiling. There are many front-end compiling tools. Just take a shot.</dd>
</dl>
