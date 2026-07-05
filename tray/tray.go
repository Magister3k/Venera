// tray.go - Системный трей для Venera
//
// Этот модуль обеспечивает интеграцию с системным треем Windows
// для управления приложением Venera из трея.
//
// Основные функции:
// - Инициализация системного трея
// - Меню трея
// - Открытие веб-интерфейса по клику
// - Управление процессами из трея
// - Уведомления в трее
//
// Использование:
// import "venera/tray"
// trayIcon := tray.NewTray(logger, cfg, processManager, webServer)
// trayIcon.Initialize()
// trayIcon.Stop()

package tray

import (
	"fmt"
	"time"

	"github.com/getlantern/systray"
	"github.com/sirupsen/logrus"
	"venera/config"
	"venera/data"
	"venera/processes"
	"venera/web"
)

// Tray - структура системного трея
type Tray struct {
	logger        *logrus.Logger
	cfg           *config.Config
	processManager *processes.ProcessManager
	webServer     *web.Server
	log           *logrus.Logger
}

// NewTray - создание нового объекта трея
func NewTray(logger *logrus.Logger, cfg *config.Config, processManager *processes.ProcessManager, webServer *web.Server) *Tray {
	return &Tray{
		logger:         logger,
		cfg:            cfg,
		processManager: processManager,
		webServer:     webServer,
		log:           logrus.WithField("module", "tray"),
	}
}

// Initialize - инициализация системного трея
func (t *Tray) Initialize() {
	t.log.Info("Инициализация системного трея")

	systray.Run(t.onReady, t.onExit)
}

// onReady - вызывается при готовности трея
func (t *Tray) onReady() {
	t.log.Info("Системный трей готов")

	// Установка иконки
	systray.SetIcon(getIcon())

	// Установка заголовка
	systray.SetTitle("Venera")

	// Создание меню
	t.createMenu()
}

// onExit - вызывается при выходе из трея
func (t *Tray) onExit() {
	t.log.Info("Выход из системного трея")
}

// createMenu - создание меню трея
func (t *Tray) createMenu() {
	// Меню открытия веб-интерфейса
	systray.AddMenuItem("Открыть веб-интерфейс", "Открыть веб-интерфейс Venera").OnClick(func() {
		url := fmt.Sprintf("http://localhost:%d", t.cfg.Generic.WebServerPort)
		t.logger.Infof("Открытие веб-интерфейса: %s", url)
		// TODO: Открыть URL в браузере по умолчанию
	})

	// Меню запуска всех процессов
	startAllMenu := systray.AddMenuItem("Запустить все процессы", "Запустить все процессы обработки")
	startAllMenu.OnClick(func() {
		t.logger.Info("Запуск всех процессов из трея")
		if err := t.processManager.StartAll(); err != nil {
			t.showNotification("Ошибка", fmt.Sprintf("Ошибка запуска процессов: %v", err))
		} else {
			t.showNotification("Успех", "Все процессы запущены")
		}
	})

	// Меню остановки всех процессов
	stopAllMenu := systray.AddMenuItem("Остановить все процессы", "Остановить все процессы обработки")
	stopAllMenu.OnClick(func() {
		t.logger.Info("Остановка всех процессов из трея")
		if err := t.processManager.StopAll(); err != nil {
			t.showNotification("Ошибка", fmt.Sprintf("Ошибка остановки процессов: %v", err))
		} else {
			t.showNotification("Успех", "Все процессы остановлены")
		}
	})

	// Меню перезапуска всех процессов
	restartAllMenu := systray.AddMenuItem("Перезапустить все процессы", "Перезапустить все процессы обработки")
	restartAllMenu.OnClick(func() {
		t.logger.Info("Перезапуск всех процессов из трея")
		if err := t.processManager.StopAll(); err != nil {
			t.showNotification("Ошибка", fmt.Sprintf("Ошибка остановки процессов: %v", err))
		}
		if err := t.processManager.StartAll(); err != nil {
			t.showNotification("Ошибка", fmt.Sprintf("Ошибка запуска процессов: %v", err))
		} else {
			t.showNotification("Успех", "Все процессы перезапущены")
		}
	})

	// Разделитель
	systray.AddMenuItemSeparator()

	// Меню статистики
	statsMenu := systray.AddMenuItem("Статистика", "Просмотр статистики системы")
	statsMenu.OnClick(func() {
		t.logger.Info("Просмотр статистики из трея")
		// TODO: Показать статистику
	})

	// Меню логов
	logsMenu := systray.AddMenuItem("Логи", "Просмотр логов")
	logsMenu.OnClick(func() {
		t.logger.Info("Просмотр логов из трея")
		// TODO: Показать логи
	})

	// Меню диагностики
	diagnoseMenu := systray.AddMenuItem("Диагностика", "Запуск диагностики системы")
	diagnoseMenu.OnClick(func() {
		t.logger.Info("Запуск диагностики из трея")
		// TODO: Запустить диагностику
	})

	// Разделитель
	systray.AddMenuItemSeparator()

	// Меню настроек
	settingsMenu := systray.AddMenuItem("Настройки", "Открыть настройки")
	settingsMenu.OnClick(func() {
		t.logger.Info("Открытие настроек из трея")
		// TODO: Открыть настройки
	})

	// Меню о программе
	aboutMenu := systray.AddMenuItem("О программе", "Информация о Venera")
	aboutMenu.OnClick(func() {
		t.logger.Info("Просмотр информации о программе")
		t.showNotification("О программе", fmt.Sprintf("Venera v1.0.0\n%s", time.Now().Format("02.01.2006")))
	})

	// Меню выхода
	exitMenu := systray.AddMenuItem("Выход", "Выход из Venera")
	exitMenu.OnClick(func() {
		t.logger.Info("Выход из Venera")
		systray.Quit()
	})
}

// showNotification - показать уведомление
func (t *Tray) showNotification(title, message string) {
	t.log.Infof("Уведомление: %s - %s", title, message)

	// TODO: Реализовать уведомление через Windows API
	// systray.SetTooltip(fmt.Sprintf("%s: %s", title, message))
}

// getIcon - получение иконки
func getIcon() []byte {
	// TODO: Загрузка иконки из файла
	return nil
}

// Stop - остановка трея
func (t *Tray) Stop() {
	t.log.Info("Остановка системного трея")

	// TODO: Реализовать корректную остановку трея
}
