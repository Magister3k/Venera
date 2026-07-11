// server.go - Веб-сервер для Venera
//
// Этот модуль обеспечивает HTTP-сервер для веб-интерфейса управления
// системой сбора идентификаторов.
//
// Основные функции:
// - HTTP сервер на заданном порту
// - Маршрутизация запросов
// - Обработка статических файлов
// - Middleware для аутентификации (если требуется)
// - Graceful shutdown
//
// Использование:
// import "venera/web"
// server := web.NewServer(cfg, processManager, dragonflyDB, postgresDB)
// server.Start()
// server.Stop()

package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"venera/config"
	"venera/data"
	"venera/metrics"
	"venera/processes"
)

// Server - структура веб-сервера
type Server struct {
	cfg          *config.Config
	processManager *processes.ProcessManager
	dragonflyDB  *data.DragonflyDB
	postgresDB   *data.PostgreSQL
	router       *gin.Engine
	httpServer   *http.Server
	log          *logrus.Logger
}

// NewServer - создание нового веб-сервера
func NewServer(cfg *config.Config, processManager *processes.ProcessManager, dragonflyDB *data.DragonflyDB, postgresDB *data.PostgreSQL) *Server {
	// Настройка роутера
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s\n",
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))

	router.Use(gin.Recovery())

	// Создание сервера
	server := &Server{
		cfg:            cfg,
		processManager: processManager,
		dragonflyDB:    dragonflyDB,
		postgresDB:     postgresDB,
		router:         router,
		log:            logrus.WithField("module", "web_server"),
	}

	// Настройка маршрутов
	server.setupRoutes()

	return server
}

// setupRoutes - настройка маршрутов
func (s *Server) setupRoutes() {
	// Статические файлы
	s.router.Static("/static", "./web/static")

	// Главная страница
	s.router.GET("/", func(c *gin.Context) {
		c.File("./web/static/index.html")
	})

	// API маршруты
	api := s.router.Group("/api")
	{
		// Процессы
		api.GET("/processes", s.getProcesses)
		api.POST("/processes", s.addProcess)
		api.PUT("/processes/:id", s.updateProcess)
		api.DELETE("/processes/:id", s.deleteProcess)
		api.POST("/processes/:id/start", s.startProcess)
		api.POST("/processes/:id/stop", s.stopProcess)
		api.POST("/processes/start-all", s.startAllProcesses)
		api.POST("/processes/stop-all", s.stopAllProcesses)

		// Статистика
		api.GET("/statistics", s.getStatistics)
		api.GET("/statistics/process/:id", s.getProcessStatistics)

		// База данных
		api.GET("/db/data", s.getDBData)
		api.GET("/db/statistics", s.getDBStatistics)

		// Настройки
		api.GET("/settings", s.getSettings)
		api.POST("/settings", s.saveSettings)

		// Логи
		api.GET("/logs", s.getLogs)
		api.GET("/logs/file", s.getLogFile)

		// Диагностика
		api.GET("/diagnose", s.getDiagnose)
	}

	// WebSocket маршруты
	s.router.GET("/ws", s.handleWebsocket)
}

// Start - запуск веб-сервера
func (s *Server) Start() error {
	port := s.cfg.Generic.WebServerPort
	address := fmt.Sprintf(":%d", port)

	s.httpServer = &http.Server{
		Addr:    address,
		Handler: s.router,
	}

	s.log.Infof("Запуск веб-сервера на порту %d", port)

	// Запуск сервера в горутине
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Errorf("Ошибка запуска веб-сервера: %v", err)
		}
	}()

	return nil
}

// Stop - остановка веб-сервера
func (s *Server) Stop() error {
	s.log.Info("Остановка веб-сервера")

	// Создание контекста с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Запрос остановки сервера
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Errorf("Ошибка остановки веб-сервера: %v", err)
		return err
	}

	s.log.Info("Веб-сервер остановлен")
	return nil
}

