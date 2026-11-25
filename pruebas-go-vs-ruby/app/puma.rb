workers Integer(ENV.fetch("WORKERS", "0"))
threads 5, 5
bind "tcp://0.0.0.0:8080"
quiet
