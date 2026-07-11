/**
 * Модуль диагностики системы для веб-интерфейса Venera
 * Предоставляет функции для запуска диагностики системы и просмотра результатов
 * 
 * Основные функции:
 * - Запуск диагностики системы через REST API
 * - Подключение к WebSocket для получения результатов диагностики в реальном времени
 * - Отображение результатов диагностики в интерактивном виде
 * - Поддержка темной/светлой темы через CSS переменные
 * - Интеграция с существующими компонентами React UI
 * 
 * @module diagnose
 */

// Глобальные переменные для хранения состояния модуля
let diagnoseWebSocket = null;
let diagnoseResults = [];
let isWebSocketConnected = false;
let isDiagnosticsRunning = false;
let currentDiagnosticsId = null;

/**
 * Инициализация модуля диагностики
 * Подключает WebSocket соединение и загружает результаты последней диагностики
 * 
 * @returns {Promise<void>}
 */
async function initDiagnoseModule() {
    console.log('[Diagnose] Инициализация модуля диагностики');
    
    // Подключение к WebSocket для получения результатов диагностики
    await connectDiagnoseWebSocket();
    
    // Загрузка результатов последней диагностики из API
    await loadDiagnoseResults();
    
    // Отрисовка интерфейса диагностики
    renderDiagnoseInterface();
    
    console.log('[Diagnose] Модуль диагностики успешно инициализирован');
}

/**
 * Подключение к WebSocket для получения результатов диагностики в реальном времени
 * 
 * @returns {Promise<void>}
 */