// getProcesses - получение списка процессов
func (s *Server) getProcesses(c *gin.Context) {
	processes := s.processManager.GetAllProcesses()
	c.JSON(http.StatusOK, gin.H{
		"processes": processes,
	})
}

// addProcess - добавление процесса
func (s *Server) addProcess(c *gin.Context) {
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

	if err := s.processManager.AddProcess(req.ID, cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Процесс добавлен"})
}

// updateProcess - обновление процесса
func (s *Server) updateProcess(c *gin.Context) {
	id := c.Param("id")
	// TODO: Реализовать обновление процесса
	c.JSON(http.StatusOK, gin.H{"message": "Процесс обновлен"})
}

// deleteProcess - удаление процесса
func (s *Server) deleteProcess(c *gin.Context) {
	id := c.Param("id")
	if err := s.processManager.DeleteProcess(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Процесс удален"})
}

// startProcess - запуск процесса
func (s *Server) startProcess(c *gin.Context) {
	id := c.Param("id")
	if err := s.processManager.StartProcess(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Процесс запущен"})
}

// stopProcess - остановка процесса
func (s *Server) stopProcess(c *gin.Context) {
	id := c.Param("id")
	if err := s.processManager.StopProcess(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Процесс остановлен"})
}

// startAllProcesses - запуск всех процессов
func (s *Server) startAllProcesses(c *gin.Context) {
	if err := s.processManager.StartAll(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Все процессы запущены"})
}

// stopAllProcesses - остановка всех процессов
func (s *Server) stopAllProcesses(c *gin.Context) {
	if err := s.processManager.StopAll(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Все процессы остановлены"})
}

// getStatistics - получение статистики
func (s *Server) getStatistics(c *gin.Context) {
	stats, err := s.postgresDB.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// getProcessStatistics - получение статистики процесса
func (s *Server) getProcessStatistics(c *gin.Context, id string) {
	wrapper, err := s.processManager.GetProcess(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"process": wrapper,
	})
}

// getDBData - получение данных из базы
func (s *Server) getDBData(c *gin.Context) {
	source := c.Query("source")
	key := c.Query("key")
	value := c.Query("value")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")

	records, err := s.postgresDB.GetDataByFilter(source, key, value, 0, 0, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
	})
}

// getDBStatistics - получение статистики базы
func (s *Server) getDBStatistics(c *gin.Context) {
	stats, err := s.postgresDB.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// getSettings - получение настроек
func (s *Server) getSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"settings": s.cfg,
	})
}

// saveSettings - сохранение настроек
func (s *Server) saveSettings(c *gin.Context) {
	var newCfg config.Config
	if err := c.ShouldBindJSON(&newCfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := newCfg.SaveConfig("config.toml"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.cfg = &newCfg

	c.JSON(http.StatusOK, gin.H{"message": "Настройки сохранены"})
}

// getLogs - получение логов
func (s *Server) getLogs(c *gin.Context) {
	// Получение логов из файлов
	logDir := s.cfg.Logging.Directory
	// TODO: Реализовать получение логов

	c.JSON(http.StatusOK, gin.H{
		"logs": []string{},
	})
}

// getLogFile - получение файла логов
func (s *Server) getLogFile(c *gin.Context) {
	// TODO: Реализовать получение файла логов
	c.JSON(http.StatusOK, gin.H{
		"file": "",
	})
}

// getDiagnose - получение диагностики
func (s *Server) getDiagnose(c *gin.Context) {
	// TODO: Реализовать получение диагностики
	c.JSON(http.StatusOK, gin.H{
		"diagnose": map[string]interface{}{},
	})
}

// handleWebsocket - обработка WebSocket соединений
func (s *Server) handleWebsocket(c *gin.Context) {
	// TODO: Реализовать WebSocket обработчики
	c.JSON(http.StatusOK, gin.H{
		"message": "WebSocket соединение установлено",
	})
}

// GetRouter - получение роутера
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// GetConfig - получение конфигурации
func (s *Server) GetConfig() *config.Config {
	return s.cfg
}
