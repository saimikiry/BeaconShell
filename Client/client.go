package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func BindShellHelp() {
	fmt.Println("[BindShell] Command list:")
	fmt.Println("\t- /BS add <ip> <port> \t\tAdd target <ip>:<port> to set; (TODO)")
	fmt.Println("\t- /BS help \t\t\tShow this list;")
	fmt.Println("\t- /BS remove <ip> <port> \tRemove target <ip>:<port> from set; (TODO)")
	fmt.Println("\t- /BS stop \t\t\tFinish session;")
	fmt.Println("\t- /BS switch <ip> <port> \tSwitch target to <ip>:<port>; (TODO)")
	fmt.Println("\t- /BS target \t\t\tShow current target.")
}

func BindShellStop() {
	fmt.Println("Shutdown...")
	os.Exit(0)
}

func BindShellSwitch(target *string, ip, port string) {
	*target = fmt.Sprintf("%s:%s", ip, port)
	conn, err := net.Dial("tcp", *target)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

}

func BindShellTarget(target string) {
	fmt.Printf("[BindShell] Current target is %s.\n", target)
}

func BindShellRequest(request []string, target *string) {
	switch request[1] {
	case "help":
		BindShellHelp()
	case "stop":
		BindShellStop()
	case "switch":
		BindShellSwitch(target, request[2], request[3])
	case "target":
		BindShellTarget(*target)
	}
}

func getCommandResult(ctx context.Context, conn net.Conn) {
	buffer := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn.Read(buffer)
			fmt.Println(string(buffer))
		}
	}
}

func main() {
	// Формирование цели
	target := fmt.Sprintf("%s:%s", os.Args[1], os.Args[2])

	// Установка соединения с целевым хостом
	conn, err := net.Dial("tcp", target)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// Создание объекта NewReader
	reader := bufio.NewReader(os.Stdin)

	//
	//buffer := make([]byte, 4096)

	// Инструктаж пользователя
	fmt.Printf("[BindShell] Successfully connected to %s.\n", target)
	fmt.Printf("[BindShell] Print command to send or \"/BS stop\" to finish BindShell.\n")
	for {
		fmt.Print("[BindShell] >> ")

		// Считывание строки
		input, _ := reader.ReadString('\n')

		// Удаление лишних символов
		input = input[:len(input)-2]

		// Разбиение команды на аргументы
		splitted_input := strings.Fields(input)

		// Проверка типа команды
		if splitted_input[0] == "/BS" {
			// Выполнение встроенной команды BindShell
			BindShellRequest(splitted_input, &target)
		} else {
			// Отправка команды на сервер
			_, err := conn.Write([]byte(input + "\n"))
			if err != nil {
				log.Fatalln(err)
				break
			}

			// Создание контекста для ожидания ответа в течение 20 миллисекунд
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()

			// Вызов горутины для ожидания ответа
			go getCommandResult(ctx, conn)

			// Фактическое ожидание в течение 50 миллисекунд
			// TODO: изучить вопрос контекста и убрать лишнее создание горутин (?)
			// Или применять далее для масштабирования выполнения команд на нескольких удаленных хостах
			// fmt.Println(runtime.NumGoroutine())
			time.Sleep(50 * time.Millisecond)
		}
	}
}
