#!/usr/bin/env python3
"""Two minimal HTTP servers for examples/scenarios/manual-cli-test (stdlib only).

Serves GET /health on:
  - 127.0.0.1:18081  (web)  — override with PERCH_MANUAL_WEB_PORT
  - 127.0.0.1:18082  (api)  — override with PERCH_MANUAL_API_PORT

If you change ports via env, update dev `status:` curls in perch.yaml to match.

Matches dev status commands in perch.yaml. Python 3.6+.

Usage (from this scenario directory or anywhere):

  python3 scripts/local_stack.py
"""

from http.server import BaseHTTPRequestHandler, HTTPServer
import errno
import os
import sys
import threading

BIND = "127.0.0.1"


def _port(env_name: str, default: int) -> int:
    raw = os.environ.get(env_name)
    if raw is None or raw == "":
        return default
    try:
        return int(raw, 10)
    except ValueError:
        print("error: %s must be an integer, got %r" % (env_name, raw), file=sys.stderr)
        sys.exit(1)


def server_specs():
    return (
        (_port("PERCH_MANUAL_WEB_PORT", 18081), "web"),
        (_port("PERCH_MANUAL_API_PORT", 18082), "api"),
    )


class ReuseHTTPServer(HTTPServer):
    """SO_REUSEADDR helps when restarting soon after Ctrl+C."""
    allow_reuse_address = True


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        label = getattr(self.server, "perch_label", "srv")
        print("[%s] %s %s" % (label, self.log_date_time_string(), fmt % args))

    def do_GET(self):
        if self.path == "/health" or self.path.startswith("/health?"):
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b"ok\n")
            return
        self.send_response(404)
        self.end_headers()


def _bind_fail_help(port: int, label: str) -> None:
    print("error: cannot bind %s:%d (%s) — address already in use." % (BIND, port, label), file=sys.stderr)
    print("  Usually another local_stack.py (or anything else) is still listening.", file=sys.stderr)
    print("  See what owns the port:", file=sys.stderr)
    print("    lsof -nP -iTCP:%d -sTCP:LISTEN" % port, file=sys.stderr)
    print("  Stop that process, or use PERCH_MANUAL_WEB_PORT / PERCH_MANUAL_API_PORT and the same ports in perch.yaml.", file=sys.stderr)


def main() -> None:
    specs = server_specs()
    httpds = []
    for port, label in specs:
        try:
            httpd = ReuseHTTPServer((BIND, port), Handler)
        except OSError as e:
            if e.errno in (errno.EADDRINUSE, getattr(errno, "WSAEADDRINUSE", -1)):
                _bind_fail_help(port, label)
                sys.exit(1)
            raise
        httpd.perch_label = label
        httpds.append(httpd)
        print("listening http://%s:%d/health (%s)" % (BIND, port, label))
    if os.environ.get("PERCH_MANUAL_WEB_PORT") or os.environ.get("PERCH_MANUAL_API_PORT"):
        print("note: custom ports — ensure perch.yaml dev status curls use these ports.", file=sys.stderr)

    for httpd in httpds:
        threading.Thread(target=httpd.serve_forever, name="perch-%s" % httpd.perch_label, daemon=True).start()

    print("Ctrl+C to stop")
    threading.Event().wait()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nstopped")
