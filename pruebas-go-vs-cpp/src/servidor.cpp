// El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en C++20, con cpp-httplib 0.63.0 (una sola cabecera,
// lo más parecido a net/http que hay en C++). El mismo contrato: /suma (CPU), /trabajo (objetos, mapa, orden y JSON a mano), /espera.
// Para «¿Go o C++?» (2026-10-24): «golang vs c++» / «golang vs c++ speed».
// Memoria a mano como haría alguien normal en C++ moderno: unique_ptr en el mapa, punteros crudos en la lista que se ordena.
#include "httplib.h"

#include <algorithm>
#include <chrono>
#include <cstdint>
#include <cstdio>
#include <memory>
#include <string>
#include <thread>
#include <unordered_map>
#include <vector>

using reloj = std::chrono::steady_clock;
static const auto ARRANQUE = reloj::now();

struct Item {
  int id;
  std::string nombre;
  int valor;
};

static long long ns_desde(reloj::time_point t0) {
  return std::chrono::duration_cast<std::chrono::nanoseconds>(reloj::now() - t0).count();
}

static int num(const httplib::Request& r, const char* k) {
  auto v = r.get_param_value(k);
  return v.empty() ? 0 : std::atoi(v.c_str());
}

int main() {
  httplib::Server srv;
  // HILOS = tamaño del pool. cpp-httplib atiende cada conexión abierta (keep-alive) con UN hilo del pool hasta que se cierra:
  // con 2 hilos solo hay 2 conexiones atendidas a la vez. Se mide con 2 (uno por núcleo) y con 32 (más que conexiones).
  const char* e = std::getenv("HILOS");
  unsigned hilos = e ? (unsigned)std::atoi(e) : std::max(2u, std::thread::hardware_concurrency());
  srv.new_task_queue = [hilos] { return new httplib::ThreadPool(hilos); };

  srv.Get("/suma", [](const httplib::Request& r, httplib::Response& w) {
    int n = num(r, "n");
    auto t0 = reloj::now();
    uint64_t s = 0;
    for (int i = 1; i <= n; i++) s += (uint64_t)i * (uint64_t)i % 1000003;
    w.set_content("{\"lang\":\"cpp\",\"n\":" + std::to_string(n) + ",\"suma\":" + std::to_string(s) +
                      ",\"ns\":" + std::to_string(ns_desde(t0)) + "}",
                  "application/json");
  });

  srv.Get("/trabajo", [](const httplib::Request& r, httplib::Response& w) {
    int n = num(r, "n");
    auto t0 = reloj::now();
    std::unordered_map<std::string, std::unique_ptr<Item>> m;
    m.reserve(n);
    for (int i = 0; i < n; i++) {
      auto it = std::make_unique<Item>(Item{i, "item-" + std::to_string(i), (i * 7919) % 1000});
      std::string clave = it->nombre;
      m.emplace(std::move(clave), std::move(it));
    }
    std::vector<Item*> lista;
    lista.reserve(m.size());
    long long total = 0;
    for (auto& [_, it] : m) {
      lista.push_back(it.get());
      total += it->valor;
    }
    std::sort(lista.begin(), lista.end(), [](const Item* a, const Item* b) {
      return a->valor != b->valor ? a->valor < b->valor : a->id < b->id;
    });
    std::string b = "{\"lang\":\"cpp\",\"n\":" + std::to_string(n) + ",\"total\":" + std::to_string(total) + ",\"primeros\":[";
    for (size_t i = 0; i < 50 && i < lista.size(); i++) {
      if (i > 0) b += ',';
      b += "{\"id\":" + std::to_string(lista[i]->id) + ",\"nombre\":\"" + lista[i]->nombre + "\",\"valor\":" +
           std::to_string(lista[i]->valor) + "}";
    }
    b += "],\"ns\":" + std::to_string(ns_desde(t0)) + "}";
    w.set_content(b, "application/json");
  });

  srv.Get("/espera", [](const httplib::Request& r, httplib::Response& w) {
    int ms = num(r, "ms");
    std::this_thread::sleep_for(std::chrono::milliseconds(ms));
    w.set_content("{\"lang\":\"cpp\",\"n\":" + std::to_string(ms) + ",\"suma\":0,\"ns\":0}", "application/json");
  });

  std::fprintf(stderr, "LISTO_EN_MS %.3f cpp hilos=%u\n", ns_desde(ARRANQUE) / 1e6, hilos);
  srv.listen("0.0.0.0", 8080);
}
