// Copiar lo que escribe el usuario en un arreglo de 8 bytes.
#include <stdio.h>
#include <string.h>

int main(int argc, char **argv) {
    (void)argc;
    struct {
        char nombre[8];
        int admin;
    } u = {"", 0};
    strcpy(u.nombre, argv[1]);
    printf("nombre=%s admin=%d\n", u.nombre, u.admin);
    return 0;
}
