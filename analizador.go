package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//comando = strings.ToLower(comando)

func analizarComandoPrincipal(entrada []string) {
	switch strings.ToLower(entrada[0]) {
	case "exec":
		analizarParametrosExec(entrada)
	case "pause":
		pausaDeLaEjecucion()
	case "mkdisk":
		analizarParametrosMkdisk(entrada)
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

func analizarParametrosMkgrp(entrada []string) {
}

func analizarParametrosLogin(entrada []string) {
	usr := "vacio"
	pwd := "vacio"
	id := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		switch aux[0] {
		case "-id":
			id = aux[1]
		case "-usr":
			usr = aux[1]
		case "-pwd":
			pwd = aux[1]
		}
	}
	if usr != "vacio" && pwd != "vacio" && id != "vacio" {
		//ir al metodo para el mkdisk
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

/*
func analizarParametrosMkfs(entrada []string) {
	id := "vacio"
	unit := "vacio"
	tipo := "vacio"
	add := 0

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		switch aux[0] {
		case "-id":
			id = aux[1]
		case "-tipo":
			if aux[1] == "fast" || aux[1] == "full" {
				tipo = aux[1]
			}
		case "-add":
			add, _ = strconv.Atoi(aux[1])
		case "-unit":
			if aux[1] == "b" || aux[1] == "k" || aux[1] == "m" {
				unit = aux[1]
			}
		}
	}
	if id != "vacio" {
		//ir al metodo para el mkdisk
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}
*/
func analizarParametrosExec(entrada []string) {
	path := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		switch aux[0] {
		case "-path":
			path = obtenerPath(entrada, i)
		}
	}

	if path != "vacio" {
		lecturaDeArchivo(path)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosUnmount(entrada []string) {
	listado := make([]string, 0)

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		match, _ := regexp.MatchString("-id([0-9]+)", strings.ToLower(aux[0]))
		if match {
			listado = append(listado, strings.ToLower(aux[1]))
		}
	}

	if len(listado) > 0 {
		//ir al metodo para leer el archivo
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosMount(entrada []string) {
	path := "vacio"
	name := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		switch strings.ToLower(aux[0]) {
		case "-path":
			path = obtenerPath(entrada, i)
		case "-name":
			name = aux[1]
		}
	}

	if path != "vacio" && name != "vacio" {
		//ir al metodo para el mkdisk
	} else {
		//mostrara en la consola las particiones montadas
	}
}

func analizarParametrosFdisk(entrada []string) {
	path := "vacio"
	size := -1
	name := "vacio"
	unit := "k"
	tipo := "p"
	fit := "wf"
	delete := "vacio"
	add := 0

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		switch strings.ToLower(aux[0]) {
		case "-size":
			size = obtenerTamanio(aux[1])
		case "-unit":
			if val := strings.ToLower(aux[1]); val == "b" || val == "k" || val == "m" {
				unit = val
			}
		case "-path":
			path = obtenerPath(entrada, i)
		case "-type":
			if val := strings.ToLower(aux[1]); val == "p" || val == "e" || val == "l" {
				tipo = val
			}
		case "-fit":
			if val := strings.ToLower(aux[1]); val == "bf" || val == "ff" || val == "wf" {
				fit = val
			}
		case "-delete":
			if val := strings.ToLower(aux[1]); val == "fast" || val == "full" {
				delete = val
			}
		case "-name":
			name = aux[1]
		case "-add":
			add, _ = strconv.Atoi(aux[1])
		}
	}
	if path != "vacio" && size > 0 && name != "vacio" {
		if delete != "vacio" && add == 0 {
			//eliminar particion
		} else if delete == "vacio" && add != 0 {
			//agregar espacio a la particion
		} else if delete == "vacio" && add == 0 {
			//crear particion
		}
		fmt.Print(unit, tipo, fit)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosRmdisk(entrada []string) {
	path := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		switch strings.ToLower(aux[0]) {
		case "-path":
			path = obtenerPath(entrada, i)
		}
	}

	if path != "vacio" {
		//ir al metodo para el rmdisk
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosMkdisk(entrada []string) {
	path := "vacio"
	size := -1
	name := "vacio"
	unit := "m"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		switch strings.ToLower(aux[0]) {
		case "-size":
			size = obtenerTamanio(aux[1])
		case "-path":
			path = obtenerPath(entrada, i)
		case "-name":
			if strings.HasSuffix(aux[1], ".dsk") {
				name = aux[1]
			}
		case "-unit":
			if val := strings.ToLower(aux[1]); val == "k" || val == "m" {
				unit = aux[1]
			}
		}
	}

	if path != "vacio" && size > 0 && name != "vacio" {
		fmt.Print(unit)
		//ir al metodo para el mkdisk
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func obtenerPath(entrada []string, posicion int) string {
	path1 := strings.Split(entrada[posicion], "->")
	if string(path1[1][0]) == "\"" {
		path := path1[1] + entrada[posicion+1]
		path = strings.ReplaceAll(path, "\"", "")
		return path
	}
	return path1[1]
}

func obtenerTamanio(entrada string) int {
	size := -1
	if val, _ := strconv.Atoi(entrada); val > 0 {
		size, _ = strconv.Atoi(entrada)
		return size
	}
	return size
}

func lecturaDeArchivo(path string) {

}

func pausaDeLaEjecucion() {
	fmt.Println("Presione la tecla enter para continuar con la ejecucion c:")
	fmt.Scanln()
}
