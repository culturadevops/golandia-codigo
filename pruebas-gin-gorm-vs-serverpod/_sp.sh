# corre el CLI de serverpod en dart:3.13.7 con un «flutter» de mentira: el CLI solo le pregunta la versión (ver _stub_flutter.log)
export PATH=$PATH:/root/.pub-cache/bin
mkdir -p /stub
printf '#!/bin/sh\necho "STUB flutter: $*" >> /w/_stub_flutter.log\necho "Flutter 3.44.0 • channel stable"\nexit 0\n' > /stub/flutter
chmod +x /stub/flutter
export PATH=/stub:$PATH
cd /w/gs_api/gs_api_server
"$@"
