---
id: store
title: 本地存储
titleEn: Store
prev: javascript/iscroll-lite.html
next: javascript/geolocation.html
source: js/util.store.js
doc: docs/javascript/store.md
---

# Store
---

跨浏览器本地存储接口封装，源自 [store.js](https://github.com/marcuswestin/store.js)，精简了针对 IE 6/7 的代码。

LocalStorage 只能存储字符串，store.js 在存取的时候会自动 `stringify`、`parse`。

## 存储接口


通过 `$.AMUI.store` 调用。

### 检测是否支持（开启） LocalStorage

爱上一匹野马之前，先想想你家有没有草原；使用之前，当然要先检测一下。

```javascript
var store = $.AMUI.store;

if (!store.enabled) {
  alert('Local storage is not supported by your browser. Please disable "Private Mode", or upgrade to a modern browser.');
  return;
}

var user = store.get('user');
// ... and so on ...
```

Safari 的`无痕浏览`模式或者用户禁用了本地存储时，`store.enabled` 将返回 `false`。

**浏览器如何禁用 LocalStorage**：

- Firefox： 地址栏输入 `about:config`, 将 `dom.storage.enabled` 的值设置为 `false`；
- Chrome: `设置` → `隐私设置` → `内容设置` → `阻止网站设置任何数据`。

### 接口列表

[LocalStorage](https://developer.mozilla.org/en-US/docs/Web/Guide/API/DOM/Storage) 受[同源策略](https://developer.mozilla.org/en-US/docs/Web/Security/Same-origin_policy)限制，以下操作仅限于同源下的数据。

- `store.set(key, val)`: 存储 `key` 的值为 `val`；
- `store.get(key)`: 获取 `key` 的值；
- `store.remove(key)`: 移除 `key` 的记录；
- `store.clear()`: 清空存储；
- `store.getAll()`: 返回所有存储；
- `store.forEach()`: 遍历所有存储。

```javascript
var store = $.AMUI.store;

// 存储 'username' 的值为 'marcus'
store.set('username', 'marcus')

// 获取 'username'
store.get('username')

// 移除 'username' 字段
store.remove('username')

// 清除所有本地存储
store.clear()

// 存储对象 - 自动调用 JSON.stringify
store.set('user', { name: 'marcus', likes: 'javascript' })

// 获取存储的对象 - 自动执行 JSON.parse
var user = store.get('user')
alert(user.name + ' likes ' + user.likes)

// 从所有存储中获取值
store.getAll().user.name == 'marcus'

// 遍历所有存储
store.forEach(function(key, val) {
  console.log(key, '==', val)
})
```

## 浏览器支持

绝大多数浏览器（包括 IE 8）都[原生支持 LocalStorage](http://caniuse.com/#search=localStorage)。

你的浏览器测试结果为： <strong id="errorOutput" class="am-text-danger"></strong>
<strong id="store-test-success" class="am-text-success"></strong>

不同浏览器对本地存储单条记录的长度限定不同，具体可以通过以下链接测试：

- [Web Storage Support Test](http://dev-test.nemikor.com/web-storage/support-test/)
- [Test of localStorage limits/quota](https://arty.name/localstorage.html)

## 注意事项

### 原生方法与 store.js 的差异

使用原生方法操作：

```javascript
localStorage.myage = 24
localStorage.myage !== 24 // true
localStorage.myage === '24' // true

localStorage.user = { name: 'marcus', likes: 'javascript' }
localStorage.user === "[object Object]" // true

localStorage.tags = ['javascript', 'localStorage', 'store.js']
localStorage.tags.length === 32 // true
localStorage.tags === "javascript,localStorage,store.js" // true
```

使用 store.js 序列化后的结果:

```javascript
store.set('myage', 24)
store.get('myage') === 24 // true

store.set('user', { name: 'marcus', likes: 'javascript' })
alert("Hi my name is " + store.get('user').name + "!") // 仍然返回对象

store.set('tags', ['javascript', 'localStorage', 'store.js'])
alert("We've got " + store.get('tags').length + " tags here") // 仍然返回数组
```

### 自动过期实现

LocalStorage 并没有提供过期时间接口，只能通过存储时间做比对实现。

```javascript
var store = $.AMUI.store;

var storeWithExpiration = {
  set: function(key, val, exp) {
    store.set(key, {val:val, exp:exp, time:new Date().getTime()});
  },

  get: function(key) {
    var info = store.get(key)
    if (!info) {
      return null;
    }

    if (new Date().getTime() - info.time > info.exp) {
      return null;
    }

    return info.val
  }
}；

storeWithExpiration.set('foo', 'bar', 1000);

setTimeout(function() {
  console.log(storeWithExpiration.get('foo'));
}, 500) // -> "bar"

setTimeout(function() {
  console.log(storeWithExpiration.get('foo'));
}, 1500) // -> null
```

## 参考资源

- [Cross domain local storage](https://github.com/zendesk/cross-storage)
- [使用 cookie 实现 LocalStorage](https://developer.mozilla.org/en-US/docs/Web/Guide/API/DOM/Storage#Compatibility)
- [localForage](https://github.com/mozilla/localForage)
- [PouchDB](https://github.com/pouchdb/pouchdb)
- [Basil.js](https://github.com/Wisembly/basil.js)

<script>
  $(function() {
    var store = $.AMUI.store;

    var tests = {
      outputError: null,
      assert: assert,
      runFirstPass: runFirstPass,
      runSecondPass: runSecondPass,
      failed: false
    };

    function assert(truthy, msg) {
      if (!truthy) {
        tests.outputError('bad assert: ' + msg);
        if (store.disabled) {
          tests.outputError('<br>Note that store.disabled == true')
        }
        tests.failed = true
      }
    }

    function runFirstPass() {
      store.clear();

      store.get('unsetValue') // see https://github.com/marcuswestin/store.js/issues/63

      store.set('foo', 'bar')
      assert(store.get('foo') == 'bar', "stored key 'foo' not equal to stored value 'bar'")

      store.remove('foo')
      assert(store.get('foo') == null, "removed key 'foo' not null")

      assert(store.set('foo', 'value') == 'value', "store#set returns the stored value")

      store.set('foo', 'bar1')
      store.set('foo', 'bar2')
      assert(store.get('foo') == 'bar2', "key 'foo' is not equal to second value set 'bar2'")

      store.set('foo', 'bar')
      store.set('bar', 'foo')
      store.remove('foo')
      assert(store.get('bar') == 'foo', "removing key 'foo' also removed key 'bar'")

      store.set('foo', 'bar');
      store.set('bar', 'foo');
      store.clear();
      assert(store.get('foo') == null && store.get('bar') == null, "keys foo and bar not cleared after store cleared")

      store.transact('foosact', function(val) {
        assert(typeof val == 'object', "new key is not an object at beginning of transaction");
        val.foo = 'foo';
      });
      store.transact('foosact', function(val) {
        assert(val.foo == 'foo', "first transaction did not register");
        val.bar = 'bar'
      });
      assert(store.get('foosact').bar == 'bar', "second transaction did not register")

      store.set('foo', {name: 'marcus', arr: [1, 2, 3]})
      assert(typeof store.get('foo') == 'object', "type of stored object 'foo' is not 'object'")
      assert(store.get('foo') instanceof Object, "stored object 'foo' is not an instance of Object")
      assert(store.get('foo').name == 'marcus', "property 'name' of stored object 'foo' is not 'marcus'")
      assert(store.get('foo').arr instanceof Array, "Array property 'arr' of stored object 'foo' is not an instance of Array")
      assert(store.get('foo').arr.length == 3, "The length of Array property 'arr' stored on object 'foo' is not 3")

      assert(store.enabled = !store.disabled, "Store.enabled is not the reverse of .disabled");

      store.remove('circularReference')
      var circularOne = {}
      var circularTwo = {one: circularOne}
      circularOne.two = circularTwo
      var threw = false
      try {
        store.set('circularReference', circularOne)
      }
      catch (e) {
        threw = true
      }
      assert(threw, "storing object with circular reference did not throw")
      assert(!store.get('circularReference'), "attempting to store object with circular reference which should have faile affected store state")

      // If plain local storage was used before store.js, we should attempt to JSON.parse them into javascript values.
      // Store values using vanilla localStorage, then read them out using store.js
      if (typeof localStorage != 'undefined') {
        var promoteValues = {
          'int': 42,
          'bool': true,
          'float': 3.141592653,
          'string': "Don't Panic",
          'odd_string': "{ZYX'} abc:;::)))"
        }
        for (key in promoteValues) {
          localStorage[key] = promoteValues[key]
        }
        for (key in promoteValues) {
          assert(store.get(key) == promoteValues[key], key + " was not correctly promoted to valid JSON")
          store.remove(key)
        }
      }

      // The following stored values get tested in doSecondPass after a page reload
      store.set('firstPassFoo', 'bar')
      store.set('firstPassObj', {woot: true})

      var all = store.getAll()
      assert(all.firstPassFoo == 'bar', 'getAll gets firstPassFoo')
      assert(countProperties(all) == 4, 'getAll gets all 4 values')
    }

    function runSecondPass() {
      assert(store.get('firstPassFoo') == 'bar', "first pass key 'firstPassFoo' not equal to stored value 'bar'")

      var all = store.getAll()
      assert(all.firstPassFoo == 'bar', "getAll still gets firstPassFoo on second pass")
      assert(countProperties(all) == 4, "getAll gets all 4 values")

      store.clear();
      assert(store.get('firstPassFoo') == null, "first pass key 'firstPassFoo' not null after store cleared")

      var all = store.getAll()
      assert(countProperties(all) == 0, "getAll returns 0 properties after store.clear() has been called")
    }

    function countProperties(obj) {
      var count = 0
      for (var key in obj) {
        if (obj.hasOwnProperty(key)) {
          count++
        }
      }
      return count
    }

    var doc = document,
      errorOutput = doc.getElementById('errorOutput'),
      isSecondPass = (doc.location.hash == '#secondPass');

    tests.outputError = function outputError(msg) {
      var prefix = (isSecondPass ? 'second' : 'first') + ' pass '
      errorOutput.appendChild(doc.createElement('div')).innerHTML = prefix + msg
    };

    try {
      if (isSecondPass) {
        tests.runSecondPass()
      } else {
        tests.runFirstPass()
      }
    } catch (e) {
      console.log(e);
      tests.assert(false, 'Tests should not throw: "' + JSON.stringify(e) + '"')
    }

    if (!tests.failed) {
      if (!isSecondPass) {
        doc.location.hash = '#secondPass';
        doc.location.reload()
      } else {
        doc.location.hash = '#';
        $('#store-test-success').html('测试通过！');
      }
    }
  });
</script>
