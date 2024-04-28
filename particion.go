package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unsafe"
)


type Particion struct {
	Estado  byte
	Tipo    byte
	Ajuste  byte
	Inicio  int64
	Tamanio int64
	Nombre  [20]byte
}

type EBR struct{
	Estado  byte
	Tipo    byte
	Inicio  int64
	Tamanio int64
	Siguiente int64
	Nombre  [20]byte
}

func insertarParticion(tamanio int64, unidad string, ruta string, tipoParticion string, ajuste string, nombre string) {

	nombreByte := [20]byte{}
	contadorParticiones := 0
	contadorExtendida := 0

	nuevaParticion := Particion{
		Estado:  '1',
		Tipo:    tipoParticion[0],
		Ajuste:  ajuste[0],
		Tamanio: obtenerTamanioParticion(tamanio, unidad),
	}
	copy(nuevaParticion.Nombre[:], nombre)

	archivo, _ := os.OpenFile(ruta, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		mensaje += "Disco no existe, no es posible crear una particion sin un disco..\n"
		return
	}

	discoAux := obtenerMBR(archivo)

	// Validaciones antes de hacer la insercion.
	for i := 0; i < 4; i++ {
		if discoAux.Particiones[i].Inicio != -1 {
			if discoAux.Particiones[i].Tipo == 'e' {
				contadorExtendida = 1
			}
			contadorParticiones++
		}
	}
	if contadorParticiones == 4 && (tipoParticion == "e" || tipoParticion == "p"){
		mensaje += "Numero de particiones insertas a llegado a su limite..\n"
		return
	}

	if contadorExtendida == 1 && tipoParticion == "e" {
		mensaje += "Numero de particiones extendidas a llegado a su limite..\n"
		return
	}

	copy(nombreByte[:], nombre)
	for i := 0; i < 4; i++ {
		if discoAux.Particiones[i].Inicio != -1 {
			if discoAux.Particiones[i].Nombre == nombreByte {
				mensaje += "No pueden existir 2 particines con el mismo nombre.\n"
				break
			}
		}
	}

	if tamanio < 0 {
		mensaje += "El tamaño de la particion no puede ser 0.\n"
		return
	}


	espacio := discoAux.Tamanio - int64(unsafe.Sizeof(discoAux))
	if tamanio > espacio {
		mensaje += "EL tamaño de la particion no puede ser mayor al disco.\n"
		return
	}

	for i := 0; i < 4; i++ {
		if tipoParticion == "p" || tipoParticion == "e" {
			if discoAux.Particiones[i].Inicio == -1 {

				nuevaParticion.Inicio = int64(unsafe.Sizeof(discoAux))
				discoAux.Particiones[0] = nuevaParticion
				escribirEnElDisco(archivo, discoAux)
				
				mensaje += "¡ Particion creada exitosamente !\n"
				mensaje += "*************************************************\n"
				mensaje += "************ DATOS PARTICIONES ************\n"
				mensaje += "Nombre : " + cadenaLimpia(discoAux.Particiones[i].Nombre[:]) + "\n"
				mensaje += "Estado : " + string(discoAux.Particiones[i].Estado) + "\n"
				mensaje += "Ajuste : " + string(discoAux.Particiones[i].Ajuste) + "\n"
				mensaje += "Inicio : " + string(discoAux.Particiones[i].Inicio) + "\n"
				mensaje += "Tamaño : " + string(discoAux.Particiones[i].Tamanio) + "\n"
				mensaje += "Tipo : " + string(discoAux.Particiones[i].Tipo) + "\n"
				mensaje += "************************************************* \n"
				break

			} else {

				posicionLibre := 0
				inicioParticionAnterior := int64(0)
				for j := 1; j < 4; j++ {
					if discoAux.Particiones[j].Inicio == -1 {

						inicioParticionAnterior = int64(discoAux.Particiones[j-1].Inicio)
						nuevaParticion.Inicio = inicioParticionAnterior + discoAux.Particiones[j-1].Tamanio
						posicionLibre = j
						break
					}
				}

				discoAux.Particiones[posicionLibre] = nuevaParticion
				escribirEnElDisco(archivo, discoAux)
				mensaje += "¡ Particion creada exitosamente !\n"
				mensaje += "*************************************************\n"
				mensaje += "************ DATOS PARTICIONES ************\n"
				mensaje += "Nombre : " + cadenaLimpia(discoAux.Particiones[i].Nombre[:]) + "\n"
				mensaje += "Estado : " + string(discoAux.Particiones[i].Estado) + "\n"
				mensaje += "Ajuste : " + string(discoAux.Particiones[i].Ajuste) + "\n"
				mensaje += "Inicio : " + string(discoAux.Particiones[i].Inicio) + "\n"
				mensaje += "Tamaño : " + string(discoAux.Particiones[i].Tamanio) + "\n"
				mensaje += "Tipo : " + string(discoAux.Particiones[i].Tipo) + "\n"
				mensaje += "*************************************************" + "\n"
				break
			}

		} else {

			if contadorExtendida == 1 {
				/* tamanioExtendida, inicioExtendida := obtenerExtendida(discoAux)

				validaTamanio := tamanioExtendida - int64(unsafe.Sizeof(EBR{}))
				tamanioParticion := tamanio - int64(unsafe.Sizeof(EBR{}))

				if validaTamanio <  tamanioParticion{
					
				} */
				fmt.Println("LOGICAS ")
				break
			} else {
				fmt.Println( "¡ Error, no existe una particion extendida !")
				break
			}

		}
	}
}

