# MkDocs 搜索插件：使用 jieba 对中文建索引，便于 lunr 检索
from __future__ import annotations

import json
import logging
import re
import subprocess
from pathlib import Path

from mkdocs.plugins import BasePlugin

log = logging.getLogger(__name__)
ZWS = "\u200b"

# lunr 在加载时把 QueryLexer.termSeparator 设为当时的 tokenizer.separator，之后 worker 只改了
# tokenizer.separator，导致查询分词仍用默认分隔符、\u200b 不生效。构建后补一行同步 termSeparator。
WORKER_PATCH_SEARCH = "lunr.tokenizer.separator = new RegExp(data.config.separator);"
WORKER_PATCH_ADD = (
    "lunr.tokenizer.separator = new RegExp(data.config.separator);\n"
    "    lunr.QueryLexer.termSeparator = lunr.tokenizer.separator;"
)
# 索引用 jieba 在中文间插 \u200b，查询也需插 \u200b 再搜才能命中
WORKER_QUERY_PATCH_SEARCH = "postMessage({ results: search(e.data.query) });"
WORKER_QUERY_PATCH_ADD = (
    "var q = e.data.query; "
    "q = q.replace(/([\\u4e00-\\u9fff]{2})(?=[\\u4e00-\\u9fff])/g, '$1\\u200b'); "
    "try { var res = search(q); postMessage({ results: res }); } catch (err) { "
    "console.error('[Search Worker] lunr.search 报错', err); postMessage({ results: [] }); }"
)
# 兼容旧补丁：无分段 / 旧调试块 → 统一为上面分段版
WORKER_QUERY_PATCH_ADD_NO_SEGMENT = (
    "var q = e.data.query; "
    "console.log('[Search Worker] 查询', 'q=', JSON.stringify(q)); "
    "try { var res = search(q); "
    "console.log('[Search Worker] 结果数=', res.length, res.length ? res.slice(0,5).map(function(r){ return r.location + '|' + (r.title||''); }) : '(无)'); "
    "postMessage({ results: res }); } catch (err) { "
    "console.error('[Search Worker] lunr.search 报错', err); postMessage({ results: [] }); }"
)
WORKER_QUERY_OLD_DEBUG_FULL = (
    "var raw = e.data.query; var q = raw.replace(/([\\u4e00-\\u9fff]{2})(?=[\\u4e00-\\u9fff])/g, '$1\\u200b'); "
    "console.log('[Search Worker] 查询', 'raw=', JSON.stringify(raw), 'segmented=', JSON.stringify(q)); "
    "try { var res = search(q); "
    "console.log('[Search Worker] 结果数=', res.length); postMessage({ results: res }); } catch (err) { "
    "console.error('[Search Worker] lunr.search 报错', err); postMessage({ results: [] }); }"
)
WORKER_INDEX_DEBUG_SEARCH = "allowSearch = true;\n  postMessage({config: data.config});"
WORKER_INDEX_DEBUG_ADD = (
    "allowSearch = true;\n"
    "  console.log('[Search Worker] 索引就绪', 'separator=', (data.config && data.config.separator) || '', "
    "'doc数=', (data.docs && data.docs.length) || 0, "
    "'首条location=', data.docs[0] && data.docs[0].location, "
    "'首条text前80字符(含\\\\u200b?)=', data.docs[0] && data.docs[0].text ? JSON.stringify(data.docs[0].text.substring(0,80)) : '');\n"
    "  postMessage({config: data.config});"
)
# main.js 默认 min_search_length-1=2，导致 2 字（如「银行」）被拦截不发送，改为至少允许 2 字
MAIN_MIN_LENGTH_SEARCH = "min_search_length = e.data.config.min_search_length-1;"
MAIN_MIN_LENGTH_ADD = "min_search_length = Math.min(1, (e.data.config.min_search_length || 3) - 1);"

