var $ = require('jquery');
var UI = require('./core');

/**
 * @ver 1.1.0
 * @via https://github.com/aralejs/qrcode/blob/master/src/qrcode.js
 * @license http://aralejs.org/
 */

var qrcodeAlgObjCache = [];

/**
 * 二维码构造函数，主要用于绘制
 * @param  {Object} opt 传递参数
 * @return {String} qrcode
 * @constructor
 */
var QRCode = function(opt) {
  if (typeof opt === 'string') { // 只编码ASCII字符串
    opt = {
      text: opt
    };
  }
  // 设置默认参数
  this.options = $.extend({}, {
    text: "",
    render: "",
    width: 256,
    height: 256,
    correctLevel: 3,
    background: "#ffffff",
    foreground: "#000000"
  }, opt);

  // 使用QRCodeAlg创建二维码结构
  var qrCodeAlg = null;
  for (var i = 0, l = qrcodeAlgObjCache.length; i < l; i++) {
    if (qrcodeAlgObjCache[i].text == this.options.text && qrcodeAlgObjCache[i].text.correctLevel == this.options.correctLevel) {
      qrCodeAlg = qrcodeAlgObjCache[i].obj;
      break;
    }
  }
  if (i == l) {
    qrCodeAlg = new QRCodeAlg(this.options.text, this.options.correctLevel);
    qrcodeAlgObjCache.push({
      text: this.options.text,
      correctLevel: this.options.correctLevel,
      obj: qrCodeAlg
    });
  }

  if (this.options.render) {
    switch (this.options.render) {
      case "canvas":
        return this.createCanvas(qrCodeAlg);
      case "table":
        return this.createTable(qrCodeAlg);
      case "svg":
        return this.createSVG(qrCodeAlg);
      default:
        return this.createDefault(qrCodeAlg);
    }
  }
  return this.createDefault(qrCodeAlg);
};
/**
 * 使用Canvas来画二维码
 * @return {}
 */

QRCode.prototype.createDefault = function(qrCodeAlg) {
  var canvas = document.createElement('canvas');
  if (canvas.getContext)
    return this.createCanvas(qrCodeAlg);
  if (!!document.createElementNS && !!document.createElementNS(SVG_NS, 'svg').createSVGRect)
    return this.createSVG(qrCodeAlg);
  return this.createTable(qrCodeAlg);
};
QRCode.prototype.createCanvas = function(qrCodeAlg) {
  //创建canvas节点
  var canvas = document.createElement('canvas');
  canvas.width = this.options.width;
  canvas.height = this.options.height;
  var ctx = canvas.getContext('2d');

  //计算每个点的长宽
  var tileW = (this.options.width / qrCodeAlg.getModuleCount()).toPrecision(4);
  var tileH = this.options.height / qrCodeAlg.getModuleCount().toPrecision(4);

  //绘制
  for (var row = 0; row < qrCodeAlg.getModuleCount(); row++) {
    for (var col = 0; col < qrCodeAlg.getModuleCount(); col++) {
      ctx.fillStyle = qrCodeAlg.modules[row][col] ? this.options.foreground : this.options.background;
      var w = (Math.ceil((col + 1) * tileW) - Math.floor(col * tileW));
      var h = (Math.ceil((row + 1) * tileW) - Math.floor(row * tileW));
      ctx.fillRect(Math.round(col * tileW), Math.round(row * tileH), w, h);
    }
  }
  //返回绘制的节点
  return canvas;
};
/**
 * 使用table来绘制二维码
 * @return {}
 */
QRCode.prototype.createTable = function(qrCodeAlg) {
  //创建table节点
  var s = [];
  s.push('<table style="border:0px; margin:0px; padding:0px; border-collapse:collapse; background-color: ' +
  this.options.background +
  ';">');
  // 计算每个节点的长宽；取整，防止点之间出现分离
  var tileW = -1, tileH = -1, caculateW = -1, caculateH = -1;
  tileW = caculateW = Math.floor(this.options.width / qrCodeAlg.getModuleCount());
  tileH = caculateH = Math.floor(this.options.height / qrCodeAlg.getModuleCount());
  if (caculateW <= 0) {
    if (qrCodeAlg.getModuleCount() < 80) {
      tileW = 2;
    } else {
      tileW = 1;
    }
  }
  if (caculateH <= 0) {
    if (qrCodeAlg.getModuleCount() < 80) {
      tileH = 2;
    } else {
      tileH = 1;
    }
  }

  // 绘制二维码
  foreTd = '<td style="border:0px; margin:0px; padding:0px; width:' + tileW + 'px; background-color: ' + this.options.foreground + '"></td>',
    backTd = '<td style="border:0px; margin:0px; padding:0px; width:' + tileW + 'px; background-color: ' + this.options.background + '"></td>',
    l = qrCodeAlg.getModuleCount();

  for (var row = 0; row < l; row++) {
    s.push('<tr style="border:0px; margin:0px; padding:0px; height: ' + tileH + 'px">');
    for (var col = 0; col < l; col++) {
      s.push(qrCodeAlg.modules[row][col] ? foreTd : backTd);
    }
    s.push('</tr>');
  }
  s.push('</table>');
  var span = document.createElement("span");
  span.innerHTML = s.join('');

  return span.firstChild;
};

/**
 * 使用SVG开绘制二维码
 * @return {}
 */
QRCode.prototype.createSVG = function(qrCodeAlg) {
  var x, dx, y, dy,
    moduleCount = qrCodeAlg.getModuleCount(),
    scale = this.options.height / this.options.width,
    svg = '<svg xmlns="http://www.w3.org/2000/svg" '
      + 'width="' + this.options.width + 'px" height="' + this.options.height + 'px" '
      + 'viewbox="0 0 ' + moduleCount * 10 + ' ' + moduleCount * 10 * scale + '">',
    rectHead = '<path ',
    foreRect = ' style="stroke-width:0.5;stroke:' + this.options.foreground
      + ';fill:' + this.options.foreground + ';"></path>',
    backRect = ' style="stroke-width:0.5;stroke:' + this.options.background
      + ';fill:' + this.options.background + ';"></path>';

  // draw in the svg
  for (var row = 0; row < moduleCount; row++) {
    for (var col = 0; col < moduleCount; col++) {
      x = col * 10;
      y = row * 10 * scale;
      dx = (col + 1) * 10;
      dy = (row + 1) * 10 * scale;

      svg += rectHead + 'd="M ' + x + ',' + y
      + ' L ' + dx + ',' + y
      + ' L ' + dx + ',' + dy
      + ' L ' + x + ',' + dy
      + ' Z"';

      svg += qrCodeAlg.modules[row][col] ? foreRect : backRect;
    }
  }

  svg += '</svg>';

  // return just built svg
  return $(svg)[0];
};

/**
 * 获取单个字符的utf8编码
 * unicode BMP平面约65535个字符
 * @param {num} code
 * return {array}
 */
function unicodeFormat8(code) {
  // 1 byte
  if (code < 128) {
    return [code];
    // 2 bytes
  } else if (code < 2048) {
    c0 = 192 + (code >> 6);
    c1 = 128 + (code & 63);
    return [c0, c1];
    // 3 bytes
  } else {
    c0 = 224 + (code >> 12);
    c1 = 128 + (code >> 6 & 63);
    c2 = 128 + (code & 63);
    return [c0, c1, c2];
  }
}

/**
 * 获取字符串的utf8编码字节串
 * @param {string} string
 * @return {array}
 */
function getUTF8Bytes(string) {
  var utf8codes = [];
  for (var i = 0; i < string.length; i++) {
    var code = string.charCodeAt(i);
    var utf8 = unicodeFormat8(code);
    for (var j = 0; j < utf8.length; j++) {
      utf8codes.push(utf8[j]);
    }
  }
  return utf8codes;
}

/**
 * 二维码算法实现
 * @param {string} data              要编码的信息字符串
 * @param {num} errorCorrectLevel 纠错等级
 */
function QRCodeAlg(data, errorCorrectLevel) {
  this.typeNumber = -1; // 版本
  this.errorCorrectLevel = errorCorrectLevel;
  this.modules = null;  // 二维矩阵，存放最终结果
  this.moduleCount = 0; // 矩阵大小
  this.dataCache = null; // 数据缓存
  this.rsBlocks = null; // 版本数据信息
  this.totalDataCount = -1; // 可使用的数据量
  this.data = data;
  this.utf8bytes = getUTF8Bytes(data);
  this.make();
}

