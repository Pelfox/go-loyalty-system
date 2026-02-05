package internal

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config задаёт структуру конфигурации приложения.
type Config struct {
	// RunAddr - адрес, на котором необходимо запустить HTTP-сервер.
	RunAddr string `mapstructure:"run_addr"`
	// DatabaseURI - адрес подключения к базе данных PostgreSQL.
	DatabaseURI string `mapstructure:"database_uri"`
	// AccrualAddr - адрес для подключения к внешней системе лояльности.
	AccrualAddr string `mapstructure:"accrual_addr"`
}

// bindField привязывает поле из Config к указанной переменной окружения, а
// также консольному флагу, задавая значение по-умолчанию и описание (для флагов).
func bindField(key string, env string, flag string, defaultValue string, usage string) error {
	viper.SetDefault(key, defaultValue)
	if err := viper.BindEnv(key, env); err != nil {
		return err
	}

	pflag.StringP(flag, flag, defaultValue, usage)
	if err := viper.BindPFlag(key, pflag.Lookup(flag)); err != nil {
		return err
	}

	return nil
}

// LoadConfig загружает конфигурацию из переменных окружения и флагов (или же
// использует заданные значения по-умолчанию), и запаковывает их в структуру
// Config.
//
// Может вернуть ошибку, если не удаётся привязать поле к переменным или флагу,
// а также если не удаётся распаковать значения для конфигурации. Приоритет
// распределён так: флаги -> переменные окружения -> значения по умолчанию.
func LoadConfig() (Config, error) {
	var config Config

	if err := bindField("run_addr", "RUN_ADDRESS", "a", ":8080", "Адрес HTTP-сервера"); err != nil {
		return config, err
	}
	if err := bindField("database_uri", "DATABASE_URI", "d", "", "Адрес подключения к PostgreSQL"); err != nil {
		return config, err
	}
	if err := bindField("accrual_addr", "ACCRUAL_SYSTEM_ADDRESS", "r", "", "Адрес внутренней системы лояльности"); err != nil {
		return config, err
	}

	pflag.Parse()
	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}
