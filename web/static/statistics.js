/* ========================================
   Модуль отображения статистики и метрик
   для веб-интерфейса системы Venera
   ======================================== */

// Глобальные переменные
let systemMetrics = null; // Системные метрики
let processesMetrics = []; // Метрики процессов
let metricsRefreshInterval = null; // Интервал обновления метрик
let metricsSocket = null; // WebSocket соединение для метрик

// Инициализация модуля статистики
function initStatisticsModule() {
    console.log('Инициализация модуля статистики и метрик');
    
    // Подключение к WebSocket для получения метрик в реальном времени
    connectMetricsWebSocket();
    
    // Загрузка системных и процессных метрик при инициализации
    loadSystemMetrics();
    loadProcessesMetrics();
    
    // Установка интервала обновления (каждые 3 секунды)
    metricsRefreshInterval = setInterval(() => {
        loadSystemMetrics();
        loadProcessesMetrics();
    }, 3000);
    
    // Настройка обработчиков событий интерфейса
    setupStatisticsEventListeners();
}

// Подключение к WebSocket для метрик
function connectMetricsWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/metrics`;
    
    metricsSocket = new WebSocket(wsUrl);
    
    metricsSocket.onopen = function() {
        console.log('WebSocket подключение к метрикам установлено');
        showToast('Подключение к WebSocket для метрик установлено', 'success');
    };
    
    metricsSocket.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleMetricsWebSocketMessage(data);
        } catch (error) {
            console.error('Ошибка при обработке WebSocket сообщения для метрик:', error);
        }
    };
    
    metricsSocket.onclose = function() {
        console.log('WebSocket подключение к метрикам закрыто');
        // Попытка переподключения через 5 секунд
        setTimeout(connectMetricsWebSocket, 5000);
    };
    
    metricsSocket.onerror = function(error) {
        console.error('Ошибка WebSocket подключения к метрикам:', error);
        showToast('Ошибка подключения к WebSocket для метрик', 'error');
    };
}

// Обработка WebSocket сообщений для метрик
function handleMetricsWebSocketMessage(data) {
    switch (data.type) {
        case 'system':
            // Обновление системных метрик
            systemMetrics = data.metrics;
            renderSystemMetrics(systemMetrics);
            updateSystemCharts(systemMetrics);
            break;
        case 'process':
            // Обновление метрик процесса
            updateProcessMetricsInList(data.processId, data.metrics);
            break;
        case 'all':
            // Обновление всех метрик
            if (data.systemMetrics) {
                systemMetrics = data.systemMetrics;
                renderSystemMetrics(systemMetrics);
                updateSystemCharts(systemMetrics);
            }
            if (data.processMetrics) {
                processesMetrics = data.processMetrics;
                renderProcessesMetrics(processesMetrics);
            }
            break;
        default:
            console.log('Неизвестный тип метрик:', data.type);
    }
}

// Загрузка системных метрик
function loadSystemMetrics() {
    fetch('/api/metrics/system')
        .then(response => {
            if (!response.ok) {
                throw new Error(`Ошибка загрузки системных метрик: ${response.statusText}`);
            }
            return response.json();
        })
        .then(data => {
            systemMetrics = data.metrics || null;
            renderSystemMetrics(systemMetrics);
            updateSystemCharts(systemMetrics);
        })
        .catch(error => {
            console.error('Ошибка при загрузке системных метрик:', error);
        });
}

// Загрузка метрик процессов
function loadProcessesMetrics() {
    fetch('/api/metrics/processes')
        .then(response => {
            if (!response.ok) {
                throw new Error(`Ошибка загрузки метрик процессов: ${response.statusText}`);
            }
            return response.json();
        })
        .then(data => {
            processesMetrics = data.metrics || [];
            renderProcessesMetrics(processesMetrics);
        })
        .catch(error => {
            console.error('Ошибка при загрузке метрик процессов:', error);
        });
}

// Отрисовка системных метрик
function renderSystemMetrics(metrics) {
    if (!metrics) {
        return;
    }
    
    // Обновление общих показателей
    updateSystemStatItem('cpu-usage', `${metrics.cpuUsage.toFixed(1)}%`);
    updateSystemStatItem('ram-usage', `${formatBytes(metrics.ramUsed)} / ${formatBytes(metrics.ramTotal)}`);
    updateSystemStatItem('disk-usage', `${formatBytes(metrics.diskUsed)} / ${formatBytes(metrics.diskTotal)}`);
    updateSystemStatItem('network-rx', formatBytes(metrics.networkRx));
    updateSystemStatItem('network-tx', formatBytes(metrics.networkTx));
    updateSystemStatItem('processes-count', metrics.processesCount || 0);
    updateSystemStatItem('identifiers-total', metrics.identifiersTotal || 0);
    updateSystemStatItem('identifiers-daily', metrics.identifiersDaily || 0);
}

// Обновление элемента статистики
function updateSystemStatItem(elementId, value) {
    const element = document.getElementById(elementId);
    if (element) {
        element.textContent = value;
    }
}

// Отрисовка метрик процессов
function renderProcessesMetrics(metrics) {
    const container = document.getElementById('processes-metrics-container');
    
    if (!container) {
        return;
    }
    
    if (metrics.length === 0) {
        container.innerHTML = '<p class="text-center text-secondary">Нет доступных метрик процессов</p>';
        return;
    }
    
    // Сортировка процессов по скорости обработки
    metrics.sort((a, b) => (b.rate || 0) - (a.rate || 0));
    
    let html = '<div class="card-grid">';
    
    metrics.forEach(metric => {
        html += createProcessMetricsCard(metric);
    });
    
    html += '</div>';
    container.innerHTML = html;
}

// Создание HTML карточки метрик процесса
function createProcessMetricsCard(metric) {
    const progressColor = getProgressColor(metric.rate || 0);
    
    return `
        <div class="card">
            <div class="card-header">
                <h3>${escapeHtml(metric.name || 'Без названия')}</h3>
            </div>
            <div class="card-body">
                <div class="form-group">
                    <label class="form-label">Скорость обработки</label>
                    <div class="progress-container">
                        <div class="progress-bar ${progressColor}" style="width: ${Math.min((metric.rate || 0) / 100 * 100, 100)}%"></div>
                    </div>
                    <div class="text-center mt-2">${metric.rate || 0} ид/сек</div>
                </div>
                <div class="form-group">
                    <label class="form-label">Обработано всего</label>
                    <div class="form-control" style="background: transparent;">${metric.totalProcessed || 0}</div>
                </div>
                <div class="form-group">
                    <label class="form-label">В очереди</label>
                    <div class="form-control" style="background: transparent;">${metric.queueSize || 0}</div>
                </div>
                <div class="form-group">
                    <label class="form-label">Ошибок</label>
                    <div class="form-control" style="background: transparent;">${metric.errors || 0}</div>
                </div>
                <div class="form-group">
                    <label class="form-label">Потребление CPU</label>
                    <div class="form-control" style="background: transparent;">${metric.cpuUsage ? metric.cpuUsage.toFixed(1) : 0}%</div>
                </div>
                <div class="form-group">
                    <label class="form-label">Потребление RAM</label>
                    <div class="form-control" style="background: transparent;">${metric.ramUsage ? formatBytes(metric.ramUsage) : '0'}</div>
                </div>
            </div>
        </div>
    `;
}

// Обновление метрик процесса в списке
function updateProcessMetricsInList(processId, metrics) {
    const processMetrics = processesMetrics.find(m => m.id === processId);
    
    if (processMetrics) {
        Object.assign(processMetrics, metrics);
        renderProcessesMetrics(processesMetrics);
    }
}

// Получение цвета прогресс бара
function getProgressColor(value) {
    if (value >= 100) return 'success';
    if (value >= 50) return 'warning';
    return '';
}

// Обновление графиков системных метрик
function updateSystemCharts(metrics) {
    if (!metrics) {
        return;
    }
    
    // Обновление графика загрузки CPU
    updateChart('cpu-chart', [
        { label: 'CPU', value: metrics.cpuUsage, color: '#3498db' }
    ]);
    
    // Обновление графика загрузки RAM
    updateChart('ram-chart', [
        { label: 'RAM', value: metrics.ramUsage, color: '#2ecc71' }
    ]);
    
    // Обновление графика загрузки диска
    updateChart('disk-chart', [
        { label: 'Disk', value: metrics.diskUsage, color: '#f39c12' }
    ]);
}

// Обновление кругового графика
function updateChart(elementId, data) {
    const canvas = document.getElementById(elementId);
    
    if (!canvas) {
        return;
    }
    
    const ctx = canvas.getContext('2d');
    
    if (!ctx) {
        return;
    }
    
    // Очистка канваса
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    
    // Настройки
    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = Math.min(centerX, centerY) - 20;
    
    // Вычисление общего значения
    const total = data.reduce((sum, item) => sum + item.value, 0);
    
    // Рисование секторов
    let currentAngle = -Math.PI / 2;
    
    data.forEach(item => {
        const sliceAngle = (item.value / total) * 2 * Math.PI;
        
        ctx.beginPath();
        ctx.moveTo(centerX, centerY);
        ctx.arc(centerX, centerY, radius, currentAngle, currentAngle + sliceAngle);
        ctx.closePath();
        
        ctx.fillStyle = item.color;
        ctx.fill();
        
        currentAngle += sliceAngle;
    });
    
    // Рисование центрального текста
    ctx.fillStyle = '#333';
    ctx.font = 'bold 20px Segoe UI';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    
    // Пример: отображение значения CPU
    if (elementId === 'cpu-chart' && data.length > 0) {
        ctx.fillText(`${data[0].value.toFixed(1)}%`, centerX, centerY);
    } else if (elementId === 'ram-chart' && data.length > 0) {
        ctx.fillText(`${data[0].value.toFixed(1)}%`, centerX, centerY);
    } else if (elementId === 'disk-chart' && data.length > 0) {
        ctx.fillText(`${data[0].value.toFixed(1)}%`, centerX, centerY);
    }
}

// Форматирование размеров в байтах
function formatBytes(bytes) {
    if (bytes === 0) return '0 Б';
    
    const k = 1024;
    const sizes = ['Б', 'КБ', 'МБ', 'ГБ', 'ТБ'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Настройка обработчиков событий интерфейса
function setupStatisticsEventListeners() {
    // Кнопка обновления метрик
    const refreshBtn = document.getElementById('refresh-metrics-btn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', () => {
            loadSystemMetrics();
            loadProcessesMetrics();
        });
    }
    
    // Кнопка экспорта метрик
    const exportBtn = document.getElementById('export-metrics-btn');
    if (exportBtn) {
        exportBtn.addEventListener('click', exportMetrics);
    }
}

// Экспорт метрик
function exportMetrics() {
    const data = {
        system: systemMetrics,
        processes: processesMetrics,
        timestamp: new Date().toISOString()
    };
    
    const json = JSON.stringify(data, null, 2);
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = `metrics-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    showToast('Метрики успешно экспортированы', 'success');
}

// Очистка ресурсов при переключении страницы
function cleanupStatisticsModule() {
    console.log('Очистка модуля статистики и метрик');
    
    if (metricsRefreshInterval) {
        clearInterval(metricsRefreshInterval);
        metricsRefreshInterval = null;
    }
    
    if (metricsSocket) {
        metricsSocket.close();
        metricsSocket = null;
    }
}

// Экспорт функций для глобального использования
window.initStatisticsModule = initStatisticsModule;
window.cleanupStatisticsModule = cleanupStatisticsModule;
window.loadSystemMetrics = loadSystemMetrics;
window.loadProcessesMetrics = loadProcessesMetrics;
window.renderSystemMetrics = renderSystemMetrics;
window.renderProcessesMetrics = renderProcessesMetrics;
window.updateSystemCharts = updateSystemCharts;
window.formatBytes = formatBytes;
window.exportMetrics = exportMetrics;