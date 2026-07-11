// alerts.go - Управление алертами для Venera
//
// Этот модуль обеспечивает загрузку и управление алертами
// из файла generic.alr, а также создание алертов программно.
//
// Основные функции:
// - Загрузка алертов из файла
// - Создание алертов через код
// - Проверка условий алертов
// - Срабатывание алертов
//
// Использование:
// import "venera/notify"
// alerts := LoadAlertsFile("settings/generic.alr")

package notify

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

"// Alert - структура алерта\ntype Alert struct {\n\tID          string                 `toml:\"id\"`\n\tName        string                 `toml:\"name\"`\n\tDescription string                 `toml:\"description\"`\n\tCondition   string                 `toml:\"condition\"`\n\tAction      string                 `toml:\"action\"`\n\tSeverity    string                 `toml:\"severity\"`\n\tEnabled     bool                   `toml:\"enabled\"`\n\tLastTrigger time.Time              `toml:\"last_trigger\"`\n\tMetadata    map[string]interface{} `toml:\"metadata\"`\n}\n\n// AlertsManager - менеджер алертов\ntype AlertsManager struct {\n\talerts   map[string]*Alert\n\tlogger   *logrus.Logger\n\tnotifier *Notifier\n}"}
	alerts   map[string]*Alert
	logger   *logrus.Logger
	notifier *Notifier
}

// NewAlertsManager - создание нового менеджера алертов
func NewAlertsManager(logger *logrus.Logger) *AlertsManager {
	return &AlertsManager{
		alerts:   make(map[string]*Alert),
		logger:   logger,
		notifier: GetNotifier(),
	}
}

// LoadAlertsFile - загрузка алертов из файла
func (am *AlertsManager) LoadAlertsFile(filePath string) error {
	// Открытие файла
	file, err := os.Open(filePath)
	if err != nil {
		// Файл не найден - это нормально
		return nil
	}
	defer file.Close()

	// Карта для хранения алертов
	alerts := make(map[string]*Alert)
	currentAlertID := ""
	currentAlert := &Alert{}

	// Чтение файла построчно
	buf := make([]byte, 65536)
	for {
		n, err := file.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("ошибка чтения файла алертов: %w", err)
		}

		if n == 0 {
			break
		}

		lines := strings.Split(string(buf[:n]), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Проверка на начало нового алерта
			if strings.HasPrefix(line, "alert_id") {
				// Сохранить предыдущий алерт, если есть
				if currentAlertID != "" && currentAlert.Name != "" {
					alerts[currentAlertID] = currentAlert
				}

				// Создать новый алерт
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					currentAlertID = strings.TrimSpace(parts[1])
					currentAlert = &Alert{
						ID:        currentAlertID,
						Enabled:   true,
						Metadata:  make(map[string]interface{}),
					}
				}
			} else if currentAlertID != "" && strings.Contains(line, "=") {
				// Параметр алерта
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])

					switch key {
					case "name":
						currentAlert.Name = value
					case "description":
						currentAlert.Description = value
					case "condition":
						currentAlert.Condition = value
					case "action":
						currentAlert.Action = value
					case "severity":
						currentAlert.Severity = value
					case "enabled":
						currentAlert.Enabled = strings.ToLower(value) == "true"
					default:
						currentAlert.Metadata[key] = value
					}
				}
			}
		}
	}

	// Сохранить последний алерт
	if currentAlertID != "" && currentAlert.Name != "" {
		alerts[currentAlertID] = currentAlert
	}

	// Сохранить алерты
	am.alerts = alerts
	am.logger.Infof("Загружено %d алертов", len(alerts))
	return nil
}

// AddAlert - добавить алерт
func (am *AlertsManager) AddAlert(alert *Alert) error {
	if alert.ID == "" {
		alert.ID = generateAlertID()
	}

	am.alerts[alert.ID] = alert
	am.logger.WithFields(logrus.Fields{
		"alert_id":   alert.ID,
		"alert_name": alert.Name,
	}).Info("Алерт добавлен")
	return nil
}

// RemoveAlert - удалить алерт
func (am *AlertsManager) RemoveAlert(id string) error {
	if _, exists := am.alerts[id]; !exists {
		return fmt.Errorf("алерт с ID %s не найден", id)
	}

	delete(am.alerts, id)
	am.logger.WithFields(logrus.Fields{
		"alert_id": id,
	}).Info("Алерт удален")
	return nil
}

// GetAlert - получить алерт по ID
func (am *AlertsManager) GetAlert(id string) (*Alert, error) {
	alert, exists := am.alerts[id]
	if !exists {
		return nil, fmt.Errorf("алерт с ID %s не найден", id)
	}
	return alert, nil
}

