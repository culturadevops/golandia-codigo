#!/usr/bin/env bash
# Dos pulsaciones de la barra espaciadora, de medio segundo cada una, como las haría una persona.
sleep 1.5
W=$(DISPLAY=:99 xdotool search --sync --onlyvisible --name "Lección 5" | head -1)
DISPLAY=:99 xdotool windowactivate --sync "$W" 2>/dev/null; DISPLAY=:99 xdotool windowfocus "$W" 2>/dev/null
for i in 1 2; do
  DISPLAY=:99 xdotool keydown space; sleep 0.5; DISPLAY=:99 xdotool keyup space; sleep 0.5
done
