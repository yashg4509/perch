#!/usr/bin/env python3
"""Minimal HTTP server for examples/scenarios/full-stack dev custom status checks.

Listens on 127.0.0.1:18080 and serves GET /health with 200 + body "ok".
Requires Python 3.6+ (stdlib only).

Usage (from any directory):

  python3 scripts/dev_health_server.py
"""

from http.server import BaseHTTPRequestHandler, HTTPServer

HOST = "127.0.0.1"
PORT = 18080


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        print("[%s] %s" % (self.log_date_time_string(), fmt % args))

    def do_GET(self):
        if self.path == "/health" or self.path.startswith("/health?"):
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b"ok\n")
            return
        self.send_response(404)
        self.end_headers()


if __name__ == "__main__":
    httpd = HTTPServer((HOST, PORT), Handler)
    print("perch example health server on http://%s:%s/health (Ctrl+C to stop)" % (HOST, PORT))
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\nstopped")
