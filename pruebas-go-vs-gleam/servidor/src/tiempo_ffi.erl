-module(tiempo_ffi).
-export([ahora_ns/0, programadores/0]).
ahora_ns() -> erlang:monotonic_time(nanosecond).
programadores() -> erlang:system_info(schedulers_online).
