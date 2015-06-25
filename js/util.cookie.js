'use strict';

var UI = require('./core');

var cookie = {
  get: function(name) {
    var cookieName = encodeURIComponent(name) + '=';
    var cookieStart = document.cookie.indexOf(cookieName);
    var cookieValue = null;
    var cookieEnd;

    if (cookieStart > -1) {
      cookieEnd = document.cookie.indexOf(';', cookieStart);
      if (cookieEnd == -1) {
        cookieEnd = document.cookie.length;
      }
      cookieValue = decodeURIComponent(document.cookie.substring(cookieStart +
      cookieName.length, cookieEnd));
    }

    return cookieValue;
  },

  set: function(name, value, expires, path, domain, secure) {
    var cookieText = encodeURIComponent(name) + '=' +
      encodeURIComponent(value);

    if (expires instanceof Date) {
      cookieText += '; expires=' + expires.toUTCString();
    }

    if (path) {
      cookieText += '; path=' + path;
    }

    if (domain) {
      cookieText += '; domain=' + domain;
    }

    if (secure) {
      cookieText += '; secure';
    }

    document.cookie = cookieText;
  },

  unset: function(name, path, domain, secure) {
    this.set(name, '', new Date(0), path, domain, secure);
  }
};

UI.utils = UI.utils || {};

module.exports = UI.utils.cookie = cookie;