# HTML 里是 const base_url，无法在 main.js 里重赋值。改为用「当前脚本 URL」推导 search 目录，本地/生产都同源。
MAIN_ANCHOR = "function getSearchTermFromLocation()"
MAIN_SEARCH_BASE = (
    "var _searchBase=(function(){"
    "var s=document.currentScript&&document.currentScript.src;"
    "if(s)return s.replace(/\\/main\\.js$/i,'/');"
    "var b=typeof base_url!=='undefined'?base_url:'';"
    "return(b.slice(-1)==='/'?b:b+'/')+'search/';"
    "})();\n"
    "function getSearchTermFromLocation()"
)
MAIN_WORKER_URL_OLD = 'joinUrl(base_url, "search/worker.js")'
MAIN_WORKER_URL_NEW = '(_searchBase+"worker.js")'
# path 以 / 开头时原样返回会变成 origin+path，丢失 base 路径（如 /double-entry-generator），导致 404
MAIN_JOINURL_SEARCH = (
    "  if (path.substring(0, 1) === \"/\") {\n"
    "    // path starts with `/`. Thus it is absolute.\n"
    "    return path;\n"
    "  }"
)
MAIN_JOINURL_ADD = (
    "  if (path.substring(0, 1) === \"/\") {\n"
    "    return (base.replace(/\\/$/, '') || base) + path;\n"
    "  }"
)
# 曾误写成 /\\/$/ 导致 $ 被当 flag，需修正为 /\/$/
MAIN_JOINURL_BROKEN = "return (base.replace(/\\\\\\\\/$/, '') || base) + path;"
MAIN_JOINURL_FIXED = "return (base.replace(/\\/$/, '') || base) + path;"
# 移除曾误打的“重写 base_url”IIFE（const 不可重写，无效）
MAIN_OLD_BASEURL_IIFE = (
    "(function(){"
    "try{if(typeof base_url!=='undefined'&&/^https?:\\/\\//.test(base_url)){"
    "base_url=location.origin+new URL(base_url).pathname;}catch(e){}})();\n"
)
# 本地（localhost/127.0.0.1）时搜索结果链接用当前 origin，不跳转到线上
MAIN_LINK_BASE_ANCHOR = "})();\nfunction getSearchTermFromLocation()"
MAIN_LINK_BASE_INSERT = (
    "})();\n"
    "var _linkBase=(function(){"
    "var h=window.location.hostname;"
    "if(h!=='localhost'&&h!=='127.0.0.1')return typeof base_url!=='undefined'?base_url:window.location.origin+'/';"
    "try{return window.location.origin+(new URL(base_url).pathname.replace(/\\/$/,'')||'/');}catch(e){return typeof base_url!=='undefined'?base_url:window.location.origin+'/';}"
    "})();\n"
    "function getSearchTermFromLocation()"
)
MAIN_FORMATRESULT_OLD = "joinUrl(base_url, location)"
MAIN_FORMATRESULT_NEW = "joinUrl(_linkBase, location)"


def _segment_chinese(text: str) -> str:
    try:
        import jieba
    except ImportError:
        return text

    def replace_chinese(m: re.Match) -> str:
        return ZWS.join(jieba.cut(m.group(0)))

    return re.sub(r"[\u4e00-\u9fff]+", replace_chinese, text)


def _patched_add_entry(self, title: str | None, text: str, loc: str) -> None:
    text = text.replace("\u00a0", " ")
    text = re.sub(r"[ \t\n\r\f\v]+", " ", text.strip())
    text = _segment_chinese(text)
    if title:
        title = _segment_chinese(title)
    self._entries.append({"title": title, "text": text, "location": loc})


