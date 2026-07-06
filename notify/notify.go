// notify.go - Уведомления для системы Venera
//
// Этот модуль обеспечивает систему уведомлений для приложения Venera,
// включая уведомления в системном трее, всплывающие окна и
// отправку уведомлений в Windows Event Log.
//
// Основные функции:
// - Отправка уведомлений в системный трей
// - Всплывающие уведомления (popups)
// - Отправка уведомлений в Windows Event Log
// - Создание алертов через код
// - Управление уведомлениями
//
// Использование:
// import "venera/notify"
// notify.SendNotification("Информация", "Данные успешно сохранены")
// notify.SendWarning("Предупреждение", "Дисковое пространство заполняется")
// notify.SendError("Ошибка", "Не удалось подключиться к базе данных")

package notify

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

// NotificationType - тип уведомления
type NotificationType int

const (
	// Info - информационное уведомление
	Info NotificationType = iota
	// Warning - предупреждение
	Warning
	// Error - ошибка
	Error
	// Success - успешное завершение
	Success
)

// Notification - структура уведомления
type Notification struct {
	Title   string
	Message string
	Type    NotificationType
	Timeout time.Duration
}

// Notifier - структура уведомлений
type Notifier struct {
	logger  *logrus.Logger
	tray    TrayInterface
	enabled bool
}

// TrayInterface - интерфейс для работы с системным трей
type TrayInterface interface {
	Notify(title, message string, typ TrayNotificationType) error
}

// TrayNotificationType - тип уведомления в трее
type TrayNotificationType int

const (
	// TrayInfo - информационное уведомление в трее
	TrayInfo TrayNotificationType = iota
	// TrayWarning - предупреждение в трее
	TrayWarning
	// TrayError - ошибка в трее
	TrayError
)

// NewNotifier - создание нового нотификатора
func NewNotifier(logger *logrus.Logger, tray TrayInterface) *Notifier {
	return &Notifier{
		logger:  logger,
		tray:    tray,
		enabled: true,
	}
}

// Enable - включение уведомлений
func (n *Notifier) Enable() {
	n.enabled = true
}

// Disable - отключение уведотлений
func (n *Notifier) Disable() {
	n.enabled = false
}

// IsEnabled - проверка включены ли уведомления
func (n *Notifier) IsEnabled() bool {
	return n.enabled
}

// SendNotification - отправка уведомления
func (n *Notifier) SendNotification(title, message string) error {
	return n.SendNotificationWithTimeout(title, message, 5*time.Second)
}

// SendNotificationWithTimeout - отправка уведомления с таймаутом
func (n *Notifier) SendNotificationWithTimeout(title, message string, timeout time.Duration) error {
	if !n.enabled {
		return nil
	}

	n.logger.WithFields(logrus.Fields{
		"notification_title":   title,
		"notification_message": message,
	}).Debug("Отправка уведомления")

	// Отправка уведомления в системный трей
	if n.tray != nil {
		_ = n.tray.Notify(title, message, TrayInfo)
	}

	// Отправка уведомления в Windows Event Log
	n.sendToEventLog(title, message, windows.EVENTLOG_INFORMATION_TYPE)

	return nil
}

// SendWarning - отправка предупреждения
func (n *Notifier) SendWarning(title, message string) error {
	return n.SendWarningWithTimeout(title, message, 5*time.Second)
}

// SendWarningWithTimeout - отправка предупреждения с таймаутом
func (n *Notifier) SendWarningWithTimeout(title, message string, timeout time.Duration) error {
	if !n.enabled {
		return nil
	}

	n.logger.WithFields(logrus.Fields{
		"notification_title":   title,
		"notification_message": message,
	}).Warn("Отправка предупреждения")

	// Отправка уведомления в системный трей
	if n.tray != nil {
		_ = n.tray.Notify(title, message, TrayWarning)
	}

	// Отправка уведомления в Windows Event Log
	n.sendToEventLog(title, message, windows.EVENTLOG_WARNING_TYPE)

	return nil
}

// SendError - отправка ошибки
func (n *Notifier) SendError(title, message string) error {
	return n.SendErrorWithTimeout(title, message, 5*time.Second)
}

// SendErrorWithTimeout - отправка ошибки с таймаутом
func (n *Notifier) SendErrorWithTimeout(title, message string, timeout time.Duration) error {
	if !n.enabled {
		return nil
	}

	n.logger.WithFields(logrus.Fields{
		"notification_title":   title,
		"notification_message": message,
	}).Error("Отправка ошибки")

	// Отправка уведомления в системный трей
	if n.tray != nil {
		_ = n.tray.Notify(title, message, TrayError)
	}

	// Отправка уведомления в Windows Event Log
	n.sendToEventLog(title, message, windows.EVENTLOG_ERROR_TYPE)

	return nil
}

// SendSuccess - отправка успешного завершения
func (n *Notifier) SendSuccess(title, message string) error {
	return n.SendSuccessWithTimeout(title, message, 5*time.Second)
}

// SendSuccessWithTimeout - отправка успешного завершения с таймаутом
func (n *Notifier) SendSuccessWithTimeout(title, message string, timeout time.Duration) error {
	if !n.enabled {
		return nil
	}

	n.logger.WithFields(logrus.Fields{
		"notification_title":   title,
		"notification_message": message,
	}).Info("Отправка успешного завершения")

	// Отправка уведомления в системный трей
	if n.tray != nil {
		_ = n.tray.Notify(title, message, TrayInfo)
	}

	// Отправка уведомления в Windows Event Log
	n.sendToEventLog(title, message, windows.EVENTLOG_INFORMATION_TYPE)

	return nil
}

