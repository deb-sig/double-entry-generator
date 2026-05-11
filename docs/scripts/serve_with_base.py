#!/usr/bin/env python3
"""在 /double-entry-generator/ 下提供 site 静态文件，便于与生产环境路径一致。"""
from __future__ import annotations

import sys
from functools import partial
from http.server import HTTPServer, SimpleHTTPRequestHandler
from pathlib import Path

BASE_PATH = "/double-entry-generator"
SITE_DIR = Path(__file__).resolve().parent.parent / "site"
PORT = 8000


class BasePathHandler(SimpleHTTPRequestHandler):
    def __init__(self, request, client_address, server, *, directory=None):
        self.base_path = BASE_PATH.rstrip("/")
        super().__init__(request, client_address, server, directory=directory or str(SITE_DIR))

    def translate_path(self, path):
        path = path.split("?", 1)[0].split("#", 1)[0]
        if path.startswith(self.base_path):
            path = path[len(self.base_path) :] or "/"
        if not path.startswith("/"):
            path = "/" + path
        return super().translate_path(path)

    def log_message(self, format, *args):
        sys.stderr.write("%s - - [%s] %s\n" % (self.address_string(), self.log_date_time_string(), format % args))


def main():
    if not SITE_DIR.is_dir():
        print("错误: site 目录不存在，请先执行 make build", file=sys.stderr)
        sys.exit(1)
    handler = partial(BasePathHandler, directory=str(SITE_DIR))
    server = HTTPServer(("", PORT), handler)
    print(f"预览地址: http://127.0.0.1:8000{BASE_PATH}/")
    print("(与生产路径一致，Ctrl+C 停止)")
    server.serve_forever()


if __name__ == "__main__":
    main()
