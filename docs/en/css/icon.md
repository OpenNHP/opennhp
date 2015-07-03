# Icon
---

Amaze UI Icon is using [Font Awesome](http://fontawesome.io/icons/)(Upgraded to 4.3 in Amaze UI 2.2.0), which contains almost all commonly used icons.

## Usage

### Adding Class

Add `am-icon-{icon-name}` class to HTML elements.

`````html
<span class="am-icon-qq"> QQ</span>
<span class="am-icon-weixin"> Wechat</span>

`````

```html
<span class="am-icon-qq"> QQ</span>
<span class="am-icon-weixin"> Wechat</span>
```
### Use Mixin

__LESS Users__ can add icons using mixin:

1. Use `.am-icon-font` in the element to set the font;
2. Set `content` to be the variable of the icon you want to use, like `content: @fa-var-{icon name}`。

`````html
<span class="doc-icon-custom"> Weibo</span>
`````

```html
<span class="doc-icon-custom"> Weibo</span>
```

```css
.doc-icon-custom {
  &:before {
    .am-icon-font;
    content: @fa-var-weibo;
  }
}
```

### Modify Font Path

~~We currently use font file from [Staticfile CDN](http://staticfile.org/) (suppot HTTPS), but you can use your own file; ~~**The CDN has been replaced by local file in the compiled CSS.**

- **Using LESS**: Set `@fa-font-path` and replace the default value. For example, set `@fa-font-path: "../fonts";`. This variable is defined in `icon.less`.
- **Using CSS**： Find and replace `//dn-staticfile.qbox.me/font-awesome/4.2.0/fonts/`。

## Size

- `.am-icon-sm`: Scale up 150%
- `.am-icon-md`: Scale up 200%
- `.am-icon-lg`: Scale up 250%

`````html
<p><span class="am-icon-home"></span> Default Size </p>
<p><span class="am-icon-home am-icon-sm"></span> .am-icon-sm</p>
<p><span class="am-icon-home am-icon-md"></span> .am-icon-md</p>
<p><span class="am-icon-home am-icon-lg"></span> .am-icon-lg</p>
`````
```html
<p><span class="am-icon-home"></span> Default Size</p>
<p><span class="am-icon-home am-icon-sm"></span> .am-icon-sm</p>
<p><span class="am-icon-home am-icon-md"></span> .am-icon-md</p>
<p><span class="am-icon-home am-icon-lg"></span> .am-icon-lg</p>
```

## Icon button

Add the `.am-btn-icon` class to Icon to make it an Icon button.

The color of button can be changed by add following classes:

- `.am-primary`
- `.am-secondary`
- `.am-success`
- `.am-warning`
- `.am-danger`

`````html
<a href="" class="am-icon-btn am-icon-twitter"></a>
<a href="" class="am-icon-btn am-icon-facebook"></a>
<a href="" class="am-icon-btn am-icon-github"></a>
<a href="" class="am-icon-btn am-primary am-icon-qq"></a>
<a href="" class="am-icon-btn am-secondary am-icon-drupal"></a>
<a href="" class="am-icon-btn am-success am-icon-shield"></a>
<a href="" class="am-icon-btn am-warning am-icon-warning"></a>
<a href="" class="am-icon-btn am-danger am-icon-youtube"></a>
`````
```html
<a href="" class="am-icon-btn am-icon-twitter"></a>
<a href="" class="am-icon-btn am-icon-facebook"></a>
<a href="" class="am-icon-btn am-icon-github"></a>
<a href="" class="am-icon-btn am-primary am-icon-qq"></a>
<a href="" class="am-icon-btn am-secondary am-icon-drupal"></a>
<a href="" class="am-icon-btn am-success am-icon-shield"></a>
<a href="" class="am-icon-btn am-warning am-icon-warning"></a>
<a href="" class="am-icon-btn am-danger am-icon-youtube"></a>
```

## Spinning 

Attention: In Chrome and Firefox, spinning animation can be applied only to elements with `display: inline-block;` or `display: block;`.

`````html
<i class="am-icon-spinner am-icon-spin"></i>
<i class="am-icon-refresh am-icon-spin"></i>
<i class="am-icon-circle-o-notch am-icon-spin"></i>
<i class="am-icon-cog am-icon-spin"></i>
<i class="am-icon-gear am-icon-spin"></i>
`````

```html
<i class="am-icon-spinner am-icon-spin"></i>
<i class="am-icon-refresh am-icon-spin"></i>
<i class="am-icon-circle-o-notch am-icon-spin"></i>
<i class="am-icon-cog am-icon-spin"></i>
<i class="am-icon-gear am-icon-spin"></i>
```

**v2.3** New Animation：

`````html
<i class="am-icon-spinner am-icon-pulse"></i>
`````
```html
<i class="am-icon-spinner am-icon-pulse"></i>
```

## Fix Height

Icons in FontAwesome are slightly different in height. Add the `.am-icon-fw` class to fix this problem(**New feature in v2.3**)

`````html
<ul>
  <li><i class="am-icon-qq am-icon-fw"></i> QQ</li>
  <li><i class="am-icon-skype am-icon-fw"></i> Skype</li>
  <li><i class="am-icon-github am-icon-fw"></i> GitHub</li>
  <li><i class="am-icon-cc-amex am-icon-fw"></i> Amex</li>
</ul>
`````


## Copy Icon

Display two small buttons when mouse hover on the icon:

- `class`: Copy the classes' name. Used when DOM can be modified.
- `style`: Copy the styles, Used when DOM can't be modified.

```css
{
  .am-icon-font;
  content: @fa-var-copy;
}
```

## Problems

### Some user agents can't display icon font correctly

There are some user agents that can't correctly display Icon Font. This is probably because these user agents can't handle the 5-digit Hex coding of Icon Font in pseudo element `content`. You can find more information [Here](http://www.cnblogs.com/ljack/p/3751678.html), and test your user agent [Here](http://www.w3cmark.com/demo/iconfont.html).

Two Solutions:

- ~~**Use 4-digit coding**：We can't do that in Amaze UI, because there are over 500 icons, but you may do it yourself.~~
- **Directly use Icon Font's code in HTML**.

`````html
<span>&#xf09b; What a fuck.</span>
`````
```html
<span>&#xf09b; What a fuck.</span>
```

Amaze UI is designed for modern browsers. Even though we provide support to IE8/9, we can't change the basic structure and spend too much energy and coffee to fully support them. If you have some problem about icons' incorrect display on some less popular browsers, we really suggest you to consider the cost and benifit of solving this problem.

**v2.3 update**:

A developer provided an solution to the solve the problem of icon display. In `v2.3`, styles for icons have been changed to this:

```css
/* Solution of icons' incorrect display in Android:*/
[class*='am-icon-']:before {
  display: inline-block;
  font: normal normal normal 14px/1 FontAwesome;
  font-size: inherit;
  text-rendering: auto;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}　
```

## List of All Icons

`````html
<section id="new-4-3">
  <h2 class="doc-icon-hd">New in Font Awesome 4.3 (Used in Amaze UI 2.2) </h2>
  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/bed"><i class="am-icon-bed"></i> bed</a></li>
    <li><a href="http://fontawesome.io/icon/buysellads"><i class="am-icon-buysellads"></i> buysellads</a></li>
    <li><a href="http://fontawesome.io/icon/cart-arrow-down"><i class="am-icon-cart-arrow-down"></i> cart-arrow-down</a></li>
    <li><a href="http://fontawesome.io/icon/cart-plus"><i class="am-icon-cart-plus"></i> cart-plus</a></li>
    <li><a href="http://fontawesome.io/icon/connectdevelop"><i class="am-icon-connectdevelop"></i> connectdevelop</a></li>
    <li><a href="http://fontawesome.io/icon/dashcube"><i class="am-icon-dashcube"></i> dashcube</a></li>
    <li><a href="http://fontawesome.io/icon/diamond"><i class="am-icon-diamond"></i> diamond</a></li>
    <li><a href="http://fontawesome.io/icon/facebook-official"><i class="am-icon-facebook-official"></i> facebook-official</a></li>
    <li><a href="http://fontawesome.io/icon/forumbee"><i class="am-icon-forumbee"></i> forumbee</a></li>
    <li><a href="http://fontawesome.io/icon/heartbeat"><i class="am-icon-heartbeat"></i> heartbeat</a></li>
    <li><a href="http://fontawesome.io/icon/bed"><i class="am-icon-hotel"></i> hotel <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/leanpub"><i class="am-icon-leanpub"></i> leanpub</a></li>
    <li><a href="http://fontawesome.io/icon/mars"><i class="am-icon-mars"></i> mars</a></li>
    <li><a href="http://fontawesome.io/icon/mars-double"><i class="am-icon-mars-double"></i> mars-double</a></li>
    <li><a href="http://fontawesome.io/icon/mars-stroke"><i class="am-icon-mars-stroke"></i> mars-stroke</a></li>
    <li><a href="http://fontawesome.io/icon/mars-stroke-h"><i class="am-icon-mars-stroke-h"></i> mars-stroke-h</a></li>
    <li><a href="http://fontawesome.io/icon/mars-stroke-v"><i class="am-icon-mars-stroke-v"></i> mars-stroke-v</a></li>
    <li><a href="http://fontawesome.io/icon/medium"><i class="am-icon-medium"></i> medium</a></li>
    <li><a href="http://fontawesome.io/icon/mercury"><i class="am-icon-mercury"></i> mercury</a></li>
    <li><a href="http://fontawesome.io/icon/motorcycle"><i class="am-icon-motorcycle"></i> motorcycle</a></li>
    <li><a href="http://fontawesome.io/icon/neuter"><i class="am-icon-neuter"></i> neuter</a></li>
    <li><a href="http://fontawesome.io/icon/pinterest-p"><i class="am-icon-pinterest-p"></i> pinterest-p</a></li>
    <li><a href="http://fontawesome.io/icon/sellsy"><i class="am-icon-sellsy"></i> sellsy</a></li>
    <li><a href="http://fontawesome.io/icon/server"><i class="am-icon-server"></i> server</a></li>
    <li><a href="http://fontawesome.io/icon/ship"><i class="am-icon-ship"></i> ship</a></li>
    <li><a href="http://fontawesome.io/icon/shirtsinbulk"><i class="am-icon-shirtsinbulk"></i> shirtsinbulk</a></li>
    <li><a href="http://fontawesome.io/icon/simplybuilt"><i class="am-icon-simplybuilt"></i> simplybuilt</a></li>
    <li><a href="http://fontawesome.io/icon/skyatlas"><i class="am-icon-skyatlas"></i> skyatlas</a></li>
    <li><a href="http://fontawesome.io/icon/street-view"><i class="am-icon-street-view"></i> street-view</a></li>
    <li><a href="http://fontawesome.io/icon/subway"><i class="am-icon-subway"></i> subway</a></li>
    <li><a href="http://fontawesome.io/icon/train"><i class="am-icon-train"></i> train</a></li>
    <li><a href="http://fontawesome.io/icon/transgender"><i class="am-icon-transgender"></i> transgender</a></li>
    <li><a href="http://fontawesome.io/icon/transgender-alt"><i class="am-icon-transgender-alt"></i> transgender-alt</a></li>
    <li><a href="http://fontawesome.io/icon/user-plus"><i class="am-icon-user-plus"></i> user-plus</a></li>
    <li><a href="http://fontawesome.io/icon/user-secret"><i class="am-icon-user-secret"></i> user-secret</a></li>
    <li><a href="http://fontawesome.io/icon/user-times"><i class="am-icon-user-times"></i> user-times</a></li>
    <li><a href="http://fontawesome.io/icon/venus"><i class="am-icon-venus"></i> venus</a></li>
    <li><a href="http://fontawesome.io/icon/venus-double"><i class="am-icon-venus-double"></i> venus-double</a></li>
    <li><a href="http://fontawesome.io/icon/venus-mars"><i class="am-icon-venus-mars"></i> venus-mars</a></li>
    <li><a href="http://fontawesome.io/icon/viacoin"><i class="am-icon-viacoin"></i> viacoin</a></li>
    <li><a href="http://fontawesome.io/icon/whatsapp"><i class="am-icon-whatsapp"></i> whatsapp</a></li>
    </ul>
</section>

<section id="new">
  <h2 class="doc-icon-hd">40 New Icons in 4.2</h2>
  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/angellist"><i class="am-icon-angellist"></i> angellist</a></li>
    <li><a href="http://fontawesome.io/icon/area-chart"><i class="am-icon-area-chart"></i> area-chart</a></li>
    <li><a href="http://fontawesome.io/icon/at"><i class="am-icon-at"></i> at</a></li>
    <li><a href="http://fontawesome.io/icon/bell-slash"><i class="am-icon-bell-slash"></i> bell-slash</a></li>
    <li><a href="http://fontawesome.io/icon/bell-slash-o"><i class="am-icon-bell-slash-o"></i> bell-slash-o</a></li>
    <li><a href="http://fontawesome.io/icon/bicycle"><i class="am-icon-bicycle"></i> bicycle</a></li>
    <li><a href="http://fontawesome.io/icon/binoculars"><i class="am-icon-binoculars"></i> binoculars</a></li>
    <li><a href="http://fontawesome.io/icon/birthday-cake"><i class="am-icon-birthday-cake"></i> birthday-cake</a></li>
    <li><a href="http://fontawesome.io/icon/bus"><i class="am-icon-bus"></i> bus</a></li>
    <li><a href="http://fontawesome.io/icon/calculator"><i class="am-icon-calculator"></i> calculator</a></li>
    <li><a href="http://fontawesome.io/icon/cc"><i class="am-icon-cc"></i> cc</a></li>
    <li><a href="http://fontawesome.io/icon/cc-amex"><i class="am-icon-cc-amex"></i> cc-amex</a></li>
    <li><a href="http://fontawesome.io/icon/cc-discover"><i class="am-icon-cc-discover"></i> cc-discover</a></li>
    <li><a href="http://fontawesome.io/icon/cc-mastercard"><i class="am-icon-cc-mastercard"></i> cc-mastercard</a></li>
    <li><a href="http://fontawesome.io/icon/cc-paypal"><i class="am-icon-cc-paypal"></i> cc-paypal</a></li>
    <li><a href="http://fontawesome.io/icon/cc-stripe"><i class="am-icon-cc-stripe"></i> cc-stripe</a></li>
    <li><a href="http://fontawesome.io/icon/cc-visa"><i class="am-icon-cc-visa"></i> cc-visa</a></li>
    <li><a href="http://fontawesome.io/icon/copyright"><i class="am-icon-copyright"></i> copyright</a></li>
    <li><a href="http://fontawesome.io/icon/eyedropper"><i class="am-icon-eyedropper"></i> eyedropper</a></li>
    <li><a href="http://fontawesome.io/icon/futbol-o"><i class="am-icon-futbol-o"></i> futbol-o</a></li>
    <li><a href="http://fontawesome.io/icon/google-wallet"><i class="am-icon-google-wallet"></i> google-wallet</a></li>
    <li><a href="http://fontawesome.io/icon/ils"><i class="am-icon-ils"></i> ils</a></li>
    <li><a href="http://fontawesome.io/icon/ioxhost"><i class="am-icon-ioxhost"></i> ioxhost</a></li>
    <li><a href="http://fontawesome.io/icon/lastfm"><i class="am-icon-lastfm"></i> lastfm</a></li>
    <li><a href="http://fontawesome.io/icon/lastfm-square"><i class="am-icon-lastfm-square"></i> lastfm-square</a></li>
    <li><a href="http://fontawesome.io/icon/line-chart"><i class="am-icon-line-chart"></i> line-chart</a></li>
    <li><a href="http://fontawesome.io/icon/meanpath"><i class="am-icon-meanpath"></i> meanpath</a></li>
    <li><a href="http://fontawesome.io/icon/newspaper-o"><i class="am-icon-newspaper-o"></i> newspaper-o</a></li>
    <li><a href="http://fontawesome.io/icon/paint-brush"><i class="am-icon-paint-brush"></i> paint-brush</a></li>
    <li><a href="http://fontawesome.io/icon/paypal"><i class="am-icon-paypal"></i> paypal</a></li>

    <li><a href="http://fontawesome.io/icon/pie-chart"><i class="am-icon-pie-chart"></i> pie-chart</a></li>

    <li><a href="http://fontawesome.io/icon/plug"><i class="am-icon-plug"></i> plug</a></li>

    <li><a href="http://fontawesome.io/icon/ils"><i class="am-icon-shekel"></i> shekel <span
        class="text-muted">(alias)</span></a></li>

    <li><a href="http://fontawesome.io/icon/ils"><i class="am-icon-sheqel"></i> sheqel <span
        class="text-muted">(alias)</span></a></li>

    <li><a href="http://fontawesome.io/icon/slideshare"><i class="am-icon-slideshare"></i> slideshare</a></li>

    <li><a href="http://fontawesome.io/icon/futbol-o"><i class="am-icon-soccer-ball-o"></i> soccer-ball-o <span
        class="text-muted">(alias)</span></a></li>

    <li><a href="http://fontawesome.io/icon/toggle-off"><i class="am-icon-toggle-off"></i> toggle-off</a></li>

    <li><a href="http://fontawesome.io/icon/toggle-on"><i class="am-icon-toggle-on"></i> toggle-on</a></li>

    <li><a href="http://fontawesome.io/icon/trash"><i class="am-icon-trash"></i> trash</a></li>

    <li><a href="http://fontawesome.io/icon/tty"><i class="am-icon-tty"></i> tty</a></li>

    <li><a href="http://fontawesome.io/icon/twitch"><i class="am-icon-twitch"></i> twitch</a></li>

    <li><a href="http://fontawesome.io/icon/wifi"><i class="am-icon-wifi"></i> wifi</a></li>

    <li><a href="http://fontawesome.io/icon/yelp"><i class="am-icon-yelp"></i> yelp</a></li>
  </ul>
</section>


<section id="web-application">
<h2 class="doc-icon-hd">Web Application Icons</h2>

<ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
<li><a href="http://fontawesome.io/icon/adjust"><i class="am-icon-adjust"></i> adjust</a></li>
<li><a href="http://fontawesome.io/icon/anchor"><i class="am-icon-anchor"></i> anchor</a></li>
<li><a href="http://fontawesome.io/icon/archive"><i class="am-icon-archive"></i> archive</a></li>
<li><a href="http://fontawesome.io/icon/area-chart"><i class="am-icon-area-chart"></i> area-chart</a></li>
<li><a href="http://fontawesome.io/icon/arrows"><i class="am-icon-arrows"></i> arrows</a></li>
<li><a href="http://fontawesome.io/icon/arrows-h"><i class="am-icon-arrows-h"></i> arrows-h</a></li>
<li><a href="http://fontawesome.io/icon/arrows-v"><i class="am-icon-arrows-v"></i> arrows-v</a></li>
<li><a href="http://fontawesome.io/icon/asterisk"><i class="am-icon-asterisk"></i> asterisk</a></li>
<li><a href="http://fontawesome.io/icon/at"><i class="am-icon-at"></i> at</a></li>
<li><a href="http://fontawesome.io/icon/car"><i class="am-icon-automobile"></i> automobile <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/ban"><i class="am-icon-ban"></i> ban</a></li>
<li><a href="http://fontawesome.io/icon/university"><i class="am-icon-bank"></i> bank <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/bar-chart"><i class="am-icon-bar-chart"></i> bar-chart</a></li>
<li><a href="http://fontawesome.io/icon/bar-chart"><i class="am-icon-bar-chart-o"></i> bar-chart-o <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/barcode"><i class="am-icon-barcode"></i> barcode</a></li>
<li><a href="http://fontawesome.io/icon/bars"><i class="am-icon-bars"></i> bars</a></li>
<li><a href="http://fontawesome.io/icon/beer"><i class="am-icon-beer"></i> beer</a></li>
<li><a href="http://fontawesome.io/icon/bell"><i class="am-icon-bell"></i> bell</a></li>
<li><a href="http://fontawesome.io/icon/bell-o"><i class="am-icon-bell-o"></i> bell-o</a></li>
<li><a href="http://fontawesome.io/icon/bell-slash"><i class="am-icon-bell-slash"></i> bell-slash</a></li>
<li><a href="http://fontawesome.io/icon/bell-slash-o"><i class="am-icon-bell-slash-o"></i> bell-slash-o</a></li>
<li><a href="http://fontawesome.io/icon/bicycle"><i class="am-icon-bicycle"></i> bicycle</a></li>
<li><a href="http://fontawesome.io/icon/binoculars"><i class="am-icon-binoculars"></i> binoculars</a></li>
<li><a href="http://fontawesome.io/icon/birthday-cake"><i class="am-icon-birthday-cake"></i> birthday-cake</a></li>
<li><a href="http://fontawesome.io/icon/bolt"><i class="am-icon-bolt"></i> bolt</a></li>
<li><a href="http://fontawesome.io/icon/bomb"><i class="am-icon-bomb"></i> bomb</a></li>
<li><a href="http://fontawesome.io/icon/book"><i class="am-icon-book"></i> book</a></li>
<li><a href="http://fontawesome.io/icon/bookmark"><i class="am-icon-bookmark"></i> bookmark</a></li>
<li><a href="http://fontawesome.io/icon/bookmark-o"><i class="am-icon-bookmark-o"></i> bookmark-o</a></li>
<li><a href="http://fontawesome.io/icon/briefcase"><i class="am-icon-briefcase"></i> briefcase</a></li>
<li><a href="http://fontawesome.io/icon/bug"><i class="am-icon-bug"></i> bug</a></li>
<li><a href="http://fontawesome.io/icon/building"><i class="am-icon-building"></i> building</a></li>
<li><a href="http://fontawesome.io/icon/building-o"><i class="am-icon-building-o"></i> building-o</a></li>
<li><a href="http://fontawesome.io/icon/bullhorn"><i class="am-icon-bullhorn"></i> bullhorn</a></li>
<li><a href="http://fontawesome.io/icon/bullseye"><i class="am-icon-bullseye"></i> bullseye</a></li>
<li><a href="http://fontawesome.io/icon/bus"><i class="am-icon-bus"></i> bus</a></li>
<li><a href="http://fontawesome.io/icon/taxi"><i class="am-icon-cab"></i> cab <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/calculator"><i class="am-icon-calculator"></i> calculator</a></li>
<li><a href="http://fontawesome.io/icon/calendar"><i class="am-icon-calendar"></i> calendar</a></li>
<li><a href="http://fontawesome.io/icon/calendar-o"><i class="am-icon-calendar-o"></i> calendar-o</a></li>
<li><a href="http://fontawesome.io/icon/camera"><i class="am-icon-camera"></i> camera</a></li>
<li><a href="http://fontawesome.io/icon/camera-retro"><i class="am-icon-camera-retro"></i> camera-retro</a></li>
<li><a href="http://fontawesome.io/icon/car"><i class="am-icon-car"></i> car</a></li>
<li><a href="http://fontawesome.io/icon/caret-square-o-down"><i class="am-icon-caret-square-o-down"></i> caret-square-o-down</a></li>
<li><a href="http://fontawesome.io/icon/caret-square-o-left"><i class="am-icon-caret-square-o-left"></i> caret-square-o-left</a></li>
<li><a href="http://fontawesome.io/icon/caret-square-o-right"><i class="am-icon-caret-square-o-right"></i> caret-square-o-right</a></li>
<li><a href="http://fontawesome.io/icon/caret-square-o-up"><i class="am-icon-caret-square-o-up"></i> caret-square-o-up</a></li>
<li><a href="http://fontawesome.io/icon/cc"><i class="am-icon-cc"></i> cc</a></li>
<li><a href="http://fontawesome.io/icon/certificate"><i class="am-icon-certificate"></i> certificate</a></li>
<li><a href="http://fontawesome.io/icon/check"><i class="am-icon-check"></i> check</a></li>
<li><a href="http://fontawesome.io/icon/check-circle"><i class="am-icon-check-circle"></i> check-circle</a></li>
<li><a href="http://fontawesome.io/icon/check-circle-o"><i class="am-icon-check-circle-o"></i> check-circle-o</a></li>
<li><a href="http://fontawesome.io/icon/check-square"><i class="am-icon-check-square"></i> check-square</a></li>
<li><a href="http://fontawesome.io/icon/check-square-o"><i class="am-icon-check-square-o"></i> check-square-o</a></li>
<li><a href="http://fontawesome.io/icon/child"><i class="am-icon-child"></i> child</a></li>
<li><a href="http://fontawesome.io/icon/circle"><i class="am-icon-circle"></i> circle</a></li>
<li><a href="http://fontawesome.io/icon/circle-o"><i class="am-icon-circle-o"></i> circle-o</a></li>
<li><a href="http://fontawesome.io/icon/circle-o-notch"><i class="am-icon-circle-o-notch"></i> circle-o-notch</a></li>
<li><a href="http://fontawesome.io/icon/circle-thin"><i class="am-icon-circle-thin"></i> circle-thin</a></li>
<li><a href="http://fontawesome.io/icon/clock-o"><i class="am-icon-clock-o"></i> clock-o</a></li>
<li><a href="http://fontawesome.io/icon/times"><i class="am-icon-close"></i> close <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/cloud"><i class="am-icon-cloud"></i> cloud</a></li>
<li><a href="http://fontawesome.io/icon/cloud-download"><i class="am-icon-cloud-download"></i> cloud-download</a></li>
<li><a href="http://fontawesome.io/icon/cloud-upload"><i class="am-icon-cloud-upload"></i> cloud-upload</a></li>
<li><a href="http://fontawesome.io/icon/code"><i class="am-icon-code"></i> code</a></li>
<li><a href="http://fontawesome.io/icon/code-fork"><i class="am-icon-code-fork"></i> code-fork</a></li>
<li><a href="http://fontawesome.io/icon/coffee"><i class="am-icon-coffee"></i> coffee</a></li>
<li><a href="http://fontawesome.io/icon/cog"><i class="am-icon-cog"></i> cog</a></li>
<li><a href="http://fontawesome.io/icon/cogs"><i class="am-icon-cogs"></i> cogs</a></li>
<li><a href="http://fontawesome.io/icon/comment"><i class="am-icon-comment"></i> comment</a></li>
<li><a href="http://fontawesome.io/icon/comment-o"><i class="am-icon-comment-o"></i> comment-o</a></li>
<li><a href="http://fontawesome.io/icon/comments"><i class="am-icon-comments"></i> comments</a></li>
<li><a href="http://fontawesome.io/icon/comments-o"><i class="am-icon-comments-o"></i> comments-o</a></li>
<li><a href="http://fontawesome.io/icon/compass"><i class="am-icon-compass"></i> compass</a></li>
<li><a href="http://fontawesome.io/icon/copyright"><i class="am-icon-copyright"></i> copyright</a></li>
<li><a href="http://fontawesome.io/icon/credit-card"><i class="am-icon-credit-card"></i> credit-card</a></li>
<li><a href="http://fontawesome.io/icon/crop"><i class="am-icon-crop"></i> crop</a></li>
<li><a href="http://fontawesome.io/icon/crosshairs"><i class="am-icon-crosshairs"></i> crosshairs</a></li>
<li><a href="http://fontawesome.io/icon/cube"><i class="am-icon-cube"></i> cube</a></li>
<li><a href="http://fontawesome.io/icon/cubes"><i class="am-icon-cubes"></i> cubes</a></li>
<li><a href="http://fontawesome.io/icon/cutlery"><i class="am-icon-cutlery"></i> cutlery</a></li>
<li><a href="http://fontawesome.io/icon/tachometer"><i class="am-icon-dashboard"></i> dashboard <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/database"><i class="am-icon-database"></i> database</a></li>
<li><a href="http://fontawesome.io/icon/desktop"><i class="am-icon-desktop"></i> desktop</a></li>
<li><a href="http://fontawesome.io/icon/dot-circle-o"><i class="am-icon-dot-circle-o"></i> dot-circle-o</a></li>
<li><a href="http://fontawesome.io/icon/download"><i class="am-icon-download"></i> download</a></li>
<li><a href="http://fontawesome.io/icon/pencil-square-o"><i class="am-icon-edit"></i> edit <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/ellipsis-h"><i class="am-icon-ellipsis-h"></i> ellipsis-h</a></li>
<li><a href="http://fontawesome.io/icon/ellipsis-v"><i class="am-icon-ellipsis-v"></i> ellipsis-v</a></li>
<li><a href="http://fontawesome.io/icon/envelope"><i class="am-icon-envelope"></i> envelope</a></li>
<li><a href="http://fontawesome.io/icon/envelope-o"><i class="am-icon-envelope-o"></i> envelope-o</a></li>
<li><a href="http://fontawesome.io/icon/envelope-square"><i class="am-icon-envelope-square"></i> envelope-square</a></li>
<li><a href="http://fontawesome.io/icon/eraser"><i class="am-icon-eraser"></i> eraser</a></li>
<li><a href="http://fontawesome.io/icon/exchange"><i class="am-icon-exchange"></i> exchange</a></li>
<li><a href="http://fontawesome.io/icon/exclamation"><i class="am-icon-exclamation"></i> exclamation</a></li>
<li><a href="http://fontawesome.io/icon/exclamation-circle"><i class="am-icon-exclamation-circle"></i> exclamation-circle</a></li>
<li><a href="http://fontawesome.io/icon/exclamation-triangle"><i class="am-icon-exclamation-triangle"></i> exclamation-triangle</a></li>
<li><a href="http://fontawesome.io/icon/external-link"><i class="am-icon-external-link"></i> external-link</a></li>
<li><a href="http://fontawesome.io/icon/external-link-square"><i class="am-icon-external-link-square"></i> external-link-square</a></li>
<li><a href="http://fontawesome.io/icon/eye"><i class="am-icon-eye"></i> eye</a></li>
<li><a href="http://fontawesome.io/icon/eye-slash"><i class="am-icon-eye-slash"></i> eye-slash</a></li>
<li><a href="http://fontawesome.io/icon/eyedropper"><i class="am-icon-eyedropper"></i> eyedropper</a></li>
<li><a href="http://fontawesome.io/icon/fax"><i class="am-icon-fax"></i> fax</a></li>
<li><a href="http://fontawesome.io/icon/female"><i class="am-icon-female"></i> female</a></li>
<li><a href="http://fontawesome.io/icon/fighter-jet"><i class="am-icon-fighter-jet"></i> fighter-jet</a></li>
<li><a href="http://fontawesome.io/icon/file-archive-o"><i class="am-icon-file-archive-o"></i> file-archive-o</a></li>
<li><a href="http://fontawesome.io/icon/file-audio-o"><i class="am-icon-file-audio-o"></i> file-audio-o</a></li>
<li><a href="http://fontawesome.io/icon/file-code-o"><i class="am-icon-file-code-o"></i> file-code-o</a></li>
<li><a href="http://fontawesome.io/icon/file-excel-o"><i class="am-icon-file-excel-o"></i> file-excel-o</a></li>
<li><a href="http://fontawesome.io/icon/file-image-o"><i class="am-icon-file-image-o"></i> file-image-o</a></li>
<li><a href="http://fontawesome.io/icon/file-video-o"><i class="am-icon-file-movie-o"></i> file-movie-o <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/file-pdf-o"><i class="am-icon-file-pdf-o"></i> file-pdf-o</a></li>
<li><a href="http://fontawesome.io/icon/file-image-o"><i class="am-icon-file-photo-o"></i> file-photo-o <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/file-image-o"><i class="am-icon-file-picture-o"></i> file-picture-o <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/file-powerpoint-o"><i class="am-icon-file-powerpoint-o"></i> file-powerpoint-o</a></li>
<li><a href="http://fontawesome.io/icon/file-audio-o"><i class="am-icon-file-sound-o"></i> file-sound-o <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/file-video-o"><i class="am-icon-file-video-o"></i> file-video-o</a></li>
<li><a href="http://fontawesome.io/icon/file-word-o"><i class="am-icon-file-word-o"></i> file-word-o</a></li>
<li><a href="http://fontawesome.io/icon/file-archive-o"><i class="am-icon-file-zip-o"></i> file-zip-o <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/film"><i class="am-icon-film"></i> film</a></li>
<li><a href="http://fontawesome.io/icon/filter"><i class="am-icon-filter"></i> filter</a></li>
<li><a href="http://fontawesome.io/icon/fire"><i class="am-icon-fire"></i> fire</a></li>
<li><a href="http://fontawesome.io/icon/fire-extinguisher"><i class="am-icon-fire-extinguisher"></i> fire-extinguisher</a></li>
<li><a href="http://fontawesome.io/icon/flag"><i class="am-icon-flag"></i> flag</a></li>
<li><a href="http://fontawesome.io/icon/flag-checkered"><i class="am-icon-flag-checkered"></i> flag-checkered</a></li>
<li><a href="http://fontawesome.io/icon/flag-o"><i class="am-icon-flag-o"></i> flag-o</a></li>
<li><a href="http://fontawesome.io/icon/bolt"><i class="am-icon-flash"></i> flash <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/flask"><i class="am-icon-flask"></i> flask</a></li>
<li><a href="http://fontawesome.io/icon/folder"><i class="am-icon-folder"></i> folder</a></li>
<li><a href="http://fontawesome.io/icon/folder-o"><i class="am-icon-folder-o"></i> folder-o</a></li>
<li><a href="http://fontawesome.io/icon/folder-open"><i class="am-icon-folder-open"></i> folder-open</a></li>
<li><a href="http://fontawesome.io/icon/folder-open-o"><i class="am-icon-folder-open-o"></i> folder-open-o</a></li>
<li><a href="http://fontawesome.io/icon/frown-o"><i class="am-icon-frown-o"></i> frown-o</a></li>
<li><a href="http://fontawesome.io/icon/futbol-o"><i class="am-icon-futbol-o"></i> futbol-o</a></li>
<li><a href="http://fontawesome.io/icon/gamepad"><i class="am-icon-gamepad"></i> gamepad</a></li>
<li><a href="http://fontawesome.io/icon/gavel"><i class="am-icon-gavel"></i> gavel</a></li>
<li><a href="http://fontawesome.io/icon/cog"><i class="am-icon-gear"></i> gear <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/cogs"><i class="am-icon-gears"></i> gears <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/gift"><i class="am-icon-gift"></i> gift</a></li>
<li><a href="http://fontawesome.io/icon/glass"><i class="am-icon-glass"></i> glass</a></li>
<li><a href="http://fontawesome.io/icon/globe"><i class="am-icon-globe"></i> globe</a></li>
<li><a href="http://fontawesome.io/icon/graduation-cap"><i class="am-icon-graduation-cap"></i> graduation-cap</a></li>
<li><a href="http://fontawesome.io/icon/users"><i class="am-icon-group"></i> group <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/hdd-o"><i class="am-icon-hdd-o"></i> hdd-o</a></li>
<li><a href="http://fontawesome.io/icon/headphones"><i class="am-icon-headphones"></i> headphones</a></li>
<li><a href="http://fontawesome.io/icon/heart"><i class="am-icon-heart"></i> heart</a></li>
<li><a href="http://fontawesome.io/icon/heart-o"><i class="am-icon-heart-o"></i> heart-o</a></li>
<li><a href="http://fontawesome.io/icon/history"><i class="am-icon-history"></i> history</a></li>
<li><a href="http://fontawesome.io/icon/home"><i class="am-icon-home"></i> home</a></li>
<li><a href="http://fontawesome.io/icon/picture-o"><i class="am-icon-image"></i> image <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/inbox"><i class="am-icon-inbox"></i> inbox</a></li>
<li><a href="http://fontawesome.io/icon/info"><i class="am-icon-info"></i> info</a></li>
<li><a href="http://fontawesome.io/icon/info-circle"><i class="am-icon-info-circle"></i> info-circle</a></li>
<li><a href="http://fontawesome.io/icon/university"><i class="am-icon-institution"></i> institution <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/key"><i class="am-icon-key"></i> key</a></li>
<li><a href="http://fontawesome.io/icon/keyboard-o"><i class="am-icon-keyboard-o"></i> keyboard-o</a></li>
<li><a href="http://fontawesome.io/icon/language"><i class="am-icon-language"></i> language</a></li>
<li><a href="http://fontawesome.io/icon/laptop"><i class="am-icon-laptop"></i> laptop</a></li>
<li><a href="http://fontawesome.io/icon/leaf"><i class="am-icon-leaf"></i> leaf</a></li>
<li><a href="http://fontawesome.io/icon/gavel"><i class="am-icon-legal"></i> legal <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/lemon-o"><i class="am-icon-lemon-o"></i> lemon-o</a></li>
<li><a href="http://fontawesome.io/icon/level-down"><i class="am-icon-level-down"></i> level-down</a></li>
<li><a href="http://fontawesome.io/icon/level-up"><i class="am-icon-level-up"></i> level-up</a></li>
<li><a href="http://fontawesome.io/icon/life-ring"><i class="am-icon-life-bouy"></i> life-bouy <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/life-ring"><i class="am-icon-life-buoy"></i> life-buoy <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/life-ring"><i class="am-icon-life-ring"></i> life-ring</a></li>
<li><a href="http://fontawesome.io/icon/life-ring"><i class="am-icon-life-saver"></i> life-saver <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/lightbulb-o"><i class="am-icon-lightbulb-o"></i> lightbulb-o</a></li>
<li><a href="http://fontawesome.io/icon/line-chart"><i class="am-icon-line-chart"></i> line-chart</a></li>
<li><a href="http://fontawesome.io/icon/location-arrow"><i class="am-icon-location-arrow"></i> location-arrow</a></li>
<li><a href="http://fontawesome.io/icon/lock"><i class="am-icon-lock"></i> lock</a></li>
<li><a href="http://fontawesome.io/icon/magic"><i class="am-icon-magic"></i> magic</a></li>
<li><a href="http://fontawesome.io/icon/magnet"><i class="am-icon-magnet"></i> magnet</a></li>
<li><a href="http://fontawesome.io/icon/share"><i class="am-icon-mail-forward"></i> mail-forward <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/reply"><i class="am-icon-mail-reply"></i> mail-reply <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/reply-all"><i class="am-icon-mail-reply-all"></i> mail-reply-all <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/male"><i class="am-icon-male"></i> male</a></li>
<li><a href="http://fontawesome.io/icon/map-marker"><i class="am-icon-map-marker"></i> map-marker</a></li>
<li><a href="http://fontawesome.io/icon/meh-o"><i class="am-icon-meh-o"></i> meh-o</a></li>
<li><a href="http://fontawesome.io/icon/microphone"><i class="am-icon-microphone"></i> microphone</a></li>
<li><a href="http://fontawesome.io/icon/microphone-slash"><i class="am-icon-microphone-slash"></i> microphone-slash</a></li>
<li><a href="http://fontawesome.io/icon/minus"><i class="am-icon-minus"></i> minus</a></li>
<li><a href="http://fontawesome.io/icon/minus-circle"><i class="am-icon-minus-circle"></i> minus-circle</a></li>
<li><a href="http://fontawesome.io/icon/minus-square"><i class="am-icon-minus-square"></i> minus-square</a></li>
<li><a href="http://fontawesome.io/icon/minus-square-o"><i class="am-icon-minus-square-o"></i> minus-square-o</a></li>
<li><a href="http://fontawesome.io/icon/mobile"><i class="am-icon-mobile"></i> mobile</a></li>
<li><a href="http://fontawesome.io/icon/mobile"><i class="am-icon-mobile-phone"></i> mobile-phone <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/money"><i class="am-icon-money"></i> money</a></li>
<li><a href="http://fontawesome.io/icon/moon-o"><i class="am-icon-moon-o"></i> moon-o</a></li>
<li><a href="http://fontawesome.io/icon/graduation-cap"><i class="am-icon-mortar-board"></i> mortar-board <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/music"><i class="am-icon-music"></i> music</a></li>
<li><a href="http://fontawesome.io/icon/bars"><i class="am-icon-navicon"></i> navicon <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/newspaper-o"><i class="am-icon-newspaper-o"></i> newspaper-o</a></li>
<li><a href="http://fontawesome.io/icon/paint-brush"><i class="am-icon-paint-brush"></i> paint-brush</a></li>
<li><a href="http://fontawesome.io/icon/paper-plane"><i class="am-icon-paper-plane"></i> paper-plane</a></li>
<li><a href="http://fontawesome.io/icon/paper-plane-o"><i class="am-icon-paper-plane-o"></i> paper-plane-o</a></li>
<li><a href="http://fontawesome.io/icon/paw"><i class="am-icon-paw"></i> paw</a></li>
<li><a href="http://fontawesome.io/icon/pencil"><i class="am-icon-pencil"></i> pencil</a></li>
<li><a href="http://fontawesome.io/icon/pencil-square"><i class="am-icon-pencil-square"></i> pencil-square</a></li>
<li><a href="http://fontawesome.io/icon/pencil-square-o"><i class="am-icon-pencil-square-o"></i> pencil-square-o</a></li>
<li><a href="http://fontawesome.io/icon/phone"><i class="am-icon-phone"></i> phone</a></li>
<li><a href="http://fontawesome.io/icon/phone-square"><i class="am-icon-phone-square"></i> phone-square</a></li>
<li><a href="http://fontawesome.io/icon/picture-o"><i class="am-icon-photo"></i> photo <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/picture-o"><i class="am-icon-picture-o"></i> picture-o</a></li>
<li><a href="http://fontawesome.io/icon/pie-chart"><i class="am-icon-pie-chart"></i> pie-chart</a></li>
<li><a href="http://fontawesome.io/icon/plane"><i class="am-icon-plane"></i> plane</a></li>
<li><a href="http://fontawesome.io/icon/plug"><i class="am-icon-plug"></i> plug</a></li>
<li><a href="http://fontawesome.io/icon/plus"><i class="am-icon-plus"></i> plus</a></li>
<li><a href="http://fontawesome.io/icon/plus-circle"><i class="am-icon-plus-circle"></i> plus-circle</a></li>
<li><a href="http://fontawesome.io/icon/plus-square"><i class="am-icon-plus-square"></i> plus-square</a></li>
<li><a href="http://fontawesome.io/icon/plus-square-o"><i class="am-icon-plus-square-o"></i> plus-square-o</a></li>
<li><a href="http://fontawesome.io/icon/power-off"><i class="am-icon-power-off"></i> power-off</a></li>
<li><a href="http://fontawesome.io/icon/print"><i class="am-icon-print"></i> print</a></li>
<li><a href="http://fontawesome.io/icon/puzzle-piece"><i class="am-icon-puzzle-piece"></i> puzzle-piece</a></li>
<li><a href="http://fontawesome.io/icon/qrcode"><i class="am-icon-qrcode"></i> qrcode</a></li>
<li><a href="http://fontawesome.io/icon/question"><i class="am-icon-question"></i> question</a></li>
<li><a href="http://fontawesome.io/icon/question-circle"><i class="am-icon-question-circle"></i> question-circle</a></li>
<li><a href="http://fontawesome.io/icon/quote-left"><i class="am-icon-quote-left"></i> quote-left</a></li>
<li><a href="http://fontawesome.io/icon/quote-right"><i class="am-icon-quote-right"></i> quote-right</a></li>
<li><a href="http://fontawesome.io/icon/random"><i class="am-icon-random"></i> random</a></li>
<li><a href="http://fontawesome.io/icon/recycle"><i class="am-icon-recycle"></i> recycle</a></li>
<li><a href="http://fontawesome.io/icon/refresh"><i class="am-icon-refresh"></i> refresh</a></li>
<li><a href="http://fontawesome.io/icon/times"><i class="am-icon-remove"></i> remove <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/bars"><i class="am-icon-reorder"></i> reorder <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/reply"><i class="am-icon-reply"></i> reply</a></li>
<li><a href="http://fontawesome.io/icon/reply-all"><i class="am-icon-reply-all"></i> reply-all</a></li>
<li><a href="http://fontawesome.io/icon/retweet"><i class="am-icon-retweet"></i> retweet</a></li>
<li><a href="http://fontawesome.io/icon/road"><i class="am-icon-road"></i> road</a></li>
<li><a href="http://fontawesome.io/icon/rocket"><i class="am-icon-rocket"></i> rocket</a></li>
<li><a href="http://fontawesome.io/icon/rss"><i class="am-icon-rss"></i> rss</a></li>
<li><a href="http://fontawesome.io/icon/rss-square"><i class="am-icon-rss-square"></i> rss-square</a></li>
<li><a href="http://fontawesome.io/icon/search"><i class="am-icon-search"></i> search</a></li>
<li><a href="http://fontawesome.io/icon/search-minus"><i class="am-icon-search-minus"></i> search-minus</a></li>
<li><a href="http://fontawesome.io/icon/search-plus"><i class="am-icon-search-plus"></i> search-plus</a></li>
<li><a href="http://fontawesome.io/icon/paper-plane"><i class="am-icon-send"></i> send <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/paper-plane-o"><i class="am-icon-send-o"></i> send-o <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/share"><i class="am-icon-share"></i> share</a></li>
<li><a href="http://fontawesome.io/icon/share-alt"><i class="am-icon-share-alt"></i> share-alt</a></li>
<li><a href="http://fontawesome.io/icon/share-alt-square"><i class="am-icon-share-alt-square"></i> share-alt-square</a></li>
<li><a href="http://fontawesome.io/icon/share-square"><i class="am-icon-share-square"></i> share-square</a></li>
<li><a href="http://fontawesome.io/icon/share-square-o"><i class="am-icon-share-square-o"></i> share-square-o</a></li>
<li><a href="http://fontawesome.io/icon/shield"><i class="am-icon-shield"></i> shield</a></li>
<li><a href="http://fontawesome.io/icon/shopping-cart"><i class="am-icon-shopping-cart"></i> shopping-cart</a></li>
<li><a href="http://fontawesome.io/icon/sign-in"><i class="am-icon-sign-in"></i> sign-in</a></li>
<li><a href="http://fontawesome.io/icon/sign-out"><i class="am-icon-sign-out"></i> sign-out</a></li>
<li><a href="http://fontawesome.io/icon/signal"><i class="am-icon-signal"></i> signal</a></li>
<li><a href="http://fontawesome.io/icon/sitemap"><i class="am-icon-sitemap"></i> sitemap</a></li>
<li><a href="http://fontawesome.io/icon/sliders"><i class="am-icon-sliders"></i> sliders</a></li>
<li><a href="http://fontawesome.io/icon/smile-o"><i class="am-icon-smile-o"></i> smile-o</a></li>
<li><a href="http://fontawesome.io/icon/futbol-o"><i class="am-icon-soccer-ball-o"></i> soccer-ball-o <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/sort"><i class="am-icon-sort"></i> sort</a></li>
<li><a href="http://fontawesome.io/icon/sort-alpha-asc"><i class="am-icon-sort-alpha-asc"></i> sort-alpha-asc</a></li>
<li><a href="http://fontawesome.io/icon/sort-alpha-desc"><i class="am-icon-sort-alpha-desc"></i> sort-alpha-desc</a></li>
<li><a href="http://fontawesome.io/icon/sort-amount-asc"><i class="am-icon-sort-amount-asc"></i> sort-amount-asc</a></li>
<li><a href="http://fontawesome.io/icon/sort-amount-desc"><i class="am-icon-sort-amount-desc"></i> sort-amount-desc</a></li>
<li><a href="http://fontawesome.io/icon/sort-asc"><i class="am-icon-sort-asc"></i> sort-asc</a></li>
<li><a href="http://fontawesome.io/icon/sort-desc"><i class="am-icon-sort-desc"></i> sort-desc</a></li>
<li><a href="http://fontawesome.io/icon/sort-desc"><i class="am-icon-sort-down"></i> sort-down <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/sort-numeric-asc"><i class="am-icon-sort-numeric-asc"></i> sort-numeric-asc</a></li>
<li><a href="http://fontawesome.io/icon/sort-numeric-desc"><i class="am-icon-sort-numeric-desc"></i> sort-numeric-desc</a></li>
<li><a href="http://fontawesome.io/icon/sort-asc"><i class="am-icon-sort-up"></i> sort-up <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/space-shuttle"><i class="am-icon-space-shuttle"></i> space-shuttle</a></li>
<li><a href="http://fontawesome.io/icon/spinner"><i class="am-icon-spinner"></i> spinner</a></li>
<li><a href="http://fontawesome.io/icon/spoon"><i class="am-icon-spoon"></i> spoon</a></li>
<li><a href="http://fontawesome.io/icon/square"><i class="am-icon-square"></i> square</a></li>
<li><a href="http://fontawesome.io/icon/square-o"><i class="am-icon-square-o"></i> square-o</a></li>
<li><a href="http://fontawesome.io/icon/star"><i class="am-icon-star"></i> star</a></li>
<li><a href="http://fontawesome.io/icon/star-half"><i class="am-icon-star-half"></i> star-half</a></li>
<li><a href="http://fontawesome.io/icon/star-half-o"><i class="am-icon-star-half-empty"></i> star-half-empty <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/star-half-o"><i class="am-icon-star-half-full"></i> star-half-full <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/star-half-o"><i class="am-icon-star-half-o"></i> star-half-o</a></li>
<li><a href="http://fontawesome.io/icon/star-o"><i class="am-icon-star-o"></i> star-o</a></li>
<li><a href="http://fontawesome.io/icon/suitcase"><i class="am-icon-suitcase"></i> suitcase</a></li>
<li><a href="http://fontawesome.io/icon/sun-o"><i class="am-icon-sun-o"></i> sun-o</a></li>
<li><a href="http://fontawesome.io/icon/life-ring"><i class="am-icon-support"></i> support <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/tablet"><i class="am-icon-tablet"></i> tablet</a></li>
<li><a href="http://fontawesome.io/icon/tachometer"><i class="am-icon-tachometer"></i> tachometer</a></li>
<li><a href="http://fontawesome.io/icon/tag"><i class="am-icon-tag"></i> tag</a></li>
<li><a href="http://fontawesome.io/icon/tags"><i class="am-icon-tags"></i> tags</a></li>
<li><a href="http://fontawesome.io/icon/tasks"><i class="am-icon-tasks"></i> tasks</a></li>
<li><a href="http://fontawesome.io/icon/taxi"><i class="am-icon-taxi"></i> taxi</a></li>
<li><a href="http://fontawesome.io/icon/terminal"><i class="am-icon-terminal"></i> terminal</a></li>
<li><a href="http://fontawesome.io/icon/thumb-tack"><i class="am-icon-thumb-tack"></i> thumb-tack</a></li>
<li><a href="http://fontawesome.io/icon/thumbs-down"><i class="am-icon-thumbs-down"></i> thumbs-down</a></li>
<li><a href="http://fontawesome.io/icon/thumbs-o-down"><i class="am-icon-thumbs-o-down"></i> thumbs-o-down</a></li>
<li><a href="http://fontawesome.io/icon/thumbs-o-up"><i class="am-icon-thumbs-o-up"></i> thumbs-o-up</a></li>
<li><a href="http://fontawesome.io/icon/thumbs-up"><i class="am-icon-thumbs-up"></i> thumbs-up</a></li>
<li><a href="http://fontawesome.io/icon/ticket"><i class="am-icon-ticket"></i> ticket</a></li>
<li><a href="http://fontawesome.io/icon/times"><i class="am-icon-times"></i> times</a></li>
<li><a href="http://fontawesome.io/icon/times-circle"><i class="am-icon-times-circle"></i> times-circle</a></li>
<li><a href="http://fontawesome.io/icon/times-circle-o"><i class="am-icon-times-circle-o"></i> times-circle-o</a></li>
<li><a href="http://fontawesome.io/icon/tint"><i class="am-icon-tint"></i> tint</a></li>
<li><a href="http://fontawesome.io/icon/caret-square-o-down"><i class="am-icon-toggle-down"></i> toggle-down <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/caret-square-o-left"><i class="am-icon-toggle-left"></i> toggle-left <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/toggle-off"><i class="am-icon-toggle-off"></i> toggle-off</a></li>
<li><a href="http://fontawesome.io/icon/toggle-on"><i class="am-icon-toggle-on"></i> toggle-on</a></li>
<li><a href="http://fontawesome.io/icon/caret-square-o-right"><i class="am-icon-toggle-right"></i> toggle-right <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/caret-square-o-up"><i class="am-icon-toggle-up"></i> toggle-up <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/trash"><i class="am-icon-trash"></i> trash</a></li>
<li><a href="http://fontawesome.io/icon/trash-o"><i class="am-icon-trash-o"></i> trash-o</a></li>
<li><a href="http://fontawesome.io/icon/tree"><i class="am-icon-tree"></i> tree</a></li>
<li><a href="http://fontawesome.io/icon/trophy"><i class="am-icon-trophy"></i> trophy</a></li>
<li><a href="http://fontawesome.io/icon/truck"><i class="am-icon-truck"></i> truck</a></li>
<li><a href="http://fontawesome.io/icon/tty"><i class="am-icon-tty"></i> tty</a></li>
<li><a href="http://fontawesome.io/icon/umbrella"><i class="am-icon-umbrella"></i> umbrella</a></li>
<li><a href="http://fontawesome.io/icon/university"><i class="am-icon-university"></i> university</a></li>
<li><a href="http://fontawesome.io/icon/unlock"><i class="am-icon-unlock"></i> unlock</a></li>
<li><a href="http://fontawesome.io/icon/unlock-alt"><i class="am-icon-unlock-alt"></i> unlock-alt</a></li>
<li><a href="http://fontawesome.io/icon/sort"><i class="am-icon-unsorted"></i> unsorted <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/upload"><i class="am-icon-upload"></i> upload</a></li>
<li><a href="http://fontawesome.io/icon/user"><i class="am-icon-user"></i> user</a></li>
<li><a href="http://fontawesome.io/icon/users"><i class="am-icon-users"></i> users</a></li>
<li><a href="http://fontawesome.io/icon/video-camera"><i class="am-icon-video-camera"></i> video-camera</a></li>
<li><a href="http://fontawesome.io/icon/volume-down"><i class="am-icon-volume-down"></i> volume-down</a></li>
<li><a href="http://fontawesome.io/icon/volume-off"><i class="am-icon-volume-off"></i> volume-off</a></li>
<li><a href="http://fontawesome.io/icon/volume-up"><i class="am-icon-volume-up"></i> volume-up</a></li>
<li><a href="http://fontawesome.io/icon/exclamation-triangle"><i class="am-icon-warning"></i> warning <span class="text-muted">(alias)</span></a></li>
<li><a href="http://fontawesome.io/icon/wheelchair"><i class="am-icon-wheelchair"></i> wheelchair</a></li>
<li><a href="http://fontawesome.io/icon/wifi"><i class="am-icon-wifi"></i> wifi</a></li>
<li><a href="http://fontawesome.io/icon/wrench"><i class="am-icon-wrench"></i> wrench</a></li>
</ul>
</section>

<section id="file-type">
  <h2 class="doc-icon-hd">File Type Icons</h2>

  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/file"><i class="am-icon-file"></i> file</a></li>
    <li><a href="http://fontawesome.io/icon/file-archive-o"><i class="am-icon-file-archive-o"></i> file-archive-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-audio-o"><i class="am-icon-file-audio-o"></i> file-audio-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-code-o"><i class="am-icon-file-code-o"></i> file-code-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-excel-o"><i class="am-icon-file-excel-o"></i> file-excel-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-image-o"><i class="am-icon-file-image-o"></i> file-image-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-video-o"><i class="am-icon-file-movie-o"></i> file-movie-o <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/file-o"><i class="am-icon-file-o"></i> file-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-pdf-o"><i class="am-icon-file-pdf-o"></i> file-pdf-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-image-o"><i class="am-icon-file-photo-o"></i> file-photo-o <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/file-image-o"><i class="am-icon-file-picture-o"></i> file-picture-o <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/file-powerpoint-o"><i class="am-icon-file-powerpoint-o"></i> file-powerpoint-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-audio-o"><i class="am-icon-file-sound-o"></i> file-sound-o <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/file-text"><i class="am-icon-file-text"></i> file-text</a></li>
    <li><a href="http://fontawesome.io/icon/file-text-o"><i class="am-icon-file-text-o"></i> file-text-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-video-o"><i class="am-icon-file-video-o"></i> file-video-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-word-o"><i class="am-icon-file-word-o"></i> file-word-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-archive-o"><i class="am-icon-file-zip-o"></i> file-zip-o <span class="text-muted">(alias)</span></a></li>
  </ul>
</section>

<section id="spinner">
  <h2 class="doc-icon-hd">Spinner Icons</h2>

  <div class="am-alert am-alert-success">
    <i class="am-icon-info-circle lg li"></i>
    These icons work great with the <code>am-icon-spin</code> class. Check out the
    <a href="http://fontawesome.io/examples/#spinning" class="alert-link">spinning icons example</a>.
  </div>

    <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
      <li><a href="http://fontawesome.io/icon/circle-o-notch"><i class="am-icon-circle-o-notch"></i> circle-o-notch</a></li>
      <li><a href="http://fontawesome.io/icon/cog"><i class="am-icon-cog"></i> cog</a></li>
      <li><a href="http://fontawesome.io/icon/cog"><i class="am-icon-gear"></i> gear <span class="text-muted">(alias)</span></a></li>
      <li><a href="http://fontawesome.io/icon/refresh"><i class="am-icon-refresh"></i> refresh</a></li>
      <li><a href="http://fontawesome.io/icon/spinner"><i class="am-icon-spinner"></i> spinner</a></li>
    </ul>
</section>

<section id="form-control">
  <h2 class="doc-icon-hd">Form Control Icons</h2>

  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">

    <li><a href="http://fontawesome.io/icon/check-square"><i class="am-icon-check-square"></i> check-square</a></li>
    <li><a href="http://fontawesome.io/icon/check-square-o"><i class="am-icon-check-square-o"></i> check-square-o</a></li>
    <li><a href="http://fontawesome.io/icon/circle"><i class="am-icon-circle"></i> circle</a></li>
    <li><a href="http://fontawesome.io/icon/circle-o"><i class="am-icon-circle-o"></i> circle-o</a></li>
    <li><a href="http://fontawesome.io/icon/dot-circle-o"><i class="am-icon-dot-circle-o"></i> dot-circle-o</a></li>
    <li><a href="http://fontawesome.io/icon/minus-square"><i class="am-icon-minus-square"></i> minus-square</a></li>
    <li><a href="http://fontawesome.io/icon/minus-square-o"><i class="am-icon-minus-square-o"></i> minus-square-o</a></li>
    <li><a href="http://fontawesome.io/icon/plus-square"><i class="am-icon-plus-square"></i> plus-square</a></li>
    <li><a href="http://fontawesome.io/icon/plus-square-o"><i class="am-icon-plus-square-o"></i> plus-square-o</a></li>
    <li><a href="http://fontawesome.io/icon/square"><i class="am-icon-square"></i> square</a></li>
    <li><a href="http://fontawesome.io/icon/square-o"><i class="am-icon-square-o"></i> square-o</a></li>
  </ul>
</section>

<section id="payment">
  <h2 class="doc-icon-hd">Payment Icons</h2>

  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/cc-amex"><i class="am-icon-cc-amex"></i> cc-amex</a></li>
    <li><a href="http://fontawesome.io/icon/cc-discover"><i class="am-icon-cc-discover"></i> cc-discover</a></li>
    <li><a href="http://fontawesome.io/icon/cc-mastercard"><i class="am-icon-cc-mastercard"></i> cc-mastercard</a></li>
    <li><a href="http://fontawesome.io/icon/cc-paypal"><i class="am-icon-cc-paypal"></i> cc-paypal</a></li>
    <li><a href="http://fontawesome.io/icon/cc-stripe"><i class="am-icon-cc-stripe"></i> cc-stripe</a></li>
    <li><a href="http://fontawesome.io/icon/cc-visa"><i class="am-icon-cc-visa"></i> cc-visa</a></li>
    <li><a href="http://fontawesome.io/icon/credit-card"><i class="am-icon-credit-card"></i> credit-card</a></li>
    <li><a href="http://fontawesome.io/icon/google-wallet"><i class="am-icon-google-wallet"></i> google-wallet</a></li>
    <li><a href="http://fontawesome.io/icon/paypal"><i class="am-icon-paypal"></i> paypal</a></li>
  </ul>
</section>

<section id="chart">
  <h2 class="doc-icon-hd">Chart Icons</h2>
  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/area-chart"><i class="am-icon-area-chart"></i> area-chart</a></li>
    <li><a href="http://fontawesome.io/icon/bar-chart"><i class="am-icon-bar-chart"></i> bar-chart</a></li>
    <li><a href="http://fontawesome.io/icon/bar-chart"><i class="am-icon-bar-chart-o"></i> bar-chart-o <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/line-chart"><i class="am-icon-line-chart"></i> line-chart</a></li>
    <li><a href="http://fontawesome.io/icon/pie-chart"><i class="am-icon-pie-chart"></i> pie-chart</a></li>
  </ul>
</section>

<section id="currency">
  <h2 class="doc-icon-hd">Currency Icons</h2>

  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/btc"><i class="am-icon-bitcoin"></i> bitcoin <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/btc"><i class="am-icon-btc"></i> btc</a></li>
    <li><a href="http://fontawesome.io/icon/jpy"><i class="am-icon-cny"></i> cny <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/usd"><i class="am-icon-dollar"></i> dollar <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/eur"><i class="am-icon-eur"></i> eur</a></li>
    <li><a href="http://fontawesome.io/icon/eur"><i class="am-icon-euro"></i> euro <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/gbp"><i class="am-icon-gbp"></i> gbp</a></li>
    <li><a href="http://fontawesome.io/icon/ils"><i class="am-icon-ils"></i> ils</a></li>
    <li><a href="http://fontawesome.io/icon/inr"><i class="am-icon-inr"></i> inr</a></li>
    <li><a href="http://fontawesome.io/icon/jpy"><i class="am-icon-jpy"></i> jpy</a></li>
    <li><a href="http://fontawesome.io/icon/krw"><i class="am-icon-krw"></i> krw</a></li>
    <li><a href="http://fontawesome.io/icon/money"><i class="am-icon-money"></i> money</a></li>
    <li><a href="http://fontawesome.io/icon/jpy"><i class="am-icon-rmb"></i> rmb <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/rub"><i class="am-icon-rouble"></i> rouble <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/rub"><i class="am-icon-rub"></i> rub</a></li>
    <li><a href="http://fontawesome.io/icon/rub"><i class="am-icon-ruble"></i> ruble <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/inr"><i class="am-icon-rupee"></i> rupee <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/ils"><i class="am-icon-shekel"></i> shekel <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/ils"><i class="am-icon-sheqel"></i> sheqel <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/try"><i class="am-icon-try"></i> try</a></li>
    <li><a href="http://fontawesome.io/icon/try"><i class="am-icon-turkish-lira"></i> turkish-lira <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/usd"><i class="am-icon-usd"></i> usd</a></li>
    <li><a href="http://fontawesome.io/icon/krw"><i class="am-icon-won"></i> won <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/jpy"><i class="am-icon-yen"></i> yen <span class="text-muted">(alias)</span></a></li>
  </ul>
</section>

<section id="text-editor">
  <h2 class="doc-icon-hd">Text Editor Icons</h2>

  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/align-center"><i class="am-icon-align-center"></i> align-center</a></li>
    <li><a href="http://fontawesome.io/icon/align-justify"><i class="am-icon-align-justify"></i> align-justify</a></li>
    <li><a href="http://fontawesome.io/icon/align-left"><i class="am-icon-align-left"></i> align-left</a></li>
    <li><a href="http://fontawesome.io/icon/align-right"><i class="am-icon-align-right"></i> align-right</a></li>
    <li><a href="http://fontawesome.io/icon/bold"><i class="am-icon-bold"></i> bold</a></li>
    <li><a href="http://fontawesome.io/icon/link"><i class="am-icon-chain"></i> chain <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/chain-broken"><i class="am-icon-chain-broken"></i> chain-broken</a></li>
    <li><a href="http://fontawesome.io/icon/clipboard"><i class="am-icon-clipboard"></i> clipboard</a></li>
    <li><a href="http://fontawesome.io/icon/columns"><i class="am-icon-columns"></i> columns</a></li>
    <li><a href="http://fontawesome.io/icon/files-o"><i class="am-icon-copy"></i> copy <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/scissors"><i class="am-icon-cut"></i> cut <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/outdent"><i class="am-icon-dedent"></i> dedent <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/eraser"><i class="am-icon-eraser"></i> eraser</a></li>
    <li><a href="http://fontawesome.io/icon/file"><i class="am-icon-file"></i> file</a></li>
    <li><a href="http://fontawesome.io/icon/file-o"><i class="am-icon-file-o"></i> file-o</a></li>
    <li><a href="http://fontawesome.io/icon/file-text"><i class="am-icon-file-text"></i> file-text</a></li>
    <li><a href="http://fontawesome.io/icon/file-text-o"><i class="am-icon-file-text-o"></i> file-text-o</a></li>
    <li><a href="http://fontawesome.io/icon/files-o"><i class="am-icon-files-o"></i> files-o</a></li>
    <li><a href="http://fontawesome.io/icon/floppy-o"><i class="am-icon-floppy-o"></i> floppy-o</a></li>
    <li><a href="http://fontawesome.io/icon/font"><i class="am-icon-font"></i> font</a></li>
    <li><a href="http://fontawesome.io/icon/header"><i class="am-icon-header"></i> header</a></li>
    <li><a href="http://fontawesome.io/icon/indent"><i class="am-icon-indent"></i> indent</a></li>
    <li><a href="http://fontawesome.io/icon/italic"><i class="am-icon-italic"></i> italic</a></li>
    <li><a href="http://fontawesome.io/icon/link"><i class="am-icon-link"></i> link</a></li>
    <li><a href="http://fontawesome.io/icon/list"><i class="am-icon-list"></i> list</a></li>
    <li><a href="http://fontawesome.io/icon/list-alt"><i class="am-icon-list-alt"></i> list-alt</a></li>
    <li><a href="http://fontawesome.io/icon/list-ol"><i class="am-icon-list-ol"></i> list-ol</a></li>
    <li><a href="http://fontawesome.io/icon/list-ul"><i class="am-icon-list-ul"></i> list-ul</a></li>
    <li><a href="http://fontawesome.io/icon/outdent"><i class="am-icon-outdent"></i> outdent</a></li>
    <li><a href="http://fontawesome.io/icon/paperclip"><i class="am-icon-paperclip"></i> paperclip</a></li>
    <li><a href="http://fontawesome.io/icon/paragraph"><i class="am-icon-paragraph"></i> paragraph</a></li>
    <li><a href="http://fontawesome.io/icon/clipboard"><i class="am-icon-paste"></i> paste <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/repeat"><i class="am-icon-repeat"></i> repeat</a></li>
    <li><a href="http://fontawesome.io/icon/undo"><i class="am-icon-rotate-left"></i> rotate-left <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/repeat"><i class="am-icon-rotate-right"></i> rotate-right <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/floppy-o"><i class="am-icon-save"></i> save <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/scissors"><i class="am-icon-scissors"></i> scissors</a></li>
    <li><a href="http://fontawesome.io/icon/strikethrough"><i class="am-icon-strikethrough"></i> strikethrough</a></li>
    <li><a href="http://fontawesome.io/icon/subscript"><i class="am-icon-subscript"></i> subscript</a></li>
    <li><a href="http://fontawesome.io/icon/superscript"><i class="am-icon-superscript"></i> superscript</a></li>
    <li><a href="http://fontawesome.io/icon/table"><i class="am-icon-table"></i> table</a></li>
    <li><a href="http://fontawesome.io/icon/text-height"><i class="am-icon-text-height"></i> text-height</a></li>
    <li><a href="http://fontawesome.io/icon/text-width"><i class="am-icon-text-width"></i> text-width</a></li>
    <li><a href="http://fontawesome.io/icon/th"><i class="am-icon-th"></i> th</a></li>
    <li><a href="http://fontawesome.io/icon/th-large"><i class="am-icon-th-large"></i> th-large</a></li>
    <li><a href="http://fontawesome.io/icon/th-list"><i class="am-icon-th-list"></i> th-list</a></li>
    <li><a href="http://fontawesome.io/icon/underline"><i class="am-icon-underline"></i> underline</a></li>
    <li><a href="http://fontawesome.io/icon/undo"><i class="am-icon-undo"></i> undo</a></li>
    <li><a href="http://fontawesome.io/icon/chain-broken"><i class="am-icon-unlink"></i> unlink <span class="text-muted">(alias)</span></a></li>
  </ul>
</section>

<section id="directional">
  <h2 class="doc-icon-hd">Directional Icons</h2>

  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/angle-double-down"><i class="am-icon-angle-double-down"></i> angle-double-down</a></li>
    <li><a href="http://fontawesome.io/icon/angle-double-left"><i class="am-icon-angle-double-left"></i> angle-double-left</a></li>
    <li><a href="http://fontawesome.io/icon/angle-double-right"><i class="am-icon-angle-double-right"></i> angle-double-right</a></li>
    <li><a href="http://fontawesome.io/icon/angle-double-up"><i class="am-icon-angle-double-up"></i> angle-double-up</a></li>
    <li><a href="http://fontawesome.io/icon/angle-down"><i class="am-icon-angle-down"></i> angle-down</a></li>
    <li><a href="http://fontawesome.io/icon/angle-left"><i class="am-icon-angle-left"></i> angle-left</a></li>
    <li><a href="http://fontawesome.io/icon/angle-right"><i class="am-icon-angle-right"></i> angle-right</a></li>
    <li><a href="http://fontawesome.io/icon/angle-up"><i class="am-icon-angle-up"></i> angle-up</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-circle-down"><i class="am-icon-arrow-circle-down"></i> arrow-circle-down</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-circle-left"><i class="am-icon-arrow-circle-left"></i> arrow-circle-left</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-circle-o-down"><i class="am-icon-arrow-circle-o-down"></i> arrow-circle-o-down</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-circle-o-left"><i class="am-icon-arrow-circle-o-left"></i> arrow-circle-o-left</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-circle-o-right"><i class="am-icon-arrow-circle-o-right"></i> arrow-circle-o-right</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-circle-o-up"><i class="am-icon-arrow-circle-o-up"></i> arrow-circle-o-up</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-circle-right"><i class="am-icon-arrow-circle-right"></i> arrow-circle-right</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-circle-up"><i class="am-icon-arrow-circle-up"></i> arrow-circle-up</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-down"><i class="am-icon-arrow-down"></i> arrow-down</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-left"><i class="am-icon-arrow-left"></i> arrow-left</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-right"><i class="am-icon-arrow-right"></i> arrow-right</a></li>
    <li><a href="http://fontawesome.io/icon/arrow-up"><i class="am-icon-arrow-up"></i> arrow-up</a></li>
    <li><a href="http://fontawesome.io/icon/arrows"><i class="am-icon-arrows"></i> arrows</a></li>
    <li><a href="http://fontawesome.io/icon/arrows-alt"><i class="am-icon-arrows-alt"></i> arrows-alt</a></li>
    <li><a href="http://fontawesome.io/icon/arrows-h"><i class="am-icon-arrows-h"></i> arrows-h</a></li>
    <li><a href="http://fontawesome.io/icon/arrows-v"><i class="am-icon-arrows-v"></i> arrows-v</a></li>
    <li><a href="http://fontawesome.io/icon/caret-down"><i class="am-icon-caret-down"></i> caret-down</a></li>
    <li><a href="http://fontawesome.io/icon/caret-left"><i class="am-icon-caret-left"></i> caret-left</a></li>
    <li><a href="http://fontawesome.io/icon/caret-right"><i class="am-icon-caret-right"></i> caret-right</a></li>
    <li><a href="http://fontawesome.io/icon/caret-square-o-down"><i class="am-icon-caret-square-o-down"></i> caret-square-o-down</a></li>
    <li><a href="http://fontawesome.io/icon/caret-square-o-left"><i class="am-icon-caret-square-o-left"></i> caret-square-o-left</a></li>
    <li><a href="http://fontawesome.io/icon/caret-square-o-right"><i class="am-icon-caret-square-o-right"></i> caret-square-o-right</a></li>
    <li><a href="http://fontawesome.io/icon/caret-square-o-up"><i class="am-icon-caret-square-o-up"></i> caret-square-o-up</a></li>
    <li><a href="http://fontawesome.io/icon/caret-up"><i class="am-icon-caret-up"></i> caret-up</a></li>
    <li><a href="http://fontawesome.io/icon/chevron-circle-down"><i class="am-icon-chevron-circle-down"></i> chevron-circle-down</a></li>
    <li><a href="http://fontawesome.io/icon/chevron-circle-left"><i class="am-icon-chevron-circle-left"></i> chevron-circle-left</a></li>
    <li><a href="http://fontawesome.io/icon/chevron-circle-right"><i class="am-icon-chevron-circle-right"></i> chevron-circle-right</a></li>
    <li><a href="http://fontawesome.io/icon/chevron-circle-up"><i class="am-icon-chevron-circle-up"></i> chevron-circle-up</a></li>
    <li><a href="http://fontawesome.io/icon/chevron-down"><i class="am-icon-chevron-down"></i> chevron-down</a></li>
    <li><a href="http://fontawesome.io/icon/chevron-left"><i class="am-icon-chevron-left"></i> chevron-left</a></li>
    <li><a href="http://fontawesome.io/icon/chevron-right"><i class="am-icon-chevron-right"></i> chevron-right</a></li>
    <li><a href="http://fontawesome.io/icon/chevron-up"><i class="am-icon-chevron-up"></i> chevron-up</a></li>
    <li><a href="http://fontawesome.io/icon/hand-o-down"><i class="am-icon-hand-o-down"></i> hand-o-down</a></li>
    <li><a href="http://fontawesome.io/icon/hand-o-left"><i class="am-icon-hand-o-left"></i> hand-o-left</a></li>
    <li><a href="http://fontawesome.io/icon/hand-o-right"><i class="am-icon-hand-o-right"></i> hand-o-right</a></li>
    <li><a href="http://fontawesome.io/icon/hand-o-up"><i class="am-icon-hand-o-up"></i> hand-o-up</a></li>
    <li><a href="http://fontawesome.io/icon/long-arrow-down"><i class="am-icon-long-arrow-down"></i> long-arrow-down</a></li>
    <li><a href="http://fontawesome.io/icon/long-arrow-left"><i class="am-icon-long-arrow-left"></i> long-arrow-left</a></li>
    <li><a href="http://fontawesome.io/icon/long-arrow-right"><i class="am-icon-long-arrow-right"></i> long-arrow-right</a></li>
    <li><a href="http://fontawesome.io/icon/long-arrow-up"><i class="am-icon-long-arrow-up"></i> long-arrow-up</a></li>
    <li><a href="http://fontawesome.io/icon/caret-square-o-down"><i class="am-icon-toggle-down"></i> toggle-down <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/caret-square-o-left"><i class="am-icon-toggle-left"></i> toggle-left <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/caret-square-o-right"><i class="am-icon-toggle-right"></i> toggle-right <span class="text-muted">(alias)</span></a></li>
    <li><a href="http://fontawesome.io/icon/caret-square-o-up"><i class="am-icon-toggle-up"></i> toggle-up <span class="text-muted">(alias)</span></a></li>
  </ul>
</section>

<section id="video-player">
  <h2 class="doc-icon-hd">Video Player Icons</h2>

  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/arrows-alt"><i class="am-icon-arrows-alt"></i> arrows-alt</a></li>
    <li><a href="http://fontawesome.io/icon/backward"><i class="am-icon-backward"></i> backward</a></li>
    <li><a href="http://fontawesome.io/icon/compress"><i class="am-icon-compress"></i> compress</a></li>
    <li><a href="http://fontawesome.io/icon/eject"><i class="am-icon-eject"></i> eject</a></li>
    <li><a href="http://fontawesome.io/icon/expand"><i class="am-icon-expand"></i> expand</a></li>
    <li><a href="http://fontawesome.io/icon/fast-backward"><i class="am-icon-fast-backward"></i> fast-backward</a></li>
    <li><a href="http://fontawesome.io/icon/fast-forward"><i class="am-icon-fast-forward"></i> fast-forward</a></li>
    <li><a href="http://fontawesome.io/icon/forward"><i class="am-icon-forward"></i> forward</a></li>
    <li><a href="http://fontawesome.io/icon/pause"><i class="am-icon-pause"></i> pause</a></li>
    <li><a href="http://fontawesome.io/icon/play"><i class="am-icon-play"></i> play</a></li>
    <li><a href="http://fontawesome.io/icon/play-circle"><i class="am-icon-play-circle"></i> play-circle</a></li>
    <li><a href="http://fontawesome.io/icon/play-circle-o"><i class="am-icon-play-circle-o"></i> play-circle-o</a></li>
    <li><a href="http://fontawesome.io/icon/step-backward"><i class="am-icon-step-backward"></i> step-backward</a></li>
    <li><a href="http://fontawesome.io/icon/step-forward"><i class="am-icon-step-forward"></i> step-forward</a></li>
    <li><a href="http://fontawesome.io/icon/stop"><i class="am-icon-stop"></i> stop</a></li>
    <li><a href="http://fontawesome.io/icon/youtube-play"><i class="am-icon-youtube-play"></i> youtube-play</a></li>
  </ul>
</section>

<section id="brand">
  <h2 class="doc-icon-hd">Brand Icons</h2>

  <div class="am-alert am-alert-success">
    <ul class="margin-bottom-none padding-left-lg">
      <li>All brand icons are trademarks of their respective owners.</li>
      <li>The use of these trademarks does not indicate endorsement of the trademark holder by Font Awesome, nor vice
        versa.
      </li>
    </ul>
  </div>

  <div class="am-alert am-alert-warning">
    <h4><i class="am-icon-warning"></i> Warning!</h4>
    Apparently, Adblock Plus can remove Font Awesome brand icons with their "Remove Social
    Media Buttons" setting. We will not use hacks to force them to display. Please
    <a href="https://adblockplus.org/en/bugs" class="alert-link">report an issue with Adblock Plus</a> if you believe
    this to be
    an error. To work around this, you'll need to modify the social icon class names.
  </div>

      <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
        <li><a href="http://fontawesome.io/icon/adn"><i class="am-icon-adn"></i> adn</a></li>
        <li><a href="http://fontawesome.io/icon/android"><i class="am-icon-android"></i> android</a></li>
        <li><a href="http://fontawesome.io/icon/angellist"><i class="am-icon-angellist"></i> angellist</a></li>
        <li><a href="http://fontawesome.io/icon/apple"><i class="am-icon-apple"></i> apple</a></li>
        <li><a href="http://fontawesome.io/icon/behance"><i class="am-icon-behance"></i> behance</a></li>
        <li><a href="http://fontawesome.io/icon/behance-square"><i class="am-icon-behance-square"></i> behance-square</a></li>
        <li><a href="http://fontawesome.io/icon/bitbucket"><i class="am-icon-bitbucket"></i> bitbucket</a></li>
        <li><a href="http://fontawesome.io/icon/bitbucket-square"><i class="am-icon-bitbucket-square"></i> bitbucket-square</a></li>
        <li><a href="http://fontawesome.io/icon/btc"><i class="am-icon-bitcoin"></i> bitcoin <span class="text-muted">(alias)</span></a></li>
        <li><a href="http://fontawesome.io/icon/btc"><i class="am-icon-btc"></i> btc</a></li>
        <li><a href="http://fontawesome.io/icon/cc-amex"><i class="am-icon-cc-amex"></i> cc-amex</a></li>
        <li><a href="http://fontawesome.io/icon/cc-discover"><i class="am-icon-cc-discover"></i> cc-discover</a></li>
        <li><a href="http://fontawesome.io/icon/cc-mastercard"><i class="am-icon-cc-mastercard"></i> cc-mastercard</a></li>
        <li><a href="http://fontawesome.io/icon/cc-paypal"><i class="am-icon-cc-paypal"></i> cc-paypal</a></li>
        <li><a href="http://fontawesome.io/icon/cc-stripe"><i class="am-icon-cc-stripe"></i> cc-stripe</a></li>
        <li><a href="http://fontawesome.io/icon/cc-visa"><i class="am-icon-cc-visa"></i> cc-visa</a></li>
        <li><a href="http://fontawesome.io/icon/codepen"><i class="am-icon-codepen"></i> codepen</a></li>
        <li><a href="http://fontawesome.io/icon/css3"><i class="am-icon-css3"></i> css3</a></li>
        <li><a href="http://fontawesome.io/icon/delicious"><i class="am-icon-delicious"></i> delicious</a></li>
        <li><a href="http://fontawesome.io/icon/deviantart"><i class="am-icon-deviantart"></i> deviantart</a></li>
        <li><a href="http://fontawesome.io/icon/digg"><i class="am-icon-digg"></i> digg</a></li>
        <li><a href="http://fontawesome.io/icon/dribbble"><i class="am-icon-dribbble"></i> dribbble</a></li>
        <li><a href="http://fontawesome.io/icon/dropbox"><i class="am-icon-dropbox"></i> dropbox</a></li>
        <li><a href="http://fontawesome.io/icon/drupal"><i class="am-icon-drupal"></i> drupal</a></li>
        <li><a href="http://fontawesome.io/icon/empire"><i class="am-icon-empire"></i> empire</a></li>
        <li><a href="http://fontawesome.io/icon/facebook"><i class="am-icon-facebook"></i> facebook</a></li>
        <li><a href="http://fontawesome.io/icon/facebook-square"><i class="am-icon-facebook-square"></i> facebook-square</a></li>
        <li><a href="http://fontawesome.io/icon/flickr"><i class="am-icon-flickr"></i> flickr</a></li>
        <li><a href="http://fontawesome.io/icon/foursquare"><i class="am-icon-foursquare"></i> foursquare</a></li>
        <li><a href="http://fontawesome.io/icon/empire"><i class="am-icon-ge"></i> ge <span class="text-muted">(alias)</span></a></li>
        <li><a href="http://fontawesome.io/icon/git"><i class="am-icon-git"></i> git</a></li>
        <li><a href="http://fontawesome.io/icon/git-square"><i class="am-icon-git-square"></i> git-square</a></li>
        <li><a href="http://fontawesome.io/icon/github"><i class="am-icon-github"></i> github</a></li>
        <li><a href="http://fontawesome.io/icon/github-alt"><i class="am-icon-github-alt"></i> github-alt</a></li>
        <li><a href="http://fontawesome.io/icon/github-square"><i class="am-icon-github-square"></i> github-square</a></li>
        <li><a href="http://fontawesome.io/icon/gittip"><i class="am-icon-gittip"></i> gittip</a></li>
        <li><a href="http://fontawesome.io/icon/google"><i class="am-icon-google"></i> google</a></li>
        <li><a href="http://fontawesome.io/icon/google-plus"><i class="am-icon-google-plus"></i> google-plus</a></li>
        <li><a href="http://fontawesome.io/icon/google-plus-square"><i class="am-icon-google-plus-square"></i> google-plus-square</a></li>
        <li><a href="http://fontawesome.io/icon/google-wallet"><i class="am-icon-google-wallet"></i> google-wallet</a></li>
        <li><a href="http://fontawesome.io/icon/hacker-news"><i class="am-icon-hacker-news"></i> hacker-news</a></li>
        <li><a href="http://fontawesome.io/icon/html5"><i class="am-icon-html5"></i> html5</a></li>
        <li><a href="http://fontawesome.io/icon/instagram"><i class="am-icon-instagram"></i> instagram</a></li>
        <li><a href="http://fontawesome.io/icon/ioxhost"><i class="am-icon-ioxhost"></i> ioxhost</a></li>
        <li><a href="http://fontawesome.io/icon/joomla"><i class="am-icon-joomla"></i> joomla</a></li>
        <li><a href="http://fontawesome.io/icon/jsfiddle"><i class="am-icon-jsfiddle"></i> jsfiddle</a></li>
        <li><a href="http://fontawesome.io/icon/lastfm"><i class="am-icon-lastfm"></i> lastfm</a></li>
        <li><a href="http://fontawesome.io/icon/lastfm-square"><i class="am-icon-lastfm-square"></i> lastfm-square</a></li>
        <li><a href="http://fontawesome.io/icon/linkedin"><i class="am-icon-linkedin"></i> linkedin</a></li>
        <li><a href="http://fontawesome.io/icon/linkedin-square"><i class="am-icon-linkedin-square"></i> linkedin-square</a></li>
        <li><a href="http://fontawesome.io/icon/linux"><i class="am-icon-linux"></i> linux</a></li>
        <li><a href="http://fontawesome.io/icon/maxcdn"><i class="am-icon-maxcdn"></i> maxcdn</a></li>
        <li><a href="http://fontawesome.io/icon/meanpath"><i class="am-icon-meanpath"></i> meanpath</a></li>
        <li><a href="http://fontawesome.io/icon/openid"><i class="am-icon-openid"></i> openid</a></li>
        <li><a href="http://fontawesome.io/icon/pagelines"><i class="am-icon-pagelines"></i> pagelines</a></li>
        <li><a href="http://fontawesome.io/icon/paypal"><i class="am-icon-paypal"></i> paypal</a></li>
        <li><a href="http://fontawesome.io/icon/pied-piper"><i class="am-icon-pied-piper"></i> pied-piper</a></li>
        <li><a href="http://fontawesome.io/icon/pied-piper-alt"><i class="am-icon-pied-piper-alt"></i> pied-piper-alt</a></li>
        <li><a href="http://fontawesome.io/icon/pinterest"><i class="am-icon-pinterest"></i> pinterest</a></li>
        <li><a href="http://fontawesome.io/icon/pinterest-square"><i class="am-icon-pinterest-square"></i> pinterest-square</a></li>
        <li><a href="http://fontawesome.io/icon/qq"><i class="am-icon-qq"></i> qq</a></li>
        <li><a href="http://fontawesome.io/icon/rebel"><i class="am-icon-ra"></i> ra <span class="text-muted">(alias)</span></a></li>
        <li><a href="http://fontawesome.io/icon/rebel"><i class="am-icon-rebel"></i> rebel</a></li>
        <li><a href="http://fontawesome.io/icon/reddit"><i class="am-icon-reddit"></i> reddit</a></li>
        <li><a href="http://fontawesome.io/icon/reddit-square"><i class="am-icon-reddit-square"></i> reddit-square</a></li>
        <li><a href="http://fontawesome.io/icon/renren"><i class="am-icon-renren"></i> renren</a></li>
        <li><a href="http://fontawesome.io/icon/share-alt"><i class="am-icon-share-alt"></i> share-alt</a></li>
        <li><a href="http://fontawesome.io/icon/share-alt-square"><i class="am-icon-share-alt-square"></i> share-alt-square</a></li>
        <li><a href="http://fontawesome.io/icon/skype"><i class="am-icon-skype"></i> skype</a></li>
        <li><a href="http://fontawesome.io/icon/slack"><i class="am-icon-slack"></i> slack</a></li>
        <li><a href="http://fontawesome.io/icon/slideshare"><i class="am-icon-slideshare"></i> slideshare</a></li>
        <li><a href="http://fontawesome.io/icon/soundcloud"><i class="am-icon-soundcloud"></i> soundcloud</a></li>
        <li><a href="http://fontawesome.io/icon/spotify"><i class="am-icon-spotify"></i> spotify</a></li>
        <li><a href="http://fontawesome.io/icon/stack-exchange"><i class="am-icon-stack-exchange"></i> stack-exchange</a></li>
        <li><a href="http://fontawesome.io/icon/stack-overflow"><i class="am-icon-stack-overflow"></i> stack-overflow</a></li>
        <li><a href="http://fontawesome.io/icon/steam"><i class="am-icon-steam"></i> steam</a></li>
        <li><a href="http://fontawesome.io/icon/steam-square"><i class="am-icon-steam-square"></i> steam-square</a></li>
        <li><a href="http://fontawesome.io/icon/stumbleupon"><i class="am-icon-stumbleupon"></i> stumbleupon</a></li>
        <li><a href="http://fontawesome.io/icon/stumbleupon-circle"><i class="am-icon-stumbleupon-circle"></i> stumbleupon-circle</a></li>
        <li><a href="http://fontawesome.io/icon/tencent-weibo"><i class="am-icon-tencent-weibo"></i> tencent-weibo</a></li>
        <li><a href="http://fontawesome.io/icon/trello"><i class="am-icon-trello"></i> trello</a></li>
        <li><a href="http://fontawesome.io/icon/tumblr"><i class="am-icon-tumblr"></i> tumblr</a></li>
        <li><a href="http://fontawesome.io/icon/tumblr-square"><i class="am-icon-tumblr-square"></i> tumblr-square</a></li>
        <li><a href="http://fontawesome.io/icon/twitch"><i class="am-icon-twitch"></i> twitch</a></li>
        <li><a href="http://fontawesome.io/icon/twitter"><i class="am-icon-twitter"></i> twitter</a></li>
        <li><a href="http://fontawesome.io/icon/twitter-square"><i class="am-icon-twitter-square"></i> twitter-square</a></li>
        <li><a href="http://fontawesome.io/icon/vimeo-square"><i class="am-icon-vimeo-square"></i> vimeo-square</a></li>
        <li><a href="http://fontawesome.io/icon/vine"><i class="am-icon-vine"></i> vine</a></li>
        <li><a href="http://fontawesome.io/icon/vk"><i class="am-icon-vk"></i> vk</a></li>
        <li><a href="http://fontawesome.io/icon/weixin"><i class="am-icon-wechat"></i> wechat <span class="text-muted">(alias)</span></a></li>
        <li><a href="http://fontawesome.io/icon/weibo"><i class="am-icon-weibo"></i> weibo</a></li>
        <li><a href="http://fontawesome.io/icon/weixin"><i class="am-icon-weixin"></i> weixin</a></li>
        <li><a href="http://fontawesome.io/icon/windows"><i class="am-icon-windows"></i> windows</a></li>
        <li><a href="http://fontawesome.io/icon/wordpress"><i class="am-icon-wordpress"></i> wordpress</a></li>
        <li><a href="http://fontawesome.io/icon/xing"><i class="am-icon-xing"></i> xing</a></li>
        <li><a href="http://fontawesome.io/icon/xing-square"><i class="am-icon-xing-square"></i> xing-square</a></li>
        <li><a href="http://fontawesome.io/icon/yahoo"><i class="am-icon-yahoo"></i> yahoo</a></li>
        <li><a href="http://fontawesome.io/icon/yelp"><i class="am-icon-yelp"></i> yelp</a></li>
        <li><a href="http://fontawesome.io/icon/youtube"><i class="am-icon-youtube"></i> youtube</a></li>
        <li><a href="http://fontawesome.io/icon/youtube-play"><i class="am-icon-youtube-play"></i> youtube-play</a></li>
        <li><a href="http://fontawesome.io/icon/youtube-square"><i class="am-icon-youtube-square"></i> youtube-square</a></li>
      </ul>
</section>

<section id="medical">
  <h2 class="doc-icon-hd">Medical Icons</h2>

  <ul class="doc-icon-list am-avg-sm-2 am-avg-md-3 am-avg-lg-4">
    <li><a href="http://fontawesome.io/icon/ambulance"><i class="am-icon-ambulance"></i> ambulance</a></li>
    <li><a href="http://fontawesome.io/icon/h-square"><i class="am-icon-h-square"></i> h-square</a></li>
    <li><a href="http://fontawesome.io/icon/hospital-o"><i class="am-icon-hospital-o"></i> hospital-o</a></li>
    <li><a href="http://fontawesome.io/icon/medkit"><i class="am-icon-medkit"></i> medkit</a></li>
    <li><a href="http://fontawesome.io/icon/plus-square"><i class="am-icon-plus-square"></i> plus-square</a></li>
    <li><a href="http://fontawesome.io/icon/stethoscope"><i class="am-icon-stethoscope"></i> stethoscope</a></li>
    <li><a href="http://fontawesome.io/icon/user-md"><i class="am-icon-user-md"></i> user-md</a></li>
    <li><a href="http://fontawesome.io/icon/wheelchair"><i class="am-icon-wheelchair"></i> wheelchair</a></li>
  </ul>
</section>

<script>
$(function() {
  if (!window.ZeroClipboard) return;

  $('.doc-icon-list li').each(function(index, item) {
    $(item).find('a').on('click', function(e) {
      e.preventDefault();
    });

    $(item).on('mouseenter', function() {
      $(item).addClass('copy-active');
    }).on('mouseleave', function() {
      setTimeout(function() {
        $(item).removeClass('copy-active');
      }, 3000);
    });
    var iconClass = $(this).find('i').attr('class'),
        iconName = iconClass.replace('am-icon-', '');

    var style = '{ \n';
    style += '  .am-icon-font; \n';
    style += '  content: @fa-var-' + iconName + ';\n';
    style += '}';

    var iconAction = '<span class="demo-icon-actions"><i class="am-badge am-badge-primary demo-copy-icon" data-clipboard-text="' + iconClass + '" title="复制 Icon Class">class</i><i class="am-badge am-badge-success demo-copy-icon" data-clipboard-text="' + style + '" title="复制 Icon Style">style</i>';
    $(this).append($(iconAction));
  });

  var clip = new ZeroClipboard($('.demo-copy-icon'));

  clip.on('aftercopy', function(e) {
    console && console.log("Copied to clipboard:\n" + e.data['text/plain']);
  });
});
</script>
`````
