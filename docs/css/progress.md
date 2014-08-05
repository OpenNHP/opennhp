# Progress
---

进度条组件，`.am-progress` 为容器，`.am-progress-bar` 为进度显示信息。

## 基本样式

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
    
## 进度条颜色

进度条默认为全局主色，在进度条上添加相应的类可设置的颜色：

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

## 进度条高度

在 `.am-progress` 添加 `.am-progress-xs` `.am-progress-sm` 可以设置进度条高度。

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

## 进度条条纹

在进度条容器上添加 `.am-progress-striped` 显示条纹效果，可结合进度条颜色 class 使用。

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

## 进度条动画

进度条容器上添加 `.am-active` 激活进度条动画（CSS Animation）。

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

## 进度条堆叠

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
