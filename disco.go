package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type disco struct {
}

//mkdisk
func crearDisco(size int, path, name, unit string) {
	fmt.Println(size, " ", path, " ", name, " ", unit)

	//conversion a tamaño en kilobytes o megabytes
	if unit == "k" {
		size = size * 1024
	} else {
		size = size * 1048576
	}

	//creacion de carpetas necesarias para el almacenamiento del archivo
	exec.Command("mkdir", "-p", path).Output()

	//creacion o apertura del archivo
	archivo, _ := os.Create(path + "/" + name)
	defer archivo.Close()

	//llenado del archivo con valores de cero para obtener el tamaño especificado
	fmt.Println(size)
	for i := 0; i < size; i++ {
		if err := binary.Write(archivo, binary.LittleEndian, uint8(0)); err != nil {
			panic(err)
		}
	}
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
