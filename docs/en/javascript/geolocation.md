# Geolocation
---

An cencapsulation of HTML5 Geolocation.

## HTML5 API Introduction

A `navigator.geolocation` object has following methods: 

```
// Get current position
.getCurrentPosition(successCallback, [errorCallback,options]);

// Watch the change of position
.watchPosition(successCallback, [errorCallback,options]);

// Clear watch
.clearWatch(watchId);
```

The `successCallback` in first two methods is required.

### Parameters of success callback function

```js
var successCallback = function(position) {

};
```

A position object is returned when position is got successfully.

```js
{
  coords: {
    latitude:,
    longitude:: ,
    altitude: ,
    accuracy: ,
    altitudeAccuracy: ,
    heading: , // The angle between NORTH and the direction of device in clockwise.
    speed: '' 
  }
  timestamp: ''
}
```

### The PositionError object returned when error happens

```js
{
  code: 2
  message: ""
}

// PERMISSION_DENIED = 1;
// POSITION_UNAVAILABLE = 2;
// TIMEOUT = 3;
```

### Options

- `enableHighAccuracy`: Whether enable high accuracy. `Boolean`. Default value is `false`. Response time and traffic will increase if high accuracy is enabled on mobile devices;
- `timeout`: Response timeout(ms). Default value is `0`, which means no timeout.
- `maximumAge`: Cache time(ms). Default value is `0`, which means no cache.


## Geolocation in Amaze UI

Constructor of Geolocation in Amaze UI can be called using `jQuery.AMUI.Geolocation`:

```js
var geo = new jQuery.AMUI.Geolocation();
```

The options above can be passed to constructor.

```js
var geo = new jQuery.AMUI.Geolocation({
    enableHighAccuracy: true,
    timeout: 5000,
    maximumAge: 60000
 });
```

There are three methods in Geolocation instance：

- **`.get(options)`**：
  - Correspond to the `getCurrentPosition()` method in original API;
  - The `options` here are just like the options above. The options set here will replace the options in this instance;
  - Return a [jQuery dealy Promise object](http://api.jquery.com/category/deferred-object/).

  ```js
  var geo = new jQuery.AMUI.Geolocation({
    enableHighAccuracy: true,
    timeout: 5000,
    maximumAge: 60000
  });

  geo.get().then(function(position){
    // Success callback. 'position' is the position object returned by get()
  }, function(err) {
    // Error callback. 'err' is the error message
  });
  ```

- **`.watch(options)`**：
  - Correspond to the `watchPosition()` method in original API;
  - Besides the three `options` above, success callback function must be given using `options.done` (error callback function `options.fail` is optional);
  - Return `watchID`；
  - **Considering the limit of power and data traffic, we suggest not to use this method on mobile devices.**

  ```js
  var geo = new jQuery.AMUI.Geolocation({
    enableHighAccuracy: true,
    timeout: 5000,
    maximumAge: 60000
  });

  geo.watch({
    done: function(position){
      // Success callback. 'position' is the position object returned by get()
    },
    fail: function(err) {
      // Error callback. 'err' is the error message
    }
  });
  ```

- `.clearWatch()`: Clear the `watchID` of this instance.

## Examples

**Please use Safari if you are mac user.** Geolocation may not function in other browser because of the security and privacy problem.

### Get current location and display on Baidu map.

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
    var contentString = 'Your location：\n\t Latitude ' + position.coords.latitude +
      '，\n\t Longitude ' + position.coords.longitude + '，\n\t Accuracy ' +
      position.coords.accuracy;
    var map = new BMap.Map('doc-geo-demo');
    var point = new BMap.Point(position.coords.longitude, position.coords.latitude);
    map.centerAndZoom(point, 15);
    map.addControl(new BMap.MapTypeControl());
    map.enableScrollWheelZoom(true);
    var marker = new BMap.Marker(point);  // Create marker
    map.addOverlay(marker);               // Add marker to the map
    marker.setAnimation(BMAP_ANIMATION_BOUNCE); //Add bouncing animation
    map.panTo(point);

    marker.addEventListener('click', function() {
      alert(contentString); // Show alert when clicking on marker
    });
  }, function(err) {
    $demoArea.html('Error occurs when trying to get geoloacion. Error message:<br>' + err);
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
    var contentString = 'Your location：\n\t Latitude ' + position.coords.latitude +
      '，\n\t Longitude ' + position.coords.longitude + '，\n\t Accuracy ' +
      position.coords.accuracy;
    var map = new BMap.Map('doc-geo-demo');
    var point = new BMap.Point(position.coords.longitude, position.coords.latitude);
    map.centerAndZoom(point, 15);
    map.addControl(new BMap.MapTypeControl());
    map.enableScrollWheelZoom(true);
    var marker = new BMap.Marker(point);  // Create marker
    map.addOverlay(marker);               // Add marker to the map
    marker.setAnimation(BMAP_ANIMATION_BOUNCE); //Add bouncing animation
    map.panTo(point);

    marker.addEventListener('click', function() {
      alert(contentString); // Show alert when clicking on marker
    });
  }, function(err) {
    $demoArea.html('Error occurs when trying to get geoloacion. Error message:<br>' + err);
    console.log(err);
  });
});

  var $watch = $('#doc-geo-watch');
  var $clear = $('#doc-geo-clear');

  $watch.on('click', function() {
    alert('Start watching. Please open the console');
    geolocation.watch({
      done: function(position) {
        // console.log(position.coords);
        // console.log(position.timestamp);
        console.log('watchID: ' + geolocation.watchID);
        console.log('Your location：\n\t Latitude ' + position.coords.latitude +
        '，\n\t Longtitude ' + position.coords.longitude);
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

### Watching Geolocation

`````html
<button type="button" class="am-btn am-btn-primary" id="doc-geo-watch">Start Watching</button>
<button type="button" class="am-btn am-btn-warning" id="doc-geo-clear">Clear Watching</button>
`````
```js
$(function() {
  var geolocation = new $.AMUI.Geolocation();

  var $watch = $('#doc-geo-watch');
  var $clear = $('#doc-geo-clear');

  $watch.on('click', function() {
    alert('Start watching. Please open the console');
    geolocation.watch({
      done: function(position) {
        console.log('watchID: ' + geolocation.watchID);
        console.log('Your location：\n\t Latitude ' + position.coords.latitude +
        '，\n\t Longtitude ' + position.coords.longitude);
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

## Reference:

- [W3C Geolocation API Specification](http://www.w3.org/TR/geolocation-API/)
- [MDN - Geolocation](https://developer.mozilla.org/en-US/docs/Web/API/Geolocation)
