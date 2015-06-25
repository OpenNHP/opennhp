'use strict';

var UI = require('./core');

/* jshint -W101, -W106 */
/**
 * Add to Homescreen v3.2.2
 * (c) 2015 Matteo Spinelli
 * @license: http://cubiq.org/license
 */

// Check for addEventListener browser support (prevent errors in IE<9)
var _eventListener = 'addEventListener' in window;

// Check if document is loaded, needed by autostart
var _DOMReady = false;
if (document.readyState === 'complete') {
  _DOMReady = true;
} else if (_eventListener) {
  window.addEventListener('load', loaded, false);
}

function loaded() {
  window.removeEventListener('load', loaded, false);
  _DOMReady = true;
}

// regex used to detect if app has been added to the homescreen
var _reSmartURL = /\/ath(\/)?$/;
var _reQueryString = /([\?&]ath=[^&]*$|&ath=[^&]*(&))/;

// singleton
var _instance;
function ath(options) {
  _instance = _instance || new ath.Class(options);

  return _instance;
}

// message in all supported languages
ath.intl = {
  en_us: {
    ios: 'To add this web app to the home screen: tap %icon and then <strong>Add to Home Screen</strong>.',
    android: 'To add this web app to the home screen open the browser option menu and tap on <strong>Add to homescreen</strong>. <small>The menu can be accessed by pressing the menu hardware button if your device has one, or by tapping the top right menu icon <span class="ath-action-icon">icon</span>.</small>'
  },

  zh_cn: {
    ios: '如要把应用程式加至主屏幕,请点击%icon, 然后<strong>加至主屏幕</strong>',
    android: 'To add this web app to the home screen open the browser option menu and tap on <strong>Add to homescreen</strong>. <small>The menu can be accessed by pressing the menu hardware button if your device has one, or by tapping the top right menu icon <span class="ath-action-icon">icon</span>.</small>'
  },

  zh_tw: {
    ios: '如要把應用程式加至主屏幕, 請點擊%icon, 然後<strong>加至主屏幕</strong>.',
    android: 'To add this web app to the home screen open the browser option menu and tap on <strong>Add to homescreen</strong>. <small>The menu can be accessed by pressing the menu hardware button if your device has one, or by tapping the top right menu icon <span class="ath-action-icon">icon</span>.</small>'
  }
};

// Add 2 characters language support (Android mostly)
for (var lang in ath.intl) {
  ath.intl[lang.substr(0, 2)] = ath.intl[lang];
}

// default options
ath.defaults = {
  appID: 'org.cubiq.addtohome',		// local storage name (no need to change)
  fontSize: 15,				// base font size, used to properly resize the popup based on viewport scale factor
  debug: false,				// override browser checks
  logging: false,				// log reasons for showing or not showing to js console; defaults to true when debug is true
  modal: false,				// prevent further actions until the message is closed
  mandatory: false,			// you can't proceed if you don't add the app to the homescreen
  autostart: true,			// show the message automatically
  skipFirstVisit: false,		// show only to returning visitors (ie: skip the first time you visit)
  startDelay: 1,				// display the message after that many seconds from page load
  lifespan: 15,				// life of the message in seconds
  displayPace: 1440,			// minutes before the message is shown again (0: display every time, default 24 hours)
  maxDisplayCount: 0,			// absolute maximum number of times the message will be shown to the user (0: no limit)
  icon: true,					// add touch icon to the message
  message: '',				// the message can be customized
  validLocation: [],			// list of pages where the message will be shown (array of regexes)
  onInit: null,				// executed on instance creation
  onShow: null,				// executed when the message is shown
  onRemove: null,				// executed when the message is removed
  onAdd: null,				// when the application is launched the first time from the homescreen (guesstimate)
  onPrivate: null,			// executed if user is in private mode
  privateModeOverride: false,	// show the message even in private mode (very rude)
  detectHomescreen: false		// try to detect if the site has been added to the homescreen (false | true | 'hash' | 'queryString' | 'smartURL')
};

// browser info and capability
var _ua = window.navigator.userAgent;

var _nav = window.navigator;
_extend(ath, {
  hasToken: document.location.hash == '#ath' || _reSmartURL.test(document.location.href) || _reQueryString.test(document.location.search),
  isRetina: window.devicePixelRatio && window.devicePixelRatio > 1,
  isIDevice: (/iphone|ipod|ipad/i).test(_ua),
  isMobileChrome: _ua.indexOf('Android') > -1 && (/Chrome\/[.0-9]*/).test(_ua) && _ua.indexOf("Version") == -1,
  isMobileIE: _ua.indexOf('Windows Phone') > -1,
  language: _nav.language && _nav.language.toLowerCase().replace('-', '_') || ''
});

