/**
 * Модуль отображения логов для веб-интерфейса Venera
 * Предоставляет функции для просмотра системных и процессных логов в реальном времени
 * 
 * Основные функции:
 * - Подключение к WebSocket для получения логов в реальном времени
 * - REST API взаимодействие с сервером: GET /api/logs/system, GET /api/logs/process/:id
 * - Фильтрация логов по уровню, источнику и времени
 * - Поддержка темной/светлой темы через CSS переменные
 * - Интеграция с существующими компонентами React UI
 * 
 * @module logs
 */

// Глобальные переменные для хранения состояния модуля
let systemLogsWebSocket = null;
let processLogsWebSocket = null;
let systemLogsData = [];
let processLogsData = {};
let isSystemWebSocketConnected = false;
let isProcessWebSocketConnected = false;
let currentProcessId = null;
let logFilters = {
    level: 'all',
    source: 'all',
    timeFrom: null,
    timeTo: null
};

/**
 * Инициализация модуля логов
 * Подключает WebSocket соединения и загружает текущие логи
 * 
 * @returns {Promise<void>}
 */
async function initLogsModule() {
    console.log('[Logs] Инициализация модуля логов');
    
    // Подключение к WebSocket для получения системных логов
    await connectSystemLogsWebSocket();
    
    // Подключение к WebSocket для получения логов процессов
    await connectProcessLogsWebSocket();
    
    // Загрузка текущих логов из API
    await loadSystemLogs();
    
    // Отрисовка интерфейса логов
    renderLogsInterface();
    
    console.log('[Logs] Модуль логов успешно инициализирован');
}

/**
 * Подключение к WebSocket для получения системных логов в реальном времени
 * 
 * @returns {Promise<void>}
 */
async function connectSystemLogsWebSocket() {
    try {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/logs/system`;
        
        systemLogsWebSocket = new WebSocket(wsUrl);
        
        systemLogsWebSocket.onopen = () => {
            console.log('[Logs] WebSocket соединение для системных логов установлено');
            isSystemWebSocketConnected = true;
        };
        
        systemLogsWebSocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                
                if (message.type === 'log_update') {
                    console.log('[Logs] Получено обновление системного лога', message.data);
                    addSystemLog(message.data);
                }
            } catch (error) {
                console.error('[Logs] Ошибка при обработке WebSocket сообщения:', error);
            }
        };
        
        systemLogsWebSocket.onclose = () => {
            console.log('[Logs] WebSocket соединение для системных логов закрыто');
            isSystemWebSocketConnected = false;
            
            // Попытка переподключения через 5 секунд
            setTimeout(connectSystemLogsWebSocket, 5000);
        };
        
        systemLogsWebSocket.onerror = (error) => {
            console.error('[Logs] Ошибка WebSocket соединения для системных логов:', error);
        };
    } catch (error) {
        console.error('[Logs] Ошибка подключения к WebSocket для системных логов:', error);
        isSystemWebSocketConnected = false;
    }
}

/**
 * Подключение к WebSocket для получения логов процессов в реальном времени
 * 
 * @returns {Promise<void>}
 */
async function connectProcessLogsWebSocket() {
    try {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/logs/process`;
        
        processLogsWebSocket = new WebSocket(wsUrl);
        
        processLogsWebSocket.onopen = () => {
            console.log('[Logs] WebSocket соединение для логов процессов установлено');
            isProcessWebSocketConnected = true;
        };
        
        processLogsWebSocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                
                if (message.type === 'process_log_update') {
                    console.log('[Logs] Получено обновление лога процесса', message.data);
                    
                    if (!processLogsData[message.data.processId]) {
                        processLogsData[message.data.processId] = [];
                    }
                    
                    processLogsData[message.data.processId].push(message.data);
                    
                    if (currentProcessId === message.data.processId) {
                        updateProcessLogsView();
                    }
                }
            } catch (error) {
                console.error('[Logs] Ошибка при обработке WebSocket сообщения:', error);
            }
        };
        
        processLogsWebSocket.onclose = () => {
            console.log('[Logs] WebSocket соединение для логов процессов закрыто');
            isProcessWebSocketConnected = false;
            
            // Попытка переподключения через 5 секунд
            setTimeout(connectProcessLogsWebSocket, 5000);
        };
        
        processLogsWebSocket.onerror = (error) => {
            console.error('[Logs] Ошибка WebSocket соединения для логов процессов:', error);
        };
    } catch (error) {
        console.error('[Logs] Ошибка подключения к WebSocket для логов процессов:', error);
        isProcessWebSocketConnected = false;
    }
}

