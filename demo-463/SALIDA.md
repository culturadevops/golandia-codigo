$ cat datos   # lo que escribió el usuario
Comentario: <script>robar(document.cookie)</script>
Web:        javascript:robar(document.cookie)
Nombre:     Ana"; robar(document.cookie); var x="

$ go run . text   # text/template
<p><script>robar(document.cookie)</script></p>
<a href="javascript:robar(document.cookie)">mi web</a>
<script>var nombre = "Ana"; robar(document.cookie); var x="";</script>

$ go run . html   # html/template
<p>&lt;script&gt;robar(document.cookie)&lt;/script&gt;</p>
<a href="#ZgotmplZ">mi web</a>
<script>var nombre = "Ana\u0022; robar(document.cookie); var x=\u0022";</script>

$ diff <(go run . text) <(go run . html) | grep -c "^>"   # líneas que cambian
3