// falls back to en_us if language is unsupported
ath.language = ath.language && ath.language in ath.intl ? ath.language : 'en_us';

ath.isMobileSafari = ath.isIDevice && _ua.indexOf('Safari') > -1 && _ua.indexOf('CriOS') < 0;
ath.OS = ath.isIDevice ? 'ios' : ath.isMobileChrome ? 'android' : ath.isMobileIE ? 'windows' : 'unsupported';

ath.OSVersion = _ua.match(/(OS|Android) (\d+[_\.]\d+)/);
ath.OSVersion = ath.OSVersion && ath.OSVersion[2] ? +ath.OSVersion[2].replace('_', '.') : 0;

ath.isStandalone = 'standalone' in window.navigator && window.navigator.standalone;
ath.isTablet = (ath.isMobileSafari && _ua.indexOf('iPad') > -1) || (ath.isMobileChrome && _ua.indexOf('Mobile') < 0);

ath.isCompatible = (ath.isMobileSafari && ath.OSVersion >= 6) || ath.isMobileChrome;	// TODO: add winphone

var _defaultSession = {
  lastDisplayTime: 0,			// last time we displayed the message
  returningVisitor: false,	// is this the first time you visit
  displayCount: 0,			// number of times the message has been shown
  optedout: false,			// has the user opted out
  added: false				// has been actually added to the homescreen
};

ath.removeSession = function(appID) {
  try {
    if (!localStorage) {
      throw new Error('localStorage is not defined');
    }

    localStorage.removeItem(appID || ath.defaults.appID);
  } catch (e) {
    // we are most likely in private mode
  }
};

ath.doLog = function(logStr) {
  if (this.options.logging) {
    console.log(logStr);
  }
};

