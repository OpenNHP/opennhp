# Panel
---

Usually, panel is used to arrange content with title.

## Default Style

The default `.am-panel` has basic shadow, padding and margin. Add `.am-panel-default` to have default border. The content should be placed in `.am-panel-bd`.

`````html
<div class="am-panel am-panel-default">
    <div class="am-panel-bd">This is a basic panel.</div>
</div>
`````

```html
<div class="am-panel am-panel-default">
    <div class="am-panel-bd">This is a basic panel.</div>
</div>
```
## Panel with Title

Title can be put in `.am-panel-hd`, but we suggest to use `h1` - `h6` and add the `.am-panel-title` class.

`````html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">Title</div>
  <div class="am-panel-bd">
    Content
  </div>
</div>

<section class="am-panel am-panel-default">
  <header class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </header>
  <div class="am-panel-bd">
    Content
  </div>
</section>
`````

```html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">Title</div>
  <div class="am-panel-bd">
    Content
  </div>
</div>

<section class="am-panel am-panel-default">
  <header class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </header>
  <div class="am-panel-bd">
    Content
  </div>
</section>
```

## Color

Use following classes to set color of panel.

- `.am-panel-primary`
- `.am-panel-secondary`
- `.am-panel-success`
- `.am-panel-warning`
- `.am-panel-danger`

`````html
<div class="am-panel am-panel-primary">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </div>
  <div class="am-panel-bd">
    Content
  </div>
</div>

<div class="am-panel am-panel-secondary">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </div>
  <div class="am-panel-bd">
    Content
  </div>
</div>

<div class="am-panel am-panel-success">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </div>
  <div class="am-panel-bd">
    Content
  </div>
</div>

<div class="am-panel am-panel-warning">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </div>
  <div class="am-panel-bd">
    Content
  </div>
</div>

<div class="am-panel am-panel-danger">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </div>
  <div class="am-panel-bd">
    Content
  </div>
</div>
`````

```html
<div class="am-panel am-panel-primary">...</div>
<div class="am-panel am-panel-secondary">...</div>
<div class="am-panel am-panel-success">...</div>
<div class="am-panel am-panel-warning">...</div>
<div class="am-panel am-panel-danger">...</div>
```

## Footer

Use footer `.am-panel-footer` for secondary information. Footer won't inherite styles from `.am-panel-primary` - `.am-panel-danger`.

`````html
<section class="am-panel am-panel-default">
  <main class="am-panel-bd">
    Title
  </main>
  <footer class="am-panel-footer">Footer</footer>
</section>
`````

```html
<section class="am-panel am-panel-default">
  <main class="am-panel-bd">
    Title
  </main>
  <footer class="am-panel-footer">Footer</footer>
</section>
```

## With Other Components

### With Table

Use `.am-table` in `.am-panel`. (Not in `.am-panel-bd`)

`````html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </div>
  <div class="am-panel-bd">
    Other contents.
  </div>
  <table class="am-table">
    <thead>
    <tr>
      <th>Name</th>
      <th>URL</th>
      <th>Created Date</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td>Amaze UI</td>
      <td>amazeui.org</td>
      <td>2014-01-01</td>
    </tr>
    <tr>
      <td>Amaze UI</td>
      <td>amazeui.org</td>
      <td>2014-01-01</td>
    </tr>
    <tr>
      <td>Amaze UI</td>
      <td>amazeui.org</td>
      <td>2014-01-01</td>
    </tr>
    </tbody>
  </table>
  <div class="am-panel-footer">Footer</div>
</div>
`````

```html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </div>
  <div class="am-panel-bd">
    Other contents.
  </div>
  <table class="am-table">
    ...
  </table>
  <div class="am-panel-footer">...</div>
</div>
```

### With List

`````html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </div>
  <div class="am-panel-bd">
    Other contents
  </div>

  <ul class="am-list am-list-static">
    <li>Devouring Time,blunt thou the lion'paws,</li>
    <li>And make the earth devour her own sweet brood;</li>
    <li>Pluck the keen teeth from the fierce tiger's jaws,</li>
    <li>And burn the long--liv'd phoenix in her blood;</li>
  </ul>
  <div class="am-panel-footer">...</div>
</div>
`````
```html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">Title</h3>
  </div>
  <div class="am-panel-bd">
    Other contents
  </div>

  <ul class="am-list am-list-static">
    <li>...</li>
  </ul>
  <div class="am-panel-footer">...</div>
</div>
```

## With Group

Use panel in `.am-panel-group`.

`````html
<div class="am-panel-group">
  <section class="am-panel am-panel-default">
    <header class="am-panel-hd">Title</header>
    <main class="am-panel-bd">
      Content
    </main>
  </section>

  <section class="am-panel am-panel-default">
    <header class="am-panel-hd">Title</header>
    <div class="am-panel-collapse">
      <main class="am-panel-bd">
        Content
      </main>
    </div>
  </section>

  <section class="am-panel am-panel-default">
    <header class="am-panel-hd">Title</header>
    <main class="am-panel-bd">
      Content
    </main>
  </section>
</div>
`````
