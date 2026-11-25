#!/bin/sh
cd /d/demos
echo "### ruby -wc precio.rb   (revisa la sintaxis)"; ruby -wc precio.rb 2>&1; echo "[código de salida: $?]"
echo "### ruby precio.rb"; ruby precio.rb 2>&1; echo "[código de salida: $?]"
