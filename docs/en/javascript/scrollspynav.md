---
id: scrollspynav
title: 滚动侦测导航
titleEn: ScrollSpyNav
prev: javascript/scrollspy.html
next: javascript/smooth-scroll.html
source: js/ui.scrollspynav.js
doc: docs/javascript/scrollspynav.md
---

# ScrollSpyNav
---

ScrollSpyNav can be used to achieve in-page navigation. When scrolling, the link with corresponding archor will be given a highlight style (`.am-active` by default).

## Examples

`````html
<style>
  .scrollspy-nav {
    top: 0;
    z-index: 100;
    background: #0e90d2;
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
    color: #eee;
    padding: 10px 20px;
    display: inline-block;
  }

  .scrollspy-nav a.am-active {
    color: #fff;
    font-weight: bold;
  }

  .am-panel {
    margin-top: 20px;
  }

</style>
<nav class="scrollspy-nav" data-am-scrollspy-nav="{offsetTop: 45}" data-am-sticky>
  <ul>
    <li><a href="#about">关于棉花糖</a></li>
    <li><a href="#team">成员介绍</a></li>
    <li><a href="#ep">首张 EP</a></li>
    <li><a href="#second">再见王子</a></li>
    <li><a href="#third">第三张</a></li>
  </ul>
</nav>

<div class="am-panel am-panel-default" id="about">
  <div class="am-panel-hd">棉花糖[katncandix2]</div>
  <div class="am-panel-bd">
    <p>棉花糖[katncandix2]，2007年5月30日街头出身，为两人团体，由团长沈圣哲(圣哲)以及主唱庄鹃瑛(小球)所组成。之所以是街头出身，始于一开始棉花糖[katncandix2]为台北市文化局核准街头艺人，从街头开始演出，往他们的音乐梦想勇敢飞行!!!。棉花糖藉由无数次的演出，一起在烈日下、狂风中、还有那没完没了的雨天里。和所有人一创造热血青春。棉花糖用年轻的眼睛看着现实世界，音乐充满温暖、力量和希望，他们用努力作为实现梦想的强心针，用音乐开创梦想的大世纪，这是属于棉花糖的记忆、一段记录 勇敢的故事。清新的城市民谣，软绵绵却有着温暖的力量，在喧嚣吵杂的城市里，还有一个温暖组合，直接将力量打在你的心脏!“棉花糖katncandix2”，街头艺人。在好天气或坏天气里快乐唱歌，在马路边、小公园或是热闹的柏油路面上，实践音乐的梦想.曾经在台湾,大陆进行数千次的街头演唱，感动了无数青年的心.触动了无数人内心的感伤和爱情观2008年1月11日，发行首张创作EP 2375。2009年5月1日发行首张创作专辑「小飞行」。 2010年5月14日发行第二张创作专辑「再见王子」。
    </p>
  </div>
</div>


<div class="am-panel am-panel-default" id="team">
  <div class="am-panel-hd">成员介绍</div>
  <div class="am-panel-bd">
    <p>主唱：小球主要为棉花糖[katncandix2]的歌词创作人.叫小球原因是因为小球以前喜欢穿宽宽的衣裤,风大时跑起来像球.因为小球比较怕冷，容易感冒 所以平时都会加个围巾.</p>
    <p>团长：圣哲主要为棉花糖[katncandix2]的制作人以及曲创作人，平时演出时以木吉他为主要乐器,大家可以叫他老板哦~</p>
  </div>
</div>

<div class="am-panel am-panel-default" id="ep">
  <div class="am-panel-hd">首张 EP</div>
  <div class="am-panel-bd">
    <h3>2375</h3>
    <p>EP曲目 介绍</p>

    <p>INTRO [家]</p>

    <p>2375导入式引言。每个人都有想到达的地方，想完成的梦想， 那是你未来想去的家，只要没有放弃，我相信有一天。</p>

    <p>[2375]</p>

    <p>轻松玩吉他手小毛热血跨刀， 2008年初的第一道梦想光芒， 舒服温暖直入你的心脏，不管经过了多久的时间，终于会到达。</p>

    <p>[X2]</p>

    <p>友情X2，勇气X2。好朋友必备默契之歌，一起看见所有的漂亮。</p>

    <p>[你的力量]</p>

    <p>棉花糖式情歌，最简单的旋律，最深刻的情歌，给失恋的你。</p>

    <p>OUTRO</p>

    <p>[幸福的花]</p>

    <p>幸福的花开了，走过的路就不苦了。幸福的花开了，我们就不哭了。</p>
  </div>
</div>


<div class="am-panel am-panel-default" id="second">
  <div class="am-panel-hd">再见王子</div>
  <div class="am-panel-bd">
        <p>来自街头的声音。飘浮在城市天空的“棉花糖”。 柔软却刚强，义无反顾地行进。 棉花糖[katncandix2] 小球+圣哲 第2张全创作专辑 [再见王子] 再见梦想。再见初恋。再见泪水。再见昨天。 必须向天真的自己说再见，才能勇敢地冒险找到未来。 幻灭之后的蜕变，是重生的那一刻。 “小飞行”好评后 严选《再见王子》《好日子》《怎么说呢?》《回不去的旅人》 全新格局 10篇刻印成长的青春创作 词+曲+创作+制作+演唱 棉花糖 棉花糖的音乐形成自成一格的清透感 。</p>
  </div>
</div>

<div class="am-panel am-panel-default" id="third">
  <div class="am-panel-hd">第三张</div>
  <div class="am-panel-bd">
    <p>
      一切看似美好的过程之中，有许多不同却深刻的故事，跟随着生命和宇宙的运行发生中。悲伤的、快乐的时常已无以名状。棉花糖将过程中产生的讯息，刻划成专辑里的十一首歌曲。棉花糖此时还能笑着说这张专辑：「我们走进了黑暗的入口，开始寻找希望与光明，当真正找到了透进光线的出口并走出时，我们早已伤痕无数、甚至被种植黑暗，热情是唯一能抵抗的道具。」我说，那是一段不被了解的路程。</p>

    <p>每个人都是不被了解的怪人，你/你的与众不同，其实与众相同。也许将自我置放在一个不被了解的过程中，才能真正面对前所未见的情感释放与获得。再确认一次：你/你的与众不同，其实与众相同。</p>
  </div>
</div>

`````
```html
<nav class="scrollspy-nav" data-am-scrollspy-nav="{offsetTop: 45}" data-am-sticky>
  <ul>
    <li><a href="#about">关于棉花糖</a></li>
    <li><a href="#team">成员介绍</a></li>
    <li><a href="#ep">首张 EP</a></li>
    <li><a href="#second">再见王子</a></li>
    <li><a href="#third">第三张</a></li>
  </ul>
</nav>
```

## Usage

### Using Data API

Add `data-am-scrollspy-nav` attribute.


### Using JS

Initialize through `$().scrollspynav(options)`.


### Options

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Option</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>options.className.active</code></td>
    <td>String. The highlight class name. Default class is <code>am-active</code></td>
  </tr>
  <tr>
    <td><code>options.smooth</code></td>
    <td>Boolean. Scroll smoothly when archer is clicked. Default value is <code>true</code></td>
  </tr>
  </tbody>
</table>

When using Data API, the API format is `data-am-scrollspy-nav="{className: {active: 'am-active'}}"`.
