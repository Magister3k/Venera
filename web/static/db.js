/* ========================================
   Модуль управления базой данных
   для веб-интерфейса системы Venera
   ======================================== */

// Глобальные переменные
let dbTables = []; // Список таблиц базы данных
let dbConnection = null; // WebSocket соединение для базы данных
let currentTable = null; // Текущая выбранная таблица
let currentPage = 1; // Текущая страница
let pageSize = 50; // Размер страницы

// Инициализация модуля базы данных
function initDatabaseModule() {
    console.log('Инициализация модуля управления базой данных');
    
    // Подключение к WebSocket для получения обновлений в реальном времени
    connectDatabaseWebSocket();
    
    // Загрузка списка таблиц при инициализации
    loadDatabaseTables();
    
    // Настройка обработчиков событий интерфейса
    setupDatabaseEventListeners();
}

// Подключение к WebSocket для базы данных
function connectDatabaseWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/db`;
    
    dbConnection = new WebSocket(wsUrl);
    
    dbConnection.onopen = function() {
        console.log('WebSocket подключение к базе данных установлено');
        showToast('Подключение к WebSocket для базы данных установлено', 'success');
    };
    
    dbConnection.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleDatabaseWebSocketMessage(data);
        } catch (error) {
            console.error('Ошибка при обработке WebSocket сообщения для базы данных:', error);
        }
    };
    
    dbConnection.onclose = function() {
        console.log('WebSocket подключение к базе данных закрыто');
        // Попытка переподключения через 5 секунд
        setTimeout(connectDatabaseWebSocket, 5000);
    };
    
    dbConnection.onerror = function(error) {
        console.error('Ошибка WebSocket подключения к базе данных:', error);
        showToast('Ошибка подключения к WebSocket для базы данных', 'error');
    };
}

// Обработка WebSocket сообщений для базы данных
function handleDatabaseWebSocketMessage(data) {
    switch (data.action) {
        case 'tables_update':
            // Обновление списка таблиц
            loadDatabaseTables();
            break;
        case 'records_update':
            // Обновление записей в текущей таблице
            if (currentTable) {
                loadTableRecords(currentTable.name, currentPage);
            }
            break;
        case 'query_result':
            // Результат выполнения запроса
            renderQueryResult(data.result);
            break;
        default:
            console.log('Неизвестное действие WebSocket для базы данных:', data.action);
    }
}

// Загрузка списка таблиц базы данных
function loadDatabaseTables() {
    fetch('/api/db/tables')
        .then(response => {
            if (!response.ok) {
                throw new Error(`Ошибка загрузки таблиц базы данных: ${response.statusText}`);
            }
            return response.json();
        })
        .then(data => {
            dbTables = data.tables || [];
            renderDatabaseTables(dbTables);
        })
        .catch(error => {
            console.error('Ошибка при загрузке таблиц базы данных:', error);
            showToast(`Ошибка загрузки таблиц базы данных: ${error.message}`, 'error');
        });
}

// Отрисовка списка таблиц базы данных
function renderDatabaseTables(tables) {
    const container = document.getElementById('db-tables-container');
    
    if (!container) {
        console.error('Контейнер для таблиц базы данных не найден');
        return;
    }
    
    if (tables.length === 0) {
        container.innerHTML = `
            <div class="card">
                <div class="card-body">
                    <p class="text-center text-secondary">Нет доступных таблиц в базе данных</p>
                </div>
            </div>
        `;
        return;
    }
    
    // Сортировка таблиц по имени
    tables.sort((a, b) => a.name.localeCompare(b.name));
    
    // Формирование HTML карточек таблиц
    let html = '<div class="card-grid">';
    
    tables.forEach(table => {
        html += createTableCard(table);
    });
    
    html += '</div>';
    container.innerHTML = html;
}

// Создание HTML карточки таблицы
function createTableCard(table) {
    return `
        <div class="card db-table-card" data-table-name="${table.name}">
            <div class="card-header">
                <h3>${escapeHtml(table.name)}</h3>
                <span class="badge bg-primary">${table.count || 0} записей</span>
            </div>
            <div class="card-body">
                <div class="form-group">
                    <label class="form-label">Столбцов</label>
                    <div class="form-control" style="background: transparent;">${table.columns || 0}</div>
                </div>
                <div class="form-group">
                    <label class="form-label">Размер</label>
                    <div class="form-control" style="background: transparent;">${formatBytes(table.size || 0)}</div>
                </div>
                <div class="form-group">
                    <label class="form-label">Последнее обновление</label>
                    <div class="form-control" style="background: transparent;">${formatDateTime(table.lastUpdate)}</div>
                </div>
                <div class="card-footer">
                    <button class="btn btn-primary btn-sm" onclick="viewTableRecords('${table.name}')">
                        <span>📋</span> Просмотреть
                    </button>
                    <button class="btn btn-secondary btn-sm" onclick="exportTable('${table.name}')">
                        <span>📥</span> Экспорт
                    </button>
                    <button class="btn btn-danger btn-sm" onclick="dropTable('${table.name}')">
                        <span>🗑️</span> Удалить
                    </button>
                </div>
            </div>
        </div>
    `;
}

// Просмотр записей таблицы
function viewTableRecords(tableName) {
    currentTable = dbTables.find(t => t.name === tableName);
    
    if (!currentTable) {
        showToast('Таблица не найдена', 'error');
        return;
    }
    
    // Загрузка записей
    loadTableRecords(tableName, 1);
    
    // Показать контейнер для просмотра записей
    document.getElementById('db-records-container').style.display = 'block';
    document.getElementById('db-query-container').style.display = 'none';
    document.getElementById('db-tables-container').style.display = 'none';
}

// Загрузка записей таблицы
function loadTableRecords(tableName, page) {
    currentPage = page;
    
    const params = new URLSearchParams({
        table: tableName,
        page: page,
        limit: pageSize
    });
    
    fetch(`/api/db/records?${params.toString()}`)
        .then(response => {
            if (!response.ok) {
                throw new Error(`Ошибка загрузки записей таблицы: ${response.statusText}`);
            }
            return response.json();
        })
        .then(data => {
            renderTableRecords(tableName, data.records, data.total, page, data.pageSize);
        })
        .catch(error => {
            console.error('Ошибка при загрузке записей таблицы:', error);
            showToast(`Ошибка загрузки записей таблицы: ${error.message}`, 'error');
        });
}

// Отрисовка записей таблицы
function renderTableRecords(tableName, records, total, page, pageSize) {
    const container = document.getElementById('db-records-container');
    
    if (!container) {
        return;
    }
    
    let html = `
        <div class="card">
            <div class="card-header">
                <div class="d-flex justify-content-between align-items-center">
                    <h3>Записи таблицы: ${escapeHtml(tableName)}</h3>
                    <div>
                        <button class="btn btn-primary btn-sm" onclick="refreshTableRecords('${tableName}')">
                            <span>🔄</span> Обновить
                        </button>
                        <button class="btn btn-secondary btn-sm" onclick="exportTable('${tableName}')">
                            <span>📥</span> Экспорт
                        </button>
                    </div>
                </div>
            </div>
            <div class="card-body">
                <div class="table-responsive">
                    <table class="table table-striped">
                        <thead>
                            <tr>
                                <th>#</th>
    `;
    
    // Добавление заголовков столбцов
    if (records.length > 0) {
        const columns = Object.keys(records[0]);
        columns.forEach(column => {
            html += `<th>${escapeHtml(column)}</th>`;
        });
    }
    
    html += `
                            </tr>
                        </thead>
                        <tbody>
    `;
    
    // Добавление строк с записями
    if (records.length === 0) {
        html += `
                            <tr>
                                <td colspan="100%" class="text-center text-secondary">Нет доступных записей</td>
                            </tr>
        `;
    } else {
        records.forEach((record, index) => {
            html += `<tr data-record-id="${record.id || index}">`;
            html += `<td>${(page - 1) * pageSize + index + 1}</td>`;
            
            const columns = Object.keys(record);
            columns.forEach(column => {
                const value = record[column];
                const formattedValue = typeof value === 'object' 
                    ? JSON.stringify(value) 
                    : String(value);
                html += `<td>${escapeHtml(formattedValue)}</td>`;
            });
            
            html += `</tr>`;
        });
    }
    
    html += `
                        </tbody>
                    </table>
                </div>
    `;
    
    // Добавление пагинации
    const totalPages = Math.ceil(total / pageSize);
    if (totalPages > 1) {
        html += `
                <div class="pagination-container">
                    <div class="pagination">
                        <button class="btn btn-secondary btn-sm ${page === 1 ? 'disabled' : ''}" 
                                onclick="loadTableRecords('${tableName}', ${page - 1})">
                            <span>←</span> Назад
                        </button>
                        <span class="pagination-info">
                            Страница ${page} из ${totalPages} (${total} записей)
                        </span>
                        <button class="btn btn-secondary btn-sm ${page === totalPages ? 'disabled' : ''}" 
                                onclick="loadTableRecords('${tableName}', ${page + 1})">
                            Вперед <span>→</span>
                        </button>
                    </div>
                </div>
        `;
    }
    
    html += `
            </div>
        </div>
    `;
    
    container.innerHTML = html;
}

// Обновление записей таблицы
function refreshTableRecords(tableName) {
    loadTableRecords(tableName, currentPage);
}

// Выполнение SQL-запроса
function executeQuery() {
    const queryInput = document.getElementById('sql-query-input');
    
    if (!queryInput) {
        return;
    }
    
    const query = queryInput.value.trim();
    
    if (!query) {
        showToast('Введите SQL-запрос', 'warning');
        return;
    }
    
    fetch('/api/db/query', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ query: query })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка выполнения запроса: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        renderQueryResult(data);
    })
    .catch(error => {
        console.error('Ошибка при выполнении SQL-запроса:', error);
        showToast(`Ошибка выполнения запроса: ${error.message}`, 'error');
    });
}

// Отрисовка результата запроса
function renderQueryResult(result) {
    const container = document.getElementById('db-query-container');
    
    if (!container) {
        return;
    }
    
    let html = `
        <div class="card">
            <div class="card-header">
                <h3>Результат запроса</h3>
            </div>
            <div class="card-body">
                <div class="form-group">
                    <label class="form-label">Строк затронуто</label>
                    <div class="form-control" style="background: transparent;">${result.affectedRows || 0}</div>
                </div>
    `;
    
    if (result.rows && result.rows.length > 0) {
        html += `
                <div class="table-responsive">
                    <table class="table table-striped">
                        <thead>
                            <tr>
                                <th>#</th>
        `;
        
        // Добавление заголовков столбцов
        const columns = Object.keys(result.rows[0]);
        columns.forEach(column => {
            html += `<th>${escapeHtml(column)}</th>`;
        });
        
        html += `
                            </tr>
                        </thead>
                        <tbody>
        `;
        
        // Добавление строк с результатами
        result.rows.forEach((row, index) => {
            html += `<tr>`;
            html += `<td>${index + 1}</td>`;
            
            columns.forEach(column => {
                const value = row[column];
                const formattedValue = typeof value === 'object' 
                    ? JSON.stringify(value) 
                    : String(value);
                html += `<td>${escapeHtml(formattedValue)}</td>`;
            });
            
            html += `</tr>`;
        });
        
        html += `
                        </tbody>
                    </table>
                </div>
        `;
    } else if (result.affectedRows > 0) {
        html += `
                <div class="alert alert-success">
                    Запрос успешно выполнен. Затронуто строк: ${result.affectedRows}
                </div>
        `;
    } else {
        html += `
                <p class="text-center text-secondary">Нет данных для отображения</p>
        `;
    }
    
    html += `
            </div>
        </div>
    `;
    
    container.innerHTML = html;
}

// Экспорт таблицы
function exportTable(tableName) {
    if (!confirm(`Вы уверены, что хотите экспортировать таблицу "${tableName}"?`)) {
        return;
    }
    
    const a = document.createElement('a');
    a.href = `/api/db/tables/${encodeURIComponent(tableName)}/export`;
    a.download = `${tableName}-${new Date().toISOString().split('T')[0]}.sql`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    
    showToast(`Таблица "${tableName}" экспортирована`, 'success');
}

// Удаление таблицы
function dropTable(tableName) {
    if (!confirm(`Вы уверены, что хотите удалить таблицу "${tableName}"? Это действие необратимо!`)) {
        return;
    }
    
    fetch(`/api/db/tables/${encodeURIComponent(tableName)}`, {
        method: 'DELETE'
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка удаления таблицы: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showToast(`Таблица "${tableName}" успешно удалена`, 'success');
        loadDatabaseTables();
    })
    .catch(error => {
        console.error('Ошибка при удалении таблицы:', error);
        showToast(`Ошибка удаления таблицы: ${error.message}`, 'error');
    });
}

// Настройка обработчиков событий интерфейса
function setupDatabaseEventListeners() {
    // Кнопка обновления списка таблиц
    const refreshTablesBtn = document.getElementById('refresh-db-tables-btn');
    if (refreshTablesBtn) {
        refreshTablesBtn.addEventListener('click', loadDatabaseTables);
    }
    
    // Кнопка выполнения SQL-запроса
    const executeQueryBtn = document.getElementById('execute-query-btn');
    if (executeQueryBtn) {
        executeQueryBtn.addEventListener('click', executeQuery);
    }
    
    // Обработка нажатия Enter в поле SQL-запроса
    const sqlQueryInput = document.getElementById('sql-query-input');
    if (sqlQueryInput) {
        sqlQueryInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
                executeQuery();
            }
        });
    }
    
    // Кнопка импорта данных
    const importBtn = document.getElementById('import-data-btn');
    if (importBtn) {
        importBtn.addEventListener('click', showImportModal);
    }
    
    // Кнопка создания новой таблицы
    const createTableBtn = document.getElementById('create-table-btn');
    if (createTableBtn) {
        createTableBtn.addEventListener('click', showCreateTableModal);
    }
}

// Показ модального окна импорта данных
function showImportModal() {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h2>Импорт данных в базу данных</h2>
                <button class="modal-close" onclick="this.closest('.modal').remove()">&times;</button>
            </div>
            <div class="modal-body">
                <form id="import-data-form" enctype="multipart/form-data">
                    <div class="form-group">
                        <label class="form-label required">Файл для импорта</label>
                        <input type="file" class="form-control" name="file" accept=".sql,.json,.csv" required>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Таблица назначения</label>
                        <select class="form-control" name="targetTable">
                            <option value="">Выберите таблицу</option>
                            ${dbTables.map(t => `<option value="${t.name}">${t.name}</option>`).join('')}
                        </select>
                    </div>
                </form>
            </div>
            <div class="modal-footer">
                <button class="btn btn-secondary" onclick="this.closest('.modal').remove()">Отмена</button>
                <button class="btn btn-success" onclick="submitImportData()">Импортировать</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    setTimeout(() => modal.classList.add('show'), 10);
}

// Отправка формы импорта данных
function submitImportData() {
    const form = document.getElementById('import-data-form');
    if (!form) {
        showToast('Форма не найдена', 'error');
        return;
    }
    
    const formData = new FormData(form);
    
    fetch('/api/db/import', {
        method: 'POST',
        body: formData
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка импорта данных: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showToast('Данные успешно импортированы', 'success');
        document.querySelector('.modal').remove();
        loadDatabaseTables();
    })
    .catch(error => {
        console.error('Ошибка при импорте данных:', error);
        showToast(`Ошибка импорта данных: ${error.message}`, 'error');
    });
}

// Показ модального окна создания новой таблицы
function showCreateTableModal() {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h2>Создание новой таблицы</h2>
                <button class="modal-close" onclick="this.closest('.modal').remove()">&times;</button>
            </div>
            <div class="modal-body">
                <form id="create-table-form">
                    <div class="form-group">
                        <label class="form-label required">Имя таблицы</label>
                        <input type="text" class="form-control" name="name" required placeholder="имя_таблицы">
                    </div>
                    <div class="form-group">
                        <label class="form-label">Описание</label>
                        <input type="text" class="form-control" name="description" placeholder="Описание таблицы">
                    </div>
                    <div class="form-group">
                        <label class="form-label required">Конфигурация столбцов (JSON)</label>
                        <textarea class="form-control" name="columns" style="height: 200px; font-family: monospace;" required placeholder='[
    {"name": "id", "type": "BIGINT", "primary": true, "auto_increment": true},
    {"name": "identifier", "type": "VARCHAR(255)", "nullable": false},
    {"name": "created_at", "type": "TIMESTAMP", "default": "CURRENT_TIMESTAMP"}
]'></textarea>
                    </div>
                </form>
            </div>
            <div class="modal-footer">
                <button class="btn btn-secondary" onclick="this.closest('.modal').remove()">Отмена</button>
                <button class="btn btn-success" onclick="submitCreateTable()">Создать</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    setTimeout(() => modal.classList.add('show'), 10);
}

// Отправка формы создания таблицы
function submitCreateTable() {
    const form = document.getElementById('create-table-form');
    if (!form) {
        showToast('Форма не найдена', 'error');
        return;
    }
    
    const formData = {
        name: form.elements.name.value,
        description: form.elements.description.value,
        columns: JSON.parse(form.elements.columns.value)
    };
    
    fetch('/api/db/tables', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка создания таблицы: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showToast('Таблица успешно создана', 'success');
        document.querySelector('.modal').remove();
        loadDatabaseTables();
    })
    .catch(error => {
        console.error('Ошибка при создании таблицы:', error);
        showToast(`Ошибка создания таблицы: ${error.message}`, 'error');
    });
}

// Очистка ресурсов при переключении страницы
function cleanupDatabaseModule() {
    console.log('Очистка модуля управления базой данных');
    
    if (dbConnection) {
        dbConnection.close();
        dbConnection = null;
    }
}

// Экспорт функций для глобального использования
window.initDatabaseModule = initDatabaseModule;
window.cleanupDatabaseModule = cleanupDatabaseModule;
window.loadDatabaseTables = loadDatabaseTables;
window.renderDatabaseTables = renderDatabaseTables;
window.viewTableRecords = viewTableRecords;
window.loadTableRecords = loadTableRecords;
window.renderTableRecords = renderTableRecords;
window.refreshTableRecords = refreshTableRecords;
window.executeQuery = executeQuery;
window.renderQueryResult = renderQueryResult;
window.exportTable = exportTable;
window.dropTable = dropTable;