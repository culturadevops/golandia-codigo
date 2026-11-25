# -*- coding: utf-8 -*-
"""cliente.py <orden> ... contra Azurite (AzureWebJobsStorage en el entorno).
  cola <nombre>                   crea la cola
  mandar <cola> <texto>           manda un mensaje en base64 (lo que espera el disparador de cola de Functions por defecto)
  blobs <contenedor>              lista los blobs y enseña el contenido de cada uno
  contar <cola>                   mensajes aproximados en la cola
"""
import base64
import os
import sys

from azure.storage.blob import BlobServiceClient
from azure.storage.queue import QueueClient

CS = os.environ["AzureWebJobsStorage"]
orden = sys.argv[1]
if orden == "cola":
    q = QueueClient.from_connection_string(CS, sys.argv[2])
    try:
        q.create_queue()
        print(f"cola {sys.argv[2]} creada")
    except Exception as e:  # ya existe
        print(f"cola {sys.argv[2]}: {type(e).__name__}")
elif orden == "mandar":
    q = QueueClient.from_connection_string(CS, sys.argv[2])
    m = q.send_message(base64.b64encode(sys.argv[3].encode()).decode())
    print(f"mensaje {m.id} en {sys.argv[2]}: {sys.argv[3]}")
elif orden == "blobs":
    c = BlobServiceClient.from_connection_string(CS).get_container_client(sys.argv[2])
    if not c.exists():
        print(f"el contenedor {sys.argv[2]} no existe")
    else:
        for b in c.list_blobs():
            print(f"{b.name}: {c.download_blob(b.name).readall().decode()}")
elif orden == "contar":
    q = QueueClient.from_connection_string(CS, sys.argv[2])
    print(q.get_queue_properties().approximate_message_count)
