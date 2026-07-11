// tshark.go - Интеграция с Tshark для сбора пакетов
//
// Этот модуль обеспечивает взаимодействие с Tshark для захвата пакетов
// и экспорта данных в формате JSON.
//
// Основные функции:
// - Запуск Tshark с параметрами
// - Обработка потоков данных из Tshark
// - Поддержка разных типов источников (сетевой, папка, файл)
// - Обработка ошибок Tshark
// - Корректное управление дескрипторами (stdin/stdout/stderr)
//
// Использование:
// import "venera/processes"
// tshark := NewTshark(cfg)
// tshark.Start()
// tshark.Stop()

package processes

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Tshark - структура управления Tshark
type Tshark struct {
	cfg         TsharkConfig
	cmd         *exec.Cmd
	ctx         context.Context
	cancel      context.CancelFunc
	outputChan  chan string
	errorChan   chan error
	wg          sync.WaitGroup
	log         *logrus.Logger
	stdoutPipe  io.ReadCloser // Сохраняем pipe для корректного закрытия
	stderrPipe  io.ReadCloser // Сохраняем pipe для корректного закрытия
}

// TsharkConfig - конфигурация Tshark
type TsharkConfig struct {
	Path      string
	Type      string // "network", "folder", "file"
	IP        string
	Port      int
	PathInput string
	Filter    string
}

// NewTshark - создание нового экземпляра Tshark
func NewTshark(cfg TsharkConfig) *Tshark {
	return &Tshark{
		outputChan: make(chan string, 1000),
		errorChan:  make(chan error, 10),
		log:        logrus.WithField("module", "tshark"),
	}
}

// Start - запуск Tshark
func (t *Tshark) Start() error {
	// Проверка наличия файла
	if _, err := exec.LookPath(t.cfg.Path); err != nil {
		return fmt.Errorf("tshark не найден: %w", err)
	}

	// Создание контекста
	t.ctx, t.cancel = context.WithCancel(context.Background())

	// Формирование команды
	args := t.buildArgs()
	t.log.Infof("Запуск Tshark с аргументами: %v", args)

	// Создание команды
	t.cmd = exec.CommandContext(t.ctx, t.cfg.Path, args...)

	// Настройка вывода
	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ошибка создания pipe для stdout: %w", err)
	}
	t.stdoutPipe = stdout // Сохраняем pipe для корректного закрытия

	stderr, err := t.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("ошибка создания pipe для stderr: %w", err)
	}
	t.stderrPipe = stderr // Сохраняем pipe для корректного закрытия

	// Запуск команды
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("ошибка запуска Tshark: %w", err)
	}

	// Запуск горутин для чтения вывода
	t.wg.Add(2)
	go t.readOutput(stdout)
	go t.readError(stderr)

	t.log.Info("Tshark запущен успешно")
	return nil
}

// Stop - остановка Tshark
//
// Корректное управление ресурсами:
// 1. Отмена контекста приводит к завершению всех горутин через ctx.Done()
// 2. defer reader.Close() в горутинах гарантирует закрытие pipe'ов
// 3. Закрытие каналов outputChan/errorChan после завершения горутин
func (t *Tshark) Stop() error {
	t.log.Info("Остановка Tshark")

	// Отмена контекста - горутины readOutput/readError завершатся через ctx.Done()
	if t.cancel != nil {
		t.cancel()
	}

	// Ожидание завершения горутин
	// defer reader.Close() в горутинах закроет pipe'ы автоматически
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	// Таймаут ожидания (5 секунд)
	select {
	case <-done:
		t.log.Info("Горутины Tshark успешно завершены")
	case <-time.After(5 * time.Second):
		t.log.Warn("Превышено время ожидания завершения Tshark, принудительная остановка")
		if t.cmd != nil && t.cmd.Process != nil {
			t.cmd.Process.Kill()
			t.log.Info("Процесс Tshark принудительно завершен")
		}
		// Ждем завершения горутин после Kill
		t.wg.Wait()
	}

	// Закрытие каналов после завершения горутин
	close(t.outputChan)
	close(t.errorChan)

	t.log.Info("Tshark остановлен")
	return nil
}

// buildArgs - формирование аргументов команды
func (t *Tshark) buildArgs() []string {
	var args []string

	// Общие аргументы
	args = append(args, "-i", "-") // Входной поток из stdin
	args = append(args, "-T", "json") // Формат JSON
	args = append(args, "-e", "frame.time_epoch") // Время фиксации

	// В зависимости от типа источника
	switch t.cfg.Type {
	case "network":
		args = append(args, "-f", fmt.Sprintf("udp port %d", t.cfg.Port))
		// Дополнительные фильтры для сетевого источника
		if t.cfg.IP != "" {
			args = append(args, "-Y", fmt.Sprintf("ip.addr == %s", t.cfg.IP))
		}
	case "folder":
		// Для папки используем файлы как вход
		args = append(args, "-r", t.cfg.PathInput)
	case "file":
		// Для отдельного файла
		args = append(args, "-r", t.cfg.PathInput)
	}

	// Дополнительные фильтры
	if t.cfg.Filter != "" {
		args = append(args, "-Y", t.cfg.Filter)
	}

	return args
}

// readOutput - чтение стандартного вывода
func (t *Tshark) readOutput(reader io.ReadCloser) {
	defer t.wg.Done()
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			select {
			case t.outputChan <- line:
			case <-t.ctx.Done():
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errorChan <- fmt.Errorf("ошибка чтения вывода: %w", err):
		case <-t.ctx.Done():
		}
	}
}

// readError - чтение стандартного вывода ошибок
func (t *Tshark) readError(reader io.ReadCloser) {
	defer t.wg.Done()
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			t.log.Warnf("Tshark stderr: %s", line)
			select {
			case t.errorChan <- fmt.Errorf("tshark error: %s", line):
			case <-t.ctx.Done():
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errorChan <- fmt.Errorf("ошибка чтения ошибок: %w", err):
		case <-t.ctx.Done():
		}
	}
}

// OutputChan - получение канала вывода
func (t *Tshark) OutputChan() <-chan string {
	return t.outputChan
}

// ErrorChan - получение канала ошибок
func (t *Tshark) ErrorChan() <-chan error {
	return t.errorChan
}

// IsRunning - проверка запущенности
func (t *Tshark) IsRunning() bool {
	if t.cmd == nil {
		return false
	}
	return t.cmd.ProcessState == nil || !t.cmd.ProcessState.Exited()
}

// GetPID - получение PID процесса
func (t *Tshark) GetPID() int {
	if t.cmd != nil && t.cmd.Process != nil {
		return t.cmd.Process.Pid
	}
	return -1
}

// Restart - перезапуск Tshark
func (t *Tshark) Restart() error {
	if err := t.Stop(); err != nil {
		t.log.Warnf("Ошибка остановки Tshark: %v", err)
	}
	return t.Start()
}

// GetConfig - получение конфигурации
func (t *Tshark) GetConfig() TsharkConfig {
	return t.cfg
}
