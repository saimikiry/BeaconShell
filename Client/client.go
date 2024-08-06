package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Структура цели
type Target struct {
	name string
	conn net.Conn
}

// Предоставляет список встроенных команд (вызов: /BS help).
func BindShellHelp() {
	fmt.Println("[BindShell] Command list:")
	fmt.Println("\t- /BS add <ip> <port> \t\tAdd target <ip>:<port> to set;")
	fmt.Println("\t- /BS help \t\t\tShow this list;")
	fmt.Println("\t- /BS remove <index> \tRemove target from set;")
	fmt.Println("\t- /BS stop \t\t\tFinish session;")
	fmt.Println("\t- /BS targets \t\t\tShow current targets list.")
}

// Завершает работу программы (вызов: /BS stop)
func BindShellStop(targets *[]Target) {
	// Завершение всех активных соединений
	finishAllSessions(targets)

	// Завершение программы
	fmt.Println("Shutdown...")
	os.Exit(0)
}

// Выводит список целевых хостов (вызов: /BS targets)
func BindShellTargets(targets *[]Target) {
	fmt.Println("[BindShell] Current targets:")
	for i := 0; i < len(*targets); i++ {
		fmt.Printf("\t[%d] %s\n", i, (*targets)[i].name)
	}
}

// Удаляет выбранный хост из списка целей и завершает с ним сессию (вызов: /BS remove)
func BindShellRemove(targets *[]Target, idx int) {
	// Завершение сессии удаляемого хоста
	finishSession((*targets)[idx])

	fmt.Printf("[BindShell] %s << Host removed from target list.\n", (*targets)[idx].name)

	// Удаление хоста из списка целей
	*targets = append((*targets)[:idx], (*targets)[idx+1:]...)
}

// Устанавливает соединение с выбранным хостом и добавляет его в список целей (вызов: /BS add)
func BindShellAdd(targets *[]Target, target_name string) {
	addSession(targets, target_name)
}

func addSession(targets *[]Target, target_name string) {
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

func finishSession(target Target) {
	if target.conn.Close() == nil {
		fmt.Printf("[BindShell] %s << Connection successfully terminated.\n", target.name)
	}
}

func finishAllSessions(targets *[]Target) {
	for i := 0; i < len(*targets); i++ {
		finishSession((*targets)[i])
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

		// Добавление сессии
		addSession(targets, target_name)
	}
}

func BindShellRequest(request []string, targets *[]Target) {
	switch request[1] {
	case "add":
		BindShellAdd(targets, request[2])
		//fmt.Print(1)
	case "help":
		BindShellHelp()
	case "remove":
		idx, err := strconv.Atoi(request[2])
		if err != nil {
			fmt.Println("[BindShell] The index is an invalid number!")
		} else if idx >= len(*targets) || idx < 0 {
			fmt.Println("[BindShell] The index is out of bounds!")
		}
		BindShellRemove(targets, idx)
	case "stop":
		BindShellStop(targets)
	case "targets":
		BindShellTargets(targets)
	}
}

func getCommandResult(ctx context.Context, mtx *sync.Mutex, target Target) {
	buffer := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			target.conn.Read(buffer)
			(*mtx).Lock()
			fmt.Printf("[%s]\n", target.name)
			fmt.Println(string(buffer))
			(*mtx).Unlock()
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

	// Создание мьютекса для корректировки порядка io
	mtx := sync.Mutex{}

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
				go getCommandResult(ctx, &mtx, targets[i])
			}

			// Фактическое ожидание в течение 50 миллисекунд
			// TODO: изучить вопрос контекста и убрать лишнее создание горутин (?)
			// Или применять далее для масштабирования выполнения команд на нескольких удаленных хостах
			//fmt.Printf("R: %d\n", runtime.NumGoroutine())
			time.Sleep(100 * time.Millisecond)
		}
	}
}