QRCodeAlg.prototype = {
  constructor: QRCodeAlg,
  /**
   * 获取二维码矩阵大小
   * @return {num} 矩阵大小
   */
  getModuleCount: function() {
    return this.moduleCount;
  },
  /**
   * 编码
   */
  make: function() {
    this.getRightType();
    this.dataCache = this.createData();
    this.createQrcode();
  },
  /**
   * 设置二位矩阵功能图形
   * @param  {bool} test 表示是否在寻找最好掩膜阶段
   * @param  {num} maskPattern 掩膜的版本
   */
  makeImpl: function(maskPattern) {

    this.moduleCount = this.typeNumber * 4 + 17;
    this.modules = new Array(this.moduleCount);

    for (var row = 0; row < this.moduleCount; row++) {

      this.modules[row] = new Array(this.moduleCount);

    }
    this.setupPositionProbePattern(0, 0);
    this.setupPositionProbePattern(this.moduleCount - 7, 0);
    this.setupPositionProbePattern(0, this.moduleCount - 7);
    this.setupPositionAdjustPattern();
    this.setupTimingPattern();
    this.setupTypeInfo(true, maskPattern);

    if (this.typeNumber >= 7) {
      this.setupTypeNumber(true);
    }
    this.mapData(this.dataCache, maskPattern);
  },
  /**
   * 设置二维码的位置探测图形
   * @param  {num} row 探测图形的中心横坐标
   * @param  {num} col 探测图形的中心纵坐标
   */
  setupPositionProbePattern: function(row, col) {

    for (var r = -1; r <= 7; r++) {

      if (row + r <= -1 || this.moduleCount <= row + r) continue;

      for (var c = -1; c <= 7; c++) {

        if (col + c <= -1 || this.moduleCount <= col + c) continue;

        if ((0 <= r && r <= 6 && (c == 0 || c == 6)) || (0 <= c && c <= 6 && (r == 0 || r == 6)) || (2 <= r && r <= 4 && 2 <= c && c <= 4)) {
          this.modules[row + r][col + c] = true;
        } else {
          this.modules[row + r][col + c] = false;
        }
      }
    }
  },
  /**
   * 创建二维码
   * @return {[type]} [description]
   */
  createQrcode: function() {

    var minLostPoint = 0;
    var pattern = 0;
    var bestModules = null;

    for (var i = 0; i < 8; i++) {

      this.makeImpl(i);

      var lostPoint = QRUtil.getLostPoint(this);
      if (i == 0 || minLostPoint > lostPoint) {
        minLostPoint = lostPoint;
        pattern = i;
        bestModules = this.modules;
      }
    }
    this.modules = bestModules;
    this.setupTypeInfo(false, pattern);

    if (this.typeNumber >= 7) {
      this.setupTypeNumber(false);
    }

  },
  /**
   * 设置定位图形
   * @return {[type]} [description]
   */
  setupTimingPattern: function() {

    for (var r = 8; r < this.moduleCount - 8; r++) {
      if (this.modules[r][6] != null) {
        continue;
      }
      this.modules[r][6] = (r % 2 == 0);

      if (this.modules[6][r] != null) {
        continue;
      }
      this.modules[6][r] = (r % 2 == 0);
    }
  },
  /**
   * 设置矫正图形
   * @return {[type]} [description]
   */
  setupPositionAdjustPattern: function() {

    var pos = QRUtil.getPatternPosition(this.typeNumber);

    for (var i = 0; i < pos.length; i++) {

      for (var j = 0; j < pos.length; j++) {

        var row = pos[i];
        var col = pos[j];

        if (this.modules[row][col] != null) {
          continue;
        }

        for (var r = -2; r <= 2; r++) {

          for (var c = -2; c <= 2; c++) {

            if (r == -2 || r == 2 || c == -2 || c == 2 || (r == 0 && c == 0)) {
              this.modules[row + r][col + c] = true;
            } else {
              this.modules[row + r][col + c] = false;
            }
          }
        }
      }
    }
  },
  /**
   * 设置版本信息（7以上版本才有）
   * @param  {bool} test 是否处于判断最佳掩膜阶段
   * @return {[type]}      [description]
   */
  setupTypeNumber: function(test) {

    var bits = QRUtil.getBCHTypeNumber(this.typeNumber);

    for (var i = 0; i < 18; i++) {
      var mod = (!test && ((bits >> i) & 1) == 1);
      this.modules[Math.floor(i / 3)][i % 3 + this.moduleCount - 8 - 3] = mod;
      this.modules[i % 3 + this.moduleCount - 8 - 3][Math.floor(i / 3)] = mod;
    }
  },
  /**
   * 设置格式信息（纠错等级和掩膜版本）
   * @param  {bool} test
   * @param  {num} maskPattern 掩膜版本
   * @return {}
   */
  setupTypeInfo: function(test, maskPattern) {

    var data = (QRErrorCorrectLevel[this.errorCorrectLevel] << 3) | maskPattern;
    var bits = QRUtil.getBCHTypeInfo(data);

    // vertical
    for (var i = 0; i < 15; i++) {

      var mod = (!test && ((bits >> i) & 1) == 1);

      if (i < 6) {
        this.modules[i][8] = mod;
      } else if (i < 8) {
        this.modules[i + 1][8] = mod;
      } else {
        this.modules[this.moduleCount - 15 + i][8] = mod;
      }

      // horizontal
      var mod = (!test && ((bits >> i) & 1) == 1);

      if (i < 8) {
        this.modules[8][this.moduleCount - i - 1] = mod;
      } else if (i < 9) {
        this.modules[8][15 - i - 1 + 1] = mod;
      } else {
        this.modules[8][15 - i - 1] = mod;
      }
    }

    // fixed module
    this.modules[this.moduleCount - 8][8] = (!test);

  },
  /**
   * 数据编码
   * @return {[type]} [description]
   */
  createData: function() {
    var buffer = new QRBitBuffer();
    var lengthBits = this.typeNumber > 9 ? 16 : 8;
    buffer.put(4, 4); //添加模式
    buffer.put(this.utf8bytes.length, lengthBits);
    for (var i = 0, l = this.utf8bytes.length; i < l; i++) {
      buffer.put(this.utf8bytes[i], 8);
    }
    if (buffer.length + 4 <= this.totalDataCount * 8) {
      buffer.put(0, 4);
    }

    // padding
    while (buffer.length % 8 != 0) {
      buffer.putBit(false);
    }

    // padding
    while (true) {

      if (buffer.length >= this.totalDataCount * 8) {
        break;
      }
      buffer.put(QRCodeAlg.PAD0, 8);

      if (buffer.length >= this.totalDataCount * 8) {
        break;
      }
      buffer.put(QRCodeAlg.PAD1, 8);
    }
    return this.createBytes(buffer);
  },
  /**
   * 纠错码编码
   * @param  {buffer} buffer 数据编码
   * @return {[type]}
   */
  createBytes: function(buffer) {

    var offset = 0;

    var maxDcCount = 0;
    var maxEcCount = 0;

    var length = this.rsBlock.length / 3;

    var rsBlocks = new Array();

    for (var i = 0; i < length; i++) {

      var count = this.rsBlock[i * 3 + 0];
      var totalCount = this.rsBlock[i * 3 + 1];
      var dataCount = this.rsBlock[i * 3 + 2];

      for (var j = 0; j < count; j++) {
        rsBlocks.push([dataCount, totalCount]);
      }
    }

    var dcdata = new Array(rsBlocks.length);
    var ecdata = new Array(rsBlocks.length);

    for (var r = 0; r < rsBlocks.length; r++) {

      var dcCount = rsBlocks[r][0];
      var ecCount = rsBlocks[r][1] - dcCount;

      maxDcCount = Math.max(maxDcCount, dcCount);
      maxEcCount = Math.max(maxEcCount, ecCount);

      dcdata[r] = new Array(dcCount);

      for (var i = 0; i < dcdata[r].length; i++) {
        dcdata[r][i] = 0xff & buffer.buffer[i + offset];
      }
      offset += dcCount;

      var rsPoly = QRUtil.getErrorCorrectPolynomial(ecCount);
      var rawPoly = new QRPolynomial(dcdata[r], rsPoly.getLength() - 1);

      var modPoly = rawPoly.mod(rsPoly);
      ecdata[r] = new Array(rsPoly.getLength() - 1);
      for (var i = 0; i < ecdata[r].length; i++) {
        var modIndex = i + modPoly.getLength() - ecdata[r].length;
        ecdata[r][i] = (modIndex >= 0) ? modPoly.get(modIndex) : 0;
      }
    }

    var data = new Array(this.totalDataCount);
    var index = 0;

    for (var i = 0; i < maxDcCount; i++) {
      for (var r = 0; r < rsBlocks.length; r++) {
        if (i < dcdata[r].length) {
          data[index++] = dcdata[r][i];
        }
      }
    }

    for (var i = 0; i < maxEcCount; i++) {
      for (var r = 0; r < rsBlocks.length; r++) {
        if (i < ecdata[r].length) {
          data[index++] = ecdata[r][i];
        }
      }
    }

    return data;

  },
  /**
   * 布置模块，构建最终信息
   * @param  {} data
   * @param  {} maskPattern
   * @return {}
   */
  mapData: function(data, maskPattern) {

    var inc = -1;
    var row = this.moduleCount - 1;
    var bitIndex = 7;
    var byteIndex = 0;

    for (var col = this.moduleCount - 1; col > 0; col -= 2) {

      if (col == 6) col--;

      while (true) {

        for (var c = 0; c < 2; c++) {

          if (this.modules[row][col - c] == null) {

            var dark = false;

            if (byteIndex < data.length) {
              dark = (((data[byteIndex] >>> bitIndex) & 1) == 1);
            }

            var mask = QRUtil.getMask(maskPattern, row, col - c);

            if (mask) {
              dark = !dark;
            }

            this.modules[row][col - c] = dark;
            bitIndex--;

            if (bitIndex == -1) {
              byteIndex++;
              bitIndex = 7;
            }
          }
        }

        row += inc;

        if (row < 0 || this.moduleCount <= row) {
          row -= inc;
          inc = -inc;
          break;
        }
      }
    }
  }

};
/**
 * 填充字段
 */
