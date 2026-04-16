package main

import (
	"flag"
	"log"
	"os"

	"github.com/max/wiki-llm/internal/extractor"
	"github.com/max/wiki-llm/internal/renderer"
	"github.com/max/wiki-llm/internal/writer"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < 1 {
		printRootHelp()
		return 1
	}

	switch args[0] {
	case "-h", "--help", "help":
		printRootHelp()
		return 0
	case "generate":
		return runGenerate(args[1:])
	default:
		log.Printf("wiki-llm: comando desconocido: %s\n", args[0])
		printRootHelp()
		return 1
	}
}

func runGenerate(args []string) int {
	if len(args) < 1 {
		printGenerateHelp()
		return 1
	}

	switch args[0] {
	case "-h", "--help", "help":
		printGenerateHelp()
		return 0
	case "api":
		return runGenerateAPI(args[1:])
	default:
		log.Printf("wiki-llm generate: subcomando desconocido: %s\n", args[0])
		printGenerateHelp()
		return 1
	}
}

func runGenerateAPI(args []string) int {
	fs := flag.NewFlagSet("wiki-llm generate api", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var source string
	var output string
	var showHelp bool

	fs.StringVar(&source, "source", "", "ruta al archivo OpenAPI")
	fs.StringVar(&output, "output", "", "directorio de salida")
	fs.BoolVar(&showHelp, "h", false, "muestra ayuda")
	fs.BoolVar(&showHelp, "help", false, "muestra ayuda")
	fs.Usage = printGenerateAPIHelp

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if showHelp {
		printGenerateAPIHelp()
		return 0
	}

	if fs.NArg() > 0 {
		log.Printf("wiki-llm generate api: argumentos no reconocidos: %v\n", fs.Args())
		printGenerateAPIHelp()
		return 1
	}

	if source == "" || output == "" {
		log.Println("wiki-llm generate api: --source y --output son obligatorios")
		printGenerateAPIHelp()
		return 1
	}

	doc, err := extractor.ExtractOpenAPI(source)
	if err != nil {
		log.Printf("wiki-llm generate api: error extrayendo openapi: %v\n", err)
		return 1
	}

	log.Printf("wiki-llm generate api: OpenAPI cargado correctamente")
	log.Printf("  source: %s", source)
	log.Printf("  output: %s", output)
	log.Printf("  title: %s", doc.Title)
	log.Printf("  version: %s", doc.Version)
	log.Printf("  endpoints extraidos: %d", len(doc.Endpoints))

	files := renderer.RenderAPI(doc)
	if err := writer.WriteFiles(output, files); err != nil {
		log.Printf("wiki-llm generate api: error escribiendo salida: %v\n", err)
		return 1
	}

	log.Printf("  archivos generados: %d", len(files))
	log.Println("wiki-llm generate api: generacion completada")
	return 0
}

func printRootHelp() {
	log.Println("wiki-llm: genera documentacion tecnica desde fuentes de API")
	log.Println("")
	log.Println("Uso:")
	log.Println("  wiki-llm <comando> [opciones]")
	log.Println("")
	log.Println("Comandos disponibles:")
	log.Println("  generate   Ejecuta comandos de generacion")
	log.Println("")
	log.Println("Opciones globales:")
	log.Println("  -h, --help Muestra esta ayuda")
}

func printGenerateHelp() {
	log.Println("Uso:")
	log.Println("  wiki-llm generate <subcomando> [opciones]")
	log.Println("")
	log.Println("Descripcion:")
	log.Println("  Comandos de generacion de documentacion.")
	log.Println("")
	log.Println("Subcomandos:")
	log.Println("  api        Genera documentacion para una API")
	log.Println("")
	log.Println("Opciones:")
	log.Println("  -h, --help Muestra esta ayuda")
}

func printGenerateAPIHelp() {
	log.Println("Uso:")
	log.Println("  wiki-llm generate api --source <archivo-openapi> --output <directorio>")
	log.Println("")
	log.Println("Descripcion:")
	log.Println("  Valida el contrato CLI para la generacion de API (bootstrap).")
	log.Println("")
	log.Println("Opciones:")
	log.Println("  --source   Ruta al archivo OpenAPI (obligatorio)")
	log.Println("  --output   Directorio de salida (obligatorio)")
	log.Println("  -h, --help Muestra esta ayuda")
}