/**
 * Загрузка системных логов из REST API
 * 
 * @returns {Promise<void>}
 */
async function loadSystemLogs() {
    try {
        const params = new URLSearchParams();
        
        if (logFilters.level !== 'all') {
            params.append('level', logFilters.level);
        }
        
        if (logFilters.source !== 'all') {
            params.append('source', logFilters.source);
        }
        
        if (logFilters.timeFrom) {
            params.append('from', logFilters.timeFrom.toISOString());
        }
        
        if (logFilters.timeTo) {
            params.append('to', logFilters.timeTo.toISOString());
        }
        
        const endpoint = `/api/logs/system${params.toString() ? '?' + params.toString() : ''}`;
        const response = await fetch(endpoint);
        
        if (!response.ok) {
            throw new Error(`Ошибка загрузки системных логов: ${response.status} ${response.statusText}`);
        }
        
        systemLogsData = await response.json();
        
        // Отрисовать системные логи
        updateSystemLogsView();
        
        console.log('[Logs] Системные логи успешно загружены', systemLogsData.length, 'записей');
    } catch (error) {
        console.error('[Logs] Ошибка при загрузке системных логов:', error);
        // Показать сообщение об ошибке
        showNotification('Ошибка при загрузке системных логов', 'error');
    }
}

/**
 * Загрузка логов конкретного процесса из REST API
 * 
 * @param {string} processId Идентификатор процесса
 * @returns {Promise<void>}
 */
async function loadProcessLogs(processId) {
    try {
        currentProcessId = processId;
        
        const response = await fetch(`/api/logs/process/${processId}`);
        
        if (!response.ok) {
            throw new Error(`Ошибка загрузки логов процесса: ${response.status} ${response.statusText}`);
        }
        
        const logs = await response.json();
        processLogsData[processId] = logs;
        
        // Отрисовать логи процесса
        updateProcessLogsView();
        
        console.log('[Logs] Логи процесса успешно загружены', processId, logs.length, 'записей');
    } catch (error) {
        console.error('[Logs] Ошибка при загрузке логов процесса:', error);
        // Показать сообщение об ошибке
        showNotification('Ошибка при загрузке логов процесса', 'error');
    }
}

/**
 * Добавление системного лога в список
 * 
 * @param {Object} log Объект лога
 * @returns {void}
 */
function addSystemLog(log) {
    systemLogsData.unshift(log);
    
    // Ограничить количество логов в памяти
    if (systemLogsData.length > 1000) {
        systemLogsData = systemLogsData.slice(0, 1000);
    }
    
    // Отрисовать системные логи
    updateSystemLogsView();
}

/**
 * Отрисовка интерфейса логов
 * 
 * @returns {void}
 */
