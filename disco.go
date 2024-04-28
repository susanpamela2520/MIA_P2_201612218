package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type MBR struct {
	Tamanio       int64
	FechaCreacion [20]byte
	NumeroRandom  int16
	TipoAjuste    byte
	Particiones   [4]Particion
}

var identificador int16 = 1;
var listaDiscos []string;


func crearDisco(tamanio int64, ajuste string, unidades string, ruta string) {

	listaDiscos = append(listaDiscos, ruta)

	particionVacia := Particion{Inicio: -1}
	discoAux := MBR{
		Tamanio:      obtenerTamanioDisco(tamanio, unidades),
		NumeroRandom: identificador,
		TipoAjuste:   ajuste[0],
		Particiones: [4]Particion{
			particionVacia,
			particionVacia,
			particionVacia,
			particionVacia}}

	copy(discoAux.FechaCreacion[:], obtenerFecha())

	exec.Command("mkdir", "-p", ruta).Output()
	exec.Command("rmdir", ruta).Output()

	if _, err := os.Stat(ruta); err == nil {
		mensaje += "El archivo ya existe, vuelva a intentarlo...\n"
		return
	}

	archivo, _ := os.Create(ruta)

	defer archivo.Close()


	buffer := bytes.NewBuffer([]byte{})               
	binary.Write(buffer, binary.BigEndian, uint8(0))  
	archivo.Write(buffer.Bytes())

	archivo.Seek(discoAux.Tamanio-int64(1), 0)
	archivo.Write(buffer.Bytes())


	archivo.Seek(0, 0)
	buffer.Reset()

	binary.Write(buffer, binary.BigEndian, &discoAux)
	archivo.Write(buffer.Bytes())

	identificador++

	mensaje += "¡ EL disco se a creado exitosamente !\n"

}

func eliminarDisco(ruta string) {

	fmt.Println("¿Esta seguro que desea eliminar el disco (si/no)?")
	condicion := "no"
	fmt.Scanln(&condicion)

	if condicion == "si"{
		error := os.Remove(ruta)
		if error != nil {
			fmt.Println("¡ Error al eleminar el disco, intentelo nuevamente !\n")
		} else {
			fmt.Println("¡ Disco eliminado exitosamente !\n")
		}
	}else{
		fmt.Println("Operacion cancelada con exito...\n")
	}
}


func obtenerDisco(path string) *os.File {
	if _, err := os.Stat(path); err == nil {
		archivo, _ := os.OpenFile(path, os.O_RDWR, 0644)
		return archivo
	}
	return nil
}

func obtenerFecha() string {

	tiempo := time.Now()

	fecha := fmt.Sprintf("%02d-%02d-%d %02d:%02d:%02d", tiempo.Day(), tiempo.Month(), tiempo.Year(),
		tiempo.Hour(), tiempo.Minute(), tiempo.Second())

	return fecha

}

func obtenerTamanioDisco(tamanio int64, unidades string) int64 {

	if (strings.Compare(unidades, "k")) == 0 {
		return int64(tamanio * 1024)
	} else if (strings.Compare(unidades, "m")) == 0 {
		return int64(tamanio * 1048576)
	}
	return int64(-1)
}
