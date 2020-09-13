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
			analizarParametrosLogin(entrada)
		case "logout":
			cerrarSesion()
		case "mkgrp":
			analizarParametrosMkgrp(entrada)
		case "rmgrp":
			analizarParametrosRmgrp(entrada)
		case "mkusr":
			analizarParametrosMkusr(entrada)
		case "rmusr":
			analizarParametrosRmusr(entrada)
		case "chmod":
		case "mkfile":
			analizarParametrosMkfile(entrada)
		case "cat":
			analizarParametrosCat(entrada)
		case "rm":
		case "edit":
			analizarParametrosEdit(entrada)
		case "ren":
			analizarParametrosRen(entrada)
		case "mkdir":
			analizarParametrosMkdir(entrada)
		case "cp":
			//analizarParametrosCp(entrada)
		case "mv":
			//analizarParametrosMv(entrada)
		case "find":
		case "chown":
		case "chgrp":
		case "loss":
			analizarParametrosLoss(entrada)
		case "recovery":
			analizarParametrosRecovery(entrada)
		case "rep":
			analizarParametrosRep(entrada)
		default:
			fmt.Println("El comando ingresado no es valido")
		}
	} else {
		fmt.Println("Has echo un comentario")
	}
}

func analizarParametrosCp(entrada []string) {
	id := "vacio"
	path := "vacio"
	destino := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-dest":
				destino = obtenerPath(entrada, i)
			case "-path":
				path = obtenerPath(entrada, i)
			case "-id":
				id = strings.ToLower(aux[1])
			}
		} else {
			break
		}
	}

	if id != "vacio" && path != "vacio" && destino != "vacio" {
		copiarArchivosOCarpetas(id, path, destino)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosMv(entrada []string) {

}

func analizarParametrosRen(entrada []string) {
	id := "vacio"
	path := "vacio"
	name := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-name":
				name = obtenerPath(entrada, i)
			case "-path":
				path = obtenerPath(entrada, i)
			case "-id":
				id = strings.ToLower(aux[1])
			}
		} else {
			break
		}
	}

	if id != "vacio" && path != "vacio" && name != "vacio" {
		modificarNombre(id, path, name)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosRm(entrada []string) {
	id := "vacio"
	path := "vacio"
	especial := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-rf":
				especial = "rf"
			case "-path":
				path = obtenerPath(entrada, i)
			case "-id":
				id = strings.ToLower(aux[1])
			}
		} else {
			break
		}
	}

	if id != "vacio" && path != "vacio" {
		//eliminar
		fmt.Println(especial)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosEdit(entrada []string) {
	id := "vacio"
	path := "vacio"
	size := -1
	cont := ""

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-path":
				path = obtenerPath(entrada, i)
			case "-id":
				id = strings.ToLower(aux[1])
			case "-size":
				if val, _ := strconv.Atoi(aux[1]); val >= 0 {
					size, _ = strconv.Atoi(aux[1])
				} else {
					fmt.Println("El parametro size debe de ser igual o mayor a cero")
					return
				}
			case "-cont":
				cont = obtenerPath(entrada, i)
			}
		} else {
			break
		}
	}

	if id != "vacio" && path != "vacio" {
		modificarContenidoArchivo(id, path, cont, int64(size))
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosCat(entrada []string) {
	listado := make([]string, 0)
	id := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			match, _ := regexp.MatchString("-file([0-9]+)", strings.ToLower(aux[0]))
			if match {
				listado = append(listado, obtenerPath(entrada, i))
			}
			if aux[0] == "-id" {
				id = strings.ToLower(aux[1])
			}
		} else {
			break
		}
	}

	if len(listado) > 0 && id != "vacio" {
		for i := 0; i < len(listado); i++ {
			mostrarContenidoArchivo(id, listado[i])
		}
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosRmusr(entrada []string) {
	id := "vacio"
	usr := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-id":
				id = strings.ToLower(aux[1])
			case "-usr":
				usr = obtenerPath(entrada, i)
			}
		} else {
			break
		}
	}

	if id != "vacio" && usr != "vacio" {
		eliminarUsuario(id, usr)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosMkusr(entrada []string) {
	id := "vacio"
	usr := "vacio"
	pwd := "vacio"
	grp := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-id":
				id = strings.ToLower(aux[1])
			case "-usr":
				usr = obtenerPath(entrada, i)
			case "-pwd":
				pwd = obtenerPath(entrada, i)
			case "-grp":
				grp = obtenerPath(entrada, i)
			}
		} else {
			break
		}
	}

	if id != "vacio" && usr != "vacio" && pwd != "vacio" && grp != "vacio" {
		crearUsuario(id, usr, pwd, grp)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosLoss(entrada []string) {
	id := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-id":
				id = strings.ToLower(aux[1])
			}
		} else {
			break
		}
	}

	if id != "vacio" {
		simulacionPerdida(id)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosRecovery(entrada []string) {
	id := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-id":
				id = strings.ToLower(aux[1])
			}
		} else {
			break
		}
	}

	if id != "vacio" {
		recuperarSistema(id)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosMkfile(entrada []string) {
	id := "vacio"
	path := "vacio"
	especial := "vacio"
	size := -1
	cont := ""

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-p":
				especial = "p"
			case "-path":
				path = obtenerPath(entrada, i)
			case "-id":
				id = strings.ToLower(aux[1])
			case "-size":
				if val, _ := strconv.Atoi(aux[1]); val >= 0 {
					size, _ = strconv.Atoi(aux[1])
				} else {
					fmt.Println("El parametro size debe de ser igual o mayor a cero")
					return
				}
			case "-cont":
				cont = obtenerPath(entrada, i)
			}
		} else {
			break
		}
	}

	if id != "vacio" && path != "vacio" {
		crearArchivo(id, especial, path, cont, int64(size), 1)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosRep(entrada []string) {
	nombre := "vacio"
	id := "vacio"
	path := "vacio"
	ruta := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-ruta":
				ruta = obtenerPath(entrada, i)
			case "-path":
				path = obtenerPath(entrada, i)
			case "-id":
				id = strings.ToLower(aux[1])
			case "-nombre":
				if val := strings.ToLower(aux[1]); val == "mbr" || val == "disk" || val == "sb" || val == "bm_arbdir" || val == "bm_detdir" || val == "bm_inode" || val == "bm_block" || val == "bitacora" || val == "directorio" || val == "tree_file" || val == "tree_directorio" || val == "tree_complete" || val == "ls" {
					nombre = val
				}
			case "-name":
				if val := strings.ToLower(aux[1]); val == "mbr" || val == "disk" || val == "sb" || val == "bm_arbdir" || val == "bm_detdir" || val == "bm_inode" || val == "bm_block" || val == "bitacora" || val == "directorio" || val == "tree_file" || val == "tree_directorio" || val == "tree_complete" || val == "ls" {
					nombre = val
				}
			}
		} else {
			break
		}
	}

	if id != "vacio" && path != "vacio" && nombre != "vacio" {
		desicionReporte(id, path, ruta, nombre)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosMkdir(entrada []string) {
	id := "vacio"
	path := "vacio"
	especial := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-p":
				especial = "p"
			case "-path":
				path = obtenerPath(entrada, i)
			case "-id":
				id = strings.ToLower(aux[1])
			}
		} else {
			break
		}
	}

	if id != "vacio" && path != "vacio" {
		crearAVD(id, especial, path, 1)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosMkgrp(entrada []string) {
	id := "vacio"
	name := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-id":
				id = strings.ToLower(aux[1])
			case "-name":
				name = obtenerPath(entrada, i)
			}
		} else {
			break
		}
	}

	if id != "vacio" && name != "vacio" {
		crearGrupo(id, name)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosRmgrp(entrada []string) {
	id := "vacio"
	name := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-id":
				id = strings.ToLower(aux[1])
			case "-name":
				name = obtenerPath(entrada, i)
			}
		} else {
			break
		}
	}

	if id != "vacio" && name != "vacio" {
		eliminarGrupo(id, name)
	} else {
		fmt.Println("El comando ingresado no es valido")
	}
}

