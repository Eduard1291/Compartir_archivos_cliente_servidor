package main

import (
	"crypto/md5"
	"fmt"
	"net"
)

const (
	ok       = "acept"
	ipserver = "192.168.1.29:8500" // ip del servidor (cambiar de acerdo donde se ejecute el servidor)
)

var s string
var rchan1 []string

// enviar un slice de bytes a una direccion
func marcar(ms []byte, dr string) {
	conn, err := net.Dial("tcp", dr)
	if err != nil {
		fmt.Println(err)
	}
	check := md5.Sum(ms) //agregamos el md5 para detectar modificaciones del mensage
	ms = append(ms, check[:]...)

	conn.Write(ms)
}

//gorrutina para estar a la espera de la comunicación con los clientes
func recive(d chan []byte) {
	b := make([]byte, 10000000)            // tamaño maximo del buffer de reccepcion
	sv, err := net.Listen("tcp", ipserver) // conecarce al puerto en espera de comunicacion
	if err != nil {
		fmt.Println(err)
	}
	for {
		conn, err := sv.Accept()
		if err != nil {
			fmt.Println("error al recivir")
		}
		bl, err := conn.Read(b)
		sum := b[bl-16 : bl]
		sumch := md5.Sum(b[:bl-16])
		if string(sum) == string(sumch[:]) { //si el mensage llega correcto segimos adelante
			lendr := b[bl-17]
			marcar([]byte(ok), string(b[bl-int(lendr)-17:bl-17]))
			d <- b[:bl-16]
		} else if string(sum) != string(sumch[:]) { // si el mensage no llego correcto enviamos que repita
			lendr := b[bl-17]
			dr := b[bl-int(lendr)-17 : bl-17]
			marcar([]byte("repit"), string(dr))
			fmt.Println("error mensage")
		}
		b = make([]byte, 10000000)

	}

}

// busca si un estring esta en un slice de string
func buscar(reg []string, ob string) int {
	estat := 0
	for i := 0; i < len(reg); i++ {
		if reg[i] == ob {
			estat = 1
		}
	}
	return estat
}

//remueve un string de un slice de string
func remove(reg []string, ob string) []string {
	var st []string
	for i := 0; i < len(reg); i++ {
		if reg[i] != ob {
			st = append(st, reg[i])
		}

	}
	return st
}

// enviamos el buffer y finalizamos la carga
func enviararchivo(archivo []byte, dr string, tipo []byte) {

	mss := []byte("carga")
	mss = append(mss, archivo...)
	marcar(mss, dr)

	ms := []byte("final")
	ms = append(ms, tipo...)
	marcar(ms, dr)

}

// gorrutina para responder en respuesta de los comandos del transmisor
func server(d chan []byte) {
	var rchan1 []string // regstro de usarios del canal 1 (se almacenan direcciones ip)
	var rchan2 []string //regstro de usarios del canal 2 (se almacenan direcciones ip)
	var archv []byte    // bufer de almacenamiento de datos  del canal 1
	var archv2 []byte   // bufer de almacenamiento de datos  del canal 2

	for {
		ms := <-d
		cm := ms[:5]
		lendr := ms[len(ms)-1]
		dr := ms[len(ms)-int(lendr)-1 : len(ms)-1]
		datos := ms[5 : len(ms)-int(lendr)-1]
		//suscrivirse a canal 2
		if string(cm) == "chan1" {
			sdr := string(dr)
			st := buscar(rchan1, sdr)
			if st != 1 {
				rchan1 = append(rchan1, sdr)

			}
			fmt.Println(rchan1)
		}
		//suscrivirse a canal 2
		if string(cm) == "chan2" {
			sdr := string(dr)
			st := buscar(rchan2, sdr)
			if st != 1 {
				rchan2 = append(rchan2, sdr)

			}
			fmt.Println(rchan2)
		}
		///////////////////////////////////////////////
		sdr := string(dr)
		st := buscar(rchan1, sdr)
		st2 := buscar(rchan2, sdr)
		//subir canal 1
		if string(cm) == "subir" && st == 1 {
			archv = append(archv, datos...)
		}
		// subir canal 2
		if string(cm) == "subir" && st2 == 1 {
			archv2 = append(archv2, datos...)
		}
		//finalisar la escritura de un archivo a los canales suscrito
		if string(cm) == "final" && st == 1 {
			// se envia al canal 1
			for i := 0; i < len(rchan1); i++ {
				if rchan1[i] != sdr {
					enviararchivo(archv, rchan1[i], datos)
					fmt.Println("se envio el archivo a " + rchan1[i])
				}

			}
			archv = make([]byte, 0)
		}
		//finalisar la escritura de un archivo a los canales suscrito
		if string(cm) == "final" && st2 == 1 {

			// se envia al canal 2
			for i := 0; i < len(rchan2); i++ {
				if rchan2[i] != sdr {
					//ioutil.WriteFile("imagen.jpg", archv2, 0666)
					enviararchivo(archv2, rchan2[i], datos)
					fmt.Println("se envio el archivo a " + rchan2[i])
				}

			}
			archv2 = make([]byte, 0)

		}

		// salir del canal 1
		if string(cm) == "drop1" {
			rchan1 = remove(rchan1, sdr)
			fmt.Println("registro canal 1")
			fmt.Println(rchan1)
		}
		// salir del  canal 2
		if string(cm) == "drop2" {
			rchan2 = remove(rchan2, sdr)
			fmt.Println("registro canal 2")
			fmt.Println(rchan2)
		}

	}
}

func main() {
	d := make(chan []byte) // canal de comunicacion entre las rutinas del receptor y el servidor
	go recive(d)
	go server(d)
	for { //cerrar el server
		fmt.Scan(&s)
		if s == "exit" {
			break
		}

	}
}
