FROM dunglas/frankenphp:1.13.3-php8.5
ENV SERVER_NAME=":8080"
COPY public/ /app/public/
# El JIT de PHP 8 viene apagado en la imagen oficial (opcache.jit=disable). Esta carpeta solo se lee si el contenedor arranca
# con PHP_INI_SCAN_DIR=":/app/jit" (los dos puntos = «además de la carpeta de siempre»): así la misma imagen sirve con y sin JIT.
COPY jit/ /app/jit/
EXPOSE 8080
