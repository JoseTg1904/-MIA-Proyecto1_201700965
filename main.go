package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Ingrese un comando: ")
	comando, _ := reader.ReadString('\n')
	comando = strings.ToLower(comando)
	analizarComandoPrincipal(strings.Split(comando, " "))
}
