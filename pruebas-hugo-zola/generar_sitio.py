# -*- coding: utf-8 -*-
"""Genera el MISMO sitio para Hugo y para Zola: N páginas Markdown sintéticas (semilla fija), 20 etiquetas, y plantillas
equivalentes (base, portada, sección, página, lista de etiquetas, etiqueta, 404). Lo único que cambia entre los dos árboles
es la sintaxis del front matter de las etiquetas y el lenguaje de plantillas (Go templates contra Tera).

    python3 generar_sitio.py <N> <destino>      → <destino>/hugo y <destino>/zola
"""
import os
import random
import sys
from datetime import date, timedelta

PAL = ("go rust servidor memoria hilo canal compilador binario prueba error tiempo red archivo proceso tarea cola carga "
       "contenedor imagen plantilla página sitio etiqueta lista mapa orden texto cliente puerto registro caché disco "
       "módulo paquete función método tipo valor puntero estructura interfaz rasgo ciclo vida préstamo dueño").split()
ETIQ = [f"tema-{i:02d}" for i in range(20)]


def frase(r, n):
    w = [r.choice(PAL) for _ in range(n)]
    return w[0].capitalize() + " " + " ".join(w[1:]) + "."


def cuerpo(r):
    b = []
    for s in range(4):
        b.append(f"## Sección {s + 1}: {frase(r, 3)[:-1]}\n")
        b.append(" ".join(frase(r, r.randint(8, 16)) for _ in range(5)) + "\n")
        b.append(f"Un párrafo con **negrita**, *cursiva*, `código` y un [enlace](https://example.com/{r.randint(1, 999)}).\n")
        if s == 1:
            b.append("\n".join(f"- {frase(r, 5)}" for _ in range(5)) + "\n")
        if s == 2:
            b.append("```go\nfunc main() {\n\tfmt.Println(\"hola\")\n}\n```\n")
        if s == 3:
            b.append("| clave | valor |\n|---|---|\n" + "\n".join(f"| {r.choice(PAL)} | {r.randint(1, 9999)} |" for _ in range(4)) + "\n")
    return "\n".join(b)


def escribir(ruta, texto):
    os.makedirs(os.path.dirname(ruta), exist_ok=True)
    with open(ruta, "w", encoding="utf-8", newline="\n") as f:
        f.write(texto)


HUGO_CFG = '''baseURL = "https://example.com/"
title = "Banco"
languageCode = "es"
enableRobotsTXT = true
titleCaseStyle = "none"
disableKinds = ["rss"]
[taxonomies]
  tag = "tags"
[markup.highlight]
  codeFences = false
[markup.goldmark.renderer]
  unsafe = false
'''

HUGO_T = {
    "layouts/baseof.html": '''<!doctype html>
<html lang="es"><head><meta charset="utf-8"><title>{{ .Title }} · {{ site.Title }}</title></head>
<body><nav><a href="/">Inicio</a> · <a href="/posts/">Posts</a> · <a href="/tags/">Etiquetas</a></nav>
<main>{{ block "main" . }}{{ end }}</main></body></html>
''',
    "layouts/home.html": '''{{ define "main" }}<h1>{{ site.Title }}</h1>
<ul>{{ range first 20 (where site.RegularPages "Section" "posts") }}<li><a href="{{ .Permalink }}">{{ .Title }}</a></li>{{ end }}</ul>{{ end }}
''',
    "layouts/section.html": '''{{ define "main" }}<h1>{{ .Title }}</h1>
<ul>{{ range .Pages }}<li><a href="{{ .Permalink }}">{{ .Title }}</a> <time>{{ .Date.Format "2006-01-02" }}</time></li>{{ end }}</ul>{{ end }}
''',
    "layouts/page.html": '''{{ define "main" }}<article><h1>{{ .Title }}</h1>
<p class="meta"><time>{{ .Date.Format "2006-01-02" }}</time> · {{ .WordCount }} palabras · {{ .ReadingTime }} min</p>
<p class="tags">{{ range .GetTerms "tags" }}<a href="{{ .Permalink }}">{{ .LinkTitle }}</a> {{ end }}</p>
<div class="contenido">{{ .Content }}</div></article>{{ end }}
''',
    "layouts/taxonomy.html": '''{{ define "main" }}<h1>Etiquetas</h1>
<ul>{{ range .Data.Terms.Alphabetical }}<li><a href="{{ .Page.Permalink }}">{{ .Page.LinkTitle }}</a> ({{ .Count }})</li>{{ end }}</ul>{{ end }}
''',
    "layouts/term.html": '''{{ define "main" }}<h1>{{ .Title }}</h1>
<ul>{{ range .Pages }}<li><a href="{{ .Permalink }}">{{ .Title }}</a></li>{{ end }}</ul>{{ end }}
''',
    "layouts/404.html": '''{{ define "main" }}<h1>No encontrado</h1>{{ end }}
''',
}

