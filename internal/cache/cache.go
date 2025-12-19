package cache

import (
	"sync"
	"time"
)

// CacheEntry representa un valor almacenado en el cache
type CacheEntry struct {
	Value      interface{} // Valor almacenado
	ExpiresAt  int64       // Timestamp de expiración (0 = sin expiración)
	LastAccess int64       // Timestamp del último acceso (para LRU)
}

// CacheEngine es el motor principal del cache
type CacheEngine struct {
	data       map[string]*CacheEntry // Almacenamiento clave-valor
	mu         sync.RWMutex           // Mutex para concurrencia segura
	maxEntries int                    // Límite máximo de entradas (para LRU)
	stopClean  chan bool              // Canal para detener el barrido periódico
	logFile    string                 // Archivo de log para persistencia (opcional)
}

// NewCacheEngine crea una nueva instancia del motor de cache
func NewCacheEngine(maxEntries int) *CacheEngine {
	if maxEntries <= 0 {
		maxEntries = 1000 // Valor por defecto
	}

	cache := &CacheEngine{
		data:       make(map[string]*CacheEntry),
		maxEntries: maxEntries,
		stopClean:  make(chan bool),
	}

	// Iniciar barrido periódico de claves expiradas
	go cache.periodicCleanup()

	return cache
}

// Set almacena un valor en el cache
func (c *CacheEngine) Set(key string, value interface{}) {
	c.mu.Lock()

	// Si alcanzamos el límite, ejecutar eviction (LRU)
	if len(c.data) >= c.maxEntries {
		c.evictLRU()
	}

	now := time.Now().UnixNano() // Usar nanosegundos para mejor precisión
	c.data[key] = &CacheEntry{
		Value:      value,
		ExpiresAt:  0, // Sin expiración por defecto
		LastAccess: now,
	}

	// Registrar operación en log si está habilitado
	logFile := c.logFile
	c.mu.Unlock()

	if logFile != "" {
		// Importar persistence causaría dependencia circular, así que el logging
		// se maneja desde el CLI
	}
}

// Get obtiene un valor del cache
func (c *CacheEngine) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Verificar si la clave ha expirado
	now := time.Now().Unix()
	if entry.ExpiresAt > 0 && entry.ExpiresAt <= now {
		delete(c.data, key)
		return nil, false
	}

	// Actualizar último acceso (para LRU) usando nanosegundos
	entry.LastAccess = time.Now().UnixNano()
	return entry.Value, true
}

// Delete elimina una clave del cache
func (c *CacheEngine) Delete(key string) bool {
	c.mu.Lock()

	_, exists := c.data[key]
	if exists {
		delete(c.data, key)
	}

	// Registrar operación en log si está habilitado
	logFile := c.logFile
	c.mu.Unlock()

	if exists && logFile != "" {
		// El logging se maneja desde el CLI
	}

	return exists
}

// Expire establece un tiempo de expiración para una clave
func (c *CacheEngine) Expire(key string, seconds int) bool {
	c.mu.Lock()

	entry, exists := c.data[key]
	if !exists {
		c.mu.Unlock()
		return false
	}

	expiresAt := time.Now().Unix() + int64(seconds)
	entry.ExpiresAt = expiresAt

	// Registrar operación en log si está habilitado
	logFile := c.logFile
	c.mu.Unlock()

	if logFile != "" {
		// El logging se maneja desde el CLI
	}

	return true
}

// evictLRU elimina la entrada menos recientemente usada
func (c *CacheEngine) evictLRU() {
	var oldestKey string
	var oldestTime int64 = time.Now().UnixNano()

	// Buscar la clave con el acceso más antiguo
	for key, entry := range c.data {
		if entry.LastAccess < oldestTime {
			oldestTime = entry.LastAccess
			oldestKey = key
		}
	}

	// Eliminar la entrada más antigua
	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

// periodicCleanup ejecuta un barrido periódico para eliminar claves expiradas
func (c *CacheEngine) periodicCleanup() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanExpired()
		case <-c.stopClean:
			return
		}
	}
}

// cleanExpired elimina todas las claves expiradas
func (c *CacheEngine) cleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().Unix()
	for key, entry := range c.data {
		if entry.ExpiresAt > 0 && entry.ExpiresAt <= now {
			delete(c.data, key)
		}
	}
}

// Size retorna el número de entradas en el cache
func (c *CacheEngine) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// MaxEntries retorna el límite máximo de entradas
func (c *CacheEngine) MaxEntries() int {
	return c.maxEntries
}

// Close detiene los procesos en segundo plano
func (c *CacheEngine) Close() {
	close(c.stopClean)
}

// EnableLogging habilita el logging de operaciones en tiempo real
func (c *CacheEngine) EnableLogging(filename string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logFile = filename
}

// DisableLogging deshabilita el logging de operaciones
func (c *CacheEngine) DisableLogging() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logFile = ""
}

// GetLogFile retorna el archivo de log actual
func (c *CacheEngine) GetLogFile() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logFile
}

// ExportData retorna una copia segura de los datos para persistencia
func (c *CacheEngine) ExportData() map[string]*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	copy := make(map[string]*CacheEntry)
	for k, v := range c.data {
		// Hacemos una copia del puntero para evitar condiciones de carrera si se modifica el entry
		entryCopy := *v
		copy[k] = &entryCopy
	}
	return copy
}

// ImportData restaura datos masivamente (útil para snapshots)
func (c *CacheEngine) ImportData(data map[string]*CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = data
}
