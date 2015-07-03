# List
---

## Default Style

### Link

Add the `.am-list` class to `<ul>`.

`````html
<ul class="am-list">
  <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
  <li><a href="#">And make the earth devour her own sweet brood;</a></li>
  <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
  <li><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
</ul>
`````

```html
<ul class="am-list">
  <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
  <li><a href="#">And make the earth devour her own sweet brood;</a></li>
  <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
  <li><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
</ul>
```

### Overflow

Add the `.am-text-truncate` class to `<a>` can cut the overflow text and add `...` to the end.

`````html
<ul class="am-list">
  <li><a href="#" class="am-text-truncate">Devouring Time,blunt thou the lion'paws, And make the earth devour her own sweet brood; Pluck the keen teeth from the fierce tiger's jaws, And burn the long--liv'd phoenix in her blood;</a></li>
</ul>
`````
```html
<ul class="am-list">
  <li><a href="#" class="am-text-truncate">Devouring Time,blunt thou the lion'paws, And make the earth devour her own sweet brood; Pluck the keen teeth from the fierce tiger's jaws, And burn the long--liv'd phoenix in her blood;</a></li>
</ul>
```



### Pure Text

Add the `.am-list-static` class to `.am-list`.

`````html
<ul class="am-list am-list-static">
  <li>Devouring Time,blunt thou the lion'paws,</li>
  <li>And make the earth devour her own sweet brood;</li>
  <li>Pluck the keen teeth from the fierce tiger's jaws,</li>
  <li>And burn the long--liv'd phoenix in her blood;</li>
</ul>
`````
```html
<ul class="am-list am-list-static">
  <li>Devouring Time,blunt thou the lion'paws,</li>
  <li>And make the earth devour her own sweet brood;</li>
  <li>Pluck the keen teeth from the fierce tiger's jaws,</li>
  <li>And burn the long--liv'd phoenix in her blood;</li>
</ul>
```

## Style Modifier

### Border

Add the `.am-list-border` class to container.

`````html
<ul class="am-list am-list-static am-list-border">
  <li>Devouring Time,blunt thou the lion'paws,</li>
  <li>And make the earth devour her own sweet brood;</li>
  <li>Pluck the keen teeth from the fierce tiger's jaws,</li>
  <li>And burn the long--liv'd phoenix in her blood;</li>
</ul>

<ul class="am-list am-list-border">
  <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
  <li><a href="#">And make the earth devour her own sweet brood;</a></li>
  <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
  <li><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
</ul>
`````

```html
<ul class="am-list am-list-static am-list-border">
  <li>Devouring Time,blunt thou the lion'paws,</li>
  <li>And make the earth devour her own sweet brood;</li>
  <li>Pluck the keen teeth from the fierce tiger's jaws,</li>
  <li>And burn the long--liv'd phoenix in her blood;</li>
</ul>

<ul class="am-list am-list-border">
  <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
  <li><a href="#">And make the earth devour her own sweet brood;</a></li>
  <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
  <li><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
</ul>
```

### Striped

Add the `.am-list-striped` class.

`````html
<ul class="am-list am-list-static am-list-border am-list-striped">
  <li>Devouring Time,blunt thou the lion'paws,</li>
  <li>And make the earth devour her own sweet brood;</li>
  <li>Pluck the keen teeth from the fierce tiger's jaws,</li>
  <li>And burn the long--liv'd phoenix in her blood;</li>
</ul>
`````
```html
<ul class="am-list am-list-static am-list-border am-list-striped">
  <li>Devouring Time,blunt thou the lion'paws,</li>
  <li>And make the earth devour her own sweet brood;</li>
  <li>Pluck the keen teeth from the fierce tiger's jaws,</li>
  <li>And burn the long--liv'd phoenix in her blood;</li>
</ul>
```

<!--### Activated -->

## Using with Other Components

### With Badge

`````html
<ul class="am-list am-list-static am-list-border">
  <li>
    <span class="am-badge am-badge-success">YES</span> <span class="am-badge am-badge-danger">NO</span>
    Devouring Time,blunt thou the lion'paws,</li>
  <li>
    <span class="am-badge">17</span>
    And make the earth devour her own sweet brood;</li>
  <li><span class="am-badge">5</span>Pluck the keen teeth from the fierce tiger's jaws,</li>
</ul>
`````
```html
<ul class="am-list am-list-static am-list-border">
  <li>
    <span class="am-badge am-badge-success">YES</span> <span class="am-badge am-badge-danger">NO</span>
    Devouring Time,blunt thou the lion'paws,</li>
  </ul>
```

### 添加 ICON

`````html
<ul class="am-list am-list-static am-list-border">
  <li>
    <i class="am-icon-home am-icon-fw"></i>
    Devouring Time,blunt thou the lion'paws,</li>
  <li>
    <i class="am-icon-book am-icon-fw"></i>
    And make the earth devour her own sweet brood;</li>
  <li><i class="am-icon-pencil am-icon-fw"></i>Pluck the keen teeth from the fierce tiger's jaws,</li>
</ul>

<ul class="am-list am-list-border">
  <li><a href="#"><i class="am-icon-home am-icon-fw"></i>
    Devouring Time,blunt thou the lion'paws,</a></li>
  <li><a href="#"> <i class="am-icon-book am-icon-fw"></i>
    And make the earth devour her own sweet brood;</a></li>
  <li><a href="#"><i class="am-icon-pencil am-icon-fw"></i>Pluck the keen teeth from the fierce tiger's jaws,</a></li>
</ul>
`````

```html
<ul class="am-list am-list-static am-list-border">
  <li>
    <i class="am-icon-home am-icon-fw"></i>
    Devouring Time,blunt thou the lion'paws,</li>
  <li>
    <i class="am-icon-book am-icon-fw"></i>
    And make the earth devour her own sweet brood;</li>
  <li><i class="am-icon-pencil am-icon-fw"></i>Pluck the keen teeth from the fierce tiger's jaws,</li>
</ul>

<ul class="am-list am-list-border">
  <li><a href="#"><i class="am-icon-home am-icon-fw"></i>
    Devouring Time,blunt thou the lion'paws,</a></li>
  <li><a href="#"> <i class="am-icon-book am-icon-fw"></i>
    And make the earth devour her own sweet brood;</a></li>
  <li><a href="#"><i class="am-icon-pencil am-icon-fw"></i>Pluck the keen teeth from the fierce tiger's jaws,</a></li>
</ul>
```

### With Panel

See [Panel](/css/panel) for detail.