class SearchJiebaPlugin(BasePlugin):
    def on_config(self, config, **kwargs):
        try:
            import jieba
        except ImportError:
            log.warning("未安装 jieba，中文搜索将不可用。请运行: uv add jieba")
            return config

        from mkdocs.contrib.search import search_index

        if not getattr(search_index.SearchIndex, "_jieba_patched", False):
            search_index.SearchIndex._add_entry = _patched_add_entry
            search_index.SearchIndex._jieba_patched = True
            log.info("已启用搜索中文分词 (jieba)")
        return config

    def on_post_build(self, config, **kwargs):
        """构建后：用 CJK 友好索引重建，再修补 worker 与 main.js。"""
        site_dir = Path(config["site_dir"]).resolve()
        _rebuild_search_index_cjk(site_dir, log)
        _apply_search_patches(site_dir, log)

    @staticmethod
    def patch_worker_in_dir(site_dir: str | Path) -> int:
        """对指定 site 目录下的所有 search/worker.js 打补丁（可与构建后手动调用）。"""
        return patch_worker_in_dir(site_dir)

    @staticmethod
    def rebuild_search_index_cjk(site_dir: str | Path) -> int:
        """对指定 site 下所有 search_index.json 用 CJK 友好 trimmer 重建 index（可与构建后手动调用）。"""
        return _rebuild_search_index_cjk(Path(site_dir).resolve(), None)


def _rebuild_search_index_cjk(site_path: Path, log: logging.Logger | None = None) -> int:
    """用保留 CJK 的 trimmer 重建 search_index.json 中的 index，解决默认 \\W trim 把中文删光的问题。"""
    site_path = site_path.resolve()
    # 脚本在插件包根目录，__file__ 在 mkdocs_search_jieba/ 子目录下
    script = Path(__file__).resolve().parent.parent / "prebuild-index-cjk.js"
    if not script.is_file():
        if log:
            log.warning("prebuild-index-cjk.js 不存在，跳过索引重建")
        return 0
    try:
        import mkdocs.contrib.search as _search_mod
        search_dir = Path(_search_mod.__file__).resolve().parent
    except Exception:
        if log:
            log.warning("无法定位 mkdocs.contrib.search 目录，跳过索引重建")
        return 0
    rebuilt = 0
    for path in site_path.rglob("search_index.json"):
        try:
            raw = path.read_text(encoding="utf-8")
            data = json.loads(raw)
        except Exception as e:
            if log:
                log.warning("无法读取 %s: %s", path, e)
            continue
        if "docs" not in data or "config" not in data:
            continue
        inp = json.dumps({"config": data["config"], "docs": data["docs"]}, ensure_ascii=False)
        try:
            proc = subprocess.run(
                ["node", str(script), str(search_dir)],
                input=inp.encode("utf-8"),
                capture_output=True,
                timeout=120,
                cwd=str(search_dir),
            )
        except (OSError, subprocess.TimeoutExpired) as e:
            if log:
                log.warning("执行 prebuild-index-cjk.js 失败: %s", e)
            continue
        if proc.returncode != 0:
            if log:
                log.warning("prebuild-index-cjk.js 退出 %s: %s", proc.returncode, proc.stderr.decode("utf-8", errors="replace")[:200])
            continue
        try:
            new_index = json.loads(proc.stdout.decode("utf-8"))
        except json.JSONDecodeError as e:
            if log:
                log.warning("prebuild-index-cjk.js 输出非合法 JSON: %s", e)
            continue
        data["index"] = new_index
        path.write_text(json.dumps(data, ensure_ascii=False, separators=(",", ":")), encoding="utf-8")
        rebuilt += 1
    if rebuilt and log:
        log.info("已用 CJK 友好 trimmer 重建 %d 个 search 索引", rebuilt)
    return rebuilt


