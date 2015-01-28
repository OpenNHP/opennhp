# Geolocation
---

HTML5 Geolocation 封装。


## HTML5 API 简介

`navigator.geolocation` 对象有以下方法：

```
// 获取当前位置
.getCurrentPosition(successCallback, [[errorCallback,]options]);

// 监视位置变化
.watchPosition(successCallback, [[errorCallback,]options]);

// 清除监视
.clearWatch(watchId);
```

前两个方法的成功回调函数 `successCallback` 必需。

### 获取位置成功时处理函数的参数

```js
var successCallback = function(position) {

};
```

获取位置成功时返回一个位置对象，可以在成功回调函数中调用：

```js
{
  coords: {
    latitude:, // 维度
    longitude:: , // 经度
    altitude: , // 海拔
    accuracy: , // 精确度
    altitudeAccuracy: , // 海拔精确度
    heading: , // 设备正北顺时针前进的方位
    speed: '' // 设备外部环境的移动速度（m/s）
  }
  timestamp: '' // 获得位置的时间戳
}
```

### 发生错误时返回 PositionError 对象

```js
{
  code: 2
  message: ""
}

// PERMISSION_DENIED = 1;
// POSITION_UNAVAILABLE = 2;
// TIMEOUT = 3;
```

### 选项

- `enableHighAccuracy`: 是否启用高精度，`Boolean`，默认为 `false`，开启以后响应时间会变长，在移动设备上会消耗更多流量；
- `timeout`: 响应超时时间（毫秒），默认为 `0`，即不限制时长；
- `maximumAge`: 缓存时间（毫秒），默认为 `0`，即不混存，每次都重新获取。


## Amaze UI 封装的 Geolocation

通过 `jQuery.AMUI.Geolocation` 可以访问 Amaze UI 封装的 Geolocation 构造函数：

```js
var geo = new jQuery.AMUI.Geolocation();
```

可以将上述 `enableHighAccuracy`、`timeout`、`maximumAge` 三个选项传递给构造函数。

```js
var geo = new jQuery.AMUI.Geolocation({
    enableHighAccuracy: true,
    timeout: 5000,
    maximumAge: 60000
 });
```

Geolocation 实例有三个方法：

