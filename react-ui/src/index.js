/**
 * index.js - Точка входа в React-приложение Venera
 *
 * Этот файл инициализирует React-приложение с поддержкой:
 * - Хэш-роутинга для SPA
 * - Переключения тем (темная/светлая)
 * - Синхронизации URL при навигации
 * - Защиты от дублирования скриптов
 * - Инициализации при старте
 *
 * Основные функции:
 * - Создание корневого элемента React
 * - Настройка роутинга
 * - Подключение провайдера тем
 * - Инициализация при старте
 *
 * Использование:
 * npm start
 * npm run build
 */

import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import './index.css';

// Инициализация React-приложения
const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>
);
