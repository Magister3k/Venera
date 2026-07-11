/**
 * App.js - Основной компонент React-приложения Venera
 *
 * Этот компонент обеспечивает:
 * - Основную структуру приложения
 * - Навигацию между разделами
 * - Управление темой (темная/светлая)
 * - Рендеринг компонентов на основе роутинга
 *
 * Основные функции:
 * - Создание навигационной панели
 * - Переключение тем
 * - Рендеринг разделов
 * - Управление состоянием темы
 *
 * Использование:
 * Импортируется в index.js
 */

import React, { useState, useEffect } from 'react';
import { Routes, Route, Link, useLocation } from 'react-router-dom';
import ProcessesPage from './components/ProcessesPage';
import StatisticsPage from './components/StatisticsPage';
import DatabasePage from './components/DatabasePage';
import SettingsPage from './components/SettingsPage';
import LogsPage from './components/LogsPage';
import DiagnosePage from './components/DiagnosePage';
import './App.css';

// Компонент навигационной панели
const Navbar = ({ theme, toggleTheme }) => {
  const location = useLocation();
  
  const navItems = [
    { path: '/processes', label: 'Процессы', icon: 'settings_applications' },
    { path: '/statistics', label: 'Статистика', icon: 'show_chart' },
    { path: '/db', label: 'База данных', icon: 'database' },
    { path: '/settings', label: 'Настройки', icon: 'settings' },
    { path: '/logs', label: 'Логи', icon: 'list_alt' },
    { path: '/diagnose', label: 'Диагностика', icon: 'medical_services' },
  ];

  return (
    <nav className="navbar">
      <div className="navbar-brand">
        <h1>Venera</h1>
      </div>
      <div className="navbar-nav">
        {navItems.map((item) => (
          <Link
            key={item.path}
            to={item.path}
            className={`nav-item ${location.pathname === item.path ? 'active' : ''}`}
          >
            <span className="nav-icon">{item.icon}</span>
            {item.label}
          </Link>
        ))}
      </div>
      <div className="navbar-theme">
        <button
          className="btn btn-secondary theme-toggle"
          onClick={toggleTheme}
          title={`Переключить на ${theme === 'light' ? 'темную' : 'светлую'} тему`}
        >
          {theme === 'light' ? '🌙' : '☀️'}
        </button>
      </div>
    </nav>
  );
};

// Компонент контента
const Content = () => {
  return (
    <div className="content">
      <Routes>
        <Route path="/processes" element={<ProcessesPage />} />
        <Route path="/statistics" element={<StatisticsPage />} />
        <Route path="/db" element={<DatabasePage />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="/logs" element={<LogsPage />} />
        <Route path="/diagnose" element={<DiagnosePage />} />
        <Route path="/" element={<ProcessesPage />} />
      </Routes>
    </div>
  );
};

// Основной компонент приложения
function App() {
  const [theme, setTheme] = useState(() => {
    // Восстановление темы из localStorage
    const savedTheme = localStorage.getItem('theme');
    return savedTheme || 'light';
  });

  // Применение темы
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  // Переключение темы
  const toggleTheme = () => {
    setTheme(prevTheme => prevTheme === 'light' ? 'dark' : 'light');
  };

  return (
    <div className="app">
      <Navbar theme={theme} toggleTheme={toggleTheme} />
      <Content />
    </div>
  );
}

export default App;
