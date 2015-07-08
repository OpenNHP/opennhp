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

Store exposes a simple API for cross browser local storage. Via [store.js](https://github.com/marcuswestin/store.js), and codes for IE 6/7 are removed.

LocalStorage can only store strings. store.js will automatically `stringify` and `parse` when store and load the data.

## Store Interface


Call store interface through `$.AMUI.store`.

### Check Support

Check whether browser support LocalStorage(or whether it is enabled).

```javascript
var store = $.AMUI.store;

if (!store.enabled) {
  alert('Local storage is not supported by your browser. Please disable "Private Mode", or upgrade to a modern browser.');
  return;
}

var user = store.get('user');
// ... and so on ...
```

`store.enabled` will return `false` when using `private` mode in Safari or localstore is disabled.

**How to disable LocalStorage in browser:**

- Firefox： Enter `about:config` as url. Then set `dom.storage.enabled` to `false`;
- Chrome: `Setting` → `Privacy` → `Content Setting` → ` Block sites from setting any data.`.s

### Interfaces

[LocalStorage](https://developer.mozilla.org/en-US/docs/Web/Guide/API/DOM/Storage) is limited by [Same Origin Policy](https://developer.mozilla.org/en-US/docs/Web/Security/Same-origin_policy), so the following operation is only avaliable for data from same origin.

- `store.set(key, val)`: Store `key`, `val` pair;
- `store.get(key)`: Get the `value ` of `key`;
- `store.remove(key)`: Delete record with `key`;
- `store.clear()`: Clear storage;
- `store.getAll()`: Return all records;
- `store.forEach()`: Apply to all records;

```javascript
var store = $.AMUI.store;

// Store 'marcus' as the value of 'username'.
store.set('username', 'marcus')

// Get 'username'.
store.get('username')

// Remove 'username'.
store.remove('username')

// Clear the storage.
store.clear()

// Store an object - Use JSON.stringify.
store.set('user', { name: 'marcus', likes: 'javascript' })

// Get an object - Use JSON.parse.
var user = store.get('user')
alert(user.name + ' likes ' + user.likes)

// Get username from all records.
store.getAll().user.name == 'marcus'

// Apply function to all records.
store.forEach(function(key, val) {
  console.log(key, '==', val)
})
```

## Support

Most of the browsers including IE 8 provide native support to[ LocalStorage](http://caniuse.com/#search=localStorage)。

Your browser: <strong id="errorOutput" class="am-text-danger"></strong>
<strong id="store-test-success" class="am-text-success"></strong>

Different browsers have different maximum length of single record. This can be tested with following tests:

- [Web Storage Support Test](http://dev-test.nemikor.com/web-storage/support-test/)
- [Test of localStorage limits/quota](https://arty.name/localstorage.html)

## Attention

### Difference between original localstorage and store.js

Original localstorage:

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

store.js:

```javascript
store.set('myage', 24)
store.get('myage') === 24 // true

store.set('user', { name: 'marcus', likes: 'javascript' })
alert("Hi my name is " + store.get('user').name + "!") // Object

store.set('tags', ['javascript', 'localStorage', 'store.js'])
alert("We've got " + store.get('tags').length + " tags here") // Array
```

### Timeout

LocalStorage doesn't have timeout interface, but we can achieve the same effect by comparing current time with stored time.

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

## Reference

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
        $('#store-test-success').html('Pass test!');
      }
    }
  });
</script>
