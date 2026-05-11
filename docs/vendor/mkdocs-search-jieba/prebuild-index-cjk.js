#!/usr/bin/env node
/**
 * 用「保留 CJK」的 trimmer 重建 lunr 索引，解决默认 trimmer 用 \\W 把中文整词删掉导致索引无中文的问题。
 * 用法: node prebuild-index-cjk.js <path-to-mkdocs-contrib-search>  < search_index_input.json > index_only.json
 * stdin: { "config": {...}, "docs": [...] }  （与 MkDocs 生成的 search_index 格式一致）
 * stdout: 仅 index 的 JSON，由插件合并回 search_index.json
 */
var path = require('path');
var searchDir = process.argv[2];
if (!searchDir) {
  process.stderr.write('Usage: node prebuild-index-cjk.js <mkdocs-contrib-search-dir>\n');
  process.exit(1);
}
var lunr = require(path.join(searchDir, 'templates', 'search', 'lunr.js'));

var stdin = process.stdin;
var buffer = [];

stdin.resume();
stdin.setEncoding('utf8');
stdin.on('data', function (data) { buffer.push(data); });
stdin.on('end', function () {
  var data;
  try {
    data = JSON.parse(buffer.join(''));
  } catch (e) {
    process.stderr.write('JSON parse error: ' + e.message + '\n');
    process.exit(1);
  }
  var lang = (data.config && data.config.lang && data.config.lang.length) ? data.config.lang : ['en'];
  if (lang.length > 1 || lang[0] !== 'en') {
    try {
      require(path.join(searchDir, 'lunr-language', 'lunr.stemmer.support.js'))(lunr);
      if (lang.length > 1) require(path.join(searchDir, 'lunr-language', 'lunr.multi.js'))(lunr);
      if (lang.indexOf('ja') >= 0 || lang.indexOf('jp') >= 0) require(path.join(searchDir, 'lunr-language', 'tinyseg.js'))(lunr);
      for (var i = 0; i < lang.length; i++) {
        if (lang[i] !== 'en') require(path.join(searchDir, 'lunr-language', 'lunr.' + lang[i] + '.js'))(lunr);
      }
    } catch (e) {}
  }
  if (data.config && data.config.separator && data.config.separator.length) {
    lunr.tokenizer.separator = new RegExp(data.config.separator);
  }

  lunr.trimmer = function (token) {
    return token.update(function (s) {
      var t = s.replace(/^\W+/, '').replace(/\W+$/, '');
      return (t.length > 0) ? t : s;
    });
  };

  var idx = lunr(function () {
    if (lang.length === 1 && lang[0] !== 'en' && lunr[lang[0]]) this.use(lunr[lang[0]]);
    else if (lang.length > 1) this.use(lunr.multiLanguage.apply(null, lang));
    this.field('title');
    this.field('text');
    this.ref('location');
    data.docs.forEach(function (doc) { this.add(doc); }, this);
  });

  process.stdout.write(JSON.stringify(idx));
});
