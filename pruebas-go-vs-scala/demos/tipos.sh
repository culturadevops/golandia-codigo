#!/bin/sh
cp /d/demos/Estado.scala /tmp/ && cd /tmp
echo "### scala-cli compile Estado.scala"; scala-cli compile --server=false Estado.scala > /tmp/o 2>&1; c=$?; sed 's/[[0-9;]*m//g' /tmp/o | grep -vE '^(Downloading|Downloaded|Compiling|Compiled|Starting|Checking|Checked)'; echo "[código de salida: $c]"
echo "### scala-cli compile -Werror Estado.scala"; scala-cli compile --server=false -Werror Estado.scala > /tmp/o 2>&1; c=$?; sed 's/[[0-9;]*m//g' /tmp/o | grep -vE '^(Downloading|Downloaded|Compiling|Compiled|Starting|Checking|Checked)'; echo "[código de salida: $c]"
echo "### scala-cli run Estado.scala"; scala-cli run --server=false Estado.scala > /tmp/o 2>&1; c=$?; grep -vE '^(Downloading|Downloaded|Compiling|Compiled|Starting|Warning|  |\[warn\])' /tmp/o | tail -2; echo "[código de salida: $c]"