ZOLA_CFG = '''base_url = "https://example.com"
title = "Banco"
default_language = "es"
compile_sass = false
build_search_index = false
generate_feeds = false
taxonomies = [ { name = "tags" } ]
'''

ZOLA_T = {
    "templates/base.html": '''<!doctype html>
<html lang="es"><head><meta charset="utf-8"><title>{% block titulo %}{% endblock titulo %} · {{ config.title }}</title></head>
<body><nav><a href="/">Inicio</a> · <a href="/posts/">Posts</a> · <a href="/tags/">Etiquetas</a></nav>
<main>{% block main %}{% endblock main %}</main></body></html>
''',
    "templates/index.html": '''{% extends "base.html" %}{% block titulo %}{{ config.title }}{% endblock titulo %}{% block main %}<h1>{{ config.title }}</h1>
{% set posts = get_section(path="posts/_index.md") %}<ul>{% for p in posts.pages %}{% if loop.index <= 20 %}<li><a href="{{ p.permalink | safe }}">{{ p.title }}</a></li>{% endif %}{% endfor %}</ul>{% endblock main %}
''',
    "templates/section.html": '''{% extends "base.html" %}{% block titulo %}{{ section.title }}{% endblock titulo %}{% block main %}<h1>{{ section.title }}</h1>
<ul>{% for p in section.pages %}<li><a href="{{ p.permalink | safe }}">{{ p.title }}</a> <time>{{ p.date | date(format="%Y-%m-%d") }}</time></li>{% endfor %}</ul>{% endblock main %}
''',
    "templates/page.html": '''{% extends "base.html" %}{% block titulo %}{{ page.title }}{% endblock titulo %}{% block main %}<article><h1>{{ page.title }}</h1>
<p class="meta"><time>{{ page.date | date(format="%Y-%m-%d") }}</time> · {{ page.word_count }} palabras · {{ page.reading_time }} min</p>
<p class="tags">{% for t in page.taxonomies.tags %}<a href="{{ get_taxonomy_url(kind="tags", name=t) | safe }}">{{ t }}</a> {% endfor %}</p>
<div class="contenido">{{ page.content | safe }}</div></article>{% endblock main %}
''',
    "templates/taxonomy_list.html": '''{% extends "base.html" %}{% block titulo %}Etiquetas{% endblock titulo %}{% block main %}<h1>Etiquetas</h1>
<ul>{% for t in terms %}<li><a href="{{ t.permalink | safe }}">{{ t.name }}</a> ({{ t.pages | length }})</li>{% endfor %}</ul>{% endblock main %}
''',
    "templates/taxonomy_single.html": '''{% extends "base.html" %}{% block titulo %}{{ term.name }}{% endblock titulo %}{% block main %}<h1>{{ term.name }}</h1>
<ul>{% for p in term.pages %}<li><a href="{{ p.permalink | safe }}">{{ p.title }}</a></li>{% endfor %}</ul>{% endblock main %}
''',
    "templates/404.html": '''{% extends "base.html" %}{% block titulo %}404{% endblock titulo %}{% block main %}<h1>No encontrado</h1>{% endblock main %}
''',
}


def main():
    n, destino = int(sys.argv[1]), sys.argv[2]
    h, z = os.path.join(destino, "hugo"), os.path.join(destino, "zola")
    escribir(f"{h}/hugo.toml", HUGO_CFG)
    for k, v in HUGO_T.items():
        escribir(f"{h}/{k}", v)
    escribir(f"{z}/zola.toml", ZOLA_CFG)
    for k, v in ZOLA_T.items():
        escribir(f"{z}/{k}", v)
    escribir(f"{h}/content/posts/_index.md", '+++\ntitle = "Posts"\n+++\n')
    escribir(f"{z}/content/_index.md", '+++\ntitle = "Banco"\n+++\n')
    escribir(f"{z}/content/posts/_index.md", '+++\ntitle = "Posts"\nsort_by = "date"\n+++\n')
    r = random.Random(2026)
    d0 = date(2020, 1, 1)
    for i in range(n):
        titulo = f"Nota {i:06d}: {frase(r, 4)[:-1]}"
        dia = d0 + timedelta(days=i % 2000, seconds=0)
        fecha = f"{dia.isoformat()}T{(i // 2000) % 24:02d}:{i % 60:02d}:00Z"
        tags = sorted(r.sample(ETIQ, 3))
        texto = cuerpo(r)
        lt = ", ".join(f'"{t}"' for t in tags)
        escribir(f"{h}/content/posts/nota-{i:06d}.md", f'+++\ntitle = "{titulo}"\ndate = {fecha}\ntags = [{lt}]\n+++\n\n{texto}')
        escribir(f"{z}/content/posts/nota-{i:06d}.md",
                 f'+++\ntitle = "{titulo}"\ndate = {fecha}\n[taxonomies]\ntags = [{lt}]\n+++\n\n{texto}')


if __name__ == "__main__":
    main()
