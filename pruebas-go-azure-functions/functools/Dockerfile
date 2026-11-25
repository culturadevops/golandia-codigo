# Azure Functions Core Tools (lo que instala `npm i -g azure-functions-core-tools@4`), versión fijada. Trae su propio host.
FROM node:22.21.1-bookworm-slim
RUN apt-get update -qq && apt-get install -y -qq --no-install-recommends libicu72 ca-certificates && rm -rf /var/lib/apt/lists/* \
 && npm i -g azure-functions-core-tools@4.17.0 --unsafe-perm true && func --version