func borrarParticion(tipoEliminar string, ruta string, nombre string){
	//obtencion del archivo que simula el disco
	archivo := obtenerDisco(ruta)
	defer archivo.Close()

	//verificacion de existencia del archivo
	if archivo == nil {
		fmt.Println( "El disco aun no a sido creado")
		return
	}

	mbrAux := obtenerMBR(archivo)
	nombreByte := [20]byte{}
	copy(nombreByte[:], nombre)

	fmt.Println("¿Esta seguro que desea eliminar el disco (si/no)?")
	condicion := "no"
	fmt.Scanln(&condicion)

	if condicion == "si" && condicion != ""{
		for i := 0; i < 4; i++ {
			if mbrAux.Particiones[i].Nombre == nombreByte{
				
				mbrAux.Particiones[i].Ajuste = ' '
				mbrAux.Particiones[i].Estado = ' '
				mbrAux.Particiones[i].Inicio = -1
				mbrAux.Particiones[i].Nombre = [20]byte{}
				mbrAux.Particiones[i].Tamanio = -1
				mbrAux.Particiones[i].Tipo = ' '
				
				escribirEnElDisco(archivo, mbrAux)
				fmt.Println("Particion borrada exitosamente !!! ")
				break
			}
		}
	}else{
		fmt.Println("Opcion cancelada o vacia, intentelo de nuevo")
	}
}

func agregarParticion(ruta string, nombre string, tamanio int64, unidad string){
	
	archivo := obtenerDisco(ruta)
	defer archivo.Close()

	
	if archivo == nil {
		fmt.Println("El disco aun no a sido creado")
		return
	}

	mbrAux := obtenerMBR(archivo)
	nombreByte := [20]byte{}
	copy(nombreByte[:], nombre)
	tamanioNuevo := obtenerTamanioDisco(tamanio, unidad)
	cambioRealizado := false

	for i := 0; i < 4; i++{
		if mbrAux.Particiones[i].Inicio != -1{
			if mbrAux.Particiones[i].Nombre == nombreByte{
				if tamanioNuevo > 0 {
					validacion := mbrAux.Tamanio - (mbrAux.Particiones[i].Tamanio + mbrAux.Particiones[i].Inicio)
					if validacion >= tamanioNuevo {
						mbrAux.Particiones[i].Tamanio += tamanioNuevo
						cambioRealizado = true
					}else{
						fmt.Println("El nuevo tamanio de la particion no puede ser mayor al tamanio del disco")
						break
					}
				}else {
					validacion := mbrAux.Particiones[i].Tamanio + tamanioNuevo
					if validacion < 0 {
						fmt.Println("El nuevo tamanio de la particion no puede ser negativo...")
						break
					}
					mbrAux.Particiones[i].Tamanio += tamanioNuevo
					cambioRealizado = true;
				}
			}
		}
	}

	if(cambioRealizado){
		escribirEnElDisco(archivo, mbrAux)
		fmt.Println("Modificacion realizada exitosamente")
	}else{
		fmt.Println("Error, no es posible realizar el cambio, vuelva a intentarlo..")
	}
	
}

func obtenerTamanioParticion(tamanio int64, unidades string) int64 {

	if (strings.Compare(unidades, "k")) == 0 {
		return int64(tamanio * 1024)
	} else if (strings.Compare(unidades, "b")) == 0 {
		return int64(tamanio)
	} else if (strings.Compare(unidades, "m")) == 0 {
		return int64(tamanio * 1048576)
	}

	return int64(-1)
}

func obtenerExtendida(discoAux MBR) (int64, int64){

	tamanioExtendida := int64(0)
	inicioExtendida := int64(0)

	for i := 0; i < 4; i++ {
		if discoAux.Particiones[i].Inicio != -1 && discoAux.Particiones[i].Tipo == 'e' {
			inicioExtendida = discoAux.Particiones[i].Inicio
			tamanioExtendida = discoAux.Particiones[i].Tamanio
			break
		}
	}
	return tamanioExtendida, inicioExtendida
}

func obtenerMBR(archivo *os.File) MBR {

	discoAux := MBR{}
	contenido := make([]byte, int(unsafe.Sizeof(discoAux)))
	archivo.Seek(0, 0)
	archivo.Read(contenido)
	buffer := bytes.NewBuffer(contenido)
	binary.Read(buffer, binary.BigEndian, &discoAux)

	return discoAux
}

func escribirEnElDisco(archivo *os.File, discoAux MBR) {

	archivo.Seek(0, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, &discoAux)
	archivo.Write(buffer.Bytes())
}

func escribirEnElDisco2(archivo *os.File, discoAux MBR, posicion int64) {

	archivo.Seek(posicion, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, &discoAux)
	archivo.Write(buffer.Bytes())
}