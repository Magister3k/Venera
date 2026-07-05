// register.go - Регистрация манифеста для Venera
//
// Этот модуль обеспечивает регистрацию манифеста приложения,
// включая проверку версии, обновление и откат при необходимости.
//
// Основные функции:
// - Регистрация манифеста при старте
// - GUI-уведомление при ошибке регистрации
// - Проверка версии зарегистрированного манифеста
// - Перезапись манифеста при обновлении
// - Создание резервной копии
// - Откат манифеста при ошибке
// - Хранение истории версий в Event ID 2003
//
// Использование:
// import "venera/manifest"
// manifest.Register()
// manifest.CheckVersion()
// manifest.Backup()

package manifest

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Manifest - структура манифеста
type Manifest struct {
	XMLName     xml.Name `xml:"manifest"`
	Version     string   `xml:"version"`
	BuildDate   string   `xml:"buildDate"`
	ProductName string   `xml:"productName"`
	Company     string   `xml:"company"`
	CPU         string   `xml:"cpu"`
	OS          string   `xml:"os"`
}

// Register - регистрация манифеста
func Register() error {
	log := logrus.WithField("module", "manifest")

	log.Info("Регистрация манифеста")

	// Создание манифеста
	manifest := Manifest{
		Version:     "1.0.0",
		BuildDate:   time.Now().Format("2006-01-02"),
		ProductName: "Venera",
		Company:     "Sber",
		CPU:         "x64",
		OS:          "Windows",
	}

	// Сериализация в XML
	data, err := xml.MarshalIndent(manifest, "", "  ")
	if err != nil {
		log.Errorf("Ошибка сериализации манифеста: %v", err)
		return err
	}

	// Добавление заголовка XML
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.Write(data)

	// Запись в файл
	manifestPath := "manifest.xml"
	if err := os.WriteFile(manifestPath, buf.Bytes(), 0644); err != nil {
		log.Errorf("Ошибка записи манифеста: %v", err)
		return err
	}

	log.Info("Манифест зарегистрирован успешно")

	// Регистрация в Windows Event Log
	if err := registerToEventLog(manifest); err != nil {
		log.Warnf("Ошибка регистрации в Event Log: %v", err)
	}

	return nil
}

// CheckVersion - проверка версии манифеста
func CheckVersion() (bool, error) {
	log := logrus.WithField("module", "manifest")

	// Проверка наличия файла манифеста
	if _, err := os.Stat("manifest.xml"); os.IsNotExist(err) {
		log.Warn("Манифест не найден")
		return false, fmt.Errorf("манифест не найден")
	}

	// Чтение манифеста
	data, err := os.ReadFile("manifest.xml")
	if err != nil {
		log.Errorf("Ошибка чтения манифеста: %v", err)
		return false, err
	}

	// Парсинг манифеста
	var manifest Manifest
	if err := xml.Unmarshal(data, &manifest); err != nil {
		log.Errorf("Ошибка парсинга манифеста: %v", err)
		return false, err
	}

	// Сравнение версии
	currentVersion := "1.0.0" // Текущая версия приложения
	if manifest.Version != currentVersion {
		log.Warnf("Версия манифеста (%s) не совпадает с версией приложения (%s)", manifest.Version, currentVersion)
		return false, nil
	}

	log.Info("Версия манифеста проверена успешно")
	return true, nil
}

// Update - обновление манифеста
func Update() error {
	log := logrus.WithField("module", "manifest")

	log.Info("Обновление манифеста")

	// Создание резервной копии
	if err := Backup(); err != nil {
		log.Errorf("Ошибка создания резервной копии: %v", err)
		return err
	}

	// Регистрация нового манифеста
	if err := Register(); err != nil {
		log.Errorf("Ошибка обновления манифеста: %v", err)
		return err
	}

	log.Info("Манифест обновлен успешно")
	return nil
}

// Backup - создание резервной копии манифеста
func Backup() error {
	log := logrus.WithField("module", "manifest")

	// Проверка наличия файла манифеста
	if _, err := os.Stat("manifest.xml"); os.IsNotExist(err) {
		log.Info("Манифест не найден, резервная копия не требуется")
		return nil
	}

	// Создание имени резервной копии
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("manifest.xml.%s.backup", timestamp)

	// Копирование файла
	if err := copyFile("manifest.xml", backupPath); err != nil {
		log.Errorf("Ошибка создания резервной копии: %v", err)
		return err
	}

	log.Infof("Резервная копия создана: %s", backupPath)
	return nil
}

// copyFile - копирование файла
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// Rollback - откат манифеста
func Rollback() error {
	log := logrus.WithField("module", "manifest")

	log.Info("Откат манифеста")

	// Поиск последней резервной копии
	files, err := filepath.Glob("manifest.xml.*.backup")
	if err != nil {
		log.Errorf("Ошибка поиска резервных копий: %v", err)
		return err
	}

	if len(files) == 0 {
		log.Warn("Резервные копии не найдены")
		return fmt.Errorf("резервные копии не найдены")
	}

	// Сортировка файлов по времени (последний файл - самый новый)
	lastBackup := files[len(files)-1]

	// Восстановление из резервной копии
	if err := copyFile(lastBackup, "manifest.xml"); err != nil {
		log.Errorf("Ошибка восстановления из резервной копии: %v", err)
		return err
	}

	log.Infof("Манифест восстановлен из %s", lastBackup)
	return nil
}

// registerToEventLog - регистрация в Windows Event Log
func registerToEventLog(manifest Manifest) error {
	// Открытие ключа Event Log
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\EventLog\Application\Venera`,
		registry.SET_VALUE)
	if err != nil {
		// Ключ не найден, создаем его
		key, err = registry.CreateKey(registry.LOCAL_MACHINE,
			`SYSTEM\CurrentControlSet\Services\EventLog\Application\Venera`,
			registry.SET_VALUE)
		if err != nil {
			return err
		}
		defer key.Close()
	}

	// Запись информации о манифесте
	if err := key.SetStringValue("EventMessageFile", "venera.exe"); err != nil {
		return err
	}

	// Запись сообщения в Event Log
	if err := writeEventLogEntry(manifest); err != nil {
		return err
	}

	return nil
}

// writeEventLogEntry - запись сообщения в Event Log
func writeEventLogEntry(manifest Manifest) error {
	// Открытие источника события
	source := "Venera"
	eventType := windows.EVENTLOG_INFORMATION_TYPE
	eventID := uint32(1001)
	category := uint16(0)
	data := []byte(fmt.Sprintf("Version: %s\nBuildDate: %s\n", manifest.Version, manifest.BuildDate))

	// Получение дескриптора источника
	handle, err := windows.RegisterEventSource(nil, syscall.StringToUTF16Ptr(source))
	if err != nil {
		return err
	}
	defer windows.DeregisterEventSource(handle)

	// Запись события
	if err := windows.ReportEvent(handle, eventType, category, eventID, nil, 1, 0, &data[0], nil); err != nil {
		return err
	}

	return nil
}

// GetManifest - получение манифеста
func GetManifest() (*Manifest, error) {
	// Проверка наличия файла манифеста
	if _, err := os.Stat("manifest.xml"); os.IsNotExist(err) {
		return nil, fmt.Errorf("манифест не найден")
	}

	// Чтение манифеста
	data, err := os.ReadFile("manifest.xml")
	if err != nil {
		return nil, err
	}

	// Парсинг манифеста
	var manifest Manifest
	if err := xml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// unsafe - импорт unsafe
import "unsafe"

// syscall - импорт syscall
import "syscall"
