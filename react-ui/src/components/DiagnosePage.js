/**
 * DiagnosePage.js - Страница диагностики
 *
 * Этот компонент обеспечивает:
 * - Запуск диагностики системы
 * - Отображение результатов диагностики
 * - Экспорт отчетов
 * - Создание архивов для отправки
 *
 * Основные функции:
 * - AJAX-запросы для запуска диагностики
 * - Отображение результатов в табличном виде
 * - Экспорт отчетов в PDF
 * - Создание архивов с логами и конфигурацией
 *
 * Использование:
 * Импортируется в App.js
 */

import React, { useState } from 'react';
import axios from 'axios';

const DiagnosePage = () => {
  const [diagnosis, setDiagnosis] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Запуск диагностики
  const runDiagnosis = async () => {
    setLoading(true);
    setError(null);
    setDiagnosis(null);

    try {
      const response = await axios.post('/api/diagnose/run');
      setDiagnosis(response.data);
    } catch (err) {
      setError('Ошибка диагностики: ' + (err.response?.data?.message || err.message));
    } finally {
      setLoading(false);
    }
  };

  // Экспорт отчета в PDF
  const exportPDF = async () => {
    try {
      const response = await axios.post('/api/diagnose/export/pdf', diagnosis, {
        responseType: 'blob'
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'venera_diagnosis.pdf');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (err) {
      setError('Ошибка экспорта отчета: ' + (err.response?.data?.message || err.message));
    }
  };

  // Создание архива
  const createArchive = async () => {
    try {
      const response = await axios.post('/api/diagnose/archive', null, {
        responseType: 'blob'
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'venera_diagnosis_archive.gz');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (err) {
      setError('Ошибка создания архива: ' + (err.response?.data?.message || err.message));
    }
  };

  // Форматирование статуса
  const formatStatus = (status) => {
    const statuses = {
      ok: '✅ OK',
      warning: '⚠️ Внимание',
      error: '❌ Ошибка',
      info: 'ℹ️ Информация'
    };
    return statuses[status] || status;
  };

  // Форматирование значения
  const formatValue = (value) => {
    if (typeof value === 'boolean') {
      return value ? 'Да' : 'Нет';
    }
    if (typeof value === 'number') {
      return value.toLocaleString('ru-RU');
    }
    return value;
  };

  return (
    <div className="diagnose-page">
      <div className="page-header">
        <h2>Диагностика системы</h2>
        <div className="page-actions">
          <button className="btn btn-primary" onClick={runDiagnosis} disabled={loading}>
            {loading ? 'Запуск диагностики...' : 'Запустить диагностику'}
          </button>
          {diagnosis && (
            <>
              <button className="btn btn-secondary" onClick={exportPDF}>Экспорт в PDF</button>
              <button className="btn btn-outline" onClick={createArchive}>Создать архив</button>
            </>
          )}
        </div>
      </div>

      {/* Результаты диагностики */}
      {diagnosis && (
        <div className="card">
          <div className="card-header">
            <h3>Результаты диагностики</h3>
          </div>
          <div className="diagnosis-details">
            {/* Версия приложения */}
            <div className="diagnosis-section">
              <h4>Версия приложения</h4>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Версия:</span>
                <span className="diagnosis-value">{diagnosis.appVersion || 'Неизвестно'}</span>
              </div>
            </div>

            {/* Конфигурация */}
            <div className="diagnosis-section">
              <h4>Конфигурация</h4>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Файл конфигурации:</span>
                <span className={`diagnosis-value ${diagnosis.configFileExists ? 'ok' : 'error'}`}>
                  {formatStatus(diagnosis.configFileExists ? 'ok' : 'error')}
                </span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Путь:</span>
                <span className="diagnosis-value">{diagnosis.configPath || 'Не указан'}</span>
              </div>
            </div>

            {/* Манифест */}
            <div className="diagnosis-section">
              <h4>Регистрация манифеста</h4>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Зарегистрирован:</span>
                <span className={`diagnosis-value ${diagnosis.manifestRegistered ? 'ok' : 'error'}`}>
                  {formatStatus(diagnosis.manifestRegistered ? 'ok' : 'error')}
                </span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Версия манифеста:</span>
                <span className="diagnosis-value">{diagnosis.manifestVersion || 'Неизвестно'}</span>
              </div>
            </div>

            {/* Ресурсы */}
            <div className="diagnosis-section">
              <h4>Системные ресурсы</h4>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Объем RAM:</span>
                <span className="diagnosis-value">{diagnosis.ramTotal ? `${(diagnosis.ramTotal / 1024 / 1024 / 1024).toFixed(1)} GB` : 'Неизвестно'}</span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Свободно RAM:</span>
                <span className="diagnosis-value">{diagnosis.ramFree ? `${(diagnosis.ramFree / 1024 / 1024 / 1024).toFixed(1)} GB` : 'Неизвестно'}</span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Свободно на диске (PostgreSQL):</span>
                <span className={`diagnosis-value ${diagnosis.diskPostgresFree ? (diagnosis.diskPostgresFree > 10737418240 ? 'ok' : 'warning') : 'error'}`}>
                  {diagnosis.diskPostgresFree ? `${(diagnosis.diskPostgresFree / 1024 / 1024 / 1024).toFixed(1)} GB` : 'Неизвестно'}
                </span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Свободно на диске (DragonflyDB):</span>
                <span className={`diagnosis-value ${diagnosis.diskDragonflyFree ? (diagnosis.diskDragonflyFree > 10737418240 ? 'ok' : 'warning') : 'error'}`}>
                  {diagnosis.diskDragonflyFree ? `${(diagnosis.diskDragonflyFree / 1024 / 1024 / 1024).toFixed(1)} GB` : 'Неизвестно'}
                </span>
              </div>
            </div>

            {/* Сетевые адаптеры */}
            <div className="diagnosis-section">
              <h4>Сетевые адаптеры</h4>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Количество:</span>
                <span className="diagnosis-value">{diagnosis.networkAdapters?.length || 0}</span>
              </div>
              {diagnosis.networkAdapters?.map((adapter, index) => (
                <div key={index} className="diagnosis-subitem">
                  <span className="diagnosis-label">• {adapter.name}:</span>
                  <span className="diagnosis-value">{adapter.ip}</span>
                </div>
              ))}
            </div>

            {/* Исполняемые файлы */}
            <div className="diagnosis-section">
              <h4>Исполняемые файлы</h4>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Podman:</span>
                <span className={`diagnosis-value ${diagnosis.podmanExists ? 'ok' : 'error'}`}>
                  {formatStatus(diagnosis.podmanExists ? 'ok' : 'error')}
                </span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Tshark:</span>
                <span className={`diagnosis-value ${diagnosis.tsharkExists ? 'ok' : 'error'}`}>
                  {formatStatus(diagnosis.tsharkExists ? 'ok' : 'error')}
                </span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Образ DragonflyDB:</span>
                <span className={`diagnosis-value ${diagnosis.dragonflyImageExists ? 'ok' : 'error'}`}>
                  {formatStatus(diagnosis.dragonflyImageExists ? 'ok' : 'error')}
                </span>
              </div>
            </div>

            {/* Подключения к базам данных */}
            <div className="diagnosis-section">
              <h4>Подключения к базам данных</h4>
              <div className="diagnosis-item">
                <span className="diagnosis-label">PostgreSQL:</span>
                <span className={`diagnosis-value ${diagnosis.postgresConnected ? 'ok' : 'error'}`}>
                  {formatStatus(diagnosis.postgresConnected ? 'ok' : 'error')}
                </span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">DragonflyDB:</span>
                <span className={`diagnosis-value ${diagnosis.dragonflyConnected ? 'ok' : 'error'}`}>
                  {formatStatus(diagnosis.dragonflyConnected ? 'ok' : 'error')}
                </span>
              </div>
            </div>

            {/* Служба */}
            <div className="diagnosis-section">
              <h4>Служба</h4>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Установлена:</span>
                <span className={`diagnosis-value ${diagnosis.serviceInstalled ? 'ok' : 'info'}`}>
                  {formatStatus(diagnosis.serviceInstalled ? 'ok' : 'info')}
                </span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Режим запуска:</span>
                <span className="diagnosis-value">{diagnosis.startType || 'Неизвестно'}</span>
              </div>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Состояние:</span>
                <span className="diagnosis-value">{diagnosis.serviceStatus || 'Неизвестно'}</span>
              </div>
            </div>

            {/* История событий */}
            <div className="diagnosis-section">
              <h4>Последние события (10)</h4>
              {diagnosis.eventLogs?.length === 0 ? (
                <div className="diagnosis-item">
                  <span className="diagnosis-value">Нет событий</span>
                </div>
              ) : (
                diagnosis.eventLogs?.map((event, index) => (
                  <div key={index} className="diagnosis-subitem">
                    <span className={`diagnosis-value ${event.level === 'Error' ? 'error' : 'warning'}`}>
                      [{event.level}] {event.message}
                    </span>
                  </div>
                ))
              )}
            </div>

            {/* Статус режима */}
            <div className="diagnosis-section">
              <h4>Режим работы</h4>
              <div className="diagnosis-item">
                <span className="diagnosis-label">Текущий режим:</span>
                <span className="diagnosis-value">{diagnosis.currentMode || 'Неизвестно'}</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Ошибка */}
      {error && (
        <div className="card">
          <div className="card-header">
            <h3>Ошибка</h3>
          </div>
          <div className="alert alert-error">{error}</div>
        </div>
      )}

      {/* Пустое состояние */}
      {!diagnosis && !error && !loading && (
        <div className="card">
          <div className="card-header">
            <h3>Информация</h3>
          </div>
          <div className="empty-state">
            <p>Нажмите "Запустить диагностику" для проверки системы</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default DiagnosePage;