QRCodeAlg.PAD0 = 0xEC;
QRCodeAlg.PAD1 = 0x11;


//---------------------------------------------------------------------
// 纠错等级对应的编码
//---------------------------------------------------------------------

var QRErrorCorrectLevel = [1, 0, 3, 2];

//---------------------------------------------------------------------
// 掩膜版本
//---------------------------------------------------------------------

var QRMaskPattern = {
  PATTERN000: 0,
  PATTERN001: 1,
  PATTERN010: 2,
  PATTERN011: 3,
  PATTERN100: 4,
  PATTERN101: 5,
  PATTERN110: 6,
  PATTERN111: 7
};

//---------------------------------------------------------------------
// 工具类
//---------------------------------------------------------------------

var QRUtil = {

  /*
   每个版本矫正图形的位置
   */
  PATTERN_POSITION_TABLE: [
    [],
    [6, 18],
    [6, 22],
    [6, 26],
    [6, 30],
    [6, 34],
    [6, 22, 38],
    [6, 24, 42],
    [6, 26, 46],
    [6, 28, 50],
    [6, 30, 54],
    [6, 32, 58],
    [6, 34, 62],
    [6, 26, 46, 66],
    [6, 26, 48, 70],
    [6, 26, 50, 74],
    [6, 30, 54, 78],
    [6, 30, 56, 82],
    [6, 30, 58, 86],
    [6, 34, 62, 90],
    [6, 28, 50, 72, 94],
    [6, 26, 50, 74, 98],
    [6, 30, 54, 78, 102],
    [6, 28, 54, 80, 106],
    [6, 32, 58, 84, 110],
    [6, 30, 58, 86, 114],
    [6, 34, 62, 90, 118],
    [6, 26, 50, 74, 98, 122],
    [6, 30, 54, 78, 102, 126],
    [6, 26, 52, 78, 104, 130],
    [6, 30, 56, 82, 108, 134],
    [6, 34, 60, 86, 112, 138],
    [6, 30, 58, 86, 114, 142],
    [6, 34, 62, 90, 118, 146],
    [6, 30, 54, 78, 102, 126, 150],
    [6, 24, 50, 76, 102, 128, 154],
    [6, 28, 54, 80, 106, 132, 158],
    [6, 32, 58, 84, 110, 136, 162],
    [6, 26, 54, 82, 110, 138, 166],
    [6, 30, 58, 86, 114, 142, 170]
  ],

  G15: (1 << 10) | (1 << 8) | (1 << 5) | (1 << 4) | (1 << 2) | (1 << 1) | (1 << 0),
  G18: (1 << 12) | (1 << 11) | (1 << 10) | (1 << 9) | (1 << 8) | (1 << 5) | (1 << 2) | (1 << 0),
  G15_MASK: (1 << 14) | (1 << 12) | (1 << 10) | (1 << 4) | (1 << 1),

  /*
   BCH编码格式信息
   */
  getBCHTypeInfo: function(data) {
    var d = data << 10;
    while (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G15) >= 0) {
      d ^= (QRUtil.G15 << (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G15)));
    }
    return ((data << 10) | d) ^ QRUtil.G15_MASK;
  },
  /*
   BCH编码版本信息
   */
  getBCHTypeNumber: function(data) {
    var d = data << 12;
    while (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G18) >= 0) {
      d ^= (QRUtil.G18 << (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G18)));
    }
    return (data << 12) | d;
  },
  /*
   获取BCH位信息
   */
  getBCHDigit: function(data) {

    var digit = 0;

    while (data != 0) {
      digit++;
      data >>>= 1;
    }

    return digit;
  },
  /*
   获取版本对应的矫正图形位置
   */
  getPatternPosition: function(typeNumber) {
    return QRUtil.PATTERN_POSITION_TABLE[typeNumber - 1];
  },
  /*
   掩膜算法
   */
  getMask: function(maskPattern, i, j) {

    switch (maskPattern) {

      case QRMaskPattern.PATTERN000:
        return (i + j) % 2 == 0;
      case QRMaskPattern.PATTERN001:
        return i % 2 == 0;
      case QRMaskPattern.PATTERN010:
        return j % 3 == 0;
      case QRMaskPattern.PATTERN011:
        return (i + j) % 3 == 0;
      case QRMaskPattern.PATTERN100:
        return (Math.floor(i / 2) + Math.floor(j / 3)) % 2 == 0;
      case QRMaskPattern.PATTERN101:
        return (i * j) % 2 + (i * j) % 3 == 0;
      case QRMaskPattern.PATTERN110:
        return ((i * j) % 2 + (i * j) % 3) % 2 == 0;
      case QRMaskPattern.PATTERN111:
        return ((i * j) % 3 + (i + j) % 2) % 2 == 0;

      default:
        throw new Error("bad maskPattern:" + maskPattern);
    }
  },
  /*
   获取RS的纠错多项式
   */
  getErrorCorrectPolynomial: function(errorCorrectLength) {

    var a = new QRPolynomial([1], 0);

    for (var i = 0; i < errorCorrectLength; i++) {
      a = a.multiply(new QRPolynomial([1, QRMath.gexp(i)], 0));
    }

    return a;
  },
  /*
   获取评价
   */
  getLostPoint: function(qrCode) {

    var moduleCount = qrCode.getModuleCount(),
      lostPoint = 0,
      darkCount = 0;

    for (var row = 0; row < moduleCount; row++) {

      var sameCount = 0;
      var head = qrCode.modules[row][0];

      for (var col = 0; col < moduleCount; col++) {

        var current = qrCode.modules[row][col];

        //level 3 评价
        if (col < moduleCount - 6) {
          if (current && !qrCode.modules[row][col + 1] && qrCode.modules[row][col + 2] && qrCode.modules[row][col + 3] && qrCode.modules[row][col + 4] && !qrCode.modules[row][col + 5] && qrCode.modules[row][col + 6]) {
            if (col < moduleCount - 10) {
              if (qrCode.modules[row][col + 7] && qrCode.modules[row][col + 8] && qrCode.modules[row][col + 9] && qrCode.modules[row][col + 10]) {
                lostPoint += 40;
              }
            } else if (col > 3) {
              if (qrCode.modules[row][col - 1] && qrCode.modules[row][col - 2] && qrCode.modules[row][col - 3] && qrCode.modules[row][col - 4]) {
                lostPoint += 40;
              }
            }

          }
        }

        //level 2 评价
        if ((row < moduleCount - 1) && (col < moduleCount - 1)) {
          var count = 0;
          if (current) count++;
          if (qrCode.modules[row + 1][col]) count++;
          if (qrCode.modules[row][col + 1]) count++;
          if (qrCode.modules[row + 1][col + 1]) count++;
          if (count == 0 || count == 4) {
            lostPoint += 3;
          }
        }

        //level 1 评价
        if (head ^ current) {
          sameCount++;
        } else {
          head = current;
          if (sameCount >= 5) {
            lostPoint += (3 + sameCount - 5);
          }
          sameCount = 1;
        }

        //level 4 评价
        if (current) {
          darkCount++;
        }

      }
    }

    for (var col = 0; col < moduleCount; col++) {

      var sameCount = 0;
      var head = qrCode.modules[0][col];

      for (var row = 0; row < moduleCount; row++) {

        var current = qrCode.modules[row][col];

        //level 3 评价
        if (row < moduleCount - 6) {
          if (current && !qrCode.modules[row + 1][col] && qrCode.modules[row + 2][col] && qrCode.modules[row + 3][col] && qrCode.modules[row + 4][col] && !qrCode.modules[row + 5][col] && qrCode.modules[row + 6][col]) {
            if (row < moduleCount - 10) {
              if (qrCode.modules[row + 7][col] && qrCode.modules[row + 8][col] && qrCode.modules[row + 9][col] && qrCode.modules[row + 10][col]) {
                lostPoint += 40;
              }
            } else if (row > 3) {
              if (qrCode.modules[row - 1][col] && qrCode.modules[row - 2][col] && qrCode.modules[row - 3][col] && qrCode.modules[row - 4][col]) {
                lostPoint += 40;
              }
            }
          }
        }

        //level 1 评价
        if (head ^ current) {
          sameCount++;
        } else {
          head = current;
          if (sameCount >= 5) {
            lostPoint += (3 + sameCount - 5);
          }
          sameCount = 1;
        }

      }
    }

    // LEVEL4

    var ratio = Math.abs(100 * darkCount / moduleCount / moduleCount - 50) / 5;
    lostPoint += ratio * 10;

    return lostPoint;
  }

};


//---------------------------------------------------------------------
// QRMath使用的数学工具
//---------------------------------------------------------------------

var QRMath = {
  /*
   将n转化为a^m
   */
  glog: function(n) {

    if (n < 1) {
      throw new Error("glog(" + n + ")");
    }

    return QRMath.LOG_TABLE[n];
  },
  /*
   将a^m转化为n
   */
  gexp: function(n) {

    while (n < 0) {
      n += 255;
    }

    while (n >= 256) {
      n -= 255;
    }

    return QRMath.EXP_TABLE[n];
  },

  EXP_TABLE: new Array(256),

  LOG_TABLE: new Array(256)

};

