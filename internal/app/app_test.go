package app

import (
	"impulse_yadro/internal/storage"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseTime_ValidFormat(t *testing.T) {
	_, err := parseTime("12:34:56.789")
	if err != nil {
		t.Errorf("parseTime() error = %v, want nil", err)
	}
}

func TestParseTime_InvalidFormat(t *testing.T) {
	_, err := parseTime("invalid-time")
	if err == nil {
		t.Error("parseTime() should return error for invalid format")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"Hour", time.Hour, "01:00:00.000"},
		{"Complex", 2*time.Hour + 30*time.Minute + 15*time.Second + 123*time.Millisecond, "02:30:15.123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.duration); got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadConfig_ValidFile(t *testing.T) {
	configContent := `{"laps":2,"lapLen":3651}`
	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	config, err := ReadConfig(tmpFile.Name())
	if err != nil {
		t.Errorf("ReadConfig() error = %v, want nil", err)
	}

	if config.Laps != 2 {
		t.Errorf("ReadConfig() laps = %d, want 2", config.Laps)
	}
}

func TestProcessingEvents_Registration(t *testing.T) {
	config := &storage.Config{Laps: 2}
	competitors := make(map[int]*storage.Competitor)
	events := []storage.Event{{EventID: 1, CompetitorID: 1}}

	ProcessingEvents(competitors, events, config)

	if len(competitors) != 1 {
		t.Fatalf("ProcessingEvents() competitors count = %d, want 1", len(competitors))
	}

	c, ok := competitors[1]
	if !ok {
		t.Fatal("ProcessingEvents() competitor not found")
	}

	if !c.Registered {
		t.Error("ProcessingEvents() competitor should be registered")
	}
}

func TestIntegration_FullCycle(t *testing.T) {
	// Создаем тестовые файлы
	configFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configFile.Name())

	eventFile, err := os.CreateTemp("", "events-*.txt")
	if err != nil {
		t.Fatalf("Failed to create events file: %v", err)
	}
	defer os.Remove(eventFile.Name())

	// Записываем тестовые данные
	if _, err := configFile.WriteString(`{"laps":2,"lapLen":3651}`); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	if _, err := eventFile.WriteString("[09:05:59.867] 1 1\n[09:06:00.000] 2 1"); err != nil {
		t.Fatalf("Failed to write events: %v", err)
	}
	configFile.Close()
	eventFile.Close()

	// Выполняем тест
	config, err := ReadConfig(configFile.Name())
	if err != nil {
		t.Fatalf("ReadConfig() error: %v", err)
	}

	events, err := ReadEvents(eventFile.Name())
	if err != nil {
		t.Fatalf("ReadEvents() error: %v", err)
	}

	competitors := make(map[int]*storage.Competitor)
	ProcessingEvents(competitors, events, &config)

	if len(competitors) != 1 {
		t.Fatalf("Expected 1 competitor, got %d", len(competitors))
	}

	// Проверяем вывод
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ResultingTable(competitors, &config)

	w.Close()
	os.Stdout = old

	output, _ := io.ReadAll(r)
	if !strings.Contains(string(output), "1 ") {
		t.Error("Integration test output doesn't contain expected results")
	}
}

func parseTestTime(t *testing.T, s string) time.Time {
	tm, err := parseTime(s)
	if err != nil {
		t.Fatalf("Failed to parse test time: %v", err)
	}
	return tm
}
