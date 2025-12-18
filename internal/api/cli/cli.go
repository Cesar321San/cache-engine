package cli

import (
	"bufio"
	"cache-engine/internal/cache"
	"cache-engine/internal/persistence"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	fmt.Println("  SAVE [archivo]       - Guardar a log")
	fmt.Println("  LOAD [archivo]       - Cargar desde log")
	fmt.Println("  SNAPSHOT [archivo]   - Guardar snapshot")
	fmt.Println("  LOADSNAPSHOT [archivo] - Cargar snapshot")
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
				fmt.Println("OK")
			} else {
				fmt.Println("Clave no encontrada")
			}

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

		case "SNAPSHOT":
			filename := "cache_snapshot.json"
			if len(parts) > 1 {
				filename = parts[1]
			}
			if err := persistence.Snapshot(cacheEngine, filename); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Snapshot guardado en %s\n", filename)
			}

		case "LOADSNAPSHOT":
			filename := "cache_snapshot.json"
			if len(parts) > 1 {
				filename = parts[1]
			}
			if err := persistence.LoadSnapshot(cacheEngine, filename); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Snapshot cargado desde %s\n", filename)
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
