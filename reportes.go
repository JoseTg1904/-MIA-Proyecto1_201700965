package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func desicionReporte(id, path, ruta, nombre string) {
	switch nombre {
	case "mbr":
		reporteMBR(id, path)
	case "disk":
		reporteDisk(id, path)
	case "sb":
		reporteSb(id, path)
	case "bm_arbdir":
		reporteBMAVD(id, path)
	case "bm_detdir":
		reporteBMDD(id, path)
	case "bm_inode":
		reporteBMINodos(id, path)
	case "bm_block":
		reporteBMBloques(id, path)
	case "bitacora":
	case "directorio":
	case "tree_file":
	case "tree_directorio":
	case "tree_complete":
	case "ls":
	}
}

func reporteMBR(id, path string) {
	vacia := particion{}

	disco, _, _ := obtenerDiscoMontado(id)
	mbr := obtenerMBR(disco)

	dot := "digraph MBR{\n"
	dot += "\"mbr\" [shape=\"plaintext\" label = < <table> <tr> <td>Clave</td> <td>Valor</td> </tr>\n"
	dot += "<tr><td> Tamanio</td> <td>" + strconv.Itoa(int(mbr.Tamanio)) + " </td></tr>\n"
	dot += "<tr><td> Fecha Creacion </td><td> " + retornarStringLimpio(mbr.Creacion[:]) + "</td></tr>\n"
	dot += "<tr><td> Random </td><td> " + strconv.Itoa(int(mbr.Random)) + "</td></tr>\n"

	for i := 0; i < 4; i++ {
		if mbr.Particiones[i] != vacia {
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Estado </td><td> " + string(mbr.Particiones[i].Estado) + " </td></tr>\n"
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Tipo </td><td> " + string(mbr.Particiones[i].Tipo) + " </td></tr>\n"
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Ajuste </td><td> " + retornarStringLimpio(mbr.Particiones[i].Ajuste[:]) + " </td></tr>\n"
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Inicio </td><td> " + strconv.Itoa(int(mbr.Particiones[i].Inicio)) + " </td></tr>\n"
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Nombre </td><td> " + retornarStringLimpio(mbr.Particiones[i].Nombre[:]) + " </td></tr>\n"
		} else {
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Estado </td><td> -1 </td></tr> \n"
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Tipo </td><td> -1 </td></tr> \n"
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Ajuste </td><td> -1 </td></tr> \n"
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Inicio </td><td> -1 </td></tr> \n"
			dot += "<tr><td> Particion" + strconv.Itoa(i+1) + "_Nombre </td><td> -1 </td></tr> \n "
		}
	}

	dot += "</table>> ]\n }"
	pathSinComillas := strings.ReplaceAll(path, "\"", "")
	aux := strings.Split(pathSinComillas, ".")
	pathDot := aux[0] + ".dot"
	pathImagen := aux[0] + ".png"
	archivoSalida, _ := os.Create(pathDot)
	archivoSalida.WriteString(dot)
	archivoSalida.Close()

	exec.Command("dot", pathDot, "-Tpng", "-o", pathImagen).Output()
}

func reporteDisk(id, path string) {
	vacia := particion{}

	disco, _, _ := obtenerDiscoMontado(id)
	mbr := obtenerMBR(disco)

	tamañoDisponible := float64(mbr.Tamanio)

	nombre := strings.Split(disco.Name(), "/")
	nombre = strings.Split(nombre[len(nombre)-1], ".")

	dot := "digraph Disk{\n"
	dot += "\"nombre\" [shape=\"plaintext\" label = \"" + nombre[0] + "\"]\n"
	dot += "\"disco\" [shape=\"plaintext\" label = <<table><tr><td> MBR </td>\n"
	for i := 0; i < 4; i++ {
		if mbr.Particiones[i] != vacia {
			tamañoDisponible -= float64(mbr.Particiones[i].Tamanio)
			porcentaje := float64((float64(mbr.Particiones[i].Tamanio) * 100) / float64(mbr.Tamanio))
			porcentajeConvertido := fmt.Sprintf("%.3f", porcentaje)
			dot += "<td> " + retornarStringLimpio(mbr.Particiones[i].Nombre[:]) + " <br /> " + string(mbr.Particiones[i].Tipo) + "<br />" + porcentajeConvertido + "% </td>\n"
		}
	}

	if tamañoDisponible > 0 {
		porcentaje := float64((float64(tamañoDisponible) * 100) / float64(mbr.Tamanio))
		porcentajeConvertido := fmt.Sprintf("%.3f", porcentaje)
		dot += "<td> Libre <br />" + porcentajeConvertido + "% </td>\n"
	}

	dot += "</tr>\n</table>>]\n}"

	pathSinComillas := strings.ReplaceAll(path, "\"", "")
	aux := strings.Split(pathSinComillas, ".")
	pathDot := aux[0] + ".dot"
	pathImagen := aux[0] + ".png"
	archivoSalida, _ := os.Create(pathDot)
	archivoSalida.WriteString(dot)
	archivoSalida.Close()

	exec.Command("dot", pathDot, "-Tpng", "-o", pathImagen).Output()

}

