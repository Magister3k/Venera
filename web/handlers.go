// handlers.go - HTTP обработчики для веб-интерфейса Venera
//
// Этот модуль содержит HTTP обработчики для всех API endpoints
// веб-интерфейса системы сбора идентификаторов.
//
// Основные функции:
// - Обработка запросов процессов (CRUD)
// - Обработка запросов настроек (CRUD)
// - Обработка запросов статистики
// - Обработка запросов базы данных
// - Обработка запросов логов
// - Обработка запросов диагностики
//
// Использование:
// import "venera/web"
// server := web.NewServer(cfg, processManager, dragonflyDB, postgresDB)
// server.setupRoutes()

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"venera/config"
	"venera/data"
	"venera/processes"
)

// Handlers - структура обработчиков
type Handlers struct {
	cfg            *config.Config
	processManager *processes.ProcessManager
	dragonflyDB    *data.DragonflyDB
	postgresDB     *data.PostgreSQL
	log            *logrus.Logger
}

// NewHandlers - создание новых обработчиков
func NewHandlers(cfg *config.Config, processManager *processes.ProcessManager, dragonflyDB *data.DragonflyDB, postgresDB *data.PostgreSQL) *Handlers {
	return &Handlers{
		cfg:            cfg,
		processManager: processManager,
		dragonflyDB:    dragonflyDB,
		postgresDB:     postgresDB,
		log:            logrus.WithField("module", "handlers"),
	}
}

// SetupRoutes - настройка маршрутов
func (h *Handlers) SetupRoutes(router *gin.Engine) {
	// API маршруты
	api := router.Group("/api")
	{
		// Процессы
		api.GET("/processes", h.getProcesses)
		api.POST("/processes", h.addProcess)
		api.PUT("/processes/:id", h.updateProcess)
		api.DELETE("/processes/:id", h.deleteProcess)
		api.POST("/processes/:id/start", h.startProcess)
		api.POST("/processes/:id/stop", h.stopProcess)
		api.POST("/processes/start-all", h.startAllProcesses)
		api.POST("/processes/stop-all", h.stopAllProcesses)

		// Статистика
		api.GET("/statistics", h.getStatistics)
		api.GET("/statistics/process/:id", h.getProcessStatistics)

		// База данных
		api.GET("/db/data", h.getDBData)
		api.GET("/db/statistics", h.getDBStatistics)

		// Настройки
		api.GET("/settings", h.getSettings)
		api.POST("/settings", h.saveSettings)

		// Логи
		api.GET("/logs", h.getLogs)
		api.GET("/logs/file", h.getLogFile)

		// Диагностика
		api.GET("/diagnose", h.getDiagnose)

		// Выполнение команд
		api.POST("/commands/:command", h.executeCommand)
	}
}

// getProcesses - получение списка процессов
func (h *Handlers) getProcesses(c *gin.Context) {
	processes := h.processManager.GetAllProcesses()

	c.JSON(http.StatusOK, gin.H{
		"processes": processes,
		"count":     len(processes),
	})
}