def _apply_search_patches(site_path: Path, log: logging.Logger | None = None) -> int:
    """对 site 下所有 search/ 打补丁：worker.js（QueryLexer、查询分段、索引日志）、main.js（_searchBase、min_search_length）。"""
    site_path = site_path.resolve()
    patched = 0
    for search_dir in site_path.rglob("worker.js"):
        if search_dir.parent.name != "search":
            continue
        search_dir = search_dir.parent
        # worker.js
        worker_path = search_dir / "worker.js"
        if worker_path.is_file():
            try:
                text = worker_path.read_text(encoding="utf-8")
            except Exception as e:
                if log:
                    log.warning("无法读取 worker.js %s: %s", worker_path, e)
                continue
            new_text = text
            if WORKER_PATCH_SEARCH in new_text and WORKER_PATCH_ADD not in new_text:
                new_text = new_text.replace(WORKER_PATCH_SEARCH, WORKER_PATCH_ADD, 1)
            if WORKER_QUERY_PATCH_ADD not in new_text:
                if WORKER_QUERY_PATCH_ADD_NO_SEGMENT in new_text:
                    new_text = new_text.replace(WORKER_QUERY_PATCH_ADD_NO_SEGMENT, WORKER_QUERY_PATCH_ADD, 1)
                elif WORKER_QUERY_OLD_DEBUG_FULL in new_text:
                    new_text = new_text.replace(WORKER_QUERY_OLD_DEBUG_FULL, WORKER_QUERY_PATCH_ADD, 1)
                elif WORKER_QUERY_PATCH_SEARCH in new_text:
                    new_text = new_text.replace(WORKER_QUERY_PATCH_SEARCH, WORKER_QUERY_PATCH_ADD, 1)
            if WORKER_INDEX_DEBUG_SEARCH in new_text and WORKER_INDEX_DEBUG_ADD not in new_text:
                new_text = new_text.replace(WORKER_INDEX_DEBUG_SEARCH, WORKER_INDEX_DEBUG_ADD, 1)
            if new_text != text:
                worker_path.write_text(new_text, encoding="utf-8")
                patched += 1
        # main.js
        main_path = search_dir / "main.js"
        if main_path.is_file():
            try:
                text = main_path.read_text(encoding="utf-8")
            except Exception:
                continue
            changed = False
            if MAIN_MIN_LENGTH_SEARCH in text and MAIN_MIN_LENGTH_ADD not in text:
                text = text.replace(MAIN_MIN_LENGTH_SEARCH, MAIN_MIN_LENGTH_ADD, 1)
                changed = True
            # 去掉旧版“重写 base_url”的 IIFE（const 不可重写，无效）
            if (MAIN_OLD_BASEURL_IIFE + MAIN_ANCHOR) in text:
                text = text.replace(MAIN_OLD_BASEURL_IIFE + MAIN_ANCHOR, MAIN_ANCHOR, 1)
                changed = True
            if MAIN_ANCHOR in text and MAIN_SEARCH_BASE not in text:
                text = text.replace(MAIN_ANCHOR, MAIN_SEARCH_BASE, 1)
                changed = True
            if MAIN_WORKER_URL_OLD in text:
                text = text.replace(MAIN_WORKER_URL_OLD, MAIN_WORKER_URL_NEW, 2)
                changed = True
            if MAIN_JOINURL_BROKEN in text:
                text = text.replace(MAIN_JOINURL_BROKEN, MAIN_JOINURL_FIXED, 1)
                changed = True
            elif MAIN_JOINURL_SEARCH in text and MAIN_JOINURL_ADD not in text:
                text = text.replace(MAIN_JOINURL_SEARCH, MAIN_JOINURL_ADD, 1)
                changed = True
            if MAIN_LINK_BASE_ANCHOR in text and "var _linkBase" not in text:
                text = text.replace(MAIN_LINK_BASE_ANCHOR, MAIN_LINK_BASE_INSERT, 1)
                changed = True
            if MAIN_FORMATRESULT_OLD in text and MAIN_FORMATRESULT_NEW not in text:
                text = text.replace(MAIN_FORMATRESULT_OLD, MAIN_FORMATRESULT_NEW, 1)
                changed = True
            if changed:
                main_path.write_text(text, encoding="utf-8")
                patched += 1
    if patched and log:
        log.info("已修补 search 资源 %d 处（worker + main.js）", patched)
    return patched


def patch_worker_in_dir(site_dir: str | Path) -> int:
    """对指定 site 目录下的 search/worker.js 与 main.js 打补丁。"""
    return _apply_search_patches(Path(site_dir).resolve(), None)
