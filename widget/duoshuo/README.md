# Duoshuo
---

多说评论系统。

<div class="am-alert am-alert-danger">
  请大家使用的时候把 <code>data-ds-short-name="amazeui"</code> 中的 <code>amazeui</code> 换成自己在<a href="http://official.duoshuo.com/login/" target="_blank">多说注册</a> 的网站 ID，不然你们的评论都跑到 Amaze UI 这边来了。
</div>

<div class="am-alert am-alert-warning">
  本组件由<a href="http://dev.duoshuo.com/" target="_blank">多说</a>提供服务，更多细节参加<a href="http://dev.duoshuo.com/docs" target="_blank">官方文档</a>。
</div>

## 使用方法

设置多说域名即可，其他参数可选。

## API

```javascript
{
  "id": "",
  "className": "",
  "theme": "",
  "options": {
    "shortName": "",
  },
  "content": {
    "threadKey": "",
    "title": "",
    "url": ""
  }
}
```