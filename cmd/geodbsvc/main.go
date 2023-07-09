package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fsnotify/fsnotify"
	"github.com/kardianos/service"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"

	sqlitedb "geodbsvc/internal/database/sqlite"
	"geodbsvc/internal/loggerx"
	"geodbsvc/internal/program"
	"geodbsvc/internal/utils"
)

func main() {
	// Парсируем командную строку
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()

	// Определяем директории
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	execDir, _ := filepath.Split(execPath)

	// Читаем конфигурационный файл
	cfg := viper.New()

	cfg.SetDefault("service.name", "geodbsvc")                                                     // Имя службы
	cfg.SetDefault("service.display_name", "GeoDB Service")                                        // Отображаемое имя службы
	cfg.SetDefault("service.description", "GeoDB Service")                                         // Описание службы
	cfg.SetDefault("server.host", "")                                                              // Хост сервера
	cfg.SetDefault("server.port", 8080)                                                            // Порт сервера
	cfg.SetDefault("log.enabled", false)                                                           // Вести log-файл?
	cfg.SetDefault("log.file", filepath.Join(execDir, "logs", "geodbsvc.log"))                     // Путь до log-файла
	cfg.SetDefault("database.dsn", fmt.Sprintf("file:%s", filepath.Join(execDir, "GeoDB.sqlite"))) // Путь до базы данных SQLite

	cfg.SetConfigName("geodbsvc")
	cfg.SetConfigType("yaml")

	switch runtime.GOOS {
	case "linux":
		cfg.AddConfigPath("/etc/geodbsvc")
		cfg.AddConfigPath("$HOME/.config/geodbsvc")
	case "windows":
		cfg.AddConfigPath(filepath.Join(os.Getenv("PROGRAMDATA"), "GeoDB-Service"))
	}
	cfg.AddConfigPath(filepath.Join(execDir, "configs"))

	if err := cfg.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// Настраиваем логирование
	logger := loggerx.New(&lumberjack.Logger{
		Filename:   cfg.GetString("log.file"), // Путь к файлу лога
		MaxSize:    5,                         // Максимальный размер в мегабайтах
		MaxAge:     30,                        // Количество дней для хранения старых логов
		MaxBackups: 10,                        //
		LocalTime:  true,                      //
		Compress:   true,                      // Сжимать в gzip-архивы
	})
	logger.SetEnabled(cfg.GetBool("log.enabled"))
	log.SetOutput(logger)

	// Открываем БД
	d, err := sqlitedb.New(cfg.GetString("database.dsn"))
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	// Создаём программу
	prg := program.New(cfg, d, logger)

	// Задаём настройки для службы
	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"

	svcConfig := &service.Config{
		Name:        cfg.GetString("service.name"),
		DisplayName: cfg.GetString("service.display_name"),
		Description: cfg.GetString("service.description"),
		Option:      options,
	}
	if runtime.GOOS == "linux" {
		svcConfig.Dependencies = []string{
			"Requires=network.target",
			"After=network-online.target syslog.target",
		}
	}

	// Создаём службу
	svc, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	errs := make(chan error, 5)

	// Открываем системный логгер
	svcLogger, err := svc.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	// Вывод ошибок
	go func() {
		for {
			if err := <-errs; err != nil {
				log.Print(err)
			}
		}
	}()

	// Управление службой
	if len(*svcFlag) != 0 {
		if !utils.Contains(service.ControlAction[:], *svcFlag, true) {
			fmt.Fprintf(os.Stdout, "Valid actions: %q\n", service.ControlAction)
		} else if err := service.Control(svc, *svcFlag); err != nil {
			fmt.Fprintln(os.Stdout, err)
		}
		return
	}

	log.Println(`Used config file "` + cfg.ConfigFileUsed() + `"`)

	// Следим за изменениями конфигурационного файла
	cfg.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
		logger.SetEnabled(cfg.GetBool("log.enabled"))
	})
	cfg.WatchConfig()

	// Запускаем службу
	if err := svc.Run(); err != nil {
		svcLogger.Error(err)
	}
}
