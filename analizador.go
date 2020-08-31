package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func analizarComandoPrincipal(entrada []string) {
	if strings.Contains(entrada[0], "#") == false {
		switch strings.ToLower(entrada[0]) {
		case "exec":
			analizarParametrosExec(entrada)
		case "pause":
			pausaDeLaEjecucion()
		case "mkdisk":
			analizarParametrosMkdisk(entrada)
		case "rmdisk":
			analizarParametrosRmdisk(entrada)
		case "fdisk":
			analizarParametrosFdisk(entrada)
		case "mount":
			analizarParametrosMount(entrada)
		case "unmount":
			analizarParametrosUnmount(entrada)
		case "mkfs":
			analizarParametrosMkfs(entrada)
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
	} else {
		fmt.Println("Has echo un comentario")
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

func analizarParametrosMkfs(entrada []string) {
	id := "vacio"
	unit := "k"
	tipo := "full"
	add := 0

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-unit":
				if val := strings.ToLower(aux[1]); val == "b" || val == "k" || val == "m" {
					unit = val
				}
			case "-type":
				if val := strings.ToLower(aux[1]); val == "fast" || val == "full" {
					tipo = val
				}
			case "-add":
				add, _ = strconv.Atoi(aux[1])
			case "-id":
				id = strings.ToLower(aux[1])
			}
		} else {
			break
		}
	}

	if id != "vacio" {
		if add != 0 {
			//modifcar el tamaño del sistema de archivos
			fmt.Println(unit)
		} else {
			//formatear la particion
			fmt.Println(tipo)
		}
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosExec(entrada []string) {
	path := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch aux[0] {
			case "-path":
				path = obtenerPath(entrada, i)
			}
		} else {
			break
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
		if strings.Contains(aux[0], "#") == false {
			match, _ := regexp.MatchString("-id([0-9]+)", strings.ToLower(aux[0]))
			if match {
				listado = append(listado, strings.ToLower(aux[1]))
			}
		} else {
			break
		}
	}

	if len(listado) > 0 {
		for i := 0; i < len(listado); i++ {
			desmontarParticion(listado[i])
		}
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosMount(entrada []string) {
	path := "vacio"
	name := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-path":
				path = obtenerPath(entrada, i)
			case "-name":
				name = aux[1]
			}
		} else {
			break
		}
	}

	if path != "vacio" && name != "vacio" {
		montarParticion(path, name)
	} else {
		mostrarParticionesMontadas()
	}
}

func analizarParametrosFdisk(entrada []string) {
	path := "vacio"
	size := int64(-1)
	name := "vacio"
	unit := "k"
	tipo := "p"
	fit := "wf"
	delete := "vacio"
	add := 0

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
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
		} else {
			break
		}
	}
	if path != "vacio" && name != "vacio" {
		if delete != "vacio" && add == 0 {
			eliminarParticion(path, name, delete)
		} else if delete == "vacio" && add != 0 {
			modificarTamanioParticion(size, unit, path, name)
		} else if delete == "vacio" && size > 0 {
			crearParticion(size, unit, path, tipo, fit, name)
		}

	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosRmdisk(entrada []string) {
	path := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-path":
				path = obtenerPath(entrada, i)
			}
		} else {
			break
		}
	}

	if path != "vacio" {
		eliminarDisco(path)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosMkdisk(entrada []string) {
	path := "vacio"
	size := int64(-1)
	name := "vacio"
	unit := "m"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-size":
				size = obtenerTamanio(aux[1])
			case "-path":
				path = obtenerPath(entrada, i)
			case "-name":
				if val := strings.ToLower(aux[1]); strings.HasSuffix(val, ".dsk") {
					name = aux[1]
				}
			case "-unit":
				if val := strings.ToLower(aux[1]); val == "k" || val == "m" {
					unit = aux[1]
				}
			}
		} else {
			break
		}
	}

	if path != "vacio" && size > 0 && name != "vacio" {
		crearDisco(size, path, name, unit)
	} else {
		if size == -1 {
			fmt.Println("El tamaño del disco ingresado es invalido")
		} else {
			fmt.Println("Revise los parametros del comando")
		}
	}
}

func obtenerPath(entrada []string, posicion int) string {
	path1 := strings.Split(entrada[posicion], "->")
	if string(path1[1][0]) == "\"" {
		path := path1[1] + entrada[posicion+1]
		path = strings.ReplaceAll(path, "\"", "")
		path = strings.ReplaceAll(path, "\n", "")
		return path
	}
	path1[1] = strings.ReplaceAll(path1[1], "\n", "")
	return path1[1]
}

func obtenerTamanio(entrada string) int64 {
	size := -1
	if val, _ := strconv.Atoi(entrada); val > 0 {
		size, _ = strconv.Atoi(entrada)
		size1 := int64(size)
		return size1
	}
	return int64(-1)
}

func lecturaDeArchivo(path string) {
	bandera := false
	contenido := ""
	linea := ""
	file, err := os.Open(path)

	if err != nil {
		fmt.Println("Error en la apertura del archivo")
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "\\*") {
			contenido += scanner.Text()
			bandera = true
		} else {
			linea = scanner.Text()
			if bandera {
				bandera = false
				linea = contenido + linea
				contenido = ""
				linea = strings.ReplaceAll(linea, "\\*", "")
			}
			linea = strings.ReplaceAll(linea, "\n", "")
			fmt.Println(linea)
			analizarComandoPrincipal(strings.Split(linea, " "))
		}
	}
}

func pausaDeLaEjecucion() {
	fmt.Print("Presione la tecla enter para continuar con la ejecucion c:")
	val := ""
	fmt.Scanln(&val)
}
