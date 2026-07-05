// win_service.go - Windows служба для Venera
//
// Этот модуль обеспечивает реализацию Windows службы для приложения Venera,
// включая установку, запуск, остановку и удаление службы.
//
// Основные функции:
// - Реализация Windows сервиса
// - Установка/удаление службы
// - Запуск/остановка службы
// - Обработка команд службы
// - Права администратора
//
// Использование:
// import "venera/services"
// services.InstallService("VeneraSrv", "C:\\Venera\\venera.exe")
// services.RunService("VeneraSrv", handler)

package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Service - структура службы
type Service struct {
	Name        string
	DisplayName string
	Path        string
	AutoStart   bool
	Description string
}

// InstallService - установка службы
func InstallService(name, path string) error {
	// Проверка прав администратора
	if !isAdmin() {
		return fmt.Errorf("для установки службы необходимы права администратора")
	}

	// Проверка существования файла
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("файл приложения не найден: %s", path)
	}

	// Установка службы через sc.exe
	cmd := exec.Command("sc.exe", "create", name,
		"binPath=", filepath.Join(filepath.Dir(path), filepath.Base(path)),
		"start=", "auto",
		"obj=", "LocalSystem",
		"depend=", "TCP/IP",
		" DisplayName=", "Venera Service")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка установки службы: %w, output: %s", err, string(output))
	}

	// Настройка описания службы
	if err := setServiceDescription(name, "Служба сбора идентификаторов в потоке пакетных данных"); err != nil {
		return fmt.Errorf("ошибка установки описания службы: %w", err)
	}

	return nil
}

// UninstallService - удаление службы
func UninstallService(name string) error {
	// Проверка прав администратора
	if !isAdmin() {
		return fmt.Errorf("для удаления службы необходимы права администратора")
	}

	// Остановка службы, если она запущена
	if isServiceRunning(name) {
		if err := StopService(name); err != nil {
			return fmt.Errorf("ошибка остановки службы перед удалением: %w", err)
		}
	}

	// Удаление службы через sc.exe
	cmd := exec.Command("sc.exe", "delete", name)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка удаления службы: %w, output: %s", err, string(output))
	}

	return nil
}

// StartService - запуск службы
func StartService(name string) error {
	// Открытие SCM
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("ошибка открытия SCM: %w", err)
	}
	defer windows.CloseServiceHandle(scm)

	// Открытие службы
	service, err := windows.OpenService(scm, syscall.StringToUTF16Ptr(name), windows.SERVICE_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("ошибка открытия службы: %w", err)
	}
	defer windows.CloseServiceHandle(service)

	// Запуск службы
	if err := windows.StartService(service, 0, nil); err != nil {
		return fmt.Errorf("ошибка запуска службы: %w", err)
	}

	return nil
}

// StopService - остановка службы
func StopService(name string) error {
	// Открытие SCM
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("ошибка открытия SCM: %w", err)
	}
	defer windows.CloseServiceHandle(scm)

	// Открытие службы
	service, err := windows.OpenService(scm, syscall.StringToUTF16Ptr(name), windows.SERVICE_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("ошибка открытия службы: %w", err)
	}
	defer windows.CloseServiceHandle(service)

	// Получение статуса службы
	var status windows.SERVICE_STATUS
	if err := windows.ControlService(service, windows.SERVICE_CONTROL_STOP, &status); err != nil {
		return fmt.Errorf("ошибка остановки службы: %w", err)
	}

	// Ожидание остановки
	for i := 0; i < 30; i++ {
		if err := windows.QueryServiceStatus(service, &status); err != nil {
			return fmt.Errorf("ошибка проверки статуса службы: %w", err)
		}

		if status.CurrentState == windows.SERVICE_STOPPED {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("время ожидания остановки службы истекло")
}

// isServiceRunning - проверка запущенности службы
func isServiceRunning(name string) bool {
	// Открытие SCM
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return false
	}
	defer windows.CloseServiceHandle(scm)

	// Открытие службы
	service, err := windows.OpenService(scm, syscall.StringToUTF16Ptr(name), windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return false
	}
	defer windows.CloseServiceHandle(service)

	// Получение статуса службы
	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(service, &status); err != nil {
		return false
	}

	return status.CurrentState == windows.SERVICE_RUNNING
}

