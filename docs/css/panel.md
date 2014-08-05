# Panel
---

面板组件带有轮廓、常用来放置带标题和文字的内容块。


## 基本样式

默认的 `.am-panel` 提供基本的阴影和边距，默认边框添加 `.am-panel-default`，内容放在 `.am-panel-bd` 里面。

`````html
<div class="am-panel am-panel-default">
    <div class="am-panel-bd">这是一个基本的面板组件。</div>
</div>
`````

```html
<div class="am-panel am-panel-default">
    <div class="am-panel-bd">这是一个基本的面板组件。</div>
</div>
```
## 带标题的面板

`.am-panel-hd` 用来放置标题，建议使用 `h1` - `h6` 并添加 `.am-panel-title` class，更加语义化。

`````html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">面板标题</div>
  <div class="am-panel-bd">
    面板内容
  </div>
</div>

<section class="am-panel am-panel-default">
  <header class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </header>
  <div class="am-panel-bd">
    面板内容
  </div>
</section>
`````

```html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">面板标题</div>
  <div class="am-panel-bd">
    面板内容
  </div>
</div>

<section class="am-panel am-panel-default">
  <header class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </header>
  <div class="am-panel-bd">
    面板内容
  </div>
</section>
```

## 面板颜色

添加不同的一下类可以设置不同的颜色。

- `.am-panel-primary`
- `.am-panel-secondary`
- `.am-panel-success`
- `.am-panel-warning`
- `.am-panel-danger`

`````html
<div class="am-panel am-panel-primary">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </div>
  <div class="am-panel-bd">
    面板内容
  </div>
</div>

<div class="am-panel am-panel-secondary">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </div>
  <div class="am-panel-bd">
    面板内容
  </div>
</div>

<div class="am-panel am-panel-success">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </div>
  <div class="am-panel-bd">
    面板内容
  </div>
</div>

<div class="am-panel am-panel-warning">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </div>
  <div class="am-panel-bd">
    面板内容
  </div>
</div>

<div class="am-panel am-panel-danger">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </div>
  <div class="am-panel-bd">
    面板内容
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

## 面板页脚

面板页脚 `.am-panel-footer`，用于放置次要信息。页脚不会继承 `.am-panel-primary` - `.am-panel-danger` 等颜色样式。

`````html
<section class="am-panel am-panel-default">
  <main class="am-panel-bd">
    面板内容
  </main>
  <footer class="am-panel-footer">面板页脚</footer>
</section>
`````

```html
<section class="am-panel am-panel-default">
  <main class="am-panel-bd">
    面板内容
  </main>
  <footer class="am-panel-footer">面板页脚</footer>
</section>
```

## 组合使用

### 面板嵌套表格

将没有边框的表格 (`.am-table`) 直接放在 `.am-panel` 下面（不是放在 `.am-panel-bd` 里面）。

`````html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </div>
  <div class="am-panel-bd">
    这里是面板其他内容。
  </div>
  <table class="am-table">
    <thead>
    <tr>
      <th>名称</th>
      <th>网址</th>
      <th>创建时间</th>
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
  <div class="am-panel-footer">面板页脚</div>
</div>
`````

```html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </div>
  <div class="am-panel-bd">
    <p>这里是面板其他内容。</p>
  </div>
  <table class="am-table">
    ...
  </table>
  <div class="am-panel-footer">...</div>
</div>
```

### 面板嵌套列表

`````html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </div>
  <div class="am-panel-bd">
    这里是面板其他内容。
  </div>

  <ul class="am-list am-list-static">
    <li>每个人都有一个死角， 自己走不出来，别人也闯不进去。</li>
    <li>我把最深沉的秘密放在那里。</li>
    <li>你不懂我，我不怪你。</li>
    <li>每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</li>
  </ul>
  <div class="am-panel-footer">...</div>
</div>
`````
```html
<div class="am-panel am-panel-default">
  <div class="am-panel-hd">
    <h3 class="am-panel-title">面板标题</h3>
  </div>
  <div class="am-panel-bd">
    这里是面板其他内容。
  </div>

  <ul class="am-list am-list-static">
    <li>...</li>
  </ul>
  <div class="am-panel-footer">...</div>
</div>
```

## 面板群组

将多个面板放在 `.am-panel-group` 里面，可结合 JS 制作折叠面板（手风琴面板）。

`````html
<div class="am-panel-group">
  <section class="am-panel am-panel-default">
    <header class="am-panel-hd">面板标题</header>
    <main class="am-panel-bd">
      面板内容
    </main>
  </section>

  <section class="am-panel am-panel-default">
    <header class="am-panel-hd">面板标题</header>
    <div class="am-panel-collapse">
      <main class="am-panel-bd">
        面板内容
      </main>
    </div>
  </section>

  <section class="am-panel am-panel-default">
    <header class="am-panel-hd">面板标题</header>
    <main class="am-panel-bd">
      面板内容
    </main>
  </section>
</div>
`````
