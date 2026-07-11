/* ========================================
   Модуль управления процессами обработки
   для веб-интерфейса системы Venera
   ======================================== */

// Глобальные переменные
let processesSocket = null; // WebSocket соединение для процессов
let processesList = []; // Список процессов
let processesRefreshInterval = null; // Интервал обновления списка процессов

// Инициализация модуля процессов
function initProcessesModule() {
    console.log('Инициализация модуля процессов обработки');
    
    // Подключение к WebSocket для получения обновлений в реальном времени
    connectProcessesWebSocket();
    
    // Загрузка списка процессов при инициализации
    loadProcessesList();
    
    // Установка интервала обновления (каждые 5 секунд)
    processesRefreshInterval = setInterval(loadProcessesList, 5000);
    
    // Обработчики событий интерфейса
    setupProcessesEventListeners();
}

// Подключение к WebSocket для процессов
function connectProcessesWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/processes`;
    
    processesSocket = new WebSocket(wsUrl);
    
    processesSocket.onopen = function() {
        console.log('WebSocket подключение к процессам установлено');
        showToast('Подключение к WebSocket для процессов установлено', 'success');
    };
    
    processesSocket.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleProcessesWebSocketMessage(data);
        } catch (error) {
            console.error('Ошибка при обработке WebSocket сообщения для процессов:', error);
        }
    };
    
    processesSocket.onclose = function() {
        console.log('WebSocket подключение к процессам закрыто');
        // Попытка переподключения через 5 секунд
        setTimeout(connectProcessesWebSocket, 5000);
    };
    
    processesSocket.onerror = function(error) {
        console.error('Ошибка WebSocket подключения к процессам:', error);
        showToast('Ошибка подключения к WebSocket для процессов', 'error');
    };
}

// Обработка WebSocket сообщений для процессов
function handleProcessesWebSocketMessage(data) {
    switch (data.action) {
        case 'update':
            // Обновление списка процессов
            loadProcessesList();
            break;
        case 'status_change':
            // Изменение статуса процесса
            updateProcessStatus(data.processId, data.status);
            break;
        case 'metrics':
            // Обновление метрик процесса
            updateProcessMetrics(data.processId, data.metrics);
            break;
        default:
            console.log('Неизвестное действие WebSocket для процессов:', data.action);
    }
}

// Загрузка списка процессов
function loadProcessesList() {
    fetch('/api/processes')
        .then(response => {
            if (!response.ok) {
                throw new Error(`Ошибка загрузки списка процессов: ${response.statusText}`);
            }
            return response.json();
        })
        .then(data => {
            processesList = data.processes || [];
            renderProcessesList(processesList);
        })
        .catch(error => {
            console.error('Ошибка при загрузке списка процессов:', error);
            showToast(`Ошибка загрузки списка процессов: ${error.message}`, 'error');
        });
}

// Отрисовка списка процессов
function renderProcessesList(processes) {
    const container = document.getElementById('processes-container');
    
    if (!container) {
        console.error('Контейнер для процессов не найден');
        return;
    }
    
    if (processes.length === 0) {
        container.innerHTML = `
            <div class="card">
                <div class="card-body">
                    <p class="text-center text-secondary">Нет активных процессов обработки</p>
                </div>
            </div>
        `;
        return;
    }
    
    // Сортировка процессов по приоритету и имени
    processes.sort((a, b) => {
        if (a.priority !== b.priority) {
            return b.priority - a.priority;
        }
        return a.name.localeCompare(b.name);
    });
    
    // Формирование HTML карточек процессов
    let html = '<div class="card-grid">';
    
    processes.forEach(process => {
        html += createProcessCard(process);
    });
    
    html += '</div>';
    container.innerHTML = html;
}

// Создание HTML карточки процесса
function createProcessCard(process) {
    const statusClass = getProcessStatusClass(process.status);
    const statusText = getProcessStatusText(process.status);
    
    return `
        <div class="card process-card" data-process-id="${process.id}">
            <div class="process-card-header">
                <h3>${escapeHtml(process.name)}</h3>
                <span class="status-indicator ${statusClass}">
                    <span class="dot"></span>
                    ${statusText}
                </span>
            </div>
            <div class="process-card-body">
                <div class="process-info">
                    <p><strong>ID:</strong> ${process.id}</p>
                    <p><strong>Тип:</strong> ${escapeHtml(process.type)}</p>
                    <p><strong>Источник:</strong> ${escapeHtml(process.source)}</p>
                    <p><strong>Приоритет:</strong> ${process.priority}</p>
                    <p><strong>Запущен:</strong> ${formatDateTime(process.startTime)}</p>
                </div>
                <div class="process-metrics">
                    <div class="metric-item">
                        <span class="metric-label">Обработано</span>
                        <span class="metric-value">${process.metrics.totalProcessed || 0}</span>
                    </div>
                    <div class="metric-item">
                        <span class="metric-label">В очереди</span>
                        <span class="metric-value">${process.metrics.queueSize || 0}</span>
                    </div>
                    <div class="metric-item">
                        <span class="metric-label">Ошибок</span>
                        <span class="metric-value">${process.metrics.errors || 0}</span>
                    </div>
                    <div class="metric-item">
                        <span class="metric-label">Скорость</span>
                        <span class="metric-value">${process.metrics.rate || 0} / сек</span>
                    </div>
                </div>
                <div class="process-actions">
                    <div class="btn-group">
                        ${process.status === 'running' 
                            ? `<button class="btn btn-danger btn-sm" onclick="stopProcess('${process.id}')">
                                <span>🛑</span> Остановить
                            </button>`
                            : `<button class="btn btn-success btn-sm" onclick="startProcess('${process.id}')">
                                <span>▶</span> Запустить
                            </button>`
                        }
                        <button class="btn btn-secondary btn-sm" onclick="restartProcess('${process.id}')">
                            <span>🔄</span> Перезапустить
                        </button>
                        <button class="btn btn-secondary btn-sm" onclick="viewProcessDetails('${process.id}')">
                            <span>📋</span> Подробнее
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `;
}

// Получение CSS класса для статуса процесса
function getProcessStatusClass(status) {
    switch (status) {
        case 'running':
            return 'active';
        case 'stopped':
        case 'stopped_error':
            return 'inactive';
        case 'starting':
        case 'stopping':
            return 'warning';
        case 'error':
            return 'error';
        default:
            return 'info';
    }
}

// Получение текстового представления статуса процесса
function getProcessStatusText(status) {
    const statusMap = {
        'running': 'Работает',
        'stopped': 'Остановлен',
        'stopped_error': 'Остановлен с ошибкой',
        'starting': 'Запуск...',
        'stopping': 'Остановка...',
        'error': 'Ошибка',
        'idle': 'Ожидание',
        'paused': 'Пауза'
    };
    
    return statusMap[status] || status;
}

// Обновление статуса процесса в интерфейсе
function updateProcessStatus(processId, status) {
    const processCard = document.querySelector(`.process-card[data-process-id="${processId}"]`);
    
    if (processCard) {
        const statusIndicator = processCard.querySelector('.status-indicator');
        if (statusIndicator) {
            statusIndicator.className = `status-indicator ${getProcessStatusClass(status)}`;
            statusIndicator.innerHTML = `
                <span class="dot"></span>
                ${getProcessStatusText(status)}
            `;
        }
    }
}

// Обновление метрик процесса в интерфейсе
function updateProcessMetrics(processId, metrics) {
    const processCard = document.querySelector(`.process-card[data-process-id="${processId}"]`);
    
    if (processCard) {
        const metricValues = processCard.querySelectorAll('.metric-value');
        if (metricValues.length >= 4) {
            metricValues[0].textContent = metrics.totalProcessed || 0;
            metricValues[1].textContent = metrics.queueSize || 0;
            metricValues[2].textContent = metrics.errors || 0;
            metricValues[3].textContent = `${metrics.rate || 0} / сек`;
        }
    }
}

// Запуск процесса по ID
function startProcess(processId) {
    fetch(`/api/processes/${processId}/start`, {
        method: 'POST'
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка запуска процесса: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showToast(`Процесс ${processId} успешно запущен`, 'success');
        loadProcessesList();
    })
    .catch(error => {
        console.error('Ошибка при запуске процесса:', error);
        showToast(`Ошибка запуска процесса: ${error.message}`, 'error');
    });
}

// Остановка процесса по ID
function stopProcess(processId) {
    fetch(`/api/processes/${processId}/stop`, {
        method: 'POST'
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка остановки процесса: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showToast(`Процесс ${processId} успешно остановлен`, 'success');
        loadProcessesList();
    })
    .catch(error => {
        console.error('Ошибка при остановке процесса:', error);
        showToast(`Ошибка остановки процесса: ${error.message}`, 'error');
    });
}

// Перезапуск процесса по ID
function restartProcess(processId) {
    fetch(`/api/processes/${processId}/restart`, {
        method: 'POST'
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка перезапуска процесса: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showToast(`Процесс ${processId} успешно перезапущен`, 'success');
        loadProcessesList();
    })
    .catch(error => {
        console.error('Ошибка при перезапуске процесса:', error);
        showToast(`Ошибка перезапуска процесса: ${error.message}`, 'error');
    });
}

// Просмотр деталей процесса
function viewProcessDetails(processId) {
    const process = processesList.find(p => p.id === processId);
    
    if (!process) {
        showToast('Процесс не найден', 'error');
        return;
    }
    
    // Показать модальное окно с деталями
    showProcessDetailsModal(process);
}

// Показ модального окна с деталями процесса
function showProcessDetailsModal(process) {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h2>Детали процесса: ${escapeHtml(process.name)}</h2>
                <button class="modal-close" onclick="this.closest('.modal').remove()">&times;</button>
            </div>
            <div class="modal-body">
                <div class="card">
                    <div class="card-header">
                        <h3>Общая информация</h3>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="form-label">ID процесса</label>
                            <div class="form-control" style="background: transparent;">${process.id}</div>
                        </div>
                        <div class="form-group">
                            <label class="form-label">Тип процесса</label>
                            <div class="form-control" style="background: transparent;">${escapeHtml(process.type)}</div>
                        </div>
                        <div class="form-group">
                            <label class="form-label">Источник данных</label>
                            <div class="form-control" style="background: transparent;">${escapeHtml(process.source)}</div>
                        </div>
                        <div class="form-group">
                            <label class="form-label">Приоритет</label>
                            <div class="form-control" style="background: transparent;">${process.priority}</div>
                        </div>
                    </div>
                </div>
                
                <div class="card">
                    <div class="card-header">
                        <h3>Статус и метрики</h3>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="form-label">Текущий статус</label>
                            <span class="status-indicator ${getProcessStatusClass(process.status)}">
                                <span class="dot"></span>
                                ${getProcessStatusText(process.status)}
                            </span>
                        </div>
                        <div class="form-group">
                            <label class="form-label">Время запуска</label>
                            <div class="form-control" style="background: transparent;">${formatDateTime(process.startTime)}</div>
                        </div>
                        <div class="form-group">
                            <label class="form-label">Время остановки</label>
                            <div class="form-control" style="background: transparent;">${process.stopTime ? formatDateTime(process.stopTime) : 'Не останавливался'}</div>
                        </div>
                        <div class="form-group">
                            <label class="form-label">Обработано идентификаторов</label>
                            <div class="form-control" style="background: transparent;">${process.metrics.totalProcessed || 0}</div>
                        </div>
                        <div class="form-group">
                            <label class="form-label">Ошибок обработки</label>
                            <div class="form-control" style="background: transparent;">${process.metrics.errors || 0}</div>
                        </div>
                        <div class="form-group">
                            <label class="form-label">Размер очереди</label>
                            <div class="form-control" style="background: transparent;">${process.metrics.queueSize || 0}</div>
                        </div>
                        <div class="form-group">
                            <label class="form-label">Скорость обработки</label>
                            <div class="form-control" style="background: transparent;">${process.metrics.rate || 0} ид/сек</div>
                        </div>
                    </div>
                </div>
                
                <div class="card">
                    <div class="card-header">
                        <h3>Конфигурация</h3>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="form-label">Параметры процесса</label>
                            <textarea class="form-control" style="height: 150px; font-family: monospace; font-size: 0.875rem;" readonly>${JSON.stringify(process.config, null, 2)}</textarea>
                        </div>
                    </div>
                </div>
                
                <div class="card">
                    <div class="card-header">
                        <h3>Логи процесса</h3>
                    </div>
                    <div class="card-body">
                        <div class="console-log" id="process-logs-${process.id}">
                            <div class="log-entry level-info">
                                <span class="timestamp">${new Date().toISOString()}</span>
                                <span>Загрузка логов...</span>
                            </div>
                        </div>
                        <button class="btn btn-primary btn-sm" style="margin-top: 10px;" onclick="loadProcessLogs('${process.id}')">
                            <span>🔄</span> Обновить логи
                        </button>
                    </div>
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn btn-secondary" onclick="this.closest('.modal').remove()">Закрыть</button>
                ${process.status === 'running' 
                    ? `<button class="btn btn-danger" onclick="stopProcess('${process.id}'); this.closest('.modal').remove()">Остановить</button>`
                    : `<button class="btn btn-success" onclick="startProcess('${process.id}'); this.closest('.modal').remove()">Запустить</button>`
                }
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    setTimeout(() => modal.classList.add('show'), 10);
    
    // Загрузка логов процесса
    loadProcessLogs(process.id);
}

// Загрузка логов процесса
function loadProcessLogs(processId) {
    const logContainer = document.getElementById(`process-logs-${processId}`);
    
    if (!logContainer) {
        return;
    }
    
    logContainer.innerHTML = '<div class="spinner" style="margin: 20px auto;"></div>';
    
    fetch(`/api/processes/${processId}/logs`)
        .then(response => {
            if (!response.ok) {
                throw new Error(`Ошибка загрузки логов: ${response.statusText}`);
            }
            return response.json();
        })
        .then(data => {
            logContainer.innerHTML = '';
            
            if (!data.logs || data.logs.length === 0) {
                logContainer.innerHTML = '<p class="text-center text-secondary">Нет доступных логов</p>';
                return;
            }
            
            data.logs.forEach(log => {
                const logEntry = document.createElement('div');
                logEntry.className = `log-entry level-${log.level}`;
                logEntry.innerHTML = `
                    <span class="timestamp">${log.timestamp}</span>
                    <span>${escapeHtml(log.message)}</span>
                `;
                logContainer.appendChild(logEntry);
            });
            
            // Прокрутка вниз
            logContainer.scrollTop = logContainer.scrollHeight;
        })
        .catch(error => {
            console.error('Ошибка при загрузке логов процесса:', error);
            logContainer.innerHTML = `
                <div class="log-entry level-error">
                    <span class="timestamp">${new Date().toISOString()}</span>
                    <span>Ошибка загрузки логов: ${error.message}</span>
                </div>
            `;
        });
}

// Настройка обработчиков событий интерфейса
function setupProcessesEventListeners() {
    // Кнопка обновления списка процессов
    const refreshBtn = document.getElementById('refresh-processes-btn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', loadProcessesList);
    }
    
    // Кнопка добавления нового процесса
    const addProcessBtn = document.getElementById('add-process-btn');
    if (addProcessBtn) {
        addProcessBtn.addEventListener('click', showAddProcessModal);
    }
    
    // Кнопка массовой остановки всех процессов
    const stopAllBtn = document.getElementById('stop-all-processes-btn');
    if (stopAllBtn) {
        stopAllBtn.addEventListener('click', stopAllProcesses);
    }
    
    // Кнопка массового запуска всех процессов
    const startAllBtn = document.getElementById('start-all-processes-btn');
    if (startAllBtn) {
        startAllBtn.addEventListener('click', startAllProcesses);
    }
}

// Показ модального окна добавления нового процесса
function showAddProcessModal() {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h2>Добавить новый процесс обработки</h2>
                <button class="modal-close" onclick="this.closest('.modal').remove()">&times;</button>
            </div>
            <div class="modal-body">
                <form id="add-process-form">
                    <div class="form-group">
                        <label class="form-label required">Название процесса</label>
                        <input type="text" class="form-control" name="name" required placeholder="Введите название процесса">
                    </div>
                    <div class="form-group">
                        <label class="form-label required">Тип процесса</label>
                        <select class="form-control" name="type" required>
                            <option value="packet_stream">Поток пакетов (network)</option>
                            <option value="folder_scan">Сканирование папки (folder)</option>
                            <option value="file_read">Чтение файла (file)</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label class="form-label required">Источник данных</label>
                        <input type="text" class="form-control" name="source" required placeholder="IP:PORT или путь к папке/файлу">
                    </div>
                    <div class="form-group">
                        <label class="form-label">Приоритет (1-10)</label>
                        <input type="number" class="form-control" name="priority" min="1" max="10" value="5" placeholder="Приоритет обработки">
                    </div>
                    <div class="form-group">
                        <label class="form-label">Конфигурация (JSON)</label>
                        <textarea class="form-control" name="config" style="height: 200px; font-family: monospace;" placeholder='{"param1": "value1", "param2": "value2"}'></textarea>
                    </div>
                </form>
            </div>
            <div class="modal-footer">
                <button class="btn btn-secondary" onclick="this.closest('.modal').remove()">Отмена</button>
                <button class="btn btn-success" onclick="submitAddProcess()">Добавить</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    setTimeout(() => modal.classList.add('show'), 10);
}

// Отправка формы добавления процесса
function submitAddProcess() {
    const form = document.getElementById('add-process-form');
    if (!form) {
        showToast('Форма не найдена', 'error');
        return;
    }
    
    const formData = {
        name: form.elements.name.value,
        type: form.elements.type.value,
        source: form.elements.source.value,
        priority: parseInt(form.elements.priority.value) || 5,
        config: form.elements.config.value ? JSON.parse(form.elements.config.value) : {}
    };
    
    fetch('/api/processes', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка добавления процесса: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showToast('Процесс успешно добавлен', 'success');
        document.querySelector('.modal').remove();
        loadProcessesList();
    })
    .catch(error => {
        console.error('Ошибка при добавлении процесса:', error);
        showToast(`Ошибка добавления процесса: ${error.message}`, 'error');
    });
}

// Остановка всех процессов
function stopAllProcesses() {
    if (!confirm('Вы уверены, что хотите остановить все процессы?')) {
        return;
    }
    
    fetch('/api/processes/stop-all', {
        method: 'POST'
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка остановки всех процессов: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showToast('Все процессы успешно остановлены', 'success');
        loadProcessesList();
    })
    .catch(error => {
        console.error('Ошибка при остановке всех процессов:', error);
        showToast(`Ошибка остановки всех процессов: ${error.message}`, 'error');
    });
}

// Запуск всех процессов
function startAllProcesses() {
    if (!confirm('Вы уверены, что хотите запустить все процессы?')) {
        return;
    }
    
    fetch('/api/processes/start-all', {
        method: 'POST'
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка запуска всех процессов: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showToast('Все процессы успешно запущены', 'success');
        loadProcessesList();
    })
    .catch(error => {
        console.error('Ошибка при запуске всех процессов:', error);
        showToast(`Ошибка запуска всех процессов: ${error.message}`, 'error');
    });
}

// Очистка ресурсов при переключении страницы
function cleanupProcessesModule() {
    console.log('Очистка модуля процессов обработки');
    
    if (processesRefreshInterval) {
        clearInterval(processesRefreshInterval);
        processesRefreshInterval = null;
    }
    
    if (processesSocket) {
        processesSocket.close();
        processesSocket = null;
    }
}

// Экспорт функций для глобального использования
window.initProcessesModule = initProcessesModule;
window.cleanupProcessesModule = cleanupProcessesModule;
window.startProcess = startProcess;
window.stopProcess = stopProcess;
window.restartProcess = restartProcess;
window.viewProcessDetails = viewProcessDetails;
window.loadProcessLogs = loadProcessLogs;
window.showAddProcessModal = showAddProcessModal;
window.submitAddProcess = submitAddProcess;
window.stopAllProcesses = stopAllProcesses;
window.startAllProcesses = startAllProcesses;
