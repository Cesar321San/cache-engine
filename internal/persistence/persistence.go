package persistence

import (
	"cache-engine/internal/cache"
	"encoding/json"
	"fmt"
	"os"
)

// LogEntry representa una operación en el log
type LogEntry struct {
	Operation string      `json:"operation"` // SET, DEL, EXPIRE
	Key       string      `json:"key"`
	Value     interface{} `json:"value,omitempty"`
	ExpiresAt int64       `json:"expires_at,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

const (
	DefaultLogFile = "cache.log"
)

// SaveToLog guarda el estado actual del cache en formato JSON append-only
func SaveToLog(c *cache.CacheEngine, filename string) error {
	if filename == "" {
		filename = DefaultLogFile
	}

	// Abrir archivo en modo append
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir archivo de log: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	// Obtener datos de forma segura
	data := c.ExportData()

	// Escribir todas las entradas actuales
	for key, entry := range data {
		logEntry := LogEntry{
			Operation: "SET",
			Key:       key,
			Value:     entry.Value,
			ExpiresAt: entry.ExpiresAt,
			Timestamp: entry.LastAccess,
		}

		if err := encoder.Encode(logEntry); err != nil {
			return fmt.Errorf("error al escribir entrada: %v", err)
		}
	}

	return nil
}

// LoadFromLog carga el estado del cache desde el archivo de log
func LoadFromLog(c *cache.CacheEngine, filename string) error {
	if filename == "" {
		filename = DefaultLogFile
	}

	// Verificar si el archivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("archivo de log no existe: %s", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error al abrir archivo de log: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	// Leer y aplicar cada operación del log
	for {
		var logEntry LogEntry
		if err := decoder.Decode(&logEntry); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("error al leer entrada del log: %v", err)
		}

		// Aplicar operación según el tipo
		switch logEntry.Operation {
		case "SET":
			c.Set(logEntry.Key, logEntry.Value)
			if logEntry.ExpiresAt > 0 {
				// Calcular segundos restantes
				seconds := int(logEntry.ExpiresAt - logEntry.Timestamp)
				if seconds > 0 {
					c.Expire(logEntry.Key, seconds)
				}
			}
		case "DEL":
			c.Delete(logEntry.Key)
		case "EXPIRE":
			seconds := int(logEntry.ExpiresAt - logEntry.Timestamp)
			if seconds > 0 {
				c.Expire(logEntry.Key, seconds)
			}
		}
	}

	return nil
}

// Snapshot guarda un snapshot completo del estado actual
func Snapshot(c *cache.CacheEngine, filename string) error {
	if filename == "" {
		filename = "cache_snapshot.json"
	}

	// Obtener datos de forma segura
	data := c.ExportData()

	// Serializar a JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error al serializar snapshot: %v", err)
	}

	// Escribir a archivo
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("error al escribir snapshot: %v", err)
	}

	return nil
}

// LoadSnapshot carga un snapshot completo
func LoadSnapshot(c *cache.CacheEngine, filename string) error {
	if filename == "" {
		filename = "cache_snapshot.json"
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error al leer snapshot: %v", err)
	}

	snapshot := make(map[string]*cache.CacheEntry)
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("error al deserializar snapshot: %v", err)
	}

	// Necesitamos una forma de cargar los datos al cache
	// Por ahora, usamos Set para cada entrada
	for key, entry := range snapshot {
		c.Set(key, entry.Value)
		if entry.ExpiresAt > 0 {
			seconds := int(entry.ExpiresAt - entry.LastAccess)
			if seconds > 0 {
				c.Expire(key, seconds)
			}
		}
	}

	return nil
}
