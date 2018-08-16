#!/usr/bin/python3
"""
usage: ./github-webhook-server.py [<port>]
"""
from http.server import BaseHTTPRequestHandler, HTTPServer
import json
import subprocess

class S(BaseHTTPRequestHandler):
    def _set_headers(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()

    def do_GET(self):
        self._set_headers()
        self.wfile.write('"OK"'.encode('utf-8'))

    def do_HEAD(self):
        self._set_headers()

    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        body = self.rfile.read(content_length).decode('utf-8')
        data = json.loads(body)
        repository = data.get('repository')
        if repository != None and repository.get('full_name') == 'karmakaze/quicklog':
            if data.get('ref') == 'refs/heads/master':
                exec_cmd("git pull -r")
                exec_cmd("make rebuild")
                exec_cmd("sudo service quicklog restart")
        self._set_headers()
        self.wfile.write('"OK"'.encode('utf-8'))

def exec_cmd(cmdline):
    print(cmdline)
    subprocess.call(cmdline.split(' '))

def run(server_class=HTTPServer, handler_class=S, port=8000):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print('Starting httpd...')
    httpd.serve_forever()

if __name__ == "__main__":
    from sys import argv

    if len(argv) == 2:
        run(port=int(argv[1]))
    else:
        run(port=8954)

run()
