// Usar memoria después de liberarla. Compila sin una sola advertencia.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(void) {
    char *saldo = malloc(16);
    strcpy(saldo, "saldo=100");
    free(saldo);
    char *otro = malloc(16);
    strcpy(otro, "saldo=999999");
    printf("leo lo liberado: %s\n", saldo);
    free(otro);
    return 0;
}
