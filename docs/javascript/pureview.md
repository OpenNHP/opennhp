# PureView
---

> P.ure...
> ~ some one like u

## 使用演示

### 从链接中获取图片
    
`````html
<ul data-am-widget="gallery" class="am-gallery sm-block-grid-2
  md-block-grid-3 lg-block-grid-4 am-gallery-default" data-am-pureview="{target: 'a'}">
  <li>
    <div class="am-gallery-item">
      <a href="http://cn.bing.com/az/hprichv/LondonTrainStation_GettyRR_139321755_ZH-CN742316019.jpg" title="远方 有一个地方 那里种有我们的梦想">
        <img src="http://cn.bing.com/az/hprichv/LondonTrainStation_GettyRR_139321755_ZH-CN742316019.jpg"
             alt="远方 有一个地方 那里种有我们的梦想" />
        <h3 class="am-gallery-title">远方 有一个地方 那里种有我们的梦想</h3>
        <div class="am-gallery-desc">2375-09-26</div>
      </a>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <a href="http://s.cn.bing.net/az/hprichbg/rb/CardinalsBerries_ZH-CN10679090179_1366x768.jpg">
        <img src="http://s.cn.bing.net/az/hprichbg/rb/CardinalsBerries_ZH-CN10679090179_1366x768.jpg"
             alt="某天 也许会相遇 相遇在这个好地方" />
        <h3 class="am-gallery-title">某天 也许会相遇 相遇在这个好地方</h3>
        <div class="am-gallery-desc">2375-09-26</div>
      </a>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <a href="http://s.cn.bing.net/az/hprichbg/rb/QingdaoJiaozhou_ZH-CN10690497202_1366x768.jpg">
        <img src="http://s.cn.bing.net/az/hprichbg/rb/QingdaoJiaozhou_ZH-CN10690497202_1366x768.jpg"
             alt="不要太担心 只因为我相信" />
        <h3 class="am-gallery-title">不要太担心 只因为我相信</h3>
        <div class="am-gallery-desc">2375-09-26</div>
      </a>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <a href="http://s.cn.bing.net/az/hprichbg/rb/FennecFox_ZH-CN13720911949_1366x768.jpg">
        <img src="http://s.cn.bing.net/az/hprichbg/rb/FennecFox_ZH-CN13720911949_1366x768.jpg"
             alt="终会走过这条遥远的道路" />
        <h3 class="am-gallery-title">终会走过这条遥远的道路</h3>
        <div class="am-gallery-desc">2375-09-26</div>
      </a>
    </div>
  </li>
</ul>
`````

### 从 `data-rel` 中获取图片

`````html
<div data-am-pureview>
  <img src="http://amui.qiniudn.com/bw-2014-06-19.jpg?imageView2/0/w/120" data-rel="http://www.yi1000.com/uploadfile/image/20140519/20140519180561186118.jpg" alt="哇哇"/>
</div>
