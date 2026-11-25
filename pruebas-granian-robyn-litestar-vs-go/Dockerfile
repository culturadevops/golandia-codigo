# Una imagen por framework: docker build --build-arg FW=fastapi -t golandia-rust-py-fastapi:1 .
# La misma base que el vídeo 770 de Cultura DevOps (python:3.13.17-slim). Versiones fijadas en req/<fw>.txt
# (pip freeze de la primera construcción; las directas en req/<fw>.in). FastAPI y Litestar traen uvicorn Y granian:
# la misma imagen y la misma app, cambia solo el servidor (SRV).
FROM python:3.13.17-slim
ARG FW
WORKDIR /app
COPY req/${FW}.txt /app/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt
COPY app/ /app/
ENV FW=${FW} WORKERS=2 PYTHONUNBUFFERED=1
EXPOSE 8080
ENTRYPOINT ["sh", "/app/arrancar.sh"]