func reporteSb(id, path string) {
	disco, _, inicio := obtenerDiscoMontado(id)
	sb := obtenerSuperBoot(disco, int64(inicio))

	dot := "digraph SB{\n"
	dot += "\"sb\" [shape=\"plaintext\" label = < <table> <tr> <td>Clave</td> <td>Valor</td> </tr>\n"
	dot += "<tr><td> Nombre disco </td> <td>" + retornarStringLimpio(sb.Nombre[:]) + " </td></tr>\n"
	dot += "<tr><td> No. AVD </td><td> " + strconv.Itoa(int(sb.NoAVD)) + "</td></tr>\n"
	dot += "<tr><td> No. DD </td><td> " + strconv.Itoa(int(sb.NoDD)) + "</td></tr>\n"
	dot += "<tr><td> No. I-nodos </td><td> " + strconv.Itoa(int(sb.NoINodos)) + "</td></tr>\n"
	dot += "<tr><td> No. Bloques </td><td> " + strconv.Itoa(int(sb.NoBloques)) + "</td></tr>\n"
	dot += "<tr><td> No. AVD libres </td><td> " + strconv.Itoa(int(sb.NoAVDLibres)) + "</td></tr>\n"
	dot += "<tr><td> No. DD libres </td><td> " + strconv.Itoa(int(sb.NoDDLibres)) + "</td></tr>\n"
	dot += "<tr><td> No. I-nodos libres </td><td> " + strconv.Itoa(int(sb.NoINodosLibres)) + "</td></tr>\n"
	dot += "<tr><td> No. Bloques libres </td><td> " + strconv.Itoa(int(sb.NoBloquesLibres)) + "</td></tr>\n"
	dot += "<tr><td> Fecha de Creacion </td><td> " + retornarStringLimpio(sb.Creacion[:]) + "</td></tr>\n"
	dot += "<tr><td> Ultimo montaje </td><td> " + retornarStringLimpio(sb.UltimoMontaje[:]) + "</td></tr>\n"
	dot += "<tr><td> No. Montajes </td><td> " + strconv.Itoa(int(sb.ContadorMontajes)) + "</td></tr>\n"
	dot += "<tr><td> Inicio bit map AVD </td><td> " + strconv.Itoa(int(sb.InicioBitMapAVD)) + "</td></tr>\n"
	dot += "<tr><td> Inicio AVD </td><td> " + strconv.Itoa(int(sb.InicioAVD)) + "</td></tr>\n"
	dot += "<tr><td> Inicio bit map DD </td><td> " + strconv.Itoa(int(sb.InicioBitMapDD)) + "</td></tr>\n"
	dot += "<tr><td> Inicio DD </td><td> " + strconv.Itoa(int(sb.InicioDD)) + "</td></tr>\n"
	dot += "<tr><td> Inicio bit map I-nodos </td><td> " + strconv.Itoa(int(sb.InicioBitMapINodo)) + "</td></tr>\n"
	dot += "<tr><td> Inicio I-nodos </td><td> " + strconv.Itoa(int(sb.InicioINodo)) + "</td></tr>\n"
	dot += "<tr><td> Inicio bit map bloques </td><td> " + strconv.Itoa(int(sb.InicioBitMapBloques)) + "</td></tr>\n"
	dot += "<tr><td> Inicio bloques </td><td> " + strconv.Itoa(int(sb.InicioBLoques)) + "</td></tr>\n"
	dot += "<tr><td> Inicio bitacora </td><td> " + strconv.Itoa(int(sb.InicioLog)) + "</td></tr>\n"
	dot += "<tr><td> Inicio bit map AVD </td><td> " + strconv.Itoa(int(sb.InicioBitMapAVD)) + "</td></tr>\n"
	dot += "<tr><td> Tamaño AVD </td><td> " + strconv.Itoa(int(sb.TamanioAVD)) + "</td></tr>\n"
	dot += "<tr><td> Tamaño DD </td><td> " + strconv.Itoa(int(sb.TamanioDD)) + "</td></tr>\n"
	dot += "<tr><td> Tamaño I-nodo </td><td> " + strconv.Itoa(int(sb.TamanioINodo)) + "</td></tr>\n"
	dot += "<tr><td> Tamaño bloque </td><td> " + strconv.Itoa(int(sb.TamanioBloque)) + "</td></tr>\n"
	dot += "<tr><td> Primer bit libre AVD </td><td> " + strconv.Itoa(int(sb.BitLibreAVD)) + "</td></tr>\n"
	dot += "<tr><td> Primer bit libre DD </td><td> " + strconv.Itoa(int(sb.BitLibreDD)) + "</td></tr>\n"
	dot += "<tr><td> Primer bit libre I-nodo </td><td> " + strconv.Itoa(int(sb.BitLibreINodo)) + "</td></tr>\n"
	dot += "<tr><td> Primer bit libre bloque </td><td> " + strconv.Itoa(int(sb.BitLibreBloque)) + "</td></tr>\n"
	dot += "<tr><td> No. Carnet </td><td> " + strconv.Itoa(int(sb.Carnet)) + "</td></tr>\n"
	dot += "</table>> ]\n }"

	pathSinComillas := strings.ReplaceAll(path, "\"", "")
	aux := strings.Split(pathSinComillas, ".")
	pathDot := aux[0] + ".dot"
	pathImagen := aux[0] + ".png"
	archivoSalida, _ := os.Create(pathDot)
	archivoSalida.WriteString(dot)
	archivoSalida.Close()

	exec.Command("dot", pathDot, "-Tpng", "-o", pathImagen).Output()
}

