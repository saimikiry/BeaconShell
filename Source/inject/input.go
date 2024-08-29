package main

import (
	"fmt"
)

func main() {
	var op1, op2 int
	var oper string
	fmt.Println("!Welcome to Fake Powerful Calc!")

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
