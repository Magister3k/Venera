// main.go - Точка входа в приложение Venera
//
// Этот файл содержит основную логику запуска приложения Venera,
// включая обработку командной строки, проверку прав администратора,
// запуск в режиме системы трей или службы Windows.
//
// Основные функции:
// - Обработка командной строки с поддержкой всех параметров
// - Проверка прав администратора с уведомлением при отсутствии
// - Запуск в режиме системы трей или службы Windows (настраивается в config.toml)
// - Проверка наличия уже запущенного процесса
// - Проверка наличия файла конфигурации
// - Инициализация всех модулей: логирование, базы данных, веб-сервер
// - Управление процессами обработки данных
//
// Использование:
// go run main.go
// venera.exe --help
// venera.exe --diagnose
// venera.exe --install_srv
// venera.exe --remove_cachedb
// venera.exe --create_pg_db
//
// Параметры командной строки:
// -h, --help                 Показать справку
// -v, --version              Показать версию программы
// -d, --diagnose             Запустить диагностику системы
// -c, --create_cachedb       Создать контейнер с СУБД DragonflyDB
// -r, --remove_cachedb       Удалить контейнер с СУБД DragonflyDB
// -p, --create_pg_db         Создать базу в СУБД PostgreSQL
// -i, --install_srv          Установить программу в качестве службы Windows
// -u, --uninstall_srv        Удалить службу программы из системы

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"venera/config"
	"venera/data"
	"venera/diagnose"
	"venera/logging"
	"venera/manifest"
	"venera/notify"
	"venera/processes"
	"venera/services"
	"venera/tray"
	"venera/web"
)

// Version - текущая версия приложения
const Version = "1.0.0"

// Global variables - глобальные переменные приложения
var (
	cfg            *config.Config
	dragonflyDB    *data.DragonflyDB
	postgresDB     *data.PostgreSQL
	dataFilter     *data.DataFilter
	controls       map[string]string
	alerts         map[string]map[string]interface{}
	processManager *processes.ProcessManager
	logger         *logging.Logger
	webServer      *web.Server
	trayIcon       *tray.Tray
)

