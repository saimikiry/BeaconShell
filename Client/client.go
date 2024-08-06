package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
)

type Target struct {
	name string
	conn net.Conn
}

// Предоставляет список встроенных команд (вызов: /BS help).
func BindShellHelp() {
	fmt.Println("[BindShell] Command list:")
	fmt.Println("\t- /BS add <ip> <port> \t\tAdd target <ip>:<port> to set; (TODO)")
	fmt.Println("\t- /BS help \t\t\tShow this list;")
	fmt.Println("\t- /BS remove <ip> <port> \tRemove target <ip>:<port> from set; (TODO)")
	fmt.Println("\t- /BS stop \t\t\tFinish session;")
	fmt.Println("\t- /BS targets \t\t\tShow current targets list.")
}

// Завершает работу (вызов: /BS stop)
func BindShellStop(targets *[]Target) {
	// Завершение всех активных соединений
	finishAllSessions(targets)

	// Завергение программы
	fmt.Println("Shutdown...")
	os.Exit(0)
}

func BindShellTargets(targets *[]Target) {
	fmt.Println("[BindShell] Current targets:")
	for i := 0; i < len(*targets); i++ {
		fmt.Printf("\t- %s\n", (*targets)[i].name)
	}
}

func finishAllSessions(targets *[]Target) {
	for i := 0; i < len(*targets); i++ {
		if (*targets)[i].conn.Close() == nil {
			fmt.Printf("[BindShell] %s << Connection successfully terminated.\n", (*targets)[i].name)
		}
	}
}

func BindShellLoadTargets(targets *[]Target, targets_file string) {
	// Завершение всех активных соединений
	finishAllSessions(targets)

	// Очистка списка целей
	*targets = []Target{}

	// Открытие файла с новыми целями
	file, err := os.Open(targets_file)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	// Создание сканера для считывания
	scanner := bufio.NewScanner(file)

	// Построчное считывание файла
	for scanner.Scan() {
		// Получение имени хоста
		target_name := scanner.Text()

		// Установка соединения с целевым хостом
		conn, err := net.Dial("tcp", target_name)
		if err != nil {
			log.Fatalln(err)
		} else {
			fmt.Printf("[BindShell] %s << Connection successfully established.\n", target_name)
		}

		// Добавление хоста в список целей
		*targets = append(*targets, Target{name: target_name, conn: conn})
	}
}

func BindShellRequest(request []string, targets *[]Target) {
	switch request[1] {
	case "help":
		BindShellHelp()
	case "stop":
		BindShellStop(targets)
	case "target":
		BindShellTargets(targets)
	}
}

func getCommandResult(ctx context.Context, target Target) {
	buffer := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			target.conn.Read(buffer)
			fmt.Printf("%s\n", target.name)
			fmt.Println(string(buffer))
		}
	}
}

func main() {
	// Инициализация контейнера целей
	targets := make([]Target, 0, 1)

	// Считывание целей из файла
	BindShellLoadTargets(&targets, os.Args[1])

	// Создание объекта NewReader стандартного потока ввода
	reader := bufio.NewReader(os.Stdin)

	// Инструктаж пользователя
	fmt.Printf("[BindShell] Print \"/BS help\" to get info.\n")
	for {
		// Приглашение к вводу команды
		fmt.Print("[BindShell] >> ")

		// Считывание команды пользователя
		input, _ := reader.ReadString('\n')

		// Удаление лишних символов
		input = input[:len(input)-2]

		// Разбиение команды на аргументы
		splitted_input := strings.Fields(input)

		// Проверка типа команды
		if splitted_input[0] == "/BS" {
			// Выполнение встроенной команды BindShell
			BindShellRequest(splitted_input, &targets)
		} else {
			// Выполнение команды для каждого хоста
			for i := 0; i < len(targets); i++ {
				// Отправка команды на целевой хост
				_, err := targets[i].conn.Write([]byte(input + "\n"))
				if err != nil {
					log.Fatalln(err)
					break
				}

				// Создание контекста для ожидания ответа в течение 20 миллисекунд
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
				defer cancel()

				// Вызов горутины для ожидания ответа
				go getCommandResult(ctx, targets[i])
			}

			// Фактическое ожидание в течение 50 миллисекунд
			// TODO: изучить вопрос контекста и убрать лишнее создание горутин (?)
			// Или применять далее для масштабирования выполнения команд на нескольких удаленных хостах
			fmt.Printf("R: %d\n", runtime.NumGoroutine())
			time.Sleep(50 * time.Millisecond)
		}
	}
}