for (var i = 0; i < 8; i++) {
  QRMath.EXP_TABLE[i] = 1 << i;
}
for (var i = 8; i < 256; i++) {
  QRMath.EXP_TABLE[i] = QRMath.EXP_TABLE[i - 4] ^ QRMath.EXP_TABLE[i - 5] ^ QRMath.EXP_TABLE[i - 6] ^ QRMath.EXP_TABLE[i - 8];
}
for (var i = 0; i < 255; i++) {
  QRMath.LOG_TABLE[QRMath.EXP_TABLE[i]] = i;
}

//---------------------------------------------------------------------
// QRPolynomial 多项式
//---------------------------------------------------------------------
/**
 * 多项式类
 * @param {Array} num   系数
 * @param {num} shift a^shift
 */
function QRPolynomial(num, shift) {

  if (num.length == undefined) {
    throw new Error(num.length + "/" + shift);
  }

  var offset = 0;

  while (offset < num.length && num[offset] == 0) {
    offset++;
  }

  this.num = new Array(num.length - offset + shift);
  for (var i = 0; i < num.length - offset; i++) {
    this.num[i] = num[i + offset];
  }
}

QRPolynomial.prototype = {

  get: function(index) {
    return this.num[index];
  },

  getLength: function() {
    return this.num.length;
  },
  /**
   * 多项式乘法
   * @param  {QRPolynomial} e 被乘多项式
   * @return {[type]}   [description]
   */
  multiply: function(e) {

    var num = new Array(this.getLength() + e.getLength() - 1);

    for (var i = 0; i < this.getLength(); i++) {
      for (var j = 0; j < e.getLength(); j++) {
        num[i + j] ^= QRMath.gexp(QRMath.glog(this.get(i)) + QRMath.glog(e.get(j)));
      }
    }

    return new QRPolynomial(num, 0);
  },
  /**
   * 多项式模运算
   * @param  {QRPolynomial} e 模多项式
   * @return {}
   */
  mod: function(e) {
    var tl = this.getLength(),
      el = e.getLength();
    if (tl - el < 0) {
      return this;
    }
    var num = new Array(tl);
    for (var i = 0; i < tl; i++) {
      num[i] = this.get(i);
    }
    while (num.length >= el) {
      var ratio = QRMath.glog(num[0]) - QRMath.glog(e.get(0));

      for (var i = 0; i < e.getLength(); i++) {
        num[i] ^= QRMath.gexp(QRMath.glog(e.get(i)) + ratio);
      }
      while (num[0] == 0) {
        num.shift();
      }
    }
    return new QRPolynomial(num, 0);
  }
};

//---------------------------------------------------------------------
// RS_BLOCK_TABLE
//---------------------------------------------------------------------
/*
 二维码各个版本信息[块数, 每块中的数据块数, 每块中的信息块数]
 */
var RS_BLOCK_TABLE = [

  // L
  // M
  // Q
  // H

  // 1
  [1, 26, 19],
  [1, 26, 16],
  [1, 26, 13],
  [1, 26, 9],

  // 2
  [1, 44, 34],
  [1, 44, 28],
  [1, 44, 22],
  [1, 44, 16],

  // 3
  [1, 70, 55],
  [1, 70, 44],
  [2, 35, 17],
  [2, 35, 13],

  // 4
  [1, 100, 80],
  [2, 50, 32],
  [2, 50, 24],
  [4, 25, 9],

  // 5
  [1, 134, 108],
  [2, 67, 43],
  [2, 33, 15, 2, 34, 16],
  [2, 33, 11, 2, 34, 12],

  // 6
  [2, 86, 68],
  [4, 43, 27],
  [4, 43, 19],
  [4, 43, 15],

  // 7
  [2, 98, 78],
  [4, 49, 31],
  [2, 32, 14, 4, 33, 15],
  [4, 39, 13, 1, 40, 14],

  // 8
  [2, 121, 97],
  [2, 60, 38, 2, 61, 39],
  [4, 40, 18, 2, 41, 19],
  [4, 40, 14, 2, 41, 15],

  // 9
  [2, 146, 116],
  [3, 58, 36, 2, 59, 37],
  [4, 36, 16, 4, 37, 17],
  [4, 36, 12, 4, 37, 13],

  // 10
  [2, 86, 68, 2, 87, 69],
  [4, 69, 43, 1, 70, 44],
  [6, 43, 19, 2, 44, 20],
  [6, 43, 15, 2, 44, 16],

  // 11
  [4, 101, 81],
  [1, 80, 50, 4, 81, 51],
  [4, 50, 22, 4, 51, 23],
  [3, 36, 12, 8, 37, 13],

  // 12
  [2, 116, 92, 2, 117, 93],
  [6, 58, 36, 2, 59, 37],
  [4, 46, 20, 6, 47, 21],
  [7, 42, 14, 4, 43, 15],

  // 13
  [4, 133, 107],
  [8, 59, 37, 1, 60, 38],
  [8, 44, 20, 4, 45, 21],
  [12, 33, 11, 4, 34, 12],

  // 14
  [3, 145, 115, 1, 146, 116],
  [4, 64, 40, 5, 65, 41],
  [11, 36, 16, 5, 37, 17],
  [11, 36, 12, 5, 37, 13],

  // 15
  [5, 109, 87, 1, 110, 88],
  [5, 65, 41, 5, 66, 42],
  [5, 54, 24, 7, 55, 25],
  [11, 36, 12],

  // 16
  [5, 122, 98, 1, 123, 99],
  [7, 73, 45, 3, 74, 46],
  [15, 43, 19, 2, 44, 20],
  [3, 45, 15, 13, 46, 16],

  // 17
  [1, 135, 107, 5, 136, 108],
  [10, 74, 46, 1, 75, 47],
  [1, 50, 22, 15, 51, 23],
  [2, 42, 14, 17, 43, 15],

  // 18
  [5, 150, 120, 1, 151, 121],
  [9, 69, 43, 4, 70, 44],
  [17, 50, 22, 1, 51, 23],
  [2, 42, 14, 19, 43, 15],

  // 19
  [3, 141, 113, 4, 142, 114],
  [3, 70, 44, 11, 71, 45],
  [17, 47, 21, 4, 48, 22],
  [9, 39, 13, 16, 40, 14],

  // 20
  [3, 135, 107, 5, 136, 108],
  [3, 67, 41, 13, 68, 42],
  [15, 54, 24, 5, 55, 25],
  [15, 43, 15, 10, 44, 16],

  // 21
  [4, 144, 116, 4, 145, 117],
  [17, 68, 42],
  [17, 50, 22, 6, 51, 23],
  [19, 46, 16, 6, 47, 17],

  // 22
  [2, 139, 111, 7, 140, 112],
  [17, 74, 46],
  [7, 54, 24, 16, 55, 25],
  [34, 37, 13],

  // 23
  [4, 151, 121, 5, 152, 122],
  [4, 75, 47, 14, 76, 48],
  [11, 54, 24, 14, 55, 25],
  [16, 45, 15, 14, 46, 16],

  // 24
  [6, 147, 117, 4, 148, 118],
  [6, 73, 45, 14, 74, 46],
  [11, 54, 24, 16, 55, 25],
  [30, 46, 16, 2, 47, 17],

  // 25
  [8, 132, 106, 4, 133, 107],
  [8, 75, 47, 13, 76, 48],
  [7, 54, 24, 22, 55, 25],
  [22, 45, 15, 13, 46, 16],

  // 26
  [10, 142, 114, 2, 143, 115],
  [19, 74, 46, 4, 75, 47],
  [28, 50, 22, 6, 51, 23],
  [33, 46, 16, 4, 47, 17],

  // 27
  [8, 152, 122, 4, 153, 123],
  [22, 73, 45, 3, 74, 46],
  [8, 53, 23, 26, 54, 24],
  [12, 45, 15, 28, 46, 16],

  // 28
  [3, 147, 117, 10, 148, 118],
  [3, 73, 45, 23, 74, 46],
  [4, 54, 24, 31, 55, 25],
  [11, 45, 15, 31, 46, 16],

  // 29
  [7, 146, 116, 7, 147, 117],
  [21, 73, 45, 7, 74, 46],
  [1, 53, 23, 37, 54, 24],
  [19, 45, 15, 26, 46, 16],

  // 30
  [5, 145, 115, 10, 146, 116],
  [19, 75, 47, 10, 76, 48],
  [15, 54, 24, 25, 55, 25],
  [23, 45, 15, 25, 46, 16],

  // 31
  [13, 145, 115, 3, 146, 116],
  [2, 74, 46, 29, 75, 47],
  [42, 54, 24, 1, 55, 25],
  [23, 45, 15, 28, 46, 16],

  // 32
  [17, 145, 115],
  [10, 74, 46, 23, 75, 47],
  [10, 54, 24, 35, 55, 25],
  [19, 45, 15, 35, 46, 16],

  // 33
  [17, 145, 115, 1, 146, 116],
  [14, 74, 46, 21, 75, 47],
  [29, 54, 24, 19, 55, 25],
  [11, 45, 15, 46, 46, 16],

  // 34
  [13, 145, 115, 6, 146, 116],
  [14, 74, 46, 23, 75, 47],
  [44, 54, 24, 7, 55, 25],
  [59, 46, 16, 1, 47, 17],

  // 35
  [12, 151, 121, 7, 152, 122],
  [12, 75, 47, 26, 76, 48],
  [39, 54, 24, 14, 55, 25],
  [22, 45, 15, 41, 46, 16],

  // 36
  [6, 151, 121, 14, 152, 122],
  [6, 75, 47, 34, 76, 48],
  [46, 54, 24, 10, 55, 25],
  [2, 45, 15, 64, 46, 16],

  // 37
  [17, 152, 122, 4, 153, 123],
  [29, 74, 46, 14, 75, 47],
  [49, 54, 24, 10, 55, 25],
  [24, 45, 15, 46, 46, 16],

  // 38
  [4, 152, 122, 18, 153, 123],
  [13, 74, 46, 32, 75, 47],
  [48, 54, 24, 14, 55, 25],
  [42, 45, 15, 32, 46, 16],

  // 39
  [20, 147, 117, 4, 148, 118],
  [40, 75, 47, 7, 76, 48],
  [43, 54, 24, 22, 55, 25],
  [10, 45, 15, 67, 46, 16],

  // 40
  [19, 148, 118, 6, 149, 119],
  [18, 75, 47, 31, 76, 48],
  [34, 54, 24, 34, 55, 25],
  [20, 45, 15, 61, 46, 16]
];