- **`.get(options)`**：
  - 对应原生 API 的 `getCurrentPosition()` 方法；
  - `options` 选项同上，这里设置的选项会覆盖该实例的选项；
  - 返回一个[jQuery 延迟 Promise 对象](http://api.jquery.com/category/deferred-object/)。

  ```js
  var geo = new jQuery.AMUI.Geolocation({
    enableHighAccuracy: true,
    timeout: 5000,
    maximumAge: 60000
  });

  geo.get().then(function(position){
    // 成功回调，position 为返回的位置对象
  }, function(err) {
    // 不支持或者发生错误时回调，err 为错误提示信息
  });
  ```

- **`.watch(options)`**：
  - 对应原生 API 的 `getCurrentPosition()` 方法；
  - `options` 除三个选项以外，还必须通过 `options.done` 设置成功时的回调函数（失败回调 `options.fail` 可选）；
  - 返回 `watchID`；
  - **出于电量、流量消耗等考虑，不建议在移动设备上使用此方法**。

  ```js
  var geo = new jQuery.AMUI.Geolocation({
    enableHighAccuracy: true,
    timeout: 5000,
    maximumAge: 60000
  });

  geo.watch({
    done: function(position){
      // 成功回调，position 为返回的位置对象
    },
    fail: function(err) {
      // 不支持或者发生错误时回调，err 为错误提示信息
    }
  });
  ```

- `.clearWatch()`：清除当前实例的 `watchID`。

## 使用示例

**使用 Mac 的用户请使用 Safari 浏览器查看**，其他浏览器可能由于安全性和隐私设置无法使用定位服务。

### 获取当前位置并在百度地图上显示

`````html
<div id="doc-geo-demo" style="width: 100%; height: 400px;">
</div>
`````

```html
<div id="doc-geo-demo" style="width: 100%; height: 400px;"></div>

<script src="http://api.map.baidu.com/api?v=2.0&ak=WVAXZ05oyNRXS5egLImmentg"></script>
```

```js
$(function() {
  var geolocation = new $.AMUI.Geolocation();
  var $demoArea = $('#doc-geo-demo');

  geolocation.get({timeout: 7000}).then(function(position){
    // console.log(position.coords);
    var contentString = '你的位置：\n\t纬度 ' + position.coords.latitude +
      '，\n\t经度 ' + position.coords.longitude + '，\n\t精确度 ' +
      position.coords.accuracy;
    var map = new BMap.Map('doc-geo-demo');
    var point = new BMap.Point(position.coords.longitude, position.coords.latitude);
    map.centerAndZoom(point, 15);
    map.addControl(new BMap.MapTypeControl());
    map.enableScrollWheelZoom(true);
    var marker = new BMap.Marker(point);  // 创建标注
    map.addOverlay(marker);               // 将标注添加到地图中
    marker.setAnimation(BMAP_ANIMATION_BOUNCE); //跳动的动画
    map.panTo(point);

    marker.addEventListener('click', function() {
      alert(contentString); // 点击点弹出信息
    });
  }, function(err) {
    $demoArea.html('获取地理位置时发生错误，错误信息：<br>' + err);
    console.log(err);
  });
});
```

<script src="http://api.map.baidu.com/api?v=2.0&ak=WVAXZ05oyNRXS5egLImmentg"></script>

<script>
$(function() {
  var geolocation = new $.AMUI.Geolocation();
  var $demoArea = $('#doc-geo-demo');
  geolocation.get({timeout: 7000}).then(function(position){
    // console.log(position.coords);
    var contentString = '你的位置：\n\t纬度 ' + position.coords.latitude +
      '，\n\t经度 ' + position.coords.longitude + '，\n\t精确度 ' +
      position.coords.accuracy;
    var map = new BMap.Map('doc-geo-demo');
    var point = new BMap.Point(position.coords.longitude, position.coords.latitude);
    map.centerAndZoom(point, 15);
    map.addControl(new BMap.MapTypeControl());
    map.enableScrollWheelZoom(true);
    var marker = new BMap.Marker(point);  // 创建标注
    map.addOverlay(marker);               // 将标注添加到地图中
    marker.setAnimation(BMAP_ANIMATION_BOUNCE); //跳动的动画
    map.panTo(point);

    marker.addEventListener('click', function() {
      alert(contentString); // 点击点弹出信息
    });
  }, function(err) {
    $demoArea.html('获取地理位置时发生错误，错误信息：<br>' + err);
    console.log(err);
  });

  var $watch = $('#doc-geo-watch');
  var $clear = $('#doc-geo-clear');

  $watch.on('click', function() {
    alert('开始监控，请打开控制台查看。');
    geolocation.watch({
      done: function(position) {
        // console.log(position.coords);
        // console.log(position.timestamp);
        console.log('watchID: ' + geolocation.watchID);
        console.log('你的位置：\n\t纬度 ' + position.coords.latitude +
        '，\n\t经度 ' + position.coords.longitude);
      },
      fail: function(error) {
        console.log(error);
      }
    });
  });

  $clear.on('click', function() {
    geolocation.clearWatch();
    console.log('watchID: ' + geolocation.watchID);
  });
});
</script>

### 监视位置变化

`````html
<button type="button" class="am-btn am-btn-primary" id="doc-geo-watch">开始监视位置</button>
<button type="button" class="am-btn am-btn-warning" id="doc-geo-clear">清除监视</button>
`````
```js
$(function() {
  var geolocation = new $.AMUI.Geolocation();

  var $watch = $('#doc-geo-watch');
  var $clear = $('#doc-geo-clear');

  $watch.on('click', function() {
    alert('开始监控，请打开控制台查看。');
    geolocation.watch({
      done: function(position) {
        console.log('watchID: ' + geolocation.watchID);
        console.log('你的位置：\n\t纬度 ' + position.coords.latitude +
        '，\n\t经度 ' + position.coords.longitude);
      },
      fail: function(error) {
        console.log(error);
      }
    });
  });

  $clear.on('click', function() {
    geolocation.clearWatch();
    console.log('watchID: ' + geolocation.watchID);
  });
});
```

## 参考链接

- [W3C Geolocation API Specification](http://www.w3.org/TR/geolocation-API/)
- [MDN - Geolocation](https://developer.mozilla.org/en-US/docs/Web/API/Geolocation)
