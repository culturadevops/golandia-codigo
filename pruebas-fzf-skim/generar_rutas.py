# -*- coding: utf-8 -*-
"""Genera N rutas de archivo sintéticas, deterministas (semilla fija), con forma de monorepo real:
proyecto/carpeta/…/archivo.ext, de 2 a 8 niveles. Mismo N → mismo archivo, byte a byte.

    python3 generar_rutas.py <N> <destino>
"""
import random
import sys

PROYECTOS = ["api", "web", "backend", "frontend", "infra", "tools", "mobile", "billing", "auth", "search", "payments",
             "gateway", "worker", "scheduler", "analytics", "admin", "docs", "sdk", "cli", "platform"]
CARPETAS = ["src", "internal", "pkg", "cmd", "lib", "test", "tests", "vendor", "node_modules", "components", "handlers",
            "models", "services", "utils", "config", "migrations", "scripts", "assets", "static", "templates", "core",
            "middleware", "controllers", "views", "routes", "store", "hooks", "fixtures", "mocks", "generated", "proto",
            "client", "server", "storage", "cache", "queue", "events", "metrics", "logging", "security", "users",
            "orders", "invoices", "reports", "notifications", "dashboard", "settings", "i18n", "build", "dist"]
NOMBRES = ["main", "index", "app", "server", "client", "handler", "router", "config", "utils", "helpers", "types", "model",
           "service", "repository", "controller", "middleware", "auth", "user", "order", "invoice", "payment", "report",
           "cache", "queue", "worker", "logger", "metrics", "errors", "constants", "schema", "migration", "test_utils",
           "api", "routes", "store", "reducer", "actions", "selectors", "button", "modal", "table", "form", "input",
           "layout", "header", "footer", "sidebar", "dashboard", "settings", "profile", "login", "signup", "readme",
           "dockerfile", "makefile", "package", "tsconfig", "webpack", "vite", "jest", "eslint", "prettier", "go"]
EXT = [".go", ".rs", ".py", ".ts", ".tsx", ".js", ".jsx", ".json", ".yaml", ".yml", ".md", ".toml", ".sql", ".css",
       ".html", ".proto", ".sh", ".txt", "_test.go", ".spec.ts", ".lock", ".mod", ".sum"]


def main():
    n, destino = int(sys.argv[1]), sys.argv[2]
    r = random.Random(2026)
    with open(destino, "w", encoding="utf-8", newline="\n") as f:
        for i in range(n):
            partes = [r.choice(PROYECTOS)]
            for _ in range(r.randint(1, 7)):
                partes.append(r.choice(CARPETAS))
            nombre = r.choice(NOMBRES)
            if r.random() < 0.35:
                nombre += f"_{r.randint(1, 999)}"
            partes.append(nombre + r.choice(EXT))
            f.write("/".join(partes) + "\n")


if __name__ == "__main__":
    main()
