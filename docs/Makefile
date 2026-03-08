# 文档构建：先 mkdocs build，再对 search worker 打补丁（中文搜索需与索引分词一致）
# 可从仓库任意目录执行: make -C docs serve-static
.PHONY: build serve serve-static patch-search
SITE_DIR ?= site
DOCS_ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

build:
	cd $(DOCS_ROOT) && uv run mkdocs build
	cd $(DOCS_ROOT) && uv run python scripts/patch_search_worker.py $(SITE_DIR)

# mkdocs serve 启动时会再次构建，会覆盖已打补丁的 site，导致中文搜索失效
serve: build
	cd $(DOCS_ROOT) && uv run mkdocs serve

# 用静态服务器预览已构建+打补丁的 site（中文搜索有效），与生产路径一致
serve-static: build
	cd $(DOCS_ROOT) && python3 scripts/serve_with_base.py

# 仅打补丁（在已有 site 目录时）
patch-search:
	cd $(DOCS_ROOT) && uv run python scripts/patch_search_worker.py $(SITE_DIR)
