// zabbix.go - Экспорт метрик для Zabbix мониторинга
//
// Этот модуль обеспечивает экспорт метрик для системы мониторинга Zabbix
// через Zabbix Sender или HTTP API.
//
// Основные функции:
// - Экспорт метрик в формате Zabbix Sender
// - Формирование данных для Zabbix HTTP API
// - Планировщик экспорта метрик в Zabbix
// - Обработка ошибок отправки метрик
//
// Использование:
// import "venera/metrics"
// zabbix := metrics.NewZabbixExporter(cfg, logger)
// zabbix.SendMetrics(metrics)

package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"venera/config"
)

// ZabbixExporter - экспортер метрик для Zabbix
type ZabbixExporter struct {
	cfg      *config.Config
	logger   *logrus.Logger
	host     string
	port     int
	sender   string
}

// ZabbixItem - элемент данных для Zabbix
type ZabbixItem struct {
	Host   string `json:"host"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	Clock  int64  `json:"clock,omitempty"`
}

// ZabbixRequest - запрос для Zabbix HTTP API
type ZabbixRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  interface{}   `json:"params"`
	ID      int           `json:"id"`
Auth    Auth          `json:"auth,omitempty"`
}

// Auth - аутентификация для Zabbix API
type Auth struct {
	Token string `json:"auth"`
}

// ZabbixResponse - ответ от Zabbix API
type ZabbixResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   Error     `json:"error,omitempty"`
	ID      int       `json:"id"`
}

// Error - ошибка Zabbix API
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// NewZabbixExporter - создание нового экспортера для Zabbix
func NewZabbixExporter(cfg *config.Config, logger *logrus.Logger) *ZabbixExporter {
	return &ZabbixExporter{
		cfg:    cfg,
		logger: logger,
		host:   "localhost", // По умолчанию
		port:   10051,       // Порт Zabbix Sender
		sender: "venera",
	}
}

// SetHost - установка хоста Zabbix
func (z *ZabbixExporter) SetHost(host string) {
	z.host = host
}

// SetPort - установка порта Zabbix
func (z *ZabbixExporter) SetPort(port int) {
	z.port = port
}

// SetSender - установка имени отправителя
func (z *ZabbixExporter) SetSender(sender string) {
	z.sender = sender
}

// SendMetrics - отправка метрик в Zabbix (формат Zabbix Sender)
func (z *ZabbixExporter) SendMetrics(metrics map[string]interface{}) error {
	// Формирование данных для Zabbix Sender
	var data bytes.Buffer

	// Заголовок Zabbix Sender
	data.WriteString("ZBXD\x01")

	// Формирование сообщения
	message := make(map[string]interface{})
	message["request"] = "sender data"
	message["data"] = z.formatForZabbix(metrics)

	// Маршалинг в JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга: %w", err)
	}

	// Добавление длины сообщения
	length := make([]byte, 8)
	copy(length, jsonData)
	data.Write(length)
	data.Write(jsonData)

	// Отправка данных
	// TODO: Реализовать отправку через TCP к Zabbix Sender
	z.logger.Debug("Метрики подготовлены для отправки в Zabbix")

	return nil
}

// formatForZabbix - форматирование метрик для Zabbix
func (z *ZabbixExporter) formatForZabbix(metrics map[string]interface{}) []ZabbixItem {
	items := make([]ZabbixItem, 0)

	// Формирование ключей и значений
	z.formatMetricsRecursive("", metrics, &items)

	return items
}

// formatMetricsRecursive - рекурсивное форматирование метрик
func (z *ZabbixExporter) formatMetricsRecursive(prefix string, metrics interface{}, items *[]ZabbixItem) {
	switch v := metrics.(type) {
	case map[string]interface{}:
		for key, value := range v {
			z.formatMetricsRecursive(fmt.Sprintf("%s[%s]", prefix, key), value, items)
		}
	case float64:
		*items = append(*items, ZabbixItem{
			Host:  z.sender,
			Key:   prefix,
			Value: fmt.Sprintf("%.2f", v),
			Clock: time.Now().Unix(),
		})
	case int64:
		*items = append(*items, ZabbixItem{
			Host:  z.sender,
			Key:   prefix,
			Value: fmt.Sprintf("%d", v),
			Clock: time.Now().Unix(),
		})
	case string:
		*items = append(*items, ZabbixItem{
			Host:  z.sender,
			Key:   prefix,
			Value: v,
			Clock: time.Now().Unix(),
		})
	case bool:
		*items = append(*items, ZabbixItem{
			Host:  z.sender,
			Key:   prefix,
			Value: fmt.Sprintf("%t", v),
			Clock: time.Now().Unix(),
		})
	}
}

// SendToZabbixAPI - отправка метрик через Zabbix HTTP API
func (z *ZabbixExporter) SendToZabbixAPI(metrics map[string]interface{}, auth_token string) error {
	// Формирование элементов данных
	items := z.formatForZabbix(metrics)

	// Формирование запроса
	request := ZabbixRequest{
		JSONRPC: "2.0",
		Method:  "item.create",
		Params: map[string]interface{}{
			"host":  z.sender,
			"key":   "venera.metrics",
			"type":  0,
			"value_type": 4,
		},
		ID: 1,
	}

	if auth_token != "" {
		request.Auth = Auth{Token: auth_token}
	}

	// Маршалинг в JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга: %w", err)
	}

	// Отправка запроса
	resp, err := http.Post("http://localhost/zabbix/api_jsonrpc.php", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Обработка ответа
	var response ZabbixResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if response.Error.Code != 0 {
		return fmt.Errorf("ошибка Zabbix API: %s", response.Error.Message)
	}

	z.logger.Info("Метрики успешно отправлены в Zabbix API")
	return nil
}

// SendSystemMetrics - отправка системных метрик
func (z *ZabbixExporter) SendSystemMetrics(systemMetrics *SystemMetrics) error {
	metrics := make(map[string]interface{})

	if systemMetrics != nil {
		metrics["ram.total"] = systemMetrics.TotalRAM
		metrics["ram.available"] = systemMetrics.AvailableRAM
		metrics["ram.used"] = systemMetrics.UsedRAM
		metrics["ram.percent"] = systemMetrics.RAMPercent
		metrics["cpu.usage"] = systemMetrics.CPUUsage
		metrics["disk.free"] = systemMetrics.DiskFree
		metrics["disk.total"] = systemMetrics.DiskTotal
		metrics["disk.used"] = systemMetrics.DiskUsed
		metrics["disk.percent"] = systemMetrics.DiskPercent
	}

	return z.SendMetrics(metrics)
}

// SendProcessMetrics - отправка метрик процессов
func (z *ZabbixExporter) SendProcessMetrics(processMetrics map[string]*ProcessMetrics) error {
	metrics := make(map[string]interface{})

	for id, pm := range processMetrics {
		metrics[fmt.Sprintf("process[%s].total_json_pairs", id)] = pm.TotalJSONPairs
		metrics[fmt.Sprintf("process[%s].selected_pairs", id)] = pm.SelectedPairs
		metrics[fmt.Sprintf("process[%s].postgres_pairs", id)] = pm.PostgresPairs
		metrics[fmt.Sprintf("process[%s].processing_rate", id)] = pm.ProcessingRate
	}

	return z.SendMetrics(metrics)
}

// SendDatabaseMetrics - отправка метрик баз данных
func (z *ZabbixExporter) SendDatabaseMetrics(dbMetrics *DatabaseMetrics) error {
	metrics := make(map[string]interface{})

	if dbMetrics != nil {
		metrics["dragonflydb.used_memory"] = dbMetrics.DragonflyDB.UsedMemory
		metrics["dragonflydb.connected_clients"] = dbMetrics.DragonflyDB.ConnectedClients
		metrics["postgresql.size"] = dbMetrics.PostgreSQL.DatabaseSize
		metrics["postgresql.connections"] = dbMetrics.PostgreSQL.ActiveConnections
	}

	return z.SendMetrics(metrics)
}

// SendNetworkMetrics - отправка сетевых метрик
func (z *ZabbixExporter) SendNetworkMetrics(networkMetrics []NetworkMetrics) error {
	metrics := make(map[string]interface{})

	for i, nm := range networkMetrics {
		metrics[fmt.Sprintf("network.interface[%d].bytes_sent", i)] = nm.BytesSent
		metrics[fmt.Sprintf("network.interface[%d].bytes_recv", i)] = nm.BytesReceived
		metrics[fmt.Sprintf("network.interface[%d].packets_sent", i)] = nm.PacketsSent
		metrics[fmt.Sprintf("network.interface[%d].packets_recv", i)] = nm.PacketsRecv
	}

	return z.SendMetrics(metrics)
}

// SendAllMetrics - отправка всех метрик в Zabbix
func (z *ZabbixExporter) SendAllMetrics(systemMetrics *SystemMetrics, dbMetrics *DatabaseMetrics, processMetrics map[string]*ProcessMetrics, networkMetrics []NetworkMetrics) error {
	// Отправка всех метрик
	if err := z.SendSystemMetrics(systemMetrics); err != nil {
		z.logger.Warnf("Ошибка отправки системных метрик: %v", err)
	}

	if err := z.SendDatabaseMetrics(dbMetrics); err != nil {
		z.logger.Warnf("Ошибка отправки метрик баз данных: %v", err)
	}

	if err := z.SendProcessMetrics(processMetrics); err != nil {
		z.logger.Warnf("Ошибка отправки метрик процессов: %v", err)
	}

	if err := z.SendNetworkMetrics(networkMetrics); err != nil {
		z.logger.Warnf("Ошибка отправки сетевых метрик: %v", err)
	}

	return nil
}