ath.Class = function(options) {
  // class methods
  this.doLog = ath.doLog;

  // merge default options with user config
  this.options = _extend({}, ath.defaults);
  _extend(this.options, options);
  // override defaults that are dependent on each other
  if (options.debug && (typeof options.logging === "undefined")) {
    this.options.logging = true;
  }

  // IE<9 so exit (I hate you, really)
  if (!_eventListener) {
    return;
  }

  // normalize some options
  this.options.mandatory = this.options.mandatory && ( 'standalone' in window.navigator || this.options.debug );
  this.options.modal = this.options.modal || this.options.mandatory;
  if (this.options.mandatory) {
    this.options.startDelay = -0.5;		// make the popup hasty
  }
  this.options.detectHomescreen = this.options.detectHomescreen === true ? 'hash' : this.options.detectHomescreen;

  // setup the debug environment
  if (this.options.debug) {
    ath.isCompatible = true;
    ath.OS = typeof this.options.debug == 'string' ? this.options.debug : ath.OS == 'unsupported' ? 'android' : ath.OS;
    ath.OSVersion = ath.OS == 'ios' ? '8' : '4';
  }

  // the element the message will be appended to
  this.container = document.documentElement;

  // load session
  this.session = this.getItem(this.options.appID);
  this.session = this.session ? JSON.parse(this.session) : undefined;

  // user most likely came from a direct link containing our token, we don't need it and we remove it
  if (ath.hasToken && ( !ath.isCompatible || !this.session )) {
    ath.hasToken = false;
    _removeToken();
  }

  // the device is not supported
  if (!ath.isCompatible) {
    this.doLog("Add to homescreen: not displaying callout because device not supported");
    return;
  }

  this.session = this.session || _defaultSession;

  // check if we can use the local storage
  try {
    if (!localStorage) {
      throw new Error('localStorage is not defined');
    }

    localStorage.setItem(this.options.appID, JSON.stringify(this.session));
    ath.hasLocalStorage = true;
  } catch (e) {
    // we are most likely in private mode
    ath.hasLocalStorage = false;

    if (this.options.onPrivate) {
      this.options.onPrivate.call(this);
    }
  }

  // check if this is a valid location
  var isValidLocation = !this.options.validLocation.length;
  for (var i = this.options.validLocation.length; i--;) {
    if (this.options.validLocation[i].test(document.location.href)) {
      isValidLocation = true;
      break;
    }
  }

  // check compatibility with old versions of add to homescreen. Opt-out if an old session is found
  if (this.getItem('addToHome')) {
    this.optOut();
  }

  // critical errors:
  if (this.session.optedout) {
    this.doLog("Add to homescreen: not displaying callout because user opted out");
    return;
  }
  if (this.session.added) {
    this.doLog("Add to homescreen: not displaying callout because already added to the homescreen");
    return;
  }
  if (!isValidLocation) {
    this.doLog("Add to homescreen: not displaying callout because not a valid location");
    return;
  }

  // check if the app is in stand alone mode
  if (ath.isStandalone) {
    // execute the onAdd event if we haven't already
    if (!this.session.added) {
      this.session.added = true;
      this.updateSession();

      if (this.options.onAdd && ath.hasLocalStorage) {	// double check on localstorage to avoid multiple calls to the custom event
        this.options.onAdd.call(this);
      }
    }

    this.doLog("Add to homescreen: not displaying callout because in standalone mode");
    return;
  }

  // (try to) check if the page has been added to the homescreen
  if (this.options.detectHomescreen) {
    // the URL has the token, we are likely coming from the homescreen
    if (ath.hasToken) {
      _removeToken();		// we don't actually need the token anymore, we remove it to prevent redistribution

      // this is called the first time the user opens the app from the homescreen
      if (!this.session.added) {
        this.session.added = true;
        this.updateSession();

        if (this.options.onAdd && ath.hasLocalStorage) {	// double check on localstorage to avoid multiple calls to the custom event
          this.options.onAdd.call(this);
        }
      }

      this.doLog("Add to homescreen: not displaying callout because URL has token, so we are likely coming from homescreen");
      return;
    }

    // URL doesn't have the token, so add it
    if (this.options.detectHomescreen == 'hash') {
      history.replaceState('', window.document.title, document.location.href + '#ath');
    } else if (this.options.detectHomescreen == 'smartURL') {
      history.replaceState('', window.document.title, document.location.href.replace(/(\/)?$/, '/ath$1'));
    } else {
      history.replaceState('', window.document.title, document.location.href + (document.location.search ? '&' : '?' ) + 'ath=');
    }
  }

  // check if this is a returning visitor
  if (!this.session.returningVisitor) {
    this.session.returningVisitor = true;
    this.updateSession();

    // we do not show the message if this is your first visit
    if (this.options.skipFirstVisit) {
      this.doLog("Add to homescreen: not displaying callout because skipping first visit");
      return;
    }
  }

  // we do no show the message in private mode
  if (!this.options.privateModeOverride && !ath.hasLocalStorage) {
    this.doLog("Add to homescreen: not displaying callout because browser is in private mode");
    return;
  }

  // all checks passed, ready to display
  this.ready = true;

  if (this.options.onInit) {
    this.options.onInit.call(this);
  }

  if (this.options.autostart) {
    this.doLog("Add to homescreen: autostart displaying callout");
    this.show();
  }
};

