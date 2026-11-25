FROM python:3.13.17-slim AS build
RUN apt-get update -qq && apt-get install -y -qq --no-install-recommends gcc libc6-dev >/dev/null
RUN pip install --no-cache-dir -q mojo==1.2.0
COPY src /src
RUN mojo build /src/computo.mojo -o /computo \
 && mkdir /libs && cd /usr/local/lib/python3.13/site-packages/modular/lib \
 && cp libKGENCompilerRTShared.so libMSupportGlobals.so libAsyncRTRuntimeGlobals.so /libs/
FROM debian:trixie-slim AS final
COPY --from=build /libs /opt/mojo/lib
COPY --from=build /computo /computo
ENV LD_LIBRARY_PATH=/opt/mojo/lib
ENTRYPOINT ["/computo"]
FROM final AS medir
RUN apt-get update -qq && apt-get install -y -qq time >/dev/null && rm -rf /var/lib/apt/lists/*
ENTRYPOINT []
