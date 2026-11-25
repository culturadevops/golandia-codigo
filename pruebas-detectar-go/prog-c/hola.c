/* El mismo hola mundo, en C, para tener con qué contrastar. */
#include <stdio.h>
#include <unistd.h>

int main(void) {
    char host[256];
    gethostname(host, sizeof(host));
    printf("{\"estado\":\"vivo\",\"host\":\"%s\"}\n", host);
    return 0;
}
