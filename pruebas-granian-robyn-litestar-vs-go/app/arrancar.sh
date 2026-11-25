# Cada variante con su servidor. WORKERS procesos (por defecto 2: uno por CPU del contenedor).
#   SRV=uvicorn  uvicorn 0.54 (uvloop + httptools), el servidor del vídeo 770 de Cultura DevOps
#   SRV=granian  Granian 2.8 (servidor HTTP en Rust), interfaz ASGI, la misma app
#   SRV=robyn    Robyn 0.88 (runtime en Rust): --processes WORKERS --workers HILOS
set -e
case "$SRV" in
  uvicorn) exec uvicorn "srv_$FW:app" --host 0.0.0.0 --port 8080 --workers "$WORKERS" --log-level warning --no-access-log ;;
  granian) exec granian --interface asgi --host 0.0.0.0 --port 8080 --workers "$WORKERS" --log-level warning "srv_$FW:app" ;;
  robyn)   exec python srv_robyn.py --processes "$WORKERS" --workers "${HILOS:-1}" --log-level WARNING --disable-openapi ;;
  *)       echo "SRV desconocido: $SRV" >&2; exit 1 ;;
esac
