package report

import (
	"fmt"
	"go-scanner/internal/scanner"
	"os"
	"sort"
	"text/tabwriter" //permite imprimir tablas alienadas :p
)

// lee el slice de resultados y los muestra formateados en consola
func PrintResults(results []scanner.ScanResult) {
	// Agrupar por host
	resultsByHost := make(map[string][]scanner.ScanResult)
	var hosts []string

	//stream processing
	for _, res := range results {
		if res.IsOpen {
			if _, exists := resultsByHost[res.Host]; !exists {
				hosts = append(hosts, res.Host)
			}
			resultsByHost[res.Host] = append(resultsByHost[res.Host], res)
		}
	}

	// Ordenar hosts para output consistente
	sort.Strings(hosts)

	fmt.Println("\n--- scanning results ---")
	if len(hosts) == 0 {
		fmt.Println("no open ports found.")
		return
	}

	for _, host := range hosts {
		openPorts := resultsByHost[host]

		fmt.Printf("\nTarget: %s\n", host)

		//ordenamiento de resultados (ports)
		sort.Slice(openPorts, func(i, j int) bool {
			return openPorts[i].Port < openPorts[j].Port
		})

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
			fmt.Fprintln(w, "PORT\tSERVICE\tBANNER")
			for _, res := range openPorts {
				fmt.Fprintf(w, "%d\t%s\t%s\n", res.Port, res.Service, res.Banner)
			}

		} else { //si no -> simplemente no se muestra el apartado BANNER
			fmt.Fprintln(w, "PORT\tSERVICE")
			for _, res := range openPorts {
				fmt.Fprintf(w, "%d\t%s\n", res.Port, res.Service)
			}
		}
		w.Flush()
	}
	fmt.Println("------------------------------")
}