ath.Class.prototype = {
  // event type to method conversion
  events: {
    load: '_delayedShow',
    error: '_delayedShow',
    orientationchange: 'resize',
    resize: 'resize',
    scroll: 'resize',
    click: 'remove',
    touchmove: '_preventDefault',
    transitionend: '_removeElements',
    webkitTransitionEnd: '_removeElements',
    MSTransitionEnd: '_removeElements'
  },

  handleEvent: function(e) {
    var type = this.events[e.type];
    if (type) {
      this[type](e);
    }
  },

  show: function(force) {
    // in autostart mode wait for the document to be ready
    if (this.options.autostart && !_DOMReady) {
      setTimeout(this.show.bind(this), 50);
      // we are not displaying callout because DOM not ready, but don't log that because
      // it would log too frequently
      return;
    }

    // message already on screen
    if (this.shown) {
      this.doLog("Add to homescreen: not displaying callout because already shown on screen");
      return;
    }

    var now = Date.now();
    var lastDisplayTime = this.session.lastDisplayTime;

    if (force !== true) {
      // this is needed if autostart is disabled and you programmatically call the show() method
      if (!this.ready) {
        this.doLog("Add to homescreen: not displaying callout because not ready");
        return;
      }

      // we obey the display pace (prevent the message to popup too often)
      if (now - lastDisplayTime < this.options.displayPace * 60000) {
        this.doLog("Add to homescreen: not displaying callout because displayed recently");
        return;
      }

      // obey the maximum number of display count
      if (this.options.maxDisplayCount && this.session.displayCount >= this.options.maxDisplayCount) {
        this.doLog("Add to homescreen: not displaying callout because displayed too many times already");
        return;
      }
    }

    this.shown = true;

    // increment the display count
    this.session.lastDisplayTime = now;
    this.session.displayCount++;
    this.updateSession();

    // try to get the highest resolution application icon
    if (!this.applicationIcon) {
      if (ath.OS == 'ios') {
        this.applicationIcon = document.querySelector('head link[rel^=apple-touch-icon][sizes="152x152"],head link[rel^=apple-touch-icon][sizes="144x144"],head link[rel^=apple-touch-icon][sizes="120x120"],head link[rel^=apple-touch-icon][sizes="114x114"],head link[rel^=apple-touch-icon]');
      } else {
        this.applicationIcon = document.querySelector('head link[rel^="shortcut icon"][sizes="196x196"],head link[rel^=apple-touch-icon]');
      }
    }

    var message = '';

    if (typeof this.options.message == 'object' && ath.language in this.options.message) {		// use custom language message
      message = this.options.message[ath.language][ath.OS];
    } else if (typeof this.options.message == 'object' && ath.OS in this.options.message) {		// use custom os message
      message = this.options.message[ath.OS];
    } else if (this.options.message in ath.intl) {				// you can force the locale
      message = ath.intl[this.options.message][ath.OS];
    } else if (this.options.message !== '') {						// use a custom message
      message = this.options.message;
    } else if (ath.OS in ath.intl[ath.language]) {				// otherwise we use our message
      message = ath.intl[ath.language][ath.OS];
    }

    // add the action icon
    message = '<p>' + message.replace('%icon', '<span class="ath-action-icon">icon</span>') + '</p>';

    // create the message container
    this.viewport = document.createElement('div');
    this.viewport.className = 'ath-viewport';
    if (this.options.modal) {
      this.viewport.className += ' ath-modal';
    }
    if (this.options.mandatory) {
      this.viewport.className += ' ath-mandatory';
    }
    this.viewport.style.position = 'absolute';

    // create the actual message element
    this.element = document.createElement('div');
    this.element.className = 'ath-container ath-' + ath.OS + ' ath-' + ath.OS + (ath.OSVersion + '').substr(0, 1) + ' ath-' + (ath.isTablet ? 'tablet' : 'phone');
    this.element.style.cssText = '-webkit-transition-property:-webkit-transform,opacity;-webkit-transition-duration:0s;-webkit-transition-timing-function:ease-out;transition-property:transform,opacity;transition-duration:0s;transition-timing-function:ease-out;';
    this.element.style.webkitTransform = 'translate3d(0,-' + window.innerHeight + 'px,0)';
    this.element.style.transform = 'translate3d(0,-' + window.innerHeight + 'px,0)';

    // add the application icon
    if (this.options.icon && this.applicationIcon) {
      this.element.className += ' ath-icon';
      this.img = document.createElement('img');
      this.img.className = 'ath-application-icon';
      this.img.addEventListener('load', this, false);
      this.img.addEventListener('error', this, false);

      this.img.src = this.applicationIcon.href;
      this.element.appendChild(this.img);
    }

    this.element.innerHTML += message;

    // we are not ready to show, place the message out of sight
    this.viewport.style.left = '-99999em';

    // attach all elements to the DOM
    this.viewport.appendChild(this.element);
    this.container.appendChild(this.viewport);

    // if we don't have to wait for an image to load, show the message right away
    if (this.img) {
      this.doLog("Add to homescreen: not displaying callout because waiting for img to load");
    } else {
      this._delayedShow();
    }
  },

  _delayedShow: function(e) {
    setTimeout(this._show.bind(this), this.options.startDelay * 1000 + 500);
  },

  _show: function() {
    var that = this;

    // update the viewport size and orientation
    this.updateViewport();

    // reposition/resize the message on orientation change
    window.addEventListener('resize', this, false);
    window.addEventListener('scroll', this, false);
    window.addEventListener('orientationchange', this, false);

    if (this.options.modal) {
      // lock any other interaction
      document.addEventListener('touchmove', this, true);
    }

    // Enable closing after 1 second
    if (!this.options.mandatory) {
      setTimeout(function() {
        that.element.addEventListener('click', that, true);
      }, 1000);
    }

    // kick the animation
    setTimeout(function() {
      that.element.style.webkitTransitionDuration = '1.2s';
      that.element.style.transitionDuration = '1.2s';
      that.element.style.webkitTransform = 'translate3d(0,0,0)';
      that.element.style.transform = 'translate3d(0,0,0)';
    }, 0);

    // set the destroy timer
    if (this.options.lifespan) {
      this.removeTimer = setTimeout(this.remove.bind(this), this.options.lifespan * 1000);
    }

    // fire the custom onShow event
    if (this.options.onShow) {
      this.options.onShow.call(this);
    }
  },

  remove: function() {
    clearTimeout(this.removeTimer);

    // clear up the event listeners
    if (this.img) {
      this.img.removeEventListener('load', this, false);
      this.img.removeEventListener('error', this, false);
    }

    window.removeEventListener('resize', this, false);
    window.removeEventListener('scroll', this, false);
    window.removeEventListener('orientationchange', this, false);
    document.removeEventListener('touchmove', this, true);
    this.element.removeEventListener('click', this, true);

    // remove the message element on transition end
    this.element.addEventListener('transitionend', this, false);
    this.element.addEventListener('webkitTransitionEnd', this, false);
    this.element.addEventListener('MSTransitionEnd', this, false);

    // start the fade out animation
    this.element.style.webkitTransitionDuration = '0.3s';
    this.element.style.opacity = '0';
  },

  _removeElements: function() {
    this.element.removeEventListener('transitionend', this, false);
    this.element.removeEventListener('webkitTransitionEnd', this, false);
    this.element.removeEventListener('MSTransitionEnd', this, false);

    // remove the message from the DOM
    this.container.removeChild(this.viewport);

    this.shown = false;

    // fire the custom onRemove event
    if (this.options.onRemove) {
      this.options.onRemove.call(this);
    }
  },

  updateViewport: function() {
    if (!this.shown) {
      return;
    }

    this.viewport.style.width = window.innerWidth + 'px';
    this.viewport.style.height = window.innerHeight + 'px';
    this.viewport.style.left = window.scrollX + 'px';
    this.viewport.style.top = window.scrollY + 'px';

    var clientWidth = document.documentElement.clientWidth;

    this.orientation = clientWidth > document.documentElement.clientHeight ? 'landscape' : 'portrait';

    var screenWidth = ath.OS == 'ios' ? this.orientation == 'portrait' ? screen.width : screen.height : screen.width;
    this.scale = screen.width > clientWidth ? 1 : screenWidth / window.innerWidth;

    this.element.style.fontSize = this.options.fontSize / this.scale + 'px';
  },

  resize: function() {
    clearTimeout(this.resizeTimer);
    this.resizeTimer = setTimeout(this.updateViewport.bind(this), 100);
  },

  updateSession: function() {
    if (ath.hasLocalStorage === false) {
      return;
    }

    if (localStorage) {
      localStorage.setItem(this.options.appID, JSON.stringify(this.session));
    }
  },

  clearSession: function() {
    this.session = _defaultSession;
    this.updateSession();
  },

  getItem: function(item) {
    try {
      if (!localStorage) {
        throw new Error('localStorage is not defined');
      }

      return localStorage.getItem(item);
    } catch (e) {
      // Preventing exception for some browsers when fetching localStorage key
      ath.hasLocalStorage = false;
    }
  },

  optOut: function() {
    this.session.optedout = true;
    this.updateSession();
  },

  optIn: function() {
    this.session.optedout = false;
    this.updateSession();
  },

  clearDisplayCount: function() {
    this.session.displayCount = 0;
    this.updateSession();
  },

  _preventDefault: function(e) {
    e.preventDefault();
    e.stopPropagation();
  }
};

// utility
function _extend(target, obj) {
  for (var i in obj) {
    target[i] = obj[i];
  }

  return target;
}

function _removeToken() {
  if (document.location.hash == '#ath') {
    history.replaceState('', window.document.title, document.location.href.split('#')[0]);
  }

  if (_reSmartURL.test(document.location.href)) {
    history.replaceState('', window.document.title, document.location.href.replace(_reSmartURL, '$1'));
  }

  if (_reQueryString.test(document.location.search)) {
    history.replaceState('', window.document.title, document.location.href.replace(_reQueryString, '$2'));
  }
}

/* jshint +W101, +W106 */

ath.VERSION = '3.2.2';

module.exports = UI.addToHomescreen = ath;
