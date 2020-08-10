package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Ingrese un comando: ")
		comando, _ := reader.ReadString('\n')
		if validacion := strings.Split(comando, " "); strings.Contains(validacion[len(validacion)-1], "\\*") {
			fmt.Print("Ingrese la continuacion del comando: ")
			continuacion, _ := reader.ReadString('\n')
			comando = comando + continuacion
			comando = strings.ReplaceAll(comando, "\\*", "")
		}
		comando = strings.ReplaceAll(comando, "\n", "")
		analizarComandoPrincipal(strings.Split(comando, " "))
	}
}
