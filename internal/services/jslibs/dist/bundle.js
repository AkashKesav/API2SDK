var API2SDK_JS_LIBS = (() => {
  var __create = Object.create;
  var __defProp = Object.defineProperty;
  var __defProps = Object.defineProperties;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropDescs = Object.getOwnPropertyDescriptors;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __getOwnPropSymbols = Object.getOwnPropertySymbols;
  var __getProtoOf = Object.getPrototypeOf;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __propIsEnum = Object.prototype.propertyIsEnumerable;
  var __defNormalProp = (obj, key, value) => key in obj ? __defProp(obj, key, { enumerable: true, configurable: true, writable: true, value }) : obj[key] = value;
  var __spreadValues = (a, b) => {
    for (var prop in b || (b = {}))
      if (__hasOwnProp.call(b, prop))
        __defNormalProp(a, prop, b[prop]);
    if (__getOwnPropSymbols)
      for (var prop of __getOwnPropSymbols(b)) {
        if (__propIsEnum.call(b, prop))
          __defNormalProp(a, prop, b[prop]);
      }
    return a;
  };
  var __spreadProps = (a, b) => __defProps(a, __getOwnPropDescs(b));
  var __objRest = (source, exclude) => {
    var target = {};
    for (var prop in source)
      if (__hasOwnProp.call(source, prop) && exclude.indexOf(prop) < 0)
        target[prop] = source[prop];
    if (source != null && __getOwnPropSymbols)
      for (var prop of __getOwnPropSymbols(source)) {
        if (exclude.indexOf(prop) < 0 && __propIsEnum.call(source, prop))
          target[prop] = source[prop];
      }
    return target;
  };
  var __commonJS = (cb, mod) => function __require() {
    return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
    // If the importer is in node compatibility mode or this is not an ESM
    // file that has been converted to a CommonJS file using a Babel-
    // compatible transform (i.e. "__esModule" has not been set), then set
    // "default" to the CommonJS "module.exports" for node compatibility.
    isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
    mod
  ));

  // fs.js
  var require_fs = __commonJS({
    "fs.js"(exports, module) {
      console.log("[API2SDK FS Shim] fs module loaded");
      module.exports = {
        promises: {
          writeFile: async (filePath, data) => {
            console.warn(`[API2SDK FS Shim] fs.promises.writeFile called for ${filePath}. This should ideally not happen.`);
          },
          readFile: async (filePath) => {
            const pathStr = String(filePath);
            if (typeof filePath !== "string") {
              try {
                const jsonContent = JSON.stringify(filePath);
                console.warn(`[API2SDK FS Shim] fs.promises.readFile called with an object. Serialized and returning as JSON string. Content (first 100): ${jsonContent.substring(0, 100)}`);
                return Promise.resolve(jsonContent);
              } catch (e) {
                console.error(`[API2SDK FS Shim] fs.promises.readFile called with an object that could not be serialized: ${pathStr}`, e);
                return Promise.reject(new Error(`[API2SDK FS Shim] Could not serialize object passed to readFile: ${e.message}`));
              }
            }
            if (filePath === "" || pathStr.startsWith("{") || pathStr.startsWith("[")) {
              console.warn(`[API2SDK FS Shim] fs.promises.readFile called with string content that looks like JSON: '${pathStr.substring(0, 100)}...'. Returning content directly.`);
              return Promise.resolve(pathStr);
            }
            console.error(`[API2SDK FS Shim] fs.promises.readFile: True file system access attempted for path '${pathStr}' and is not available in this environment.`);
            return Promise.reject(new Error(`[API2SDK FS Shim] File system access not available for path: ${pathStr}`));
          }
          // Add other fs functions if they are reported as missing by the library
          // For example, existsSync, statSync, etc.
          // stat: async (path) => {
          //   console.warn(`[API2SDK FS Shim] fs.promises.stat called for ${String(path)}. Returning dummy stats.`);
          //   return Promise.resolve({
          //     isFile: () => true,
          //     isDirectory: () => false,
          //     isBlockDevice: () => false,
          //     isCharacterDevice: () => false,
          //     isSymbolicLink: () => false,
          //     isFIFO: () => false,
          //     isSocket: () => false,
          //     size: 0,
          //   });
          // },
          // access: async (path, mode) => {
          //   console.warn(`[API2SDK FS Shim] fs.promises.access called for ${String(path)}. Assuming accessible.`);
          //   return Promise.resolve();
          // }
        }
        // Synchronous stubs if needed, though esbuild browser platform should handle some of this.
        // readFileSync: (path, options) => {
        //   console.error(`[API2SDK FS Shim] fs.readFileSync: File system access not available for path: ${String(path)}`);
        //   throw new Error(`[API2SDK FS Shim] File system access not available for path: ${String(path)} (readFileSync)`);
        // },
        // writeFileSync: (path, data, options) => {
        //   console.warn(`[API2SDK FS Shim] fs.writeFileSync called for path '${String(path)}'. This operation is not supported and will be a no-op.`);
        // },
      };
      console.log("[API2SDK FS Shim] fs module (readFile/writeFile promises) shim loaded with improved logic.");
    }
  });

  // node_modules/js-yaml/lib/common.js
  var require_common = __commonJS({
    "node_modules/js-yaml/lib/common.js"(exports, module) {
      "use strict";
      function isNothing(subject) {
        return typeof subject === "undefined" || subject === null;
      }
      function isObject(subject) {
        return typeof subject === "object" && subject !== null;
      }
      function toArray(sequence) {
        if (Array.isArray(sequence))
          return sequence;
        else if (isNothing(sequence))
          return [];
        return [sequence];
      }
      function extend(target, source) {
        var index, length, key, sourceKeys;
        if (source) {
          sourceKeys = Object.keys(source);
          for (index = 0, length = sourceKeys.length; index < length; index += 1) {
            key = sourceKeys[index];
            target[key] = source[key];
          }
        }
        return target;
      }
      function repeat(string, count) {
        var result = "", cycle;
        for (cycle = 0; cycle < count; cycle += 1) {
          result += string;
        }
        return result;
      }
      function isNegativeZero(number) {
        return number === 0 && Number.NEGATIVE_INFINITY === 1 / number;
      }
      module.exports.isNothing = isNothing;
      module.exports.isObject = isObject;
      module.exports.toArray = toArray;
      module.exports.repeat = repeat;
      module.exports.isNegativeZero = isNegativeZero;
      module.exports.extend = extend;
    }
  });

  // node_modules/js-yaml/lib/exception.js
  var require_exception = __commonJS({
    "node_modules/js-yaml/lib/exception.js"(exports, module) {
      "use strict";
      function formatError(exception, compact) {
        var where = "", message = exception.reason || "(unknown reason)";
        if (!exception.mark)
          return message;
        if (exception.mark.name) {
          where += 'in "' + exception.mark.name + '" ';
        }
        where += "(" + (exception.mark.line + 1) + ":" + (exception.mark.column + 1) + ")";
        if (!compact && exception.mark.snippet) {
          where += "\n\n" + exception.mark.snippet;
        }
        return message + " " + where;
      }
      function YAMLException(reason, mark) {
        Error.call(this);
        this.name = "YAMLException";
        this.reason = reason;
        this.mark = mark;
        this.message = formatError(this, false);
        if (Error.captureStackTrace) {
          Error.captureStackTrace(this, this.constructor);
        } else {
          this.stack = new Error().stack || "";
        }
      }
      YAMLException.prototype = Object.create(Error.prototype);
      YAMLException.prototype.constructor = YAMLException;
      YAMLException.prototype.toString = function toString(compact) {
        return this.name + ": " + formatError(this, compact);
      };
      module.exports = YAMLException;
    }
  });

  // node_modules/js-yaml/lib/snippet.js
  var require_snippet = __commonJS({
    "node_modules/js-yaml/lib/snippet.js"(exports, module) {
      "use strict";
      var common = require_common();
      function getLine(buffer, lineStart, lineEnd, position, maxLineLength) {
        var head = "";
        var tail = "";
        var maxHalfLength = Math.floor(maxLineLength / 2) - 1;
        if (position - lineStart > maxHalfLength) {
          head = " ... ";
          lineStart = position - maxHalfLength + head.length;
        }
        if (lineEnd - position > maxHalfLength) {
          tail = " ...";
          lineEnd = position + maxHalfLength - tail.length;
        }
        return {
          str: head + buffer.slice(lineStart, lineEnd).replace(/\t/g, "\u2192") + tail,
          pos: position - lineStart + head.length
          // relative position
        };
      }
      function padStart(string, max) {
        return common.repeat(" ", max - string.length) + string;
      }
      function makeSnippet(mark, options) {
        options = Object.create(options || null);
        if (!mark.buffer)
          return null;
        if (!options.maxLength)
          options.maxLength = 79;
        if (typeof options.indent !== "number")
          options.indent = 1;
        if (typeof options.linesBefore !== "number")
          options.linesBefore = 3;
        if (typeof options.linesAfter !== "number")
          options.linesAfter = 2;
        var re = /\r?\n|\r|\0/g;
        var lineStarts = [0];
        var lineEnds = [];
        var match;
        var foundLineNo = -1;
        while (match = re.exec(mark.buffer)) {
          lineEnds.push(match.index);
          lineStarts.push(match.index + match[0].length);
          if (mark.position <= match.index && foundLineNo < 0) {
            foundLineNo = lineStarts.length - 2;
          }
        }
        if (foundLineNo < 0)
          foundLineNo = lineStarts.length - 1;
        var result = "", i, line;
        var lineNoLength = Math.min(mark.line + options.linesAfter, lineEnds.length).toString().length;
        var maxLineLength = options.maxLength - (options.indent + lineNoLength + 3);
        for (i = 1; i <= options.linesBefore; i++) {
          if (foundLineNo - i < 0)
            break;
          line = getLine(
            mark.buffer,
            lineStarts[foundLineNo - i],
            lineEnds[foundLineNo - i],
            mark.position - (lineStarts[foundLineNo] - lineStarts[foundLineNo - i]),
            maxLineLength
          );
          result = common.repeat(" ", options.indent) + padStart((mark.line - i + 1).toString(), lineNoLength) + " | " + line.str + "\n" + result;
        }
        line = getLine(mark.buffer, lineStarts[foundLineNo], lineEnds[foundLineNo], mark.position, maxLineLength);
        result += common.repeat(" ", options.indent) + padStart((mark.line + 1).toString(), lineNoLength) + " | " + line.str + "\n";
        result += common.repeat("-", options.indent + lineNoLength + 3 + line.pos) + "^\n";
        for (i = 1; i <= options.linesAfter; i++) {
          if (foundLineNo + i >= lineEnds.length)
            break;
          line = getLine(
            mark.buffer,
            lineStarts[foundLineNo + i],
            lineEnds[foundLineNo + i],
            mark.position - (lineStarts[foundLineNo] - lineStarts[foundLineNo + i]),
            maxLineLength
          );
          result += common.repeat(" ", options.indent) + padStart((mark.line + i + 1).toString(), lineNoLength) + " | " + line.str + "\n";
        }
        return result.replace(/\n$/, "");
      }
      module.exports = makeSnippet;
    }
  });

  // node_modules/js-yaml/lib/type.js
  var require_type = __commonJS({
    "node_modules/js-yaml/lib/type.js"(exports, module) {
      "use strict";
      var YAMLException = require_exception();
      var TYPE_CONSTRUCTOR_OPTIONS = [
        "kind",
        "multi",
        "resolve",
        "construct",
        "instanceOf",
        "predicate",
        "represent",
        "representName",
        "defaultStyle",
        "styleAliases"
      ];
      var YAML_NODE_KINDS = [
        "scalar",
        "sequence",
        "mapping"
      ];
      function compileStyleAliases(map) {
        var result = {};
        if (map !== null) {
          Object.keys(map).forEach(function(style) {
            map[style].forEach(function(alias) {
              result[String(alias)] = style;
            });
          });
        }
        return result;
      }
      function Type(tag, options) {
        options = options || {};
        Object.keys(options).forEach(function(name) {
          if (TYPE_CONSTRUCTOR_OPTIONS.indexOf(name) === -1) {
            throw new YAMLException('Unknown option "' + name + '" is met in definition of "' + tag + '" YAML type.');
          }
        });
        this.options = options;
        this.tag = tag;
        this.kind = options["kind"] || null;
        this.resolve = options["resolve"] || function() {
          return true;
        };
        this.construct = options["construct"] || function(data) {
          return data;
        };
        this.instanceOf = options["instanceOf"] || null;
        this.predicate = options["predicate"] || null;
        this.represent = options["represent"] || null;
        this.representName = options["representName"] || null;
        this.defaultStyle = options["defaultStyle"] || null;
        this.multi = options["multi"] || false;
        this.styleAliases = compileStyleAliases(options["styleAliases"] || null);
        if (YAML_NODE_KINDS.indexOf(this.kind) === -1) {
          throw new YAMLException('Unknown kind "' + this.kind + '" is specified for "' + tag + '" YAML type.');
        }
      }
      module.exports = Type;
    }
  });

  // node_modules/js-yaml/lib/schema.js
  var require_schema = __commonJS({
    "node_modules/js-yaml/lib/schema.js"(exports, module) {
      "use strict";
      var YAMLException = require_exception();
      var Type = require_type();
      function compileList(schema, name) {
        var result = [];
        schema[name].forEach(function(currentType) {
          var newIndex = result.length;
          result.forEach(function(previousType, previousIndex) {
            if (previousType.tag === currentType.tag && previousType.kind === currentType.kind && previousType.multi === currentType.multi) {
              newIndex = previousIndex;
            }
          });
          result[newIndex] = currentType;
        });
        return result;
      }
      function compileMap() {
        var result = {
          scalar: {},
          sequence: {},
          mapping: {},
          fallback: {},
          multi: {
            scalar: [],
            sequence: [],
            mapping: [],
            fallback: []
          }
        }, index, length;
        function collectType(type) {
          if (type.multi) {
            result.multi[type.kind].push(type);
            result.multi["fallback"].push(type);
          } else {
            result[type.kind][type.tag] = result["fallback"][type.tag] = type;
          }
        }
        for (index = 0, length = arguments.length; index < length; index += 1) {
          arguments[index].forEach(collectType);
        }
        return result;
      }
      function Schema(definition) {
        return this.extend(definition);
      }
      Schema.prototype.extend = function extend(definition) {
        var implicit = [];
        var explicit = [];
        if (definition instanceof Type) {
          explicit.push(definition);
        } else if (Array.isArray(definition)) {
          explicit = explicit.concat(definition);
        } else if (definition && (Array.isArray(definition.implicit) || Array.isArray(definition.explicit))) {
          if (definition.implicit)
            implicit = implicit.concat(definition.implicit);
          if (definition.explicit)
            explicit = explicit.concat(definition.explicit);
        } else {
          throw new YAMLException("Schema.extend argument should be a Type, [ Type ], or a schema definition ({ implicit: [...], explicit: [...] })");
        }
        implicit.forEach(function(type) {
          if (!(type instanceof Type)) {
            throw new YAMLException("Specified list of YAML types (or a single Type object) contains a non-Type object.");
          }
          if (type.loadKind && type.loadKind !== "scalar") {
            throw new YAMLException("There is a non-scalar type in the implicit list of a schema. Implicit resolving of such types is not supported.");
          }
          if (type.multi) {
            throw new YAMLException("There is a multi type in the implicit list of a schema. Multi tags can only be listed as explicit.");
          }
        });
        explicit.forEach(function(type) {
          if (!(type instanceof Type)) {
            throw new YAMLException("Specified list of YAML types (or a single Type object) contains a non-Type object.");
          }
        });
        var result = Object.create(Schema.prototype);
        result.implicit = (this.implicit || []).concat(implicit);
        result.explicit = (this.explicit || []).concat(explicit);
        result.compiledImplicit = compileList(result, "implicit");
        result.compiledExplicit = compileList(result, "explicit");
        result.compiledTypeMap = compileMap(result.compiledImplicit, result.compiledExplicit);
        return result;
      };
      module.exports = Schema;
    }
  });

  // node_modules/js-yaml/lib/type/str.js
  var require_str = __commonJS({
    "node_modules/js-yaml/lib/type/str.js"(exports, module) {
      "use strict";
      var Type = require_type();
      module.exports = new Type("tag:yaml.org,2002:str", {
        kind: "scalar",
        construct: function(data) {
          return data !== null ? data : "";
        }
      });
    }
  });

  // node_modules/js-yaml/lib/type/seq.js
  var require_seq = __commonJS({
    "node_modules/js-yaml/lib/type/seq.js"(exports, module) {
      "use strict";
      var Type = require_type();
      module.exports = new Type("tag:yaml.org,2002:seq", {
        kind: "sequence",
        construct: function(data) {
          return data !== null ? data : [];
        }
      });
    }
  });

  // node_modules/js-yaml/lib/type/map.js
  var require_map = __commonJS({
    "node_modules/js-yaml/lib/type/map.js"(exports, module) {
      "use strict";
      var Type = require_type();
      module.exports = new Type("tag:yaml.org,2002:map", {
        kind: "mapping",
        construct: function(data) {
          return data !== null ? data : {};
        }
      });
    }
  });

  // node_modules/js-yaml/lib/schema/failsafe.js
  var require_failsafe = __commonJS({
    "node_modules/js-yaml/lib/schema/failsafe.js"(exports, module) {
      "use strict";
      var Schema = require_schema();
      module.exports = new Schema({
        explicit: [
          require_str(),
          require_seq(),
          require_map()
        ]
      });
    }
  });

  // node_modules/js-yaml/lib/type/null.js
  var require_null = __commonJS({
    "node_modules/js-yaml/lib/type/null.js"(exports, module) {
      "use strict";
      var Type = require_type();
      function resolveYamlNull(data) {
        if (data === null)
          return true;
        var max = data.length;
        return max === 1 && data === "~" || max === 4 && (data === "null" || data === "Null" || data === "NULL");
      }
      function constructYamlNull() {
        return null;
      }
      function isNull(object) {
        return object === null;
      }
      module.exports = new Type("tag:yaml.org,2002:null", {
        kind: "scalar",
        resolve: resolveYamlNull,
        construct: constructYamlNull,
        predicate: isNull,
        represent: {
          canonical: function() {
            return "~";
          },
          lowercase: function() {
            return "null";
          },
          uppercase: function() {
            return "NULL";
          },
          camelcase: function() {
            return "Null";
          },
          empty: function() {
            return "";
          }
        },
        defaultStyle: "lowercase"
      });
    }
  });

  // node_modules/js-yaml/lib/type/bool.js
  var require_bool = __commonJS({
    "node_modules/js-yaml/lib/type/bool.js"(exports, module) {
      "use strict";
      var Type = require_type();
      function resolveYamlBoolean(data) {
        if (data === null)
          return false;
        var max = data.length;
        return max === 4 && (data === "true" || data === "True" || data === "TRUE") || max === 5 && (data === "false" || data === "False" || data === "FALSE");
      }
      function constructYamlBoolean(data) {
        return data === "true" || data === "True" || data === "TRUE";
      }
      function isBoolean(object) {
        return Object.prototype.toString.call(object) === "[object Boolean]";
      }
      module.exports = new Type("tag:yaml.org,2002:bool", {
        kind: "scalar",
        resolve: resolveYamlBoolean,
        construct: constructYamlBoolean,
        predicate: isBoolean,
        represent: {
          lowercase: function(object) {
            return object ? "true" : "false";
          },
          uppercase: function(object) {
            return object ? "TRUE" : "FALSE";
          },
          camelcase: function(object) {
            return object ? "True" : "False";
          }
        },
        defaultStyle: "lowercase"
      });
    }
  });

  // node_modules/js-yaml/lib/type/int.js
  var require_int = __commonJS({
    "node_modules/js-yaml/lib/type/int.js"(exports, module) {
      "use strict";
      var common = require_common();
      var Type = require_type();
      function isHexCode(c) {
        return 48 <= c && c <= 57 || 65 <= c && c <= 70 || 97 <= c && c <= 102;
      }
      function isOctCode(c) {
        return 48 <= c && c <= 55;
      }
      function isDecCode(c) {
        return 48 <= c && c <= 57;
      }
      function resolveYamlInteger(data) {
        if (data === null)
          return false;
        var max = data.length, index = 0, hasDigits = false, ch;
        if (!max)
          return false;
        ch = data[index];
        if (ch === "-" || ch === "+") {
          ch = data[++index];
        }
        if (ch === "0") {
          if (index + 1 === max)
            return true;
          ch = data[++index];
          if (ch === "b") {
            index++;
            for (; index < max; index++) {
              ch = data[index];
              if (ch === "_")
                continue;
              if (ch !== "0" && ch !== "1")
                return false;
              hasDigits = true;
            }
            return hasDigits && ch !== "_";
          }
          if (ch === "x") {
            index++;
            for (; index < max; index++) {
              ch = data[index];
              if (ch === "_")
                continue;
              if (!isHexCode(data.charCodeAt(index)))
                return false;
              hasDigits = true;
            }
            return hasDigits && ch !== "_";
          }
          if (ch === "o") {
            index++;
            for (; index < max; index++) {
              ch = data[index];
              if (ch === "_")
                continue;
              if (!isOctCode(data.charCodeAt(index)))
                return false;
              hasDigits = true;
            }
            return hasDigits && ch !== "_";
          }
        }
        if (ch === "_")
          return false;
        for (; index < max; index++) {
          ch = data[index];
          if (ch === "_")
            continue;
          if (!isDecCode(data.charCodeAt(index))) {
            return false;
          }
          hasDigits = true;
        }
        if (!hasDigits || ch === "_")
          return false;
        return true;
      }
      function constructYamlInteger(data) {
        var value = data, sign = 1, ch;
        if (value.indexOf("_") !== -1) {
          value = value.replace(/_/g, "");
        }
        ch = value[0];
        if (ch === "-" || ch === "+") {
          if (ch === "-")
            sign = -1;
          value = value.slice(1);
          ch = value[0];
        }
        if (value === "0")
          return 0;
        if (ch === "0") {
          if (value[1] === "b")
            return sign * parseInt(value.slice(2), 2);
          if (value[1] === "x")
            return sign * parseInt(value.slice(2), 16);
          if (value[1] === "o")
            return sign * parseInt(value.slice(2), 8);
        }
        return sign * parseInt(value, 10);
      }
      function isInteger(object) {
        return Object.prototype.toString.call(object) === "[object Number]" && (object % 1 === 0 && !common.isNegativeZero(object));
      }
      module.exports = new Type("tag:yaml.org,2002:int", {
        kind: "scalar",
        resolve: resolveYamlInteger,
        construct: constructYamlInteger,
        predicate: isInteger,
        represent: {
          binary: function(obj) {
            return obj >= 0 ? "0b" + obj.toString(2) : "-0b" + obj.toString(2).slice(1);
          },
          octal: function(obj) {
            return obj >= 0 ? "0o" + obj.toString(8) : "-0o" + obj.toString(8).slice(1);
          },
          decimal: function(obj) {
            return obj.toString(10);
          },
          /* eslint-disable max-len */
          hexadecimal: function(obj) {
            return obj >= 0 ? "0x" + obj.toString(16).toUpperCase() : "-0x" + obj.toString(16).toUpperCase().slice(1);
          }
        },
        defaultStyle: "decimal",
        styleAliases: {
          binary: [2, "bin"],
          octal: [8, "oct"],
          decimal: [10, "dec"],
          hexadecimal: [16, "hex"]
        }
      });
    }
  });

  // node_modules/js-yaml/lib/type/float.js
  var require_float = __commonJS({
    "node_modules/js-yaml/lib/type/float.js"(exports, module) {
      "use strict";
      var common = require_common();
      var Type = require_type();
      var YAML_FLOAT_PATTERN = new RegExp(
        // 2.5e4, 2.5 and integers
        "^(?:[-+]?(?:[0-9][0-9_]*)(?:\\.[0-9_]*)?(?:[eE][-+]?[0-9]+)?|\\.[0-9_]+(?:[eE][-+]?[0-9]+)?|[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"
      );
      function resolveYamlFloat(data) {
        if (data === null)
          return false;
        if (!YAML_FLOAT_PATTERN.test(data) || // Quick hack to not allow integers end with `_`
        // Probably should update regexp & check speed
        data[data.length - 1] === "_") {
          return false;
        }
        return true;
      }
      function constructYamlFloat(data) {
        var value, sign;
        value = data.replace(/_/g, "").toLowerCase();
        sign = value[0] === "-" ? -1 : 1;
        if ("+-".indexOf(value[0]) >= 0) {
          value = value.slice(1);
        }
        if (value === ".inf") {
          return sign === 1 ? Number.POSITIVE_INFINITY : Number.NEGATIVE_INFINITY;
        } else if (value === ".nan") {
          return NaN;
        }
        return sign * parseFloat(value, 10);
      }
      var SCIENTIFIC_WITHOUT_DOT = /^[-+]?[0-9]+e/;
      function representYamlFloat(object, style) {
        var res;
        if (isNaN(object)) {
          switch (style) {
            case "lowercase":
              return ".nan";
            case "uppercase":
              return ".NAN";
            case "camelcase":
              return ".NaN";
          }
        } else if (Number.POSITIVE_INFINITY === object) {
          switch (style) {
            case "lowercase":
              return ".inf";
            case "uppercase":
              return ".INF";
            case "camelcase":
              return ".Inf";
          }
        } else if (Number.NEGATIVE_INFINITY === object) {
          switch (style) {
            case "lowercase":
              return "-.inf";
            case "uppercase":
              return "-.INF";
            case "camelcase":
              return "-.Inf";
          }
        } else if (common.isNegativeZero(object)) {
          return "-0.0";
        }
        res = object.toString(10);
        return SCIENTIFIC_WITHOUT_DOT.test(res) ? res.replace("e", ".e") : res;
      }
      function isFloat(object) {
        return Object.prototype.toString.call(object) === "[object Number]" && (object % 1 !== 0 || common.isNegativeZero(object));
      }
      module.exports = new Type("tag:yaml.org,2002:float", {
        kind: "scalar",
        resolve: resolveYamlFloat,
        construct: constructYamlFloat,
        predicate: isFloat,
        represent: representYamlFloat,
        defaultStyle: "lowercase"
      });
    }
  });

  // node_modules/js-yaml/lib/schema/json.js
  var require_json = __commonJS({
    "node_modules/js-yaml/lib/schema/json.js"(exports, module) {
      "use strict";
      module.exports = require_failsafe().extend({
        implicit: [
          require_null(),
          require_bool(),
          require_int(),
          require_float()
        ]
      });
    }
  });

  // node_modules/js-yaml/lib/schema/core.js
  var require_core = __commonJS({
    "node_modules/js-yaml/lib/schema/core.js"(exports, module) {
      "use strict";
      module.exports = require_json();
    }
  });

  // node_modules/js-yaml/lib/type/timestamp.js
  var require_timestamp = __commonJS({
    "node_modules/js-yaml/lib/type/timestamp.js"(exports, module) {
      "use strict";
      var Type = require_type();
      var YAML_DATE_REGEXP = new RegExp(
        "^([0-9][0-9][0-9][0-9])-([0-9][0-9])-([0-9][0-9])$"
      );
      var YAML_TIMESTAMP_REGEXP = new RegExp(
        "^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)(?:[Tt]|[ \\t]+)([0-9][0-9]?):([0-9][0-9]):([0-9][0-9])(?:\\.([0-9]*))?(?:[ \\t]*(Z|([-+])([0-9][0-9]?)(?::([0-9][0-9]))?))?$"
      );
      function resolveYamlTimestamp(data) {
        if (data === null)
          return false;
        if (YAML_DATE_REGEXP.exec(data) !== null)
          return true;
        if (YAML_TIMESTAMP_REGEXP.exec(data) !== null)
          return true;
        return false;
      }
      function constructYamlTimestamp(data) {
        var match, year, month, day, hour, minute, second, fraction = 0, delta = null, tz_hour, tz_minute, date;
        match = YAML_DATE_REGEXP.exec(data);
        if (match === null)
          match = YAML_TIMESTAMP_REGEXP.exec(data);
        if (match === null)
          throw new Error("Date resolve error");
        year = +match[1];
        month = +match[2] - 1;
        day = +match[3];
        if (!match[4]) {
          return new Date(Date.UTC(year, month, day));
        }
        hour = +match[4];
        minute = +match[5];
        second = +match[6];
        if (match[7]) {
          fraction = match[7].slice(0, 3);
          while (fraction.length < 3) {
            fraction += "0";
          }
          fraction = +fraction;
        }
        if (match[9]) {
          tz_hour = +match[10];
          tz_minute = +(match[11] || 0);
          delta = (tz_hour * 60 + tz_minute) * 6e4;
          if (match[9] === "-")
            delta = -delta;
        }
        date = new Date(Date.UTC(year, month, day, hour, minute, second, fraction));
        if (delta)
          date.setTime(date.getTime() - delta);
        return date;
      }
      function representYamlTimestamp(object) {
        return object.toISOString();
      }
      module.exports = new Type("tag:yaml.org,2002:timestamp", {
        kind: "scalar",
        resolve: resolveYamlTimestamp,
        construct: constructYamlTimestamp,
        instanceOf: Date,
        represent: representYamlTimestamp
      });
    }
  });

  // node_modules/js-yaml/lib/type/merge.js
  var require_merge = __commonJS({
    "node_modules/js-yaml/lib/type/merge.js"(exports, module) {
      "use strict";
      var Type = require_type();
      function resolveYamlMerge(data) {
        return data === "<<" || data === null;
      }
      module.exports = new Type("tag:yaml.org,2002:merge", {
        kind: "scalar",
        resolve: resolveYamlMerge
      });
    }
  });

  // node_modules/js-yaml/lib/type/binary.js
  var require_binary = __commonJS({
    "node_modules/js-yaml/lib/type/binary.js"(exports, module) {
      "use strict";
      var Type = require_type();
      var BASE64_MAP = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=\n\r";
      function resolveYamlBinary(data) {
        if (data === null)
          return false;
        var code, idx, bitlen = 0, max = data.length, map = BASE64_MAP;
        for (idx = 0; idx < max; idx++) {
          code = map.indexOf(data.charAt(idx));
          if (code > 64)
            continue;
          if (code < 0)
            return false;
          bitlen += 6;
        }
        return bitlen % 8 === 0;
      }
      function constructYamlBinary(data) {
        var idx, tailbits, input = data.replace(/[\r\n=]/g, ""), max = input.length, map = BASE64_MAP, bits = 0, result = [];
        for (idx = 0; idx < max; idx++) {
          if (idx % 4 === 0 && idx) {
            result.push(bits >> 16 & 255);
            result.push(bits >> 8 & 255);
            result.push(bits & 255);
          }
          bits = bits << 6 | map.indexOf(input.charAt(idx));
        }
        tailbits = max % 4 * 6;
        if (tailbits === 0) {
          result.push(bits >> 16 & 255);
          result.push(bits >> 8 & 255);
          result.push(bits & 255);
        } else if (tailbits === 18) {
          result.push(bits >> 10 & 255);
          result.push(bits >> 2 & 255);
        } else if (tailbits === 12) {
          result.push(bits >> 4 & 255);
        }
        return new Uint8Array(result);
      }
      function representYamlBinary(object) {
        var result = "", bits = 0, idx, tail, max = object.length, map = BASE64_MAP;
        for (idx = 0; idx < max; idx++) {
          if (idx % 3 === 0 && idx) {
            result += map[bits >> 18 & 63];
            result += map[bits >> 12 & 63];
            result += map[bits >> 6 & 63];
            result += map[bits & 63];
          }
          bits = (bits << 8) + object[idx];
        }
        tail = max % 3;
        if (tail === 0) {
          result += map[bits >> 18 & 63];
          result += map[bits >> 12 & 63];
          result += map[bits >> 6 & 63];
          result += map[bits & 63];
        } else if (tail === 2) {
          result += map[bits >> 10 & 63];
          result += map[bits >> 4 & 63];
          result += map[bits << 2 & 63];
          result += map[64];
        } else if (tail === 1) {
          result += map[bits >> 2 & 63];
          result += map[bits << 4 & 63];
          result += map[64];
          result += map[64];
        }
        return result;
      }
      function isBinary(obj) {
        return Object.prototype.toString.call(obj) === "[object Uint8Array]";
      }
      module.exports = new Type("tag:yaml.org,2002:binary", {
        kind: "scalar",
        resolve: resolveYamlBinary,
        construct: constructYamlBinary,
        predicate: isBinary,
        represent: representYamlBinary
      });
    }
  });

  // node_modules/js-yaml/lib/type/omap.js
  var require_omap = __commonJS({
    "node_modules/js-yaml/lib/type/omap.js"(exports, module) {
      "use strict";
      var Type = require_type();
      var _hasOwnProperty = Object.prototype.hasOwnProperty;
      var _toString = Object.prototype.toString;
      function resolveYamlOmap(data) {
        if (data === null)
          return true;
        var objectKeys = [], index, length, pair, pairKey, pairHasKey, object = data;
        for (index = 0, length = object.length; index < length; index += 1) {
          pair = object[index];
          pairHasKey = false;
          if (_toString.call(pair) !== "[object Object]")
            return false;
          for (pairKey in pair) {
            if (_hasOwnProperty.call(pair, pairKey)) {
              if (!pairHasKey)
                pairHasKey = true;
              else
                return false;
            }
          }
          if (!pairHasKey)
            return false;
          if (objectKeys.indexOf(pairKey) === -1)
            objectKeys.push(pairKey);
          else
            return false;
        }
        return true;
      }
      function constructYamlOmap(data) {
        return data !== null ? data : [];
      }
      module.exports = new Type("tag:yaml.org,2002:omap", {
        kind: "sequence",
        resolve: resolveYamlOmap,
        construct: constructYamlOmap
      });
    }
  });

  // node_modules/js-yaml/lib/type/pairs.js
  var require_pairs = __commonJS({
    "node_modules/js-yaml/lib/type/pairs.js"(exports, module) {
      "use strict";
      var Type = require_type();
      var _toString = Object.prototype.toString;
      function resolveYamlPairs(data) {
        if (data === null)
          return true;
        var index, length, pair, keys, result, object = data;
        result = new Array(object.length);
        for (index = 0, length = object.length; index < length; index += 1) {
          pair = object[index];
          if (_toString.call(pair) !== "[object Object]")
            return false;
          keys = Object.keys(pair);
          if (keys.length !== 1)
            return false;
          result[index] = [keys[0], pair[keys[0]]];
        }
        return true;
      }
      function constructYamlPairs(data) {
        if (data === null)
          return [];
        var index, length, pair, keys, result, object = data;
        result = new Array(object.length);
        for (index = 0, length = object.length; index < length; index += 1) {
          pair = object[index];
          keys = Object.keys(pair);
          result[index] = [keys[0], pair[keys[0]]];
        }
        return result;
      }
      module.exports = new Type("tag:yaml.org,2002:pairs", {
        kind: "sequence",
        resolve: resolveYamlPairs,
        construct: constructYamlPairs
      });
    }
  });

  // node_modules/js-yaml/lib/type/set.js
  var require_set = __commonJS({
    "node_modules/js-yaml/lib/type/set.js"(exports, module) {
      "use strict";
      var Type = require_type();
      var _hasOwnProperty = Object.prototype.hasOwnProperty;
      function resolveYamlSet(data) {
        if (data === null)
          return true;
        var key, object = data;
        for (key in object) {
          if (_hasOwnProperty.call(object, key)) {
            if (object[key] !== null)
              return false;
          }
        }
        return true;
      }
      function constructYamlSet(data) {
        return data !== null ? data : {};
      }
      module.exports = new Type("tag:yaml.org,2002:set", {
        kind: "mapping",
        resolve: resolveYamlSet,
        construct: constructYamlSet
      });
    }
  });

  // node_modules/js-yaml/lib/schema/default.js
  var require_default = __commonJS({
    "node_modules/js-yaml/lib/schema/default.js"(exports, module) {
      "use strict";
      module.exports = require_core().extend({
        implicit: [
          require_timestamp(),
          require_merge()
        ],
        explicit: [
          require_binary(),
          require_omap(),
          require_pairs(),
          require_set()
        ]
      });
    }
  });

  // node_modules/js-yaml/lib/loader.js
  var require_loader = __commonJS({
    "node_modules/js-yaml/lib/loader.js"(exports, module) {
      "use strict";
      var common = require_common();
      var YAMLException = require_exception();
      var makeSnippet = require_snippet();
      var DEFAULT_SCHEMA = require_default();
      var _hasOwnProperty = Object.prototype.hasOwnProperty;
      var CONTEXT_FLOW_IN = 1;
      var CONTEXT_FLOW_OUT = 2;
      var CONTEXT_BLOCK_IN = 3;
      var CONTEXT_BLOCK_OUT = 4;
      var CHOMPING_CLIP = 1;
      var CHOMPING_STRIP = 2;
      var CHOMPING_KEEP = 3;
      var PATTERN_NON_PRINTABLE = /[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x84\x86-\x9F\uFFFE\uFFFF]|[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?:[^\uD800-\uDBFF]|^)[\uDC00-\uDFFF]/;
      var PATTERN_NON_ASCII_LINE_BREAKS = /[\x85\u2028\u2029]/;
      var PATTERN_FLOW_INDICATORS = /[,\[\]\{\}]/;
      var PATTERN_TAG_HANDLE = /^(?:!|!!|![a-z\-]+!)$/i;
      var PATTERN_TAG_URI = /^(?:!|[^,\[\]\{\}])(?:%[0-9a-f]{2}|[0-9a-z\-#;\/\?:@&=\+\$,_\.!~\*'\(\)\[\]])*$/i;
      function _class(obj) {
        return Object.prototype.toString.call(obj);
      }
      function is_EOL(c) {
        return c === 10 || c === 13;
      }
      function is_WHITE_SPACE(c) {
        return c === 9 || c === 32;
      }
      function is_WS_OR_EOL(c) {
        return c === 9 || c === 32 || c === 10 || c === 13;
      }
      function is_FLOW_INDICATOR(c) {
        return c === 44 || c === 91 || c === 93 || c === 123 || c === 125;
      }
      function fromHexCode(c) {
        var lc;
        if (48 <= c && c <= 57) {
          return c - 48;
        }
        lc = c | 32;
        if (97 <= lc && lc <= 102) {
          return lc - 97 + 10;
        }
        return -1;
      }
      function escapedHexLen(c) {
        if (c === 120) {
          return 2;
        }
        if (c === 117) {
          return 4;
        }
        if (c === 85) {
          return 8;
        }
        return 0;
      }
      function fromDecimalCode(c) {
        if (48 <= c && c <= 57) {
          return c - 48;
        }
        return -1;
      }
      function simpleEscapeSequence(c) {
        return c === 48 ? "\0" : c === 97 ? "\x07" : c === 98 ? "\b" : c === 116 ? "	" : c === 9 ? "	" : c === 110 ? "\n" : c === 118 ? "\v" : c === 102 ? "\f" : c === 114 ? "\r" : c === 101 ? "\x1B" : c === 32 ? " " : c === 34 ? '"' : c === 47 ? "/" : c === 92 ? "\\" : c === 78 ? "\x85" : c === 95 ? "\xA0" : c === 76 ? "\u2028" : c === 80 ? "\u2029" : "";
      }
      function charFromCodepoint(c) {
        if (c <= 65535) {
          return String.fromCharCode(c);
        }
        return String.fromCharCode(
          (c - 65536 >> 10) + 55296,
          (c - 65536 & 1023) + 56320
        );
      }
      var simpleEscapeCheck = new Array(256);
      var simpleEscapeMap = new Array(256);
      for (i = 0; i < 256; i++) {
        simpleEscapeCheck[i] = simpleEscapeSequence(i) ? 1 : 0;
        simpleEscapeMap[i] = simpleEscapeSequence(i);
      }
      var i;
      function State(input, options) {
        this.input = input;
        this.filename = options["filename"] || null;
        this.schema = options["schema"] || DEFAULT_SCHEMA;
        this.onWarning = options["onWarning"] || null;
        this.legacy = options["legacy"] || false;
        this.json = options["json"] || false;
        this.listener = options["listener"] || null;
        this.implicitTypes = this.schema.compiledImplicit;
        this.typeMap = this.schema.compiledTypeMap;
        this.length = input.length;
        this.position = 0;
        this.line = 0;
        this.lineStart = 0;
        this.lineIndent = 0;
        this.firstTabInLine = -1;
        this.documents = [];
      }
      function generateError(state, message) {
        var mark = {
          name: state.filename,
          buffer: state.input.slice(0, -1),
          // omit trailing \0
          position: state.position,
          line: state.line,
          column: state.position - state.lineStart
        };
        mark.snippet = makeSnippet(mark);
        return new YAMLException(message, mark);
      }
      function throwError(state, message) {
        throw generateError(state, message);
      }
      function throwWarning(state, message) {
        if (state.onWarning) {
          state.onWarning.call(null, generateError(state, message));
        }
      }
      var directiveHandlers = {
        YAML: function handleYamlDirective(state, name, args) {
          var match, major, minor;
          if (state.version !== null) {
            throwError(state, "duplication of %YAML directive");
          }
          if (args.length !== 1) {
            throwError(state, "YAML directive accepts exactly one argument");
          }
          match = /^([0-9]+)\.([0-9]+)$/.exec(args[0]);
          if (match === null) {
            throwError(state, "ill-formed argument of the YAML directive");
          }
          major = parseInt(match[1], 10);
          minor = parseInt(match[2], 10);
          if (major !== 1) {
            throwError(state, "unacceptable YAML version of the document");
          }
          state.version = args[0];
          state.checkLineBreaks = minor < 2;
          if (minor !== 1 && minor !== 2) {
            throwWarning(state, "unsupported YAML version of the document");
          }
        },
        TAG: function handleTagDirective(state, name, args) {
          var handle, prefix;
          if (args.length !== 2) {
            throwError(state, "TAG directive accepts exactly two arguments");
          }
          handle = args[0];
          prefix = args[1];
          if (!PATTERN_TAG_HANDLE.test(handle)) {
            throwError(state, "ill-formed tag handle (first argument) of the TAG directive");
          }
          if (_hasOwnProperty.call(state.tagMap, handle)) {
            throwError(state, 'there is a previously declared suffix for "' + handle + '" tag handle');
          }
          if (!PATTERN_TAG_URI.test(prefix)) {
            throwError(state, "ill-formed tag prefix (second argument) of the TAG directive");
          }
          try {
            prefix = decodeURIComponent(prefix);
          } catch (err) {
            throwError(state, "tag prefix is malformed: " + prefix);
          }
          state.tagMap[handle] = prefix;
        }
      };
      function captureSegment(state, start, end, checkJson) {
        var _position, _length, _character, _result;
        if (start < end) {
          _result = state.input.slice(start, end);
          if (checkJson) {
            for (_position = 0, _length = _result.length; _position < _length; _position += 1) {
              _character = _result.charCodeAt(_position);
              if (!(_character === 9 || 32 <= _character && _character <= 1114111)) {
                throwError(state, "expected valid JSON character");
              }
            }
          } else if (PATTERN_NON_PRINTABLE.test(_result)) {
            throwError(state, "the stream contains non-printable characters");
          }
          state.result += _result;
        }
      }
      function mergeMappings(state, destination, source, overridableKeys) {
        var sourceKeys, key, index, quantity;
        if (!common.isObject(source)) {
          throwError(state, "cannot merge mappings; the provided source object is unacceptable");
        }
        sourceKeys = Object.keys(source);
        for (index = 0, quantity = sourceKeys.length; index < quantity; index += 1) {
          key = sourceKeys[index];
          if (!_hasOwnProperty.call(destination, key)) {
            destination[key] = source[key];
            overridableKeys[key] = true;
          }
        }
      }
      function storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, startLine, startLineStart, startPos) {
        var index, quantity;
        if (Array.isArray(keyNode)) {
          keyNode = Array.prototype.slice.call(keyNode);
          for (index = 0, quantity = keyNode.length; index < quantity; index += 1) {
            if (Array.isArray(keyNode[index])) {
              throwError(state, "nested arrays are not supported inside keys");
            }
            if (typeof keyNode === "object" && _class(keyNode[index]) === "[object Object]") {
              keyNode[index] = "[object Object]";
            }
          }
        }
        if (typeof keyNode === "object" && _class(keyNode) === "[object Object]") {
          keyNode = "[object Object]";
        }
        keyNode = String(keyNode);
        if (_result === null) {
          _result = {};
        }
        if (keyTag === "tag:yaml.org,2002:merge") {
          if (Array.isArray(valueNode)) {
            for (index = 0, quantity = valueNode.length; index < quantity; index += 1) {
              mergeMappings(state, _result, valueNode[index], overridableKeys);
            }
          } else {
            mergeMappings(state, _result, valueNode, overridableKeys);
          }
        } else {
          if (!state.json && !_hasOwnProperty.call(overridableKeys, keyNode) && _hasOwnProperty.call(_result, keyNode)) {
            state.line = startLine || state.line;
            state.lineStart = startLineStart || state.lineStart;
            state.position = startPos || state.position;
            throwError(state, "duplicated mapping key");
          }
          if (keyNode === "__proto__") {
            Object.defineProperty(_result, keyNode, {
              configurable: true,
              enumerable: true,
              writable: true,
              value: valueNode
            });
          } else {
            _result[keyNode] = valueNode;
          }
          delete overridableKeys[keyNode];
        }
        return _result;
      }
      function readLineBreak(state) {
        var ch;
        ch = state.input.charCodeAt(state.position);
        if (ch === 10) {
          state.position++;
        } else if (ch === 13) {
          state.position++;
          if (state.input.charCodeAt(state.position) === 10) {
            state.position++;
          }
        } else {
          throwError(state, "a line break is expected");
        }
        state.line += 1;
        state.lineStart = state.position;
        state.firstTabInLine = -1;
      }
      function skipSeparationSpace(state, allowComments, checkIndent) {
        var lineBreaks = 0, ch = state.input.charCodeAt(state.position);
        while (ch !== 0) {
          while (is_WHITE_SPACE(ch)) {
            if (ch === 9 && state.firstTabInLine === -1) {
              state.firstTabInLine = state.position;
            }
            ch = state.input.charCodeAt(++state.position);
          }
          if (allowComments && ch === 35) {
            do {
              ch = state.input.charCodeAt(++state.position);
            } while (ch !== 10 && ch !== 13 && ch !== 0);
          }
          if (is_EOL(ch)) {
            readLineBreak(state);
            ch = state.input.charCodeAt(state.position);
            lineBreaks++;
            state.lineIndent = 0;
            while (ch === 32) {
              state.lineIndent++;
              ch = state.input.charCodeAt(++state.position);
            }
          } else {
            break;
          }
        }
        if (checkIndent !== -1 && lineBreaks !== 0 && state.lineIndent < checkIndent) {
          throwWarning(state, "deficient indentation");
        }
        return lineBreaks;
      }
      function testDocumentSeparator(state) {
        var _position = state.position, ch;
        ch = state.input.charCodeAt(_position);
        if ((ch === 45 || ch === 46) && ch === state.input.charCodeAt(_position + 1) && ch === state.input.charCodeAt(_position + 2)) {
          _position += 3;
          ch = state.input.charCodeAt(_position);
          if (ch === 0 || is_WS_OR_EOL(ch)) {
            return true;
          }
        }
        return false;
      }
      function writeFoldedLines(state, count) {
        if (count === 1) {
          state.result += " ";
        } else if (count > 1) {
          state.result += common.repeat("\n", count - 1);
        }
      }
      function readPlainScalar(state, nodeIndent, withinFlowCollection) {
        var preceding, following, captureStart, captureEnd, hasPendingContent, _line, _lineStart, _lineIndent, _kind = state.kind, _result = state.result, ch;
        ch = state.input.charCodeAt(state.position);
        if (is_WS_OR_EOL(ch) || is_FLOW_INDICATOR(ch) || ch === 35 || ch === 38 || ch === 42 || ch === 33 || ch === 124 || ch === 62 || ch === 39 || ch === 34 || ch === 37 || ch === 64 || ch === 96) {
          return false;
        }
        if (ch === 63 || ch === 45) {
          following = state.input.charCodeAt(state.position + 1);
          if (is_WS_OR_EOL(following) || withinFlowCollection && is_FLOW_INDICATOR(following)) {
            return false;
          }
        }
        state.kind = "scalar";
        state.result = "";
        captureStart = captureEnd = state.position;
        hasPendingContent = false;
        while (ch !== 0) {
          if (ch === 58) {
            following = state.input.charCodeAt(state.position + 1);
            if (is_WS_OR_EOL(following) || withinFlowCollection && is_FLOW_INDICATOR(following)) {
              break;
            }
          } else if (ch === 35) {
            preceding = state.input.charCodeAt(state.position - 1);
            if (is_WS_OR_EOL(preceding)) {
              break;
            }
          } else if (state.position === state.lineStart && testDocumentSeparator(state) || withinFlowCollection && is_FLOW_INDICATOR(ch)) {
            break;
          } else if (is_EOL(ch)) {
            _line = state.line;
            _lineStart = state.lineStart;
            _lineIndent = state.lineIndent;
            skipSeparationSpace(state, false, -1);
            if (state.lineIndent >= nodeIndent) {
              hasPendingContent = true;
              ch = state.input.charCodeAt(state.position);
              continue;
            } else {
              state.position = captureEnd;
              state.line = _line;
              state.lineStart = _lineStart;
              state.lineIndent = _lineIndent;
              break;
            }
          }
          if (hasPendingContent) {
            captureSegment(state, captureStart, captureEnd, false);
            writeFoldedLines(state, state.line - _line);
            captureStart = captureEnd = state.position;
            hasPendingContent = false;
          }
          if (!is_WHITE_SPACE(ch)) {
            captureEnd = state.position + 1;
          }
          ch = state.input.charCodeAt(++state.position);
        }
        captureSegment(state, captureStart, captureEnd, false);
        if (state.result) {
          return true;
        }
        state.kind = _kind;
        state.result = _result;
        return false;
      }
      function readSingleQuotedScalar(state, nodeIndent) {
        var ch, captureStart, captureEnd;
        ch = state.input.charCodeAt(state.position);
        if (ch !== 39) {
          return false;
        }
        state.kind = "scalar";
        state.result = "";
        state.position++;
        captureStart = captureEnd = state.position;
        while ((ch = state.input.charCodeAt(state.position)) !== 0) {
          if (ch === 39) {
            captureSegment(state, captureStart, state.position, true);
            ch = state.input.charCodeAt(++state.position);
            if (ch === 39) {
              captureStart = state.position;
              state.position++;
              captureEnd = state.position;
            } else {
              return true;
            }
          } else if (is_EOL(ch)) {
            captureSegment(state, captureStart, captureEnd, true);
            writeFoldedLines(state, skipSeparationSpace(state, false, nodeIndent));
            captureStart = captureEnd = state.position;
          } else if (state.position === state.lineStart && testDocumentSeparator(state)) {
            throwError(state, "unexpected end of the document within a single quoted scalar");
          } else {
            state.position++;
            captureEnd = state.position;
          }
        }
        throwError(state, "unexpected end of the stream within a single quoted scalar");
      }
      function readDoubleQuotedScalar(state, nodeIndent) {
        var captureStart, captureEnd, hexLength, hexResult, tmp, ch;
        ch = state.input.charCodeAt(state.position);
        if (ch !== 34) {
          return false;
        }
        state.kind = "scalar";
        state.result = "";
        state.position++;
        captureStart = captureEnd = state.position;
        while ((ch = state.input.charCodeAt(state.position)) !== 0) {
          if (ch === 34) {
            captureSegment(state, captureStart, state.position, true);
            state.position++;
            return true;
          } else if (ch === 92) {
            captureSegment(state, captureStart, state.position, true);
            ch = state.input.charCodeAt(++state.position);
            if (is_EOL(ch)) {
              skipSeparationSpace(state, false, nodeIndent);
            } else if (ch < 256 && simpleEscapeCheck[ch]) {
              state.result += simpleEscapeMap[ch];
              state.position++;
            } else if ((tmp = escapedHexLen(ch)) > 0) {
              hexLength = tmp;
              hexResult = 0;
              for (; hexLength > 0; hexLength--) {
                ch = state.input.charCodeAt(++state.position);
                if ((tmp = fromHexCode(ch)) >= 0) {
                  hexResult = (hexResult << 4) + tmp;
                } else {
                  throwError(state, "expected hexadecimal character");
                }
              }
              state.result += charFromCodepoint(hexResult);
              state.position++;
            } else {
              throwError(state, "unknown escape sequence");
            }
            captureStart = captureEnd = state.position;
          } else if (is_EOL(ch)) {
            captureSegment(state, captureStart, captureEnd, true);
            writeFoldedLines(state, skipSeparationSpace(state, false, nodeIndent));
            captureStart = captureEnd = state.position;
          } else if (state.position === state.lineStart && testDocumentSeparator(state)) {
            throwError(state, "unexpected end of the document within a double quoted scalar");
          } else {
            state.position++;
            captureEnd = state.position;
          }
        }
        throwError(state, "unexpected end of the stream within a double quoted scalar");
      }
      function readFlowCollection(state, nodeIndent) {
        var readNext = true, _line, _lineStart, _pos, _tag = state.tag, _result, _anchor = state.anchor, following, terminator, isPair, isExplicitPair, isMapping, overridableKeys = /* @__PURE__ */ Object.create(null), keyNode, keyTag, valueNode, ch;
        ch = state.input.charCodeAt(state.position);
        if (ch === 91) {
          terminator = 93;
          isMapping = false;
          _result = [];
        } else if (ch === 123) {
          terminator = 125;
          isMapping = true;
          _result = {};
        } else {
          return false;
        }
        if (state.anchor !== null) {
          state.anchorMap[state.anchor] = _result;
        }
        ch = state.input.charCodeAt(++state.position);
        while (ch !== 0) {
          skipSeparationSpace(state, true, nodeIndent);
          ch = state.input.charCodeAt(state.position);
          if (ch === terminator) {
            state.position++;
            state.tag = _tag;
            state.anchor = _anchor;
            state.kind = isMapping ? "mapping" : "sequence";
            state.result = _result;
            return true;
          } else if (!readNext) {
            throwError(state, "missed comma between flow collection entries");
          } else if (ch === 44) {
            throwError(state, "expected the node content, but found ','");
          }
          keyTag = keyNode = valueNode = null;
          isPair = isExplicitPair = false;
          if (ch === 63) {
            following = state.input.charCodeAt(state.position + 1);
            if (is_WS_OR_EOL(following)) {
              isPair = isExplicitPair = true;
              state.position++;
              skipSeparationSpace(state, true, nodeIndent);
            }
          }
          _line = state.line;
          _lineStart = state.lineStart;
          _pos = state.position;
          composeNode(state, nodeIndent, CONTEXT_FLOW_IN, false, true);
          keyTag = state.tag;
          keyNode = state.result;
          skipSeparationSpace(state, true, nodeIndent);
          ch = state.input.charCodeAt(state.position);
          if ((isExplicitPair || state.line === _line) && ch === 58) {
            isPair = true;
            ch = state.input.charCodeAt(++state.position);
            skipSeparationSpace(state, true, nodeIndent);
            composeNode(state, nodeIndent, CONTEXT_FLOW_IN, false, true);
            valueNode = state.result;
          }
          if (isMapping) {
            storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, _line, _lineStart, _pos);
          } else if (isPair) {
            _result.push(storeMappingPair(state, null, overridableKeys, keyTag, keyNode, valueNode, _line, _lineStart, _pos));
          } else {
            _result.push(keyNode);
          }
          skipSeparationSpace(state, true, nodeIndent);
          ch = state.input.charCodeAt(state.position);
          if (ch === 44) {
            readNext = true;
            ch = state.input.charCodeAt(++state.position);
          } else {
            readNext = false;
          }
        }
        throwError(state, "unexpected end of the stream within a flow collection");
      }
      function readBlockScalar(state, nodeIndent) {
        var captureStart, folding, chomping = CHOMPING_CLIP, didReadContent = false, detectedIndent = false, textIndent = nodeIndent, emptyLines = 0, atMoreIndented = false, tmp, ch;
        ch = state.input.charCodeAt(state.position);
        if (ch === 124) {
          folding = false;
        } else if (ch === 62) {
          folding = true;
        } else {
          return false;
        }
        state.kind = "scalar";
        state.result = "";
        while (ch !== 0) {
          ch = state.input.charCodeAt(++state.position);
          if (ch === 43 || ch === 45) {
            if (CHOMPING_CLIP === chomping) {
              chomping = ch === 43 ? CHOMPING_KEEP : CHOMPING_STRIP;
            } else {
              throwError(state, "repeat of a chomping mode identifier");
            }
          } else if ((tmp = fromDecimalCode(ch)) >= 0) {
            if (tmp === 0) {
              throwError(state, "bad explicit indentation width of a block scalar; it cannot be less than one");
            } else if (!detectedIndent) {
              textIndent = nodeIndent + tmp - 1;
              detectedIndent = true;
            } else {
              throwError(state, "repeat of an indentation width identifier");
            }
          } else {
            break;
          }
        }
        if (is_WHITE_SPACE(ch)) {
          do {
            ch = state.input.charCodeAt(++state.position);
          } while (is_WHITE_SPACE(ch));
          if (ch === 35) {
            do {
              ch = state.input.charCodeAt(++state.position);
            } while (!is_EOL(ch) && ch !== 0);
          }
        }
        while (ch !== 0) {
          readLineBreak(state);
          state.lineIndent = 0;
          ch = state.input.charCodeAt(state.position);
          while ((!detectedIndent || state.lineIndent < textIndent) && ch === 32) {
            state.lineIndent++;
            ch = state.input.charCodeAt(++state.position);
          }
          if (!detectedIndent && state.lineIndent > textIndent) {
            textIndent = state.lineIndent;
          }
          if (is_EOL(ch)) {
            emptyLines++;
            continue;
          }
          if (state.lineIndent < textIndent) {
            if (chomping === CHOMPING_KEEP) {
              state.result += common.repeat("\n", didReadContent ? 1 + emptyLines : emptyLines);
            } else if (chomping === CHOMPING_CLIP) {
              if (didReadContent) {
                state.result += "\n";
              }
            }
            break;
          }
          if (folding) {
            if (is_WHITE_SPACE(ch)) {
              atMoreIndented = true;
              state.result += common.repeat("\n", didReadContent ? 1 + emptyLines : emptyLines);
            } else if (atMoreIndented) {
              atMoreIndented = false;
              state.result += common.repeat("\n", emptyLines + 1);
            } else if (emptyLines === 0) {
              if (didReadContent) {
                state.result += " ";
              }
            } else {
              state.result += common.repeat("\n", emptyLines);
            }
          } else {
            state.result += common.repeat("\n", didReadContent ? 1 + emptyLines : emptyLines);
          }
          didReadContent = true;
          detectedIndent = true;
          emptyLines = 0;
          captureStart = state.position;
          while (!is_EOL(ch) && ch !== 0) {
            ch = state.input.charCodeAt(++state.position);
          }
          captureSegment(state, captureStart, state.position, false);
        }
        return true;
      }
      function readBlockSequence(state, nodeIndent) {
        var _line, _tag = state.tag, _anchor = state.anchor, _result = [], following, detected = false, ch;
        if (state.firstTabInLine !== -1)
          return false;
        if (state.anchor !== null) {
          state.anchorMap[state.anchor] = _result;
        }
        ch = state.input.charCodeAt(state.position);
        while (ch !== 0) {
          if (state.firstTabInLine !== -1) {
            state.position = state.firstTabInLine;
            throwError(state, "tab characters must not be used in indentation");
          }
          if (ch !== 45) {
            break;
          }
          following = state.input.charCodeAt(state.position + 1);
          if (!is_WS_OR_EOL(following)) {
            break;
          }
          detected = true;
          state.position++;
          if (skipSeparationSpace(state, true, -1)) {
            if (state.lineIndent <= nodeIndent) {
              _result.push(null);
              ch = state.input.charCodeAt(state.position);
              continue;
            }
          }
          _line = state.line;
          composeNode(state, nodeIndent, CONTEXT_BLOCK_IN, false, true);
          _result.push(state.result);
          skipSeparationSpace(state, true, -1);
          ch = state.input.charCodeAt(state.position);
          if ((state.line === _line || state.lineIndent > nodeIndent) && ch !== 0) {
            throwError(state, "bad indentation of a sequence entry");
          } else if (state.lineIndent < nodeIndent) {
            break;
          }
        }
        if (detected) {
          state.tag = _tag;
          state.anchor = _anchor;
          state.kind = "sequence";
          state.result = _result;
          return true;
        }
        return false;
      }
      function readBlockMapping(state, nodeIndent, flowIndent) {
        var following, allowCompact, _line, _keyLine, _keyLineStart, _keyPos, _tag = state.tag, _anchor = state.anchor, _result = {}, overridableKeys = /* @__PURE__ */ Object.create(null), keyTag = null, keyNode = null, valueNode = null, atExplicitKey = false, detected = false, ch;
        if (state.firstTabInLine !== -1)
          return false;
        if (state.anchor !== null) {
          state.anchorMap[state.anchor] = _result;
        }
        ch = state.input.charCodeAt(state.position);
        while (ch !== 0) {
          if (!atExplicitKey && state.firstTabInLine !== -1) {
            state.position = state.firstTabInLine;
            throwError(state, "tab characters must not be used in indentation");
          }
          following = state.input.charCodeAt(state.position + 1);
          _line = state.line;
          if ((ch === 63 || ch === 58) && is_WS_OR_EOL(following)) {
            if (ch === 63) {
              if (atExplicitKey) {
                storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
                keyTag = keyNode = valueNode = null;
              }
              detected = true;
              atExplicitKey = true;
              allowCompact = true;
            } else if (atExplicitKey) {
              atExplicitKey = false;
              allowCompact = true;
            } else {
              throwError(state, "incomplete explicit mapping pair; a key node is missed; or followed by a non-tabulated empty line");
            }
            state.position += 1;
            ch = following;
          } else {
            _keyLine = state.line;
            _keyLineStart = state.lineStart;
            _keyPos = state.position;
            if (!composeNode(state, flowIndent, CONTEXT_FLOW_OUT, false, true)) {
              break;
            }
            if (state.line === _line) {
              ch = state.input.charCodeAt(state.position);
              while (is_WHITE_SPACE(ch)) {
                ch = state.input.charCodeAt(++state.position);
              }
              if (ch === 58) {
                ch = state.input.charCodeAt(++state.position);
                if (!is_WS_OR_EOL(ch)) {
                  throwError(state, "a whitespace character is expected after the key-value separator within a block mapping");
                }
                if (atExplicitKey) {
                  storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
                  keyTag = keyNode = valueNode = null;
                }
                detected = true;
                atExplicitKey = false;
                allowCompact = false;
                keyTag = state.tag;
                keyNode = state.result;
              } else if (detected) {
                throwError(state, "can not read an implicit mapping pair; a colon is missed");
              } else {
                state.tag = _tag;
                state.anchor = _anchor;
                return true;
              }
            } else if (detected) {
              throwError(state, "can not read a block mapping entry; a multiline key may not be an implicit key");
            } else {
              state.tag = _tag;
              state.anchor = _anchor;
              return true;
            }
          }
          if (state.line === _line || state.lineIndent > nodeIndent) {
            if (atExplicitKey) {
              _keyLine = state.line;
              _keyLineStart = state.lineStart;
              _keyPos = state.position;
            }
            if (composeNode(state, nodeIndent, CONTEXT_BLOCK_OUT, true, allowCompact)) {
              if (atExplicitKey) {
                keyNode = state.result;
              } else {
                valueNode = state.result;
              }
            }
            if (!atExplicitKey) {
              storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, _keyLine, _keyLineStart, _keyPos);
              keyTag = keyNode = valueNode = null;
            }
            skipSeparationSpace(state, true, -1);
            ch = state.input.charCodeAt(state.position);
          }
          if ((state.line === _line || state.lineIndent > nodeIndent) && ch !== 0) {
            throwError(state, "bad indentation of a mapping entry");
          } else if (state.lineIndent < nodeIndent) {
            break;
          }
        }
        if (atExplicitKey) {
          storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
        }
        if (detected) {
          state.tag = _tag;
          state.anchor = _anchor;
          state.kind = "mapping";
          state.result = _result;
        }
        return detected;
      }
      function readTagProperty(state) {
        var _position, isVerbatim = false, isNamed = false, tagHandle, tagName, ch;
        ch = state.input.charCodeAt(state.position);
        if (ch !== 33)
          return false;
        if (state.tag !== null) {
          throwError(state, "duplication of a tag property");
        }
        ch = state.input.charCodeAt(++state.position);
        if (ch === 60) {
          isVerbatim = true;
          ch = state.input.charCodeAt(++state.position);
        } else if (ch === 33) {
          isNamed = true;
          tagHandle = "!!";
          ch = state.input.charCodeAt(++state.position);
        } else {
          tagHandle = "!";
        }
        _position = state.position;
        if (isVerbatim) {
          do {
            ch = state.input.charCodeAt(++state.position);
          } while (ch !== 0 && ch !== 62);
          if (state.position < state.length) {
            tagName = state.input.slice(_position, state.position);
            ch = state.input.charCodeAt(++state.position);
          } else {
            throwError(state, "unexpected end of the stream within a verbatim tag");
          }
        } else {
          while (ch !== 0 && !is_WS_OR_EOL(ch)) {
            if (ch === 33) {
              if (!isNamed) {
                tagHandle = state.input.slice(_position - 1, state.position + 1);
                if (!PATTERN_TAG_HANDLE.test(tagHandle)) {
                  throwError(state, "named tag handle cannot contain such characters");
                }
                isNamed = true;
                _position = state.position + 1;
              } else {
                throwError(state, "tag suffix cannot contain exclamation marks");
              }
            }
            ch = state.input.charCodeAt(++state.position);
          }
          tagName = state.input.slice(_position, state.position);
          if (PATTERN_FLOW_INDICATORS.test(tagName)) {
            throwError(state, "tag suffix cannot contain flow indicator characters");
          }
        }
        if (tagName && !PATTERN_TAG_URI.test(tagName)) {
          throwError(state, "tag name cannot contain such characters: " + tagName);
        }
        try {
          tagName = decodeURIComponent(tagName);
        } catch (err) {
          throwError(state, "tag name is malformed: " + tagName);
        }
        if (isVerbatim) {
          state.tag = tagName;
        } else if (_hasOwnProperty.call(state.tagMap, tagHandle)) {
          state.tag = state.tagMap[tagHandle] + tagName;
        } else if (tagHandle === "!") {
          state.tag = "!" + tagName;
        } else if (tagHandle === "!!") {
          state.tag = "tag:yaml.org,2002:" + tagName;
        } else {
          throwError(state, 'undeclared tag handle "' + tagHandle + '"');
        }
        return true;
      }
      function readAnchorProperty(state) {
        var _position, ch;
        ch = state.input.charCodeAt(state.position);
        if (ch !== 38)
          return false;
        if (state.anchor !== null) {
          throwError(state, "duplication of an anchor property");
        }
        ch = state.input.charCodeAt(++state.position);
        _position = state.position;
        while (ch !== 0 && !is_WS_OR_EOL(ch) && !is_FLOW_INDICATOR(ch)) {
          ch = state.input.charCodeAt(++state.position);
        }
        if (state.position === _position) {
          throwError(state, "name of an anchor node must contain at least one character");
        }
        state.anchor = state.input.slice(_position, state.position);
        return true;
      }
      function readAlias(state) {
        var _position, alias, ch;
        ch = state.input.charCodeAt(state.position);
        if (ch !== 42)
          return false;
        ch = state.input.charCodeAt(++state.position);
        _position = state.position;
        while (ch !== 0 && !is_WS_OR_EOL(ch) && !is_FLOW_INDICATOR(ch)) {
          ch = state.input.charCodeAt(++state.position);
        }
        if (state.position === _position) {
          throwError(state, "name of an alias node must contain at least one character");
        }
        alias = state.input.slice(_position, state.position);
        if (!_hasOwnProperty.call(state.anchorMap, alias)) {
          throwError(state, 'unidentified alias "' + alias + '"');
        }
        state.result = state.anchorMap[alias];
        skipSeparationSpace(state, true, -1);
        return true;
      }
      function composeNode(state, parentIndent, nodeContext, allowToSeek, allowCompact) {
        var allowBlockStyles, allowBlockScalars, allowBlockCollections, indentStatus = 1, atNewLine = false, hasContent = false, typeIndex, typeQuantity, typeList, type, flowIndent, blockIndent;
        if (state.listener !== null) {
          state.listener("open", state);
        }
        state.tag = null;
        state.anchor = null;
        state.kind = null;
        state.result = null;
        allowBlockStyles = allowBlockScalars = allowBlockCollections = CONTEXT_BLOCK_OUT === nodeContext || CONTEXT_BLOCK_IN === nodeContext;
        if (allowToSeek) {
          if (skipSeparationSpace(state, true, -1)) {
            atNewLine = true;
            if (state.lineIndent > parentIndent) {
              indentStatus = 1;
            } else if (state.lineIndent === parentIndent) {
              indentStatus = 0;
            } else if (state.lineIndent < parentIndent) {
              indentStatus = -1;
            }
          }
        }
        if (indentStatus === 1) {
          while (readTagProperty(state) || readAnchorProperty(state)) {
            if (skipSeparationSpace(state, true, -1)) {
              atNewLine = true;
              allowBlockCollections = allowBlockStyles;
              if (state.lineIndent > parentIndent) {
                indentStatus = 1;
              } else if (state.lineIndent === parentIndent) {
                indentStatus = 0;
              } else if (state.lineIndent < parentIndent) {
                indentStatus = -1;
              }
            } else {
              allowBlockCollections = false;
            }
          }
        }
        if (allowBlockCollections) {
          allowBlockCollections = atNewLine || allowCompact;
        }
        if (indentStatus === 1 || CONTEXT_BLOCK_OUT === nodeContext) {
          if (CONTEXT_FLOW_IN === nodeContext || CONTEXT_FLOW_OUT === nodeContext) {
            flowIndent = parentIndent;
          } else {
            flowIndent = parentIndent + 1;
          }
          blockIndent = state.position - state.lineStart;
          if (indentStatus === 1) {
            if (allowBlockCollections && (readBlockSequence(state, blockIndent) || readBlockMapping(state, blockIndent, flowIndent)) || readFlowCollection(state, flowIndent)) {
              hasContent = true;
            } else {
              if (allowBlockScalars && readBlockScalar(state, flowIndent) || readSingleQuotedScalar(state, flowIndent) || readDoubleQuotedScalar(state, flowIndent)) {
                hasContent = true;
              } else if (readAlias(state)) {
                hasContent = true;
                if (state.tag !== null || state.anchor !== null) {
                  throwError(state, "alias node should not have any properties");
                }
              } else if (readPlainScalar(state, flowIndent, CONTEXT_FLOW_IN === nodeContext)) {
                hasContent = true;
                if (state.tag === null) {
                  state.tag = "?";
                }
              }
              if (state.anchor !== null) {
                state.anchorMap[state.anchor] = state.result;
              }
            }
          } else if (indentStatus === 0) {
            hasContent = allowBlockCollections && readBlockSequence(state, blockIndent);
          }
        }
        if (state.tag === null) {
          if (state.anchor !== null) {
            state.anchorMap[state.anchor] = state.result;
          }
        } else if (state.tag === "?") {
          if (state.result !== null && state.kind !== "scalar") {
            throwError(state, 'unacceptable node kind for !<?> tag; it should be "scalar", not "' + state.kind + '"');
          }
          for (typeIndex = 0, typeQuantity = state.implicitTypes.length; typeIndex < typeQuantity; typeIndex += 1) {
            type = state.implicitTypes[typeIndex];
            if (type.resolve(state.result)) {
              state.result = type.construct(state.result);
              state.tag = type.tag;
              if (state.anchor !== null) {
                state.anchorMap[state.anchor] = state.result;
              }
              break;
            }
          }
        } else if (state.tag !== "!") {
          if (_hasOwnProperty.call(state.typeMap[state.kind || "fallback"], state.tag)) {
            type = state.typeMap[state.kind || "fallback"][state.tag];
          } else {
            type = null;
            typeList = state.typeMap.multi[state.kind || "fallback"];
            for (typeIndex = 0, typeQuantity = typeList.length; typeIndex < typeQuantity; typeIndex += 1) {
              if (state.tag.slice(0, typeList[typeIndex].tag.length) === typeList[typeIndex].tag) {
                type = typeList[typeIndex];
                break;
              }
            }
          }
          if (!type) {
            throwError(state, "unknown tag !<" + state.tag + ">");
          }
          if (state.result !== null && type.kind !== state.kind) {
            throwError(state, "unacceptable node kind for !<" + state.tag + '> tag; it should be "' + type.kind + '", not "' + state.kind + '"');
          }
          if (!type.resolve(state.result, state.tag)) {
            throwError(state, "cannot resolve a node with !<" + state.tag + "> explicit tag");
          } else {
            state.result = type.construct(state.result, state.tag);
            if (state.anchor !== null) {
              state.anchorMap[state.anchor] = state.result;
            }
          }
        }
        if (state.listener !== null) {
          state.listener("close", state);
        }
        return state.tag !== null || state.anchor !== null || hasContent;
      }
      function readDocument(state) {
        var documentStart = state.position, _position, directiveName, directiveArgs, hasDirectives = false, ch;
        state.version = null;
        state.checkLineBreaks = state.legacy;
        state.tagMap = /* @__PURE__ */ Object.create(null);
        state.anchorMap = /* @__PURE__ */ Object.create(null);
        while ((ch = state.input.charCodeAt(state.position)) !== 0) {
          skipSeparationSpace(state, true, -1);
          ch = state.input.charCodeAt(state.position);
          if (state.lineIndent > 0 || ch !== 37) {
            break;
          }
          hasDirectives = true;
          ch = state.input.charCodeAt(++state.position);
          _position = state.position;
          while (ch !== 0 && !is_WS_OR_EOL(ch)) {
            ch = state.input.charCodeAt(++state.position);
          }
          directiveName = state.input.slice(_position, state.position);
          directiveArgs = [];
          if (directiveName.length < 1) {
            throwError(state, "directive name must not be less than one character in length");
          }
          while (ch !== 0) {
            while (is_WHITE_SPACE(ch)) {
              ch = state.input.charCodeAt(++state.position);
            }
            if (ch === 35) {
              do {
                ch = state.input.charCodeAt(++state.position);
              } while (ch !== 0 && !is_EOL(ch));
              break;
            }
            if (is_EOL(ch))
              break;
            _position = state.position;
            while (ch !== 0 && !is_WS_OR_EOL(ch)) {
              ch = state.input.charCodeAt(++state.position);
            }
            directiveArgs.push(state.input.slice(_position, state.position));
          }
          if (ch !== 0)
            readLineBreak(state);
          if (_hasOwnProperty.call(directiveHandlers, directiveName)) {
            directiveHandlers[directiveName](state, directiveName, directiveArgs);
          } else {
            throwWarning(state, 'unknown document directive "' + directiveName + '"');
          }
        }
        skipSeparationSpace(state, true, -1);
        if (state.lineIndent === 0 && state.input.charCodeAt(state.position) === 45 && state.input.charCodeAt(state.position + 1) === 45 && state.input.charCodeAt(state.position + 2) === 45) {
          state.position += 3;
          skipSeparationSpace(state, true, -1);
        } else if (hasDirectives) {
          throwError(state, "directives end mark is expected");
        }
        composeNode(state, state.lineIndent - 1, CONTEXT_BLOCK_OUT, false, true);
        skipSeparationSpace(state, true, -1);
        if (state.checkLineBreaks && PATTERN_NON_ASCII_LINE_BREAKS.test(state.input.slice(documentStart, state.position))) {
          throwWarning(state, "non-ASCII line breaks are interpreted as content");
        }
        state.documents.push(state.result);
        if (state.position === state.lineStart && testDocumentSeparator(state)) {
          if (state.input.charCodeAt(state.position) === 46) {
            state.position += 3;
            skipSeparationSpace(state, true, -1);
          }
          return;
        }
        if (state.position < state.length - 1) {
          throwError(state, "end of the stream or a document separator is expected");
        } else {
          return;
        }
      }
      function loadDocuments(input, options) {
        input = String(input);
        options = options || {};
        if (input.length !== 0) {
          if (input.charCodeAt(input.length - 1) !== 10 && input.charCodeAt(input.length - 1) !== 13) {
            input += "\n";
          }
          if (input.charCodeAt(0) === 65279) {
            input = input.slice(1);
          }
        }
        var state = new State(input, options);
        var nullpos = input.indexOf("\0");
        if (nullpos !== -1) {
          state.position = nullpos;
          throwError(state, "null byte is not allowed in input");
        }
        state.input += "\0";
        while (state.input.charCodeAt(state.position) === 32) {
          state.lineIndent += 1;
          state.position += 1;
        }
        while (state.position < state.length - 1) {
          readDocument(state);
        }
        return state.documents;
      }
      function loadAll(input, iterator, options) {
        if (iterator !== null && typeof iterator === "object" && typeof options === "undefined") {
          options = iterator;
          iterator = null;
        }
        var documents = loadDocuments(input, options);
        if (typeof iterator !== "function") {
          return documents;
        }
        for (var index = 0, length = documents.length; index < length; index += 1) {
          iterator(documents[index]);
        }
      }
      function load(input, options) {
        var documents = loadDocuments(input, options);
        if (documents.length === 0) {
          return void 0;
        } else if (documents.length === 1) {
          return documents[0];
        }
        throw new YAMLException("expected a single document in the stream, but found more");
      }
      module.exports.loadAll = loadAll;
      module.exports.load = load;
    }
  });

  // node_modules/js-yaml/lib/dumper.js
  var require_dumper = __commonJS({
    "node_modules/js-yaml/lib/dumper.js"(exports, module) {
      "use strict";
      var common = require_common();
      var YAMLException = require_exception();
      var DEFAULT_SCHEMA = require_default();
      var _toString = Object.prototype.toString;
      var _hasOwnProperty = Object.prototype.hasOwnProperty;
      var CHAR_BOM = 65279;
      var CHAR_TAB = 9;
      var CHAR_LINE_FEED = 10;
      var CHAR_CARRIAGE_RETURN = 13;
      var CHAR_SPACE = 32;
      var CHAR_EXCLAMATION = 33;
      var CHAR_DOUBLE_QUOTE = 34;
      var CHAR_SHARP = 35;
      var CHAR_PERCENT = 37;
      var CHAR_AMPERSAND = 38;
      var CHAR_SINGLE_QUOTE = 39;
      var CHAR_ASTERISK = 42;
      var CHAR_COMMA = 44;
      var CHAR_MINUS = 45;
      var CHAR_COLON = 58;
      var CHAR_EQUALS = 61;
      var CHAR_GREATER_THAN = 62;
      var CHAR_QUESTION = 63;
      var CHAR_COMMERCIAL_AT = 64;
      var CHAR_LEFT_SQUARE_BRACKET = 91;
      var CHAR_RIGHT_SQUARE_BRACKET = 93;
      var CHAR_GRAVE_ACCENT = 96;
      var CHAR_LEFT_CURLY_BRACKET = 123;
      var CHAR_VERTICAL_LINE = 124;
      var CHAR_RIGHT_CURLY_BRACKET = 125;
      var ESCAPE_SEQUENCES = {};
      ESCAPE_SEQUENCES[0] = "\\0";
      ESCAPE_SEQUENCES[7] = "\\a";
      ESCAPE_SEQUENCES[8] = "\\b";
      ESCAPE_SEQUENCES[9] = "\\t";
      ESCAPE_SEQUENCES[10] = "\\n";
      ESCAPE_SEQUENCES[11] = "\\v";
      ESCAPE_SEQUENCES[12] = "\\f";
      ESCAPE_SEQUENCES[13] = "\\r";
      ESCAPE_SEQUENCES[27] = "\\e";
      ESCAPE_SEQUENCES[34] = '\\"';
      ESCAPE_SEQUENCES[92] = "\\\\";
      ESCAPE_SEQUENCES[133] = "\\N";
      ESCAPE_SEQUENCES[160] = "\\_";
      ESCAPE_SEQUENCES[8232] = "\\L";
      ESCAPE_SEQUENCES[8233] = "\\P";
      var DEPRECATED_BOOLEANS_SYNTAX = [
        "y",
        "Y",
        "yes",
        "Yes",
        "YES",
        "on",
        "On",
        "ON",
        "n",
        "N",
        "no",
        "No",
        "NO",
        "off",
        "Off",
        "OFF"
      ];
      var DEPRECATED_BASE60_SYNTAX = /^[-+]?[0-9_]+(?::[0-9_]+)+(?:\.[0-9_]*)?$/;
      function compileStyleMap(schema, map) {
        var result, keys, index, length, tag, style, type;
        if (map === null)
          return {};
        result = {};
        keys = Object.keys(map);
        for (index = 0, length = keys.length; index < length; index += 1) {
          tag = keys[index];
          style = String(map[tag]);
          if (tag.slice(0, 2) === "!!") {
            tag = "tag:yaml.org,2002:" + tag.slice(2);
          }
          type = schema.compiledTypeMap["fallback"][tag];
          if (type && _hasOwnProperty.call(type.styleAliases, style)) {
            style = type.styleAliases[style];
          }
          result[tag] = style;
        }
        return result;
      }
      function encodeHex(character) {
        var string, handle, length;
        string = character.toString(16).toUpperCase();
        if (character <= 255) {
          handle = "x";
          length = 2;
        } else if (character <= 65535) {
          handle = "u";
          length = 4;
        } else if (character <= 4294967295) {
          handle = "U";
          length = 8;
        } else {
          throw new YAMLException("code point within a string may not be greater than 0xFFFFFFFF");
        }
        return "\\" + handle + common.repeat("0", length - string.length) + string;
      }
      var QUOTING_TYPE_SINGLE = 1;
      var QUOTING_TYPE_DOUBLE = 2;
      function State(options) {
        this.schema = options["schema"] || DEFAULT_SCHEMA;
        this.indent = Math.max(1, options["indent"] || 2);
        this.noArrayIndent = options["noArrayIndent"] || false;
        this.skipInvalid = options["skipInvalid"] || false;
        this.flowLevel = common.isNothing(options["flowLevel"]) ? -1 : options["flowLevel"];
        this.styleMap = compileStyleMap(this.schema, options["styles"] || null);
        this.sortKeys = options["sortKeys"] || false;
        this.lineWidth = options["lineWidth"] || 80;
        this.noRefs = options["noRefs"] || false;
        this.noCompatMode = options["noCompatMode"] || false;
        this.condenseFlow = options["condenseFlow"] || false;
        this.quotingType = options["quotingType"] === '"' ? QUOTING_TYPE_DOUBLE : QUOTING_TYPE_SINGLE;
        this.forceQuotes = options["forceQuotes"] || false;
        this.replacer = typeof options["replacer"] === "function" ? options["replacer"] : null;
        this.implicitTypes = this.schema.compiledImplicit;
        this.explicitTypes = this.schema.compiledExplicit;
        this.tag = null;
        this.result = "";
        this.duplicates = [];
        this.usedDuplicates = null;
      }
      function indentString(string, spaces) {
        var ind = common.repeat(" ", spaces), position = 0, next = -1, result = "", line, length = string.length;
        while (position < length) {
          next = string.indexOf("\n", position);
          if (next === -1) {
            line = string.slice(position);
            position = length;
          } else {
            line = string.slice(position, next + 1);
            position = next + 1;
          }
          if (line.length && line !== "\n")
            result += ind;
          result += line;
        }
        return result;
      }
      function generateNextLine(state, level) {
        return "\n" + common.repeat(" ", state.indent * level);
      }
      function testImplicitResolving(state, str) {
        var index, length, type;
        for (index = 0, length = state.implicitTypes.length; index < length; index += 1) {
          type = state.implicitTypes[index];
          if (type.resolve(str)) {
            return true;
          }
        }
        return false;
      }
      function isWhitespace(c) {
        return c === CHAR_SPACE || c === CHAR_TAB;
      }
      function isPrintable(c) {
        return 32 <= c && c <= 126 || 161 <= c && c <= 55295 && c !== 8232 && c !== 8233 || 57344 <= c && c <= 65533 && c !== CHAR_BOM || 65536 <= c && c <= 1114111;
      }
      function isNsCharOrWhitespace(c) {
        return isPrintable(c) && c !== CHAR_BOM && c !== CHAR_CARRIAGE_RETURN && c !== CHAR_LINE_FEED;
      }
      function isPlainSafe(c, prev, inblock) {
        var cIsNsCharOrWhitespace = isNsCharOrWhitespace(c);
        var cIsNsChar = cIsNsCharOrWhitespace && !isWhitespace(c);
        return (
          // ns-plain-safe
          (inblock ? (
            // c = flow-in
            cIsNsCharOrWhitespace
          ) : cIsNsCharOrWhitespace && c !== CHAR_COMMA && c !== CHAR_LEFT_SQUARE_BRACKET && c !== CHAR_RIGHT_SQUARE_BRACKET && c !== CHAR_LEFT_CURLY_BRACKET && c !== CHAR_RIGHT_CURLY_BRACKET) && c !== CHAR_SHARP && !(prev === CHAR_COLON && !cIsNsChar) || isNsCharOrWhitespace(prev) && !isWhitespace(prev) && c === CHAR_SHARP || prev === CHAR_COLON && cIsNsChar
        );
      }
      function isPlainSafeFirst(c) {
        return isPrintable(c) && c !== CHAR_BOM && !isWhitespace(c) && c !== CHAR_MINUS && c !== CHAR_QUESTION && c !== CHAR_COLON && c !== CHAR_COMMA && c !== CHAR_LEFT_SQUARE_BRACKET && c !== CHAR_RIGHT_SQUARE_BRACKET && c !== CHAR_LEFT_CURLY_BRACKET && c !== CHAR_RIGHT_CURLY_BRACKET && c !== CHAR_SHARP && c !== CHAR_AMPERSAND && c !== CHAR_ASTERISK && c !== CHAR_EXCLAMATION && c !== CHAR_VERTICAL_LINE && c !== CHAR_EQUALS && c !== CHAR_GREATER_THAN && c !== CHAR_SINGLE_QUOTE && c !== CHAR_DOUBLE_QUOTE && c !== CHAR_PERCENT && c !== CHAR_COMMERCIAL_AT && c !== CHAR_GRAVE_ACCENT;
      }
      function isPlainSafeLast(c) {
        return !isWhitespace(c) && c !== CHAR_COLON;
      }
      function codePointAt(string, pos) {
        var first = string.charCodeAt(pos), second;
        if (first >= 55296 && first <= 56319 && pos + 1 < string.length) {
          second = string.charCodeAt(pos + 1);
          if (second >= 56320 && second <= 57343) {
            return (first - 55296) * 1024 + second - 56320 + 65536;
          }
        }
        return first;
      }
      function needIndentIndicator(string) {
        var leadingSpaceRe = /^\n* /;
        return leadingSpaceRe.test(string);
      }
      var STYLE_PLAIN = 1;
      var STYLE_SINGLE = 2;
      var STYLE_LITERAL = 3;
      var STYLE_FOLDED = 4;
      var STYLE_DOUBLE = 5;
      function chooseScalarStyle(string, singleLineOnly, indentPerLevel, lineWidth, testAmbiguousType, quotingType, forceQuotes, inblock) {
        var i;
        var char = 0;
        var prevChar = null;
        var hasLineBreak = false;
        var hasFoldableLine = false;
        var shouldTrackWidth = lineWidth !== -1;
        var previousLineBreak = -1;
        var plain = isPlainSafeFirst(codePointAt(string, 0)) && isPlainSafeLast(codePointAt(string, string.length - 1));
        if (singleLineOnly || forceQuotes) {
          for (i = 0; i < string.length; char >= 65536 ? i += 2 : i++) {
            char = codePointAt(string, i);
            if (!isPrintable(char)) {
              return STYLE_DOUBLE;
            }
            plain = plain && isPlainSafe(char, prevChar, inblock);
            prevChar = char;
          }
        } else {
          for (i = 0; i < string.length; char >= 65536 ? i += 2 : i++) {
            char = codePointAt(string, i);
            if (char === CHAR_LINE_FEED) {
              hasLineBreak = true;
              if (shouldTrackWidth) {
                hasFoldableLine = hasFoldableLine || // Foldable line = too long, and not more-indented.
                i - previousLineBreak - 1 > lineWidth && string[previousLineBreak + 1] !== " ";
                previousLineBreak = i;
              }
            } else if (!isPrintable(char)) {
              return STYLE_DOUBLE;
            }
            plain = plain && isPlainSafe(char, prevChar, inblock);
            prevChar = char;
          }
          hasFoldableLine = hasFoldableLine || shouldTrackWidth && (i - previousLineBreak - 1 > lineWidth && string[previousLineBreak + 1] !== " ");
        }
        if (!hasLineBreak && !hasFoldableLine) {
          if (plain && !forceQuotes && !testAmbiguousType(string)) {
            return STYLE_PLAIN;
          }
          return quotingType === QUOTING_TYPE_DOUBLE ? STYLE_DOUBLE : STYLE_SINGLE;
        }
        if (indentPerLevel > 9 && needIndentIndicator(string)) {
          return STYLE_DOUBLE;
        }
        if (!forceQuotes) {
          return hasFoldableLine ? STYLE_FOLDED : STYLE_LITERAL;
        }
        return quotingType === QUOTING_TYPE_DOUBLE ? STYLE_DOUBLE : STYLE_SINGLE;
      }
      function writeScalar(state, string, level, iskey, inblock) {
        state.dump = function() {
          if (string.length === 0) {
            return state.quotingType === QUOTING_TYPE_DOUBLE ? '""' : "''";
          }
          if (!state.noCompatMode) {
            if (DEPRECATED_BOOLEANS_SYNTAX.indexOf(string) !== -1 || DEPRECATED_BASE60_SYNTAX.test(string)) {
              return state.quotingType === QUOTING_TYPE_DOUBLE ? '"' + string + '"' : "'" + string + "'";
            }
          }
          var indent = state.indent * Math.max(1, level);
          var lineWidth = state.lineWidth === -1 ? -1 : Math.max(Math.min(state.lineWidth, 40), state.lineWidth - indent);
          var singleLineOnly = iskey || state.flowLevel > -1 && level >= state.flowLevel;
          function testAmbiguity(string2) {
            return testImplicitResolving(state, string2);
          }
          switch (chooseScalarStyle(
            string,
            singleLineOnly,
            state.indent,
            lineWidth,
            testAmbiguity,
            state.quotingType,
            state.forceQuotes && !iskey,
            inblock
          )) {
            case STYLE_PLAIN:
              return string;
            case STYLE_SINGLE:
              return "'" + string.replace(/'/g, "''") + "'";
            case STYLE_LITERAL:
              return "|" + blockHeader(string, state.indent) + dropEndingNewline(indentString(string, indent));
            case STYLE_FOLDED:
              return ">" + blockHeader(string, state.indent) + dropEndingNewline(indentString(foldString(string, lineWidth), indent));
            case STYLE_DOUBLE:
              return '"' + escapeString(string, lineWidth) + '"';
            default:
              throw new YAMLException("impossible error: invalid scalar style");
          }
        }();
      }
      function blockHeader(string, indentPerLevel) {
        var indentIndicator = needIndentIndicator(string) ? String(indentPerLevel) : "";
        var clip = string[string.length - 1] === "\n";
        var keep = clip && (string[string.length - 2] === "\n" || string === "\n");
        var chomp = keep ? "+" : clip ? "" : "-";
        return indentIndicator + chomp + "\n";
      }
      function dropEndingNewline(string) {
        return string[string.length - 1] === "\n" ? string.slice(0, -1) : string;
      }
      function foldString(string, width) {
        var lineRe = /(\n+)([^\n]*)/g;
        var result = function() {
          var nextLF = string.indexOf("\n");
          nextLF = nextLF !== -1 ? nextLF : string.length;
          lineRe.lastIndex = nextLF;
          return foldLine(string.slice(0, nextLF), width);
        }();
        var prevMoreIndented = string[0] === "\n" || string[0] === " ";
        var moreIndented;
        var match;
        while (match = lineRe.exec(string)) {
          var prefix = match[1], line = match[2];
          moreIndented = line[0] === " ";
          result += prefix + (!prevMoreIndented && !moreIndented && line !== "" ? "\n" : "") + foldLine(line, width);
          prevMoreIndented = moreIndented;
        }
        return result;
      }
      function foldLine(line, width) {
        if (line === "" || line[0] === " ")
          return line;
        var breakRe = / [^ ]/g;
        var match;
        var start = 0, end, curr = 0, next = 0;
        var result = "";
        while (match = breakRe.exec(line)) {
          next = match.index;
          if (next - start > width) {
            end = curr > start ? curr : next;
            result += "\n" + line.slice(start, end);
            start = end + 1;
          }
          curr = next;
        }
        result += "\n";
        if (line.length - start > width && curr > start) {
          result += line.slice(start, curr) + "\n" + line.slice(curr + 1);
        } else {
          result += line.slice(start);
        }
        return result.slice(1);
      }
      function escapeString(string) {
        var result = "";
        var char = 0;
        var escapeSeq;
        for (var i = 0; i < string.length; char >= 65536 ? i += 2 : i++) {
          char = codePointAt(string, i);
          escapeSeq = ESCAPE_SEQUENCES[char];
          if (!escapeSeq && isPrintable(char)) {
            result += string[i];
            if (char >= 65536)
              result += string[i + 1];
          } else {
            result += escapeSeq || encodeHex(char);
          }
        }
        return result;
      }
      function writeFlowSequence(state, level, object) {
        var _result = "", _tag = state.tag, index, length, value;
        for (index = 0, length = object.length; index < length; index += 1) {
          value = object[index];
          if (state.replacer) {
            value = state.replacer.call(object, String(index), value);
          }
          if (writeNode(state, level, value, false, false) || typeof value === "undefined" && writeNode(state, level, null, false, false)) {
            if (_result !== "")
              _result += "," + (!state.condenseFlow ? " " : "");
            _result += state.dump;
          }
        }
        state.tag = _tag;
        state.dump = "[" + _result + "]";
      }
      function writeBlockSequence(state, level, object, compact) {
        var _result = "", _tag = state.tag, index, length, value;
        for (index = 0, length = object.length; index < length; index += 1) {
          value = object[index];
          if (state.replacer) {
            value = state.replacer.call(object, String(index), value);
          }
          if (writeNode(state, level + 1, value, true, true, false, true) || typeof value === "undefined" && writeNode(state, level + 1, null, true, true, false, true)) {
            if (!compact || _result !== "") {
              _result += generateNextLine(state, level);
            }
            if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
              _result += "-";
            } else {
              _result += "- ";
            }
            _result += state.dump;
          }
        }
        state.tag = _tag;
        state.dump = _result || "[]";
      }
      function writeFlowMapping(state, level, object) {
        var _result = "", _tag = state.tag, objectKeyList = Object.keys(object), index, length, objectKey, objectValue, pairBuffer;
        for (index = 0, length = objectKeyList.length; index < length; index += 1) {
          pairBuffer = "";
          if (_result !== "")
            pairBuffer += ", ";
          if (state.condenseFlow)
            pairBuffer += '"';
          objectKey = objectKeyList[index];
          objectValue = object[objectKey];
          if (state.replacer) {
            objectValue = state.replacer.call(object, objectKey, objectValue);
          }
          if (!writeNode(state, level, objectKey, false, false)) {
            continue;
          }
          if (state.dump.length > 1024)
            pairBuffer += "? ";
          pairBuffer += state.dump + (state.condenseFlow ? '"' : "") + ":" + (state.condenseFlow ? "" : " ");
          if (!writeNode(state, level, objectValue, false, false)) {
            continue;
          }
          pairBuffer += state.dump;
          _result += pairBuffer;
        }
        state.tag = _tag;
        state.dump = "{" + _result + "}";
      }
      function writeBlockMapping(state, level, object, compact) {
        var _result = "", _tag = state.tag, objectKeyList = Object.keys(object), index, length, objectKey, objectValue, explicitPair, pairBuffer;
        if (state.sortKeys === true) {
          objectKeyList.sort();
        } else if (typeof state.sortKeys === "function") {
          objectKeyList.sort(state.sortKeys);
        } else if (state.sortKeys) {
          throw new YAMLException("sortKeys must be a boolean or a function");
        }
        for (index = 0, length = objectKeyList.length; index < length; index += 1) {
          pairBuffer = "";
          if (!compact || _result !== "") {
            pairBuffer += generateNextLine(state, level);
          }
          objectKey = objectKeyList[index];
          objectValue = object[objectKey];
          if (state.replacer) {
            objectValue = state.replacer.call(object, objectKey, objectValue);
          }
          if (!writeNode(state, level + 1, objectKey, true, true, true)) {
            continue;
          }
          explicitPair = state.tag !== null && state.tag !== "?" || state.dump && state.dump.length > 1024;
          if (explicitPair) {
            if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
              pairBuffer += "?";
            } else {
              pairBuffer += "? ";
            }
          }
          pairBuffer += state.dump;
          if (explicitPair) {
            pairBuffer += generateNextLine(state, level);
          }
          if (!writeNode(state, level + 1, objectValue, true, explicitPair)) {
            continue;
          }
          if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
            pairBuffer += ":";
          } else {
            pairBuffer += ": ";
          }
          pairBuffer += state.dump;
          _result += pairBuffer;
        }
        state.tag = _tag;
        state.dump = _result || "{}";
      }
      function detectType(state, object, explicit) {
        var _result, typeList, index, length, type, style;
        typeList = explicit ? state.explicitTypes : state.implicitTypes;
        for (index = 0, length = typeList.length; index < length; index += 1) {
          type = typeList[index];
          if ((type.instanceOf || type.predicate) && (!type.instanceOf || typeof object === "object" && object instanceof type.instanceOf) && (!type.predicate || type.predicate(object))) {
            if (explicit) {
              if (type.multi && type.representName) {
                state.tag = type.representName(object);
              } else {
                state.tag = type.tag;
              }
            } else {
              state.tag = "?";
            }
            if (type.represent) {
              style = state.styleMap[type.tag] || type.defaultStyle;
              if (_toString.call(type.represent) === "[object Function]") {
                _result = type.represent(object, style);
              } else if (_hasOwnProperty.call(type.represent, style)) {
                _result = type.represent[style](object, style);
              } else {
                throw new YAMLException("!<" + type.tag + '> tag resolver accepts not "' + style + '" style');
              }
              state.dump = _result;
            }
            return true;
          }
        }
        return false;
      }
      function writeNode(state, level, object, block, compact, iskey, isblockseq) {
        state.tag = null;
        state.dump = object;
        if (!detectType(state, object, false)) {
          detectType(state, object, true);
        }
        var type = _toString.call(state.dump);
        var inblock = block;
        var tagStr;
        if (block) {
          block = state.flowLevel < 0 || state.flowLevel > level;
        }
        var objectOrArray = type === "[object Object]" || type === "[object Array]", duplicateIndex, duplicate;
        if (objectOrArray) {
          duplicateIndex = state.duplicates.indexOf(object);
          duplicate = duplicateIndex !== -1;
        }
        if (state.tag !== null && state.tag !== "?" || duplicate || state.indent !== 2 && level > 0) {
          compact = false;
        }
        if (duplicate && state.usedDuplicates[duplicateIndex]) {
          state.dump = "*ref_" + duplicateIndex;
        } else {
          if (objectOrArray && duplicate && !state.usedDuplicates[duplicateIndex]) {
            state.usedDuplicates[duplicateIndex] = true;
          }
          if (type === "[object Object]") {
            if (block && Object.keys(state.dump).length !== 0) {
              writeBlockMapping(state, level, state.dump, compact);
              if (duplicate) {
                state.dump = "&ref_" + duplicateIndex + state.dump;
              }
            } else {
              writeFlowMapping(state, level, state.dump);
              if (duplicate) {
                state.dump = "&ref_" + duplicateIndex + " " + state.dump;
              }
            }
          } else if (type === "[object Array]") {
            if (block && state.dump.length !== 0) {
              if (state.noArrayIndent && !isblockseq && level > 0) {
                writeBlockSequence(state, level - 1, state.dump, compact);
              } else {
                writeBlockSequence(state, level, state.dump, compact);
              }
              if (duplicate) {
                state.dump = "&ref_" + duplicateIndex + state.dump;
              }
            } else {
              writeFlowSequence(state, level, state.dump);
              if (duplicate) {
                state.dump = "&ref_" + duplicateIndex + " " + state.dump;
              }
            }
          } else if (type === "[object String]") {
            if (state.tag !== "?") {
              writeScalar(state, state.dump, level, iskey, inblock);
            }
          } else if (type === "[object Undefined]") {
            return false;
          } else {
            if (state.skipInvalid)
              return false;
            throw new YAMLException("unacceptable kind of an object to dump " + type);
          }
          if (state.tag !== null && state.tag !== "?") {
            tagStr = encodeURI(
              state.tag[0] === "!" ? state.tag.slice(1) : state.tag
            ).replace(/!/g, "%21");
            if (state.tag[0] === "!") {
              tagStr = "!" + tagStr;
            } else if (tagStr.slice(0, 18) === "tag:yaml.org,2002:") {
              tagStr = "!!" + tagStr.slice(18);
            } else {
              tagStr = "!<" + tagStr + ">";
            }
            state.dump = tagStr + " " + state.dump;
          }
        }
        return true;
      }
      function getDuplicateReferences(object, state) {
        var objects = [], duplicatesIndexes = [], index, length;
        inspectNode(object, objects, duplicatesIndexes);
        for (index = 0, length = duplicatesIndexes.length; index < length; index += 1) {
          state.duplicates.push(objects[duplicatesIndexes[index]]);
        }
        state.usedDuplicates = new Array(length);
      }
      function inspectNode(object, objects, duplicatesIndexes) {
        var objectKeyList, index, length;
        if (object !== null && typeof object === "object") {
          index = objects.indexOf(object);
          if (index !== -1) {
            if (duplicatesIndexes.indexOf(index) === -1) {
              duplicatesIndexes.push(index);
            }
          } else {
            objects.push(object);
            if (Array.isArray(object)) {
              for (index = 0, length = object.length; index < length; index += 1) {
                inspectNode(object[index], objects, duplicatesIndexes);
              }
            } else {
              objectKeyList = Object.keys(object);
              for (index = 0, length = objectKeyList.length; index < length; index += 1) {
                inspectNode(object[objectKeyList[index]], objects, duplicatesIndexes);
              }
            }
          }
        }
      }
      function dump(input, options) {
        options = options || {};
        var state = new State(options);
        if (!state.noRefs)
          getDuplicateReferences(input, state);
        var value = input;
        if (state.replacer) {
          value = state.replacer.call({ "": value }, "", value);
        }
        if (writeNode(state, 0, value, true, true))
          return state.dump + "\n";
        return "";
      }
      module.exports.dump = dump;
    }
  });

  // node_modules/js-yaml/index.js
  var require_js_yaml = __commonJS({
    "node_modules/js-yaml/index.js"(exports, module) {
      "use strict";
      var loader = require_loader();
      var dumper = require_dumper();
      function renamed(from, to) {
        return function() {
          throw new Error("Function yaml." + from + " is removed in js-yaml 4. Use yaml." + to + " instead, which is now safe by default.");
        };
      }
      module.exports.Type = require_type();
      module.exports.Schema = require_schema();
      module.exports.FAILSAFE_SCHEMA = require_failsafe();
      module.exports.JSON_SCHEMA = require_json();
      module.exports.CORE_SCHEMA = require_core();
      module.exports.DEFAULT_SCHEMA = require_default();
      module.exports.load = loader.load;
      module.exports.loadAll = loader.loadAll;
      module.exports.dump = dumper.dump;
      module.exports.YAMLException = require_exception();
      module.exports.types = {
        binary: require_binary(),
        float: require_float(),
        map: require_map(),
        null: require_null(),
        pairs: require_pairs(),
        set: require_set(),
        timestamp: require_timestamp(),
        bool: require_bool(),
        int: require_int(),
        merge: require_merge(),
        omap: require_omap(),
        seq: require_seq(),
        str: require_str()
      };
      module.exports.safeLoad = renamed("safeLoad", "load");
      module.exports.safeLoadAll = renamed("safeLoadAll", "loadAll");
      module.exports.safeDump = renamed("safeDump", "dump");
    }
  });

  // node_modules/marked/lib/marked.js
  var require_marked = __commonJS({
    "node_modules/marked/lib/marked.js"(exports, module) {
      (function(global, factory) {
        typeof exports === "object" && typeof module !== "undefined" ? module.exports = factory() : typeof define === "function" && define.amd ? define(factory) : (global = typeof globalThis !== "undefined" ? globalThis : global || self, global.marked = factory());
      })(exports, function() {
        "use strict";
        function _defineProperties(target, props) {
          for (var i = 0; i < props.length; i++) {
            var descriptor = props[i];
            descriptor.enumerable = descriptor.enumerable || false;
            descriptor.configurable = true;
            if ("value" in descriptor)
              descriptor.writable = true;
            Object.defineProperty(target, descriptor.key, descriptor);
          }
        }
        function _createClass(Constructor, protoProps, staticProps) {
          if (protoProps)
            _defineProperties(Constructor.prototype, protoProps);
          if (staticProps)
            _defineProperties(Constructor, staticProps);
          return Constructor;
        }
        function _unsupportedIterableToArray(o, minLen) {
          if (!o)
            return;
          if (typeof o === "string")
            return _arrayLikeToArray(o, minLen);
          var n = Object.prototype.toString.call(o).slice(8, -1);
          if (n === "Object" && o.constructor)
            n = o.constructor.name;
          if (n === "Map" || n === "Set")
            return Array.from(o);
          if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))
            return _arrayLikeToArray(o, minLen);
        }
        function _arrayLikeToArray(arr, len) {
          if (len == null || len > arr.length)
            len = arr.length;
          for (var i = 0, arr2 = new Array(len); i < len; i++)
            arr2[i] = arr[i];
          return arr2;
        }
        function _createForOfIteratorHelperLoose(o, allowArrayLike) {
          var it = typeof Symbol !== "undefined" && o[Symbol.iterator] || o["@@iterator"];
          if (it)
            return (it = it.call(o)).next.bind(it);
          if (Array.isArray(o) || (it = _unsupportedIterableToArray(o)) || allowArrayLike && o && typeof o.length === "number") {
            if (it)
              o = it;
            var i = 0;
            return function() {
              if (i >= o.length)
                return {
                  done: true
                };
              return {
                done: false,
                value: o[i++]
              };
            };
          }
          throw new TypeError("Invalid attempt to iterate non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
        }
        var defaults$5 = { exports: {} };
        function getDefaults$1() {
          return {
            baseUrl: null,
            breaks: false,
            extensions: null,
            gfm: true,
            headerIds: true,
            headerPrefix: "",
            highlight: null,
            langPrefix: "language-",
            mangle: true,
            pedantic: false,
            renderer: null,
            sanitize: false,
            sanitizer: null,
            silent: false,
            smartLists: false,
            smartypants: false,
            tokenizer: null,
            walkTokens: null,
            xhtml: false
          };
        }
        function changeDefaults$1(newDefaults) {
          defaults$5.exports.defaults = newDefaults;
        }
        defaults$5.exports = {
          defaults: getDefaults$1(),
          getDefaults: getDefaults$1,
          changeDefaults: changeDefaults$1
        };
        var escapeTest = /[&<>"']/;
        var escapeReplace = /[&<>"']/g;
        var escapeTestNoEncode = /[<>"']|&(?!#?\w+;)/;
        var escapeReplaceNoEncode = /[<>"']|&(?!#?\w+;)/g;
        var escapeReplacements = {
          "&": "&amp;",
          "<": "&lt;",
          ">": "&gt;",
          '"': "&quot;",
          "'": "&#39;"
        };
        var getEscapeReplacement = function getEscapeReplacement2(ch) {
          return escapeReplacements[ch];
        };
        function escape$2(html, encode) {
          if (encode) {
            if (escapeTest.test(html)) {
              return html.replace(escapeReplace, getEscapeReplacement);
            }
          } else {
            if (escapeTestNoEncode.test(html)) {
              return html.replace(escapeReplaceNoEncode, getEscapeReplacement);
            }
          }
          return html;
        }
        var unescapeTest = /&(#(?:\d+)|(?:#x[0-9A-Fa-f]+)|(?:\w+));?/ig;
        function unescape$1(html) {
          return html.replace(unescapeTest, function(_, n) {
            n = n.toLowerCase();
            if (n === "colon")
              return ":";
            if (n.charAt(0) === "#") {
              return n.charAt(1) === "x" ? String.fromCharCode(parseInt(n.substring(2), 16)) : String.fromCharCode(+n.substring(1));
            }
            return "";
          });
        }
        var caret = /(^|[^\[])\^/g;
        function edit$1(regex, opt) {
          regex = regex.source || regex;
          opt = opt || "";
          var obj = {
            replace: function replace(name, val) {
              val = val.source || val;
              val = val.replace(caret, "$1");
              regex = regex.replace(name, val);
              return obj;
            },
            getRegex: function getRegex() {
              return new RegExp(regex, opt);
            }
          };
          return obj;
        }
        var nonWordAndColonTest = /[^\w:]/g;
        var originIndependentUrl = /^$|^[a-z][a-z0-9+.-]*:|^[?#]/i;
        function cleanUrl$1(sanitize, base, href) {
          if (sanitize) {
            var prot;
            try {
              prot = decodeURIComponent(unescape$1(href)).replace(nonWordAndColonTest, "").toLowerCase();
            } catch (e) {
              return null;
            }
            if (prot.indexOf("javascript:") === 0 || prot.indexOf("vbscript:") === 0 || prot.indexOf("data:") === 0) {
              return null;
            }
          }
          if (base && !originIndependentUrl.test(href)) {
            href = resolveUrl(base, href);
          }
          try {
            href = encodeURI(href).replace(/%25/g, "%");
          } catch (e) {
            return null;
          }
          return href;
        }
        var baseUrls = {};
        var justDomain = /^[^:]+:\/*[^/]*$/;
        var protocol = /^([^:]+:)[\s\S]*$/;
        var domain = /^([^:]+:\/*[^/]*)[\s\S]*$/;
        function resolveUrl(base, href) {
          if (!baseUrls[" " + base]) {
            if (justDomain.test(base)) {
              baseUrls[" " + base] = base + "/";
            } else {
              baseUrls[" " + base] = rtrim$1(base, "/", true);
            }
          }
          base = baseUrls[" " + base];
          var relativeBase = base.indexOf(":") === -1;
          if (href.substring(0, 2) === "//") {
            if (relativeBase) {
              return href;
            }
            return base.replace(protocol, "$1") + href;
          } else if (href.charAt(0) === "/") {
            if (relativeBase) {
              return href;
            }
            return base.replace(domain, "$1") + href;
          } else {
            return base + href;
          }
        }
        var noopTest$1 = {
          exec: function noopTest2() {
          }
        };
        function merge$2(obj) {
          var i = 1, target, key;
          for (; i < arguments.length; i++) {
            target = arguments[i];
            for (key in target) {
              if (Object.prototype.hasOwnProperty.call(target, key)) {
                obj[key] = target[key];
              }
            }
          }
          return obj;
        }
        function splitCells$1(tableRow, count) {
          var row = tableRow.replace(/\|/g, function(match, offset, str) {
            var escaped = false, curr = offset;
            while (--curr >= 0 && str[curr] === "\\") {
              escaped = !escaped;
            }
            if (escaped) {
              return "|";
            } else {
              return " |";
            }
          }), cells = row.split(/ \|/);
          var i = 0;
          if (cells.length > count) {
            cells.splice(count);
          } else {
            while (cells.length < count) {
              cells.push("");
            }
          }
          for (; i < cells.length; i++) {
            cells[i] = cells[i].trim().replace(/\\\|/g, "|");
          }
          return cells;
        }
        function rtrim$1(str, c, invert) {
          var l = str.length;
          if (l === 0) {
            return "";
          }
          var suffLen = 0;
          while (suffLen < l) {
            var currChar = str.charAt(l - suffLen - 1);
            if (currChar === c && !invert) {
              suffLen++;
            } else if (currChar !== c && invert) {
              suffLen++;
            } else {
              break;
            }
          }
          return str.substr(0, l - suffLen);
        }
        function findClosingBracket$1(str, b) {
          if (str.indexOf(b[1]) === -1) {
            return -1;
          }
          var l = str.length;
          var level = 0, i = 0;
          for (; i < l; i++) {
            if (str[i] === "\\") {
              i++;
            } else if (str[i] === b[0]) {
              level++;
            } else if (str[i] === b[1]) {
              level--;
              if (level < 0) {
                return i;
              }
            }
          }
          return -1;
        }
        function checkSanitizeDeprecation$1(opt) {
          if (opt && opt.sanitize && !opt.silent) {
            console.warn("marked(): sanitize and sanitizer parameters are deprecated since version 0.7.0, should not be used and will be removed in the future. Read more here: https://marked.js.org/#/USING_ADVANCED.md#options");
          }
        }
        function repeatString$1(pattern, count) {
          if (count < 1) {
            return "";
          }
          var result = "";
          while (count > 1) {
            if (count & 1) {
              result += pattern;
            }
            count >>= 1;
            pattern += pattern;
          }
          return result + pattern;
        }
        var helpers = {
          escape: escape$2,
          unescape: unescape$1,
          edit: edit$1,
          cleanUrl: cleanUrl$1,
          resolveUrl,
          noopTest: noopTest$1,
          merge: merge$2,
          splitCells: splitCells$1,
          rtrim: rtrim$1,
          findClosingBracket: findClosingBracket$1,
          checkSanitizeDeprecation: checkSanitizeDeprecation$1,
          repeatString: repeatString$1
        };
        var defaults$4 = defaults$5.exports.defaults;
        var rtrim = helpers.rtrim, splitCells = helpers.splitCells, _escape = helpers.escape, findClosingBracket = helpers.findClosingBracket;
        function outputLink(cap, link, raw) {
          var href = link.href;
          var title = link.title ? _escape(link.title) : null;
          var text = cap[1].replace(/\\([\[\]])/g, "$1");
          if (cap[0].charAt(0) !== "!") {
            return {
              type: "link",
              raw,
              href,
              title,
              text
            };
          } else {
            return {
              type: "image",
              raw,
              href,
              title,
              text: _escape(text)
            };
          }
        }
        function indentCodeCompensation(raw, text) {
          var matchIndentToCode = raw.match(/^(\s+)(?:```)/);
          if (matchIndentToCode === null) {
            return text;
          }
          var indentToCode = matchIndentToCode[1];
          return text.split("\n").map(function(node) {
            var matchIndentInNode = node.match(/^\s+/);
            if (matchIndentInNode === null) {
              return node;
            }
            var indentInNode = matchIndentInNode[0];
            if (indentInNode.length >= indentToCode.length) {
              return node.slice(indentToCode.length);
            }
            return node;
          }).join("\n");
        }
        var Tokenizer_1 = /* @__PURE__ */ function() {
          function Tokenizer2(options) {
            this.options = options || defaults$4;
          }
          var _proto = Tokenizer2.prototype;
          _proto.space = function space(src) {
            var cap = this.rules.block.newline.exec(src);
            if (cap) {
              if (cap[0].length > 1) {
                return {
                  type: "space",
                  raw: cap[0]
                };
              }
              return {
                raw: "\n"
              };
            }
          };
          _proto.code = function code(src) {
            var cap = this.rules.block.code.exec(src);
            if (cap) {
              var text = cap[0].replace(/^ {1,4}/gm, "");
              return {
                type: "code",
                raw: cap[0],
                codeBlockStyle: "indented",
                text: !this.options.pedantic ? rtrim(text, "\n") : text
              };
            }
          };
          _proto.fences = function fences(src) {
            var cap = this.rules.block.fences.exec(src);
            if (cap) {
              var raw = cap[0];
              var text = indentCodeCompensation(raw, cap[3] || "");
              return {
                type: "code",
                raw,
                lang: cap[2] ? cap[2].trim() : cap[2],
                text
              };
            }
          };
          _proto.heading = function heading(src) {
            var cap = this.rules.block.heading.exec(src);
            if (cap) {
              var text = cap[2].trim();
              if (/#$/.test(text)) {
                var trimmed = rtrim(text, "#");
                if (this.options.pedantic) {
                  text = trimmed.trim();
                } else if (!trimmed || / $/.test(trimmed)) {
                  text = trimmed.trim();
                }
              }
              return {
                type: "heading",
                raw: cap[0],
                depth: cap[1].length,
                text
              };
            }
          };
          _proto.nptable = function nptable(src) {
            var cap = this.rules.block.nptable.exec(src);
            if (cap) {
              var item = {
                type: "table",
                header: splitCells(cap[1].replace(/^ *| *\| *$/g, "")),
                align: cap[2].replace(/^ *|\| *$/g, "").split(/ *\| */),
                cells: cap[3] ? cap[3].replace(/\n$/, "").split("\n") : [],
                raw: cap[0]
              };
              if (item.header.length === item.align.length) {
                var l = item.align.length;
                var i;
                for (i = 0; i < l; i++) {
                  if (/^ *-+: *$/.test(item.align[i])) {
                    item.align[i] = "right";
                  } else if (/^ *:-+: *$/.test(item.align[i])) {
                    item.align[i] = "center";
                  } else if (/^ *:-+ *$/.test(item.align[i])) {
                    item.align[i] = "left";
                  } else {
                    item.align[i] = null;
                  }
                }
                l = item.cells.length;
                for (i = 0; i < l; i++) {
                  item.cells[i] = splitCells(item.cells[i], item.header.length);
                }
                return item;
              }
            }
          };
          _proto.hr = function hr(src) {
            var cap = this.rules.block.hr.exec(src);
            if (cap) {
              return {
                type: "hr",
                raw: cap[0]
              };
            }
          };
          _proto.blockquote = function blockquote(src) {
            var cap = this.rules.block.blockquote.exec(src);
            if (cap) {
              var text = cap[0].replace(/^ *> ?/gm, "");
              return {
                type: "blockquote",
                raw: cap[0],
                text
              };
            }
          };
          _proto.list = function list(src) {
            var cap = this.rules.block.list.exec(src);
            if (cap) {
              var raw = cap[0];
              var bull = cap[2];
              var isordered = bull.length > 1;
              var list2 = {
                type: "list",
                raw,
                ordered: isordered,
                start: isordered ? +bull.slice(0, -1) : "",
                loose: false,
                items: []
              };
              var itemMatch = cap[0].match(this.rules.block.item);
              var next = false, item, space, bcurr, bnext, addBack, loose, istask, ischecked, endMatch;
              var l = itemMatch.length;
              bcurr = this.rules.block.listItemStart.exec(itemMatch[0]);
              for (var i = 0; i < l; i++) {
                item = itemMatch[i];
                raw = item;
                if (!this.options.pedantic) {
                  endMatch = item.match(new RegExp("\\n\\s*\\n {0," + (bcurr[0].length - 1) + "}\\S"));
                  if (endMatch) {
                    addBack = item.length - endMatch.index + itemMatch.slice(i + 1).join("\n").length;
                    list2.raw = list2.raw.substring(0, list2.raw.length - addBack);
                    item = item.substring(0, endMatch.index);
                    raw = item;
                    l = i + 1;
                  }
                }
                if (i !== l - 1) {
                  bnext = this.rules.block.listItemStart.exec(itemMatch[i + 1]);
                  if (!this.options.pedantic ? bnext[1].length >= bcurr[0].length || bnext[1].length > 3 : bnext[1].length > bcurr[1].length) {
                    itemMatch.splice(i, 2, itemMatch[i] + (!this.options.pedantic && bnext[1].length < bcurr[0].length && !itemMatch[i].match(/\n$/) ? "" : "\n") + itemMatch[i + 1]);
                    i--;
                    l--;
                    continue;
                  } else if (
                    // different bullet style
                    !this.options.pedantic || this.options.smartLists ? bnext[2][bnext[2].length - 1] !== bull[bull.length - 1] : isordered === (bnext[2].length === 1)
                  ) {
                    addBack = itemMatch.slice(i + 1).join("\n").length;
                    list2.raw = list2.raw.substring(0, list2.raw.length - addBack);
                    i = l - 1;
                  }
                  bcurr = bnext;
                }
                space = item.length;
                item = item.replace(/^ *([*+-]|\d+[.)]) ?/, "");
                if (~item.indexOf("\n ")) {
                  space -= item.length;
                  item = !this.options.pedantic ? item.replace(new RegExp("^ {1," + space + "}", "gm"), "") : item.replace(/^ {1,4}/gm, "");
                }
                item = rtrim(item, "\n");
                if (i !== l - 1) {
                  raw = raw + "\n";
                }
                loose = next || /\n\n(?!\s*$)/.test(raw);
                if (i !== l - 1) {
                  next = raw.slice(-2) === "\n\n";
                  if (!loose)
                    loose = next;
                }
                if (loose) {
                  list2.loose = true;
                }
                if (this.options.gfm) {
                  istask = /^\[[ xX]\] /.test(item);
                  ischecked = void 0;
                  if (istask) {
                    ischecked = item[1] !== " ";
                    item = item.replace(/^\[[ xX]\] +/, "");
                  }
                }
                list2.items.push({
                  type: "list_item",
                  raw,
                  task: istask,
                  checked: ischecked,
                  loose,
                  text: item
                });
              }
              return list2;
            }
          };
          _proto.html = function html(src) {
            var cap = this.rules.block.html.exec(src);
            if (cap) {
              return {
                type: this.options.sanitize ? "paragraph" : "html",
                raw: cap[0],
                pre: !this.options.sanitizer && (cap[1] === "pre" || cap[1] === "script" || cap[1] === "style"),
                text: this.options.sanitize ? this.options.sanitizer ? this.options.sanitizer(cap[0]) : _escape(cap[0]) : cap[0]
              };
            }
          };
          _proto.def = function def(src) {
            var cap = this.rules.block.def.exec(src);
            if (cap) {
              if (cap[3])
                cap[3] = cap[3].substring(1, cap[3].length - 1);
              var tag = cap[1].toLowerCase().replace(/\s+/g, " ");
              return {
                type: "def",
                tag,
                raw: cap[0],
                href: cap[2],
                title: cap[3]
              };
            }
          };
          _proto.table = function table(src) {
            var cap = this.rules.block.table.exec(src);
            if (cap) {
              var item = {
                type: "table",
                header: splitCells(cap[1].replace(/^ *| *\| *$/g, "")),
                align: cap[2].replace(/^ *|\| *$/g, "").split(/ *\| */),
                cells: cap[3] ? cap[3].replace(/\n$/, "").split("\n") : []
              };
              if (item.header.length === item.align.length) {
                item.raw = cap[0];
                var l = item.align.length;
                var i;
                for (i = 0; i < l; i++) {
                  if (/^ *-+: *$/.test(item.align[i])) {
                    item.align[i] = "right";
                  } else if (/^ *:-+: *$/.test(item.align[i])) {
                    item.align[i] = "center";
                  } else if (/^ *:-+ *$/.test(item.align[i])) {
                    item.align[i] = "left";
                  } else {
                    item.align[i] = null;
                  }
                }
                l = item.cells.length;
                for (i = 0; i < l; i++) {
                  item.cells[i] = splitCells(item.cells[i].replace(/^ *\| *| *\| *$/g, ""), item.header.length);
                }
                return item;
              }
            }
          };
          _proto.lheading = function lheading(src) {
            var cap = this.rules.block.lheading.exec(src);
            if (cap) {
              return {
                type: "heading",
                raw: cap[0],
                depth: cap[2].charAt(0) === "=" ? 1 : 2,
                text: cap[1]
              };
            }
          };
          _proto.paragraph = function paragraph(src) {
            var cap = this.rules.block.paragraph.exec(src);
            if (cap) {
              return {
                type: "paragraph",
                raw: cap[0],
                text: cap[1].charAt(cap[1].length - 1) === "\n" ? cap[1].slice(0, -1) : cap[1]
              };
            }
          };
          _proto.text = function text(src) {
            var cap = this.rules.block.text.exec(src);
            if (cap) {
              return {
                type: "text",
                raw: cap[0],
                text: cap[0]
              };
            }
          };
          _proto.escape = function escape2(src) {
            var cap = this.rules.inline.escape.exec(src);
            if (cap) {
              return {
                type: "escape",
                raw: cap[0],
                text: _escape(cap[1])
              };
            }
          };
          _proto.tag = function tag(src, inLink, inRawBlock) {
            var cap = this.rules.inline.tag.exec(src);
            if (cap) {
              if (!inLink && /^<a /i.test(cap[0])) {
                inLink = true;
              } else if (inLink && /^<\/a>/i.test(cap[0])) {
                inLink = false;
              }
              if (!inRawBlock && /^<(pre|code|kbd|script)(\s|>)/i.test(cap[0])) {
                inRawBlock = true;
              } else if (inRawBlock && /^<\/(pre|code|kbd|script)(\s|>)/i.test(cap[0])) {
                inRawBlock = false;
              }
              return {
                type: this.options.sanitize ? "text" : "html",
                raw: cap[0],
                inLink,
                inRawBlock,
                text: this.options.sanitize ? this.options.sanitizer ? this.options.sanitizer(cap[0]) : _escape(cap[0]) : cap[0]
              };
            }
          };
          _proto.link = function link(src) {
            var cap = this.rules.inline.link.exec(src);
            if (cap) {
              var trimmedUrl = cap[2].trim();
              if (!this.options.pedantic && /^</.test(trimmedUrl)) {
                if (!/>$/.test(trimmedUrl)) {
                  return;
                }
                var rtrimSlash = rtrim(trimmedUrl.slice(0, -1), "\\");
                if ((trimmedUrl.length - rtrimSlash.length) % 2 === 0) {
                  return;
                }
              } else {
                var lastParenIndex = findClosingBracket(cap[2], "()");
                if (lastParenIndex > -1) {
                  var start = cap[0].indexOf("!") === 0 ? 5 : 4;
                  var linkLen = start + cap[1].length + lastParenIndex;
                  cap[2] = cap[2].substring(0, lastParenIndex);
                  cap[0] = cap[0].substring(0, linkLen).trim();
                  cap[3] = "";
                }
              }
              var href = cap[2];
              var title = "";
              if (this.options.pedantic) {
                var link2 = /^([^'"]*[^\s])\s+(['"])(.*)\2/.exec(href);
                if (link2) {
                  href = link2[1];
                  title = link2[3];
                }
              } else {
                title = cap[3] ? cap[3].slice(1, -1) : "";
              }
              href = href.trim();
              if (/^</.test(href)) {
                if (this.options.pedantic && !/>$/.test(trimmedUrl)) {
                  href = href.slice(1);
                } else {
                  href = href.slice(1, -1);
                }
              }
              return outputLink(cap, {
                href: href ? href.replace(this.rules.inline._escapes, "$1") : href,
                title: title ? title.replace(this.rules.inline._escapes, "$1") : title
              }, cap[0]);
            }
          };
          _proto.reflink = function reflink(src, links) {
            var cap;
            if ((cap = this.rules.inline.reflink.exec(src)) || (cap = this.rules.inline.nolink.exec(src))) {
              var link = (cap[2] || cap[1]).replace(/\s+/g, " ");
              link = links[link.toLowerCase()];
              if (!link || !link.href) {
                var text = cap[0].charAt(0);
                return {
                  type: "text",
                  raw: text,
                  text
                };
              }
              return outputLink(cap, link, cap[0]);
            }
          };
          _proto.emStrong = function emStrong(src, maskedSrc, prevChar) {
            if (prevChar === void 0) {
              prevChar = "";
            }
            var match = this.rules.inline.emStrong.lDelim.exec(src);
            if (!match)
              return;
            if (match[3] && prevChar.match(/(?:[0-9A-Za-z\xAA\xB2\xB3\xB5\xB9\xBA\xBC-\xBE\xC0-\xD6\xD8-\xF6\xF8-\u02C1\u02C6-\u02D1\u02E0-\u02E4\u02EC\u02EE\u0370-\u0374\u0376\u0377\u037A-\u037D\u037F\u0386\u0388-\u038A\u038C\u038E-\u03A1\u03A3-\u03F5\u03F7-\u0481\u048A-\u052F\u0531-\u0556\u0559\u0560-\u0588\u05D0-\u05EA\u05EF-\u05F2\u0620-\u064A\u0660-\u0669\u066E\u066F\u0671-\u06D3\u06D5\u06E5\u06E6\u06EE-\u06FC\u06FF\u0710\u0712-\u072F\u074D-\u07A5\u07B1\u07C0-\u07EA\u07F4\u07F5\u07FA\u0800-\u0815\u081A\u0824\u0828\u0840-\u0858\u0860-\u086A\u08A0-\u08B4\u08B6-\u08C7\u0904-\u0939\u093D\u0950\u0958-\u0961\u0966-\u096F\u0971-\u0980\u0985-\u098C\u098F\u0990\u0993-\u09A8\u09AA-\u09B0\u09B2\u09B6-\u09B9\u09BD\u09CE\u09DC\u09DD\u09DF-\u09E1\u09E6-\u09F1\u09F4-\u09F9\u09FC\u0A05-\u0A0A\u0A0F\u0A10\u0A13-\u0A28\u0A2A-\u0A30\u0A32\u0A33\u0A35\u0A36\u0A38\u0A39\u0A59-\u0A5C\u0A5E\u0A66-\u0A6F\u0A72-\u0A74\u0A85-\u0A8D\u0A8F-\u0A91\u0A93-\u0AA8\u0AAA-\u0AB0\u0AB2\u0AB3\u0AB5-\u0AB9\u0ABD\u0AD0\u0AE0\u0AE1\u0AE6-\u0AEF\u0AF9\u0B05-\u0B0C\u0B0F\u0B10\u0B13-\u0B28\u0B2A-\u0B30\u0B32\u0B33\u0B35-\u0B39\u0B3D\u0B5C\u0B5D\u0B5F-\u0B61\u0B66-\u0B6F\u0B71-\u0B77\u0B83\u0B85-\u0B8A\u0B8E-\u0B90\u0B92-\u0B95\u0B99\u0B9A\u0B9C\u0B9E\u0B9F\u0BA3\u0BA4\u0BA8-\u0BAA\u0BAE-\u0BB9\u0BD0\u0BE6-\u0BF2\u0C05-\u0C0C\u0C0E-\u0C10\u0C12-\u0C28\u0C2A-\u0C39\u0C3D\u0C58-\u0C5A\u0C60\u0C61\u0C66-\u0C6F\u0C78-\u0C7E\u0C80\u0C85-\u0C8C\u0C8E-\u0C90\u0C92-\u0CA8\u0CAA-\u0CB3\u0CB5-\u0CB9\u0CBD\u0CDE\u0CE0\u0CE1\u0CE6-\u0CEF\u0CF1\u0CF2\u0D04-\u0D0C\u0D0E-\u0D10\u0D12-\u0D3A\u0D3D\u0D4E\u0D54-\u0D56\u0D58-\u0D61\u0D66-\u0D78\u0D7A-\u0D7F\u0D85-\u0D96\u0D9A-\u0DB1\u0DB3-\u0DBB\u0DBD\u0DC0-\u0DC6\u0DE6-\u0DEF\u0E01-\u0E30\u0E32\u0E33\u0E40-\u0E46\u0E50-\u0E59\u0E81\u0E82\u0E84\u0E86-\u0E8A\u0E8C-\u0EA3\u0EA5\u0EA7-\u0EB0\u0EB2\u0EB3\u0EBD\u0EC0-\u0EC4\u0EC6\u0ED0-\u0ED9\u0EDC-\u0EDF\u0F00\u0F20-\u0F33\u0F40-\u0F47\u0F49-\u0F6C\u0F88-\u0F8C\u1000-\u102A\u103F-\u1049\u1050-\u1055\u105A-\u105D\u1061\u1065\u1066\u106E-\u1070\u1075-\u1081\u108E\u1090-\u1099\u10A0-\u10C5\u10C7\u10CD\u10D0-\u10FA\u10FC-\u1248\u124A-\u124D\u1250-\u1256\u1258\u125A-\u125D\u1260-\u1288\u128A-\u128D\u1290-\u12B0\u12B2-\u12B5\u12B8-\u12BE\u12C0\u12C2-\u12C5\u12C8-\u12D6\u12D8-\u1310\u1312-\u1315\u1318-\u135A\u1369-\u137C\u1380-\u138F\u13A0-\u13F5\u13F8-\u13FD\u1401-\u166C\u166F-\u167F\u1681-\u169A\u16A0-\u16EA\u16EE-\u16F8\u1700-\u170C\u170E-\u1711\u1720-\u1731\u1740-\u1751\u1760-\u176C\u176E-\u1770\u1780-\u17B3\u17D7\u17DC\u17E0-\u17E9\u17F0-\u17F9\u1810-\u1819\u1820-\u1878\u1880-\u1884\u1887-\u18A8\u18AA\u18B0-\u18F5\u1900-\u191E\u1946-\u196D\u1970-\u1974\u1980-\u19AB\u19B0-\u19C9\u19D0-\u19DA\u1A00-\u1A16\u1A20-\u1A54\u1A80-\u1A89\u1A90-\u1A99\u1AA7\u1B05-\u1B33\u1B45-\u1B4B\u1B50-\u1B59\u1B83-\u1BA0\u1BAE-\u1BE5\u1C00-\u1C23\u1C40-\u1C49\u1C4D-\u1C7D\u1C80-\u1C88\u1C90-\u1CBA\u1CBD-\u1CBF\u1CE9-\u1CEC\u1CEE-\u1CF3\u1CF5\u1CF6\u1CFA\u1D00-\u1DBF\u1E00-\u1F15\u1F18-\u1F1D\u1F20-\u1F45\u1F48-\u1F4D\u1F50-\u1F57\u1F59\u1F5B\u1F5D\u1F5F-\u1F7D\u1F80-\u1FB4\u1FB6-\u1FBC\u1FBE\u1FC2-\u1FC4\u1FC6-\u1FCC\u1FD0-\u1FD3\u1FD6-\u1FDB\u1FE0-\u1FEC\u1FF2-\u1FF4\u1FF6-\u1FFC\u2070\u2071\u2074-\u2079\u207F-\u2089\u2090-\u209C\u2102\u2107\u210A-\u2113\u2115\u2119-\u211D\u2124\u2126\u2128\u212A-\u212D\u212F-\u2139\u213C-\u213F\u2145-\u2149\u214E\u2150-\u2189\u2460-\u249B\u24EA-\u24FF\u2776-\u2793\u2C00-\u2C2E\u2C30-\u2C5E\u2C60-\u2CE4\u2CEB-\u2CEE\u2CF2\u2CF3\u2CFD\u2D00-\u2D25\u2D27\u2D2D\u2D30-\u2D67\u2D6F\u2D80-\u2D96\u2DA0-\u2DA6\u2DA8-\u2DAE\u2DB0-\u2DB6\u2DB8-\u2DBE\u2DC0-\u2DC6\u2DC8-\u2DCE\u2DD0-\u2DD6\u2DD8-\u2DDE\u2E2F\u3005-\u3007\u3021-\u3029\u3031-\u3035\u3038-\u303C\u3041-\u3096\u309D-\u309F\u30A1-\u30FA\u30FC-\u30FF\u3105-\u312F\u3131-\u318E\u3192-\u3195\u31A0-\u31BF\u31F0-\u31FF\u3220-\u3229\u3248-\u324F\u3251-\u325F\u3280-\u3289\u32B1-\u32BF\u3400-\u4DBF\u4E00-\u9FFC\uA000-\uA48C\uA4D0-\uA4FD\uA500-\uA60C\uA610-\uA62B\uA640-\uA66E\uA67F-\uA69D\uA6A0-\uA6EF\uA717-\uA71F\uA722-\uA788\uA78B-\uA7BF\uA7C2-\uA7CA\uA7F5-\uA801\uA803-\uA805\uA807-\uA80A\uA80C-\uA822\uA830-\uA835\uA840-\uA873\uA882-\uA8B3\uA8D0-\uA8D9\uA8F2-\uA8F7\uA8FB\uA8FD\uA8FE\uA900-\uA925\uA930-\uA946\uA960-\uA97C\uA984-\uA9B2\uA9CF-\uA9D9\uA9E0-\uA9E4\uA9E6-\uA9FE\uAA00-\uAA28\uAA40-\uAA42\uAA44-\uAA4B\uAA50-\uAA59\uAA60-\uAA76\uAA7A\uAA7E-\uAAAF\uAAB1\uAAB5\uAAB6\uAAB9-\uAABD\uAAC0\uAAC2\uAADB-\uAADD\uAAE0-\uAAEA\uAAF2-\uAAF4\uAB01-\uAB06\uAB09-\uAB0E\uAB11-\uAB16\uAB20-\uAB26\uAB28-\uAB2E\uAB30-\uAB5A\uAB5C-\uAB69\uAB70-\uABE2\uABF0-\uABF9\uAC00-\uD7A3\uD7B0-\uD7C6\uD7CB-\uD7FB\uF900-\uFA6D\uFA70-\uFAD9\uFB00-\uFB06\uFB13-\uFB17\uFB1D\uFB1F-\uFB28\uFB2A-\uFB36\uFB38-\uFB3C\uFB3E\uFB40\uFB41\uFB43\uFB44\uFB46-\uFBB1\uFBD3-\uFD3D\uFD50-\uFD8F\uFD92-\uFDC7\uFDF0-\uFDFB\uFE70-\uFE74\uFE76-\uFEFC\uFF10-\uFF19\uFF21-\uFF3A\uFF41-\uFF5A\uFF66-\uFFBE\uFFC2-\uFFC7\uFFCA-\uFFCF\uFFD2-\uFFD7\uFFDA-\uFFDC]|\uD800[\uDC00-\uDC0B\uDC0D-\uDC26\uDC28-\uDC3A\uDC3C\uDC3D\uDC3F-\uDC4D\uDC50-\uDC5D\uDC80-\uDCFA\uDD07-\uDD33\uDD40-\uDD78\uDD8A\uDD8B\uDE80-\uDE9C\uDEA0-\uDED0\uDEE1-\uDEFB\uDF00-\uDF23\uDF2D-\uDF4A\uDF50-\uDF75\uDF80-\uDF9D\uDFA0-\uDFC3\uDFC8-\uDFCF\uDFD1-\uDFD5]|\uD801[\uDC00-\uDC9D\uDCA0-\uDCA9\uDCB0-\uDCD3\uDCD8-\uDCFB\uDD00-\uDD27\uDD30-\uDD63\uDE00-\uDF36\uDF40-\uDF55\uDF60-\uDF67]|\uD802[\uDC00-\uDC05\uDC08\uDC0A-\uDC35\uDC37\uDC38\uDC3C\uDC3F-\uDC55\uDC58-\uDC76\uDC79-\uDC9E\uDCA7-\uDCAF\uDCE0-\uDCF2\uDCF4\uDCF5\uDCFB-\uDD1B\uDD20-\uDD39\uDD80-\uDDB7\uDDBC-\uDDCF\uDDD2-\uDE00\uDE10-\uDE13\uDE15-\uDE17\uDE19-\uDE35\uDE40-\uDE48\uDE60-\uDE7E\uDE80-\uDE9F\uDEC0-\uDEC7\uDEC9-\uDEE4\uDEEB-\uDEEF\uDF00-\uDF35\uDF40-\uDF55\uDF58-\uDF72\uDF78-\uDF91\uDFA9-\uDFAF]|\uD803[\uDC00-\uDC48\uDC80-\uDCB2\uDCC0-\uDCF2\uDCFA-\uDD23\uDD30-\uDD39\uDE60-\uDE7E\uDE80-\uDEA9\uDEB0\uDEB1\uDF00-\uDF27\uDF30-\uDF45\uDF51-\uDF54\uDFB0-\uDFCB\uDFE0-\uDFF6]|\uD804[\uDC03-\uDC37\uDC52-\uDC6F\uDC83-\uDCAF\uDCD0-\uDCE8\uDCF0-\uDCF9\uDD03-\uDD26\uDD36-\uDD3F\uDD44\uDD47\uDD50-\uDD72\uDD76\uDD83-\uDDB2\uDDC1-\uDDC4\uDDD0-\uDDDA\uDDDC\uDDE1-\uDDF4\uDE00-\uDE11\uDE13-\uDE2B\uDE80-\uDE86\uDE88\uDE8A-\uDE8D\uDE8F-\uDE9D\uDE9F-\uDEA8\uDEB0-\uDEDE\uDEF0-\uDEF9\uDF05-\uDF0C\uDF0F\uDF10\uDF13-\uDF28\uDF2A-\uDF30\uDF32\uDF33\uDF35-\uDF39\uDF3D\uDF50\uDF5D-\uDF61]|\uD805[\uDC00-\uDC34\uDC47-\uDC4A\uDC50-\uDC59\uDC5F-\uDC61\uDC80-\uDCAF\uDCC4\uDCC5\uDCC7\uDCD0-\uDCD9\uDD80-\uDDAE\uDDD8-\uDDDB\uDE00-\uDE2F\uDE44\uDE50-\uDE59\uDE80-\uDEAA\uDEB8\uDEC0-\uDEC9\uDF00-\uDF1A\uDF30-\uDF3B]|\uD806[\uDC00-\uDC2B\uDCA0-\uDCF2\uDCFF-\uDD06\uDD09\uDD0C-\uDD13\uDD15\uDD16\uDD18-\uDD2F\uDD3F\uDD41\uDD50-\uDD59\uDDA0-\uDDA7\uDDAA-\uDDD0\uDDE1\uDDE3\uDE00\uDE0B-\uDE32\uDE3A\uDE50\uDE5C-\uDE89\uDE9D\uDEC0-\uDEF8]|\uD807[\uDC00-\uDC08\uDC0A-\uDC2E\uDC40\uDC50-\uDC6C\uDC72-\uDC8F\uDD00-\uDD06\uDD08\uDD09\uDD0B-\uDD30\uDD46\uDD50-\uDD59\uDD60-\uDD65\uDD67\uDD68\uDD6A-\uDD89\uDD98\uDDA0-\uDDA9\uDEE0-\uDEF2\uDFB0\uDFC0-\uDFD4]|\uD808[\uDC00-\uDF99]|\uD809[\uDC00-\uDC6E\uDC80-\uDD43]|[\uD80C\uD81C-\uD820\uD822\uD840-\uD868\uD86A-\uD86C\uD86F-\uD872\uD874-\uD879\uD880-\uD883][\uDC00-\uDFFF]|\uD80D[\uDC00-\uDC2E]|\uD811[\uDC00-\uDE46]|\uD81A[\uDC00-\uDE38\uDE40-\uDE5E\uDE60-\uDE69\uDED0-\uDEED\uDF00-\uDF2F\uDF40-\uDF43\uDF50-\uDF59\uDF5B-\uDF61\uDF63-\uDF77\uDF7D-\uDF8F]|\uD81B[\uDE40-\uDE96\uDF00-\uDF4A\uDF50\uDF93-\uDF9F\uDFE0\uDFE1\uDFE3]|\uD821[\uDC00-\uDFF7]|\uD823[\uDC00-\uDCD5\uDD00-\uDD08]|\uD82C[\uDC00-\uDD1E\uDD50-\uDD52\uDD64-\uDD67\uDD70-\uDEFB]|\uD82F[\uDC00-\uDC6A\uDC70-\uDC7C\uDC80-\uDC88\uDC90-\uDC99]|\uD834[\uDEE0-\uDEF3\uDF60-\uDF78]|\uD835[\uDC00-\uDC54\uDC56-\uDC9C\uDC9E\uDC9F\uDCA2\uDCA5\uDCA6\uDCA9-\uDCAC\uDCAE-\uDCB9\uDCBB\uDCBD-\uDCC3\uDCC5-\uDD05\uDD07-\uDD0A\uDD0D-\uDD14\uDD16-\uDD1C\uDD1E-\uDD39\uDD3B-\uDD3E\uDD40-\uDD44\uDD46\uDD4A-\uDD50\uDD52-\uDEA5\uDEA8-\uDEC0\uDEC2-\uDEDA\uDEDC-\uDEFA\uDEFC-\uDF14\uDF16-\uDF34\uDF36-\uDF4E\uDF50-\uDF6E\uDF70-\uDF88\uDF8A-\uDFA8\uDFAA-\uDFC2\uDFC4-\uDFCB\uDFCE-\uDFFF]|\uD838[\uDD00-\uDD2C\uDD37-\uDD3D\uDD40-\uDD49\uDD4E\uDEC0-\uDEEB\uDEF0-\uDEF9]|\uD83A[\uDC00-\uDCC4\uDCC7-\uDCCF\uDD00-\uDD43\uDD4B\uDD50-\uDD59]|\uD83B[\uDC71-\uDCAB\uDCAD-\uDCAF\uDCB1-\uDCB4\uDD01-\uDD2D\uDD2F-\uDD3D\uDE00-\uDE03\uDE05-\uDE1F\uDE21\uDE22\uDE24\uDE27\uDE29-\uDE32\uDE34-\uDE37\uDE39\uDE3B\uDE42\uDE47\uDE49\uDE4B\uDE4D-\uDE4F\uDE51\uDE52\uDE54\uDE57\uDE59\uDE5B\uDE5D\uDE5F\uDE61\uDE62\uDE64\uDE67-\uDE6A\uDE6C-\uDE72\uDE74-\uDE77\uDE79-\uDE7C\uDE7E\uDE80-\uDE89\uDE8B-\uDE9B\uDEA1-\uDEA3\uDEA5-\uDEA9\uDEAB-\uDEBB]|\uD83C[\uDD00-\uDD0C]|\uD83E[\uDFF0-\uDFF9]|\uD869[\uDC00-\uDEDD\uDF00-\uDFFF]|\uD86D[\uDC00-\uDF34\uDF40-\uDFFF]|\uD86E[\uDC00-\uDC1D\uDC20-\uDFFF]|\uD873[\uDC00-\uDEA1\uDEB0-\uDFFF]|\uD87A[\uDC00-\uDFE0]|\uD87E[\uDC00-\uDE1D]|\uD884[\uDC00-\uDF4A])/))
              return;
            var nextChar = match[1] || match[2] || "";
            if (!nextChar || nextChar && (prevChar === "" || this.rules.inline.punctuation.exec(prevChar))) {
              var lLength = match[0].length - 1;
              var rDelim, rLength, delimTotal = lLength, midDelimTotal = 0;
              var endReg = match[0][0] === "*" ? this.rules.inline.emStrong.rDelimAst : this.rules.inline.emStrong.rDelimUnd;
              endReg.lastIndex = 0;
              maskedSrc = maskedSrc.slice(-1 * src.length + lLength);
              while ((match = endReg.exec(maskedSrc)) != null) {
                rDelim = match[1] || match[2] || match[3] || match[4] || match[5] || match[6];
                if (!rDelim)
                  continue;
                rLength = rDelim.length;
                if (match[3] || match[4]) {
                  delimTotal += rLength;
                  continue;
                } else if (match[5] || match[6]) {
                  if (lLength % 3 && !((lLength + rLength) % 3)) {
                    midDelimTotal += rLength;
                    continue;
                  }
                }
                delimTotal -= rLength;
                if (delimTotal > 0)
                  continue;
                rLength = Math.min(rLength, rLength + delimTotal + midDelimTotal);
                if (Math.min(lLength, rLength) % 2) {
                  return {
                    type: "em",
                    raw: src.slice(0, lLength + match.index + rLength + 1),
                    text: src.slice(1, lLength + match.index + rLength)
                  };
                }
                return {
                  type: "strong",
                  raw: src.slice(0, lLength + match.index + rLength + 1),
                  text: src.slice(2, lLength + match.index + rLength - 1)
                };
              }
            }
          };
          _proto.codespan = function codespan(src) {
            var cap = this.rules.inline.code.exec(src);
            if (cap) {
              var text = cap[2].replace(/\n/g, " ");
              var hasNonSpaceChars = /[^ ]/.test(text);
              var hasSpaceCharsOnBothEnds = /^ /.test(text) && / $/.test(text);
              if (hasNonSpaceChars && hasSpaceCharsOnBothEnds) {
                text = text.substring(1, text.length - 1);
              }
              text = _escape(text, true);
              return {
                type: "codespan",
                raw: cap[0],
                text
              };
            }
          };
          _proto.br = function br(src) {
            var cap = this.rules.inline.br.exec(src);
            if (cap) {
              return {
                type: "br",
                raw: cap[0]
              };
            }
          };
          _proto.del = function del(src) {
            var cap = this.rules.inline.del.exec(src);
            if (cap) {
              return {
                type: "del",
                raw: cap[0],
                text: cap[2]
              };
            }
          };
          _proto.autolink = function autolink(src, mangle2) {
            var cap = this.rules.inline.autolink.exec(src);
            if (cap) {
              var text, href;
              if (cap[2] === "@") {
                text = _escape(this.options.mangle ? mangle2(cap[1]) : cap[1]);
                href = "mailto:" + text;
              } else {
                text = _escape(cap[1]);
                href = text;
              }
              return {
                type: "link",
                raw: cap[0],
                text,
                href,
                tokens: [{
                  type: "text",
                  raw: text,
                  text
                }]
              };
            }
          };
          _proto.url = function url(src, mangle2) {
            var cap;
            if (cap = this.rules.inline.url.exec(src)) {
              var text, href;
              if (cap[2] === "@") {
                text = _escape(this.options.mangle ? mangle2(cap[0]) : cap[0]);
                href = "mailto:" + text;
              } else {
                var prevCapZero;
                do {
                  prevCapZero = cap[0];
                  cap[0] = this.rules.inline._backpedal.exec(cap[0])[0];
                } while (prevCapZero !== cap[0]);
                text = _escape(cap[0]);
                if (cap[1] === "www.") {
                  href = "http://" + text;
                } else {
                  href = text;
                }
              }
              return {
                type: "link",
                raw: cap[0],
                text,
                href,
                tokens: [{
                  type: "text",
                  raw: text,
                  text
                }]
              };
            }
          };
          _proto.inlineText = function inlineText(src, inRawBlock, smartypants2) {
            var cap = this.rules.inline.text.exec(src);
            if (cap) {
              var text;
              if (inRawBlock) {
                text = this.options.sanitize ? this.options.sanitizer ? this.options.sanitizer(cap[0]) : _escape(cap[0]) : cap[0];
              } else {
                text = _escape(this.options.smartypants ? smartypants2(cap[0]) : cap[0]);
              }
              return {
                type: "text",
                raw: cap[0],
                text
              };
            }
          };
          return Tokenizer2;
        }();
        var noopTest = helpers.noopTest, edit = helpers.edit, merge$1 = helpers.merge;
        var block$1 = {
          newline: /^(?: *(?:\n|$))+/,
          code: /^( {4}[^\n]+(?:\n(?: *(?:\n|$))*)?)+/,
          fences: /^ {0,3}(`{3,}(?=[^`\n]*\n)|~{3,})([^\n]*)\n(?:|([\s\S]*?)\n)(?: {0,3}\1[~`]* *(?:\n+|$)|$)/,
          hr: /^ {0,3}((?:- *){3,}|(?:_ *){3,}|(?:\* *){3,})(?:\n+|$)/,
          heading: /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/,
          blockquote: /^( {0,3}> ?(paragraph|[^\n]*)(?:\n|$))+/,
          list: /^( {0,3})(bull) [\s\S]+?(?:hr|def|\n{2,}(?! )(?! {0,3}bull )\n*|\s*$)/,
          html: "^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n+|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>\\n*|$)|<![A-Z][\\s\\S]*?(?:>\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>\\n*|$)|</?(tag)(?: +|\\n|/?>)[\\s\\S]*?(?:(?:\\n *)+\\n|$)|<(?!script|pre|style|textarea)([a-z][\\w-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n *)+\\n|$)|</(?!script|pre|style|textarea)[a-z][\\w-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n *)+\\n|$))",
          def: /^ {0,3}\[(label)\]: *\n? *<?([^\s>]+)>?(?:(?: +\n? *| *\n *)(title))? *(?:\n+|$)/,
          nptable: noopTest,
          table: noopTest,
          lheading: /^([^\n]+)\n {0,3}(=+|-+) *(?:\n+|$)/,
          // regex template, placeholders will be replaced according to different paragraph
          // interruption rules of commonmark and the original markdown spec:
          _paragraph: /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html| +\n)[^\n]+)*)/,
          text: /^[^\n]+/
        };
        block$1._label = /(?!\s*\])(?:\\[\[\]]|[^\[\]])+/;
        block$1._title = /(?:"(?:\\"?|[^"\\])*"|'[^'\n]*(?:\n[^'\n]+)*\n?'|\([^()]*\))/;
        block$1.def = edit(block$1.def).replace("label", block$1._label).replace("title", block$1._title).getRegex();
        block$1.bullet = /(?:[*+-]|\d{1,9}[.)])/;
        block$1.item = /^( *)(bull) ?[^\n]*(?:\n(?! *bull ?)[^\n]*)*/;
        block$1.item = edit(block$1.item, "gm").replace(/bull/g, block$1.bullet).getRegex();
        block$1.listItemStart = edit(/^( *)(bull) */).replace("bull", block$1.bullet).getRegex();
        block$1.list = edit(block$1.list).replace(/bull/g, block$1.bullet).replace("hr", "\\n+(?=\\1?(?:(?:- *){3,}|(?:_ *){3,}|(?:\\* *){3,})(?:\\n+|$))").replace("def", "\\n+(?=" + block$1.def.source + ")").getRegex();
        block$1._tag = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|section|source|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul";
        block$1._comment = /<!--(?!-?>)[\s\S]*?(?:-->|$)/;
        block$1.html = edit(block$1.html, "i").replace("comment", block$1._comment).replace("tag", block$1._tag).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex();
        block$1.paragraph = edit(block$1._paragraph).replace("hr", block$1.hr).replace("heading", " {0,3}#{1,6} ").replace("|lheading", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", block$1._tag).getRegex();
        block$1.blockquote = edit(block$1.blockquote).replace("paragraph", block$1.paragraph).getRegex();
        block$1.normal = merge$1({}, block$1);
        block$1.gfm = merge$1({}, block$1.normal, {
          nptable: "^ *([^|\\n ].*\\|.*)\\n {0,3}([-:]+ *\\|[-| :]*)(?:\\n((?:(?!\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)",
          // Cells
          table: "^ *\\|(.+)\\n {0,3}\\|?( *[-:]+[-| :]*)(?:\\n *((?:(?!\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)"
          // Cells
        });
        block$1.gfm.nptable = edit(block$1.gfm.nptable).replace("hr", block$1.hr).replace("heading", " {0,3}#{1,6} ").replace("blockquote", " {0,3}>").replace("code", " {4}[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", block$1._tag).getRegex();
        block$1.gfm.table = edit(block$1.gfm.table).replace("hr", block$1.hr).replace("heading", " {0,3}#{1,6} ").replace("blockquote", " {0,3}>").replace("code", " {4}[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", block$1._tag).getRegex();
        block$1.pedantic = merge$1({}, block$1.normal, {
          html: edit(`^ *(?:comment *(?:\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1> *(?:\\n{2,}|\\s*$)|<tag(?:"[^"]*"|'[^']*'|\\s[^'"/>\\s]*)*?/?> *(?:\\n{2,}|\\s*$))`).replace("comment", block$1._comment).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
          def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
          heading: /^(#{1,6})(.*)(?:\n+|$)/,
          fences: noopTest,
          // fences not supported
          paragraph: edit(block$1.normal._paragraph).replace("hr", block$1.hr).replace("heading", " *#{1,6} *[^\n]").replace("lheading", block$1.lheading).replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").getRegex()
        });
        var inline$1 = {
          escape: /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/,
          autolink: /^<(scheme:[^\s\x00-\x1f<>]*|email)>/,
          url: noopTest,
          tag: "^comment|^</[a-zA-Z][\\w:-]*\\s*>|^<[a-zA-Z][\\w-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>",
          // CDATA section
          link: /^!?\[(label)\]\(\s*(href)(?:\s+(title))?\s*\)/,
          reflink: /^!?\[(label)\]\[(?!\s*\])((?:\\[\[\]]?|[^\[\]\\])+)\]/,
          nolink: /^!?\[(?!\s*\])((?:\[[^\[\]]*\]|\\[\[\]]|[^\[\]])*)\](?:\[\])?/,
          reflinkSearch: "reflink|nolink(?!\\()",
          emStrong: {
            lDelim: /^(?:\*+(?:([punct_])|[^\s*]))|^_+(?:([punct*])|([^\s_]))/,
            //        (1) and (2) can only be a Right Delimiter. (3) and (4) can only be Left.  (5) and (6) can be either Left or Right.
            //        () Skip other delimiter (1) #***                   (2) a***#, a***                   (3) #***a, ***a                 (4) ***#              (5) #***#                 (6) a***a
            rDelimAst: /\_\_[^_*]*?\*[^_*]*?\_\_|[punct_](\*+)(?=[\s]|$)|[^punct*_\s](\*+)(?=[punct_\s]|$)|[punct_\s](\*+)(?=[^punct*_\s])|[\s](\*+)(?=[punct_])|[punct_](\*+)(?=[punct_])|[^punct*_\s](\*+)(?=[^punct*_\s])/,
            rDelimUnd: /\*\*[^_*]*?\_[^_*]*?\*\*|[punct*](\_+)(?=[\s]|$)|[^punct*_\s](\_+)(?=[punct*\s]|$)|[punct*\s](\_+)(?=[^punct*_\s])|[\s](\_+)(?=[punct*])|[punct*](\_+)(?=[punct*])/
            // ^- Not allowed for _
          },
          code: /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/,
          br: /^( {2,}|\\)\n(?!\s*$)/,
          del: noopTest,
          text: /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/,
          punctuation: /^([\spunctuation])/
        };
        inline$1._punctuation = "!\"#$%&'()+\\-.,/:;<=>?@\\[\\]`^{|}~";
        inline$1.punctuation = edit(inline$1.punctuation).replace(/punctuation/g, inline$1._punctuation).getRegex();
        inline$1.blockSkip = /\[[^\]]*?\]\([^\)]*?\)|`[^`]*?`|<[^>]*?>/g;
        inline$1.escapedEmSt = /\\\*|\\_/g;
        inline$1._comment = edit(block$1._comment).replace("(?:-->|$)", "-->").getRegex();
        inline$1.emStrong.lDelim = edit(inline$1.emStrong.lDelim).replace(/punct/g, inline$1._punctuation).getRegex();
        inline$1.emStrong.rDelimAst = edit(inline$1.emStrong.rDelimAst, "g").replace(/punct/g, inline$1._punctuation).getRegex();
        inline$1.emStrong.rDelimUnd = edit(inline$1.emStrong.rDelimUnd, "g").replace(/punct/g, inline$1._punctuation).getRegex();
        inline$1._escapes = /\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/g;
        inline$1._scheme = /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/;
        inline$1._email = /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/;
        inline$1.autolink = edit(inline$1.autolink).replace("scheme", inline$1._scheme).replace("email", inline$1._email).getRegex();
        inline$1._attribute = /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/;
        inline$1.tag = edit(inline$1.tag).replace("comment", inline$1._comment).replace("attribute", inline$1._attribute).getRegex();
        inline$1._label = /(?:\[(?:\\.|[^\[\]\\])*\]|\\.|`[^`]*`|[^\[\]\\`])*?/;
        inline$1._href = /<(?:\\.|[^\n<>\\])+>|[^\s\x00-\x1f]*/;
        inline$1._title = /"(?:\\"?|[^"\\])*"|'(?:\\'?|[^'\\])*'|\((?:\\\)?|[^)\\])*\)/;
        inline$1.link = edit(inline$1.link).replace("label", inline$1._label).replace("href", inline$1._href).replace("title", inline$1._title).getRegex();
        inline$1.reflink = edit(inline$1.reflink).replace("label", inline$1._label).getRegex();
        inline$1.reflinkSearch = edit(inline$1.reflinkSearch, "g").replace("reflink", inline$1.reflink).replace("nolink", inline$1.nolink).getRegex();
        inline$1.normal = merge$1({}, inline$1);
        inline$1.pedantic = merge$1({}, inline$1.normal, {
          strong: {
            start: /^__|\*\*/,
            middle: /^__(?=\S)([\s\S]*?\S)__(?!_)|^\*\*(?=\S)([\s\S]*?\S)\*\*(?!\*)/,
            endAst: /\*\*(?!\*)/g,
            endUnd: /__(?!_)/g
          },
          em: {
            start: /^_|\*/,
            middle: /^()\*(?=\S)([\s\S]*?\S)\*(?!\*)|^_(?=\S)([\s\S]*?\S)_(?!_)/,
            endAst: /\*(?!\*)/g,
            endUnd: /_(?!_)/g
          },
          link: edit(/^!?\[(label)\]\((.*?)\)/).replace("label", inline$1._label).getRegex(),
          reflink: edit(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", inline$1._label).getRegex()
        });
        inline$1.gfm = merge$1({}, inline$1.normal, {
          escape: edit(inline$1.escape).replace("])", "~|])").getRegex(),
          _extended_email: /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![-_])/,
          url: /^((?:ftp|https?):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/,
          _backpedal: /(?:[^?!.,:;*_~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_~)]+(?!$))+/,
          del: /^(~~?)(?=[^\s~])([\s\S]*?[^\s~])\1(?=[^~]|$)/,
          text: /^([`~]+|[^`~])(?:(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|https?:\/\/|ftp:\/\/|www\.|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)))/
        });
        inline$1.gfm.url = edit(inline$1.gfm.url, "i").replace("email", inline$1.gfm._extended_email).getRegex();
        inline$1.breaks = merge$1({}, inline$1.gfm, {
          br: edit(inline$1.br).replace("{2,}", "*").getRegex(),
          text: edit(inline$1.gfm.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
        });
        var rules = {
          block: block$1,
          inline: inline$1
        };
        var Tokenizer$1 = Tokenizer_1;
        var defaults$3 = defaults$5.exports.defaults;
        var block = rules.block, inline = rules.inline;
        var repeatString = helpers.repeatString;
        function smartypants(text) {
          return text.replace(/---/g, "\u2014").replace(/--/g, "\u2013").replace(/(^|[-\u2014/(\[{"\s])'/g, "$1\u2018").replace(/'/g, "\u2019").replace(/(^|[-\u2014/(\[{\u2018\s])"/g, "$1\u201C").replace(/"/g, "\u201D").replace(/\.{3}/g, "\u2026");
        }
        function mangle(text) {
          var out = "", i, ch;
          var l = text.length;
          for (i = 0; i < l; i++) {
            ch = text.charCodeAt(i);
            if (Math.random() > 0.5) {
              ch = "x" + ch.toString(16);
            }
            out += "&#" + ch + ";";
          }
          return out;
        }
        var Lexer_1 = /* @__PURE__ */ function() {
          function Lexer2(options) {
            this.tokens = [];
            this.tokens.links = /* @__PURE__ */ Object.create(null);
            this.options = options || defaults$3;
            this.options.tokenizer = this.options.tokenizer || new Tokenizer$1();
            this.tokenizer = this.options.tokenizer;
            this.tokenizer.options = this.options;
            var rules2 = {
              block: block.normal,
              inline: inline.normal
            };
            if (this.options.pedantic) {
              rules2.block = block.pedantic;
              rules2.inline = inline.pedantic;
            } else if (this.options.gfm) {
              rules2.block = block.gfm;
              if (this.options.breaks) {
                rules2.inline = inline.breaks;
              } else {
                rules2.inline = inline.gfm;
              }
            }
            this.tokenizer.rules = rules2;
          }
          Lexer2.lex = function lex(src, options) {
            var lexer = new Lexer2(options);
            return lexer.lex(src);
          };
          Lexer2.lexInline = function lexInline(src, options) {
            var lexer = new Lexer2(options);
            return lexer.inlineTokens(src);
          };
          var _proto = Lexer2.prototype;
          _proto.lex = function lex(src) {
            src = src.replace(/\r\n|\r/g, "\n").replace(/\t/g, "    ");
            this.blockTokens(src, this.tokens, true);
            this.inline(this.tokens);
            return this.tokens;
          };
          _proto.blockTokens = function blockTokens(src, tokens, top) {
            var _this = this;
            if (tokens === void 0) {
              tokens = [];
            }
            if (top === void 0) {
              top = true;
            }
            if (this.options.pedantic) {
              src = src.replace(/^ +$/gm, "");
            }
            var token, i, l, lastToken, cutSrc, lastParagraphClipped;
            while (src) {
              if (this.options.extensions && this.options.extensions.block && this.options.extensions.block.some(function(extTokenizer) {
                if (token = extTokenizer.call(_this, src, tokens)) {
                  src = src.substring(token.raw.length);
                  tokens.push(token);
                  return true;
                }
                return false;
              })) {
                continue;
              }
              if (token = this.tokenizer.space(src)) {
                src = src.substring(token.raw.length);
                if (token.type) {
                  tokens.push(token);
                }
                continue;
              }
              if (token = this.tokenizer.code(src)) {
                src = src.substring(token.raw.length);
                lastToken = tokens[tokens.length - 1];
                if (lastToken && lastToken.type === "paragraph") {
                  lastToken.raw += "\n" + token.raw;
                  lastToken.text += "\n" + token.text;
                } else {
                  tokens.push(token);
                }
                continue;
              }
              if (token = this.tokenizer.fences(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.heading(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.nptable(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.hr(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.blockquote(src)) {
                src = src.substring(token.raw.length);
                token.tokens = this.blockTokens(token.text, [], top);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.list(src)) {
                src = src.substring(token.raw.length);
                l = token.items.length;
                for (i = 0; i < l; i++) {
                  token.items[i].tokens = this.blockTokens(token.items[i].text, [], false);
                }
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.html(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (top && (token = this.tokenizer.def(src))) {
                src = src.substring(token.raw.length);
                if (!this.tokens.links[token.tag]) {
                  this.tokens.links[token.tag] = {
                    href: token.href,
                    title: token.title
                  };
                }
                continue;
              }
              if (token = this.tokenizer.table(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.lheading(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              cutSrc = src;
              if (this.options.extensions && this.options.extensions.startBlock) {
                (function() {
                  var startIndex = Infinity;
                  var tempSrc = src.slice(1);
                  var tempStart = void 0;
                  _this.options.extensions.startBlock.forEach(function(getStartIndex) {
                    tempStart = getStartIndex.call(this, tempSrc);
                    if (typeof tempStart === "number" && tempStart >= 0) {
                      startIndex = Math.min(startIndex, tempStart);
                    }
                  });
                  if (startIndex < Infinity && startIndex >= 0) {
                    cutSrc = src.substring(0, startIndex + 1);
                  }
                })();
              }
              if (top && (token = this.tokenizer.paragraph(cutSrc))) {
                lastToken = tokens[tokens.length - 1];
                if (lastParagraphClipped && lastToken.type === "paragraph") {
                  lastToken.raw += "\n" + token.raw;
                  lastToken.text += "\n" + token.text;
                } else {
                  tokens.push(token);
                }
                lastParagraphClipped = cutSrc.length !== src.length;
                src = src.substring(token.raw.length);
                continue;
              }
              if (token = this.tokenizer.text(src)) {
                src = src.substring(token.raw.length);
                lastToken = tokens[tokens.length - 1];
                if (lastToken && lastToken.type === "text") {
                  lastToken.raw += "\n" + token.raw;
                  lastToken.text += "\n" + token.text;
                } else {
                  tokens.push(token);
                }
                continue;
              }
              if (src) {
                var errMsg = "Infinite loop on byte: " + src.charCodeAt(0);
                if (this.options.silent) {
                  console.error(errMsg);
                  break;
                } else {
                  throw new Error(errMsg);
                }
              }
            }
            return tokens;
          };
          _proto.inline = function inline2(tokens) {
            var i, j, k, l2, row, token;
            var l = tokens.length;
            for (i = 0; i < l; i++) {
              token = tokens[i];
              switch (token.type) {
                case "paragraph":
                case "text":
                case "heading": {
                  token.tokens = [];
                  this.inlineTokens(token.text, token.tokens);
                  break;
                }
                case "table": {
                  token.tokens = {
                    header: [],
                    cells: []
                  };
                  l2 = token.header.length;
                  for (j = 0; j < l2; j++) {
                    token.tokens.header[j] = [];
                    this.inlineTokens(token.header[j], token.tokens.header[j]);
                  }
                  l2 = token.cells.length;
                  for (j = 0; j < l2; j++) {
                    row = token.cells[j];
                    token.tokens.cells[j] = [];
                    for (k = 0; k < row.length; k++) {
                      token.tokens.cells[j][k] = [];
                      this.inlineTokens(row[k], token.tokens.cells[j][k]);
                    }
                  }
                  break;
                }
                case "blockquote": {
                  this.inline(token.tokens);
                  break;
                }
                case "list": {
                  l2 = token.items.length;
                  for (j = 0; j < l2; j++) {
                    this.inline(token.items[j].tokens);
                  }
                  break;
                }
              }
            }
            return tokens;
          };
          _proto.inlineTokens = function inlineTokens(src, tokens, inLink, inRawBlock) {
            var _this2 = this;
            if (tokens === void 0) {
              tokens = [];
            }
            if (inLink === void 0) {
              inLink = false;
            }
            if (inRawBlock === void 0) {
              inRawBlock = false;
            }
            var token, lastToken, cutSrc;
            var maskedSrc = src;
            var match;
            var keepPrevChar, prevChar;
            if (this.tokens.links) {
              var links = Object.keys(this.tokens.links);
              if (links.length > 0) {
                while ((match = this.tokenizer.rules.inline.reflinkSearch.exec(maskedSrc)) != null) {
                  if (links.includes(match[0].slice(match[0].lastIndexOf("[") + 1, -1))) {
                    maskedSrc = maskedSrc.slice(0, match.index) + "[" + repeatString("a", match[0].length - 2) + "]" + maskedSrc.slice(this.tokenizer.rules.inline.reflinkSearch.lastIndex);
                  }
                }
              }
            }
            while ((match = this.tokenizer.rules.inline.blockSkip.exec(maskedSrc)) != null) {
              maskedSrc = maskedSrc.slice(0, match.index) + "[" + repeatString("a", match[0].length - 2) + "]" + maskedSrc.slice(this.tokenizer.rules.inline.blockSkip.lastIndex);
            }
            while ((match = this.tokenizer.rules.inline.escapedEmSt.exec(maskedSrc)) != null) {
              maskedSrc = maskedSrc.slice(0, match.index) + "++" + maskedSrc.slice(this.tokenizer.rules.inline.escapedEmSt.lastIndex);
            }
            while (src) {
              if (!keepPrevChar) {
                prevChar = "";
              }
              keepPrevChar = false;
              if (this.options.extensions && this.options.extensions.inline && this.options.extensions.inline.some(function(extTokenizer) {
                if (token = extTokenizer.call(_this2, src, tokens)) {
                  src = src.substring(token.raw.length);
                  tokens.push(token);
                  return true;
                }
                return false;
              })) {
                continue;
              }
              if (token = this.tokenizer.escape(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.tag(src, inLink, inRawBlock)) {
                src = src.substring(token.raw.length);
                inLink = token.inLink;
                inRawBlock = token.inRawBlock;
                lastToken = tokens[tokens.length - 1];
                if (lastToken && token.type === "text" && lastToken.type === "text") {
                  lastToken.raw += token.raw;
                  lastToken.text += token.text;
                } else {
                  tokens.push(token);
                }
                continue;
              }
              if (token = this.tokenizer.link(src)) {
                src = src.substring(token.raw.length);
                if (token.type === "link") {
                  token.tokens = this.inlineTokens(token.text, [], true, inRawBlock);
                }
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.reflink(src, this.tokens.links)) {
                src = src.substring(token.raw.length);
                lastToken = tokens[tokens.length - 1];
                if (token.type === "link") {
                  token.tokens = this.inlineTokens(token.text, [], true, inRawBlock);
                  tokens.push(token);
                } else if (lastToken && token.type === "text" && lastToken.type === "text") {
                  lastToken.raw += token.raw;
                  lastToken.text += token.text;
                } else {
                  tokens.push(token);
                }
                continue;
              }
              if (token = this.tokenizer.emStrong(src, maskedSrc, prevChar)) {
                src = src.substring(token.raw.length);
                token.tokens = this.inlineTokens(token.text, [], inLink, inRawBlock);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.codespan(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.br(src)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.del(src)) {
                src = src.substring(token.raw.length);
                token.tokens = this.inlineTokens(token.text, [], inLink, inRawBlock);
                tokens.push(token);
                continue;
              }
              if (token = this.tokenizer.autolink(src, mangle)) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              if (!inLink && (token = this.tokenizer.url(src, mangle))) {
                src = src.substring(token.raw.length);
                tokens.push(token);
                continue;
              }
              cutSrc = src;
              if (this.options.extensions && this.options.extensions.startInline) {
                (function() {
                  var startIndex = Infinity;
                  var tempSrc = src.slice(1);
                  var tempStart = void 0;
                  _this2.options.extensions.startInline.forEach(function(getStartIndex) {
                    tempStart = getStartIndex.call(this, tempSrc);
                    if (typeof tempStart === "number" && tempStart >= 0) {
                      startIndex = Math.min(startIndex, tempStart);
                    }
                  });
                  if (startIndex < Infinity && startIndex >= 0) {
                    cutSrc = src.substring(0, startIndex + 1);
                  }
                })();
              }
              if (token = this.tokenizer.inlineText(cutSrc, inRawBlock, smartypants)) {
                src = src.substring(token.raw.length);
                if (token.raw.slice(-1) !== "_") {
                  prevChar = token.raw.slice(-1);
                }
                keepPrevChar = true;
                lastToken = tokens[tokens.length - 1];
                if (lastToken && lastToken.type === "text") {
                  lastToken.raw += token.raw;
                  lastToken.text += token.text;
                } else {
                  tokens.push(token);
                }
                continue;
              }
              if (src) {
                var errMsg = "Infinite loop on byte: " + src.charCodeAt(0);
                if (this.options.silent) {
                  console.error(errMsg);
                  break;
                } else {
                  throw new Error(errMsg);
                }
              }
            }
            return tokens;
          };
          _createClass(Lexer2, null, [{
            key: "rules",
            get: function get() {
              return {
                block,
                inline
              };
            }
          }]);
          return Lexer2;
        }();
        var defaults$2 = defaults$5.exports.defaults;
        var cleanUrl = helpers.cleanUrl, escape$1 = helpers.escape;
        var Renderer_1 = /* @__PURE__ */ function() {
          function Renderer2(options) {
            this.options = options || defaults$2;
          }
          var _proto = Renderer2.prototype;
          _proto.code = function code(_code, infostring, escaped) {
            var lang = (infostring || "").match(/\S*/)[0];
            if (this.options.highlight) {
              var out = this.options.highlight(_code, lang);
              if (out != null && out !== _code) {
                escaped = true;
                _code = out;
              }
            }
            _code = _code.replace(/\n$/, "") + "\n";
            if (!lang) {
              return "<pre><code>" + (escaped ? _code : escape$1(_code, true)) + "</code></pre>\n";
            }
            return '<pre><code class="' + this.options.langPrefix + escape$1(lang, true) + '">' + (escaped ? _code : escape$1(_code, true)) + "</code></pre>\n";
          };
          _proto.blockquote = function blockquote(quote) {
            return "<blockquote>\n" + quote + "</blockquote>\n";
          };
          _proto.html = function html(_html) {
            return _html;
          };
          _proto.heading = function heading(text, level, raw, slugger) {
            if (this.options.headerIds) {
              return "<h" + level + ' id="' + this.options.headerPrefix + slugger.slug(raw) + '">' + text + "</h" + level + ">\n";
            }
            return "<h" + level + ">" + text + "</h" + level + ">\n";
          };
          _proto.hr = function hr() {
            return this.options.xhtml ? "<hr/>\n" : "<hr>\n";
          };
          _proto.list = function list(body, ordered, start) {
            var type = ordered ? "ol" : "ul", startatt = ordered && start !== 1 ? ' start="' + start + '"' : "";
            return "<" + type + startatt + ">\n" + body + "</" + type + ">\n";
          };
          _proto.listitem = function listitem(text) {
            return "<li>" + text + "</li>\n";
          };
          _proto.checkbox = function checkbox(checked) {
            return "<input " + (checked ? 'checked="" ' : "") + 'disabled="" type="checkbox"' + (this.options.xhtml ? " /" : "") + "> ";
          };
          _proto.paragraph = function paragraph(text) {
            return "<p>" + text + "</p>\n";
          };
          _proto.table = function table(header, body) {
            if (body)
              body = "<tbody>" + body + "</tbody>";
            return "<table>\n<thead>\n" + header + "</thead>\n" + body + "</table>\n";
          };
          _proto.tablerow = function tablerow(content) {
            return "<tr>\n" + content + "</tr>\n";
          };
          _proto.tablecell = function tablecell(content, flags) {
            var type = flags.header ? "th" : "td";
            var tag = flags.align ? "<" + type + ' align="' + flags.align + '">' : "<" + type + ">";
            return tag + content + "</" + type + ">\n";
          };
          _proto.strong = function strong(text) {
            return "<strong>" + text + "</strong>";
          };
          _proto.em = function em(text) {
            return "<em>" + text + "</em>";
          };
          _proto.codespan = function codespan(text) {
            return "<code>" + text + "</code>";
          };
          _proto.br = function br() {
            return this.options.xhtml ? "<br/>" : "<br>";
          };
          _proto.del = function del(text) {
            return "<del>" + text + "</del>";
          };
          _proto.link = function link(href, title, text) {
            href = cleanUrl(this.options.sanitize, this.options.baseUrl, href);
            if (href === null) {
              return text;
            }
            var out = '<a href="' + escape$1(href) + '"';
            if (title) {
              out += ' title="' + title + '"';
            }
            out += ">" + text + "</a>";
            return out;
          };
          _proto.image = function image(href, title, text) {
            href = cleanUrl(this.options.sanitize, this.options.baseUrl, href);
            if (href === null) {
              return text;
            }
            var out = '<img src="' + href + '" alt="' + text + '"';
            if (title) {
              out += ' title="' + title + '"';
            }
            out += this.options.xhtml ? "/>" : ">";
            return out;
          };
          _proto.text = function text(_text) {
            return _text;
          };
          return Renderer2;
        }();
        var TextRenderer_1 = /* @__PURE__ */ function() {
          function TextRenderer2() {
          }
          var _proto = TextRenderer2.prototype;
          _proto.strong = function strong(text) {
            return text;
          };
          _proto.em = function em(text) {
            return text;
          };
          _proto.codespan = function codespan(text) {
            return text;
          };
          _proto.del = function del(text) {
            return text;
          };
          _proto.html = function html(text) {
            return text;
          };
          _proto.text = function text(_text) {
            return _text;
          };
          _proto.link = function link(href, title, text) {
            return "" + text;
          };
          _proto.image = function image(href, title, text) {
            return "" + text;
          };
          _proto.br = function br() {
            return "";
          };
          return TextRenderer2;
        }();
        var Slugger_1 = /* @__PURE__ */ function() {
          function Slugger2() {
            this.seen = {};
          }
          var _proto = Slugger2.prototype;
          _proto.serialize = function serialize(value) {
            return value.toLowerCase().trim().replace(/<[!\/a-z].*?>/ig, "").replace(/[\u2000-\u206F\u2E00-\u2E7F\\'!"#$%&()*+,./:;<=>?@[\]^`{|}~]/g, "").replace(/\s/g, "-");
          };
          _proto.getNextSafeSlug = function getNextSafeSlug(originalSlug, isDryRun) {
            var slug = originalSlug;
            var occurenceAccumulator = 0;
            if (this.seen.hasOwnProperty(slug)) {
              occurenceAccumulator = this.seen[originalSlug];
              do {
                occurenceAccumulator++;
                slug = originalSlug + "-" + occurenceAccumulator;
              } while (this.seen.hasOwnProperty(slug));
            }
            if (!isDryRun) {
              this.seen[originalSlug] = occurenceAccumulator;
              this.seen[slug] = 0;
            }
            return slug;
          };
          _proto.slug = function slug(value, options) {
            if (options === void 0) {
              options = {};
            }
            var slug2 = this.serialize(value);
            return this.getNextSafeSlug(slug2, options.dryrun);
          };
          return Slugger2;
        }();
        var Renderer$1 = Renderer_1;
        var TextRenderer$1 = TextRenderer_1;
        var Slugger$1 = Slugger_1;
        var defaults$1 = defaults$5.exports.defaults;
        var unescape = helpers.unescape;
        var Parser_1 = /* @__PURE__ */ function() {
          function Parser2(options) {
            this.options = options || defaults$1;
            this.options.renderer = this.options.renderer || new Renderer$1();
            this.renderer = this.options.renderer;
            this.renderer.options = this.options;
            this.textRenderer = new TextRenderer$1();
            this.slugger = new Slugger$1();
          }
          Parser2.parse = function parse(tokens, options) {
            var parser = new Parser2(options);
            return parser.parse(tokens);
          };
          Parser2.parseInline = function parseInline(tokens, options) {
            var parser = new Parser2(options);
            return parser.parseInline(tokens);
          };
          var _proto = Parser2.prototype;
          _proto.parse = function parse(tokens, top) {
            if (top === void 0) {
              top = true;
            }
            var out = "", i, j, k, l2, l3, row, cell, header, body, token, ordered, start, loose, itemBody, item, checked, task, checkbox, ret;
            var l = tokens.length;
            for (i = 0; i < l; i++) {
              token = tokens[i];
              if (this.options.extensions && this.options.extensions.renderers && this.options.extensions.renderers[token.type]) {
                ret = this.options.extensions.renderers[token.type].call(this, token);
                if (ret !== false || !["space", "hr", "heading", "code", "table", "blockquote", "list", "html", "paragraph", "text"].includes(token.type)) {
                  out += ret || "";
                  continue;
                }
              }
              switch (token.type) {
                case "space": {
                  continue;
                }
                case "hr": {
                  out += this.renderer.hr();
                  continue;
                }
                case "heading": {
                  out += this.renderer.heading(this.parseInline(token.tokens), token.depth, unescape(this.parseInline(token.tokens, this.textRenderer)), this.slugger);
                  continue;
                }
                case "code": {
                  out += this.renderer.code(token.text, token.lang, token.escaped);
                  continue;
                }
                case "table": {
                  header = "";
                  cell = "";
                  l2 = token.header.length;
                  for (j = 0; j < l2; j++) {
                    cell += this.renderer.tablecell(this.parseInline(token.tokens.header[j]), {
                      header: true,
                      align: token.align[j]
                    });
                  }
                  header += this.renderer.tablerow(cell);
                  body = "";
                  l2 = token.cells.length;
                  for (j = 0; j < l2; j++) {
                    row = token.tokens.cells[j];
                    cell = "";
                    l3 = row.length;
                    for (k = 0; k < l3; k++) {
                      cell += this.renderer.tablecell(this.parseInline(row[k]), {
                        header: false,
                        align: token.align[k]
                      });
                    }
                    body += this.renderer.tablerow(cell);
                  }
                  out += this.renderer.table(header, body);
                  continue;
                }
                case "blockquote": {
                  body = this.parse(token.tokens);
                  out += this.renderer.blockquote(body);
                  continue;
                }
                case "list": {
                  ordered = token.ordered;
                  start = token.start;
                  loose = token.loose;
                  l2 = token.items.length;
                  body = "";
                  for (j = 0; j < l2; j++) {
                    item = token.items[j];
                    checked = item.checked;
                    task = item.task;
                    itemBody = "";
                    if (item.task) {
                      checkbox = this.renderer.checkbox(checked);
                      if (loose) {
                        if (item.tokens.length > 0 && item.tokens[0].type === "text") {
                          item.tokens[0].text = checkbox + " " + item.tokens[0].text;
                          if (item.tokens[0].tokens && item.tokens[0].tokens.length > 0 && item.tokens[0].tokens[0].type === "text") {
                            item.tokens[0].tokens[0].text = checkbox + " " + item.tokens[0].tokens[0].text;
                          }
                        } else {
                          item.tokens.unshift({
                            type: "text",
                            text: checkbox
                          });
                        }
                      } else {
                        itemBody += checkbox;
                      }
                    }
                    itemBody += this.parse(item.tokens, loose);
                    body += this.renderer.listitem(itemBody, task, checked);
                  }
                  out += this.renderer.list(body, ordered, start);
                  continue;
                }
                case "html": {
                  out += this.renderer.html(token.text);
                  continue;
                }
                case "paragraph": {
                  out += this.renderer.paragraph(this.parseInline(token.tokens));
                  continue;
                }
                case "text": {
                  body = token.tokens ? this.parseInline(token.tokens) : token.text;
                  while (i + 1 < l && tokens[i + 1].type === "text") {
                    token = tokens[++i];
                    body += "\n" + (token.tokens ? this.parseInline(token.tokens) : token.text);
                  }
                  out += top ? this.renderer.paragraph(body) : body;
                  continue;
                }
                default: {
                  var errMsg = 'Token with "' + token.type + '" type was not found.';
                  if (this.options.silent) {
                    console.error(errMsg);
                    return;
                  } else {
                    throw new Error(errMsg);
                  }
                }
              }
            }
            return out;
          };
          _proto.parseInline = function parseInline(tokens, renderer) {
            renderer = renderer || this.renderer;
            var out = "", i, token, ret;
            var l = tokens.length;
            for (i = 0; i < l; i++) {
              token = tokens[i];
              if (this.options.extensions && this.options.extensions.renderers && this.options.extensions.renderers[token.type]) {
                ret = this.options.extensions.renderers[token.type].call(this, token);
                if (ret !== false || !["escape", "html", "link", "image", "strong", "em", "codespan", "br", "del", "text"].includes(token.type)) {
                  out += ret || "";
                  continue;
                }
              }
              switch (token.type) {
                case "escape": {
                  out += renderer.text(token.text);
                  break;
                }
                case "html": {
                  out += renderer.html(token.text);
                  break;
                }
                case "link": {
                  out += renderer.link(token.href, token.title, this.parseInline(token.tokens, renderer));
                  break;
                }
                case "image": {
                  out += renderer.image(token.href, token.title, token.text);
                  break;
                }
                case "strong": {
                  out += renderer.strong(this.parseInline(token.tokens, renderer));
                  break;
                }
                case "em": {
                  out += renderer.em(this.parseInline(token.tokens, renderer));
                  break;
                }
                case "codespan": {
                  out += renderer.codespan(token.text);
                  break;
                }
                case "br": {
                  out += renderer.br();
                  break;
                }
                case "del": {
                  out += renderer.del(this.parseInline(token.tokens, renderer));
                  break;
                }
                case "text": {
                  out += renderer.text(token.text);
                  break;
                }
                default: {
                  var errMsg = 'Token with "' + token.type + '" type was not found.';
                  if (this.options.silent) {
                    console.error(errMsg);
                    return;
                  } else {
                    throw new Error(errMsg);
                  }
                }
              }
            }
            return out;
          };
          return Parser2;
        }();
        var Lexer = Lexer_1;
        var Parser = Parser_1;
        var Tokenizer = Tokenizer_1;
        var Renderer = Renderer_1;
        var TextRenderer = TextRenderer_1;
        var Slugger = Slugger_1;
        var merge = helpers.merge, checkSanitizeDeprecation = helpers.checkSanitizeDeprecation, escape = helpers.escape;
        var getDefaults = defaults$5.exports.getDefaults, changeDefaults = defaults$5.exports.changeDefaults, defaults = defaults$5.exports.defaults;
        function marked(src, opt, callback) {
          if (typeof src === "undefined" || src === null) {
            throw new Error("marked(): input parameter is undefined or null");
          }
          if (typeof src !== "string") {
            throw new Error("marked(): input parameter is of type " + Object.prototype.toString.call(src) + ", string expected");
          }
          if (typeof opt === "function") {
            callback = opt;
            opt = null;
          }
          opt = merge({}, marked.defaults, opt || {});
          checkSanitizeDeprecation(opt);
          if (callback) {
            var highlight = opt.highlight;
            var tokens;
            try {
              tokens = Lexer.lex(src, opt);
            } catch (e) {
              return callback(e);
            }
            var done = function done2(err) {
              var out;
              if (!err) {
                try {
                  if (opt.walkTokens) {
                    marked.walkTokens(tokens, opt.walkTokens);
                  }
                  out = Parser.parse(tokens, opt);
                } catch (e) {
                  err = e;
                }
              }
              opt.highlight = highlight;
              return err ? callback(err) : callback(null, out);
            };
            if (!highlight || highlight.length < 3) {
              return done();
            }
            delete opt.highlight;
            if (!tokens.length)
              return done();
            var pending = 0;
            marked.walkTokens(tokens, function(token) {
              if (token.type === "code") {
                pending++;
                setTimeout(function() {
                  highlight(token.text, token.lang, function(err, code) {
                    if (err) {
                      return done(err);
                    }
                    if (code != null && code !== token.text) {
                      token.text = code;
                      token.escaped = true;
                    }
                    pending--;
                    if (pending === 0) {
                      done();
                    }
                  });
                }, 0);
              }
            });
            if (pending === 0) {
              done();
            }
            return;
          }
          try {
            var _tokens = Lexer.lex(src, opt);
            if (opt.walkTokens) {
              marked.walkTokens(_tokens, opt.walkTokens);
            }
            return Parser.parse(_tokens, opt);
          } catch (e) {
            e.message += "\nPlease report this to https://github.com/markedjs/marked.";
            if (opt.silent) {
              return "<p>An error occurred:</p><pre>" + escape(e.message + "", true) + "</pre>";
            }
            throw e;
          }
        }
        marked.options = marked.setOptions = function(opt) {
          merge(marked.defaults, opt);
          changeDefaults(marked.defaults);
          return marked;
        };
        marked.getDefaults = getDefaults;
        marked.defaults = defaults;
        marked.use = function() {
          var _this = this;
          for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
            args[_key] = arguments[_key];
          }
          var opts = merge.apply(void 0, [{}].concat(args));
          var extensions = marked.defaults.extensions || {
            renderers: {},
            childTokens: {}
          };
          var hasExtensions;
          args.forEach(function(pack) {
            if (pack.extensions) {
              hasExtensions = true;
              pack.extensions.forEach(function(ext) {
                if (!ext.name) {
                  throw new Error("extension name required");
                }
                if (ext.renderer) {
                  var prevRenderer = extensions.renderers ? extensions.renderers[ext.name] : null;
                  if (prevRenderer) {
                    extensions.renderers[ext.name] = function() {
                      for (var _len2 = arguments.length, args2 = new Array(_len2), _key2 = 0; _key2 < _len2; _key2++) {
                        args2[_key2] = arguments[_key2];
                      }
                      var ret = ext.renderer.apply(this, args2);
                      if (ret === false) {
                        ret = prevRenderer.apply(this, args2);
                      }
                      return ret;
                    };
                  } else {
                    extensions.renderers[ext.name] = ext.renderer;
                  }
                }
                if (ext.tokenizer) {
                  if (!ext.level || ext.level !== "block" && ext.level !== "inline") {
                    throw new Error("extension level must be 'block' or 'inline'");
                  }
                  if (extensions[ext.level]) {
                    extensions[ext.level].unshift(ext.tokenizer);
                  } else {
                    extensions[ext.level] = [ext.tokenizer];
                  }
                  if (ext.start) {
                    if (ext.level === "block") {
                      if (extensions.startBlock) {
                        extensions.startBlock.push(ext.start);
                      } else {
                        extensions.startBlock = [ext.start];
                      }
                    } else if (ext.level === "inline") {
                      if (extensions.startInline) {
                        extensions.startInline.push(ext.start);
                      } else {
                        extensions.startInline = [ext.start];
                      }
                    }
                  }
                }
                if (ext.childTokens) {
                  extensions.childTokens[ext.name] = ext.childTokens;
                }
              });
            }
            if (pack.renderer) {
              (function() {
                var renderer = marked.defaults.renderer || new Renderer();
                var _loop = function _loop2(prop2) {
                  var prevRenderer = renderer[prop2];
                  renderer[prop2] = function() {
                    for (var _len3 = arguments.length, args2 = new Array(_len3), _key3 = 0; _key3 < _len3; _key3++) {
                      args2[_key3] = arguments[_key3];
                    }
                    var ret = pack.renderer[prop2].apply(renderer, args2);
                    if (ret === false) {
                      ret = prevRenderer.apply(renderer, args2);
                    }
                    return ret;
                  };
                };
                for (var prop in pack.renderer) {
                  _loop(prop);
                }
                opts.renderer = renderer;
              })();
            }
            if (pack.tokenizer) {
              (function() {
                var tokenizer = marked.defaults.tokenizer || new Tokenizer();
                var _loop2 = function _loop22(prop2) {
                  var prevTokenizer = tokenizer[prop2];
                  tokenizer[prop2] = function() {
                    for (var _len4 = arguments.length, args2 = new Array(_len4), _key4 = 0; _key4 < _len4; _key4++) {
                      args2[_key4] = arguments[_key4];
                    }
                    var ret = pack.tokenizer[prop2].apply(tokenizer, args2);
                    if (ret === false) {
                      ret = prevTokenizer.apply(tokenizer, args2);
                    }
                    return ret;
                  };
                };
                for (var prop in pack.tokenizer) {
                  _loop2(prop);
                }
                opts.tokenizer = tokenizer;
              })();
            }
            if (pack.walkTokens) {
              var walkTokens = marked.defaults.walkTokens;
              opts.walkTokens = function(token) {
                pack.walkTokens.call(_this, token);
                if (walkTokens) {
                  walkTokens(token);
                }
              };
            }
            if (hasExtensions) {
              opts.extensions = extensions;
            }
            marked.setOptions(opts);
          });
        };
        marked.walkTokens = function(tokens, callback) {
          var _loop3 = function _loop32() {
            var token = _step.value;
            callback(token);
            switch (token.type) {
              case "table": {
                for (var _iterator2 = _createForOfIteratorHelperLoose(token.tokens.header), _step2; !(_step2 = _iterator2()).done; ) {
                  var cell = _step2.value;
                  marked.walkTokens(cell, callback);
                }
                for (var _iterator3 = _createForOfIteratorHelperLoose(token.tokens.cells), _step3; !(_step3 = _iterator3()).done; ) {
                  var row = _step3.value;
                  for (var _iterator4 = _createForOfIteratorHelperLoose(row), _step4; !(_step4 = _iterator4()).done; ) {
                    var _cell = _step4.value;
                    marked.walkTokens(_cell, callback);
                  }
                }
                break;
              }
              case "list": {
                marked.walkTokens(token.items, callback);
                break;
              }
              default: {
                if (marked.defaults.extensions && marked.defaults.extensions.childTokens && marked.defaults.extensions.childTokens[token.type]) {
                  marked.defaults.extensions.childTokens[token.type].forEach(function(childTokens) {
                    marked.walkTokens(token[childTokens], callback);
                  });
                } else if (token.tokens) {
                  marked.walkTokens(token.tokens, callback);
                }
              }
            }
          };
          for (var _iterator = _createForOfIteratorHelperLoose(tokens), _step; !(_step = _iterator()).done; ) {
            _loop3();
          }
        };
        marked.parseInline = function(src, opt) {
          if (typeof src === "undefined" || src === null) {
            throw new Error("marked.parseInline(): input parameter is undefined or null");
          }
          if (typeof src !== "string") {
            throw new Error("marked.parseInline(): input parameter is of type " + Object.prototype.toString.call(src) + ", string expected");
          }
          opt = merge({}, marked.defaults, opt || {});
          checkSanitizeDeprecation(opt);
          try {
            var tokens = Lexer.lexInline(src, opt);
            if (opt.walkTokens) {
              marked.walkTokens(tokens, opt.walkTokens);
            }
            return Parser.parseInline(tokens, opt);
          } catch (e) {
            e.message += "\nPlease report this to https://github.com/markedjs/marked.";
            if (opt.silent) {
              return "<p>An error occurred:</p><pre>" + escape(e.message + "", true) + "</pre>";
            }
            throw e;
          }
        };
        marked.Parser = Parser;
        marked.parser = Parser.parse;
        marked.Renderer = Renderer;
        marked.TextRenderer = TextRenderer;
        marked.Lexer = Lexer;
        marked.lexer = Lexer.lex;
        marked.Tokenizer = Tokenizer;
        marked.Slugger = Slugger;
        marked.parse = marked;
        var marked_1 = marked;
        return marked_1;
      });
    }
  });

  // node_modules/postman-to-openapi/lib/md-utils.js
  var require_md_utils = __commonJS({
    "node_modules/postman-to-openapi/lib/md-utils.js"(exports, module) {
      "use strict";
      var marked = require_marked();
      var supHeaders = ["object", "name", "description", "example", "type", "required"];
      function parseMdTable(md) {
        const parsed = marked.lexer(md);
        const table = parsed.find((el) => el.type === "table");
        if (table == null)
          return {};
        const { header, cells } = table;
        if (!header.includes("object") || !header.includes("name"))
          return {};
        const headers = header.map((h) => supHeaders.includes(h) ? h : false);
        const tableObj = cells.reduce((accTable, cell, i) => {
          const cellObj = cell.reduce((accCell, field, index) => {
            if (headers[index]) {
              accCell[headers[index]] = field;
            }
            return accCell;
          }, {});
          accTable[cellObj.name] = cellObj;
          return accTable;
        }, {});
        return tableObj;
      }
      module.exports = { parseMdTable };
    }
  });

  // node_modules/postman-to-openapi/package.json
  var require_package = __commonJS({
    "node_modules/postman-to-openapi/package.json"(exports, module) {
      module.exports = {
        name: "postman-to-openapi",
        version: "1.7.3",
        description: "Convert postman collection to OpenAPI spec",
        main: "lib/index.js",
        scripts: {
          lint: "eslint **/*.js",
          "lint:fix": "eslint **/*.js --fix",
          "test:unit": "mocha",
          "test:unit-no-only": "npm run test:unit -- --forbid-only",
          test: "nyc npm run test:unit-no-only",
          "changelog:all": "conventional-changelog --config ./changelog.config.js -i CHANGELOG.md -s -r 0",
          changelog: "conventional-changelog --config ./changelog.config.js -i CHANGELOG.md -s",
          prepare: "husky install"
        },
        repository: {
          type: "git",
          url: "git+https://github.com/joolfe/postman-to-openapi.git"
        },
        keywords: [
          "swagger",
          "OpenAPI",
          "postman",
          "collection",
          "convert",
          "converter",
          "transform",
          "specification",
          "yml"
        ],
        author: "joolfe04@gmail.com",
        license: "MIT",
        bugs: {
          url: "https://github.com/joolfe/postman-to-openapi/issues"
        },
        homepage: "https://github.com/joolfe/postman-to-openapi#readme",
        devDependencies: {
          "@commitlint/cli": "^12.1.1",
          "@commitlint/config-conventional": "^12.1.1",
          "conventional-changelog-cli": "^2.1.1",
          eslint: "^7.24.0",
          "eslint-config-standard": "^16.0.2",
          "eslint-plugin-import": "^2.22.1",
          "eslint-plugin-node": "^11.1.0",
          "eslint-plugin-promise": "^5.1.0",
          husky: "^6.0.0",
          mocha: "^8.3.2",
          nyc: "^15.1.0"
        },
        commitlint: {
          extends: [
            "@commitlint/config-conventional"
          ]
        },
        nyc: {
          all: true,
          include: [
            "lib/**/*.js",
            "test/**/*.js"
          ],
          exclude: [],
          reporter: [
            "lcovonly",
            "html",
            "text"
          ],
          lines: 90,
          statements: 90,
          functions: 90,
          branches: 90,
          "check-coverage": true
        },
        dependencies: {
          "js-yaml": "^4.1.0",
          marked: "^2.0.3"
        },
        husky: {
          hooks: {
            "commit-msg": "commitlint -E HUSKY_GIT_PARAMS"
          }
        }
      };
    }
  });

  // node_modules/postman-to-openapi/lib/index.js
  var require_lib = __commonJS({
    "node_modules/postman-to-openapi/lib/index.js"(exports, module) {
      "use strict";
      var { promises: { writeFile, readFile } } = require_fs();
      var { dump } = require_js_yaml();
      var { parseMdTable } = require_md_utils();
      var { version } = require_package();
      async function postmanToOpenApi(input, output, {
        info = {},
        defaultTag = "default",
        pathDepth = 0,
        auth,
        servers,
        externalDocs = {},
        folders = {}
      } = {}) {
        const collectionFile = await readFile(input);
        const postmanJson = JSON.parse(collectionFile);
        const { item: items, variable = [] } = postmanJson;
        const paths = {};
        const domains = /* @__PURE__ */ new Set();
        const tags = {};
        for (let [i, element] of items.entries()) {
          while (element.item != null) {
            const { item, description: tagDesc } = element;
            const tag2 = calculateFolderTag(element, folders);
            const tagged = item.map((e) => __spreadProps(__spreadValues({}, e), { tag: tag2 }));
            tags[tag2] = tagDesc;
            items.splice(i, 1, ...tagged);
            element = tagged.length > 0 ? tagged.shift() : items[i];
          }
          const {
            request: { url, method, body, description: rawDesc, header },
            name: summary,
            tag = defaultTag,
            event: events
          } = element;
          const { path, query, protocol, host, port } = scrapeURL(url);
          domains.add(calculateDomains(protocol, host, port));
          const joinedPath = calculatePath(path, pathDepth);
          if (!paths[joinedPath])
            paths[joinedPath] = {};
          const { description, paramsMeta } = descriptionParse(rawDesc);
          paths[joinedPath][method.toLowerCase()] = __spreadProps(__spreadValues(__spreadValues(__spreadValues({
            tags: [tag],
            summary
          }, description ? { description } : {}), parseBody(body, method)), parseParameters(query, header, joinedPath, paramsMeta)), {
            responses: parseResponse(events)
          });
        }
        const openApi = __spreadProps(__spreadValues(__spreadValues(__spreadValues(__spreadValues({
          openapi: "3.0.0",
          info: compileInfo(postmanJson, info)
        }, parseExternalDocs(variable, externalDocs)), parseServers(domains, servers)), parseAuth(postmanJson, auth)), parseTags(tags)), {
          paths
        });
        const openApiYml = dump(openApi, { skipInvalid: true });
        if (output != null) {
          await writeFile(output, openApiYml, "utf8");
        }
        return openApiYml;
      }
      function calculateFolderTag({ tag, name }, { separator = " > ", concat = true }) {
        return tag && concat ? `${tag}${separator}${name}` : name;
      }
      function compileInfo(postmanJson, optsInfo) {
        const { info: { name, description: desc }, variable = [] } = postmanJson;
        const ver = getVarValue(variable, "version", "1.0.0");
        const {
          title = name,
          description = desc,
          version: version2 = ver,
          termsOfService,
          license,
          contact
        } = optsInfo;
        return __spreadValues(__spreadValues(__spreadValues({
          title,
          description,
          version: version2
        }, termsOfService ? { termsOfService } : {}), parseContact(variable, contact)), parseLicense(variable, license));
      }
      function parseLicense(variables, optsLicense = {}) {
        const nameVar = getVarValue(variables, "license.name");
        const urlVar = getVarValue(variables, "license.url");
        const { name = nameVar, url = urlVar } = optsLicense;
        return name != null ? { license: __spreadValues({ name }, url ? { url } : {}) } : {};
      }
      function parseContact(variables, optsContact = {}) {
        const nameVar = getVarValue(variables, "contact.name");
        const urlVar = getVarValue(variables, "contact.url");
        const emailVar = getVarValue(variables, "contact.email");
        const { name = nameVar, url = urlVar, email = emailVar } = optsContact;
        return [name, url, email].some((e) => e != null) ? {
          contact: __spreadValues(__spreadValues(__spreadValues({}, name ? { name } : {}), url ? { url } : {}), email ? { email } : {})
        } : {};
      }
      function parseExternalDocs(variables, optsExternalDocs) {
        const descriptionVar = getVarValue(variables, "externalDocs.description");
        const urlVar = getVarValue(variables, "externalDocs.url");
        const { description = descriptionVar, url = urlVar } = optsExternalDocs;
        return url != null ? { externalDocs: __spreadValues({ url }, description ? { description } : {}) } : {};
      }
      function parseBody(body = {}, method) {
        if (["GET", "DELETE"].includes(method))
          return {};
        const { mode, raw, options = { raw: { language: "json" } } } = body;
        let content = {};
        switch (mode) {
          case "raw": {
            const { raw: { language } } = options;
            if (language === "json") {
              content = {
                "application/json": {
                  schema: {
                    type: "object",
                    example: raw ? JSON.parse(raw) : ""
                  }
                }
              };
            } else {
              content = {
                "application/json": {
                  schema: {
                    type: "string",
                    example: raw
                  }
                }
              };
            }
            break;
          }
          case "file":
            content = {
              "text/plain": {}
            };
            break;
        }
        return { requestBody: { content } };
      }
      function parseParameters(query = [], header, paths, paramsMeta = {}) {
        let parameters = header.reduce(mapParameters("header"), []);
        parameters = query.reduce(mapParameters("query"), parameters);
        parameters.push(...extractPathParameters(paths, paramsMeta));
        return parameters.length ? { parameters } : {};
      }
      function mapParameters(type) {
        return (parameters, { key, description, value }) => {
          const required = /\[required\]/gi.test(description);
          parameters.push(__spreadValues(__spreadValues(__spreadValues({
            name: key,
            in: type,
            schema: { type: inferType(value) }
          }, required ? { required } : {}), description ? { description: description.replace(/ ?\[required\] ?/gi, "") } : {}), value ? { example: value } : {}));
          return parameters;
        };
      }
      function extractPathParameters(path, paramsMeta) {
        const matched = path.match(/{\s*[\w-]+\s*}/g) || [];
        return matched.map(
          (match) => {
            const name = match.slice(1, -1);
            const { type = "string", description, example } = paramsMeta[name] || {};
            return __spreadValues(__spreadValues({
              name,
              in: "path",
              schema: { type },
              required: true
            }, description ? { description } : {}), example ? { example } : {});
          }
        );
      }
      function getVarValue(variables, name, def = void 0) {
        const variable = variables.find(({ key }) => key === name);
        return variable ? variable.value : def;
      }
      function inferType(value) {
        if (/^\d+$/.test(value))
          return "integer";
        if (/-?\d+\.\d+/.test(value))
          return "number";
        if (/^(true|false)$/.test(value))
          return "boolean";
        return "string";
      }
      function parseAuth({ auth }, optAuth) {
        if (optAuth != null) {
          return parseOptsAuth(optAuth);
        }
        return parsePostmanAuth(auth);
      }
      function parsePostmanAuth(postmanAuth = {}) {
        const { type } = postmanAuth;
        return type != null ? {
          components: {
            securitySchemes: {
              [type + "Auth"]: {
                type: "http",
                scheme: type
              }
            }
          },
          security: [{
            [type + "Auth"]: []
          }]
        } : {};
      }
      function parseOptsAuth(optAuth) {
        const securitySchemes = {};
        const security = [];
        for (const [secName, secDefinition] of Object.entries(optAuth)) {
          const _a = secDefinition, { type, scheme } = _a, rest = __objRest(_a, ["type", "scheme"]);
          if (type === "http" && ["bearer", "basic"].includes(scheme)) {
            securitySchemes[secName] = __spreadValues({
              type: "http",
              scheme
            }, rest);
            security.push({ [secName]: [] });
          }
        }
        return {
          components: { securitySchemes },
          security
        };
      }
      function calculatePath(paths = [], pathDepth) {
        paths = paths.slice(pathDepth);
        return "/" + paths.map((path) => path.replace(/([{}])\1+/g, "$1")).join("/");
      }
      function calculateDomains(protocol, hosts, port) {
        return protocol + "://" + hosts.join(".") + (port ? `:${port}` : "");
      }
      function scrapeURL(url) {
        if (typeof url === "string" || url instanceof String) {
          const objUrl = new URL(url);
          return {
            raw: url,
            path: decodeURIComponent(objUrl.pathname).slice(1).split("/"),
            query: [],
            protocol: objUrl.protocol.slice(0, -1),
            host: decodeURIComponent(objUrl.hostname).split("."),
            port: objUrl.port
          };
        }
        return url;
      }
      function parseServers(domains, serversOpts) {
        let servers;
        if (serversOpts != null) {
          servers = serversOpts.map(({ url, description }) => ({ url, description }));
        } else {
          servers = Array.from(domains).map((domain) => ({ url: domain }));
        }
        return servers.length > 0 ? { servers } : {};
      }
      function parseTags(tagsObj) {
        const tags = Object.entries(tagsObj).map(([name, description]) => ({ name, description }));
        return tags.length > 0 ? { tags } : {};
      }
      function descriptionParse(description) {
        if (description == null)
          return { description };
        const splitDesc = description.split(/# postman-to-openapi/gi);
        if (splitDesc.length === 1)
          return { description };
        return {
          description: splitDesc[0].trim(),
          paramsMeta: parseMdTable(splitDesc[1])
        };
      }
      function parseResponse(events = []) {
        let status = 200;
        const test = events.filter((event) => event.listen === "test");
        if (test.length > 0) {
          const script = test[0].script.exec.join();
          const result = script.match(/\.response\.code\)\.to\.eql\((\d{3})\)|\.to\.have\.status\((\d{3})\)/);
          status = result && result[1] != null ? result[1] : result && result[2] != null ? result[2] : status;
        }
        return {
          [status]: {
            description: "Successful response",
            content: {
              "application/json": {}
            }
          }
        };
      }
      postmanToOpenApi.version = version;
      module.exports = postmanToOpenApi;
    }
  });

  // index.js
  var import_postman_to_openapi = __toESM(require_lib());
  globalThis.myCustomP2OFunction = async (collectionString, optionsString) => {
    try {
      console.log("[P2O JS] Received collection string (first 100 chars):", collectionString.substring(0, 100));
      console.log("[P2O JS] Received options string:", optionsString);
      const collectionObject = JSON.parse(collectionString);
      console.log("[P2O JS] Parsed collection string to object.");
      let options = optionsString ? JSON.parse(optionsString) : {};
      options.replaceVars = true;
      options.operationId = "OPERATION_NAME";
      if (collectionObject && collectionObject.variable) {
        const baseUrlVar = collectionObject.variable.find((v) => v.key === "baseUrl");
        if (baseUrlVar && baseUrlVar.value) {
          options.servers = [{ url: baseUrlVar.value }];
          console.log("[P2O JS] Added explicit server URL to options:", baseUrlVar.value);
        }
      }
      console.log("[P2O JS] Parsed and modified options:", JSON.stringify(options));
      console.log("[P2O JS] Calling p2o with RAW COLLECTION STRING and NULL for output path.");
      const result = await (0, import_postman_to_openapi.default)(collectionString, null, options);
      console.log("[P2O JS] Conversion successful. Result type:", typeof result);
      if (typeof result === "string") {
        console.log("[P2O JS] Conversion result (first 200 chars):", result.substring(0, 200));
      } else {
        console.log("[P2O JS] Conversion result (not a string):", result);
      }
      return result;
    } catch (e) {
      console.error("[P2O JS] Error during conversion:", e);
      if (e.stack) {
        console.error("[P2O JS] Error stack:", e.stack);
      }
      throw new Error(`JavaScript conversion failed: ${e.message || String(e)}`);
    }
  };
  console.log("js: myCustomP2OFunction has been set on globalThis.");
  if (typeof globalThis.myCustomP2OFunction === "function") {
    console.log("js: typeof globalThis.myCustomP2OFunction is function.");
  } else {
    console.error("js: typeof globalThis.myCustomP2OFunction is NOT a function.");
  }
})();
