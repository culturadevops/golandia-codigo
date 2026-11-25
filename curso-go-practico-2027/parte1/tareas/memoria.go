package main

import (
	"cmp"
	"context"
	"slices"
	"sync"
)

// Memoria guarda las tareas en un map. Se pierde todo al reiniciar.
type Memoria struct {
	mu     sync.Mutex
	ultimo int64
	tareas map[int64]Tarea
}

func NuevaMemoria() *Memoria {
	return &Memoria{tareas: map[int64]Tarea{}}
}

func (m *Memoria) Listar(ctx context.Context) ([]Tarea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	lista := make([]Tarea, 0, len(m.tareas))
	for _, t := range m.tareas {
		lista = append(lista, t)
	}
	slices.SortFunc(lista, func(a, b Tarea) int { return cmp.Compare(a.ID, b.ID) })
	return lista, nil
}

func (m *Memoria) Crear(ctx context.Context, titulo string) (Tarea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ultimo++
	t := Tarea{ID: m.ultimo, Titulo: titulo}
	m.tareas[t.ID] = t
	return t, nil
}

func (m *Memoria) Obtener(ctx context.Context, id int64) (Tarea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tareas[id]
	if !ok {
		return Tarea{}, ErrNoEncontrada
	}
	return t, nil
}

func (m *Memoria) Marcar(ctx context.Context, id int64, hecha bool) (Tarea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tareas[id]
	if !ok {
		return Tarea{}, ErrNoEncontrada
	}
	t.Hecha = hecha
	m.tareas[id] = t
	return t, nil
}

func (m *Memoria) Borrar(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tareas[id]; !ok {
		return ErrNoEncontrada
	}
	delete(m.tareas, id)
	return nil
}
