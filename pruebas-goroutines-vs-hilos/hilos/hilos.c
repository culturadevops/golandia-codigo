// Lo mismo con hilos del sistema (pthreads): N hilos que se quedan esperando. Mide tiempo, memoria e hilos.
#include <pthread.h>
#include <semaphore.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <errno.h>

static sem_t espera, arrancados;

static int estado(const char *campo) {
    FILE *f = fopen("/proc/self/status", "r");
    char l[256]; int n = -1; size_t k = strlen(campo);
    while (fgets(l, sizeof l, f))
        if (!strncmp(l, campo, k) && l[k] == ':') { n = atoi(l + k + 1); break; }
    fclose(f);
    return n;
}

static void *trabajo(void *x) { sem_post(&arrancados); sem_wait(&espera); return NULL; }

int main(int argc, char **argv) {
    int n = atoi(argv[1]), creados = 0, err = 0;
    pthread_t *h = malloc(sizeof(pthread_t) * n);
    sem_init(&espera, 0, 0); sem_init(&arrancados, 0, 0);
    int rss0 = estado("VmRSS");
    struct timespec a, b; clock_gettime(CLOCK_MONOTONIC, &a);
    for (; creados < n; creados++)
        if ((err = pthread_create(&h[creados], NULL, trabajo, NULL)) != 0) break;
    for (int i = 0; i < creados; i++) sem_wait(&arrancados);
    clock_gettime(CLOCK_MONOTONIC, &b);
    double ms = (b.tv_sec - a.tv_sec) * 1e3 + (b.tv_nsec - a.tv_nsec) / 1e6;
    int rss = estado("VmRSS") - rss0;
    printf("hilos      n=%d crear=%.1fms memoria=%.1fMB porcada=%.2fKB hilos_del_SO=%d",
           creados, ms, rss / 1024.0, creados ? (double)rss / creados : 0, estado("Threads"));
    if (err) printf(" FALLO en el hilo %d: %s", creados + 1, strerror(err));
    printf("\n");
    for (int i = 0; i < creados; i++) sem_post(&espera);
    for (int i = 0; i < creados; i++) pthread_join(h[i], NULL);
    return 0;
}
