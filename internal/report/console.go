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

	//verificar si se capturo algun banner
	showBanner := false
	for _, res := range openPorts {
		if res.Banner != "" {
			showBanner = true
			break
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if showBanner { //si se capturo algun banner -> agregar apartado BANNER en la tabla
		fmt.Fprintln(w, "PORT\tSTATE\tPROTOCOL\tBANNER")
		for _, res := range openPorts {
			fmt.Fprintf(w, "%d\t%s\ttcp\t%s\n", res.Port, "OPEN", res.Banner)
		}

	} else { //si no -> simplemente no se muestra el apartado BANNER
		fmt.Fprintln(w, "PORT\tSTATE\tPROTOCOL")
		for _, res := range openPorts {
			fmt.Fprintf(w, "%d\t%s\ttcp\n", res.Port, "OPEN")
		}
	}
	w.Flush()
	fmt.Println("------------------------------")
}
