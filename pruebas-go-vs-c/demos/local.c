// Devolver la dirección de una variable local.
#include <stdio.h>

int *nuevo(int v) {
    int x = v;
    return &x;
}

int main(void) {
    int *a = nuevo(42);
    int *b = nuevo(7);
    printf("a=%d b=%d\n", *a, *b);
    return 0;
}
