/**
 * ProcessesPage.js - Страница управления процессами
 *
 * Этот компонент обеспечивает:
 * - Отображение списка процессов обработки
 * - Добавление новых процессов
 * - Управление процессами (старт, стоп, удаление)
 * - Сохранение конфигурации процессов
 *
 * Основные функции:
 * - AJAX-запросы к веб-серверу для управления процессами
 * - Форма добавления процесса
 * - Таблица процессов с кнопками управления
 * - Кнопки для управления всеми процессами
 *
 * Использование:
 * Импортируется в App.js
 */

import React, { useState, useEffect } from 'react';
import axios from 'axios';

const ProcessesPage = () => {
  const [processes, setProcesses] = useState([]);
  const [newProcess, setNewProcess] = useState({
    type: 'network',
    name: '',
    ip: '',
    port: 5000,
    path: '',
    scanSubfolders: false,
    monitorNewFiles: false
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Загрузка процессов при монтировании
  useEffect(() => {
    loadProcesses();
  }, []);

  // Загрузка процессов
  const loadProcesses = async () => {
    try {
      const response = await axios.get('/api/processes');
      setProcesses(response.data);
      setError(null);
    } catch (err) {
      setError('Ошибка загрузки процессов: ' + (err.response?.data?.message || err.message));
    }
  };

  // Добавление процесса
  const addProcess = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await axios.post('/api/processes', newProcess);
      setNewProcess({
        type: 'network',
        name: '',
        ip: '',
        port: 5000,
        path: '',
        scanSubfolders: false,
        monitorNewFiles: false
      });
      await loadProcesses();
    } catch (err) {
      setError('Ошибка добавления процесса: ' + (err.response?.data?.message || err.message));
    } finally {
      setLoading(false);
    }
  };

  // Запуск процесса
  const startProcess = async (id) => {
    try {
      await axios.post(`/api/processes/${id}/start`);
      await loadProcesses();
    } catch (err) {
      setError('Ошибка запуска процесса: ' + (err.response?.data?.message || err.message));
    }
  };

  // Остановка процесса
  const stopProcess = async (id) => {
    try {
      await axios.post(`/api/processes/${id}/stop`);
      await loadProcesses();
    } catch (err) {
      setError('Ошибка остановки процесса: ' + (err.response?.data?.message || err.message));
    }
  };

  // Удаление процесса
  const deleteProcess = async (id) => {
    if (!window.confirm('Вы уверены, что хотите удалить процесс?')) {
      return;
    }
    try {
      await axios.delete(`/api/processes/${id}`);
      await loadProcesses();
    } catch (err) {
      setError('Ошибка удаления процесса: ' + (err.response?.data?.message || err.message));
    }
  };

  // Запуск всех процессов
  const startAll = async () => {
    try {
      await axios.post('/api/processes/start-all');
      await loadProcesses();
    } catch (err) {
      setError('Ошибка запуска всех процессов: ' + (err.response?.data?.message || err.message));
    }
  };

  // Остановка всех процессов
  const stopAll = async () => {
    try {
      await axios.post('/api/processes/stop-all');
      await loadProcesses();
    } catch (err) {
      setError('Ошибка остановки всех процессов: ' + (err.response?.data?.message || err.message));
    }
  };

  // Сохранение конфигурации процессов
  const saveConfig = async () => {
    try {
      await axios.post('/api/processes/config', processes);
      setError(null);
      alert('Конфигурация сохранена');
    } catch (err) {
      setError('Ошибка сохранения конфигурации: ' + (err.response?.data?.message || err.message));
    }
  };

  // Форматирование типа источника
  const formatType = (type) => {
    const types = {
      network: 'Сетевой поток',
      folder: 'Папка с файлами',
      file: 'Отдельный файл'
    };
    return types[type] || type;
  };

  // Форматирование статуса
  const formatStatus = (status) => {
    const statuses = {
      running: 'Запущен',
      stopped: 'Остановлен',
      error: 'Ошибка'
    };
    return statuses[status] || status;
  };

  return (
    <div className="processes-page">
      <div className="page-header">
        <h2>Управление процессами обработки</h2>
        <div className="page-actions">
          <button className="btn btn-primary" onClick={startAll}>Запустить все</button>
          <button className="btn btn-warning" onClick={stopAll}>Остановить все</button>
          <button className="btn btn-secondary" onClick={saveConfig}>Сохранить конфигурацию</button>
        </div>
      </div>

      {/* Форма добавления процесса */}
      <div className="card">
        <div className="card-header">
          <h3>Добавить новый процесс</h3>
        </div>
        <form onSubmit={addProcess}>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label">Тип источника</label>
              <select
                className="form-input"
                value={newProcess.type}
                onChange={(e) => setNewProcess({ ...newProcess, type: e.target.value })}
              >
                <option value="network">Сетевой поток</option>
                <option value="folder">Папка с файлами</option>
                <option value="file">Отдельный файл</option>
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Название</label>
              <input
                type="text"
                className="form-input"
                value={newProcess.name}
                onChange={(e) => setNewProcess({ ...newProcess, name: e.target.value })}
                placeholder="Введите название источника"
                required
              />
            </div>
          </div>

          {/* Поля для сетевого источника */}
          {newProcess.type === 'network' && (
            <div className="form-row network-fields">
              <div className="form-group">
                <label className="form-label">IP-адрес</label>
                <input
                  type="text"
                  className="form-input"
                  value={newProcess.ip}
                  onChange={(e) => setNewProcess({ ...newProcess, ip: e.target.value })}
                  placeholder="192.168.1.100"
                />
              </div>
              <div className="form-group">
                <label className="form-label">UDP-порт</label>
                <input
                  type="number"
                  className="form-input"
                  value={newProcess.port}
                  onChange={(e) => setNewProcess({ ...newProcess, port: parseInt(e.target.value) || 5000 })}
                  min="1"
                  max="65535"
                />
              </div>
            </div>
          )}

          {/* Поля для папки */}
          {newProcess.type === 'folder' && (
            <div className="form-row folder-fields">
              <div className="form-group">
                <label className="form-label">Путь к папке</label>
                <input
                  type="text"
                  className="form-input"
                  value={newProcess.path}
                  onChange={(e) => setNewProcess({ ...newProcess, path: e.target.value })}
                  placeholder="C:\Logs\Server"
                  required
                />
              </div>
              <div className="form-group checkbox-group">
                <label className="form-checkbox">
                  <input
                    type="checkbox"
                    checked={newProcess.scanSubfolders}
                    onChange={(e) => setNewProcess({ ...newProcess, scanSubfolders: e.target.checked })}
                  />
                  Сканировать подпапки
                </label>
                <label className="form-checkbox">
                  <input
                    type="checkbox"
                    checked={newProcess.monitorNewFiles}
                    onChange={(e) => setNewProcess({ ...newProcess, monitorNewFiles: e.target.checked })}
                  />
                  Мониторинг новых файлов
                </label>
              </div>
            </div>
          )}

          {/* Поля для отдельного файла */}
          {newProcess.type === 'file' && (
            <div className="form-row file-fields">
              <div className="form-group">
                <label className="form-label">Путь к файлу</label>
                <input
                  type="text"
                  className="form-input"
                  value={newProcess.path}
                  onChange={(e) => setNewProcess({ ...newProcess, path: e.target.value })}
                  placeholder="C:\Data\traffic.json"
                  required
                />
              </div>
            </div>
          )}

          <div className="form-actions">
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? 'Добавление...' : 'Добавить процесс'}
            </button>
          </div>
        </form>
      </div>

      {/* Список процессов */}
      <div className="card">
        <div className="card-header">
          <h3>Список процессов</h3>
        </div>
        {error && <div className="alert alert-error">{error}</div>}
        {processes.length === 0 ? (
          <div className="empty-state">
            <p>Нет активных процессов. Добавьте новый процесс выше.</p>
          </div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Тип</th>
                  <th>Название</th>
                  <th>Статус</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {processes.map((process) => (
                  <tr key={process.id}>
                    <td>{process.id}</td>
                    <td>{formatType(process.type)}</td>
                    <td>{process.name}</td>
                    <td>{formatStatus(process.status)}</td>
                    <td className="actions">
                      {process.status === 'stopped' ? (
                        <button
                          className="btn btn-success btn-sm"
                          onClick={() => startProcess(process.id)}
                        >
                          Старт
                        </button>
                      ) : (
                        <button
                          className="btn btn-warning btn-sm"
                          onClick={() => stopProcess(process.id)}
                        >
                          Стоп
                        </button>
                      )}
                      <button
                        className="btn btn-danger btn-sm"
                        onClick={() => deleteProcess(process.id)}
                      >
                        Удалить
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default ProcessesPage;
