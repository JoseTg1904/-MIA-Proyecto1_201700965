package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var numRand int16 = 0

type discoMBR struct {
	Tamanio     int64
	Creacion    [16]byte
	Random      int16
	Particiones [4]particion
}

//mkdisk
func crearDisco(size int64, path, name, unit string) {
	//conversion a tamaño en kilobytes o megabytes
	if unit == "k" {
		size *= 1024
	} else {
		size *= 1048576
	}

	//creacion de carpetas necesarias para el almacenamiento del archivo
	exec.Command("mkdir", "-p", path).Output()

	//validacion de la existencia del archivo
	if _, err := os.Stat(path + name); err == nil {
		fmt.Println("\033[1;31mEl archivo del disco ya existe\033[0m")
		return
	}

	//creacion del archivo de simulacion del disco
	archivo, _ := os.Create(path + name)
	defer archivo.Close()

	//instancia de una particion vacia para el llenado del mbr
	particionLimpia := particion{Inicio: -1}

	//instancia del mbr y llenado de datos del mismo
	mbr := discoMBR{Tamanio: size,
		Random: numRand,
		Particiones: [4]particion{particionLimpia,
			particionLimpia,
			particionLimpia,
			particionLimpia}}
	copy(mbr.Creacion[:], obtenerFecha())

	//aumento del numero random que identifica al disco
	numRand++

	//llenado del archivo con valores de cero para obtener el tamaño especificado
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))
	archivo.Write(buffer.Bytes())

	//corrimiento del puntero del archivo para alcanzar el tamaño especificado
	archivo.Seek(size-int64(1), 0)
	archivo.Write(buffer.Bytes())

	//posicionando el puntero en el inicio del archivo y limpiando el buffer de escritura
	archivo.Seek(0, 0)
	buffer.Reset()

	//escritura del struct que representa el mbr en el archivo
	binary.Write(buffer, binary.BigEndian, &mbr)
	archivo.Write(buffer.Bytes())

	fmt.Println("\033[1;32mSe a creado con exito el disco\033[0m")
}

//rmdisk
func eliminarDisco(path string) {
	fmt.Print("\033[1;33mSeguro que desea eliminar el disco ? [S/N]: \033[0m")
	val := ""
	fmt.Scanln(&val)
	if strings.ToLower(val) == "s" {
		if err := os.Remove(path); err != nil {
			fmt.Println("\033[1;31mError en la eliminacion del disco\033[0m")
		} else {
			fmt.Println("\033[1;32mEl disco a sido eliminado exitosamente\033[0m")
			desmontarDiscoEliminado(path)
		}
	} else {
		fmt.Println("\033[1;31mEl disco no a sido elminado\033[0m")
	}
}

func obtenerFecha() string {
	tiempo := time.Now()
	return strconv.Itoa(tiempo.Day()) + "/" + obtenerMes(tiempo.Month()) + "/" +
		strconv.Itoa(tiempo.Year()) + " " + strconv.Itoa(tiempo.Hour()) + ":" + strconv.Itoa(tiempo.Minute())
}

func obtenerMes(mes time.Month) string {
	salida := ""
	switch mes {
	case 1:
		salida = "01"
	case 2:
		salida = "02"
	case 3:
		salida = "03"
	case 4:
		salida = "04"
	case 5:
		salida = "05"
	case 6:
		salida = "06"
	case 7:
		salida = "07"
	case 8:
		salida = "08"
	case 9:
		salida = "09"
	case 10:
		salida = "10"
	case 11:
		salida = "11"
	case 12:
		salida = "12"
	}
	return salida
}
