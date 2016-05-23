'use strict';

var UI = require('./core');

/**
 * store.js
 * @see https://github.com/marcuswestin/store.js
 * @license https://github.com/marcuswestin/store.js/blob/master/LICENSE
 */

var store = {};
var win = (typeof window != 'undefined' ? window : global);
var localStorageName = 'localStorage';
var storage;

store.disabled = false;
store.version = '1.3.20';

store.set = function(key, value) {
};

store.get = function(key, defaultVal) {
};

store.has = function(key) {
  return store.get(key) !== undefined;
};

store.remove = function(key) {
};

store.clear = function() {
};

store.transact = function(key, defaultVal, transactionFn) {
  if (transactionFn == null) {
    transactionFn = defaultVal;
    defaultVal = null;
  }
  if (defaultVal == null) {
    defaultVal = {};
  }

  var val = store.get(key, defaultVal);
  transactionFn(val);
  store.set(key, val);
};

store.getAll = function() {
};

store.forEach = function() {
};

store.serialize = function(value) {
  return JSON.stringify(value);
};

store.deserialize = function(value) {
  if (typeof value != 'string') {
    return undefined;
  }

  try {
    return JSON.parse(value);
  } catch (e) {
    return value || undefined;
  }
};

// Functions to encapsulate questionable FireFox 3.6.13 behavior
// when about.config::dom.storage.enabled === false
// See https://github.com/marcuswestin/store.js/issues#issue/13
function isLocalStorageNameSupported() {
  try {
    return (localStorageName in win && win[localStorageName]);
  } catch (err) {
    return false;
  }
}

if (isLocalStorageNameSupported()) {
  storage = win[localStorageName];

  store.set = function(key, val) {
    if (val === undefined) {
      return store.remove(key);
    }
    storage.setItem(key, store.serialize(val));
    return val;
  };

  store.get = function(key, defaultVal) {
    var val = store.deserialize(storage.getItem(key));
    return (val === undefined ? defaultVal : val);
  };

  store.remove = function(key) {
    storage.removeItem(key);
  };

  store.clear = function() {
    storage.clear();
  };

  store.getAll = function() {
    var ret = {};
    store.forEach(function(key, val) {
      ret[key] = val;
    });
    return ret;
  };

  store.forEach = function(callback) {
    for (var i = 0; i < storage.length; i++) {
      var key = storage.key(i);
      callback(key, store.get(key));
    }
  };
}

try {
  var testKey = '__storejs__';
  store.set(testKey, testKey);
  if (store.get(testKey) != testKey) {
    store.disabled = true;
  }
  store.remove(testKey);
} catch (e) {
  store.disabled = true;
}

store.enabled = !store.disabled;

module.exports = UI.store = store;