/**
 * 根据数据获取对应版本
 * @return {[type]} [description]
 */
QRCodeAlg.prototype.getRightType = function() {
  for (var typeNumber = 1; typeNumber < 41; typeNumber++) {
    var rsBlock = RS_BLOCK_TABLE[(typeNumber - 1) * 4 + this.errorCorrectLevel];
    if (rsBlock == undefined) {
      throw new Error("bad rs block @ typeNumber:" + typeNumber + "/errorCorrectLevel:" + this.errorCorrectLevel);
    }
    var length = rsBlock.length / 3;
    var totalDataCount = 0;
    for (var i = 0; i < length; i++) {
      var count = rsBlock[i * 3 + 0];
      var dataCount = rsBlock[i * 3 + 2];
      totalDataCount += dataCount * count;
    }

    var lengthBytes = typeNumber > 9 ? 2 : 1;
    if (this.utf8bytes.length + lengthBytes < totalDataCount || typeNumber == 40) {
      this.typeNumber = typeNumber;
      this.rsBlock = rsBlock;
      this.totalDataCount = totalDataCount;
      break;
    }
  }
};

//---------------------------------------------------------------------
// QRBitBuffer
//---------------------------------------------------------------------

function QRBitBuffer() {
  this.buffer = new Array();
  this.length = 0;
}

QRBitBuffer.prototype = {

  get: function(index) {
    var bufIndex = Math.floor(index / 8);
    return ((this.buffer[bufIndex] >>> (7 - index % 8)) & 1);
  },

  put: function(num, length) {
    for (var i = 0; i < length; i++) {
      this.putBit(((num >>> (length - i - 1)) & 1));
    }
  },

  putBit: function(bit) {

    var bufIndex = Math.floor(this.length / 8);
    if (this.buffer.length <= bufIndex) {
      this.buffer.push(0);
    }

    if (bit) {
      this.buffer[bufIndex] |= (0x80 >>> (this.length % 8));
    }

    this.length++;
  }
};
/**
 * 获取单个字符的utf8编码
 * unicode BMP平面约65535个字符
 * @param {num} code
 * return {array}
 */
function unicodeFormat8(code) {
  // 1 byte
  if (code < 128) {
    return [code];
    // 2 bytes
  } else if (code < 2048) {
    c0 = 192 + (code >> 6);
    c1 = 128 + (code & 63);
    return [c0, c1];
    // 3 bytes
  } else {
    c0 = 224 + (code >> 12);
    c1 = 128 + (code >> 6 & 63);
    c2 = 128 + (code & 63);
    return [c0, c1, c2];
  }
}

/**
 * 获取字符串的utf8编码字节串
 * @param {string} string
 * @return {array}
 */
function getUTF8Bytes(string) {
  var utf8codes = [];
  for (var i = 0; i < string.length; i++) {
    var code = string.charCodeAt(i);
    var utf8 = unicodeFormat8(code);
    for (var j = 0; j < utf8.length; j++) {
      utf8codes.push(utf8[j]);
    }
  }
  return utf8codes;
}

/**
 * 二维码算法实现
 * @param {string} data              要编码的信息字符串
 * @param {num} errorCorrectLevel 纠错等级
 */
function QRCodeAlg(data, errorCorrectLevel) {
  this.typeNumber = -1; //版本
  this.errorCorrectLevel = errorCorrectLevel;
  this.modules = null;  //二维矩阵，存放最终结果
  this.moduleCount = 0; //矩阵大小
  this.dataCache = null; //数据缓存
  this.rsBlocks = null; //版本数据信息
  this.totalDataCount = -1; //可使用的数据量
  this.data = data;
  this.utf8bytes = getUTF8Bytes(data);
  this.make();
}

