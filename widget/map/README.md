# Map 地图
---

**本组件调用百度地图极速版 API，适用于触控设备，使用鼠标无法进行拖放等操作。**

如果通过地址定位不准确，可以选择使用经纬度定位，设置经纬度定位以后，地址定位会被忽略。

__经纬度获取：__打开[百度地图坐标拾取系统](http://api.map.baidu.com/lbsapi/getpoint/index.html)，在地图中找准要标识的位置，点击右上角【复制】按钮即可获取经纬度（逗号前面的为经度，后面的为纬度）。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的配置信息替换成自己的信息。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `map`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面，设置名称、坐标等信息；
- 本组件无需采集数据。

## API

```javascript
{
  // id
  "id": "",

  // 自定义 class
  "className": "",

  // 主题
  "theme": "",

  // 选项
  "options": {
    "name": "", // 坐标名称
    "address": "", // 公司地址，地图定位的地址
    "longitude": "", // 经度
    "latitude": "" // 纬度    
    "zoomControl": Boolean, // 是否开启地图缩放控件
    "scaleControl": Boolean, // 是否开启地图比例尺控件
    "setZoom": Number, // 设置地图缩放级别 3-18 
    "icon": "" // 标注图标
  }
}
```