func analizarParametrosLogin(entrada []string) {
	id := "vacio"
	usr := "vacio"
	pwd := "vacio"

	for i := 1; i < len(entrada); i++ {
		aux := strings.Split(entrada[i], "->")
		if strings.Contains(aux[0], "#") == false {
			switch strings.ToLower(aux[0]) {
			case "-id":
				id = strings.ToLower(aux[1])
			case "-usr":
				usr = obtenerPath(entrada, i)
			case "-pwd":
				pwd = obtenerPath(entrada, i)
			}
		} else {
			break
		}
	}

	if id != "vacio" && pwd != "vacio" && usr != "vacio" {
		iniciarSesion(id, usr, pwd)
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
			case "-tipo":
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
			formateoSistema(id, tipo)
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
				name = obtenerPath(entrada, i)
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
				name = obtenerPath(entrada, i)
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
	if val := path1[1]; string(val[0]) == "\"" && string(val[len(val)-1]) == "\"" {
		val = strings.ReplaceAll(val, "\"", "")
		val = strings.ReplaceAll(val, "\n", "")
		return val
	} else if val := path1[1]; string(val[0]) == "\"" {
		for i := posicion + 1; i < len(entrada); i++ {
			val += " " + entrada[i]
			if strings.Contains(entrada[i], "\"") {
				break
			}
		}
		val = strings.ReplaceAll(val, "\"", "")
		val = strings.ReplaceAll(val, "\n", "")
		return val
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
	fmt.Print("\nPresione la tecla enter para continuar con la ejecucion c:\n")
	val := ""
	fmt.Scanln(&val)
}
