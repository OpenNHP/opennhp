# PureView
---

> P.ure...
> ~ some one like u

## 使用演示

### 从链接中获取图片

`````html
<ul data-am-widget="gallery" class="am-gallery am-avg-sm-2
  am-avg-md-3 am-avg-lg-4 am-gallery-default"
    data-am-pureview="{target: 'a'}" id="doc-pv-gallery">
  <li>
    <div class="am-gallery-item">
      <a href="http://7jpqbr.com1.z0.glb.clouddn.com/pure-1.jpg" title="远方 有一个地方 那里种有我们的梦想">
        <img src="http://7jpqbr.com1.z0.glb.clouddn.com/pure-1.jpg?imageView2/0/w/640"
             alt="远方 有一个地方 那里种有我们的梦想" />
        <h3 class="am-gallery-title">远方 有一个地方 那里种有我们的梦想</h3>
        <div class="am-gallery-desc">2375-09-26</div>
      </a>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <a href="http://7jpqbr.com1.z0.glb.clouddn.com/pure-2.jpg">
        <img src="http://7jpqbr.com1.z0.glb.clouddn.com/pure-2.jpg?imageView2/0/w/640"
             alt="某天 也许会相遇 相遇在这个好地方" />
        <h3 class="am-gallery-title">某天 也许会相遇 相遇在这个好地方</h3>
        <div class="am-gallery-desc">2375-09-26</div>
      </a>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <a href="http://7jpqbr.com1.z0.glb.clouddn.com/pure-3.jpg">
        <img src="http://7jpqbr.com1.z0.glb.clouddn.com/pure-3.jpg?imageView2/0/w/640"
             alt="不要太担心 只因为我相信" />
        <h3 class="am-gallery-title">不要太担心 只因为我相信</h3>
        <div class="am-gallery-desc">2375-09-26</div>
      </a>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <a href="http://7jpqbr.com1.z0.glb.clouddn.com/pure-4.jpg">
        <img src="http://7jpqbr.com1.z0.glb.clouddn.com/pure-4.jpg?imageView2/0/w/640"
             alt="终会走过这条遥远的道路" />
        <h3 class="am-gallery-title">终会走过这条遥远的道路</h3>
        <div class="am-gallery-desc">2375-09-26</div>
      </a>
    </div>
  </li>
</ul>
`````
`````html
<button class="am-btn am-btn-primary" id="doc-pv-append">随机插入一个图片</button>
`````

<script>
  $(function() {
    var $gallery = $('#doc-pv-gallery');
    var $items = $gallery.find('li');

    $('#doc-pv-append').on('click', function() {
      var random = Math.round(Math.random() * 3);
      $items.eq(random).clone(false).find('a').
        removeAttr('data-am-pureviewed').end().appendTo($gallery);
    });
  });
</script>

### 从 `data-rel` 中获取图片

#### BT 的图片

<div data-am-pureview>
  <img src="http://7jpqbr.com1.z0.glb.clouddn.com/bw-2014-06-19.jpg?imageView2/0/w/120" data-rel="http://www.yi1000.com/uploadfile/image/20140519/20140519180561186118.jpg" alt="哇哇"/>
</div>

#### 图片速度比较慢的

<div data-am-pureview>
  <img src="https://farm3.staticflickr.com/2948/15348772291_f0016e18ef_z.jpg" data-rel="https://farm3.staticflickr.com/2948/15348772291_bb0f3af931_k.jpg" alt="哇哇，看看加载动画"/>
</div>
