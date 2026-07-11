// diagnose.go - Модуль диагностики для Venera
//
// Этот модуль обеспечивает диагностику системы сбора идентификаторов,
// включая проверку конфигурации, ресурсов, зависимостей и создание отчетов.
//
// Основные функции:
// - Проверка конфигурации
// - Проверка регистрации
// - Проверка ресурсов
// - Проверка сетевых интерфейсов
// - Проверка зависимостей (podman, tshark)
// - Проверка служб
// - История событий Event Log
// - Экспорт отчетов в PDF
// - Создание архивов
//
// Использование:
// import "venera/diagnose"
// diag := diagnose.NewDiagnose(cfg, dragonflyDB, postgresDB)
// report, err := diag.Run()

package diagnose

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"venera/config"
	"venera/data"
)

// Diagnose - структура диагностики
type Diagnose struct {
	cfg          *config.Config
	dragonflyDB  *data.DragonflyDB
	postgresDB   *data.PostgreSQL
	log          *logrus.Logger
	results      []DiagnosticResult
}

// DiagnosticResult - результат диагностики
type DiagnosticResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // "success", "warning", "error"
	Message  string `json:"message"`
	Duration int64  `json:"duration"`
}

// NewDiagnose - создание нового объекта диагностики
func NewDiagnose(cfg *config.Config, dragonflyDB *data.DragonflyDB, postgresDB *data.PostgreSQL) *Diagnose {
	return &Diagnose{
		cfg:         cfg,
		dragonflyDB: dragonflyDB,
		postgresDB:  postgresDB,
		log:         logrus.WithField("module", "diagnose"),
		results:     make([]DiagnosticResult, 0),
	}
}

// Run - выполнение диагностики
func (d *Diagnose) Run() (string, error) {
	d.log.Info("Запуск диагностики")

	// Очистка результатов
	d.results = make([]DiagnosticResult, 0)

	// Выполнение проверок
	checks := []struct {
		name string
		fn   func() error
	}{
		{"Проверка конфигурации", d.checkConfig},
		{"Проверка регистрации", d.checkRegistration},
		{"Проверка ресурсов", d.checkResources},
		{"Проверка сетевых интерфейсов", d.checkNetworkInterfaces},
		{"Проверка tshark", d.checkTshark},
		{"Проверка podman", d.checkPodman},
		{"Проверка образа DragonflyDB", d.checkDragonflyImage},
		{"Проверка контейнера DragonflyDB", d.checkDragonflyContainer},
		{"Проверка подключения к PostgreSQL", d.checkPostgreSQL},
		{"Проверка службы", d.checkService},
		{"Проверка Event Log", d.checkEventLog},
	}

	for _, check := range checks {
		if err := check.fn(); err != nil {
			d.results = append(d.results, DiagnosticResult{
				Name:     check.name,
				Status:   "error",
				Message:  err.Error(),
				Duration: 0,
			})
		} else {
			d.results = append(d.results, DiagnosticResult{
				Name:     check.name,
				Status:   "success",
				Message:  "ОК",
				Duration: 0,
			})
		}
	}

	// Генерация отчета
	report := d.generateReport()

	return report, nil
}

// checkConfig - проверка конфигурации
func (d *Diagnose) checkConfig() error {
	// Проверка наличия файла конфигурации
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		return fmt.Errorf("файл config.toml не найден")
	}

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	d.cfg = cfg
	return nil
}