function renderLogsInterface() {
    const container = document.getElementById('logs-container');
    
    if (!container) {
        console.error('[Logs] Элемент logs-container не найден');
        return;
    }
    
    container.innerHTML = `
        <div class="logs-interface">
            <h2 class="logs-title">Логи системы</h2>
            
            <div class="logs-filters">
                <div class="filter-group">
                    <label for="log-filter-level" class="filter-label">Уровень:</label>
                    <select id="log-filter-level" class="filter-select">
                        <option value="all">Все</option>
                        <option value="debug">Debug</option>
                        <option value="info">Info</option>
                        <option value="warn">Warn</option>
                        <option value="error">Error</option>
                        <option value="fatal">Fatal</option>
                        <option value="panic">Panic</option>
                    </select>
                </div>
                
                <div class="filter-group">
                    <label for="log-filter-source" class="filter-label">Источник:</label>
                    <select id="log-filter-source" class="filter-select">
                        <option value="all">Все</option>
                        <option value="server">Сервер</option>
                        <option value="database">База данных</option>
                        <option value="processes">Процессы</option>
                        <option value="notifications">Уведомления</option>
                        <option value="diagnostics">Диагностика</option>
                    </select>
                </div>
                
                <div class="filter-group">
                    <label for="log-filter-refresh" class="filter-label">Обновление:</label>
                    <select id="log-filter-refresh" class="filter-select">
                        <option value="realtime">Реальное время</option>
                        <option value="manual">Вручную</option>
                    </select>
                </div>
                
                <button id="log-filter-apply" class="btn btn-secondary">Применить</button>
                <button id="log-filter-clear" class="btn btn-secondary">Очистить</button>
            </div>
            
            <div class="logs-tabs">
                <button class="tab-btn active" data-tab="system">Системные логи</button>
                <button class="tab-btn" data-tab="processes">Логи процессов</button>
            </div>
            
            <div class="logs-content">
                <div class="logs-section system-logs" data-section="system">
                    <div class="logs-list" id="system-logs-list">
                        <!-- Системные логи будут отрисованы здесь -->
                    </div>
                </div>
                
                <div class="logs-section process-logs" data-section="processes" style="display: none;">
                    <div class="logs-list" id="process-logs-list">
                        <!-- Логи процессов будут отрисованы здесь -->
                    </div>
                </div>
            </div>
        </div>
    `;
    
    // Добавление обработчиков событий
    addLogsEventListeners();
    
    // Отрисовать системные логи
    updateSystemLogsView();
}

/**
 * Отрисовка системных логов
 * 
 * @returns {void}
 */
function updateSystemLogsView() {
    const list = document.getElementById('system-logs-list');
    
    if (!list) {
        console.error('[Logs] Элемент system-logs-list не найден');
        return;
    }
    
    // Фильтрация логов
    let filteredLogs = systemLogsData;
    
    if (logFilters.level !== 'all') {
        filteredLogs = filteredLogs.filter(log => log.level === logFilters.level);
    }
    
    if (logFilters.source !== 'all') {
        filteredLogs = filteredLogs.filter(log => log.source === logFilters.source);
    }
    
    if (logFilters.timeFrom) {
        filteredLogs = filteredLogs.filter(log => new Date(log.timestamp) >= logFilters.timeFrom);
    }
    
    if (logFilters.timeTo) {
        filteredLogs = filteredLogs.filter(log => new Date(log.timestamp) <= logFilters.timeTo);
    }
    
    // Генерация HTML
    let html = '';
    
    filteredLogs.forEach(log => {
        html += `
            <div class="log-entry log-${log.level}">
                <div class="log-header">
                    <span class="log-timestamp">${formatTimestamp(log.timestamp)}</span>
                    <span class="log-level">${log.level}</span>
                    <span class="log-source">${log.source}</span>
                </div>
                <div class="log-body">
                    ${escapeHtml(log.message)}
                </div>
                ${log.details ? `<div class="log-details">${escapeHtml(JSON.stringify(log.details, null, 2))}</div>` : ''}
            </div>
        `;
    });
    
    if (filteredLogs.length === 0) {
        html = '<div class="no-logs">Логи не найдены</div>';
    }
    
    list.innerHTML = html;
    
    // Прокрутка к последнему логу
    list.scrollTop = list.scrollHeight;
}

/**
 * Отрисовка логов процессов
 * 
 * @returns {void}
 */
function updateProcessLogsView() {
    const list = document.getElementById('process-logs-list');
    
    if (!list) {
        console.error('[Logs] Элемент process-logs-list не найден');
        return;
    }
    
    const logs = currentProcessId && processLogsData[currentProcessId] ? processLogsData[currentProcessId] : [];
    
    // Генерация HTML
    let html = '';
    
    logs.forEach(log => {
        html += `
            <div class="log-entry log-${log.level}">
                <div class="log-header">
                    <span class="log-timestamp">${formatTimestamp(log.timestamp)}</span>
                    <span class="log-level">${log.level}</span>
                    <span class="log-source">process:${log.processId}</span>
                </div>
                <div class="log-body">
                    ${escapeHtml(log.message)}
                </div>
            </div>
        `;
    });
    
    if (logs.length === 0) {
        html = '<div class="no-logs">Логи процесса не найдены</div>';
    }
    
    list.innerHTML = html;
    
    // Прокрутка к последнему логу
    list.scrollTop = list.scrollHeight;
}

