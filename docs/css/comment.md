# Comment 评论组件
------

评论组件由用户头像、评论元信息、评论内容组成（有点似曾相识？没错，很像 Github 的评论列表）。

基本结构如下：

```html
<article class="am-comment"> <!-- 评论容器 -->
  <a href="">
    <img class="am-comment-avatar" alt=""/> <!-- 评论者头像 -->
  </a>

  <div class="am-comment-main"> <!-- 评论内容容器 -->
    <header class="am-comment-hd">
      <!--<h3 class="am-comment-title">评论标题</h3>-->
      <div class="am-comment-meta"> <!-- 评论元数据 -->
        <a href="#link-to-user" class="am-comment-author">..</a> <!-- 评论者 -->
        评论于 <time datetime="">...</time>
      </div>
    </header>

    <div class="am-comment-bd">...</div> <!-- 评论内容 -->
  </div>
</article>
```

其中 `.am-comment-title` 在使用中并不常见。

## 使用演示

### 单条评论

`````html
<article class="am-comment">
  <a href="#link-to-user-home">
    <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
  </a>

  <div class="am-comment-main">
    <header class="am-comment-hd">
      <!--<h3 class="am-comment-title">评论标题</h3>-->
      <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">某人</a>
        评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-12 15:30</time>
      </div>
      <div class="am-comment-actions">
        <a href=""><i class="am-icon-pencil"></i></a>
        <a href=""><i class="am-icon-close"></i></a>
      </div>
    </header>

    <div class="am-comment-bd">
      <p>
        那，那是一封写给南部母亲的信。我茫然站在骑楼下，我又看到永远的樱子走到街心。其实雨下得并不大，却是一生一世中最大的一场雨。而那封信是这样写的，年轻的樱子知不知道呢？
      </p>
      <blockquote>
        妈：我打算在下个月和樱子结婚。
      </blockquote>
    </div>
    <footer class="am-comment-footer">
      <div class="am-comment-actions">
        <a href=""><i class="am-icon-thumbs-up"></i></a>
        <a href=""><i class="am-icon-thumbs-down"></i></a>
        <a href=""><i class="am-icon-reply"></i></a>
      </div>
    </footer>
  </div>
</article>
`````

```html
<article class="am-comment">
  <a href="#link-to-user-home">
    <img src="" alt="" class="am-comment-avatar" width="48" height="48"/>
  </a>

  <div class="am-comment-main">
    <header class="am-comment-hd">
      <!--<h3 class="am-comment-title">评论标题</h3>-->
      <div class="am-comment-meta">
        <a href="#link-to-user" class="am-comment-author">某人</a>
        评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-12 15:30</time>
      </div>
    </header>

    <div class="am-comment-bd">
      ...
    </div>
  </div>
</article>
```

### 评论列表

使用 `.am-comments-list` 包裹多个 `.am-comment` 即成评论列表。

给`<ul>`元素添加`.am-comment-list`类来创建一个评论列表。


`````html
<ul class="am-comments-list am-comments-list-flip">
  <li class="am-comment">
    <a href="#link-to-user-home">
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <!--<h3 class="am-comment-title">这是我悲伤的评论</h3>-->
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">某人</a>
          评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-12 15:30</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p>
          那，那是一封写给南部母亲的信。我茫然站在骑楼下，我又看到永远的樱子走到街心。其实雨下得并不大，却是一生一世中最大的一场雨。而那封信是这样写的，年轻的樱子知不知道呢？
        </p>
        <blockquote>
          妈：我打算在下个月和樱子结婚。
        </blockquote>
      </div>
    </div>
  </li>

  <li class="am-comment">
    <a href="#link-to-user-home">
      <img src="http://www.gravatar.com/avatar/1ecedeede84a44f371b9d8d656bb4265?d=mm&s=96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">路人甲</a>
          评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-13 0:03</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p>
          She's gone 我早知道 <br/>
          她将要 从我的生命中走掉 <br/>
          不再停留多一秒
        </p>
      </div>
    </div>
  </li>

  <li class="am-comment am-comment-primary">
    <a href="#link-to-user-home">
      <img src="http://www.gravatar.com/avatar/1ecedeede84abbf371b9d8d656bb4265?d=mm&s=96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">路人乙</a>
          评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:30</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lin-to-user">@某人</a>
          撸主保重！
        </p>
      </div>
    </div>
  </li>

  <li class="am-comment am-comment-flip am-comment-secondary">
    <a href="#link-to-user-home">
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <!--<h3 class="am-comment-title">这是我悲伤的评论</h3>-->
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">某人</a>
          评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:301</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lurenyi">@路人乙</a> 朕知道了！</p>
      </div>
    </div>
  </li>

  <li class="am-comment am-comment-highlight">
    <a href="#link-to-user-home">
      <img src="http://www.gravatar.com/avatar/1ecedeede84abbf371b9d8d656bb4265?d=mm&s=96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">路人乙</a>
          评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:32</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lin-to-user">@某人</a> 艹民告退！
        </p>
      </div>
    </div>
  </li>

  <li class="am-comment am-comment-flip am-comment-danger">
    <a href="#link-to-user-home">
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">某人</a>
          评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:301</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lurenyi">@路人乙</a> 滚！</p>
      </div>
    </div>
  </li>

  <li class="am-comment am-comment-warning">
    <a href="#link-to-user-home">
      <img src="http://www.gravatar.com/avatar/1ecedeede84abbf371b9d8d656bb4265?d=mm&s=96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">路人乙</a>
          评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:32</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lin-to-user">@某人</a> 你妹！
        </p>
      </div>
    </div>
  </li>

  <li class="am-comment am-comment-flip am-comment-success">
    <a href="#link-to-user-home">
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">某人</a>
          评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:301</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lurenyi">@路人乙</a> 你妹你妹！</p>
      </div>
    </div>
  </li>
</ul>
`````

```html
<ul class="am-comments-list am-comments-list-flip">
  <li class="am-comment">
    ...
  </li>

  <li class="am-comment">
    ...
  </li>

  ...

  <li class="am-comment am-comment-flip"></li>

  <li class="am-comment am-comment-highlight"></li>
</ul>
```

### 评论内容左右对齐

在评论列表 `.am-comments-list` 上增加 `.am-comments-list-flip` class，可以使左右交错的评论列表内容左右对齐（在 `medium-up` 区间有效）。

**谢谢大家的建议**。并非所有使用场景都使用左右交错的列表，所以加了单独的 class，供用户选择。

### 评论状态

在容器上添加评论状态 class（演示见上面列表里的最后几条）。

- `.am-comment-flip` 在右边显示头像
- `.am-comment-primary` 高亮评论（边框为主色）
- `.am-comment-highlight` / `.am-comment-highlight` 高亮评论（边框为次色）
- `.am-comment-success` 高亮评论（边框为绿色）
- `.am-comment-warning` 高亮评论（边框为橙色）
- `.am-comment-danger` 高亮评论（边框为红色）

```html
<article class="am-comment am-comment-flip">
  ...
</article>

<article class="am-comment am-comment-flip">
  ...
</article>

<article class="am-comment am-comment-flip am-comment-highlight">
  ...
</article>
```
