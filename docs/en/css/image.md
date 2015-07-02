# Image
---

Define the styles for image。

## Default Style

Default style is defined in `base`.

```css
img {
  box-sizing: border-box;
  /* Codes below is removed in v2.3, find more detail in issue #502 */
  /* max-width: 100%;
  height: auto;*/
  vertical-align: middle;
  border: 0;
}
```

## Responsive Image

Max width has been removed in `v2.3` to solve [#502](https://github.com/allmobilize/amazeui/issues/502). New class `.am-img-responsive` is added. 

`````html
<img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg" class="am-img-responsive" alt=""/>
`````
```html
<img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg" class="am-img-responsive" alt=""/>
```

## Enhanced Styles

### Round Cornor

Use following classes to give round cornor to `<img>`.

- `.am-radius`     Round Cornor
- `.am-round`      Ellipse
- `.am-circle`     Circle

`````html
<p><img class="am-radius" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />
<img class="am-round" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/600/q/80" width="200" height="120"/>
<img class="am-circle" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140"/></p>
`````
```html
<p>
  <img class="am-radius" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />

  <img class="am-round" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/600/q/80" width="200" height="120"/>

  <img class="am-circle" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140"/>
</p>
```


### Border

- `.am-img-thumbnail`   Border

`````html
<img class="am-img-thumbnail" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />

<img class="am-img-thumbnail am-radius" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />

<img class="am-img-thumbnail am-circle" alt="140*140" src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/1000/h/1000/q/80" width="140" height="140" />
`````

```html
<img src="..." alt="..." class="am-img-thumbnail">
<img src="..." alt="..." class="am-img-thumbnail am-radius">
<img src="..." alt="..." class="am-img-thumbnail am-circle">
```

<!--
## Responsive Images

Add the `.am-img-responsive` class to scale the image automatically.

`````html
<img class="am-img-responsive" alt="Responsive image" src="http://www.bing.com/az/hprichbg/rb/AdelaideFrog_EN-US12171255358_1366x768.jpg" />
`````

```html
<img src="..." class="am-img-responsive" alt="Responsive image">
```-->
