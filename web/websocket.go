// websocket.go - WebSocket сервер для веб-интерфейса Venera
//
// Этот модуль обеспечивает WebSocket соединение для веб-интерфейса
// Venera, позволяя получать метрики в реальном времени.
//
// Основные функции:
// - Управление WebSocket соединениями
// - Рассылка метрик подключенным клиентам
// - Обработка входящих сообщений от клиентов
// - Планировщик отправки метрик
//
// Использование:
// import "venera/web"
// ws := web.NewWebSocketServer(cfg, logger)
// ws.Start()
// ws.Stop()

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"venera/config"
	"venera/metrics"
)

// WebSocketServer - сервер WebSocket
type WebSocketServer struct {
	cfg       *config.Config
	logger    *logrus.Logger
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	mu        sync.Mutex
	metrics   map[string]interface{}
	metricsCollector *metrics.Collector
}

// MetricMessage - сообщение с метриками
type MetricMessage struct {
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewWebSocketServer - создание нового WebSocket сервера
func NewWebSocketServer(cfg *config.Config, logger *logrus.Logger, metricsCollector *metrics.Collector) *WebSocketServer {
	return &WebSocketServer{
		cfg:              cfg,
		logger:           logger,
		upgrader:         websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Разрешаем все соединения
			},
		},
		clients:          make(map[*websocket.Conn]bool),
		metricsCollector: metricsCollector,
	}
}

// Start - запуск WebSocket сервера
func (ws *WebSocketServer) Start() {
	ws.logger.Info("Запуск WebSocket сервера")

	// Запуск планировщика отправки метрик
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			ws.broadcastMetrics()
		}
	}()
}

// Stop - остановка WebSocket сервера
func (ws *WebSocketServer) Stop() {
	ws.logger.Info("Остановка WebSocket сервера")

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Закрытие всех соединений
	for client := range ws.clients {
		client.Close()
	}

	ws.clients = make(map[*websocket.Conn]bool)
}

// HandleConnections - обработка WebSocket соединений
func (ws *WebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Апгрейд HTTP соединения до WebSocket
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Errorf("Ошибка апгрейда WebSocket: %v", err)
		return
	}

	// Регистрация клиента
	ws.mu.Lock()
	ws.clients[conn] = true
	ws.mu.Unlock()

	ws.logger.WithField("client", conn.RemoteAddr()).Info("Подключен новый клиент WebSocket")

	// Обработка входящих сообщений
	go ws.handleMessages(conn)

	// Ожидание закрытия соединения
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	// Удаление клиента
	ws.mu.Lock()
	delete(ws.clients, conn)
	ws.mu.Unlock()

	conn.Close()
	ws.logger.WithField("client", conn.RemoteAddr()).Info("Клиент отключен")
}

// handleMessages - обработка входящих сообщений от клиентов
func (ws *WebSocketServer) handleMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			ws.logger.Errorf("Ошибка чтения сообщения: %v", err)
			break
		}

		// Обработка сообщения
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			ws.logger.Warnf("Ошибка декодирования сообщения: %v", err)
			continue
		}

		// Обработка разных типов сообщений
		ws.handleClientMessage(conn, msg)
	}
}

// handleClientMessage - обработка сообщения от клиента
func (ws *WebSocketServer) handleClientMessage(conn *websocket.Conn, msg map[string]interface{}) {
	if msg["type"] == "get_metrics" {
		// Отправка текущих метрик
		ws.sendMetrics(conn)
	} else if msg["type"] == "subscribe" {
		// Подписка на обновления
		ws.logger.Debug("Клиент подписался на обновления")
	} else if msg["type"] == "unsubscribe" {
		// Отписка от обновлений
		ws.logger.Debug("Клиент отписался от обновлений")
	}
}

// broadcastMetrics - рассылка метрик всем клиентам
func (ws *WebSocketServer) broadcastMetrics() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Получение метрик
	metrics := ws.getMetrics()

	// Формирование сообщения
	message := MetricMessage{
		Type:      "metrics_update",
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      metrics,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		ws.logger.Warnf("Ошибка marshaling сообщения: %v", err)
		return
	}

	// Отправка всем клиентам
	for client := range ws.clients {
		err := client.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			ws.logger.Warnf("Ошибка отправки клиенту: %v", err)
			client.Close()
			delete(ws.clients, client)
		}
	}
}

// getMetrics - получение метрик для отправки
func (ws *WebSocketServer) getMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Получение системных метрик
	if ws.metricsCollector != nil {
		if systemMetrics, err := ws.metricsCollector.GetRAMUsage(); err == nil {
			metrics["ram"] = systemMetrics
		}
		if diskMetrics, err := ws.metricsCollector.GetDiskUsage("."); err == nil {
			metrics["disk"] = diskMetrics
		}
		metrics["cpu_usage"] = ws.metricsCollector.GetCPUUsage()
	}

	return metrics
}

// sendMetrics - отправка метрик конкретному клиенту
func (ws *WebSocketServer) sendMetrics(conn *websocket.Conn) {
	metrics := ws.getMetrics()

	message := MetricMessage{
		Type:      "metrics",
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      metrics,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		ws.logger.Warnf("Ошибка marshaling сообщения: %v", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		ws.logger.Warnf("Ошибка отправки клиенту: %v", err)
	}
}

// SendToClient - отправка сообщения конкретному клиенту
func (ws *WebSocketServer) SendToClient(conn *websocket.Conn, message MetricMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, jsonData)
}

// SendToAllClients - отправка сообщения всем клиентам
func (ws *WebSocketServer) SendToAllClients(message MetricMessage) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	jsonData, err := json.Marshal(message)
	if err != nil {
		ws.logger.Warnf("Ошибка marshaling сообщения: %v", err)
		return
	}

	for client := range ws.clients {
		err := client.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			ws.logger.Warnf("Ошибка отправки клиенту: %v", err)
			client.Close()
			delete(ws.clients, client)
		}
	}
}
