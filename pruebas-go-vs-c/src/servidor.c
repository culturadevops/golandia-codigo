/* El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en C, para «¿Go o C?» (golandia 803).
 * El mismo contrato: /suma (CPU) y /trabajo (objetos, mapa, orden y JSON armado a mano).
 * HTTP con libmicrohttpd (GNU), un grupo de HILOS hilos (variable de entorno, 2 por defecto).
 * C no trae mapa en su librería estándar: la tabla hash de abajo está escrita a mano (direccionamiento abierto, FNV-1a). */
#include <microhttpd.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

static struct timespec arranque;

static int64_t ahora_ns(void) {
    struct timespec t;
    clock_gettime(CLOCK_MONOTONIC, &t);
    return (int64_t)t.tv_sec * 1000000000LL + t.tv_nsec;
}

typedef struct {
    int id;
    char *nombre;
    int valor;
} item;

/* ── la tabla hash: texto → item* ─────────────────────────────────────────────────────────── */
typedef struct {
    char **claves;
    item **valores;
    size_t cap;
} mapa;

static uint64_t fnv1a(const char *s) {
    uint64_t h = 1469598103934665603ULL;
    for (; *s; s++) {
        h ^= (unsigned char)*s;
        h *= 1099511628211ULL;
    }
    return h;
}

static void mapa_nuevo(mapa *m, size_t n) {
    m->cap = 16;
    while (m->cap < n * 2) m->cap <<= 1;
    m->claves = calloc(m->cap, sizeof(char *));
    m->valores = calloc(m->cap, sizeof(item *));
}

static void mapa_poner(mapa *m, char *k, item *v) {
    size_t i = fnv1a(k) & (m->cap - 1);
    while (m->claves[i] && strcmp(m->claves[i], k) != 0) i = (i + 1) & (m->cap - 1);
    m->claves[i] = k;
    m->valores[i] = v;
}

static void mapa_liberar(mapa *m) {
    free(m->claves);
    free(m->valores);
}

static int comparar(const void *a, const void *b) {
    const item *x = *(item *const *)a, *y = *(item *const *)b;
    if (x->valor != y->valor) return x->valor < y->valor ? -1 : 1;
    return (x->id > y->id) - (x->id < y->id);
}

static long num(struct MHD_Connection *c, const char *k) {
    const char *v = MHD_lookup_connection_value(c, MHD_GET_ARGUMENT_KIND, k);
    return v ? strtol(v, NULL, 10) : 0;
}

static char *suma(long n) {
    int64_t t0 = ahora_ns();
    uint64_t s = 0;
    for (long i = 1; i <= n; i++) s += (uint64_t)i * (uint64_t)i % 1000003;
    char *r = malloc(160);
    snprintf(r, 160, "{\"lang\":\"c\",\"n\":%ld,\"suma\":%llu,\"ns\":%lld}", n, (unsigned long long)s,
             (long long)(ahora_ns() - t0));
    return r;
}

static char *trabajo(long n) {
    int64_t t0 = ahora_ns();
    item *items = malloc(sizeof(item) * (n ? n : 1));
    mapa m;
    mapa_nuevo(&m, n);
    for (long i = 0; i < n; i++) {
        char buf[32];
        int len = snprintf(buf, sizeof buf, "item-%ld", i);
        items[i].id = (int)i;
        items[i].nombre = malloc(len + 1);
        memcpy(items[i].nombre, buf, len + 1);
        items[i].valor = (int)((i * 7919) % 1000);
        mapa_poner(&m, items[i].nombre, &items[i]);
    }
    item **lista = malloc(sizeof(item *) * (n ? n : 1));
    long k = 0, total = 0;
    for (size_t i = 0; i < m.cap; i++)
        if (m.claves[i]) {
            lista[k++] = m.valores[i];
            total += m.valores[i]->valor;
        }
    qsort(lista, k, sizeof(item *), comparar);
    size_t cap = 8192, len = 0;
    char *r = malloc(cap);
    len += snprintf(r + len, cap - len, "{\"lang\":\"c\",\"n\":%ld,\"total\":%ld,\"primeros\":[", n, total);
    for (long i = 0; i < 50 && i < k; i++)
        len += snprintf(r + len, cap - len, "%s{\"id\":%d,\"nombre\":\"%s\",\"valor\":%d}", i ? "," : "", lista[i]->id,
                        lista[i]->nombre, lista[i]->valor);
    for (long i = 0; i < n; i++) free(items[i].nombre);
    free(items);
    free(lista);
    mapa_liberar(&m);
    snprintf(r + len, cap - len, "],\"ns\":%lld}", (long long)(ahora_ns() - t0));
    return r;
}

static enum MHD_Result atender(void *cls, struct MHD_Connection *c, const char *url, const char *metodo,
                               const char *version, const char *datos, size_t *tam, void **estado) {
    (void)cls; (void)metodo; (void)version; (void)datos; (void)tam; (void)estado;
    char *cuerpo;
    unsigned int codigo = MHD_HTTP_OK;
    if (strcmp(url, "/suma") == 0) cuerpo = suma(num(c, "n"));
    else if (strcmp(url, "/trabajo") == 0) cuerpo = trabajo(num(c, "n"));
    else {
        cuerpo = strdup("{}");
        codigo = MHD_HTTP_NOT_FOUND;
    }
    struct MHD_Response *r = MHD_create_response_from_buffer(strlen(cuerpo), cuerpo, MHD_RESPMEM_MUST_FREE);
    MHD_add_response_header(r, "Content-Type", "application/json");
    enum MHD_Result ok = MHD_queue_response(c, codigo, r);
    MHD_destroy_response(r);
    return ok;
}

int main(void) {
    clock_gettime(CLOCK_MONOTONIC, &arranque);
    int64_t t0 = (int64_t)arranque.tv_sec * 1000000000LL + arranque.tv_nsec;
    const char *h = getenv("HILOS");
    unsigned int hilos = h ? (unsigned int)atoi(h) : 2;
    struct MHD_Daemon *d = MHD_start_daemon(MHD_USE_INTERNAL_POLLING_THREAD | MHD_USE_EPOLL, 8080, NULL, NULL,
                                            &atender, NULL, MHD_OPTION_THREAD_POOL_SIZE, hilos,
                                            MHD_OPTION_CONNECTION_LIMIT, 4096, MHD_OPTION_END);
    if (!d) {
        fprintf(stderr, "no arrancó\n");
        return 1;
    }
    fprintf(stderr, "LISTO_EN_MS %.3f hilos=%u\n", (ahora_ns() - t0) / 1e6, hilos);
    for (;;) pause();
}
