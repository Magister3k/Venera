// callbacks.go - Модуль обратных вызовов для предотвращения зацикленности модулей
//
// Этот модуль обеспечивает механизм обратных вызовов для предотвращения зацикленности
// модулей и горутин в приложении Venera. Поддерживает обнаружение зацикленности,
// таймауты операций и корректное завершение работы.
//
// Основные функции:
// - Регистрация и выполнение callback-функций
// - Обнаружение зацикленности (deadlock detection)
// - Установка таймаутов для операций
// - Отмена операций
// - Логирование событий зацикленности
// - Интеграция с логером и метриками
// - Graceful Shutdown
//
// Использование:
// import "venera/callbacks"
//
// manager := callbacks.NewCallbackManager(logger, metrics)
// manager.RegisterCallback("deadlock", func(args ...interface{}) error {
//     // Обработка зацикленности
//     return nil
// })
// manager.ExecuteCallback("deadlock", data)
//
// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// defer cancel()
// if err := manager.DetectDeadlock(ctx, operations); err != nil {
//     // Обработка зацикленности
// }

package callbacks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"venera/metrics"
)

// CallbackFunc - тип функции обратного вызова
type CallbackFunc func(args ...interface{}) error

// DeadlockInfo - информация о зацикленности
type DeadlockInfo struct {
	Operations []string    `json:"operations"`
	Duration   time.Duration `json:"duration"`
	Timestamp  time.Time   `json:"timestamp"`
}

// CallbackManager - менеджер обратных вызовов
type CallbackManager struct {
	callbacks     map[string]CallbackFunc
	callbacksMu   sync.RWMutex
	deadlockInfo  *DeadlockInfo
	deadlockMu    sync.RWMutex
	logger        *logrus.Logger
	metrics       *metrics.Collector
	shutdownChan  chan struct{}
	shutdownOnce  sync.Once
}

// NewCallbackManager - создание нового менеджера обратных вызовов
func NewCallbackManager(logger *logrus.Logger, metrics *metrics.Collector) *CallbackManager {
	return &CallbackManager{
		callbacks:    make(map[string]CallbackFunc),
		logger:       logger.WithField("module", "callbacks"),
		metrics:      metrics,
		shutdownChan: make(chan struct{}),
	}
}

// RegisterCallback - регистрация callback-функции
// name - уникальное имя callback
// fn - функция обратного вызова
func (cm *CallbackManager) RegisterCallback(name string, fn CallbackFunc) error {
	cm.callbacksMu.Lock()
	defer cm.callbacksMu.Unlock()

	if _, exists := cm.callbacks[name]; exists {
		return fmt.Errorf("callback с именем '%s' уже зарегистрирован", name)
	}

	cm.callbacks[name] = fn
	cm.logger.Debugf("Зарегистрирован callback: %s", name)

	return nil
}

// ExecuteCallback - выполнение callback-функции
// name - имя callback
// args - аргументы для callback
func (cm *CallbackManager) ExecuteCallback(name string, args ...interface{}) error {
	cm.callbacksMu.RLock()
	fn, exists := cm.callbacks[name]
	cm.callbacksMu.RUnlock()

	if !exists {
		return fmt.Errorf("callback с именем '%s' не найден", name)
	}

	// Логирование начала выполнения
	cm.logger.Debugf("Выполнение callback: %s", name)

	// Выполнение callback
	err := fn(args...)

	// Логирование завершения
	if err != nil {
		cm.logger.Errorf("Ошибка при выполнении callback '%s': %v", name, err)
	} else {
		cm.logger.Debugf("Callback '%s' успешно выполнен", name)
	}

	return err
}

// DetectDeadlock - обнаружение зацикленности
// ctx - контекст для отмены операции
// operations - список операций для проверки
func (cm *CallbackManager) DetectDeadlock(ctx context.Context, operations []string) error {
	// Создание контекста с таймаутом для обнаружения зацикленности
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Мапа для отслеживания состояния операций
	operationStates := make(map[string]bool)
	for _, op := range operations {
		operationStates[op] = false
	}

	// Мапа для отслеживания времени начала операций
	operationStartTimes := make(map[string]time.Time)
	for _, op := range operations {
		operationStartTimes[op] = time.Now()
	}

	// Канал для уведомления о завершении операций
	done := make(chan struct{})

	// Проверка операций в отдельной горутине
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Таймаут - возможно зацикленность
				cm.handlePotentialDeadlock(operations, operationStartTimes)
				return
			case <-done:
				// Все операции завершены
				return
			}
		}
	}()

	// Проверка состояния операций
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Таймаут - зацикленность обнаружена
			return cm.reportDeadlock(operations)
		case <-done:
			// Завершение без зацикленности
			return nil
		case <-ticker.C:
			// Проверка состояния операций
			allFinished := true
			for _, op := range operations {
				if !operationStates[op] {
					allFinished = false
					break
				}
			}

			if allFinished {
				close(done)
				return nil
			}
		}
	}
}

