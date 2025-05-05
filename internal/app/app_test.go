package app

import (
	"impulse_yadro/internal/storage"
	"os"
	"testing"
	"time"
)

// Тесты для функций времени
func TestTimeFunctions(t *testing.T) {
	t.Run("ParseTime valid", func(t *testing.T) {
		_, err := parseTime("12:34:56.789")
		if err != nil {
			t.Errorf("parseTime() error = %v", err)
		}
	})

	t.Run("FormatDuration", func(t *testing.T) {
		if got := formatDuration(time.Hour); got != "01:00:00.000" {
			t.Errorf("formatDuration() = %v", got)
		}
	})
}

// Тесты для работы с конфигурацией
func TestConfig(t *testing.T) {
	configContent := `{"laps":2,"lapLen":3651}`
	tmpFile, _ := os.CreateTemp("", "config-*.json")
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(configContent)
	tmpFile.Close()

	t.Run("ReadConfig valid", func(t *testing.T) {
		_, err := ReadConfig(tmpFile.Name())
		if err != nil {
			t.Errorf("ReadConfig() error = %v", err)
		}
	})
}

// Тесты для обработки событий
func TestEvents(t *testing.T) {
	config := &storage.Config{Laps: 2}
	competitors := make(map[int]*storage.Competitor)

	t.Run("Process registration", func(t *testing.T) {
		events := []storage.Event{{EventID: 1, CompetitorID: 1}}
		ProcessingEvents(competitors, events, config)
		if len(competitors) != 1 {
			t.Error("Competitor not registered")
		}
	})
}

// Тесты для формирования результатов
func TestResults(t *testing.T) {
	config := &storage.Config{Laps: 2, FiringLines: 1}
	competitors := map[int]*storage.Competitor{
		1: {ID: 1, Registered: true, Laps: make([]storage.Lap, 2)},
	}

	t.Run("ResultingTable", func(t *testing.T) {
		ResultingTable(competitors, config) // Проверяем что не паникует
	})
}

// Интеграционный тест
func TestIntegration(t *testing.T) {
	// Создаем тестовые файлы
	configFile, _ := os.CreateTemp("", "config-*.json")
	eventFile, _ := os.CreateTemp("", "events-*.txt")
	defer os.Remove(configFile.Name())
	defer os.Remove(eventFile.Name())

	configFile.WriteString(`{"laps":2,"lapLen":3651}`)
	eventFile.WriteString("[09:05:59.867] 1 1")
	configFile.Close()
	eventFile.Close()

	// Выполняем полный цикл
	config, _ := ReadConfig(configFile.Name())
	events, _ := ReadEvents(eventFile.Name())
	competitors := make(map[int]*storage.Competitor)
	ProcessingEvents(competitors, events, &config)
	ResultingTable(competitors, &config)
}