QRCodeAlg.prototype = {
  constructor: QRCodeAlg,
  /**
   * 获取二维码矩阵大小
   * @return {num} 矩阵大小
   */
  getModuleCount: function() {
    return this.moduleCount;
  },
  /**
   * 编码
   */
  make: function() {
    this.getRightType();
    this.dataCache = this.createData();
    this.createQrcode();
  },
  /**
   * 设置二位矩阵功能图形
   * @param  {bool} test 表示是否在寻找最好掩膜阶段
   * @param  {num} maskPattern 掩膜的版本
   */
  makeImpl: function(maskPattern) {

    this.moduleCount = this.typeNumber * 4 + 17;
    this.modules = new Array(this.moduleCount);

    for (var row = 0; row < this.moduleCount; row++) {

      this.modules[row] = new Array(this.moduleCount);

    }
    this.setupPositionProbePattern(0, 0);
    this.setupPositionProbePattern(this.moduleCount - 7, 0);
    this.setupPositionProbePattern(0, this.moduleCount - 7);
    this.setupPositionAdjustPattern();
    this.setupTimingPattern();
    this.setupTypeInfo(true, maskPattern);

    if (this.typeNumber >= 7) {
      this.setupTypeNumber(true);
    }
    this.mapData(this.dataCache, maskPattern);
  },
  /**
   * 设置二维码的位置探测图形
   * @param  {num} row 探测图形的中心横坐标
   * @param  {num} col 探测图形的中心纵坐标
   */
  setupPositionProbePattern: function(row, col) {

    for (var r = -1; r <= 7; r++) {

      if (row + r <= -1 || this.moduleCount <= row + r) continue;

      for (var c = -1; c <= 7; c++) {

        if (col + c <= -1 || this.moduleCount <= col + c) continue;

        if ((0 <= r && r <= 6 && (c == 0 || c == 6)) || (0 <= c && c <= 6 && (r == 0 || r == 6)) || (2 <= r && r <= 4 && 2 <= c && c <= 4)) {
          this.modules[row + r][col + c] = true;
        } else {
          this.modules[row + r][col + c] = false;
        }
      }
    }
  },
  /**
   * 创建二维码
   * @return {[type]} [description]
   */
  createQrcode: function() {

    var minLostPoint = 0;
    var pattern = 0;
    var bestModules = null;

    for (var i = 0; i < 8; i++) {

      this.makeImpl(i);

      var lostPoint = QRUtil.getLostPoint(this);
      if (i == 0 || minLostPoint > lostPoint) {
        minLostPoint = lostPoint;
        pattern = i;
        bestModules = this.modules;
      }
    }
    this.modules = bestModules;
    this.setupTypeInfo(false, pattern);

    if (this.typeNumber >= 7) {
      this.setupTypeNumber(false);
    }

  },
  /**
   * 设置定位图形
   * @return {[type]} [description]
   */
  setupTimingPattern: function() {

    for (var r = 8; r < this.moduleCount - 8; r++) {
      if (this.modules[r][6] != null) {
        continue;
      }
      this.modules[r][6] = (r % 2 == 0);

      if (this.modules[6][r] != null) {
        continue;
      }
      this.modules[6][r] = (r % 2 == 0);
    }
  },
  /**
   * 设置矫正图形
   * @return {[type]} [description]
   */
  setupPositionAdjustPattern: function() {

    var pos = QRUtil.getPatternPosition(this.typeNumber);

    for (var i = 0; i < pos.length; i++) {

      for (var j = 0; j < pos.length; j++) {

        var row = pos[i];
        var col = pos[j];

        if (this.modules[row][col] != null) {
          continue;
        }

        for (var r = -2; r <= 2; r++) {

          for (var c = -2; c <= 2; c++) {

            if (r == -2 || r == 2 || c == -2 || c == 2 || (r == 0 && c == 0)) {
              this.modules[row + r][col + c] = true;
            } else {
              this.modules[row + r][col + c] = false;
            }
          }
        }
      }
    }
  },
  /**
   * 设置版本信息（7以上版本才有）
   * @param  {bool} test 是否处于判断最佳掩膜阶段
   * @return {[type]}      [description]
   */
  setupTypeNumber: function(test) {

    var bits = QRUtil.getBCHTypeNumber(this.typeNumber);

    for (var i = 0; i < 18; i++) {
      var mod = (!test && ((bits >> i) & 1) == 1);
      this.modules[Math.floor(i / 3)][i % 3 + this.moduleCount - 8 - 3] = mod;
      this.modules[i % 3 + this.moduleCount - 8 - 3][Math.floor(i / 3)] = mod;
    }
  },
  /**
   * 设置格式信息（纠错等级和掩膜版本）
   * @param  {bool} test
   * @param  {num} maskPattern 掩膜版本
   * @return {}
   */
  setupTypeInfo: function(test, maskPattern) {

    var data = (QRErrorCorrectLevel[this.errorCorrectLevel] << 3) | maskPattern;
    var bits = QRUtil.getBCHTypeInfo(data);

    // vertical
    for (var i = 0; i < 15; i++) {

      var mod = (!test && ((bits >> i) & 1) == 1);

      if (i < 6) {
        this.modules[i][8] = mod;
      } else if (i < 8) {
        this.modules[i + 1][8] = mod;
      } else {
        this.modules[this.moduleCount - 15 + i][8] = mod;
      }

      // horizontal
      var mod = (!test && ((bits >> i) & 1) == 1);

      if (i < 8) {
        this.modules[8][this.moduleCount - i - 1] = mod;
      } else if (i < 9) {
        this.modules[8][15 - i - 1 + 1] = mod;
      } else {
        this.modules[8][15 - i - 1] = mod;
      }
    }

    // fixed module
    this.modules[this.moduleCount - 8][8] = (!test);

  },
  /**
   * 数据编码
   * @return {[type]} [description]
   */
  createData: function() {
    var buffer = new QRBitBuffer();
    var lengthBits = this.typeNumber > 9 ? 16 : 8;
    buffer.put(4, 4); //添加模式
    buffer.put(this.utf8bytes.length, lengthBits);
    for (var i = 0, l = this.utf8bytes.length; i < l; i++) {
      buffer.put(this.utf8bytes[i], 8);
    }
    if (buffer.length + 4 <= this.totalDataCount * 8) {
      buffer.put(0, 4);
    }

    // padding
    while (buffer.length % 8 != 0) {
      buffer.putBit(false);
    }

    // padding
    while (true) {

      if (buffer.length >= this.totalDataCount * 8) {
        break;
      }
      buffer.put(QRCodeAlg.PAD0, 8);

      if (buffer.length >= this.totalDataCount * 8) {
        break;
      }
      buffer.put(QRCodeAlg.PAD1, 8);
    }
    return this.createBytes(buffer);
  },
  /**
   * 纠错码编码
   * @param  {buffer} buffer 数据编码
   * @return {[type]}
   */
  createBytes: function(buffer) {

    var offset = 0;

    var maxDcCount = 0;
    var maxEcCount = 0;

    var length = this.rsBlock.length / 3;

    var rsBlocks = new Array();

    for (var i = 0; i < length; i++) {

      var count = this.rsBlock[i * 3 + 0];
      var totalCount = this.rsBlock[i * 3 + 1];
      var dataCount = this.rsBlock[i * 3 + 2];

      for (var j = 0; j < count; j++) {
        rsBlocks.push([dataCount, totalCount]);
      }
    }

    var dcdata = new Array(rsBlocks.length);
    var ecdata = new Array(rsBlocks.length);

    for (var r = 0; r < rsBlocks.length; r++) {

      var dcCount = rsBlocks[r][0];
      var ecCount = rsBlocks[r][1] - dcCount;

      maxDcCount = Math.max(maxDcCount, dcCount);
      maxEcCount = Math.max(maxEcCount, ecCount);

      dcdata[r] = new Array(dcCount);

      for (var i = 0; i < dcdata[r].length; i++) {
        dcdata[r][i] = 0xff & buffer.buffer[i + offset];
      }
      offset += dcCount;

      var rsPoly = QRUtil.getErrorCorrectPolynomial(ecCount);
      var rawPoly = new QRPolynomial(dcdata[r], rsPoly.getLength() - 1);

      var modPoly = rawPoly.mod(rsPoly);
      ecdata[r] = new Array(rsPoly.getLength() - 1);
      for (var i = 0; i < ecdata[r].length; i++) {
        var modIndex = i + modPoly.getLength() - ecdata[r].length;
        ecdata[r][i] = (modIndex >= 0) ? modPoly.get(modIndex) : 0;
      }
    }

    var data = new Array(this.totalDataCount);
    var index = 0;

    for (var i = 0; i < maxDcCount; i++) {
      for (var r = 0; r < rsBlocks.length; r++) {
        if (i < dcdata[r].length) {
          data[index++] = dcdata[r][i];
        }
      }
    }

    for (var i = 0; i < maxEcCount; i++) {
      for (var r = 0; r < rsBlocks.length; r++) {
        if (i < ecdata[r].length) {
          data[index++] = ecdata[r][i];
        }
      }
    }

    return data;

  },
  /**
   * 布置模块，构建最终信息
   * @param  {} data
   * @param  {} maskPattern
   * @return {}
   */
  mapData: function(data, maskPattern) {

    var inc = -1;
    var row = this.moduleCount - 1;
    var bitIndex = 7;
    var byteIndex = 0;

    for (var col = this.moduleCount - 1; col > 0; col -= 2) {

      if (col == 6) col--;

      while (true) {

        for (var c = 0; c < 2; c++) {

          if (this.modules[row][col - c] == null) {

            var dark = false;

            if (byteIndex < data.length) {
              dark = (((data[byteIndex] >>> bitIndex) & 1) == 1);
            }

            var mask = QRUtil.getMask(maskPattern, row, col - c);

            if (mask) {
              dark = !dark;
            }

            this.modules[row][col - c] = dark;
            bitIndex--;

            if (bitIndex == -1) {
              byteIndex++;
              bitIndex = 7;
            }
          }
        }

        row += inc;

        if (row < 0 || this.moduleCount <= row) {
          row -= inc;
          inc = -inc;
          break;
        }
      }
    }
  }

};
/**
 * 填充字段
 */
QRCodeAlg.PAD0 = 0xEC;
QRCodeAlg.PAD1 = 0x11;


//---------------------------------------------------------------------
// 纠错等级对应的编码
//---------------------------------------------------------------------

var QRErrorCorrectLevel = [1, 0, 3, 2];

//---------------------------------------------------------------------
// 掩膜版本
//---------------------------------------------------------------------

var QRMaskPattern = {
  PATTERN000: 0,
  PATTERN001: 1,
  PATTERN010: 2,
  PATTERN011: 3,
  PATTERN100: 4,
  PATTERN101: 5,
  PATTERN110: 6,
  PATTERN111: 7
};

//---------------------------------------------------------------------
// 工具类
//---------------------------------------------------------------------