func reporteBMAVD(id, path string) {
	disco, _, inicio := obtenerDiscoMontado(id)
	sb := obtenerSuperBoot(disco, int64(inicio))

	bitMap := make([]byte, sb.NoAVD)
	contenido := make([]byte, sb.NoAVD)

	disco.Seek(int64(sb.InicioBitMapAVD), 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &bitMap)
	if a != nil {
	}

	contador := 1
	salida := "| "

	for i := 0; i < len(bitMap); i++ {
		if bitMap[i] == 0 {
			salida += "0 | "
		} else {
			salida += "1 | "
		}
		if contador == 20 {
			salida += "\n| "
			contador = 0
		}
		contador++
	}

	pathSinComillas := strings.ReplaceAll(path, "\"", "")
	archivoSalida, _ := os.Create(pathSinComillas)
	archivoSalida.WriteString(salida)
	archivoSalida.Close()
}

func reporteBMDD(id, path string) {
	disco, _, inicio := obtenerDiscoMontado(id)
	sb := obtenerSuperBoot(disco, int64(inicio))

	bitMap := make([]byte, sb.NoDD)
	contenido := make([]byte, sb.NoDD)

	disco.Seek(int64(sb.InicioBitMapDD), 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &bitMap)
	if a != nil {
	}

	contador := 1
	salida := "| "

	for i := 0; i < len(bitMap); i++ {
		if bitMap[i] == 0 {
			salida += "0 | "
		} else {
			salida += "1 | "
		}
		if contador == 20 {
			salida += "\n| "
			contador = 0
		}
		contador++
	}

	pathSinComillas := strings.ReplaceAll(path, "\"", "")
	archivoSalida, _ := os.Create(pathSinComillas)
	archivoSalida.WriteString(salida)
	archivoSalida.Close()
}

func reporteBMINodos(id, path string) {
	disco, _, inicio := obtenerDiscoMontado(id)
	sb := obtenerSuperBoot(disco, int64(inicio))

	bitMap := make([]byte, sb.NoINodos)
	contenido := make([]byte, sb.NoINodos)

	disco.Seek(int64(sb.InicioBitMapINodo), 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &bitMap)
	if a != nil {
	}

	contador := 1
	salida := "| "

	for i := 0; i < len(bitMap); i++ {
		if bitMap[i] == 0 {
			salida += "0 | "
		} else {
			salida += "1 | "
		}
		if contador == 20 {
			salida += "\n| "
			contador = 0
		}
		contador++
	}

	pathSinComillas := strings.ReplaceAll(path, "\"", "")
	archivoSalida, _ := os.Create(pathSinComillas)
	archivoSalida.WriteString(salida)
	archivoSalida.Close()
}

func reporteBMBloques(id, path string) {
	disco, _, inicio := obtenerDiscoMontado(id)
	sb := obtenerSuperBoot(disco, int64(inicio))

	bitMap := make([]byte, sb.NoBloques)
	contenido := make([]byte, sb.NoBloques)

	disco.Seek(int64(sb.InicioBitMapBloques), 0)
	_, err := disco.Read(contenido)
	if err != nil {
		fmt.Println("Error en la lectura del disco")
	}
	buffer := bytes.NewBuffer(contenido)
	a := binary.Read(buffer, binary.BigEndian, &bitMap)
	if a != nil {
	}

	contador := 1
	salida := "| "

	for i := 0; i < len(bitMap); i++ {
		if bitMap[i] == 0 {
			salida += "0 | "
		} else {
			salida += "1 | "
		}
		if contador == 20 {
			salida += "\n| "
			contador = 0
		}
		contador++
	}

	pathSinComillas := strings.ReplaceAll(path, "\"", "")
	archivoSalida, _ := os.Create(pathSinComillas)
	archivoSalida.WriteString(salida)
	archivoSalida.Close()
}

func retornarStringLimpio(entrada []byte) string {
	salida := ""
	for i := 0; i < len(entrada); i++ {
		if entrada[i] == 0 {
			salida += ""
		} else {
			salida += string(entrada[i])
		}
	}

	return salida
}