// handlePotentialDeadlock - обработка потенциальной зацикленности
func (cm *CallbackManager) handlePotentialDeadlock(operations []string, startTimes map[string]time.Time) {
	cm.deadlockMu.Lock()
	defer cm.deadlockMu.Unlock()

	// Запись информации о зацикленности
	cm.deadlockInfo = &DeadlockInfo{
		Operations: operations,
		Duration:   time.Since(startTimes[operations[0]]),
		Timestamp:  time.Now(),
	}

	// Выполнение callback зацикленности
	_ = cm.ExecuteCallback("deadlock", cm.deadlockInfo)

	// Логирование
	cm.logger.Errorf("Обнаружена потенциальная зацикленность: операции=%v, duration=%v",
		operations, cm.deadlockInfo.Duration)

	// Обновление метрик
	if cm.metrics != nil {
		cm.metrics.RecordDeadlock(operations)
	}
}

// reportDeadlock - сообщение о зацикленности
func (cm *CallbackManager) reportDeadlock(operations []string) error {
	cm.deadlockMu.Lock()
	defer cm.deadlockMu.Unlock()

	// Выполнение callback зацикленности
	if err := cm.ExecuteCallback("deadlock", operations); err != nil {
		cm.logger.Errorf("Ошибка при выполнении callback зацикленности: %v", err)
		return err
	}

	// Логирование
	cm.logger.Errorf("Зацикленность обнаружена: операции=%v", operations)

	return fmt.Errorf("обнаружена зацикленность: операции=%v", operations)
}

// SetTimeout - установка таймаута для операции
// name - имя операции
// duration - длительность таймаута
func (cm *CallbackManager) SetTimeout(name string, duration time.Duration) {
	// Создание контекста с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Логирование
	cm.logger.Debugf("Установлен таймаут для операции '%s': %v", name, duration)

	// Создание канала для уведомления о завершении
	done := make(chan struct{})

	// Запуск проверки таймаута в отдельной горутине
	go func() {
		select {
		case <-ctx.Done():
			// Таймаут
			cm.logger.Warnf("Таймаут для операции '%s' истек", name)
			_ = cm.ExecuteCallback("timeout", name)
		case <-done:
			// Операция завершена вовремя
			cm.logger.Debugf("Операция '%s' завершена вовремя", name)
		}
	}()
}

// CancelOperation - отмена операции
// name - имя операции
func (cm *CallbackManager) CancelOperation(name string) {
	// Логирование
	cm.logger.Debugf("Отмена операции: %s", name)

	// Выполнение callback отмены
	_ = cm.ExecuteCallback("cancel", name)
}

// Shutdown - корректное завершение работы менеджера
func (cm *CallbackManager) Shutdown() error {
	cm.shutdownOnce.Do(func() {
		cm.logger.Info("Остановка менеджера обратных вызовов")

		// Закрытие канала shutdown
		close(cm.shutdownChan)

		// Выполнение всех callback завершения
		_ = cm.ExecuteCallback("shutdown")
	})

	return nil
}

// GetDeadlockInfo - получение информации о зацикленности
func (cm *CallbackManager) GetDeadlockInfo() *DeadlockInfo {
	cm.deadlockMu.RLock()
	defer cm.deadlockMu.RUnlock()

	if cm.deadlockInfo == nil {
		return nil
	}

	// Возврат копии
	return &DeadlockInfo{
		Operations: append([]string{}, cm.deadlockInfo.Operations...),
		Duration:   cm.deadlockInfo.Duration,
		Timestamp:  cm.deadlockInfo.Timestamp,
	}
}

// RegisterDefaultCallbacks - регистрация callback-функций по умолчанию
func (cm *CallbackManager) RegisterDefaultCallbacks() {
	// Callback для зацикленности
	_ = cm.RegisterCallback("deadlock", func(args ...interface{}) error {
		// Обработка зацикленности
		if info, ok := args[0].(*DeadlockInfo); ok {
			cm.logger.Errorf("CB: Deadlock detected for operations: %v, duration: %v",
				info.Operations, info.Duration)
		} else if ops, ok := args[0].([]string); ok {
			cm.logger.Errorf("CB: Deadlock detected for operations: %v", ops)
		}
		return nil
	})

	// Callback для таймаута
	_ = cm.RegisterCallback("timeout", func(args ...interface{}) error {
		if name, ok := args[0].(string); ok {
			cm.logger.Warnf("CB: Timeout for operation: %s", name)
		}
		return nil
	})

	// Callback для отмены
	_ = cm.RegisterCallback("cancel", func(args ...interface{}) error {
		if name, ok := args[0].(string); ok {
			cm.logger.Debugf("CB: Cancel operation: %s", name)
		}
		return nil
	})

	// Callback для завершения
	_ = cm.RegisterCallback("shutdown", func(args ...interface{}) error {
		cm.logger.Info("CB: Shutdown completed")
		return nil
	})
}
