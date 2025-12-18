package cache

import (
	"testing"
	"time"
)

// TestSetAndGet prueba las operaciones básicas SET y GET
func TestSetAndGet(t *testing.T) {
	cache := NewCacheEngine(10)
	defer cache.Close()

	// Test SET y GET
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")

	if !exists {
		t.Error("Esperaba que la clave existiera")
	}

	if value != "value1" {
		t.Errorf("Esperaba 'value1', obtuve '%v'", value)
	}
}

// TestGetNonExistent prueba obtener una clave inexistente
func TestGetNonExistent(t *testing.T) {
	cache := NewCacheEngine(10)
	defer cache.Close()

	_, exists := cache.Get("nonexistent")
	if exists {
		t.Error("No esperaba que la clave existiera")
	}
}

// TestDelete prueba la operación DELETE
func TestDelete(t *testing.T) {
	cache := NewCacheEngine(10)
	defer cache.Close()

	cache.Set("key1", "value1")
	deleted := cache.Delete("key1")

	if !deleted {
		t.Error("Esperaba que la clave fuera eliminada")
	}

	_, exists := cache.Get("key1")
	if exists {
		t.Error("La clave no debería existir después de eliminarla")
	}
}

// TestExpire prueba la expiración de claves
func TestExpire(t *testing.T) {
	cache := NewCacheEngine(10)
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Expire("key1", 1) // Expira en 1 segundo

	// Verificar que existe antes de expirar
	_, exists := cache.Get("key1")
	if !exists {
		t.Error("La clave debería existir antes de expirar")
	}

	// Esperar a que expire
	time.Sleep(2 * time.Second)

	// Verificar que ya no existe
	_, exists = cache.Get("key1")
	if exists {
		t.Error("La clave debería haber expirado")
	}
}

// TestLRUEviction prueba la expulsión LRU
func TestLRUEviction(t *testing.T) {
	cache := NewCacheEngine(3) // Solo 3 entradas
	defer cache.Close()

	// Agregar 3 elementos
	cache.Set("key1", "value1")
	time.Sleep(100 * time.Millisecond) // Asegurar diferencia de tiempo
	cache.Set("key2", "value2")
	time.Sleep(100 * time.Millisecond)
	cache.Set("key3", "value3")
	time.Sleep(100 * time.Millisecond)

	// Acceder a key1 y key2 para actualizarlas (key3 será la más antigua)
	cache.Get("key1")
	time.Sleep(100 * time.Millisecond)
	cache.Get("key2")
	time.Sleep(100 * time.Millisecond)

	// Agregar un cuarto elemento (debería expulsar key3 que es la menos usada)
	cache.Set("key4", "value4")

	// key3 debería haber sido expulsada por LRU
	_, exists := cache.Get("key3")
	if exists {
		t.Error("key3 debería haber sido expulsada por LRU")
	}

	// Las otras claves deberían existir
	_, exists = cache.Get("key1")
	if !exists {
		t.Error("key1 debería existir")
	}

	_, exists = cache.Get("key2")
	if !exists {
		t.Error("key2 debería existir")
	}

	_, exists = cache.Get("key4")
	if !exists {
		t.Error("key4 debería existir")
	}
}

// TestConcurrency prueba operaciones concurrentes
func TestConcurrency(t *testing.T) {
	cache := NewCacheEngine(100)
	defer cache.Close()

	done := make(chan bool)

	// Lanzar múltiples goroutines escribiendo
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				key := "key" + string(rune(n*10+j))
				cache.Set(key, n*10+j)
			}
			done <- true
		}(i)
	}

	// Lanzar múltiples goroutines leyendo
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				key := "key" + string(rune(n*10+j))
				cache.Get(key)
			}
			done <- true
		}(i)
	}

	// Esperar a que terminen todas
	for i := 0; i < 20; i++ {
		<-done
	}

	// Si llegamos aquí sin crash, el test pasó
}

// TestSize prueba el método Size
func TestSize(t *testing.T) {
	cache := NewCacheEngine(10)
	defer cache.Close()

	if cache.Size() != 0 {
		t.Error("El cache debería estar vacío inicialmente")
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	if cache.Size() != 2 {
		t.Errorf("Esperaba tamaño 2, obtuve %d", cache.Size())
	}

	cache.Delete("key1")

	if cache.Size() != 1 {
		t.Errorf("Esperaba tamaño 1, obtuve %d", cache.Size())
	}
}

// TestPeriodicCleanup prueba el barrido periódico
func TestPeriodicCleanup(t *testing.T) {
	cache := NewCacheEngine(10)
	defer cache.Close()

	// Agregar claves con expiración
	cache.Set("key1", "value1")
	cache.Expire("key1", 1)

	cache.Set("key2", "value2")
	cache.Expire("key2", 1)

	// Esperar a que el cleanup periódico las elimine
	time.Sleep(3 * time.Second)

	if cache.Size() != 0 {
		t.Errorf("El cache debería estar vacío después del cleanup, tiene %d entradas", cache.Size())
	}
}
