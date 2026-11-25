# El servicio más pequeño posible: responde «ok» en /salud. Solo biblioteca estándar.
from http.server import BaseHTTPRequestHandler, HTTPServer
class H(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200); self.end_headers(); self.wfile.write(b"ok\n")
HTTPServer(("", 8080), H).serve_forever()
