# Comment
------

Comment consists of users' avatar, meta information and comment content. (Just like comment in github.)

Basic structure: 

```html
<article class="am-comment"> <!-- Container -->
  <a href="">
    <img class="am-comment-avatar" alt=""/> <!-- Avatar -->
  </a>

  <div class="am-comment-main"> <!-- Content Container -->
    <header class="am-comment-hd">
      <!--<h3 class="am-comment-title">Title</h3>-->
      <div class="am-comment-meta"> <!-- Meta -->
        <a href="#link-to-user" class="am-comment-author">..</a> <!-- Author -->
        Comment at <time datetime="">...</time>
      </div>
    </header>

    <div class="am-comment-bd">...</div> <!-- 评论内容 -->
  </div>
</article>
```

Usually we won't use `.am-comment-title`.

## Usage

### Single Comment

`````html
<article class="am-comment">
  <a href="#link-to-user-home">
    <img src="http://amui.qiniudn.com/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
  </a>

  <div class="am-comment-main">
    <header class="am-comment-hd">
      <!--<h3 class="am-comment-title">Title</h3>-->
      <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">Ernest Hemingway</a>
        Comment at <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-12 15:30</time>
      </div>
      <div class="am-comment-actions">
        <a href=""><i class="am-icon-pencil"></i></a>
        <a href=""><i class="am-icon-close"></i></a>
      </div>
    </header>

    <div class="am-comment-bd">
      <p>
        Every day is a new day. It is better to be lucky. But I would rather be exact. Then when luck comes you are ready.
      </p>
      <blockquote>
        Luck is a thing that comes in many forms and who can recognize her?
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
```

```html
<article class="am-comment">
  <a href="#link-to-user-home">
    <img src="" alt="" class="am-comment-avatar" width="48" height="48"/>
  </a>

  <div class="am-comment-main">
    <header class="am-comment-hd">
      <!--<h3 class="am-comment-title">Title</h3>-->
      <div class="am-comment-meta">
        <a href="#link-to-user" class="am-comment-author">Ernest Hemingway</a>
        Comment at <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-12 15:30</time>
      </div>
    </header>

    <div class="am-comment-bd">
      ...
    </div>
  </div>
</article>
```

### Comments list

Use `.am-comments-list` to pack multiple `.am-comment` to form a comments list.

Add the `.am-comment-list` class to `<ul>` element to create a comments list.


`````html
<ul class="am-comments-list am-comments-list-flip">
  <li class="am-comment">
    <a href="#link-to-user-home">
      <img src="http://amui.qiniudn.com/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <!--<h3 class="am-comment-title">Quotes from my novel</h3>-->
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">Ernest Hemingway</a>
          Comment at <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-12 15:30</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p>
          Every day is a new day. It is better to be lucky. But I would rather be exact. Then when luck comes you are ready.
        </p>
        <blockquote>
          Luck is a thing that comes in many forms and who can recognize her?
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
          <a href="#link-to-user" class="am-comment-author">Alen</a>
          评论于 <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-13 0:03</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p>
          She's gone; I've already known <br/>
          She will leave from my life
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
          <a href="#link-to-user" class="am-comment-author">Betty</a>
          Comment at <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:30</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lin-to-user">@Ernest Hemingway</a>
          Good luck!
        </p>
      </div>
    </div>
  </li>

  <li class="am-comment am-comment-flip am-comment-secondary">
    <a href="#link-to-user-home">
      <img src="http://amui.qiniudn.com/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <!--<h3 class="am-comment-title">Quotes from my novel</h3>-->
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">Ernest Hemingway</a>
          Comment at <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:301</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lurenyi">@Betty</a> Got it!</p>
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
          <a href="#link-to-user" class="am-comment-author">Betty</a>
          Comment at <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:32</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lin-to-user">@Ernest Hemingway</a> Bye-Bye
        </p>
      </div>
    </div>
  </li>

  <li class="am-comment am-comment-flip am-comment-danger">
    <a href="#link-to-user-home">
      <img src="http://amui.qiniudn.com/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">Ernest Hemingway</a>
          Comment at <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:301</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lurenyi">@路人乙</a> Bye</p>
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
          <a href="#link-to-user" class="am-comment-author">Betty</a>
          Comment at <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:32</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lin-to-user">@Ernest Hemingway</a> Love you.
        </p>
      </div>
    </div>
  </li>

  <li class="am-comment am-comment-flip am-comment-success">
    <a href="#link-to-user-home">
      <img src="http://amui.qiniudn.com/bw-2014-06-19.jpg?imageView/1/w/96/h/96" alt="" class="am-comment-avatar" width="48" height="48"/>
    </a>

    <div class="am-comment-main">
      <header class="am-comment-hd">
        <div class="am-comment-meta">
          <a href="#link-to-user" class="am-comment-author">Ernest Hemingway</a>
          Comment at <time datetime="2013-07-27T04:54:29-07:00" title="2013年7月27日 下午7:54 格林尼治标准时间+0800">2014-7-14 23:301</time>
        </div>
      </header>

      <div class="am-comment-bd">
        <p><a href="#lurenyi">@路人乙</a> Me too.</p>
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

### Align

Add the `.am-comments-list-flip` class to `.am-comments-list` to align the comments alternately to left and right (works in `medium-up`).

**Suggestion from users**: Not everyone like alternately align their comments, so we make this a seperate class. You may use it as your wish.

### Comment status

Add the status classes to comments (like the last comments in example above).

- `.am-comment-flip` Show avatar on the right
- `.am-comment-primary` Highlight Comment (primary border)
- `.am-comment-highlight` / `.am-comment-highlight` Highlight Comment (secondary border)
- `.am-comment-success` Highlight Comment (green border)
- `.am-comment-warning` Highlight Comment (orange border)
- `.am-comment-danger` Highlight Comment (red border)

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