func main() {
	// Обработка аргументов командной строки
	if err := handleCommandLine(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

	// Создание логгера (Singleton)
	logger = logging.GetLogger()
	if err := logger.Initialize("config.toml"); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка инициализации логгера: %v\n", err)
		os.Exit(1)
	}
	defer logger.Shutdown()

	// Логирование версии приложения
	logger.Infof("Venera version %s starting", Version)

	// Регистрация манифеста приложения
	if err := manifest.RegisterManifest(Version); err != nil {
		logger.Errorf("Ошибка регистрации манифеста: %v", err)
		// Вывод GUI-уведомления (для Windows)
		services.ShowNotification("Ошибка регистрации", "Проверьте манифест приложения")
	}

	// Проверка прав администратора
	if !isAdmin() {
		logger.Error("Приложение должно быть запущено с правами администратора")
		fmt.Println("Ошибка: Приложение должно быть запущено с правами администратора")
		fmt.Println("Пожалуйста, запустите приложение от имени администратора")
		os.Exit(1)
	}

	// Загрузка конфигурации
	var err error
	cfg, err = config.LoadConfig("config.toml")
	if err != nil {
		logger.Errorf("Ошибка загрузки конфигурации: %v", err)
		fmt.Fprintf(os.Stderr, "Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	// Проверка наличия уже запущенного процесса (механизм предотвращения дублирования)
	if hasRunningProcess() {
		logger.Info("Обнаружен уже запущенный процесс, открытие веб-интерфейса")
		openWebInterface()
		return
	}

	// Запуск Podman (если используется)
	if err := ensurePodmanRunning(); err != nil {
		logger.Warnf("Не удалось запустить Podman: %v", err)
	}

	// Инициализация подключения к DragonflyDB
	dragonflyDB, err = data.NewDragonflyDB(cfg.DragonflyDB)
	if err != nil {
		logger.Errorf("Ошибка подключения к DragonflyDB: %v", err)
		fmt.Fprintf(os.Stderr, "Ошибка подключения к DragonflyDB: %v\n", err)
		os.Exit(1)
	}
	defer dragonflyDB.Close()

	// Инициализация подключения к PostgreSQL
	postgresDB, err = data.NewPostgreSQL(cfg.PostgreSQL)
	if err != nil {
		logger.Errorf("Ошибка подключения к PostgreSQL: %v", err)
		fmt.Fprintf(os.Stderr, "Ошибка подключения к PostgreSQL: %v\n", err)
		os.Exit(1)
	}
	defer postgresDB.Close()

	// Загрузка фильтров данных
	dataFilter, err = data.NewDataFilter(cfg.GetFilterFilePath())
	if err != nil {
		logger.Warnf("Ошибка загрузки фильтров: %v", err)
		// Продолжаем работу без фильтров
		dataFilter = &data.DataFilter{Whitelist: make(map[string][]string), Blacklist: make(map[string][]string)}
	}

	// Загрузка контрольных значений
	controls, err = config.LoadControlFile(cfg.GetControlFilePath())
	if err != nil {
		logger.Warnf("Ошибка загрузки контрольных значений: %v", err)
		controls = make(map[string]string)
	}

	// Загрузка алертов
	alerts, err = config.LoadAlertsFile(cfg.GetAlertsFilePath())
	if err != nil {
		logger.Warnf("Ошибка загрузки алертов: %v", err)
		alerts = make(map[string]map[string]interface{})
	}

	// Создание менеджера процессов
	processManager = processes.NewProcessManager(cfg, dragonflyDB, postgresDB, dataFilter, controls, alerts)

	// Создание веб-сервера
	webServer = web.NewServer(cfg, processManager, dragonflyDB, postgresDB, logger)

	// Запуск в зависимости от режима
	if cfg.Generic.Mode == "service" {
		runAsService()
	} else {
		runAsTray()
	}
}

// handleCommandLine - обработка командной строки
func handleCommandLine() error {
	if len(os.Args) < 2 {
		return nil
	}

	arg := strings.ToLower(os.Args[1])

	switch arg {
	case "-h", "--help":
		printHelp()
		os.Exit(0)
	case "-v", "--version":
		fmt.Printf("Venera version %s\n", Version)
		os.Exit(0)
	case "-d", "--diagnose":
		return runDiagnose()
	case "-p", "--create_pg_db":
		return createPostgreSQLDatabase()
	case "-c", "--create_cachedb":
		return createCachedb()
	case "-r", "--remove_cachedb":
		return removeCachedb()
	case "-i", "--install_srv":
		return installService()
	case "-u", "--uninstall_srv":
		return uninstallService()
	default:
		return fmt.Errorf("неизвестный аргумент: %s", os.Args[1])
	}
}

// printHelp - вывод справки о параметрах командной строки
func printHelp() {
	fmt.Println("Venera - Система сбора идентификаторов в потоке пакетных данных")
	fmt.Println()
	fmt.Println("Использование:")
	fmt.Println("  venera.exe [опция]")
	fmt.Println()
	fmt.Println("Опции:")
	fmt.Println("  -h, --help                 Показать это сообщение")
	fmt.Println("  -v, --version              Показать версию программы")
	fmt.Println("  -d, --diagnose             Запустить диагностику системы")
	fmt.Println("  -p, --create_pg_db         Создать базу в СУБД PostgreSQL")
	fmt.Println("  -c, --create_cachedb       Создать контейнер с кэширующей СУБД")
	fmt.Println("  -r, --remove_cachedb       Удалить контейнер с кэширующей СУБД")
	fmt.Println("  -i, --install_srv          Установить программу в качестве службы Windows")
	fmt.Println("  -u, --uninstall_srv        Удалить службу программы из системы")
	fmt.Println()
	fmt.Println("Примеры:")
	fmt.Println("  venera.exe")
	fmt.Println("  venera.exe --diagnose")
	fmt.Println("  venera.exe -d")
	fmt.Println("  venera.exe --install_srv")
}

// runDiagnose - запуск диагностики системы
func runDiagnose() error {
	logger.Info("Запуск диагностики системы")

	diag := diagnose.NewDiagnose(cfg, dragonflyDB, postgresDB, logger)
	report, err := diag.Run()
	if err != nil {
		logger.Errorf("Ошибка диагностики: %v", err)
		return err
	}

	fmt.Println("Результаты диагностики:")
	fmt.Println(report)
	return nil
}

// createPostgreSQLDatabase - создание базы данных PostgreSQL
func createPostgreSQLDatabase() error {
	logger.Info("Создание базы PostgreSQL")

	postgresDB, err := data.NewPostgreSQL(cfg.PostgreSQL)
	if err != nil {
		logger.Errorf("Ошибка подключения к PostgreSQL: %v", err)
		return err
	}
	defer postgresDB.Close()

	// Создание базы данных
	if err := postgresDB.CreateDatabaseIfNotExists(cfg.PostgreSQL.Database); err != nil {
		logger.Errorf("Ошибка создания базы данных: %v", err)
		return err
	}

	logger.Info("База данных создана успешно")
	fmt.Println("База данных 'Venera' успешно создана")
	return nil
}

// createCachedb - создание контейнера DragonflyDB через Podman
func createCachedb() error {
	logger.Info("Создание контейнера cachedb")

	// Проверка запущенного процесса podman.exe
	if err := ensurePodmanRunning(); err != nil {
		logger.Errorf("Ошибка запуска Podman: %v", err)
		return err
	}

	// Создание контейнера
	if err := services.CreateCachedb(cfg.Paths.PodmanPath, cfg.Paths.DragonflyImage); err != nil {
		logger.Errorf("Ошибка создания контейнера: %v", err)
		return err
	}

	logger.Info("Контейнер cachedb создан успешно")
	fmt.Println("Контейнер cachedb успешно создан")
	return nil
}

// removeCachedb - удаление контейнера DragonflyDB
func removeCachedb() error {
	logger.Info("Удаление контейнера cachedb")

	// Удаление контейнера
	if err := services.RemoveCachedb(cfg.Paths.PodmanPath); err != nil {
		logger.Errorf("Ошибка удаления контейнера: %v", err)
		return err
	}

	logger.Info("Контейнер cachedb удален успешно")
	fmt.Println("Контейнер cachedb успешно удален")
	return nil
}

// installService - установка службы Windows
func installService() error {
	logger.Info("Установка службы Windows")

	// Проверка прав администратора
	if !isAdmin() {
		logger.Error("Для установки службы требуются права администратора")
		return fmt.Errorf("требуются права администратора")
	}

	// Установка службы
	if err := services.InstallService("VeneraSrv", services.GetExecutablePath()); err != nil {
		logger.Errorf("Ошибка установки службы: %v", err)
		return err
	}

	logger.Info("Служба установлена успешно")
	fmt.Println("Служба 'VeneraSrv' успешно установлена")
	return nil
}

// uninstallService - удаление службы Windows
func uninstallService() error {
	logger.Info("Удаление службы Windows")

	// Удаление службы
	if err := services.UninstallService("VeneraSrv"); err != nil {
		logger.Errorf("Ошибка удаления службы: %v", err)
		return err
	}

	logger.Info("Служба удалена успешно")
	fmt.Println("Служба 'VeneraSrv' успешно удалена")
	return nil
}

// isAdmin - проверка прав администратора на Windows
func isAdmin() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(),
		windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	var elevation windows.TokenElevation
	size := uint32(unsafe.Sizeof(elevation))
	err = windows.GetTokenInformation(token,
		windows.TokenElevation, (*byte)(unsafe.Pointer(&elevation)),
		size, &size)
	if err != nil {
		return false
	}

	return elevation.TokenIsElevated > 0
}

// hasRunningProcess - проверка наличия уже запущенного процесса
// Использует мьютекс для предотвращения дублирования
func hasRunningProcess() bool {
	// Создание именованного мьютекса
	mutex, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr("VeneraSingleInstance"))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(mutex)

	// Проверка, является ли мьютекс уже существующим
	if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		return true
	}

	return false
}

