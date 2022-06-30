package main

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	id       = ":7200"             //soket al que se conectara la funcion listen
	idserver = "192.168.1.29:8500" // ip del servidor (cambiar de acerdo donde se ejecute el servidor)
)

// encuentra la direccio ip del usuario
func ipuser() []byte {
	name, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
	}
	ip, err := net.LookupHost(name)
	if err != nil {
		fmt.Println(err)

	}
	x := []byte(ip[1])
	ln := make([]byte, 1)
	x = append(x, []byte(id)...)
	ln[0] = byte(len(x))
	x = append(x, ln...)
	return x
}

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
	b := make([]byte, 10000000)      // tamaño maximo del buffer de reccepcion
	sv, err := net.Listen("tcp", id) // conecarce al puerto en espera de comunicacion
	if err != nil {
		fmt.Println(err)
	}
	for {
		conn, err := sv.Accept()
		if err != nil {
			fmt.Println(err)
		}
		bl, err := conn.Read(b)
		sum := b[bl-16 : bl]
		sumch := md5.Sum(b[:bl-16])
		if string(sum) == string(sumch[:]) { //si el mensage llega correcto segimos adelante
			//fmt.Println(string(b[:5]))
			//fmt.Println(string(b[5 : bl-16]))
			d <- b[:bl-16]
		} else if string(sum) != string(sumch[:]) { // si el mensage no llego correcto enviamos que repita
			marcar([]byte("repit"), idserver)
			fmt.Println("error mensage")
		}
		b = make([]byte, 10000000)
	}
}

// enviamos el buffer y finalizamos la carga
func enviararchivo(archivo []byte, dr string, tipo []byte, idr []byte) {

	mss := []byte("subir") //comando
	mss = append(mss, archivo...)
	mss = append(mss, idr...)
	marcar(mss, dr)

	mss = []byte("final") //comando
	mss = append(mss, tipo...)
	mss = append(mss, idr...)
	fmt.Println("se envio el archivo a " + idserver)
	marcar(mss, dr)
}

// crea un string aleatorio
func radom_nombres() string {
	var nm []byte
	for i := 0; i < 7; i++ {
		tiempo := time.Now()
		semilla := int64(tiempo.UTC().UnixNano() + int64(i))
		rand.Seed(semilla)
		n := rand.Intn(25) + 65 //para qe quede en codigo ascii
		nm = append(nm, byte(n))

	}
	nombre := string(nm)
	return nombre
}

//esta a la espera para almacenar los datos del archivo y crearlo
func descargar(d chan []byte) {
	var arh []byte
	for {
		ms := <-d
		cm := ms[:5]
		datos := ms[5:]
		if string(cm) == "carga" {
			arh = append(arh, datos...)

		}
		if string(cm) == "final" {
			nombre := radom_nombres()
			//err := ioutil.WriteFile(nombre+string(datos), arh, 0644)
			err := os.WriteFile(nombre+string(datos), arh, 0666)
			if err != nil {
				fmt.Println("no se guardo el archivo")
				continue
			}
			fmt.Println("se descargo el archivo")
			arh = make([]byte, 0)
		}
	}

}

// extraemos la extencion del archivo
func exten(st string) []byte {
	var extencion []byte
	r := []byte(st)
	for i := 0; i < len(r); i++ {
		if r[i] == 46 {
			extencion = r[i:]
		}
	}
	return extencion
}
func comandos() {
	fmt.Println("comandos")
	fmt.Println("enviar archivo: send")
	fmt.Println("entrar canal 1: ch1")
	fmt.Println("entrar canal 2: ch2")
	fmt.Println("salir canal 1: desch1")
	fmt.Println("salir canal 2: desch2")
	fmt.Println("salir del programa: exit")
}

var s string
var drhost []byte

func main() {
	var path string
	d := make(chan []byte) //canal de comnicacion entreñas rutinas del receptor y la descarga
	drhost = ipuser()      //direccion del host
	comandos()
	go recive(d)
	go descargar(d)
	for {
		fmt.Scan(&s)
		if s == "exit" { //alimos del cliente
			break
		}
		if s == "ch1" { //suscrivirnos a canal 1
			ms := []byte("chan1") //comando
			idb := []byte(drhost)
			ms = append(ms, idb...)
			fmt.Println(string(ms))
			marcar(ms, idserver)
		}
		if s == "ch2" { //suscrivirnos a canal 1
			ms := []byte("chan2") //comando
			idb := []byte(drhost)
			ms = append(ms, idb...)
			fmt.Println(string(ms))
			marcar(ms, idserver)
		}
		//leemos y enviamos el archivo enviamos el archivo
		if s == "send" {
			fmt.Println("direccion del archivo")
			fmt.Scan(&path)
			//path_ex := []byte(path)
			tipo := exten(path)             //path_ex[len(path_ex)-4:]
			archv, err := os.ReadFile(path) //leer archivo
			if err != nil {
				fmt.Println("no se cargo el arcivo")
				continue
			}

			if len(archv) <= 10000000 {
				enviararchivo(archv, idserver, tipo, drhost)

			} else if len(archv) > 10000000 {
				fmt.Println("el archivo es demasiado pesado(limite 10Mb)")
			}
		}
		//salir del canal 1
		if s == "desch1" {

			ms := []byte("drop1") //comando
			idb := []byte(drhost)
			ms = append(ms, idb...)
			fmt.Println("saliste del canal 1")
			marcar(ms, idserver)
		}
		//salir del canal 2
		if s == "desch2" {

			ms := []byte("drop2") //comando
			idb := []byte(drhost)
			ms = append(ms, idb...)
			fmt.Println("saliste del canal 2")
			marcar(ms, idserver)
		}

	}
}
