/**
 * DatabasePage.js - Страница работы с базой данных
 *
 * Этот компонент обеспечивает:
 * - Просмотр данных из PostgreSQL
 * - Фильтрацию данных
 * - Поиск в выборке
 * - Экспорт данных в различные форматы
 *
 * Основные функции:
 * - AJAX-запросы для получения данных из PostgreSQL
 * - Фильтрация по источнику, ключу, значению и дате
 * - Поиск по всем полям
 * - Экспорт в Excel, CSV, JSON
 *
 * Использование:
 * Импортируется в App.js
 */

import React, { useState, useEffect } from 'react';
import axios from 'axios';

const DatabasePage = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [filter, setFilter] = useState({
    source: '',
    key: '',
    value: '',
    startDate: '',
    endDate: ''
  });
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 50,
    total: 0
  });
  const [sortBy, setSortBy] = useState('date_last');
  const [sortOrder, setSortOrder] = useState('desc');

  // Загрузка данных при монтировании
  useEffect(() => {
    loadData();
  }, [pagination.page, pagination.limit, sortBy, sortOrder, filter]);

  // Загрузка данных
  const loadData = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await axios.get('/api/database', {
        params: {
          ...filter,
          page: pagination.page,
          limit: pagination.limit,
          sortBy,
          sortOrder
        }
      });
      setData(response.data.records);
      setPagination({
        ...pagination,
        total: response.data.total
      });
    } catch (err) {
      setError('Ошибка загрузки данных: ' + (err.response?.data?.message || err.message));
    } finally {
      setLoading(false);
    }
  };

  // Изменение фильтра
  const handleFilterChange = (field, value) => {
    setFilter(prev => ({
      ...prev,
      [field]: value
    }));
    setPagination({ ...pagination, page: 1 });
  };

  // Сортировка
  const handleSort = (column) => {
    if (sortBy === column) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortBy(column);
      setSortOrder('desc');
    }
    setPagination({ ...pagination, page: 1 });
  };

  // Экспорт в Excel
  const exportExcel = async () => {
    try {
      const response = await axios.get('/api/database/export/excel', {
        params: filter,
        responseType: 'blob'
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'venera_data.xlsx');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (err) {
      setError('Ошибка экспорта в Excel: ' + (err.response?.data?.message || err.message));
    }
  };

  // Экспорт в CSV
  const exportCSV = async () => {
    try {
      const response = await axios.get('/api/database/export/csv', {
        params: filter,
        responseType: 'blob'
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'venera_data.csv');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (err) {
      setError('Ошибка экспорта в CSV: ' + (err.response?.data?.message || err.message));
    }
  };

  // Экспорт в JSON
  const exportJSON = async () => {
    try {
      const response = await axios.get('/api/database/export/json', {
        params: filter,
        responseType: 'blob'
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'venera_data.json');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (err) {
      setError('Ошибка экспорта в JSON: ' + (err.response?.data?.message || err.message));
    }
  };

  // Форматирование даты
  const formatDate = (timestamp) => {
    if (!timestamp) return '-';
    const date = new Date(timestamp * 1000);
    return date.toLocaleString('ru-RU');
  };

  // Страницы пагинации
  const totalPages = Math.ceil(pagination.total / pagination.limit);
  const pages = [];
  for (let i = 1; i <= totalPages; i++) {
    pages.push(i);
  }

  // Сортировка колонок
  const getSortIcon = (column) => {
    if (sortBy !== column) return ' ↔ ';
    return sortOrder === 'asc' ? ' ↑ ' : ' ↓ ';
  };

  return (
    <div className="database-page">
      <div className="page-header">
        <h2>База данных PostgreSQL</h2>
        <div className="page-actions">
          <button className="btn btn-primary" onClick={exportExcel}>Экспорт в Excel</button>
          <button className="btn btn-secondary" onClick={exportCSV}>Экспорт в CSV</button>
          <button className="btn btn-outline" onClick={exportJSON}>Экспорт в JSON</button>
        </div>
      </div>

      {/* Фильтры */}
      <div className="card">
        <div className="card-header">
          <h3>Фильтры</h3>
        </div>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">Источник</label>
            <input
              type="text"
              className="form-input"
              value={filter.source}
              onChange={(e) => handleFilterChange('source', e.target.value)}
              placeholder="Введите источник"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Ключ</label>
            <input
              type="text"
              className="form-input"
              value={filter.key}
              onChange={(e) => handleFilterChange('key', e.target.value)}
              placeholder="Введите ключ"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Значение</label>
            <input
              type="text"
              className="form-input"
              value={filter.value}
              onChange={(e) => handleFilterChange('value', e.target.value)}
              placeholder="Введите значение"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Начальная дата</label>
            <input
              type="date"
              className="form-input"
              value={filter.startDate}
              onChange={(e) => handleFilterChange('startDate', e.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="form-label">Конечная дата</label>
            <input
              type="date"
              className="form-input"
              value={filter.endDate}
              onChange={(e) => handleFilterChange('endDate', e.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="form-label">&nbsp;</label>
            <button className="btn btn-primary" onClick={loadData}>Применить фильтры</button>
          </div>
        </div>
      </div>

      {/* Таблица данных */}
      <div className="card">
        <div className="card-header">
          <h3>Данные из базы</h3>
        </div>
        {error && <div className="alert alert-error">{error}</div>}
        {loading ? (
          <div className="loading-state">
            <div className="spinner"></div>
            <p>Загрузка данных...</p>
          </div>
        ) : data.length === 0 ? (
          <div className="empty-state">
            <p>Нет данных, соответствующих фильтрам</p>
          </div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th onClick={() => handleSort('id')} style={{ cursor: 'pointer' }}>
                    ID{getSortIcon('id')}
                  </th>
                  <th onClick={() => handleSort('source')} style={{ cursor: 'pointer' }}>
                    Источник{getSortIcon('source')}
                  </th>
                  <th onClick={() => handleSort('key')} style={{ cursor: 'pointer' }}>
                    Ключ{getSortIcon('key')}
                  </th>
                  <th onClick={() => handleSort('value')} style={{ cursor: 'pointer' }}>
                    Значение{getSortIcon('value')}
                  </th>
                  <th onClick={() => handleSort('date_first')} style={{ cursor: 'pointer' }}>
                    Первое появление{getSortIcon('date_first')}
                  </th>
                  <th onClick={() => handleSort('date_last')} style={{ cursor: 'pointer' }}>
                    Последнее появление{getSortIcon('date_last')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {data.map((record) => (
                  <tr key={record.id}>
                    <td>{record.id}</td>
                    <td>{record.source}</td>
                    <td>{record.key}</td>
                    <td>{record.value}</td>
                    <td>{formatDate(record.date_first)}</td>
                    <td>{formatDate(record.date_last)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Пагинация */}
        <div className="pagination">
          <button
            className="btn btn-secondary"
            onClick={() => setPagination(prev => ({ ...prev, page: Math.max(1, prev.page - 1) }))}
            disabled={pagination.page === 1}
          >
            ← Назад
          </button>
          <span className="pagination-info">
            Страница {pagination.page} из {totalPages} ({pagination.total} записей)
          </span>
          <button
            className="btn btn-secondary"
            onClick={() => setPagination(prev => ({ ...prev, page: Math.min(totalPages, prev.page + 1) }))}
            disabled={pagination.page === totalPages}
          >
            Вперед →
          </button>
        </div>
      </div>
    </div>
  );
};

export default DatabasePage;
