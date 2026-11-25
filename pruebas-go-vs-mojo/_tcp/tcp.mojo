from std.net import SocketAddrV4, TCPListener, TCPStream

def main() raises:
    var listener = TCPListener(bind=.v4(SocketAddrV4(.localhost, 0)))
    var client = TCPStream(connect=listener.local_addr())
    var server, _ = listener.accept()
    client.write_all("ping".as_bytes())
    var buffer = List[Byte](length=4, fill=0)
    var n = server.read(buffer)
    print("leídos", n, "bytes")
