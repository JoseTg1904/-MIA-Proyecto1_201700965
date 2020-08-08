package main

import (
	"fmt"
	"strings"
)

func analizarComandoPrincipal(entrada []string) {
	switch entrada[0] {
	case "exec":
		analizarParametrosExec(entrada)
	case "pause":
	case "mkdisk":
	case "rmdisk":
	case "fdisk":
	case "mount":
	case "unmount":
	case "mkfs":
	case "login":
	case "logout":
	case "mkgrp":
	case "rmgrp":
	case "mkusr":
	case "rmusr":
	case "chmod":
	case "mkfile":
	case "cat":
	case "rm":
	case "edit":
	case "ren":
	case "mkdir":
	case "cp":
	case "mv":
	case "find":
	case "chown":
	case "chgrp":
	case "loss":
	case "recovery":
	case "rep":
	default:
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosExec(entrada []string) {
	path := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		switch aux[0] {
		case "-path":
			path = aux[1]
		}
	}

	if path != "vacio" {
		//ir al metodo para leer el archivo
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}