var QRUtil = {

  /*
   每个版本矫正图形的位置
   */
  PATTERN_POSITION_TABLE: [
    [],
    [6, 18],
    [6, 22],
    [6, 26],
    [6, 30],
    [6, 34],
    [6, 22, 38],
    [6, 24, 42],
    [6, 26, 46],
    [6, 28, 50],
    [6, 30, 54],
    [6, 32, 58],
    [6, 34, 62],
    [6, 26, 46, 66],
    [6, 26, 48, 70],
    [6, 26, 50, 74],
    [6, 30, 54, 78],
    [6, 30, 56, 82],
    [6, 30, 58, 86],
    [6, 34, 62, 90],
    [6, 28, 50, 72, 94],
    [6, 26, 50, 74, 98],
    [6, 30, 54, 78, 102],
    [6, 28, 54, 80, 106],
    [6, 32, 58, 84, 110],
    [6, 30, 58, 86, 114],
    [6, 34, 62, 90, 118],
    [6, 26, 50, 74, 98, 122],
    [6, 30, 54, 78, 102, 126],
    [6, 26, 52, 78, 104, 130],
    [6, 30, 56, 82, 108, 134],
    [6, 34, 60, 86, 112, 138],
    [6, 30, 58, 86, 114, 142],
    [6, 34, 62, 90, 118, 146],
    [6, 30, 54, 78, 102, 126, 150],
    [6, 24, 50, 76, 102, 128, 154],
    [6, 28, 54, 80, 106, 132, 158],
    [6, 32, 58, 84, 110, 136, 162],
    [6, 26, 54, 82, 110, 138, 166],
    [6, 30, 58, 86, 114, 142, 170]
  ],

  G15: (1 << 10) | (1 << 8) | (1 << 5) | (1 << 4) | (1 << 2) | (1 << 1) | (1 << 0),
  G18: (1 << 12) | (1 << 11) | (1 << 10) | (1 << 9) | (1 << 8) | (1 << 5) | (1 << 2) | (1 << 0),
  G15_MASK: (1 << 14) | (1 << 12) | (1 << 10) | (1 << 4) | (1 << 1),

  /*
   BCH编码格式信息
   */
  getBCHTypeInfo: function(data) {
    var d = data << 10;
    while (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G15) >= 0) {
      d ^= (QRUtil.G15 << (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G15)));
    }
    return ((data << 10) | d) ^ QRUtil.G15_MASK;
  },
  /*
   BCH编码版本信息
   */
  getBCHTypeNumber: function(data) {
    var d = data << 12;
    while (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G18) >= 0) {
      d ^= (QRUtil.G18 << (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G18)));
    }
    return (data << 12) | d;
  },
  /*
   获取BCH位信息
   */
  getBCHDigit: function(data) {

    var digit = 0;

    while (data != 0) {
      digit++;
      data >>>= 1;
    }

    return digit;
  },
  /*
   获取版本对应的矫正图形位置
   */
  getPatternPosition: function(typeNumber) {
    return QRUtil.PATTERN_POSITION_TABLE[typeNumber - 1];
  },
  /*
   掩膜算法
   */
  getMask: function(maskPattern, i, j) {

    switch (maskPattern) {

      case QRMaskPattern.PATTERN000:
        return (i + j) % 2 == 0;
      case QRMaskPattern.PATTERN001:
        return i % 2 == 0;
      case QRMaskPattern.PATTERN010:
        return j % 3 == 0;
      case QRMaskPattern.PATTERN011:
        return (i + j) % 3 == 0;
      case QRMaskPattern.PATTERN100:
        return (Math.floor(i / 2) + Math.floor(j / 3)) % 2 == 0;
      case QRMaskPattern.PATTERN101:
        return (i * j) % 2 + (i * j) % 3 == 0;
      case QRMaskPattern.PATTERN110:
        return ((i * j) % 2 + (i * j) % 3) % 2 == 0;
      case QRMaskPattern.PATTERN111:
        return ((i * j) % 3 + (i + j) % 2) % 2 == 0;

      default:
        throw new Error("bad maskPattern:" + maskPattern);
    }
  },
  /*
   获取RS的纠错多项式
   */
  getErrorCorrectPolynomial: function(errorCorrectLength) {

    var a = new QRPolynomial([1], 0);

    for (var i = 0; i < errorCorrectLength; i++) {
      a = a.multiply(new QRPolynomial([1, QRMath.gexp(i)], 0));
    }

    return a;
  },
  /*
   获取评价
   */
  getLostPoint: function(qrCode) {

    var moduleCount = qrCode.getModuleCount(),
      lostPoint = 0,
      darkCount = 0;

    for (var row = 0; row < moduleCount; row++) {

      var sameCount = 0;
      var head = qrCode.modules[row][0];

      for (var col = 0; col < moduleCount; col++) {

        var current = qrCode.modules[row][col];

        //level 3 评价
        if (col < moduleCount - 6) {
          if (current && !qrCode.modules[row][col + 1] && qrCode.modules[row][col + 2] && qrCode.modules[row][col + 3] && qrCode.modules[row][col + 4] && !qrCode.modules[row][col + 5] && qrCode.modules[row][col + 6]) {
            if (col < moduleCount - 10) {
              if (qrCode.modules[row][col + 7] && qrCode.modules[row][col + 8] && qrCode.modules[row][col + 9] && qrCode.modules[row][col + 10]) {
                lostPoint += 40;
              }
            } else if (col > 3) {
              if (qrCode.modules[row][col - 1] && qrCode.modules[row][col - 2] && qrCode.modules[row][col - 3] && qrCode.modules[row][col - 4]) {
                lostPoint += 40;
              }
            }

          }
        }

        //level 2 评价
        if ((row < moduleCount - 1) && (col < moduleCount - 1)) {
          var count = 0;
          if (current) count++;
          if (qrCode.modules[row + 1][col]) count++;
          if (qrCode.modules[row][col + 1]) count++;
          if (qrCode.modules[row + 1][col + 1]) count++;
          if (count == 0 || count == 4) {
            lostPoint += 3;
          }
        }

        //level 1 评价
        if (head ^ current) {
          sameCount++;
        } else {
          head = current;
          if (sameCount >= 5) {
            lostPoint += (3 + sameCount - 5);
          }
          sameCount = 1;
        }

        //level 4 评价
        if (current) {
          darkCount++;
        }

      }
    }

    for (var col = 0; col < moduleCount; col++) {

      var sameCount = 0;
      var head = qrCode.modules[0][col];

      for (var row = 0; row < moduleCount; row++) {

        var current = qrCode.modules[row][col];

        //level 3 评价
        if (row < moduleCount - 6) {
          if (current && !qrCode.modules[row + 1][col] && qrCode.modules[row + 2][col] && qrCode.modules[row + 3][col] && qrCode.modules[row + 4][col] && !qrCode.modules[row + 5][col] && qrCode.modules[row + 6][col]) {
            if (row < moduleCount - 10) {
              if (qrCode.modules[row + 7][col] && qrCode.modules[row + 8][col] && qrCode.modules[row + 9][col] && qrCode.modules[row + 10][col]) {
                lostPoint += 40;
              }
            } else if (row > 3) {
              if (qrCode.modules[row - 1][col] && qrCode.modules[row - 2][col] && qrCode.modules[row - 3][col] && qrCode.modules[row - 4][col]) {
                lostPoint += 40;
              }
            }
          }
        }

        //level 1 评价
        if (head ^ current) {
          sameCount++;
        } else {
          head = current;
          if (sameCount >= 5) {
            lostPoint += (3 + sameCount - 5);
          }
          sameCount = 1;
        }

      }
    }

    // LEVEL4

    var ratio = Math.abs(100 * darkCount / moduleCount / moduleCount - 50) / 5;
    lostPoint += ratio * 10;

    return lostPoint;
  }

};


//---------------------------------------------------------------------
// QRMath使用的数学工具
//---------------------------------------------------------------------

var QRMath = {
  /*
   将n转化为a^m
   */
  glog: function(n) {

    if (n < 1) {
      throw new Error("glog(" + n + ")");
    }

    return QRMath.LOG_TABLE[n];
  },
  /*
   将a^m转化为n
   */
  gexp: function(n) {

    while (n < 0) {
      n += 255;
    }

    while (n >= 256) {
      n -= 255;
    }

    return QRMath.EXP_TABLE[n];
  },

  EXP_TABLE: new Array(256),

  LOG_TABLE: new Array(256)

};

for (var i = 0; i < 8; i++) {
  QRMath.EXP_TABLE[i] = 1 << i;
}
for (var i = 8; i < 256; i++) {
  QRMath.EXP_TABLE[i] = QRMath.EXP_TABLE[i - 4] ^ QRMath.EXP_TABLE[i - 5] ^ QRMath.EXP_TABLE[i - 6] ^ QRMath.EXP_TABLE[i - 8];
}
for (var i = 0; i < 255; i++) {
  QRMath.LOG_TABLE[QRMath.EXP_TABLE[i]] = i;
}

//---------------------------------------------------------------------
// QRPolynomial 多项式
//---------------------------------------------------------------------
/**
 * 多项式类
 * @param {Array} num   系数
 * @param {num} shift a^shift
 */
function QRPolynomial(num, shift) {

  if (num.length == undefined) {
    throw new Error(num.length + "/" + shift);
  }

  var offset = 0;

  while (offset < num.length && num[offset] == 0) {
    offset++;
  }

  this.num = new Array(num.length - offset + shift);
  for (var i = 0; i < num.length - offset; i++) {
    this.num[i] = num[i + offset];
  }
}

QRPolynomial.prototype = {

  get: function(index) {
    return this.num[index];
  },

  getLength: function() {
    return this.num.length;
  },
  /**
   * 多项式乘法
   * @param  {QRPolynomial} e 被乘多项式
   * @return {[type]}   [description]
   */
  multiply: function(e) {

    var num = new Array(this.getLength() + e.getLength() - 1);

    for (var i = 0; i < this.getLength(); i++) {
      for (var j = 0; j < e.getLength(); j++) {
        num[i + j] ^= QRMath.gexp(QRMath.glog(this.get(i)) + QRMath.glog(e.get(j)));
      }
    }

    return new QRPolynomial(num, 0);
  },
  /**
   * 多项式模运算
   * @param  {QRPolynomial} e 模多项式
   * @return {}
   */
  mod: function(e) {
    var tl = this.getLength(),
      el = e.getLength();
    if (tl - el < 0) {
      return this;
    }
    var num = new Array(tl);
    for (var i = 0; i < tl; i++) {
      num[i] = this.get(i);
    }
    while (num.length >= el) {
      var ratio = QRMath.glog(num[0]) - QRMath.glog(e.get(0));

      for (var i = 0; i < e.getLength(); i++) {
        num[i] ^= QRMath.gexp(QRMath.glog(e.get(i)) + ratio);
      }
      while (num[0] == 0) {
        num.shift();
      }
    }
    return new QRPolynomial(num, 0);
  }
};

//---------------------------------------------------------------------
// RS_BLOCK_TABLE
//---------------------------------------------------------------------
/*
 二维码各个版本信息[块数, 每块中的数据块数, 每块中的信息块数]
 */
