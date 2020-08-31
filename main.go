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
		if strings.Contains(comando, "\\*") {
			for {
				fmt.Print("Ingrese la continuacion del comando: ")
				continuacion, _ := reader.ReadString('\n')
				comando = comando + continuacion
				comando = strings.ReplaceAll(comando, "\\*", "")
				if strings.Contains(continuacion, "\\*") == false {
					break
				}
			}
		}
		comando = strings.ReplaceAll(comando, "\n", "")
		analizarComandoPrincipal(strings.Split(comando, " "))
		//formateoSistema(1073741824)
	}
}