// sendToEventLog - отправка уведомления в Windows Event Log
func (n *Notifier) sendToEventLog(title, message string, eventType uint16) {
	// Открытие Event Log
	h, err := windows.RegisterEventSource(nil, windows.StringToUTF16Ptr("Venera"))
	if err != nil {
		n.logger.Errorf("Ошибка открытия Event Log: %v", err)
		return
	}
	defer windows.DeregisterEventSource(h)

	// Формирование сообщения
	msg := fmt.Sprintf("%s: %s", title, message)
	msgUTF16 := windows.StringToUTF16Ptr(msg)

	// Отправка уведомления
	err = windows.ReportEvent(h, eventType, 0, 1001, nil, 1, 0, &msgUTF16, nil)
	if err != nil {
		n.logger.Errorf("Ошибка отправки в Event Log: %v", err)
	}
}

// ShowNotification - показ всплывающего уведомления (для Windows)
func ShowNotification(title, message string) error {
	return ShowNotificationWithTimeout(title, message, 5*time.Second)
}

// ShowNotificationWithTimeout - показ всплывающего уведомления с таймаутом
func ShowNotificationWithTimeout(title, message string, timeout time.Duration) error {
	// Использование MessageBox для всплывающего уведомления
	MessageBoxPtr(nil, windows.StringToUTF16Ptr(message), windows.StringToUTF16Ptr(title),
		windows.MB_OK|windows.MB_ICONINFORMATION|windows.MB_TOPMOST)
	return nil
}

// ShowWarningNotification - показ всплывающего предупреждения
func ShowWarningNotification(title, message string) error {
	MessageBoxPtr(nil, windows.StringToUTF16Ptr(message), windows.StringToUTF16Ptr(title),
		windows.MB_OK|windows.MB_ICONWARNING|windows.MB_TOPMOST)
	return nil
}

// ShowErrorNotification - показ всплывающего сообщения об ошибке
func ShowErrorNotification(title, message string) error {
	MessageBoxPtr(nil, windows.StringToUTF16Ptr(message), windows.StringToUTF16Ptr(title),
		windows.MB_OK|windows.MB_ICONERROR|windows.MB_TOPMOST)
	return nil
}

// MessageBoxPtr - вызов MessageBox с указателями
func MessageBoxPtr(hWnd uintptr, lpText, lpCaption *uint16, uType uint32) int {
	// Получение адреса функции MessageBoxW
	user32 := syscall.NewLazyDLL("user32.dll")
	MessageBoxW := user32.NewProc("MessageBoxW")

	// Вызов функции
	r1, _, _ := MessageBoxW.Call(
		hWnd,
		uintptr(unsafe.Pointer(lpText)),
		uintptr(unsafe.Pointer(lpCaption)),
		uintptr(uType),
	)

	return int(r1)
}

// CreateAlertFromCode - создание алерта через код
func CreateAlertFromCode(alertID, name, condition, action string) *Alert {
	return &Alert{
		ID:        alertID,
		Name:      name,
		Condition: condition,
		Action:    action,
		Enabled:   true,
		Metadata:  make(map[string]string),
	}
}

// SendAlertNotification - отправка уведомления о срабатывании алерта
func (n *Notifier) SendAlertNotification(alertID, alertName string) error {
	if !n.enabled {
		return nil
	}

	message := fmt.Sprintf("Срабатывание алерта: %s (%s)", alertName, alertID)
	n.logger.WithFields(logrus.Fields{
		"alert_id":   alertID,
		"alert_name": alertName,
	}).Warn("Срабатывание алерта")

	// Отправка уведомления в системный трей
	if n.tray != nil {
		_ = n.tray.Notify("Срабатывание алерта", message, TrayWarning)
	}

	// Отправка уведомления в Windows Event Log
	n.sendToEventLog("Срабатывание алерта", message, windows.EVENTLOG_WARNING_TYPE)

	return nil
}

// SendDiskWarning - отправка предупреждения о диске
func (n *Notifier) SendDiskWarning(diskPath string, usedPercent float64) error {
	message := fmt.Sprintf("Дисковое пространство на %s заполнилось на %.1f%%", diskPath, usedPercent)
	return n.SendWarning("Предупреждение о диске", message)
}

// SendRAMWarning - отправка предупреждения о RAM
func (n *Notifier) SendRAMWarning(usedPercent float64) error {
	message := fmt.Sprintf("Использование RAM достигло %.1f%%", usedPercent)
	return n.SendWarning("Предупреждение о памяти", message)
}

// SendConnectionWarning - отправка предупреждения о подключении
func (n *Notifier) SendConnectionWarning(serviceName string) error {
	message := fmt.Sprintf("Не удалось подключиться к %s", serviceName)
	return n.SendWarning("Проблема с подключением", message)
}

// SendProcessWarning - отправка предупреждения о процессе
func (n *Notifier) SendProcessWarning(processID, processName string) error {
	message := fmt.Sprintf("Проблема с процессом: %s (%s)", processName, processID)
	return n.SendWarning("Проблема с процессом", message)
}

// SendProcessError - отправка ошибки о процессе
func (n *Notifier) SendProcessError(processID, processName string) error {
	message := fmt.Sprintf("Ошибка процесса: %s (%s)", processName, processID)
	return n.SendError("Ошибка процесса", message)
}

// SendServiceNotification - отправка уведомления о службе
func (n *Notifier) SendServiceNotification(serviceName, status string) error {
	message := fmt.Sprintf("Служба %s: %s", serviceName, status)
	return n.SendNotification("Статус службы", message)
}
