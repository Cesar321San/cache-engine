package cli

import (
	"bufio"
	"cache-engine/internal/cache"
	"cache-engine/internal/persistence"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Run ejecuta la interfaz de línea de comandos
func Run(cacheEngine *cache.CacheEngine) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Custom Cache Engine CLI ===")
	fmt.Println("Comandos disponibles:")
	fmt.Println("  SET <key> <value>    - Establecer clave-valor")
	fmt.Println("  GET <key>            - Obtener valor")
	fmt.Println("  DEL <key>            - Eliminar clave")
	fmt.Println("  EXPIRE <key> <secs>  - Establecer expiración")
	fmt.Println("  ENABLELOG [archivo]  - Habilitar logging automático")
	fmt.Println("  DISABLELOG           - Deshabilitar logging automático")
	fmt.Println("  SAVE [archivo]       - Guardar estado actual a log")
	fmt.Println("  LOAD [archivo]       - Cargar desde log")
	fmt.Println("  STATS                - Mostrar estadísticas")
	fmt.Println("  EXIT                 - Salir")
	fmt.Println()

	for {
		fmt.Print("cache> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer entrada:", err)
			continue
		}

		// Limpiar entrada
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Separar comando y argumentos
		parts := strings.Fields(input)
		command := strings.ToUpper(parts[0])

		// Procesar comando
		switch command {
		case "SET":
			if len(parts) < 3 {
				fmt.Println("Error: Uso: SET <key> <value>")
				continue
			}
			key := parts[1]
			value := strings.Join(parts[2:], " ")
			cacheEngine.Set(key, value)

			// Log automático si está habilitado
			if logFile := getLogFile(cacheEngine); logFile != "" {
				persistence.LogOperation(logFile, "SET", key, value, 0)
			}

			fmt.Println("OK")

		case "GET":
			if len(parts) < 2 {
				fmt.Println("Error: Uso: GET <key>")
				continue
			}
			key := parts[1]
			value, exists := cacheEngine.Get(key)
			if !exists {
				fmt.Println("(nil)")
			} else {
				fmt.Printf("%v\n", value)
			}

		case "DEL":
			if len(parts) < 2 {
				fmt.Println("Error: Uso: DEL <key>")
				continue
			}
			key := parts[1]
			deleted := cacheEngine.Delete(key)
			if deleted {
				// Log automático si está habilitado
				if logFile := getLogFile(cacheEngine); logFile != "" {
					persistence.LogOperation(logFile, "DEL", key, nil, 0)
				}
				fmt.Println("OK")
			} else {
				fmt.Println("Clave no encontrada")
			}

		case "EXPIRE":
			if len(parts) < 3 {
				fmt.Println("Error: Uso: EXPIRE <key> <seconds>")
				continue
			}
			key := parts[1]
			seconds, err := strconv.Atoi(parts[2])
			if err != nil {
				fmt.Println("Error: segundos debe ser un número")
				continue
			}
			success := cacheEngine.Expire(key, seconds)
			if success {
				// Log automático si está habilitado
				if logFile := getLogFile(cacheEngine); logFile != "" {
					persistence.LogOperation(logFile, "EXPIRE", key, nil, time.Now().Unix()+int64(seconds))
				}
				fmt.Println("OK")
			} else {
				fmt.Println("Clave no encontrada")
			}

		case "ENABLELOG":
			filename := persistence.DefaultLogFile
			if len(parts) > 1 {
				filename = parts[1]
			}
			cacheEngine.EnableLogging(filename)
			fmt.Printf("Logging automático habilitado en: %s\n", filename)

		case "DISABLELOG":
			cacheEngine.DisableLogging()
			fmt.Println("Logging automático deshabilitado")

		case "SAVE":
			filename := persistence.DefaultLogFile
			if len(parts) > 1 {
				filename = parts[1]
			}
			if err := persistence.SaveToLog(cacheEngine, filename); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Guardado en %s\n", filename)
			}

		case "LOAD":
			filename := persistence.DefaultLogFile
			if len(parts) > 1 {
				filename = parts[1]
			}
			if err := persistence.LoadFromLog(cacheEngine, filename); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Cargado desde %s\n", filename)
			}

		case "STATS":
			fmt.Printf("Entradas en cache: %d\n", cacheEngine.Size())
			fmt.Printf("Límite máximo: %d\n", cacheEngine.MaxEntries())

		case "EXIT":
			fmt.Println("Cerrando cache engine...")
			cacheEngine.Close()
			return

		default:
			fmt.Println("Comando desconocido. Escribe EXIT para salir.")
		}
	}
}

// getLogFile obtiene el archivo de log actual del cache de forma segura
func getLogFile(c *cache.CacheEngine) string {
	return c.GetLogFile()
}
