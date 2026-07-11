// notify.go - Система уведомлений для Venera
//
// Этот модуль обеспечивает систему уведомлений пользователя
// о событиях в системе сбора идентификаторов.
//
// Основные функции:
// - Отправка уведомлений в системный трей
// - Управление алертами
// - Обработка критических ошибок
//
// Использование:
// import "venera/notify"
// notify.SendNotification("Внимание", "Достигнут порог очереди", "warning")

package notify

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Notification - уведомление
type Notification struct {
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Type        string    `json:"type"` // info, warning, error, success
	Duration    time.Duration `json:"duration"`
	AutoClose   bool      `json:"auto_close"`
	Timestamp   time.Time `json:"timestamp"`
	ID          string    `json:"id"`
}

type Notifier struct {
	mu            sync.Mutex
	notifications []Notification
	alerts        map[string]*Alert
	log           *logrus.Logger
}

// Global notifier - глобальный уведомитель
var (
	globalNotifier *Notifier
	once           sync.Once
)

// GetNotifier - получить глобальный уведомитель
func GetNotifier() *Notifier {
	once.Do(func() {
		globalNotifier = &Notifier{
			notifications: make([]Notification, 0),
			alerts:        make(map[string]*Alert),
			log:           logrus.New(),
		}
	})
	return globalNotifier
}

// SendNotification - отправить уведомление
func SendNotification(title, message, messageType string) {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	notification := Notification{
		Title:     title,
		Message:   message,
		Type:      messageType,
		Duration:  5 * time.Second,
		AutoClose: true,
		Timestamp: time.Now(),
		ID:        generateNotificationID(),
	}

	n.notifications = append(n.notifications, notification)

	// Логирование уведомления
	switch messageType {
	case "error":
		n.log.Errorf("[%s] %s", title, message)
	case "warning":
		n.log.Warnf("[%s] %s", title, message)
	case "success":
		n.log.Infof("[%s] %s", title, message)
	default:
		n.log.Infof("[%s] %s", title, message)
	}

	// В будущем можно добавить отправку в системный трей
	// tray.SendNotification(title, message)
}

// SendCriticalNotification - отправить критическое уведомление
func SendCriticalNotification(title, message string) {
	SendNotification(title, message, "error")
}

// SendWarningNotification - отправить предупреждение
func SendWarningNotification(title, message string) {
	SendNotification(title, message, "warning")
}

// SendInfoNotification - отправить информационное уведомление
func SendInfoNotification(title, message string) {
	SendNotification(title, message, "info")
}

// SendSuccessNotification - отправить уведомление об успехе
func SendSuccessNotification(title, message string) {
	SendNotification(title, message, "success")
}

// AddAlert - добавить алерт
func AddAlert(alert *Alert) error {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	if alert.ID == "" {
		alert.ID = generateAlertID()
	}

	n.alerts[alert.ID] = alert
	return nil
}

// RemoveAlert - удалить алерт
func RemoveAlert(id string) error {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.alerts, id)
	return nil
}

// GetAlert - получить алерт по ID
func GetAlert(id string) (*Alert, error) {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	alert, exists := n.alerts[id]
	if !exists {
		return nil, nil
	}
	return alert, nil
}

// GetAllAlerts - получить все алерты
func GetAllAlerts() []*Alert {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	alerts := make([]*Alert, 0, len(n.alerts))
	for _, alert := range n.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// TriggerAlert - сработать алерт
func TriggerAlert(id string) error {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	alert, exists := n.alerts[id]
	if !exists {
		return nil
	}

	alert.LastTrigger = time.Now()

	// Выполнение действия алерта
	switch alert.Action {
	case "notification":
		SendNotification(alert.Name, alert.Description, alert.Severity)
	case "log":
		switch alert.Severity {
		case "critical":
			n.log.Errorf("ALERT [%s]: %s", alert.Name, alert.Description)
		case "high":
			n.log.Errorf("ALERT [%s]: %s", alert.Name, alert.Description)
		case "medium":
			n.log.Warnf("ALERT [%s]: %s", alert.Name, alert.Description)
		case "low":
			n.log.Infof("ALERT [%s]: %s", alert.Name, alert.Description)
		}
	}

	return nil
}

// CheckAlerts - проверить все алерты
func CheckAlerts(conditionFunc func(*Alert) bool) []*Alert {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	var triggered []*Alert
	for _, alert := range n.alerts {
		if conditionFunc(alert) {
			triggered = append(triggered, alert)
		}
	}
	return triggered
}

// generateNotificationID - генерация ID уведомления
func generateNotificationID() string {
	return "notif_" + time.Now().Format("20060102_150405_000")
}

// ClearNotifications - очистить уведомления
func ClearNotifications() {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	n.notifications = make([]Notification, 0)
}

// GetNotifications - получить уведомления
func GetNotifications() []Notification {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.notifications
}

// GetRecentNotifications - получить последние уведомления
func GetRecentNotifications(count int) []Notification {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	if count <= 0 || count > len(n.notifications) {
		count = len(n.notifications)
	}

	start := len(n.notifications) - count
	if start < 0 {
		start = 0
	}

	return n.notifications[start:]
}

// NotificationCount - количество уведомлений
func NotificationCount() int {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	return len(n.notifications)
}

// AlertCount - количество алертов
func AlertCount() int {
	n := GetNotifier()
	n.mu.Lock()
	defer n.mu.Unlock()

	return len(n.alerts)
}
