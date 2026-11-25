# Cliente de Azurite para el banco (crear la cola, mandar mensajes, leer los blobs que escribe el host).
FROM python:3.13.17-slim
RUN pip install --no-cache-dir azure-storage-queue==12.18.1 azure-storage-blob==12.31.1
COPY cliente.py /cliente.py
ENTRYPOINT ["python", "/cliente.py"]
