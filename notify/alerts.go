// alerts.go - Работа с алертами (формат CEL)
//
// Этот модуль обеспечивает загрузку, обработку и управление алертами
// в формате CEL (Common Event Language) для системы Venera.
//
// Основные функции:
// - Загрузка алертов из файла generic.alr
// - Обработка алертов по правилам CEL
// - Создание алертов через код из файла
// - Логирование алертов в централизованный лог
//
// Формат файла generic.alr:
// alert_id=001
// name=Алерт 1
// condition=...
// action=...
//
// Использование:
// import "venera/notify"
// alerts, err := LoadAlertsFile("settings/generic.alr")
// if err != nil { log.Fatal(err) }
// ProcessAlerts(alerts)

package notify

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"venera/config"
)

// Alert - структура алерта
type Alert struct {
	ID        string            `toml:"alert_id"`
	Name      string            `toml:"name"`
	Condition string            `toml:"condition"`
	Action    string            `toml:"action"`
	Enabled   bool              `toml:"enabled"`
	Metadata  map[string]string `toml:"metadata"`
}

// Alerts - коллекция алертов
type Alerts struct {
	Alerts map[string]*Alert `toml:"alerts"`
}

// NewAlerts - создание новой коллекции алертов
func NewAlerts() *Alerts {
	return &Alerts{
		Alerts: make(map[string]*Alert),
	}
}

// AddAlert - добавление алерта
func (a *Alerts) AddAlert(alert *Alert) {
	if a.Alerts == nil {
		a.Alerts = make(map[string]*Alert)
	}
	a.Alerts[alert.ID] = alert
}

// GetAlert - получение алерта по ID
func (a *Alerts) GetAlert(id string) *Alert {
	return a.Alerts[id]
}

// GetAllAlerts - получение всех алертов
func (a *Alerts) GetAllAlerts() map[string]*Alert {
	return a.Alerts
}

// EnableAlert - включение алерта
func (a *Alerts) EnableAlert(id string) error {
	if alert, ok := a.Alerts[id]; ok {
		alert.Enabled = true
		return nil
	}
	return fmt.Errorf("алерт с ID %s не найден", id)
}

// DisableAlert - отключение алерта
func (a *Alerts) DisableAlert(id string) error {
	if alert, ok := a.Alerts[id]; ok {
		alert.Enabled = false
		return nil
	}
	return fmt.Errorf("алерт с ID %s не найден", id)
}

// CheckAlert - проверка условия алерта
func (a *Alerts) CheckAlert(id string, data map[string]string) (bool, error) {
	alert, ok := a.Alerts[id]
	if !ok {
		return false, fmt.Errorf("алерт с ID %s не найден", id)
	}

	if !alert.Enabled {
		return false, nil
	}

	// Простая проверка условия (можно расширить при необходимости)
	condition := alert.Condition
	if condition == "" {
		return true, nil
	}

	// Проверка наличия ключа в данных
	parts := strings.SplitN(condition, "=", 2)
	if len(parts) == 2 {
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if dataValue, ok := data[key]; ok && dataValue == value {
			return true, nil
		}
	}

	return false, nil
}

// TriggerAlert - срабатывание алерта
func (a *Alerts) TriggerAlert(id string, logger *logrus.Logger) error {
	alert, ok := a.Alerts[id]
	if !ok {
		return fmt.Errorf("алерт с ID %s не найден", id)
	}

	if !alert.Enabled {
		return fmt.Errorf("алерт с ID %s отключен", id)
	}

	// Логирование срабатывания алерта
	logger.WithFields(logrus.Fields{
		"alert_id":   id,
		"alert_name": alert.Name,
		"action":     alert.Action,
	}).Warn("Срабатывание алерта")

	// Выполнение действия алерта (можно расширить при необходимости)
	if alert.Action != "" {
		// Здесь можно добавить выполнение действий (например, отправка уведомлений)
	}

	return nil
}

// LoadAlertsFile - загрузка алертов из файла
func LoadAlertsFile(filePath string) (*Alerts, error) {
	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return NewAlerts(), nil
	}

	// Открытие файла
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла алертов: %w", err)
	}
	defer file.Close()

	// Создание коллекции алертов
	alerts := NewAlerts()
	currentAlert := &Alert{}

	// Чтение файла построчно
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пропуск пустых строк
		if line == "" {
			continue
		}

		// Проверка на начало нового алерта
		if strings.HasPrefix(line, "alert_id") {
			// Сохраняем предыдущий алерт, если он есть
			if currentAlert.ID != "" {
				alerts.AddAlert(currentAlert)
			}
			// Создаем новый алерт
			currentAlert = &Alert{
				Metadata: make(map[string]string),
			}
			// Извлекаем ID
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				currentAlert.ID = strings.TrimSpace(parts[1])
			}
		} else if currentAlert.ID != "" && strings.Contains(line, "=") {
			// Параметр алерта
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "name":
					currentAlert.Name = value
				case "condition":
					currentAlert.Condition = value
				case "action":
					currentAlert.Action = value
				case "enabled":
					currentAlert.Enabled = strings.ToLower(value) == "true"
				default:
					currentAlert.Metadata[key] = value
				}
			}
		}
	}

	// Сохраняем последний алерт
	if currentAlert.ID != "" {
		alerts.AddAlert(currentAlert)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения файла алертов: %w", err)
	}

	return alerts, nil
}

// SaveAlertsFile - сохранение алертов в файл
func (a *Alerts) SaveAlertsFile(filePath string) error {
	// Создание файла
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла алертов: %w", err)
	}
	defer file.Close()

	// Запись алертов
	for id, alert := range a.Alerts {
		_, err := fmt.Fprintf(file, "alert_id=%s\n", id)
		if err != nil {
			return fmt.Errorf("ошибка записи алерта: %w", err)
		}

		fmt.Fprintf(file, "name=%s\n", alert.Name)
		fmt.Fprintf(file, "condition=%s\n", alert.Condition)
		fmt.Fprintf(file, "action=%s\n", alert.Action)
		fmt.Fprintf(file, "enabled=%t\n", alert.Enabled)

		for key, value := range alert.Metadata {
			fmt.Fprintf(file, "%s=%s\n", key, value)
		}
		fmt.Fprintln(file)
	}

	return nil
}

// GenerateAlertCode - генерация кода для создания алерта
func GenerateAlertCode(alertID, name, condition, action string) string {
	return fmt.Sprintf(`// Alert %s: %s
// Condition: %s
// Action: %s
alert := &notify.Alert{
    ID:        "%s",
    Name:      "%s",
    Condition: "%s",
    Action:    "%s",
    Enabled:   true,
}
process.Alerts.AddAlert(alert)`, alertID, name, condition, action, alertID, name, condition, action)
}

// CreateAlertFromConfig - создание алерта из конфигурации
func CreateAlertFromConfig(cfg *config.Config, logger *logrus.Logger) (*Alerts, error) {
	alerts := NewAlerts()

	// Загрузка алертов из файла
	alertsFile := cfg.GetAlertsFilePath()
	if _, err := os.Stat(alertsFile); err == nil {
		loadedAlerts, err := LoadAlertsFile(alertsFile)
		if err != nil {
			logger.Errorf("Ошибка загрузки алертов из %s: %v", alertsFile, err)
			return nil, err
		}
		for id, alert := range loadedAlerts.Alerts {
			alerts.Alerts[id] = alert
		}
	}

	return alerts, nil
}