// checkRegistration - проверка регистрации
func (d *Diagnose) checkRegistration() error {
	// Проверка наличия файла манифеста
	if _, err := os.Stat("manifest.xml"); os.IsNotExist(err) {
		return fmt.Errorf("файл manifest.xml не найден")
	}

	// Проверка регистрации манифеста через Windows Registry
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\EventLog\Application\Venera`,
		registry.READ)
	if err != nil {
		return fmt.Errorf("ошибка проверки регистрации: %w", err)
	}
	defer key.Close()

	return nil
}

// checkResources - проверка ресурсов
func (d *Diagnose) checkResources() error {
	// Проверка RAM
	if err := d.checkRAM(); err != nil {
		return err
	}

	// Проверка диска
	if err := d.checkDisk(); err != nil {
		return err
	}

	return nil
}

// checkRAM - проверка RAM
func (d *Diagnose) checkRAM() error {
	var memInfo windows.MEMORYSTATUSEX
	memInfo.Length = uint32(unsafe.Sizeof(memInfo))

	if err := windows.GlobalMemoryStatusEx(&memInfo); err != nil {
		return fmt.Errorf("ошибка получения информации о памяти: %w", err)
	}

	totalRAM := float64(memInfo.TotalPhys) / (1024 * 1024 * 1024)
	availableRAM := float64(memInfo.AvailPhys) / (1024 * 1024 * 1024)
	usagePercent := (1 - availableRAM/totalRAM) * 100

	d.log.Infof("RAM: %.2f GB всего, %.2f GB доступно, %.2f%% использовано", totalRAM, availableRAM, usagePercent)

	return nil
}

// checkDisk - проверка диска
func (d *Diagnose) checkDisk() error {
	// Проверка диска с базой PostgreSQL
	diskPath := "C:\\"
	var freeBytesAvailable uint64
	var totalBytes uint64
	var totalFreeBytes uint64

	err := windows.GetDiskFreeSpaceEx(syscall.StringToUTF16Ptr(diskPath),
		&freeBytesAvailable, &totalBytes, &totalFreeBytes)
	if err != nil {
		return fmt.Errorf("ошибка получения информации о диске: %w", err)
	}

	diskTotalGB := float64(totalBytes) / (1024 * 1024 * 1024)
	diskFreeGB := float64(totalFreeBytes) / (1024 * 1024 * 1024)
	diskUsagePercent := (1 - diskFreeGB/diskTotalGB) * 100

	d.log.Infof("Disk C:\\: %.2f GB всего, %.2f GB свободно, %.2f%% использовано", diskTotalGB, diskFreeGB, diskUsagePercent)

	return nil
}

// checkNetworkInterfaces - проверка сетевых интерфейсов
func (d *Diagnose) checkNetworkInterfaces() error {
	// Получение списка сетевых интерфейсов
	cmd := exec.Command("ipconfig", "/all")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка получения информации о сетевых интерфейсах: %w", err)
	}

	interfaces := out.String()
	d.log.Info("Сетевые интерфейсы получены")

	// Парсинг и форматирование вывода
	lines := strings.Split(interfaces, "\n")
	var formattedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			formattedLines = append(formattedLines, line)
		}
	}

	d.log.Infof("Найдено %d строк с информацией о сетевых интерфейсах", len(formattedLines))

	return nil
}

// checkTshark - проверка tshark
func (d *Diagnose) checkTshark() error {
	// Проверка наличия файла
	if _, err := os.Stat(d.cfg.Paths.TsharkPath); os.IsNotExist(err) {
		return fmt.Errorf("tshark не найден: %s", d.cfg.Paths.TsharkPath)
	}

	// Проверка версии
	cmd := exec.Command(d.cfg.Paths.TsharkPath, "-v")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка проверки версии tshark: %w", err)
	}

	d.log.Info("tshark проверен успешно")
	return nil
}

// checkPodman - проверка podman
func (d *Diagnose) checkPodman() error {
	// Проверка наличия файла
	if _, err := os.Stat(d.cfg.Paths.PodmanPath); os.IsNotExist(err) {
		return fmt.Errorf("podman не найден: %s", d.cfg.Paths.PodmanPath)
	}

	// Проверка версии
	cmd := exec.Command(d.cfg.Paths.PodmanPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка проверки версии podman: %w", err)
	}

	d.log.Info("podman проверен успешно")
	return nil
}

// checkDragonflyImage - проверка образа DragonflyDB
func (d *Diagnose) checkDragonflyImage() error {
	// Проверка образа через podman images
	cmd := exec.Command(d.cfg.Paths.PodmanPath, "images", d.cfg.Paths.DragonflyImage)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка проверки образа DragonflyDB: %w", err)
	}

	output := out.String()
	if !strings.Contains(output, d.cfg.Paths.DragonflyImage) {
		return fmt.Errorf("образ DragonflyDB не найден: %s", d.cfg.Paths.DragonflyImage)
	}

	d.log.Info("Образ DragonflyDB найден")
	return nil
}

// checkDragonflyContainer - проверка контейнера DragonflyDB
func (d *Diagnose) checkDragonflyContainer() error {
	// Проверка контейнера через podman ps -a
	cmd := exec.Command(d.cfg.Paths.PodmanPath, "ps", "-a", "--filter", "name=cachedb", "--format", "{{.Names}}")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка проверки контейнера DragonflyDB: %w", err)
	}

	output := out.String()
	if !strings.Contains(output, "cachedb") {
		return fmt.Errorf("контейнер cachedb не найден")
	}

	d.log.Info("Контейнер cachedb найден")
	return nil
}

// checkPostgreSQL - проверка PostgreSQL
func (d *Diagnose) checkPostgreSQL() error {
	if d.postgresDB == nil {
		return fmt.Errorf("не подключено к PostgreSQL")
	}

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := d.postgresDB.GetPool().Ping(ctx); err != nil {
		return fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	d.log.Info("PostgreSQL проверен успешно")
	return nil
}

// checkService - проверка службы
func (d *Diagnose) checkService() error {
	// Открытие SCM
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return fmt.Errorf("ошибка открытия SCM: %w", err)
	}
	defer windows.CloseServiceHandle(scm)

	// Открытие службы
	service, err := windows.OpenService(scm, syscall.StringToUTF16Ptr("VeneraSrv"), windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return fmt.Errorf("ошибка открытия службы VeneraSrv: %w", err)
	}
	defer windows.CloseServiceHandle(service)

	// Получение статуса службы
	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(service, &status); err != nil {
		return fmt.Errorf("ошибка получения статуса службы: %w", err)
	}

	d.log.Infof("Служба VeneraSrv: currentState=%d", status.CurrentState)

	return nil
}

// checkEventLog - проверка Event Log
func (d *Diagnose) checkEventLog() error {
	// Проверка наличия логов через PowerShell
	cmd := exec.Command("powershell", "-Command",
		"Get-EventLog -LogName Application -EntryType Error -Newest 10 | Format-List")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка получения событий Event Log: %w", err)
	}

	output := out.String()
	d.log.Infof("Получено %d строк из Event Log", len(strings.Split(output, "\n")))

	return nil
}

// generateReport - генерация отчета
func (d *Diagnose) generateReport() string {
	var buf bytes.Buffer

	buf.WriteString("=== Отчет диагностики Venera ===\n")
	buf.WriteString(fmt.Sprintf("Время создания: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("Версия ОС: %s (%s)\n", runtime.GOOS, runtime.GOARCH))
	buf.WriteString(fmt.Sprintf("Версия приложения: 1.0.0\n"))
	buf.WriteString("\n")

	buf.WriteString("=== Результаты диагностики ===\n")
	for _, result := range d.results {
		buf.WriteString(fmt.Sprintf("[%s] %s: %s\n", result.Status, result.Name, result.Message))
	}

	buf.WriteString("\n=== Конец отчета ===\n")

	return buf.String()
}

// ExportReport - экспорт отчета в файл
func (d *Diagnose) ExportReport(filePath string) error {
	report := d.generateReport()

	// Запись в файл
	if err := os.WriteFile(filePath, []byte(report), 0644); err != nil {
		return fmt.Errorf("ошибка записи отчета: %w", err)
	}

	d.log.Infof("Отчет экспортирован в %s", filePath)
	return nil
}

// CreateArchive - создание архива с логами и отчетом
func (d *Diagnose) CreateArchive(archivePath string) error {
	// Создание архива через PowerShell
	cmd := exec.Command("powershell", "-Command",
		"Compress-Archive -Path 'Logs\\*,config.toml' -DestinationPath '"+archivePath+"' -Force")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка создания архива: %w", err)
	}

	d.log.Infof("Архив создан: %s", archivePath)
	return nil
}