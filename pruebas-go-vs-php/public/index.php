<?php
// El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en PHP 8.5 sobre FrankenPHP 1.13.3.
// El mismo contrato: /suma (CPU), /trabajo (objetos, array asociativo, orden y JSON a mano), /espera.
// Dos modos con el MISMO archivo: «clásico» (cada petición arranca el script desde cero, como PHP de toda la vida) y «worker»
// (el script se queda en memoria y atiende peticiones en bucle). Para «¿Go o PHP?» (2026-10-24): «golang vs php» / «go vs php».
$arranque = hrtime(true);

final class Item
{
    public function __construct(public int $id, public string $nombre, public int $valor) {}
}

function n(string $k): int
{
    return isset($_GET[$k]) ? (int)$_GET[$k] : 0;
}

function suma(): string
{
    $n = n('n');
    $t0 = hrtime(true);
    $s = 0;
    for ($i = 1; $i <= $n; $i++) {
        $s += ($i * $i) % 1000003;
    }
    return '{"lang":"php","n":' . $n . ',"suma":' . $s . ',"ns":' . (hrtime(true) - $t0) . '}';
}

function trabajo(): string
{
    $n = n('n');
    $t0 = hrtime(true);
    $m = [];
    for ($i = 0; $i < $n; $i++) {
        $it = new Item($i, 'item-' . $i, ($i * 7919) % 1000);
        $m[$it->nombre] = $it;
    }
    $lista = [];
    $total = 0;
    foreach ($m as $it) {
        $lista[] = $it;
        $total += $it->valor;
    }
    usort($lista, fn(Item $a, Item $b) => $a->valor !== $b->valor ? $a->valor <=> $b->valor : $a->id <=> $b->id);
    $b = '{"lang":"php","n":' . $n . ',"total":' . $total . ',"primeros":[';
    for ($i = 0; $i < 50 && $i < count($lista); $i++) {
        if ($i > 0) {
            $b .= ',';
        }
        $b .= '{"id":' . $lista[$i]->id . ',"nombre":"' . $lista[$i]->nombre . '","valor":' . $lista[$i]->valor . '}';
    }
    return $b . '],"ns":' . (hrtime(true) - $t0) . '}';
}

function atender(): void
{
    $ruta = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH);
    if ($ruta === '/suma') {
        $cuerpo = suma();
    } elseif ($ruta === '/trabajo') {
        $cuerpo = trabajo();
    } elseif ($ruta === '/espera') {
        $ms = n('ms');
        usleep($ms * 1000);
        $cuerpo = '{"lang":"php","n":' . $ms . ',"suma":0,"ns":0}';
    } else {
        http_response_code(404);
        echo 'no';
        return;
    }
    header('Content-Type: application/json');
    echo $cuerpo;
}

if (getenv('MODO') === 'worker' && function_exists('frankenphp_handle_request')) {
    error_log(sprintf("LISTO_EN_MS %.3f php %s worker\n", (hrtime(true) - $arranque) / 1e6, PHP_VERSION));
    while (frankenphp_handle_request('atender')) {
        gc_collect_cycles();
    }
} else {
    atender();
}
