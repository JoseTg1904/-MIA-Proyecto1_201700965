package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese un comando: ")
	comando, _ := reader.ReadString('\n')
	if validacion := strings.Split(comando, " "); strings.Contains(validacion[len(validacion)-1], "\\*") {
		fmt.Print("Ingrese la continuacion del comando: ")
		continuacion, _ := reader.ReadString('\n')
		comando = comando + continuacion
		comando = strings.ReplaceAll(comando, "\n", "")
		comando = strings.ReplaceAll(comando, "\\*", "")
	}
	fmt.Println(comando)
	analizarComandoPrincipal(strings.Split(comando, " "))
}
