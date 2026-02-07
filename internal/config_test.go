package internal

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// resetState сбрасывает все предыдущие значения в источниках конфигурации
// (флаги и переменные окружения).
func resetState(t *testing.T) {
	t.Helper()
	viper.Reset()
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
}

// TestMain является входной точкой (entrypoint) для тестов конфигурации. Она
// очищает все аргументы командной строки, чтобы go test флаги не конфликтовали
// с pflag внутренне, и тесты запускались.
func TestMain(m *testing.M) {
	os.Args = []string{os.Args[0]}
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	os.Exit(m.Run())
}

// TestLoadConfig_Defaults проверяет, что при отсутствии переменных окружения и
// консольных флагов конфигурация инициализируется значениями по умолчанию.
func TestLoadConfig_Defaults(t *testing.T) {
	resetState(t)
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.RunAddr != ":8080" {
		t.Errorf("RunAddr = %q, want %q", config.RunAddr, ":8080")
	}
	if config.DatabaseURI != "" {
		t.Errorf("DatabaseURI = %q, want empty", config.DatabaseURI)
	}
	if config.AccrualAddr != "" {
		t.Errorf("AccrualAddr = %q, want empty", config.AccrualAddr)
	}
}

// TestLoadConfig_FromEnv проверяет, что значения из переменных окружения
// корректно загружаются в конфигурацию и переопределяют значения по умолчанию.
func TestLoadConfig_FromEnv(t *testing.T) {
	resetState(t)
	t.Setenv("RUN_ADDRESS", ":9090")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.RunAddr != ":9090" {
		t.Errorf("RunAddr = %q, want %q", config.RunAddr, ":9090")
	}
}

// TestLoadConfig_FromFlag проверяет, что значения, переданные через консольные
// флаги, корректно загружаются в конфигурацию и переопределяют значения по
// умолчанию.
func TestLoadConfig_FromFlag(t *testing.T) {
	resetState(t)
	os.Args = []string{
		"cmd",
		"-a", ":7070",
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.RunAddr != ":7070" {
		t.Errorf("RunAddr = %q, want %q", config.RunAddr, ":7070")
	}
}

// TestLoadConfig_PriorityFlagOverEnv проверяет приоритет консольных флагов над
// переменными окружения в случае, когда одно и то же поле конфигурации задано
// обоими способами одновременно.
func TestLoadConfig_PriorityFlagOverEnv(t *testing.T) {
	resetState(t)

	t.Setenv("RUN_ADDRESS", ":9999")
	os.Args = []string{
		"cmd",
		"--a", ":7070",
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.RunAddr != ":7070" {
		t.Errorf("RunAddr = %q, want %q", config.RunAddr, ":7070")
	}
}
