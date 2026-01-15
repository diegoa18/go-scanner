package report

import (
	"fmt"
	"go-scanner/internal/scanner"
	"os"
	"sort"
	"text/tabwriter" //permite imprimir tablas alienadas :p
)

// lee el slice de resultados y los muestra formateados en consola
func PrintResults(results []scanner.ScanResult, showAll bool) {
	// Agrupar por host
	resultsByHost := make(map[string][]scanner.ScanResult)
	var hosts []string

	//stream processing y filtrado
	for _, res := range results {
		// Logica de visualizacion:
		// - Siempre mostrar OPEN
		// - Siempre mostrar FILTERED
		// - Mostrar CLOSED solo si showAll es true

		shouldShow := false
		if res.State == scanner.PortStateOpen || res.State == scanner.PortStateFiltered {
			shouldShow = true
		} else if showAll {
			shouldShow = true
		}

		if shouldShow {
			if _, exists := resultsByHost[res.Host]; !exists {
				hosts = append(hosts, res.Host)
			}
			resultsByHost[res.Host] = append(resultsByHost[res.Host], res)
		}
	}

	// Ordenar hosts para output consistente
	sort.Strings(hosts)

	fmt.Println("\n--- scanning results ---")
	if showAll {
		fmt.Println("[Full report enabled: showing all states]") //para --all
	}

	if len(hosts) == 0 {
		fmt.Println("no ports found matching criteria.")
		return
	}

	for _, host := range hosts {
		hostResults := resultsByHost[host]

		fmt.Printf("\nTarget: %s\n", host)

		//ordenamiento de resultados (ports)
		sort.Slice(hostResults, func(i, j int) bool {
			return hostResults[i].Port < hostResults[j].Port
		})

		//verificar si se capturo algun banner para ajustar columnas
		showBanner := false
		for _, res := range hostResults {
			if res.Banner != "" {
				showBanner = true
				break
			}
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// Encabezados con STATE
		if showBanner {
			fmt.Fprintln(w, "PORT\tSTATE\tSERVICE\tBANNER")
			for _, res := range hostResults {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", res.Port, res.State, res.Service, res.Banner)
			}
		} else {
			fmt.Fprintln(w, "PORT\tSTATE\tSERVICE")
			for _, res := range hostResults {
				fmt.Fprintf(w, "%d\t%s\t%s\n", res.Port, res.State, res.Service)
			}
		}
		w.Flush()
	}
	fmt.Println("------------------------------")
}
