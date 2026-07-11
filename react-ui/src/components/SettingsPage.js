/**
 * SettingsPage.js - Страница настроек
 *
 * Этот компонент обеспечивает:
 * - Настройку параметров приложения
 * - Управление конфигурацией
 * - Проверку подключения к базам данных
 * - Сохранение настроек
 *
 * Основные функции:
 * - AJAX-запросы для получения и сохранения настроек
 * - Форма редактирования конфигурации
 * - Проверка подключения к DragonflyDB и PostgreSQL
 * - Сохранение настроек в config.toml
 *
 * Использование:
 * Импортируется в App.js
 */

import React, { useState, useEffect } from 'react';
import axios from 'axios';

const SettingsPage = () => {
  const [settings, setSettings] = useState({
    generic: {
      mode: 'tray',
      autoStartProcesses: false,
      maxProcesses: 10,
      webServerPort: 8080,
      diskWarningPercent: 90,
      ramWarningPercent: 90,
      maxQueueSize: 1000,
      processingInterval: 30
    },
    paths: {
      podmanPath: '',
      tsharkPath: '',
      filterFile: '',
      controlFile: '',
      alertsFile: '',
      dragonflyImage: '',
      dragonflyBackupPath: ''
    },
    dragonflyDB: {
      host: 'localhost',
      port: 6379,
      password: '',
      database: 0,
      batch_size: 1000,
      timeout: 30
    },
    postgresql: {
      host: 'localhost',
      port: 5432,
      database: 'Venera',
      user: '',
      password: '',
      maxConnections: 20
    },
    logging: {
      level: 'info',
      directory: 'Logs',
      maxAgeDays: 30,
      maxFiles: 10,
      compressOldLogs: true
    }
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [connectionStatus, setConnectionStatus] = useState({
    dragonfly: null,
    postgresql: null
  });

  // Загрузка настроек при монтировании
  useEffect(() => {
    loadSettings();
  }, []);

  // Загрузка настроек
  const loadSettings = async () => {
    try {
      const response = await axios.get('/api/settings');
      setSettings(response.data);
      setError(null);
    } catch (err) {
      setError('Ошибка загрузки настроек: ' + (err.response?.data?.message || err.message));
    }
  };

  // Обработка изменения поля
  const handleFieldChange = (section, field, value) => {
    setSettings(prev => ({
      ...prev,
      [section]: {
        ...prev[section],
        [field]: value
      }
    }));
  };

  // Проверка подключения к DragonflyDB
  const checkDragonflyConnection = async () => {
    try {
      const response = await axios.post('/api/settings/test-dragonfly', settings.dragonflyDB);
      setConnectionStatus(prev => ({
        ...prev,
        dragonfly: { success: true, message: response.data.message }
      }));
    } catch (err) {
      setConnectionStatus(prev => ({
        ...prev,
        dragonfly: { success: false, message: err.response?.data?.message || 'Ошибка подключения' }
      }));
    }
  };

  // Проверка подключения к PostgreSQL
  const checkPostgresqlConnection = async () => {
    try {
      const response = await axios.post('/api/settings/test-postgresql', settings.postgresql);
      setConnectionStatus(prev => ({
        ...prev,
        postgresql: { success: true, message: response.data.message }
      }));
    } catch (err) {
      setConnectionStatus(prev => ({
        ...prev,
        postgresql: { success: false, message: err.response?.data?.message || 'Ошибка подключения' }
      }));
    }
  };

  // Сохранение настроек
  const saveSettings = async () => {
    setLoading(true);
    setError(null);

    try {
      await axios.post('/api/settings/save', settings);
      setError(null);
      alert('Настройки сохранены успешно');
    } catch (err) {
      setError('Ошибка сохранения настроек: ' + (err.response?.data?.message || err.message));
    } finally {
      setLoading(false);
    }
  };

  // Форматирование пути
  const formatPath = (path) => {
    return path || '(не указан)';
  };

  return (
    <div className="settings-page">
      <div className="page-header">
        <h2>Настройки приложения</h2>
        <div className="page-actions">
          <button className="btn btn-primary" onClick={saveSettings} disabled={loading}>
            {loading ? 'Сохранение...' : 'Сохранить настройки'}
          </button>
        </div>
      </div>

      {/* Проверка подключения */}
      <div className="card">
        <div className="card-header">
          <h3>Проверка подключения к базам данных</h3>
        </div>
        <div className="connection-check">
          <div className="connection-card">
            <h4>DragonflyDB</h4>
            <div className="connection-status">
              {connectionStatus.dragonfly ? (
                connectionStatus.dragonfly.success ? (
                  <span className="status-success">✅ Подключено</span>
                ) : (
                  <span className="status-error">❌ Ошибка: {connectionStatus.dragonfly.message}</span>
                )
              ) : (
                <span className="status-default">Не проверено</span>
              )}
            </div>
            <button className="btn btn-secondary" onClick={checkDragonflyConnection}>
              Проверить подключение
            </button>
          </div>
          <div className="connection-card">
            <h4>PostgreSQL</h4>
            <div className="connection-status">
              {connectionStatus.postgresql ? (
                connectionStatus.postgresql.success ? (
                  <span className="status-success">✅ Подключено</span>
                ) : (
                  <span className="status-error">❌ Ошибка: {connectionStatus.postgresql.message}</span>
                )
              ) : (
                <span className="status-default">Не проверено</span>
              )}
            </div>
            <button className="btn btn-secondary" onClick={checkPostgresqlConnection}>
              Проверить подключение
            </button>
          </div>
        </div>
      </div>

      {/* Общие настройки */}
      <div className="card">
        <div className="card-header">
          <h3>Общие настройки</h3>
        </div>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Режим работы</label>
            <select
              className="form-input"
              value={settings.generic.mode}
              onChange={(e) => handleFieldChange('generic', 'mode', e.target.value)}
            >
              <option value="tray">Системный трей</option>
              <option value="service">Служба Windows</option>
            </select>
          </div>
          <div className="form-group">
            <label className="form-label">Автозапуск процессов</label>
            <select
              className="form-input"
              value={settings.generic.autoStartProcesses ? 'true' : 'false'}
              onChange={(e) => handleFieldChange('generic', 'autoStartProcesses', e.target.value === 'true')}
            >
              <option value="true">Да</option>
              <option value="false">Нет</option>
            </select>
          </div>
          <div className="form-group">
            <label className="form-label">Макс. процессов (1-20)</label>
            <input
              type="number"
              className="form-input"
              value={settings.generic.maxProcesses}
              onChange={(e) => handleFieldChange('generic', 'maxProcesses', parseInt(e.target.value))}
              min="1"
              max="20"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Порт веб-сервера</label>
            <input
              type="number"
              className="form-input"
              value={settings.generic.webServerPort}
              onChange={(e) => handleFieldChange('generic', 'webServerPort', parseInt(e.target.value))}
              min="1"
              max="65535"
            />
          </div>
        </div>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Предупреждение о диске (%)</label>
            <input
              type="number"
              className="form-input"
              value={settings.generic.diskWarningPercent}
              onChange={(e) => handleFieldChange('generic', 'diskWarningPercent', parseInt(e.target.value))}
              min="0"
              max="100"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Предупреждение о RAM (%)</label>
            <input
              type="number"
              className="form-input"
              value={settings.generic.ramWarningPercent}
              onChange={(e) => handleFieldChange('generic', 'ramWarningPercent', parseInt(e.target.value))}
              min="0"
              max="100"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Макс. размер очереди</label>
            <input
              type="number"
              className="form-input"
              value={settings.generic.maxQueueSize}
              onChange={(e) => handleFieldChange('generic', 'maxQueueSize', parseInt(e.target.value))}
              min="1"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Интервал обработки (сек)</label>
            <input
              type="number"
              className="form-input"
              value={settings.generic.processingInterval}
              onChange={(e) => handleFieldChange('generic', 'processingInterval', parseInt(e.target.value))}
              min="1"
            />
          </div>
        </div>
      </div>

      {/* Пути */}
      <div className="card">
        <div className="card-header">
          <h3>Пути к ресурсам</h3>
        </div>
        <div className="form-group">
          <label className="form-label">Путь к podman.exe</label>
          <input
            type="text"
            className="form-input"
            value={settings.paths.podmanPath}
            onChange={(e) => handleFieldChange('paths', 'podmanPath', e.target.value)}
            placeholder="C:\Program Files\Podman\podman.exe"
          />
        </div>
        <div className="form-group">
          <label className="form-label">Путь к tshark.exe</label>
          <input
            type="text"
            className="form-input"
            value={settings.paths.tsharkPath}
            onChange={(e) => handleFieldChange('paths', 'tsharkPath', e.target.value)}
            placeholder="C:\Program Files\Wireshark\tshark.exe"
          />
        </div>
        <div className="form-group">
          <label className="form-label">Путь к generic.flt</label>
          <input
            type="text"
            className="form-input"
            value={settings.paths.filterFile}
            onChange={(e) => handleFieldChange('paths', 'filterFile', e.target.value)}
            placeholder="settings\generic.flt"
          />
        </div>
        <div className="form-group">
          <label className="form-label">Путь к generic.ctr</label>
          <input
            type="text"
            className="form-input"
            value={settings.paths.controlFile}
            onChange={(e) => handleFieldChange('paths', 'controlFile', e.target.value)}
            placeholder="settings\generic.ctr"
          />
        </div>
        <div className="form-group">
          <label className="form-label">Путь к generic.alr</label>
          <input
            type="text"
            className="form-input"
            value={settings.paths.alertsFile}
            onChange={(e) => handleFieldChange('paths', 'alertsFile', e.target.value)}
            placeholder="settings\generic.alr"
          />
        </div>
        <div className="form-group">
          <label className="form-label">Образ DragonflyDB</label>
          <input
            type="text"
            className="form-input"
            value={settings.paths.dragonflyImage}
            onChange={(e) => handleFieldChange('paths', 'dragonflyImage', e.target.value)}
            placeholder="docker.io/dragonflydb/dragonfly:latest"
          />
        </div>
        <div className="form-group">
          <label className="form-label">Резервная папка DragonflyDB</label>
          <input
            type="text"
            className="form-input"
            value={settings.paths.dragonflyBackupPath}
            onChange={(e) => handleFieldChange('paths', 'dragonflyBackupPath', e.target.value)}
            placeholder="C:\Venera\dragonfly_backup"
          />
        </div>
      </div>

      {/* DragonflyDB */}
      <div className="card">
        <div className="card-header">
          <h3>DragonflyDB</h3>
        </div>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Хост</label>
            <input
              type="text"
              className="form-input"
              value={settings.dragonflyDB.host}
              onChange={(e) => handleFieldChange('dragonflyDB', 'host', e.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="form-label">Порт</label>
            <input
              type="number"
              className="form-input"
              value={settings.dragonflyDB.port}
              onChange={(e) => handleFieldChange('dragonflyDB', 'port', parseInt(e.target.value))}
              min="1"
              max="65535"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Пароль</label>
            <input
              type="password"
              className="form-input"
              value={settings.dragonflyDB.password}
              onChange={(e) => handleFieldChange('dragonflyDB', 'password', e.target.value)}
              placeholder="••••••"
            />
          </div>
          <div className="form-group">
            <label className="form-label">База данных (0-15)</label>
            <input
              type="number"
              className="form-input"
              value={settings.dragonflyDB.database}
              onChange={(e) => handleFieldChange('dragonflyDB', 'database', parseInt(e.target.value))}
              min="0"
              max="15"
            />
          </div>
        </div>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Размер партии</label>
            <input
              type="number"
              className="form-input"
              value={settings.dragonflyDB.batch_size}
              onChange={(e) => handleFieldChange('dragonflyDB', 'batch_size', parseInt(e.target.value))}
              min="1"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Таймаут (сек)</label>
            <input
              type="number"
              className="form-input"
              value={settings.dragonflyDB.timeout}
              onChange={(e) => handleFieldChange('dragonflyDB', 'timeout', parseInt(e.target.value))}
              min="1"
            />
          </div>
        </div>
      </div>

      {/* PostgreSQL */}
      <div className="card">
        <div className="card-header">
          <h3>PostgreSQL</h3>
        </div>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Хост</label>
            <input
              type="text"
              className="form-input"
              value={settings.postgresql.host}
              onChange={(e) => handleFieldChange('postgresql', 'host', e.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="form-label">Порт</label>
            <input
              type="number"
              className="form-input"
              value={settings.postgresql.port}
              onChange={(e) => handleFieldChange('postgresql', 'port', parseInt(e.target.value))}
              min="1"
              max="65535"
            />
          </div>
          <div className="form-group">
            <label className="form-label">База данных</label>
            <input
              type="text"
              className="form-input"
              value={settings.postgresql.database}
              onChange={(e) => handleFieldChange('postgresql', 'database', e.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="form-label">Пользователь</label>
            <input
              type="text"
              className="form-input"
              value={settings.postgresql.user}
              onChange={(e) => handleFieldChange('postgresql', 'user', e.target.value)}
            />
          </div>
        </div>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Пароль</label>
            <input
              type="password"
              className="form-input"
              value={settings.postgresql.password}
              onChange={(e) => handleFieldChange('postgresql', 'password', e.target.value)}
              placeholder="••••••"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Макс. соединений</label>
            <input
              type="number"
              className="form-input"
              value={settings.postgresql.maxConnections}
              onChange={(e) => handleFieldChange('postgresql', 'maxConnections', parseInt(e.target.value))}
              min="1"
            />
          </div>
        </div>
      </div>

      {/* Логирование */}
      <div className="card">
        <div className="card-header">
          <h3>Логирование</h3>
        </div>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Уровень логирования</label>
            <select
              className="form-input"
              value={settings.logging.level}
              onChange={(e) => handleFieldChange('logging', 'level', e.target.value)}
            >
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warn</option>
              <option value="error">Error</option>
              <option value="fatal">Fatal</option>
              <option value="panic">Panic</option>
            </select>
          </div>
          <div className="form-group">
            <label className="form-label">Директория логов</label>
            <input
              type="text"
              className="form-input"
              value={settings.logging.directory}
              onChange={(e) => handleFieldChange('logging', 'directory', e.target.value)}
              placeholder="Logs"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Макс. возраст (дней)</label>
            <input
              type="number"
              className="form-input"
              value={settings.logging.maxAgeDays}
              onChange={(e) => handleFieldChange('logging', 'maxAgeDays', parseInt(e.target.value))}
              min="1"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Макс. файлов</label>
            <input
              type="number"
              className="form-input"
              value={settings.logging.maxFiles}
              onChange={(e) => handleFieldChange('logging', 'maxFiles', parseInt(e.target.value))}
              min="1"
            />
          </div>
        </div>
        <div className="form-group">
          <label className="form-checkbox">
            <input
              type="checkbox"
              checked={settings.logging.compressOldLogs}
              onChange={(e) => handleFieldChange('logging', 'compressOldLogs', e.target.checked)}
            />
            Сжимать старые логи
          </label>
        </div>
      </div>
    </div>
  );
};

export default SettingsPage;
