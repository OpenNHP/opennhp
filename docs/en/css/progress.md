# Progress
---

Defines different styles for progress bars. Use `.am-progress` as container and `.am-progress-bar` to display information

## Default Style

`````html
<div class="am-progress">
  <div class="am-progress-bar" style="width: 30%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar" style="width: 40%">40%</div>
</div>
`````

```html
<div class="am-progress">
  <div class="am-progress-bar" style="width: 30%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar" style="width: 40%">40%</div>
</div>
```
    
## Color

Default color is the primary color. Use following classes to change the color:

- `.am-progress-bar-secondary`
- `.am-progress-bar-success`
- `.am-progress-bar-warning`
- `.am-progress-bar-danger`

`````html
<div class="am-progress">
  <div class="am-progress-bar" style="width: 15%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar am-progress-bar-secondary" style="width: 30%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar am-progress-bar-success" style="width: 45%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar am-progress-bar-warning" style="width: 60%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar am-progress-bar-danger" style="width: 75%"></div>
</div>
`````
```html
<div class="am-progress">
  <div class="am-progress-bar" style="width: 15%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar am-progress-bar-secondary" style="width: 30%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar am-progress-bar-success" style="width: 45%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar am-progress-bar-warning" style="width: 60%"></div>
</div>

<div class="am-progress">
  <div class="am-progress-bar am-progress-bar-danger" style="width: 75%"></div>
</div>
```

## Height

Add `.am-progress-xs` or `.am-progress-sm` to set the height of progress bar.

`````html
<div class="am-progress am-progress-xs">
    <div class="am-progress-bar" style="width: 80%"></div>
</div>

<div class="am-progress am-progress-sm">
    <div class="am-progress-bar" style="width: 60%"></div>
</div>

<div class="am-progress">
    <div class="am-progress-bar" style="width: 40%"></div>
</div>
`````

```html
<div class="am-progress am-progress-xs">
    <div class="am-progress-bar" style="width: 80%"></div>
</div>

<div class="am-progress am-progress-sm">
    <div class="am-progress-bar" style="width: 60%"></div>
</div>

<div class="am-progress">
    <div class="am-progress-bar" style="width: 40%"></div>
</div>
```

## Striped

To create a striped progress bar, use the `.uk-progress-striped class`. Combinable with with color modifier.

`````html
<div class="am-progress am-progress-striped">
  <div class="am-progress-bar am-progress-bar-danger" style="width: 80%"></div>
</div>
<div class="am-progress am-progress-striped">
  <div class="am-progress-bar am-progress-bar-warning" style="width: 60%"></div>
</div>
<div class="am-progress am-progress-striped">
  <div class="am-progress-bar am-progress-bar-success" style="width: 45%"></div>
</div>
<div class="am-progress am-progress-striped">
  <div class="am-progress-bar am-progress-bar-secondary" style="width: 30%"></div>
</div>

<div class="am-progress am-progress-striped">
  <div class="am-progress-bar" style="width: 15%"></div>
</div>
`````
```html
<div class="am-progress am-progress-striped">
  <div class="am-progress-bar am-progress-bar-danger" style="width: 80%"></div>
</div>
<div class="am-progress am-progress-striped">
  <div class="am-progress-bar am-progress-bar-warning" style="width: 60%"></div>
</div>
<div class="am-progress am-progress-striped">
  <div class="am-progress-bar am-progress-bar-success" style="width: 45%"></div>
</div>
<div class="am-progress am-progress-striped">
  <div class="am-progress-bar am-progress-bar-secondary" style="width: 30%"></div>
</div>

<div class="am-progress am-progress-striped">
  <div class="am-progress-bar" style="width: 15%"></div>
</div>
```

## Animation

Use `.am-active` to active the progess bar animation.

`````html
<div class="am-progress am-progress-striped am-progress-sm am-active ">
  <div class="am-progress-bar am-progress-bar-secondary"  style="width: 57%"></div>
</div>
`````
```html
<div class="am-progress am-progress-striped am-progress-sm am-active ">
  <div class="am-progress-bar am-progress-bar-secondary"  style="width: 57%"></div>
</div>
```

## Percentage Bar

`````html
<div class="am-progress">
  <div class="am-progress-bar"  style="width: 65%">Male</div>
  <div class="am-progress-bar am-progress-bar-success"  style="width: 15%">Female</div>
  <div class="am-progress-bar am-progress-bar-warning"  style="width: 20%">Other</div>
</div>
`````
```html
<div class="am-progress">
  <div class="am-progress-bar"  style="width: 65%">Male</div>
  <div class="am-progress-bar am-progress-bar-success"  style="width: 15%">Female</div>
  <div class="am-progress-bar am-progress-bar-warning"  style="width: 20%">Other</div>
</div>
```