// addProcess - добавление процесса
func (h *Handlers) addProcess(c *gin.Context) {
	var req struct {
		ID                 string `json:"id"`
		Type               string `json:"type"`
		Name               string `json:"name"`
		IP                 string `json:"ip"`
		Port               int    `json:"port"`
		Path               string `json:"path"`
		ScanSubfolders     bool   `json:"scan_subfolders"`
		MonitorNewFiles    bool   `json:"monitor_new_files"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := config.ProcessConfig{
		ID:             req.ID,
		Type:           req.Type,
		Name:           req.Name,
		IP:             req.IP,
		Port:           req.Port,
		Path:           req.Path,
		ScanSubfolders: req.ScanSubfolders,
		MonitorNewFiles: req.MonitorNewFiles,
	}

	if err := h.processManager.AddProcess(req.ID, cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Infof("Процесс %s добавлен", req.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Процесс добавлен",
		"id":      req.ID,
	})
}

// updateProcess - обновление процесса
func (h *Handlers) updateProcess(c *gin.Context) {
	id := c.Param("id")
	wrapper, err := h.processManager.GetProcess(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Name           string `json:"name"`
		IP             string `json:"ip"`
		Port           int    `json:"port"`
		Path           string `json:"path"`
		ScanSubfolders bool   `json:"scan_subfolders"`
		MonitorNewFiles bool  `json:"monitor_new_files"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновление конфигурации
	wrapper.Config.Name = req.Name
	wrapper.Config.IP = req.IP
	wrapper.Config.Port = req.Port
	wrapper.Config.Path = req.Path
	wrapper.Config.ScanSubfolders = req.ScanSubfolders
	wrapper.Config.MonitorNewFiles = req.MonitorNewFiles

	c.JSON(http.StatusOK, gin.H{
		"message": "Процесс обновлен",
		"id":      id,
	})
}

// deleteProcess - удаление процесса
func (h *Handlers) deleteProcess(c *gin.Context) {
	id := c.Param("id")
	if err := h.processManager.DeleteProcess(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Infof("Процесс %s удален", id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Процесс удален",
		"id":      id,
	})
}

// startProcess - запуск процесса
func (h *Handlers) startProcess(c *gin.Context) {
	id := c.Param("id")
	if err := h.processManager.StartProcess(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Infof("Процесс %s запущен", id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Процесс запущен",
		"id":      id,
	})
}

// stopProcess - остановка процесса
func (h *Handlers) stopProcess(c *gin.Context) {
	id := c.Param("id")
	if err := h.processManager.StopProcess(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Infof("Процесс %s остановлен", id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Процесс остановлен",
		"id":      id,
	})
}

// startAllProcesses - запуск всех процессов
func (h *Handlers) startAllProcesses(c *gin.Context) {
	if err := h.processManager.StartAll(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Все процессы запущены")

	c.JSON(http.StatusOK, gin.H{
		"message": "Все процессы запущены",
	})
}

// stopAllProcesses - остановка всех процессов
func (h *Handlers) stopAllProcesses(c *gin.Context) {
	if err := h.processManager.StopAll(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Все процессы остановлены")

	c.JSON(http.StatusOK, gin.H{
		"message": "Все процессы остановлены",
	})
}

// getStatistics - получение статистики
func (h *Handlers) getStatistics(c *gin.Context) {
	stats, err := h.postgresDB.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// getProcessStatistics - получение статистики процесса
func (h *Handlers) getProcessStatistics(c *gin.Context) {
	id := c.Param("id")
	wrapper, err := h.processManager.GetProcess(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"process": wrapper,
	})
}

// getDBData - получение данных из базы
func (h *Handlers) getDBData(c *gin.Context) {
	source := c.Query("source")
	key := c.Query("key")
	value := c.Query("value")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")

	records, err := h.postgresDB.GetDataByFilter(source, key, value, 0, 0, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"count":   len(records),
	})
}

// getDBStatistics - получение статистики базы
func (h *Handlers) getDBStatistics(c *gin.Context) {
	stats, err := h.postgresDB.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// getSettings - получение настроек
func (h *Handlers) getSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"settings": h.cfg,
	})
}

// saveSettings - сохранение настроек
func (h *Handlers) saveSettings(c *gin.Context) {
	var newCfg config.Config
	if err := c.ShouldBindJSON(&newCfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := newCfg.SaveConfig("config.toml"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cfg = &newCfg

	h.log.Info("Настройки сохранены")

	c.JSON(http.StatusOK, gin.H{
		"message": "Настройки сохранены",
	})
}

// getLogs - получение логов
func (h *Handlers) getLogs(c *gin.Context) {
	logDir := h.cfg.Logging.Directory

	// Получение списка лог-файлов
	files, err := os.ReadDir(logDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".log") || strings.HasSuffix(file.Name(), ".log.gz")) {
			logFiles = append(logFiles, file.Name())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logFiles,
		"count": len(logFiles),
	})
}

// getLogFile - получение файла логов
func (h *Handlers) getLogFile(c *gin.Context) {
	filename := c.Query("file")
	filePath := filepath.Join(h.cfg.Logging.Directory, filename)

	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Файл не найден"})
		return
	}

	// Отправка файла
	c.File(filePath)
}

// getDiagnose - получение диагностики
func (h *Handlers) getDiagnose(c *gin.Context) {
	// TODO: Реализовать получение диагностики
	c.JSON(http.StatusOK, gin.H{
		"diagnose": map[string]interface{}{},
	})
}

// executeCommand - выполнение команды
func (h *Handlers) executeCommand(c *gin.Context) {
	command := c.Param("command")

	// TODO: Реализовать выполнение команд
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Команда %s выполнена", command),
	})
}

// GetConfig - получение конфигурации
func (h *Handlers) GetConfig() *config.Config {
	return h.cfg
}
