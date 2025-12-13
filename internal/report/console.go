package report

import (
	"fmt"
	"go-scanner/internal/scanner"
	"os"
	"sort"
	"text/tabwriter" //permite imprimir tablas alienadas :p
)

// lee el canal de resultados y los muestra formateados en consola
func PrintResults(results <-chan scanner.ScanResult) { //(<-chan): el canal es solo lectura
	var openPorts []scanner.ScanResult

	//stream processing
	for res := range results {
		if res.IsOpen {
			openPorts = append(openPorts, res)
		}
	}

	//ordenamiento de resultados (ports)
	sort.Slice(openPorts, func(i, j int) bool {
		return openPorts[i].Port < openPorts[j].Port
	})

	//renderizacion de tabla
	fmt.Println("\n--- scanning results ---")
	if len(openPorts) == 0 {
		fmt.Println("no open ports found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PORT\tSTATE\tPROTOCOL")
	for _, res := range openPorts {
		fmt.Fprintf(w, "%d\t%s\ttcp\n", res.Port, "OPEN")
	}
	w.Flush()
	fmt.Println("------------------------------")
}