/**
 * Добавление обработчиков событий для интерфейса логов
 * 
 * @returns {void}
 */
function addLogsEventListeners() {
    // Фильтрация по уровню
    document.getElementById('log-filter-level').addEventListener('change', (e) => {
        logFilters.level = e.target.value;
    });
    
    // Фильтрация по источнику
    document.getElementById('log-filter-source').addEventListener('change', (e) => {
        logFilters.source = e.target.value;
    });
    
    // Переключение режима обновления
    document.getElementById('log-filter-refresh').addEventListener('change', (e) => {
        if (e.target.value === 'realtime') {
            console.log('[Logs] Режим обновления: реальное время');
        } else {
            console.log('[Logs] Режим обновления: вручную');
        }
    });
    
    // Применение фильтров
    document.getElementById('log-filter-apply').addEventListener('click', loadSystemLogs);
    
    // Очистка логов
    document.getElementById('log-filter-clear').addEventListener('click', () => {
        if (confirm('Вы уверены, что хотите очистить системные логи?')) {
            clearSystemLogs();
        }
    });
    
    // Переключение вкладок
    document.querySelectorAll('.logs-tabs .tab-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const tab = e.target.dataset.tab;
            
            // Скрыть все секции
            document.querySelectorAll('.logs-section').forEach(section => {
                section.style.display = 'none';
            });
            
            // Показать выбранную секцию
            document.querySelector(`.${tab}-logs`).style.display = 'block';
            
            // Обновить активную вкладку
            document.querySelectorAll('.logs-tabs .tab-btn').forEach(b => {
                b.classList.remove('active');
            });
            e.target.classList.add('active');
        });
    });
}

/**
 * Очистка системных логов
 * 
 * @returns {Promise<void>}
 */
async function clearSystemLogs() {
    try {
        const response = await fetch('/api/logs/clear', {
            method: 'POST'
        });
        
        if (!response.ok) {
            throw new Error(`Ошибка очистки логов: ${response.status} ${response.statusText}`);
        }
        
        systemLogsData = [];
        updateSystemLogsView();
        
        showNotification('Системные логи успешно очищены', 'success');
    } catch (error) {
        console.error('[Logs] Ошибка при очистке логов:', error);
        showNotification('Ошибка при очистке логов', 'error');
    }
}

/**
 * Форматирование метки времени
 * 
 * @param {string} timestamp Метка времени в ISO формате
 * @returns {string} Отформатированная метка времени
 */
function formatTimestamp(timestamp) {
    const date = new Date(timestamp);
    
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    const seconds = String(date.getSeconds()).padStart(2, '0');
    
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}

/**
 * Экранирование HTML
 * 
 * @param {string} html Строка для экранирования
 * @returns {string} Экранированная строка
 */
function escapeHtml(html) {
    const div = document.createElement('div');
    div.textContent = html;
    return div.innerHTML;
}

/**
 * Отображение уведомления пользователю
 * 
 * @param {string} message Текст уведомления
 * @param {string} type Тип уведомления (success, error, info)
 * @returns {void}
 */
function showNotification(message, type = 'info') {
    const container = document.getElementById('notifications-container');
    
    if (!container) {
        console.warn('[Logs] Элемент notifications-container не найден');
        return;
    }
    
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;
    
    container.appendChild(notification);
    
    // Удалить уведомление через 5 секунд
    setTimeout(() => {
        notification.remove();
    }, 5000);
}

// Экспорт функций для глобального использования
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        initLogsModule,
        connectSystemLogsWebSocket,
        connectProcessLogsWebSocket,
        loadSystemLogs,
        loadProcessLogs,
        addSystemLog,
        renderLogsInterface,
        updateSystemLogsView,
        updateProcessLogsView,
        addLogsEventListeners,
        clearSystemLogs,
        formatTimestamp,
        escapeHtml,
        showNotification
    };
}