// setServiceDescription - установка описания службы
func setServiceDescription(name, description string) error {
	// Открытие SCM
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("ошибка открытия SCM: %w", err)
	}
	defer windows.CloseServiceHandle(scm)

	// Открытие службы
	service, err := windows.OpenService(scm, syscall.StringToUTF16Ptr(name), windows.SERVICE_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("ошибка открытия службы: %w", err)
	}
	defer windows.CloseServiceHandle(service)

	// Установка описания
	descriptionUTF16 := syscall.StringToUTF16Ptr(description)
	if err := windows.ChangeServiceConfig2(service, windows.SERVICE_CONFIG_DESCRIPTION, (*byte)(unsafe.Pointer(&descriptionUTF16))); err != nil {
		return fmt.Errorf("ошибка установки описания службы: %w", err)
	}

	return nil
}

// RunService - запуск службы
func RunService(name string, handler func()) error {
	// Создание обработчика службы
	serviceHandler := &ServiceHandler{
		Name:    name,
		Handler: handler,
	}

	// Регистрация обработчика службы
	err := windows.StartServiceCtrlDispatcher(serviceHandler.dispatchTable())
	if err != nil {
		return fmt.Errorf("ошибка запуска диспетчера служб: %w", err)
	}

	return nil
}

// ServiceHandler - обработчик службы
type ServiceHandler struct {
	Name    string
	Handler func()
}

// dispatchTable - таблица диспетчера служб
func (h *ServiceHandler) dispatchTable() *windows.SERVICE_TABLE_ENTRY {
	return &windows.SERVICE_TABLE_ENTRY{
		ServiceName: syscall.StringToUTF16Ptr(h.Name),
		ServiceProc: syscall.NewCallback(h.serviceMain),
	}
}

// serviceMain - главная функция службы
func (h *ServiceHandler) serviceMain(argc uint32, argv **uint16) uintptr {
	// Регистрация обработчика управления службой
	hStatusHandle, err := windows.RegisterServiceCtrlHandler(
		syscall.StringToUTF16Ptr(h.Name),
		syscall.NewCallback(h.serviceControlHandler))
	if err != nil {
		return 1
	}

	// Установка статуса службы
	h.setServiceStatus(windows.SERVICE_START_PENDING, 0, 3000)

	// Запуск обработчика
	go func() {
		h.setServiceStatus(windows.SERVICE_RUNNING, 0, 0)
		h.Handler()
		h.setServiceStatus(windows.SERVICE_STOPPED, 0, 0)
	}()

	// Ожидание завершения
	select {}
}

// serviceControlHandler - обработчик управления службой
func (h *ServiceHandler) serviceControlHandler(control uint32) (result uint32) {
	switch control {
	case windows.SERVICE_CONTROL_STOP:
		result = windows.NO_ERROR
	case windows.SERVICE_CONTROL_SHUTDOWN:
		result = windows.NO_ERROR
	case windows.SERVICE_CONTROL_INTERROGATE:
		result = windows.NO_ERROR
	default:
		result = windows.ERROR_CALL_NOT_IMPLEMENTED
	}

	return result
}

// setServiceStatus - установка статуса службы
func (h *ServiceHandler) setServiceStatus(state, win32ExitCode, checkPoint uint32) {
	if hStatusHandle != 0 {
		status := windows.SERVICE_STATUS{
			ServiceType:        windows.SERVICE_WIN32_OWN_PROCESS,
		(CurrentState:     state,
			ControlsAccepted: windows.SERVICE_ACCEPT_STOP | windows.SERVICE_ACCEPT_SHUTDOWN,
			Win32ExitCode:    win32ExitCode,
			ServiceSpecificExitCode: 0,
			CheckPoint:       checkPoint,
			WaitHint:         3000,
		}
		windows.SetServiceStatus(hStatusHandle, &status)
	}
}

// isAdmin - проверка прав администратора
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

// unsafe - импорт unsafe
import "unsafe"

// hStatusHandle - дескриптор службы
var hStatusHandle windows.Handle
