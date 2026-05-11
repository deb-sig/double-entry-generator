# 文档构建说明

在 `docs/` 目录下执行：

1. **安装依赖**（含中文搜索插件）：`uv sync`
2. **构建**：`make build`（推荐，会顺带修补中文搜索）或 `uv run mkdocs build`
3. **本地预览**：**测试中文搜索请用** `make serve-static`（在 `docs/` 下执行；或仓库根目录执行 `make -C docs serve-static`），浏览器打开 **[http://127.0.0.1:8000/double-entry-generator/](http://127.0.0.1:8000/double-entry-generator/)**（与生产路径一致，如 `/configuration/accounts/` 等子路径均可直接访问）。若用 `uv run mkdocs serve`，其启动时会重新构建并覆盖补丁，搜索会无结果。

### 线上预览与验证（GitHub Pages）

- **线上地址**：推送并合并到 `main` / `master` 或 `feat/docs` 后，GitHub Actions 会构建并部署到 GitHub Pages。  
  - 本仓库：**https://deb-sig.github.io/double-entry-generator/**  
  - 若为 fork：`https://<你的用户名>.github.io/double-entry-generator/`

- **如何触发部署**  
  - 修改 `docs/**`、`docs/mkdocs.yml`、`docs/pyproject.toml` 或 `.github/workflows/docs.yml` 后 **push**，会触发构建。  
  - 会执行“部署”到 GitHub Pages 的分支：**main**、**master**、**feat/docs**、**docs/i18n-and-quickstart**。其他分支只跑构建（可在 Actions 里看是否成功）。  
  - 在 fork 上推 **docs/i18n-and-quickstart** 后，你的 Pages 地址会显示该分支的文档，便于给上游 PR 时贴“演示链接”。

- **如何验证线上效果**  
  1. 打开上述线上地址，看首页、侧栏导航是否正常。  
  2. 点「中文 / English」切换，检查 `/` 与 `/en/` 下对应页面是否存在、链接是否一致。  
  3. 点侧栏若干页（如 配置指南 → 账户映射、提供商 → 某银行），确认无 404。  
  4. 用站内搜索试几个关键词（含中文），看是否有结果、点结果是否跳转到正确页面（且不跳到错误域名）。

- `search_jieba` 插件通过本地 path 依赖 `vendor/mkdocs-search-jieba` 引入，`uv sync` 后即可使用。
- **中文搜索**：默认 lunr trimmer 用 `\W` 会把中文整词删掉导致索引无中文；`**make build`** 会先用 CJK 友好 trimmer 重建索引，再对 worker/main.js 打补丁（QueryLexer 同步、查询分段、min_search_length）。仅用 `mkdocs build` 时需再执行 `**make patch-search**`（会重建索引并打补丁）。
- **搜索索引**：`prebuild_index: true` 时由 Node 在构建时预生成索引，首屏即可搜、无需在浏览器内建索引。若改为 `false`，首搜时浏览器会先建索引（通常 1–3 秒，文档量越大越慢），之后搜索速度与预构建差不多；适合不想装 Node 或构建环境受限时使用。

### 搜索排错（无结果时）

1. **看 Worker 日志**（F12 控制台）：索引就绪时应看到 `separator= [\s\u200b\-]+`、`doc数`；查询时会执行 `lunr.search`，若报错会打印 `lunr.search 报错`。查询在 worker 内会自动插 `\u200b` 与索引一致。
2. **确认索引含中文**：控制台执行 `fetch(location.pathname.replace(/\/[^/]*$/, '') + '/search/search_index.json').then(r=>r.json()).then(d=>console.log('doc数', d.docs?.length, '首条text前80', d.docs?.[0]?.text?.slice(0,80)))`，首条 text 应含 `\u200b`。
3. **仍无结果**：确认已执行 `make build` 或 `make patch-search`（会重建 CJK 索引并打补丁）；若用 `mkdocs serve` 会覆盖补丁，请用 `make serve-static` 预览。

