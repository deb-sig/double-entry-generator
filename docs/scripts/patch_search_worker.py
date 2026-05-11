#!/usr/bin/env python3
"""构建后：用 CJK 友好索引重建 search_index.json，再对 worker.js/main.js 打补丁。"""
from __future__ import annotations

import sys
from pathlib import Path

# 确保使用本仓库的 vendor 插件
sys.path.insert(0, str(Path(__file__).resolve().parent.parent / "vendor" / "mkdocs-search-jieba"))

from mkdocs_search_jieba import patch_worker_in_dir, SearchJiebaPlugin

if __name__ == "__main__":
    site_dir = Path(__file__).resolve().parent.parent / "site"
    if len(sys.argv) > 1:
        site_dir = Path(sys.argv[1])
    r = SearchJiebaPlugin.rebuild_search_index_cjk(site_dir)
    if r > 0:
        print(f"已重建 {r} 个 search 索引（CJK 友好）")
    n = patch_worker_in_dir(site_dir)
    print(f"已修补 {n} 个 search worker")
    sys.exit(0)
