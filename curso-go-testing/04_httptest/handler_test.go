package httpprueba

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// La prueba de handler de siempre: httptest.NewRecorder y llamar al handler directamente.
// Es rápida y está bien... pero no pasa por el mux: la ruta mal escrita no la ve.
func TestVerPedido_Handler(t *testing.T) {
	req := httptest.NewRequest("GET", "/pedidos/42", nil)
	req.SetPathValue("id", "42") // lo que haría el mux
	rec := httptest.NewRecorder()
	VerPedido(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("código %d", rec.Code)
	}
	t.Logf("handler directo: %d %s", rec.Code, rec.Body.String())
}
