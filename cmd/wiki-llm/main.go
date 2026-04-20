package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/max/wiki-llm/internal/discoverer"
	"github.com/max/wiki-llm/internal/extractor"
	"github.com/max/wiki-llm/internal/normalizer"
	"github.com/max/wiki-llm/internal/renderer"
	"github.com/max/wiki-llm/internal/sourcecode"
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

	var sources stringSliceFlag
	var sourceType string
	var codePath string
	var output string
	var showHelp bool

	fs.Var(&sources, "source", "ruta al archivo de fuente (repetible)")
	fs.StringVar(&sourceType, "source-type", extractor.SourceTypeAuto, "tipo de fuente: auto|openapi|postman")
	fs.StringVar(&codePath, "code", "", "ruta raiz del proyecto API para descubrir swagger.json")
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

	if output == "" {
		log.Println("wiki-llm generate api: --output es obligatorio")
		printGenerateAPIHelp()
		return 1
	}

	log.Println("wiki-llm generate api: validando entradas...")
	resolvedSources, discoveryLogs, err := discoverer.ResolveGenerateAPISources(codePath, sources, sourceType)
	if err != nil {
		log.Printf("wiki-llm generate api: error resolviendo fuentes: %v\n", err)
		return 1
	}
	for _, line := range discoveryLogs {
		log.Printf("wiki-llm generate api: %s", line)
	}
	for _, source := range resolvedSources {
		if err := validateSourcePath(source.Path); err != nil {
			log.Printf("wiki-llm generate api: validacion de source fallo (%s): %v\n", source.Path, err)
			return 1
		}
	}
	if err := writer.ValidateOutputDir(output); err != nil {
		log.Printf("wiki-llm generate api: validacion de output fallo: %v\n", err)
		return 1
	}

	log.Printf("wiki-llm generate api: extrayendo fuentes (type=%s)...", sourceType)
	documents := make([]normalizer.APIDocument, 0, len(resolvedSources))
	sourcePaths := make([]string, 0, len(resolvedSources))
	for _, source := range resolvedSources {
		doc, err := extractor.ExtractSource(source.Path, source.SourceType)
		if err != nil {
			log.Printf("wiki-llm generate api: error extrayendo fuente (%s): %v\n", source.Path, err)
			return 1
		}
		documents = append(documents, doc)
		sourcePaths = append(sourcePaths, source.Path)
	}
	doc := normalizer.MergeDocuments(documents)
	doc.Title = resolveDocumentTitle(doc, codePath)

	if strings.TrimSpace(codePath) != "" {
		log.Println("wiki-llm generate api: analizando codigo fuente para enriquecer endpoints...")
		enriched, err := sourcecode.EnrichDocument(doc, codePath)
		if err != nil {
			log.Printf("wiki-llm generate api: error analizando codigo fuente: %v\n", err)
			return 1
		}
		doc = enriched
	}
	doc = normalizer.ApplyEndpointConfidence(doc)

	log.Printf("wiki-llm generate api: fuentes cargadas y fusionadas correctamente")
	log.Printf("  sources: %s", strings.Join(sourcePaths, ", "))
	log.Printf("  source-type: %s", sourceType)
	if strings.TrimSpace(codePath) != "" {
		log.Printf("  code: %s", codePath)
	}
	log.Printf("  output: %s", output)
	log.Printf("  title: %s", doc.Title)
	log.Printf("  version: %s", doc.Version)
	log.Printf("  endpoints extraidos: %d", len(doc.Endpoints))

	log.Println("wiki-llm generate api: renderizando Markdown...")
	files := renderer.RenderAPI(doc)

	log.Println("wiki-llm generate api: escribiendo archivos...")
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
	log.Println("  wiki-llm generate api --source <archivo-fuente> [--source <archivo-fuente> ...] --output <directorio>")
	log.Println("")
	log.Println("Descripcion:")
	log.Println("  Genera documentacion de API desde una o multiples fuentes soportadas.")
	log.Println("")
	log.Println("Opciones:")
	log.Println("  --source        Ruta al archivo de fuente (opcional, repetible)")
	log.Println("  --source-type   auto|openapi|postman (default: auto)")
	log.Println("  --code          Ruta raiz de proyecto para buscar swagger.json recursivamente")
	log.Println("  --output        Directorio de salida (obligatorio)")
	log.Println("  -h, --help      Muestra esta ayuda")
}

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	if s == nil {
		return ""
	}
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return fmt.Errorf("valor vacio para --source")
	}
	*s = append(*s, v)
	return nil
}

func validateSourcePath(source string) error {
	if strings.TrimSpace(source) == "" {
		return fmt.Errorf("ruta source vacia")
	}
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("no se pudo leer source: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("source debe ser archivo, se recibio directorio: %s", source)
	}
	return nil
}

func resolveDocumentTitle(doc normalizer.APIDocument, codePath string) string {
	title := strings.TrimSpace(doc.Title)
	if strings.TrimSpace(codePath) == "" {
		return title
	}

	// Solo fallback si el titulo final no fue consolidado.
	if title == "" {
		fallback := strings.TrimSpace(filepath.Base(strings.TrimRight(codePath, "/")))
		if fallback != "" {
			return fallback
		}
	}
	return title
}