// openWebInterface - открытие веб-интерфейса в браузере по умолчанию
func openWebInterface() {
	url := fmt.Sprintf("http://localhost:%d", cfg.Generic.WebServerPort)
	fmt.Printf("Открытие веб-интерфейса: %s\n", url)

	// Открытие в браузере по умолчанию
	services.OpenURL(url)
}

// ensurePodmanRunning - проверка и запуск Podman если необходимо
func ensurePodmanRunning() error {
	// Проверка наличия файла podman.exe
	if _, err := os.Stat(cfg.Paths.PodmanPath); os.IsNotExist(err) {
		return fmt.Errorf("файл podman.exe не найден: %s", cfg.Paths.PodmanPath)
	}

	// Запуск podman info для проверки работоспособности
	cmd := fmt.Sprintf("\"%s\" info", cfg.Paths.PodmanPath)
	if err := services.RunCommand(cmd); err != nil {
		return fmt.Errorf("ошибка подключения к Podman: %w", err)
	}

	return nil
}

// runAsTray - запуск в режиме системы трей
func runAsTray() {
	logger.Info("Запуск в режиме системы трей")

	// Создание иконки трея
	trayIcon = tray.NewTray(logger, cfg, processManager, webServer)
	trayIcon.Initialize()

	// Запуск веб-сервера
	webServer.Start()

	// Запуск процессов, если включено
	if cfg.Generic.AutoStartProcesses {
		logger.Info("Автозапуск процессов обработки")
		processManager.StartAll()
	}

	// Обработка сигналов остановки
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала остановки
	logger.Info("Ожидание сигнала остановки...")
	<-sigChan

	logger.Info("Получен сигнал остановки")

	// Остановка процессов
	logger.Info("Остановка процессов обработки")
	processManager.StopAll()

	// Остановка веб-сервера
	logger.Info("Остановка веб-сервера")
	webServer.Stop()

	// Остановка трея
	logger.Info("Остановка системного трея")
	trayIcon.Stop()

	logger.Info("Приложение остановлено")
}

// runAsService - запуск в режиме службы Windows
func runAsService() {
	logger.Info("Запуск в режиме службы Windows")

	// Создание иконки трея для управления службой
	trayIcon = tray.NewTray(logger, cfg, processManager, webServer)
	trayIcon.Initialize()

	// Запуск веб-сервера
	webServer.Start()

	// Запуск процессов, если включено
	if cfg.Generic.AutoStartProcesses {
		logger.Info("Автозапуск процессов обработки")
		processManager.StartAll()
	}

	// Запуск службы Windows
	services.RunService("VeneraSrv", func() {
		logger.Info("Служба запущена")

		// Обработчик остановки службы
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		<-sigChan

		logger.Info("Получен сигнал остановки службы")

		// Остановка процессов
		logger.Info("Остановка процессов обработки")
		processManager.StopAll()

		// Остановка веб-сервера
		logger.Info("Остановка веб-сервера")
		webServer.Stop()

		// Остановка трея
		logger.Info("Остановка системного трея")
		trayIcon.Stop()

		logger.Info("Служба остановлена")
	})
}

// init - инициализация приложения
func init() {
	// Настройка формата логирования по умолчанию
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	// Настройка уровня логирования по умолчанию
	logrus.SetLevel(logrus.DebugLevel)
}
