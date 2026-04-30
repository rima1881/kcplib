# kcplib
pythons kcp packages are not very compatible with go ones. this repo is to bring kcp power in go to python.
usage:
```python
from cffi import FFI
ffi = FFI()

ffi.cdef("""
    int KCPDial(const char* addr, const char* key, int dataShards, int parityShards);
    int KCPSend(int id, const char* data, int length);
    int KCPRecv(int id, char* buf, int maxLen);
    void KCPClose(int id);
""")

lib = ffi.dlopen("./kcplib.so")

conn_id = lib.KCPDial(b"your-server:8388", b"your-secret-key!", 10, 3)

msg = b"Hello!"
lib.KCPSend(conn_id, msg, len(msg))

buf = ffi.new("char[4096]")
n = lib.KCPRecv(conn_id, buf, 4096)
print("Received:", ffi.string(buf, n))

lib.KCPClose(conn_id)
```
