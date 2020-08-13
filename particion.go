package main

type particion struct {
	estado  [1]byte
	tipo    [1]byte
	ajuste  [1]byte
	inicio  int64
	tamanio int64
	nombre  [16]byte
}

type logicaEBR struct {
	estado    [1]byte
	ajuste    [1]byte
	inicio    int64
	tamanio   int64
	siguiente int64
	nombre    [16]byte
}
