# La memoria según /proc/meminfo: MemFree frente a MemAvailable.
source /work/repo/comun.sh
c 'grep -E "^(MemTotal|MemFree|MemAvailable|Cached):" /proc/meminfo'
c 'ops-mem'
c 'free -h'
