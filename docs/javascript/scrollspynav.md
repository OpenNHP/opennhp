# ScrollSpyNav
---

适用于页面内导航，窗口滚动时，锚点对应的链接添加高亮样式类（`.am-active`）。

## 使用方法

添加 `data-am-scrollspy-nav` 属性，并设置相关属性。

<table class="am-table am-table-bd am-table-striped">
  <thead>
    <tr>
      <th>属性</th>
      <th>描述</th>
    </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>data-am-scrollspy-nav="{cls:'am-active'}"</code></td>
    <td>高亮导航条类名，默认为 <code>am-active</code></td>
  </tr>
  <tr>
    <td><code>data-am-scrollspy-nav="{smooth: false}"</code></td>
    <td>点击锚点时平滑滚动，默认为 <code>true</code></td>
  </tr>
  <!--<tr>
    <td><code>data-am-scrollspy="{animation:'fade', repeat: false}"</code></td>
    <td>是否重复动画，默认为 <code>true</code></td>
  </tr>-->
  </tbody>
</table>



`````html
<style>
  .scrollspy-nav {
    top: 0;
    z-index: 100;
    background: #800080;
    width: 100%;
    padding: 0 10px;
  }

  .scrollspy-nav ul {
    margin: 0;
    padding: 0;
  }

  .scrollspy-nav li {
    display: inline-block;
    list-style: none;
  }

  .scrollspy-nav a {
    color: #ccc;
    padding: 10px 20px;
    display: inline-block;
  }

  .scrollspy-nav a.am-active {
    color: #ffe;
  }

</style>
<nav class="scrollspy-nav" data-am-scrollspy-nav data-am-sticky>
  <ul>
    <li><a href="#yunshipei">什么是云适配</a></li>
    <li><a href="#product">适配产品</a></li>
    <li><a href="#pricing">适配套餐</a></li>
    <li><a href="#cases">适配案例</a></li>
    <li><a href="#about">关于我们</a></li>
  </ul>
</nav>

<div class="am-panel am-panel-default" id="yunshipei">
  <div class="am-panel-hd">什么是云适配？</div>
  <div class="am-panel-bd">
    <p>云适配技术是一项基于云计算、利用html5进行网站跨屏适配的移动化技术解决方案，它为目标网站开发一行JS代码，并嵌入PC网站，这个JS代码通过对PC网站目标网页数据的分析和抓取，在云端完成用户当前设备的网页最佳展现方式的计算，最后在浏览器端实现html结构的重排及CSS的重新渲染，以适应移动端用户的浏览习惯。云适配版手机网站网址不变、内容同步更新、SEO优化提高，入口更多，一站解决适配和营销问题。
    </p>
  </div>
</div>


<div class="am-panel am-panel-default" id="product">
  <div class="am-panel-hd">适配产品</div>
  <div class="am-panel-bd">
    <p>图片压缩</p>
    <p> 云适配自动检测识别终端类型，根据终端屏幕大小，自动将图片缩放至最适合尺寸，大幅提高图片下载速度，提高页面打开速度，并节省用户流量
    </p>
    <p>CDN加速</p>
    <p>CDN布设图片压缩自动适配代码，极速压缩图片流量，使用户体验更加流畅</p>
  </div>
</div>

<div class="am-panel am-panel-default" id="pricing">
  <div class="am-panel-hd">适配套餐</div>
    <table class="am-table am-table-striped">
      <thead>
      <tr class="price-name">
        <th class="highlight">商务版</th>
        <th>企业版</th>
        <th>高端定制</th>
      </tr>
      </thead>
      <tbody>
      <tr>
        <td>销售型网站</td>
        <td>集团企业官网</td>
        <td>政府/高校/资讯门户、电商团购</td>
      </tr>
      <tr class="dark-color">
        <td>不限</td>
        <td>不限</td>
        <td>不限</td>
      </tr>
      <tr>
        <td>适配所有主流智能手机App内适配</td>
        <td>适配所有主流智能手机App内适配</td>
        <td>适配所有主流智能手机App内适配</td>
      </tr>
      <tr class="dark-color">
        <td>设计师定制设计（可修改两次）</td>
        <td>首席设计师定制设计、国内知名用户体验设计师打造</td>
        <td>首席设计师定制设计、国内知名用户体验设计师打造</td>
      </tr>
      <tr>
        <td>一键拨叫、一键分享（邮件、微博、微信等）、二维码推广</td>
        <td>一键拨叫、一键分享（邮件、微博、微信等）、二维码推广</td>
        <td>一键拨叫、一键分享（邮件、微博、微信等）、二维码推广</td>
      </tr>
      <tr class="dark-color">
        <td>移动地图及导航、站内留言、站内搜索</td>
        <td>移动地图及导航、站内留言、站内搜索、用户登录注册</td>
        <td>移动地图及导航、站内留言、站内搜索、用户登录注册</td>
      </tr>
      <tr>
        <td>-</td>
        <td>-</td>
        <td>移动购物车、移动支付、移动论坛、其他定制功能</td>
      </tr>
      </tbody>
    </table>
</div>


<div class="am-panel am-panel-default" id="cases">
  <div class="am-panel-hd">适配案例</div>
  <div class="am-panel-bd">
    <div class="flex-caption">
      <h3 class="title">品牌长青，联想移动访问量半月增长3倍！</h3>
      <div class="text">
        <p>背景：联想集团是一家在信息产业内多元化发展的大型企业集团，2013年，联想电脑销售量升居世界第1</p>
        <p>联想一直以“成就客户—致力于客户的满意与成功”、“创业创新—追求速度和效率，专注于对客户和公司有影响的创新”为使命，全球移动互联网时代已经全面来临，联想也开始走移动化战略，移动官网是移动化战略第一步，“专业的人做专业的事”这是联想选择和云适配合作的原因！</p>
        <p>数据：联想云适配版于2013年8月5日正式上线，在短短半个月时间内，日均IP和及日访问PV均呈3倍增长，老用户访问量增长10倍！</p>
        <p>安全：联想PC官网历史悠久，逻辑结构复杂，对安全性要求极高。云适配插入代码不侵及后台，数据安全稳定，且很短工期内上线官网移动版，PC版一直正常稳定运营。</p>
        <p>品牌：联想官网云适配版设计简洁、时尚，人性化的交互设计让用户尽享流畅、舒适的体验。另外PC版和移动版内容同步更新，内容实现无缝对接，提高了用户体验，使联想的品牌价值在移动端和PC端得到融合，最大程度展现联想作为一线品牌的高尚品质。</p>
      </div>
      <a href="http://blog.yunshipei.com/?p=76" target="_blank" class="am-btn am-btn-success">了解详情</a>
    </div>
  </div>
</div>

<div class="am-panel am-panel-default" id="about">
  <div class="am-panel-hd">关于我们
  </div>
  <div class="am-panel-bd">
      <p>AllMobilize Inc (美通云动科技有限公司)
        由前微软美国总部IE浏览器核心研发团队成员及移动互联网行业专家在美国西雅图创立，旨在解决网页在不同移动设备屏幕上的适配问题。基于国际专利技术并结合最前沿的HTML5技术，云适配解决方案可以帮助企业快速将桌面版网站适配到各种移动设备终端的屏幕上，不仅显著地提高了企业网站的用户体验以及销售转化率，而且大幅度地节省了企业开发和维护移动网站的费用。</p>

      <p>AllMobilize Inc
        获得了微软创投孵化器的支持，其领先科技已得到全球多家企业及机构的认可与信赖，客户包括全球500强企业、美国政府、国内政府机关、国内外上市公司、以及互联网标准化组织W3C。</p>

      <p>我们的使命是通过技术创新使移动上网变得越来越快捷高效。</p>
  </div>
</div>

<div class="am-panel am-panel-default">
  <div class="am-panel-hd">会不会由于我们的网站点击量大，出现云端服务器崩溃？
  </div>
  <div class="am-panel-bd">
    <p>云适配使用的是微软Azure云平台，它最大的优势就是可靠，微软能够确保平台以及服务的可靠性和安全性。</p>
  </div>
</div>
`````
```html
<div class="am-panel am-panel-default" data-am-scrollspy="{animation: 'fade'}">...</div>

<div class="am-panel am-panel-default" data-am-scrollspy="{animation: 'fade', delay: 300}">...</div>
```

## JS 调用

通过 `$().scrollspynav(options)` 设置，参数同上。


`````html
<script>
  seajs.use(['ui.scrollspynav', 'ui.smooth-scroll'], function() {
    $(function() {

    });
  });
</script>
`````