// GetAllAlerts - получить все алерты
func (am *AlertsManager) GetAllAlerts() []*Alert {
	alerts := make([]*Alert, 0, len(am.alerts))
	for _, alert := range am.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// EnableAlert - включить алерт
func (am *AlertsManager) EnableAlert(id string) error {
	alert, err := am.GetAlert(id)
	if err != nil {
		return err
	}

	alert.Enabled = true
	am.logger.WithFields(logrus.Fields{
		"alert_id": id,
	}).Info("Алерт включен")
	return nil
}

// DisableAlert - выключить алерт
func (am *AlertsManager) DisableAlert(id string) error {
	alert, err := am.GetAlert(id)
	if err != nil {
		return err
	}

	alert.Enabled = false
	am.logger.WithFields(logrus.Fields{
		"alert_id": id,
	}).Info("Алерт выключен")
	return nil
}

// TriggerAlert - сработать алерт
func (am *AlertsManager) TriggerAlert(id string) error {
	alert, err := am.GetAlert(id)
	if err != nil {
		return err
	}

	if !alert.Enabled {
		return nil
	}

	alert.LastTrigger = time.Now()

	// Выполнение действия алерта
	switch alert.Action {
	case "notification":
		am.notifier.SendNotification(alert.Name, alert.Description, alert.Severity)
	case "log":
		switch alert.Severity {
		case "critical":
			am.logger.Errorf("ALERT [%s]: %s", alert.Name, alert.Description)
		case "high":
			am.logger.Errorf("ALERT [%s]: %s", alert.Name, alert.Description)
		case "medium":
			am.logger.Warnf("ALERT [%s]: %s", alert.Name, alert.Description)
		case "low":
			am.logger.Infof("ALERT [%s]: %s", alert.Name, alert.Description)
		}
	}

	am.logger.WithFields(logrus.Fields{
		"alert_id":   id,
		"alert_name": alert.Name,
	}).Warn("Алерт сработал")
	return nil
}

// CheckCondition - проверить условие алерта
func (am *AlertsManager) CheckCondition(id string, conditionData map[string]interface{}) bool {
	alert, err := am.GetAlert(id)
	if err != nil {
		return false
	}

	if !alert.Enabled {
		return false
	}

	// TODO: Реализовать проверку условия
	// В текущей версии просто возвращаем false
	// В будущем можно добавить поддержку CEL или других выражений

	return false
}

// CheckAllAlerts - проверить все алерты
func (am *AlertsManager) CheckAllAlerts(conditionData map[string]interface{}) []*Alert {
	var triggered []*Alert

	for _, alert := range am.alerts {
		if alert.Enabled && am.checkConditionSimple(alert, conditionData) {
			am.TriggerAlert(alert.ID)
			triggered = append(triggered, alert)
		}
	}

	return triggered
}

// checkConditionSimple - простая проверка условия
func (am *AlertsManager) checkConditionSimple(alert *Alert, conditionData map[string]interface{}) bool {
	// TODO: Реализовать проверку условия
	// В текущей версии просто возвращаем false
	// В будущем можно добавить поддержку CEL или других выражений

	return false
}

// generateAlertID - генерация ID алерта
func generateAlertID() string {
	return "alert_" + time.Now().Format("20060102_150405_000")
}

// CreateAlertFromCode - создание алерта через код
func CreateAlertFromCode(alertID, name, condition, action, severity string) *Alert {
	return &Alert{
		ID:        alertID,
		Name:      name,
		Condition: condition,
		Action:    action,
		Severity:  severity,
		Enabled:   true,
		Metadata:  make(map[string]interface{}),
	}
}

// CreateDiskAlert - создание алерта для диска
func CreateDiskAlert(diskPath string, thresholdPercent int) *Alert {
	return &Alert{
		ID:          fmt.Sprintf("disk_%s", diskPath),
		Name:        fmt.Sprintf("Предупреждение о диске: %s", diskPath),
		Description: fmt.Sprintf("Дисковое пространство на %s з��полнилось более чем на %d%%", diskPath, thresholdPercent),
		Condition:   fmt.Sprintf("disk_usage_percent > %d", thresholdPercent),
		Action:      "notification",
		Severity:    "high",
		Enabled:     true,
		Metadata: map[string]interface{}{
			"disk_path":   diskPath,
			"threshold":   thresholdPercent,
			"check_type":  "disk_usage",
		},
	}
}

// CreateRAMAlert - создание алерта для RAM
func CreateRAMAlert(thresholdPercent int) *Alert {
	return &Alert{
		ID:          "ram_warning",
		Name:        "Предупреждение о памяти",
		Description: fmt.Sprintf("Использование RAM достигло более чем %d%%", thresholdPercent),
		Condition:   "ram_usage_percent > %d",
		Action:      "notification",
		Severity:    "high",
		Enabled:     true,
		Metadata: map[string]interface{}{
			"threshold":  thresholdPercent,
			"check_type": "ram_usage",
		},
	}
}

// CreateQueueAlert - создание алерта для очереди
func CreateQueueAlert(processID string, thresholdSize int) *Alert {
	return &Alert{
		ID:          fmt.Sprintf("queue_%s", processID),
		Name:        fmt.Sprintf("Очередь процесса переполнилась: %s", processID),
		Description: fmt.Sprintf("Размер очереди процесса %s достиг %d записей", processID, thresholdSize),
		Condition:   "queue_size > %d",
		Action:      "notification",
		Severity:    "medium",
		Enabled:     true,
		Metadata: map[string]interface{}{
			"process_id":  processID,
			"threshold":   thresholdSize,
			"check_type":  "queue_size",
		},
	}
}

// CreateErrorAlert - создание алерта для ошибок
func CreateErrorAlert(name, description, condition string) *Alert {
	return &Alert{
		ID:          fmt.Sprintf("error_%s", generateAlertID()),
		Name:        name,
		Description: description,
		Condition:   condition,
		Action:      "notification",
		Severity:    "critical",
		Enabled:     true,
		Metadata: map[string]interface{}{
			"check_type": "error",
		},
	}
}