RS_BLOCK_TABLE = [

  // L
  // M
  // Q
  // H

  // 1
  [1, 26, 19],
  [1, 26, 16],
  [1, 26, 13],
  [1, 26, 9],

  // 2
  [1, 44, 34],
  [1, 44, 28],
  [1, 44, 22],
  [1, 44, 16],

  // 3
  [1, 70, 55],
  [1, 70, 44],
  [2, 35, 17],
  [2, 35, 13],

  // 4
  [1, 100, 80],
  [2, 50, 32],
  [2, 50, 24],
  [4, 25, 9],

  // 5
  [1, 134, 108],
  [2, 67, 43],
  [2, 33, 15, 2, 34, 16],
  [2, 33, 11, 2, 34, 12],

  // 6
  [2, 86, 68],
  [4, 43, 27],
  [4, 43, 19],
  [4, 43, 15],

  // 7
  [2, 98, 78],
  [4, 49, 31],
  [2, 32, 14, 4, 33, 15],
  [4, 39, 13, 1, 40, 14],

  // 8
  [2, 121, 97],
  [2, 60, 38, 2, 61, 39],
  [4, 40, 18, 2, 41, 19],
  [4, 40, 14, 2, 41, 15],

  // 9
  [2, 146, 116],
  [3, 58, 36, 2, 59, 37],
  [4, 36, 16, 4, 37, 17],
  [4, 36, 12, 4, 37, 13],

  // 10
  [2, 86, 68, 2, 87, 69],
  [4, 69, 43, 1, 70, 44],
  [6, 43, 19, 2, 44, 20],
  [6, 43, 15, 2, 44, 16],

  // 11
  [4, 101, 81],
  [1, 80, 50, 4, 81, 51],
  [4, 50, 22, 4, 51, 23],
  [3, 36, 12, 8, 37, 13],

  // 12
  [2, 116, 92, 2, 117, 93],
  [6, 58, 36, 2, 59, 37],
  [4, 46, 20, 6, 47, 21],
  [7, 42, 14, 4, 43, 15],

  // 13
  [4, 133, 107],
  [8, 59, 37, 1, 60, 38],
  [8, 44, 20, 4, 45, 21],
  [12, 33, 11, 4, 34, 12],

  // 14
  [3, 145, 115, 1, 146, 116],
  [4, 64, 40, 5, 65, 41],
  [11, 36, 16, 5, 37, 17],
  [11, 36, 12, 5, 37, 13],

  // 15
  [5, 109, 87, 1, 110, 88],
  [5, 65, 41, 5, 66, 42],
  [5, 54, 24, 7, 55, 25],
  [11, 36, 12],

  // 16
  [5, 122, 98, 1, 123, 99],
  [7, 73, 45, 3, 74, 46],
  [15, 43, 19, 2, 44, 20],
  [3, 45, 15, 13, 46, 16],

  // 17
  [1, 135, 107, 5, 136, 108],
  [10, 74, 46, 1, 75, 47],
  [1, 50, 22, 15, 51, 23],
  [2, 42, 14, 17, 43, 15],

  // 18
  [5, 150, 120, 1, 151, 121],
  [9, 69, 43, 4, 70, 44],
  [17, 50, 22, 1, 51, 23],
  [2, 42, 14, 19, 43, 15],

  // 19
  [3, 141, 113, 4, 142, 114],
  [3, 70, 44, 11, 71, 45],
  [17, 47, 21, 4, 48, 22],
  [9, 39, 13, 16, 40, 14],

  // 20
  [3, 135, 107, 5, 136, 108],
  [3, 67, 41, 13, 68, 42],
  [15, 54, 24, 5, 55, 25],
  [15, 43, 15, 10, 44, 16],

  // 21
  [4, 144, 116, 4, 145, 117],
  [17, 68, 42],
  [17, 50, 22, 6, 51, 23],
  [19, 46, 16, 6, 47, 17],

  // 22
  [2, 139, 111, 7, 140, 112],
  [17, 74, 46],
  [7, 54, 24, 16, 55, 25],
  [34, 37, 13],

  // 23
  [4, 151, 121, 5, 152, 122],
  [4, 75, 47, 14, 76, 48],
  [11, 54, 24, 14, 55, 25],
  [16, 45, 15, 14, 46, 16],

  // 24
  [6, 147, 117, 4, 148, 118],
  [6, 73, 45, 14, 74, 46],
  [11, 54, 24, 16, 55, 25],
  [30, 46, 16, 2, 47, 17],

  // 25
  [8, 132, 106, 4, 133, 107],
  [8, 75, 47, 13, 76, 48],
  [7, 54, 24, 22, 55, 25],
  [22, 45, 15, 13, 46, 16],

  // 26
  [10, 142, 114, 2, 143, 115],
  [19, 74, 46, 4, 75, 47],
  [28, 50, 22, 6, 51, 23],
  [33, 46, 16, 4, 47, 17],

  // 27
  [8, 152, 122, 4, 153, 123],
  [22, 73, 45, 3, 74, 46],
  [8, 53, 23, 26, 54, 24],
  [12, 45, 15, 28, 46, 16],

  // 28
  [3, 147, 117, 10, 148, 118],
  [3, 73, 45, 23, 74, 46],
  [4, 54, 24, 31, 55, 25],
  [11, 45, 15, 31, 46, 16],

  // 29
  [7, 146, 116, 7, 147, 117],
  [21, 73, 45, 7, 74, 46],
  [1, 53, 23, 37, 54, 24],
  [19, 45, 15, 26, 46, 16],

  // 30
  [5, 145, 115, 10, 146, 116],
  [19, 75, 47, 10, 76, 48],
  [15, 54, 24, 25, 55, 25],
  [23, 45, 15, 25, 46, 16],

  // 31
  [13, 145, 115, 3, 146, 116],
  [2, 74, 46, 29, 75, 47],
  [42, 54, 24, 1, 55, 25],
  [23, 45, 15, 28, 46, 16],

  // 32
  [17, 145, 115],
  [10, 74, 46, 23, 75, 47],
  [10, 54, 24, 35, 55, 25],
  [19, 45, 15, 35, 46, 16],

  // 33
  [17, 145, 115, 1, 146, 116],
  [14, 74, 46, 21, 75, 47],
  [29, 54, 24, 19, 55, 25],
  [11, 45, 15, 46, 46, 16],

  // 34
  [13, 145, 115, 6, 146, 116],
  [14, 74, 46, 23, 75, 47],
  [44, 54, 24, 7, 55, 25],
  [59, 46, 16, 1, 47, 17],

  // 35
  [12, 151, 121, 7, 152, 122],
  [12, 75, 47, 26, 76, 48],
  [39, 54, 24, 14, 55, 25],
  [22, 45, 15, 41, 46, 16],

  // 36
  [6, 151, 121, 14, 152, 122],
  [6, 75, 47, 34, 76, 48],
  [46, 54, 24, 10, 55, 25],
  [2, 45, 15, 64, 46, 16],

  // 37
  [17, 152, 122, 4, 153, 123],
  [29, 74, 46, 14, 75, 47],
  [49, 54, 24, 10, 55, 25],
  [24, 45, 15, 46, 46, 16],

  // 38
  [4, 152, 122, 18, 153, 123],
  [13, 74, 46, 32, 75, 47],
  [48, 54, 24, 14, 55, 25],
  [42, 45, 15, 32, 46, 16],

  // 39
  [20, 147, 117, 4, 148, 118],
  [40, 75, 47, 7, 76, 48],
  [43, 54, 24, 22, 55, 25],
  [10, 45, 15, 67, 46, 16],

  // 40
  [19, 148, 118, 6, 149, 119],
  [18, 75, 47, 31, 76, 48],
  [34, 54, 24, 34, 55, 25],
  [20, 45, 15, 61, 46, 16]
];

/**
 * 根据数据获取对应版本
 * @return {[type]} [description]
 */
QRCodeAlg.prototype.getRightType = function() {
  for (var typeNumber = 1; typeNumber < 41; typeNumber++) {
    var rsBlock = RS_BLOCK_TABLE[(typeNumber - 1) * 4 + this.errorCorrectLevel];
    if (rsBlock == undefined) {
      throw new Error('bad rs block @ typeNumber:' + typeNumber + '/errorCorrectLevel:' + this.errorCorrectLevel);
    }
    var length = rsBlock.length / 3;
    var totalDataCount = 0;
    for (var i = 0; i < length; i++) {
      var count = rsBlock[i * 3 + 0];
      var dataCount = rsBlock[i * 3 + 2];
      totalDataCount += dataCount * count;
    }

    var lengthBytes = typeNumber > 9 ? 2 : 1;
    if (this.utf8bytes.length + lengthBytes < totalDataCount || typeNumber == 40) {
      this.typeNumber = typeNumber;
      this.rsBlock = rsBlock;
      this.totalDataCount = totalDataCount;
      break;
    }
  }
};

// ---------------------------------------------------------------------
// QRBitBuffer
// ---------------------------------------------------------------------

function QRBitBuffer() {
  this.buffer = new Array();
  this.length = 0;
}

QRBitBuffer.prototype = {
  get: function(index) {
    var bufIndex = Math.floor(index / 8);
    return ((this.buffer[bufIndex] >>> (7 - index % 8)) & 1);
  },

  put: function(num, length) {
    for (var i = 0; i < length; i++) {
      this.putBit(((num >>> (length - i - 1)) & 1));
    }
  },

  putBit: function(bit) {

    var bufIndex = Math.floor(this.length / 8);
    if (this.buffer.length <= bufIndex) {
      this.buffer.push(0);
    }

    if (bit) {
      this.buffer[bufIndex] |= (0x80 >>> (this.length % 8));
    }

    this.length++;
  }
};

$.fn.qrcode = function QRCodePlugin(option) {
  return this.each(function() {
    $(this).append(new QRCode(option));
  });
};

module.exports = UI.qrcode = QRCode;
