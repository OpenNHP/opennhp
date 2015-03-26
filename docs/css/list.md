# List
---

## 基本样式

### 链接列表

使用 `<ul>` 结构嵌套链接列表，添加 `.am-list`。

`````html
<ul class="am-list">
  <li><a href="#">每个人都有一个死角， 自己走不出来，别人也闯不进去。</a></li>
  <li><a href="#">我把最深沉的秘密放在那里。</a></li>
  <li><a href="#">你不懂我，我不怪你。</a></li>
  <li><a href="#">每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</a></li>
</ul>
`````

```html
<ul class="am-list">
  <li><a href="#">每个人都有一个死角， 自己走不出来，别人也闯不进去。</a></li>
  <li><a href="#">我把最深沉的秘密放在那里。</a></li>
  <li><a href="#">你不懂我，我不怪你。</a></li>
  <li><a href="#">每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</a></li>
</ul>
```

### 文字超出截断为「...」

在 `<a>` 上添加 `.am-text-truncate` class 可以实现文字超出一行时截断为 `...`。

`````html
<ul class="am-list">
  <li><a href="#" class="am-text-truncate">每个人都有一个死角， 自己走不出来，别人也闯不进去。我把最深沉的秘密放在那里。每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</a></li>
</ul>
`````
```html
<ul class="am-list">
  <li><a href="#" class="am-text-truncate">每个人都有一个死角， 自己走不出来，别人也闯不进去。我把最深沉的秘密放在那里。每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</a></li>
</ul>
```



### 纯文字列表

在 `.am-list` 的基础上添加 `.am-list-static`。

`````html
<ul class="am-list am-list-static">
  <li>每个人都有一个死角， 自己走不出来，别人也闯不进去。</li>
  <li>我把最深沉的秘密放在那里。</li>
  <li>你不懂我，我不怪你。</li>
  <li>每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</li>
</ul>
`````
```html
<ul class="am-list am-list-static">
  <li>...</li>
</ul>
```

## 样式变换

### 列表边框

在容器上添加 `.am-list-border` class。

`````html
<ul class="am-list am-list-static am-list-border">
  <li>每个人都有一个死角， 自己走不出来，别人也闯不进去。</li>
  <li>我把最深沉的秘密放在那里。</li>
  <li>你不懂我，我不怪你。</li>
  <li>每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</li>
</ul>

<ul class="am-list am-list-border">
  <li><a href="#">每个人都有一个死角， 自己走不出来，别人也闯不进去。</a></li>
  <li><a href="#">我把最深沉的秘密放在那里。</a></li>
  <li><a href="#">你不懂我，我不怪你。</a></li>
  <li><a href="#">每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</a></li>
</ul>
`````

```html
<ul class="am-list am-list-static am-list-border">
  <li>...</li>
</ul>

<ul class="am-list am-list-border">
  <li>...</li>
</ul>
```

### 斑马纹

添加 `.am-list-striped` class。

`````html
<ul class="am-list am-list-static am-list-border am-list-striped">
  <li>每个人都有一个死角， 自己走不出来，别人也闯不进去。</li>
  <li>我把最深沉的秘密放在那里。</li>
  <li>你不懂我，我不怪你。</li>
  <li>每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</li>
</ul>
`````
```html
<ul class="am-list am-list-static am-list-border am-list-striped">
  <li>...</li>
</ul>
```

<!--### 激活状态 -->

## 组合使用

### 添加 Badge

`````html
<ul class="am-list am-list-static am-list-border">
  <li>
    <span class="am-badge am-badge-success">YES</span> <span class="am-badge am-badge-danger">NO</span>
    每个人都有一个死角， 自己走不出来，别人也闯不进去。</li>
  <li>
    <span class="am-badge">17</span>
    我把最深沉的秘密放在那里。</li>
  <li><span class="am-badge">5</span>你不懂我，我不怪你。</li>
</ul>
`````
```html
<ul class="am-list am-list-static am-list-border">
  <li>
    <span class="am-badge am-badge-success">YES</span> <span class="am-badge am-badge-danger">NO</span>
    每个人都有一个死角， 自己走不出来，别人也闯不进去。</li>
  </ul>
```

### 添加 ICON

`````html
<ul class="am-list am-list-static am-list-border">
  <li>
    <i class="am-icon-home am-icon-fw"></i>
    每个人都有一个死角， 自己走不出来，别人也闯不进去。</li>
  <li>
    <i class="am-icon-book am-icon-fw"></i>
    我把最深沉的秘密放在那里。</li>
  <li><i class="am-icon-pencil am-icon-fw"></i>你不懂我，我不怪你。</li>
</ul>

<ul class="am-list am-list-border">
  <li><a href="#"><i class="am-icon-home am-icon-fw"></i>
    每个人都有一个死角， 自己走不出来，别人也闯不进去。</a></li>
  <li><a href="#"> <i class="am-icon-book am-icon-fw"></i>
    我把最深沉的秘密放在那里。</a></li>
  <li><a href="#"><i class="am-icon-pencil am-icon-fw"></i>你不懂我，我不怪你。</a></li>
</ul>
`````

```html
<ul class="am-list am-list-static am-list-border">
  <li>
    <i class="am-icon-home am-icon-fw"></i>
    每个人都有一个死角， 自己走不出来，别人也闯不进去。</li>
  <li>
    <i class="am-icon-book am-icon-fw"></i>
    我把最深沉的秘密放在那里。</li>
  <li><i class="am-icon-pencil am-icon-fw"></i>你不懂我，我不怪你。</li>
</ul>

<ul class="am-list am-list-border">
  <li><a href="#"><i class="am-icon-home am-icon-fw"></i>
    每个人都有一个死角， 自己走不出来，别人也闯不进去。</a></li>
  <li><a href="#"> <i class="am-icon-book am-icon-fw"></i>
    我把最深沉的秘密放在那里。</a></li>
  <li><a href="#"><i class="am-icon-pencil am-icon-fw"></i>你不懂我，我不怪你。</a></li>
</ul>
```

### 与 Panel 组合

见 [Panel 组件](/css/panel)。
