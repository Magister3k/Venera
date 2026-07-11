/**
 * LogsPage.js - Страница логов
 *
 * Этот компонент обеспечивает:
 * - Просмотр логов в реальном времени
 * - Фильтрацию и поиск сообщений
 * - Цветовую подсветку по типу сообщения
 * - Экспорт логов в HTML и PDF
 *
 * Основные функции:
 * - WebSocket-подключение для получения логов
 * - Фильтрация по уровню, тексту и диапазону дат
 * - Поиск с подсветкой
 * - Горячие клавиши (Ctrl+F, Esc)
 * - Сохранение фильтров в localStorage
 *
 * Использование:
 * Импортируется в App.js
 */

import React, { useState, useEffect, useRef } from 'react';
import { io } from 'socket.io-client';

const LogsPage = () => {
  const [logs, setLogs] = useState([]);
  const [filter, setFilter] = useState({
    level: '',
    text: '',
    startDate: '',
    endDate: ''
  });
  const [searchText, setSearchText] = useState('');
  const [showSearch, setShowSearch] = useState(false);
  const logsRef = useRef(null);
  const socketRef = useRef(null);
  const [selectedLevels, setSelectedLevels] = useState(new Set(['debug', 'info', 'warn', 'error']));

  // Подключение к веб-сокету при монтировании
  useEffect(() => {
    socketRef.current = io('/logs');

    socketRef.current.on('connect', () => {
      console.log('WebSocket connected');
    });

    socketRef.current.on('disconnect', () => {
      console.log('WebSocket disconnected');
    });

    socketRef.current.on('log', (log) => {
      setLogs(prev => [log, ...prev].slice(0, 1000));
    });

    return () => {
      if (socketRef.current) {
        socketRef.current.disconnect();
      }
    };
  }, []);

  // Автопрокрутка вниз при добавлении новых логов
  useEffect(() => {
    if (logsRef.current) {
      logsRef.current.scrollTop = logsRef.current.scrollHeight;
    }
  }, [logs]);

  // Загрузка истории логов
  const loadHistory = async () => {
    try {
      const response = await axios.get('/api/logs/history', {
        params: {
          ...filter,
          limit: 100
        }
      });
      setLogs(response.data.logs);
    } catch (err) {
      console.error('Ошибка загрузки истории:', err);
    }
  };

  // Применение фильтров
  const applyFilters = () => {
    loadHistory();
    // Сохранение фильтров в localStorage
    localStorage.setItem('logFilters', JSON.stringify(filter));
  };

  // Очистка поиска
  const clearSearch = () => {
    setSearchText('');
  };

  // Фильтрация логов
  const filteredLogs = logs.filter(log => {
    // Фильтр по уровню
    if (filter.level && log.level !== filter.level) {
      return false;
    }

    // Фильтр по тексту
    if (searchText && !log.message.toLowerCase().includes(searchText.toLowerCase())) {
      return false;
    }

    return true;
  });

  // Форматирование уровня
  const formatLevel = (level) => {
    const levels = {
      debug: 'Отладка',
      info: 'Информация',
      warn: 'Предупреждение',
      error: 'Ошибка',
      fatal: 'Критическая ошибка',
      panic: 'Паника'
    };
    return levels[level] || level;
  };

  // Цвет уровня
  const getLevelColor = (level) => {
    const colors = {
      debug: '#9e9e9e',
      info: '#2196f3',
      warn: '#ff9800',
      error: '#f44336',
      fatal: '#b71c1c',
      panic: '#880e4f'
    };
    return colors[level] || '#000000';
  };

  // Подсветка текста
  const highlightText = (text, searchTerm) => {
    if (!searchTerm) return text;
    const regex = new RegExp(`(${searchTerm})`, 'gi');
    const parts = text.split(regex);
    return parts.map((part, i) => {
      if (part.toLowerCase() === searchTerm.toLowerCase()) {
        return <span key={i} className="highlight">{part}</span>;
      }
      return part;
    });
  };

  // Экспорт в HTML
  const exportHTML = async () => {
    try {
      const response = await axios.get('/api/logs/export/html', {
        params: { ...filter, search: searchText },
        responseType: 'blob'
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'venera_logs.html');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (err) {
      console.error('Ошибка экспорта в HTML:', err);
    }
  };

  // Экспорт в PDF
  const exportPDF = async () => {
    try {
      const response = await axios.get('/api/logs/export/pdf', {
        params: { ...filter, search: searchText },
        responseType: 'blob'
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'venera_logs.pdf');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (err) {
      console.error('Ошибка экспорта в PDF:', err);
    }
  };

  // Обработка горячих клавиш
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.ctrlKey && e.key === 'f') {
        e.preventDefault();
        setShowSearch(true);
      }
      if (e.key === 'Escape') {
        setShowSearch(false);
        setSearchText('');
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Форматирование времени
  const formatTime = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleString('ru-RU');
  };

  return (
    <div className="logs-page">
      <div className="page-header">
        <h2>Логи приложения</h2>
        <div className="page-actions">
          <button className="btn btn-primary" onClick={exportHTML}>Экспорт в HTML</button>
          <button className="btn btn-secondary" onClick={exportPDF}>Экспорт в PDF</button>
        </div>
      </div>

      {/* Панель фильтров */}
      <div className="card">
        <div className="card-header">
          <h3>Фильтры</h3>
        </div>
        <div className="filter-bar">
          <div className="filter-group">
            <label className="filter-label">Уровень:</label>
            <select
              className="form-input"
              value={filter.level}
              onChange={(e) => setFilter(prev => ({ ...prev, level: e.target.value }))}
            >
              <option value="">Все уровни</option>
              <option value="debug">Отладка</option>
              <option value="info">Информация</option>
              <option value="warn">Предупреждение</option>
              <option value="error">Ошибка</option>
              <option value="fatal">Критическая</option>
              <option value="panic">Паника</option>
            </select>
          </div>
          <div className="filter-group">
            <label className="filter-label">Дата начала:</label>
            <input
              type="date"
              className="form-input"
              value={filter.startDate}
              onChange={(e) => setFilter(prev => ({ ...prev, startDate: e.target.value }))}
            />
          </div>
          <div className="filter-group">
            <label className="filter-label">Дата окончания:</label>
            <input
              type="date"
              className="form-input"
              value={filter.endDate}
              onChange={(e) => setFilter(prev => ({ ...prev, endDate: e.target.value }))}
            />
          </div>
          <button className="btn btn-primary" onClick={applyFilters}>Применить</button>
        </div>
      </div>

      {/* Поиск */}
      {showSearch && (
        <div className="card search-bar">
          <div className="search-group">
            <input
              type="text"
              className="form-input"
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              placeholder="Поиск в логах..."
              autoFocus
            />
            <button className="btn btn-secondary" onClick={clearSearch}>
              ✖
            </button>
          </div>
        </div>
      )}

      {/* Область логов */}
      <div className="card logs-container">
        <div className="card-header">
          <h3>Логи ({filteredLogs.length})</h3>
        </div>
        <div className="logs-content" ref={logsRef}>
          {filteredLogs.length === 0 ? (
            <div className="empty-state">
              <p>Нет логов, соответствующих фильтрам</p>
            </div>
          ) : (
            <div className="log-list">
              {filteredLogs.map((log, index) => (
                <div
                  key={index}
                  className="log-entry"
                  style={{ borderLeft: `4px solid ${getLevelColor(log.level)}` }}
                >
                  <div className="log-header">
                    <span className="log-time">{formatTime(log.timestamp)}</span>
                    <span className={`log-level log-level-${log.level}`}>
                      {formatLevel(log.level)}
                    </span>
                    <span className="log-module">{log.module || 'main'}</span>
                  </div>
                  <div className="log-message">
                    {highlightText(log.message, searchText)}
                  </div>
                  {log.fields && (
                    <div className="log-fields">
                      {Object.entries(log.fields).map(([key, value]) => (
                        <span key={key} className="log-field">
                          {key}: {value}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
        <div className="logs-footer">
          <span className="logs-info">
            Показано {filteredLogs.length} из {logs.length} записей
          </span>
          <div className="logs-controls">
            <button className="btn btn-secondary" onClick={() => setLogs([])}>Очистить</button>
          </div>
        </div>
      </div>

      {/* Подсказка */}
      <div className="card">
        <div className="card-header">
          <h3>Горячие клавиши</h3>
        </div>
        <div className="hotkeys">
          <div className="hotkey-item">
            <span className="hotkey">Ctrl+F</span>
            <span className="hotkey-desc">Показать/скрыть поиск</span>
          </div>
          <div className="hotkey-item">
            <span className="hotkey">Esc</span>
            <span className="hotkey-desc">Очистить поиск</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LogsPage;
