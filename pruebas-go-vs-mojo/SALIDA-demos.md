# SALIDA-demos — ¿Go o Mojo? (golandia 904), sin cronómetro

> 2026-11-23 07:03. Regenerable con `python recursos/pruebas-go-vs-mojo/demos.py`.

## 1 · Versión y licencia

```
$ mojo --version; pip show mojo mojo-compiler | grep -E '^(Name|Version|License)'
Mojo 1.2.0 (2016d0cd)
Name: mojo
Version: 1.2.0
License: LicenseRef-MAX-Platform-Software-License
Name: mojo-compiler
Version: 1.2.0
License: LicenseRef-MAX-Platform-Software-License
```

```
$ curl https://raw.githubusercontent.com/modular/modular/main/LICENSE | head -7
Mojo and its contributions are licensed under the Apache License v2.0
with LLVM Exceptions. See Licenses/APACHE_LICENSE.

Portions of the MAX source code and its binary distribution are licensed
under the MAX License. See Licenses/MAX_LICENSE. The
remainder of the source is under the Apache License v2.0 with LLVM
Exceptions. See Licenses/APACHE_LICENSE.
```

## 2 · ¿Servidor web? El ejemplo de TCP de la documentación de std.net

```
$ mojo run tcp.mojo   # Mojo 1.2.0, la estable
/t/tcp.mojo:1:21: error: package 'net' does not contain 'SocketAddrV4'
from std.net import SocketAddrV4, TCPListener, TCPStream
                    ^
/t/tcp.mojo:1:35: error: package 'net' does not contain 'TCPListener'
from std.net import SocketAddrV4, TCPListener, TCPStream
                                  ^
/t/tcp.mojo:1:48: error: package 'net' does not contain 'TCPStream'
from std.net import SocketAddrV4, TCPListener, TCPStream
```

```
$ pip install -U --pre mojo --index-url https://whl.modular.com/nightly/simple/ --extra-index-url https://pypi.org/simple; mojo --version; mojo run tcp.mojo   # nightly
WARNING: Running pip as the 'root' user can result in broken permissions and conflicting behaviour with the system package manager, possibly rendering your system unusable. It is recommended to use a virtual environment instead: https://pip.pypa.io/warnings/venv. Use the --root-user-action option if you know what you are doing and want to suppress this warning.

Mojo 1.3.0.dev2026112105 (f23ed0af)
leídos 4 bytes
```

## 3 · lightbug_http, el framework web de la comunidad

```
$ curl https://api.github.com/repos/Lightbug-HQ/lightbug_http
archived: True
pushed_at: 2026-05-12T18:19:17Z
última versión: v26.2.0 (2026-03-29)
```

## 4 · Lo que necesita el binario para correr

```
$ ldd computo; du -sh .../site-packages/modular
linux-vdso.so.1
	libKGENCompilerRTShared.so => /usr/local/lib/python3.13/site-packages/modular/lib/libKGENCompilerRTShared.so
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
	libm.so.6 => /lib/x86_64-linux-gnu/libm.so.6
	libstdc++.so.6 => /lib/x86_64-linux-gnu/libstdc++.so.6
	libMSupportGlobals.so => /usr/local/lib/python3.13/site-packages/modular/lib/libMSupportGlobals.so
	libAsyncRTRuntimeGlobals.so => /usr/local/lib/python3.13/site-packages/modular/lib/libAsyncRTRuntimeGlobals.so
	libgcc_s.so.1 => /lib/x86_64-linux-gnu/libgcc_s.so.1
	/lib64/ld-linux-x86-64.so.2
670M	/usr/local/lib/python3.13/site-packages/modular
```

