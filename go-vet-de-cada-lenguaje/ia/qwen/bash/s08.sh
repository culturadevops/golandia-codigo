#!/bin/bash
ORIGEN="$1"
DESTINO="$2"
find "$ORIGEN" -name "*.log" -type f -mtime +7 -exec gzip {} \; -exec mv --backup=t --target-directory="$DESTINO" {}*.gz \;