async function connectDiagnoseWebSocket() {
    try {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/diagnostics`;
        
        diagnoseWebSocket = new WebSocket(wsUrl);
        
        diagnoseWebSocket.onopen = () => {
            console.log('[Diagnose] WebSocket соединение установлено');
            isWebSocketConnected = true;
            
            // Если диагностика запущена, запросить обновление результатов
            if (isDiagnosticsRunning) {
                requestDiagnosticsUpdate();
            }
        };
        
        diagnoseWebSocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                
                if (message.type === 'diagnostics_update') {
                    console.log('[Diagnose] Получено обновление диагностики', message.data);
                    
                    if (message.data.type === 'diagnostics_start') {
                        isDiagnosticsRunning = true;
                        currentDiagnosticsId = message.data.id;
                        updateDiagnosticsStatus('running');
                    } else if (message.data.type === 'diagnostics_progress') {
                        updateDiagnosticsProgress(message.data);
                    } else if (message.data.type === 'diagnostics_complete') {
                        isDiagnosticsRunning = false;
                        currentDiagnosticsId = null;
                        diagnoseResults = message.data.results;
                        updateDiagnosticsStatus('complete');
                        renderDiagnosticsResults();
                    } else if (message.data.type === 'diagnostics_error') {
                        isDiagnosticsRunning = false;
                        currentDiagnosticsId = null;
                        updateDiagnosticsStatus('error', message.data.message);
                    }
                }
            } catch (error) {
                console.error('[Diagnose] Ошибка при обработке WebSocket сообщения:', error);
            }
        };
        
        diagnoseWebSocket.onclose = () => {
            console.log('[Diagnose] WebSocket соединение закрыто');
            isWebSocketConnected = false;
            
            // Попытка переподключения через 5 секунд
            setTimeout(connectDiagnoseWebSocket, 5000);
        };
        
        diagnoseWebSocket.onerror = (error) => {
            console.error('[Diagnose] Ошибка WebSocket соединения:', error);
        };
    } catch (error) {
        console.error('[Diagnose] Ошибка подключения к WebSocket:', error);
        isWebSocketConnected = false;
    }
}

/**
 * Загрузка результатов диагностики из REST API
 * 
 * @returns {Promise<void>}
 */
async function loadDiagnoseResults() {
    try {
        const response = await fetch('/api/diagnostics/results');
        
        if (!response.ok) {
            throw new Error(`Ошибка загрузки результатов диагностики: ${response.status} ${response.statusText}`);
        }
        
        const data = await response.json();
        diagnoseResults = data.results || [];
        
        console.log('[Diagnose] Результаты диагностики успешно загружены', diagnoseResults.length, 'записей');
    } catch (error) {
        console.error('[Diagnose] Ошибка при загрузке результатов диагностики:', error);
    }
}

/**
 * Запуск диагностики системы
 * 
 * @returns {Promise<void>}
 */
async function runDiagnostics() {
    if (isDiagnosticsRunning) {
        showNotification('Диагностика уже запущена', 'info');
        return;
    }
    
    try {
        const response = await fetch('/api/diagnostics/run', {
            method: 'POST'
        });
        
        if (!response.ok) {
            throw new Error(`Ошибка запуска диагностики: ${response.status} ${response.statusText}`);
        }
        
        const data = await response.json();
        currentDiagnosticsId = data.id;
        isDiagnosticsRunning = true;
        
        updateDiagnosticsStatus('running');
        showNotification('Диагностика успешно запущена', 'success');
        
        console.log('[Diagnose] Диагностика запущена с ID:', currentDiagnosticsId);
    } catch (error) {
        console.error('[Diagnose] Ошибка при запуске диагностики:', error);
        showNotification('Ошибка при запуске диагностики', 'error');
    }
}

/**
 * Запрос обновления результатов диагностики
 * 
 * @returns {void}
 */
function requestDiagnosticsUpdate() {
    if (diagnoseWebSocket && diagnoseWebSocket.readyState === WebSocket.OPEN) {
        diagnoseWebSocket.send(JSON.stringify({
            type: 'request_update',
            data: { id: currentDiagnosticsId }
        }));
    }
}

/**
 * Отрисовка интерфейса диагностики
 * 
 * @returns {void}
 */
function renderDiagnoseInterface() {
    const container = document.getElementById('diagnose-container');
    
    if (!container) {
        console.error('[Diagnose] Элемент diagnose-container не найден');
        return;
    }
    
    container.innerHTML = `
        <div class="diagnose-interface">
            <h2 class="diagnose-title">Диагностика системы</h2>
            
            <div class="diagnose-controls">
                <button id="run-diagnostics-btn" class="btn btn-primary" ${isDiagnosticsRunning ? 'disabled' : ''}>
                    ${isDiagnosticsRunning ? '⏳ Запуск диагностики...' : '🚀 Запустить диагностику'}
                </button>
                <button id="clear-diagnostics-btn" class="btn btn-secondary" ${diagnoseResults.length === 0 ? 'disabled' : ''}>
                    🗑️ Очистить результаты
                </button>
            </div>
            
            <div class="diagnose-status">
                <div class="status-indicator ${isDiagnosticsRunning ? 'running' : ''}">
                    <span class="status-dot"></span>
                    <span class="status-text">${isDiagnosticsRunning ? 'Диагностика запущена' : 'Готова к диагностике'}</span>
                </div>
                ${isDiagnosticsRunning ? '<div class="progress-bar"><div class="progress-fill" style="width: 0%"></div></div>' : ''}
            </div>
            
            <div class="diagnose-tabs">
                <button class="tab-btn active" data-tab="summary">Общий отчет</button>
                <button class="tab-btn" data-tab="details">Детали</button>
                <button class="tab-btn" data-tab="scripts">Скрипты</button>
            </div>
            
            <div class="diagnose-content">
                <div class="diagnose-section summary-section" data-section="summary">
                    <div class="summary-cards">
                        <div class="summary-card">
                            <div class="summary-card-title">Всего проверок</div>
                            <div class="summary-card-value" id="total-checks">0</div>
                        </div>
                        <div class="summary-card">
                            <div class="summary-card-title">Успешных</div>
                            <div class="summary-card-value success" id="success-checks">0</div>
                        </div>
                        <div class="summary-card">
                            <div class="summary-card-title">Предупреждений</div>
                            <div class="summary-card-value warning" id="warning-checks">0</div>
                        </div>
                        <div class="summary-card">
                            <div class="summary-card-title">Ошибок</div>
                            <div class="summary-card-value error" id="error-checks">0</div>
                        </div>
                    </div>
                    <div class="diagnose-list" id="summary-list">
                        <!-- Общий отчет будет отрисован здесь -->
                    </div>
                </div>
                
                <div class="diagnose-section details-section" data-section="details" style="display: none;">
                    <div class="diagnose-list" id="details-list">
                        <!-- Детали будут отрисованы здесь -->
                    </div>
                </div>
                
                <div class="diagnose-section scripts-section" data-section="scripts" style="display: none;">
                    <div class="scripts-list">
                        <h3 class="scripts-title">Доступные скрипты диагностики</h3>
                        <div class="script-item">
                            <div class="script-name">diagnose.ps1</div>
                            <div class="script-description">Консольная диагностика системы</div>
                            <button class="btn btn-secondary btn-run">▶ Запустить</button>
                        </div>
                        <div class="script-item">
                            <div class="script-name">diagnose-gui.ps1</div>
                            <div class="script-description">Диагностика с GUI (WinForms)</div>
                            <button class="btn btn-secondary btn-run">▶ Запустить</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;
    
    // Добавление обработчиков событий
    addDiagnoseEventListeners();
    
    // Отрисовать результаты
    renderDiagnosticsResults();
}

/**
 * Отрисовка результатов диагностики
 * 
 * @returns {void}
 */
function renderDiagnosticsResults() {
    // Обновление статистики
    const totalChecks = diagnoseResults.length;
    const successChecks = diagnoseResults.filter(r => r.status === 'success').length;
    const warningChecks = diagnoseResults.filter(r => r.status === 'warning').length;
    const errorChecks = diagnoseResults.filter(r => r.status === 'error').length;
    
    document.getElementById('total-checks').textContent = totalChecks;
    document.getElementById('success-checks').textContent = successChecks;
    document.getElementById('warning-checks').textContent = warningChecks;
    document.getElementById('error-checks').textContent = errorChecks;
    
    // Отрисовка списка результатов
    const summaryList = document.getElementById('summary-list');
    const detailsList = document.getElementById('details-list');
    
    if (!summaryList || !detailsList) {
        console.error('[Diagnose] Элементы summary-list или details-list не найдены');
        return;
    }
    
    let summaryHtml = '';
    let detailsHtml = '';
    
    diagnoseResults.forEach((result, index) => {
        summaryHtml += `
            <div class="diagnose-item diagnose-${result.status}">
                <div class="diagnose-header">
                    <span class="diagnose-name">${result.name}</span>
                    <span class="diagnose-status ${result.status}">${result.status}</span>
                </div>
                <div class="diagnose-body">
                    <p>${escapeHtml(result.message)}</p>
                    ${result.details ? `<details><summary>Детали</summary><pre>${escapeHtml(JSON.stringify(result.details, null, 2))}</pre></details>` : ''}
                </div>
            </div>
        `;
        
        detailsHtml += `
            <div class="diagnose-details">
                <h4 class="diagnose-details-title">${index + 1}. ${result.name}</h4>
                <div class="diagnose-details-body">
                    <div class="diagnose-details-item">
                        <span class="diagnose-details-label">Статус:</span>
                        <span class="diagnose-details-value ${result.status}">${result.status}</span>
                    </div>
                    <div class="diagnose-details-item">
                        <span class="diagnose-details-label">Сообщение:</span>
                        <span class="diagnose-details-value">${escapeHtml(result.message)}</span>
                    </div>
                    <div class="diagnose-details-item">
                        <span class="diagnose-details-label">Время:</span>
                        <span class="diagnose-details-value">${formatTimestamp(result.timestamp)}</span>
                    </div>
                    ${result.details ? `
                    <div class="diagnose-details-item">
                        <span class="diagnose-details-label">Детали:</span>
                        <pre class="diagnose-details-value">${escapeHtml(JSON.stringify(result.details, null, 2))}</pre>
                    </div>` : ''}
                </div>
            </div>
        `;
    });
    
    if (diagnoseResults.length === 0) {
        summaryHtml = '<div class="no-diagnostics">Результаты диагностики не найдены. Запустите диагностику системы.</div>';
        detailsHtml = '<div class="no-diagnostics">Результаты диагностики не найдены. Запустите диагностику системы.</div>';
    }
    
    summaryList.innerHTML = summaryHtml;
    detailsList.innerHTML = detailsHtml;
}

/**
 * Обновление статуса диагностики
 * 
 * @param {string} status Статус (running, complete, error)
 * @param {string} message Сообщение об ошибке (опционально)
 * @returns {void}
 */
function updateDiagnosticsStatus(status, message) {
    const statusText = document.querySelector('.status-text');
    const progressBar = document.querySelector('.progress-bar');
    const progressFill = document.querySelector('.progress-fill');
    const runBtn = document.getElementById('run-diagnostics-btn');
    
    if (!statusText) {
        console.error('[Diagnose] Элемент status-text не найден');
        return;
    }
    
    if (status === 'running') {
        statusText.textContent = 'Диагностика запущена';
        if (progressBar) progressBar.style.display = 'block';
        if (runBtn) runBtn.disabled = true;
    } else if (status === 'complete') {
        statusText.textContent = 'Диагностика завершена';
        if (progressBar) progressBar.style.display = 'none';
        if (runBtn) runBtn.disabled = false;
    } else if (status === 'error') {
        statusText.textContent = 'Ошибка диагностики';
        if (progressBar) progressBar.style.display = 'none';
        if (runBtn) runBtn.disabled = false;
        
        if (message) {
            showNotification(message, 'error');
        }
    }
}

/**
 * Обновление прогресса диагностики
 * 
 * @param {Object} data Данные о прогрессе
 * @returns {void}
 */
function updateDiagnosticsProgress(data) {
    const progressFill = document.querySelector('.progress-fill');
    
    if (progressFill) {
        progressFill.style.width = `${data.progress}%`;
    }
    
    if (data.current && data.total) {
        console.log(`[Diagnose] Прогресс диагностики: ${data.current}/${data.total}`);
    }
}

/**
 * Добавление обработчиков событий для интерфейса диагностики
 * 
 * @returns {void}
 */
function addDiagnoseEventListeners() {
    // Кнопка запуска диагностики
    document.getElementById('run-diagnostics-btn').addEventListener('click', runDiagnostics);
    
    // Кнопка очистки результатов
    document.getElementById('clear-diagnostics-btn').addEventListener('click', () => {
        if (confirm('Вы уверены, что хотите очистить результаты диагностики?')) {
            clearDiagnosticsResults();
        }
    });
    
    // Кнопки запуска скриптов
    document.querySelectorAll('.btn-run').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const scriptName = e.target.closest('.script-item').querySelector('.script-name').textContent;
            runScript(scriptName);
        });
    });
    
    // Переключение вкладок
    document.querySelectorAll('.diagnose-tabs .tab-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const tab = e.target.dataset.tab;
            
            // Скрыть все секции
            document.querySelectorAll('.diagnose-section').forEach(section => {
                section.style.display = 'none';
            });
            
            // Показать выбранную секцию
            document.querySelector(`.${tab}-section`).style.display = 'block';
            
            // Обновить активную вкладку
            document.querySelectorAll('.diagnose-tabs .tab-btn').forEach(b => {
                b.classList.remove('active');
            });
            e.target.classList.add('active');
        });
    });
}

