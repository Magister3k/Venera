// alert.go - Модель данных для алертов Venera
//
// Этот файл определяет структуру Alert и вспомогательные функции
// для работы с алертами в системе Venera.
//
// Алерт - это правило, которое проверяет условия и выполняет действия
// при срабатывании. Алерты используются для мониторинга состояния системы,
// обнаружения проблем и уведомления администратора.
//
// Основные возможности:
// - Определение структуры алерта с метаданными
// - Управление состоянием алерта (включен/выключен)
// - Отслеживание последнего срабатывания
// - Поддержка различных уровней критичности
//
// Использование:
// alert := &Alert{
//     ID:          "disk_warning",
//     Name:        "Недостаточно места на диске",
//     Description: "Проверка свободного места на диске C:",
//     Condition:   "disk_usage_percent > 90",
//     Action:      "notification",
//     Severity:    "high",
//     Enabled:     true,
// }
//
// see: notify/alerts.go - управление алертами
// see: notify/notify.go - отправка уведомлений

package models

import (
	"time"
)

// Alert - модель данных алерта
//
// Алерт представляет собой правило мониторинга, которое проверяет
// определенные условия и выполняет действия при их выполнении.
// Каждый алерт имеет уникальный ID, описание, условие срабатывания
// и действие, которое выполняется при срабатывании.
//
// Поля:
//   - ID: Уникальный идентификатор алерта (например, "disk_c_warning")
//   - Name: Человекочитаемое имя алерта (например, "Недостаточно места на диске")
//   - Description: Подробное описание условия и причин срабатывания
//   - Condition: Условие срабатывания (например, "disk_usage_percent > 90")
//   - Action: Действие при срабатывании ("notification", "log", "script")
//   - Severity: Критичность алерта ("critical", "high", "medium", "low")
//   - Enabled: Флаг включенности алерта (true - алерт активен)
//   - LastTrigger: Время последнего срабатывания алерта
//   - Metadata: Дополнительные метаданные в виде пар ключ-значение
//
// Пример создания алерта:
//   alert := &Alert{
//       ID:          "ram_warning",
//       Name:        "Высокое использование RAM",
//       Description: "Использование оперативной памяти превысило порог",
//       Condition:   "ram_usage_percent > 85",
//       Action:      "notification",
//       Severity:    "high",
//       Enabled:     true,
//       Metadata: map[string]interface{}{
//           "threshold": 85,
//       },
//   }
type Alert struct {
	ID          string                 `json:"id"`          // Уникальный идентификатор алерта
	Name        string                 `json:"name"`        // Имя алерта для отображения
	Description string                 `json:"description"` // Описание условия срабатывания
	Condition   string                 `json:"condition"`   // Условие срабатывания (выражение)
	Action      string                 `json:"action"`      // Действие при срабатывании
	Severity    string                 `json:"severity"`    // Критичность алерта
	Enabled     bool                   `json:"enabled"`     // Флаг включенности алерта
	LastTrigger *time.Time             `json:"last_trigger"` // Время последнего срабатывания
	Metadata    map[string]interface{} `json:"metadata"`    // Дополнительные метаданные
}

// Copy - создание копии алерта
//
// Создает глубокую копию алерта, включая копию карты Metadata.
// Это необходимо для безопасного изменения алерта без влияния на оригинал.
func (a *Alert) Copy() *Alert {
	if a == nil {
		return nil
	}

	// Создаем новую карту для метаданных
	metadata := make(map[string]interface{}, len(a.Metadata))
	for k, v := range a.Metadata {
		metadata[k] = v
	}

	return &Alert{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Condition:   a.Condition,
		Action:      a.Action,
		Severity:    a.Severity,
		Enabled:     a.Enabled,
		LastTrigger: a.LastTrigger,
		Metadata:    metadata,
	}
}

// Clone - создание новой копии алерта с новым ID
//
// Создает копию алерта и генерирует для нее новый уникальный ID.
// Это полезно для клонирования шаблонов алертов.
func (a *Alert) Clone() *Alert {
	copy := a.Copy()
	if copy != nil {
		copy.ID = generateAlertID()
	}
	return copy
}

// IsTriggered - проверка, сработал ли алерт недавно
//
// Проверяет, сработал ли алерт в последнее время (в указанном диапазоне).
// Если LastTrigger равен nil, возвращает false.
func (a *Alert) IsTriggered(within time.Duration) bool {
	if a.LastTrigger == nil {
		return false
	}
	return time.Since(*a.LastTrigger) < within
}

// Trigger - пометить алерт как сработавший
//
// Устанавливает время последнего срабатывания текущим временем.
// Этот метод должен вызываться при фактическом срабатывании алерта.
func (a *Alert) Trigger() {
	now := time.Now()
	a.LastTrigger = &now
}

// Validate - валидация алерта
//
// Проверяет, что все обязательные поля алерта заполнены корректно.
// Возвращает nil, если алерт валиден, или ошибку с описанием проблемы.
func (a *Alert) Validate() error {
	if a.ID == "" {
		return &AlertValidationError{"ID не может быть пустым"}
	}
	if a.Name == "" {
		return &AlertValidationError{"Name не может быть пустым"}
	}
	if a.Condition == "" {
		return &AlertValidationError{"Condition не может быть пустым"}
	}
	if a.Action == "" {
		return &AlertValidationError{"Action не может быть пустым"}
	}
	if !a.Enabled {
		// Неактивные алерты не требуют валидации
		return nil
	}
	return nil
}

// AlertValidationError - ошибка валидации алерта
//
// Структура ошибки, содержащая сообщение о проблеме валидации.
type AlertValidationError struct {
	Message string
}

// Error - реализация интерфейса error
func (e *AlertValidationError) Error() string {
	return e.Message
}

// Severity levels - уровни критичности алертов
const (
	SeverityCritical = "critical" // Критическая ошибка, требует немедленного вмешательства
	SeverityHigh     = "high"     // Высокая критичность, требует внимания в ближайшее время
	SeverityMedium   = "medium"   // Средняя критичность, требует планового вмешательства
	SeverityLow      = "low"      // Низкая критичность, информационное сообщение
)

// Action types - типы действий при срабатывании алерта
const (
	ActionNotification = "notification" // Отправка уведомления администратору
	ActionLog          = "log"          // Запись в лог
	ActionScript       = "script"       // Выполнение скрипта
)

// GenerateAlertID - генерация уникального ID алерта
//
// Генерирует уникальный ID на основе текущего времени.
// Используется при создании новых алертов программно.
func generateAlertID() string {
	return "alert_" + time.Now().Format("20060102_150405_000")
}
