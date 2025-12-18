package main

import (
	"cache-engine/internal/api/cli"
	"cache-engine/internal/cache"
	"flag"
	"fmt"
)

func main() {
	// Definir flags de línea de comandos
	maxEntries := flag.Int("max", 1000, "Número máximo de entradas en el cache")

	flag.Parse()

	// Crear instancia del cache
	cacheEngine := cache.NewCacheEngine(*maxEntries)

	fmt.Printf("Cache Engine iniciado (límite: %d entradas)\n", *maxEntries)
	fmt.Println("Modo: CLI")
	fmt.Println()

	// Ejecutar CLI
	cli.Run(cacheEngine)
}
