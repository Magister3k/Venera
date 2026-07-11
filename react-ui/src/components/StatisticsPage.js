/**
 * StatisticsPage.js - Страница статистики
 *
 * Этот компонент обеспечивает:
 * - Отображение статистики в реальном времени
 * - Подключение к веб-сокету для обновления данных
 * - Отображение системных метрик
 * - Метрики по процессам
 *
 * Основные функции:
 * - WebSocket-подключение для получения метрик
 * - Обновление данных в реальном времени
 * - Отображение графиков и таблиц
 * - Фильтрация и сортировка данных
 *
 * Использование:
 * Импортируется в App.js
 */

import React, { useState, useEffect, useRef } from 'react';
import { io } from 'socket.io-client';

const StatisticsPage = () => {
  const [metrics, setMetrics] = useState({
    system: {
      cpuUsage: 0,
      ramUsage: 0,
      diskUsage: 0,
      processesCount: 0
    },
    processes: {},
    network: {
      packetsReceived: 0,
      packetsDropped: 0,
      bytesReceived: 0
    }
  });
  const [wsConnected, setWsConnected] = useState(false);
  const socketRef = useRef(null);

  // Подключение к веб-сокету при монтировании
  useEffect(() => {
    socketRef.current = io('/metrics');

    socketRef.current.on('connect', () => {
      setWsConnected(true);
    });

    socketRef.current.on('disconnect', () => {
      setWsConnected(false);
    });

    socketRef.current.on('metrics', (data) => {
      setMetrics(data);
    });

    socketRef.current.on('processMetrics', (data) => {
      setMetrics(prev => ({
        ...prev,
        processes: data
      }));
    });

    // Очистка при размонтировании
    return () => {
      if (socketRef.current) {
        socketRef.current.disconnect();
      }
    };
  }, []);

  // Форматирование процентов
  const formatPercent = (value) => {
    return `${(value * 100).toFixed(1)}%`;
  };

  // Форматирование размера в байтах
  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // Форматирование количества пакетов
  const formatCount = (count) => {
    if (count >= 1000000) return `${(count / 1000000).toFixed(1)} M`;
    if (count >= 1000) return `${(count / 1000).toFixed(1)} K`;
    return count.toString();
  };

  return (
    <div className="statistics-page">
      <div className="page-header">
        <h2>Статистика в реальном времени</h2>
        <div className="connection-status">
          <span className={`status-dot ${wsConnected ? 'connected' : 'disconnected'}`}></span>
          {wsConnected ? 'Подключено к серверу' : 'Отключено от сервера'}
        </div>
      </div>

      {/* Системные метрики */}
      <div className="card">
        <div className="card-header">
          <h3>Системные метрики</h3>
        </div>
        <div className="metrics-grid">
          <div className="metric-card">
            <h4>Загрузка CPU</h4>
            <div className="metric-value">{formatPercent(metrics.system.cpuUsage)}</div>
            <div className="metric-progress">
              <div
                className="metric-bar"
                style={{ width: `${metrics.system.cpuUsage * 100}%` }}
              ></div>
            </div>
          </div>
          <div className="metric-card">
            <h4>Использование RAM</h4>
            <div className="metric-value">{formatPercent(metrics.system.ramUsage)}</div>
            <div className="metric-progress">
              <div
                className="metric-bar"
                style={{ width: `${metrics.system.ramUsage * 100}%` }}
              ></div>
            </div>
          </div>
          <div className="metric-card">
            <h4>Использование диска</h4>
            <div className="metric-value">{formatPercent(metrics.system.diskUsage)}</div>
            <div className="metric-progress">
              <div
                className="metric-bar"
                style={{ width: `${metrics.system.diskUsage * 100}%` }}
              ></div>
            </div>
          </div>
          <div className="metric-card">
            <h4>Активных процессов</h4>
            <div className="metric-value">{metrics.system.processesCount}</div>
          </div>
        </div>
      </div>

      {/* Метрики сети */}
      <div className="card">
        <div className="card-header">
          <h3>Метрики сети</h3>
        </div>
        <div className="metrics-grid">
          <div className="metric-card">
            <h4>Пакетов получено</h4>
            <div className="metric-value">{formatCount(metrics.network.packetsReceived)}</div>
          </div>
          <div className="metric-card">
            <h4>Пакетов упущено</h4>
            <div className="metric-value">{formatCount(metrics.network.packetsDropped)}</div>
          </div>
          <div className="metric-card">
            <h4>Байт получено</h4>
            <div className="metric-value">{formatBytes(metrics.network.bytesReceived)}</div>
          </div>
        </div>
      </div>

      {/* Метрики процессов */}
      <div className="card">
        <div className="card-header">
          <h3>Метрики по процессам</h3>
        </div>
        {Object.keys(metrics.processes).length === 0 ? (
          <div className="empty-state">
            <p>Нет активных процессов для отображения метрик</p>
          </div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>ID процесса</th>
                  <th>Скорость входного потока</th>
                  <th>Потреб��ение RAM</th>
                  <th>Загрузка CPU</th>
                  <th>Статус</th>
                </tr>
              </thead>
              <tbody>
                {Object.entries(metrics.processes).map(([id, process]) => (
                  <tr key={id}>
                    <td>{id}</td>
                    <td>{formatCount(process.packetsPerSecond || 0)} пакетов/сек</td>
                    <td>{formatBytes(process.ramUsage || 0)}</td>
                    <td>{formatPercent(process.cpuUsage || 0)}</td>
                    <td>{process.status || 'unknown'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Экспорт статистики */}
      <div className="card">
        <div className="card-header">
          <h3>Экспорт статистики</h3>
        </div>
        <div className="export-actions">
          <button className="btn btn-primary">Экспорт в CSV</button>
          <button className="btn btn-secondary">Экспорт в JSON</button>
          <button className="btn btn-outline">Экспорт в PDF</button>
        </div>
      </div>
    </div>
  );
};

export default StatisticsPage;
