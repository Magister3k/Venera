// network.go - Сетевые метрики для системы Venera
//
// Этот модуль обеспечивает сбор сетевых метрик для приложения Venera,
// включая информацию о сетевых интерфейсах, передаче данных,
// ошибках и других параметрах сети.
//
// Основные функции:
// - Получение информации о сетевых интерфейсах
// - Сбор метрик передачи данных (байты, пакеты)
// - Сбор метрик ошибок
// - Получение IP-адресов и сетевых настроек
//
// Использование:
// import "venera/metrics"
// networkMetrics, _ := GetNetworkMetrics()
// for _, metric := range networkMetrics {
//     fmt.Printf("Interface: %s, IP: %s\n", metric.InterfaceName, metric.IP)
// }

package metrics

import (
	"fmt"
	"net"
)

// NetworkMetrics - сетевые метрики
type NetworkMetrics struct {
	InterfaceName   string  `json:"interface_name"`
	IP              string  `json:"ip"`
	MacAddress      string  `json:"mac_address"`
	BytesSent       uint64  `json:"bytes_sent"`
	BytesReceived   uint64  `json:"bytes_received"`
	PacketsSent     uint64  `json:"packets_sent"`
	PacketsRecv     uint64  `json:"packets_recv"`
	ErrorsSent      uint64  `json:"errors_sent"`
	ErrorsRecv      uint64  `json:"errors_recv"`
	DroppedSent     uint64  `json:"dropped_sent"`
	DroppedRecv     uint64  `json:"dropped_recv"`
	Status          string  `json:"status"`
}

// GetNetworkMetrics - получение сетевых метрик
func GetNetworkMetrics() ([]NetworkMetrics, error) {
	metrics := make([]NetworkMetrics, 0)

	// Получение всех сетевых интерфейсов
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сетевых интерфейсов: %w", err)
	}

	for _, iface := range interfaces {
		metric := NetworkMetrics{
			InterfaceName: iface.Name,
			Status:        getStatus(iface.Flags),
		}

		// Получение MAC-адреса
		metric.MacAddress = iface.HardwareAddr.String()

		// Получение IP-адресов
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					metric.IP = ipNet.IP.String()
					break
				}
			}
		}

		// TODO: Получение статистики интерфейса (bytes, packets, errors)
		// Для Windows можно использовать Win32_PerfRawData_Tcpip_NetworkInterface

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetInterfaceStats - получение статистики конкретного интерфейса
func GetInterfaceStats(interfaceName string) (*NetworkMetrics, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сетевых интерфейсов: %w", err)
	}

	for _, iface := range interfaces {
		if iface.Name == interfaceName {
			metric := NetworkMetrics{
				InterfaceName: iface.Name,
				Status:        getStatus(iface.Flags),
				MacAddress:    iface.HardwareAddr.String(),
			}

			// Получение IP-адресов
			addrs, err := iface.Addrs()
			if err == nil {
				for _, addr := range addrs {
					if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
						metric.IP = ipNet.IP.String()
						break
					}
				}
			}

			// TODO: Получение статистики интерфейса
			// Для Windows можно использовать Win32_PerfRawData_Tcpip_NetworkInterface

			return &metric, nil
		}
	}

	return nil, fmt.Errorf("интерфейс %s не найден", interfaceName)
}

// GetLocalIP - получение локального IP-адреса
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}

	return "", fmt.Errorf("не удалось определить локальный IP-адрес")
}

// GetNetworkInterfaces - получение списка сетевых интерфейсов
func GetNetworkInterfaces() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сетевых интерфейсов: %w", err)
	}

	result := make([]string, 0)
	for _, iface := range interfaces {
		result = append(result, iface.Name)
	}

	return result, nil
}

// GetNetworkInterfacesWithIP - получение списка сетевых интерфейсов с IP-адресами
func GetNetworkInterfacesWithIP() (map[string]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сетевых интерфейсов: %w", err)
	}

	result := make(map[string]string)
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
					result[iface.Name] = ipNet.IP.String()
					break
				}
			}
		}
	}

	return result, nil
}

// getStatus - получение статуса интерфейса по флагам
func getStatus(flags net.Flags) string {
	if flags&net.FlagUp != 0 {
		return "up"
	}
	return "down"
}

// GetNetworkUsage - получение использования сети
func GetNetworkUsage() (uint64, uint64, error) {
	var bytesSent, bytesReceived uint64

	// TODO: Получение статистики использования сети
	// Для Windows можно использовать Win32_PerfRawData_Tcpip_NetworkInterface

	return bytesSent, bytesReceived, nil
}

// GetConnectionStats - получение статистики соединений
func GetConnectionStats() (map[string]int, error) {
	stats := make(map[string]int)

	// TODO: Получение статистики соединений
	// Для Windows можно использовать Win32_PerfRawData_Tcpip_TCPv4 и TCPv6

	stats["tcp_connections"] = 0
	stats["udp_connections"] = 0
	stats["established"] = 0
	stats["listening"] = 0

	return stats, nil
}