/**
 * Очистка результатов диагностики
 * 
 * @returns {Promise<void>}
 */
async function clearDiagnosticsResults() {
    try {
        const response = await fetch('/api/diagnostics/clear', {
            method: 'POST'
        });
        
        if (!response.ok) {
            throw new Error(`Ошибка очистки результатов диагностики: ${response.status} ${response.statusText}`);
        }
        
        diagnoseResults = [];
        renderDiagnosticsResults();
        
        showNotification('Результаты диагностики успешно очищены', 'success');
    } catch (error) {
        console.error('[Diagnose] Ошибка при очистке результатов диагностики:', error);
        showNotification('Ошибка при очистке результатов диагностики', 'error');
    }
}

/**
 * Запуск скрипта диагностики
 * 
 * @param {string} scriptName Имя скрипта
 * @returns {Promise<void>}
 */
async function runScript(scriptName) {
    try {
        const response = await fetch('/api/diagnostics/run-script', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ script: scriptName })
        });
        
        if (!response.ok) {
            throw new Error(`Ошибка запуска скрипта: ${response.status} ${response.statusText}`);
        }
        
        const data = await response.json();
        
        if (data.success) {
            showNotification(`Скрипт ${scriptName} успешно запущен`, 'success');
        } else {
            showNotification(`Ошибка при запуске скрипта ${scriptName}: ${data.message || 'Неизвестная ошибка'}`, 'error');
        }
    } catch (error) {
        console.error('[Diagnose] Ошибка при запуске скрипта:', error);
        showNotification('Ошибка при запуске скрипта', 'error');
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
        console.warn('[Diagnose] Элемент notifications-container не найден');
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
        initDiagnoseModule,
        connectDiagnoseWebSocket,
        loadDiagnoseResults,
        runDiagnostics,
        requestDiagnosticsUpdate,
        renderDiagnoseInterface,
        renderDiagnosticsResults,
        updateDiagnosticsStatus,
        updateDiagnosticsProgress,
        addDiagnoseEventListeners,
        clearDiagnosticsResults,
        runScript,
        formatTimestamp,
        escapeHtml,
        showNotification
    };
}