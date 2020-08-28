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

var num_rand int16 = 0

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
		size = size * 1024
	} else {
		size = size * 1048576
	}

	//creacion de carpetas necesarias para el almacenamiento del archivo
	pathAux := "\"" + path + "\""
	exec.Command("mkdir", "-p", pathAux).Output()

	//validacion de la existencia del archivo
	if _, err := os.Stat(path + "/" + name); err == nil {
		fmt.Println("El archivo del disco ya existe")
		return
	}

	//creacion del archivo de simulacion del disco
	archivo, _ := os.Create(path + "/" + name)
	defer archivo.Close()

	//obtencion de la fecha
	tiempo := time.Now()
	tiempoActual := strconv.Itoa(tiempo.Day()) + "/" + obtenerMes(tiempo.Month()) + "/" +
		strconv.Itoa(tiempo.Year()) + " " + strconv.Itoa(tiempo.Hour()) + ":" + strconv.Itoa(tiempo.Minute())

	//instancia del mbr y llenado de datos del mismo
	mbr := discoMBR{}
	mbr.Tamanio = size
	mbr.Random = num_rand
	num_rand++
	copy(mbr.Creacion[:], tiempoActual)

	//llenado del archivo con valores de cero para obtener el tamaño especificado
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))
	archivo.Write(buffer.Bytes())

	//corrimiento del puntero del archivo para alcanzar el tamaño especificado
	archivo.Seek(size-int64(1), 0)
	archivo.Write(buffer.Bytes())

	//escritura del struct que representa el mbr en el archivo
	archivo.Seek(0, 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, &mbr)
	archivo.Write(buffer.Bytes())

	fmt.Println("Se a creado con exito el disco")
}

//rmdisk
func eliminarDisco(path string) {
	fmt.Print("Seguro que desea eliminar el archivo ? [S/N]: ")
	val := ""
	fmt.Scanln(&val)
	if strings.ToLower(val) == "s" {

		if err := os.Remove(path); err != nil {
			fmt.Println("Error en la eliminacion del archivo")
		} else {
			fmt.Println("El archivo a sido eliminado exitsamente")
		}
	}
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
