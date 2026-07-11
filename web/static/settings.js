/**
 * Модуль управления настройками системы для веб-интерфейса Venera
 * Предоставляет функции для просмотра и редактирования конфигурации системы
 * 
 * Основные функции:
 * - Подключение к WebSocket для получения обновлений настроек в реальном времени
 * - REST API взаимодействие с сервером: GET /api/settings, POST /api/settings, GET /api/settings/:key
 * - Валидация форм на клиенте
 * - Поддержка темной/светлой темы через CSS переменные
 * - Интеграция с существующими компонентами React UI
 * 
 * @module settings
 */

// Глобальные переменные для хранения состояния модуля
let settingsWebSocket = null;
let settingsData = {};
let isWebSocketConnected = false;

/**
 * Инициализация модуля настроек
 * Подключает WebSocket соединение и загружает текущие настройки
 * 
 * @returns {Promise<void>}
 */
async function initSettingsModule() {
    console.log('[Settings] Инициализация модуля настроек');
    
    // Подключение к WebSocket для получения обновлений настроек
    await connectSettingsWebSocket();
    
    // Загрузка текущих настроек из API
    await loadSettings();
    
    // Отрисовка формы настроек
    renderSettingsForm();
    
    console.log('[Settings] Модуль настроек успешно инициализирован');
}

/**
 * Подключение к WebSocket для получения обновлений настроек в реальном времени
 * 
 * @returns {Promise<void>}
 */
async function connectSettingsWebSocket() {
    try {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/settings`;
        
        settingsWebSocket = new WebSocket(wsUrl);
        
        settingsWebSocket.onopen = () => {
            console.log('[Settings] WebSocket соединение установлено');
            isWebSocketConnected = true;
        };
        
        settingsWebSocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                
                if (message.type === 'settings_update') {
                    console.log('[Settings] Получено обновление настроек', message.data);
                    settingsData = message.data;
                    updateSettingsForm();
                }
            } catch (error) {
                console.error('[Settings] Ошибка при обработке WebSocket сообщения:', error);
            }
        };
        
        settingsWebSocket.onclose = () => {
            console.log('[Settings] WebSocket соединение закрыто');
            isWebSocketConnected = false;
            
            // Попытка переподключения через 5 секунд
            setTimeout(connectSettingsWebSocket, 5000);
        };
        
        settingsWebSocket.onerror = (error) => {
            console.error('[Settings] Ошибка WebSocket соединения:', error);
        };
    } catch (error) {
        console.error('[Settings] Ошибка подключения к WebSocket:', error);
        isWebSocketConnected = false;
    }
}

/**
 * Загрузка настроек из REST API
 * 
 * @returns {Promise<void>}
 */
async function loadSettings() {
    try {
        const response = await fetch('/api/settings');
        
        if (!response.ok) {
            throw new Error(`Ошибка загрузки настроек: ${response.status} ${response.statusText}`);
        }
        
        settingsData = await response.json();
        console.log('[Settings] Настройки успешно загружены', settingsData);
    } catch (error) {
        console.error('[Settings] Ошибка при загрузке настроек:', error);
        // Использовать настройки по умолчанию при ошибке
        settingsData = getDefaultSettings();
    }
}

/**
 * Получение настроек по умолчанию
 * 
 * @returns {Object} Объект с настройками по умолчанию
 */
function getDefaultSettings() {
    return {
        database: {
            host: 'localhost',
            port: 5432,
            name: 'venera',
            user: 'postgres',
            password: '',
            poolSize: 10,
            connectionTimeout: 30,
            sslMode: 'disable'
        },
        server: {
            host: '0.0.0.0',
            port: 8080,
            readTimeout: 15,
            writeTimeout: 15,
            maxHeaderBytes: 1048576
        },
        logging: {
            level: 'info',
            format: 'text',
            path: './logs',
            maxSize: 100,
            maxBackups: 10,
            maxAge: 30,
            compress: true
        },
        notifications: {
            enabled: false,
            type: 'email',
            email: {
                host: 'smtp.example.com',
                port: 587,
                user: 'noreply@example.com',
                password: '',
                from: 'noreply@example.com',
                to: 'admin@example.com'
            },
            webhook: {
                url: '',
                method: 'POST',
                headers: {}
            }
        },
        processes: {
            timeout: 60,
            priority: 'normal',
            maxRetries: 3,
            retryDelay: 5,
            maxConcurrent: 10,
            queueSize: 1000
        }
    };
}

/**
 * Отрисовка формы настроек
 * Создает HTML элементы для редактирования настроек
 * 
 * @returns {void}
 */
function renderSettingsForm() {
    const container = document.getElementById('settings-container');
    
    if (!container) {
        console.error('[Settings] Элемент settings-container не найден');
        return;
    }
    
    container.innerHTML = `
        <div class="settings-form">
            <h2 class="settings-title">Настройки системы</h2>
            
            <div class="settings-tabs">
                <button class="tab-btn active" data-tab="database">База данных</button>
                <button class="tab-btn" data-tab="server">Сервер</button>
                <button class="tab-btn" data-tab="logging">Логирование</button>
                <button class="tab-btn" data-tab="notifications">Уведомления</button>
                <button class="tab-btn" data-tab="processes">Процессы</button>
            </div>
            
            <div class="settings-content">
                ${renderDatabaseSection()}
                ${renderServerSection()}
                ${renderLoggingSection()}
                ${renderNotificationsSection()}
                ${renderProcessesSection()}
            </div>
            
            <div class="settings-actions">
                <button id="save-settings-btn" class="btn btn-primary">Сохранить</button>
                <button id="reset-settings-btn" class="btn btn-secondary">Сбросить</button>
                <button id="test-settings-btn" class="btn btn-secondary">Проверить</button>
            </div>
        </div>
    `;
    
    // Добавление обработчиков событий
    addSettingsEventListeners();
}

/**
 * Отрисовка секции настроек базы данных
 * 
 * @returns {string} HTML код секции
 */
function renderDatabaseSection() {
    const data = settingsData.database || getDefaultSettings().database;
    
    return `
        <div class="settings-section database-section" data-section="database">
            <h3 class="section-title">Параметры подключения к базе данных</h3>
            
            <div class="form-group">
                <label for="db-host" class="form-label">Хост базы данных</label>
                <input type="text" id="db-host" class="form-input" value="${data.host}" placeholder="localhost" required>
                <span class="form-hint">Адрес сервера базы данных PostgreSQL</span>
            </div>
            
            <div class="form-group">
                <label for="db-port" class="form-label">Порт</label>
                <input type="number" id="db-port" class="form-input" value="${data.port}" min="1" max="65535" required>
                <span class="form-hint">Порт сервера PostgreSQL (по умолчанию: 5432)</span>
            </div>
            
            <div class="form-group">
                <label for="db-name" class="form-label">Имя базы данных</label>
                <input type="text" id="db-name" class="form-input" value="${data.name}" placeholder="venera" required>
                <span class="form-hint">Название базы данных для хранения идентификаторов</span>
            </div>
            
            <div class="form-group">
                <label for="db-user" class="form-label">Пользователь</label>
                <input type="text" id="db-user" class="form-input" value="${data.user}" placeholder="postgres" required>
                <span class="form-hint">Имя пользователя для подключения к базе данных</span>
            </div>
            
            <div class="form-group">
                <label for="db-password" class="form-label">Пароль</label>
                <input type="password" id="db-password" class="form-input" value="${data.password}" placeholder="••••••">
                <span class="form-hint">Пароль для подключения к базе данных</span>
            </div>
            
            <div class="form-group">
                <label for="db-pool-size" class="form-label">Размер пула соединений</label>
                <input type="number" id="db-pool-size" class="form-input" value="${data.poolSize}" min="1" max="100" required>
                <span class="form-hint">Максимальное количество одновременных соединений (по умолчанию: 10)</span>
            </div>
            
            <div class="form-group">
                <label for="db-connection-timeout" class="form-label">Таймаут подключения (сек)</label>
                <input type="number" id="db-connection-timeout" class="form-input" value="${data.connectionTimeout}" min="1" max="300" required>
                <span class="form-hint">Время ожидания подключения к серверу (по умолчанию: 30)</span>
            </div>
            
            <div class="form-group">
                <label for="db-ssl-mode" class="form-label">Режим SSL</label>
                <select id="db-ssl-mode" class="form-select">
                    <option value="disable" ${data.sslMode === 'disable' ? 'selected' : ''}>disable</option>
                    <option value="allow" ${data.sslMode === 'allow' ? 'selected' : ''}>allow</option>
                    <option value="prefer" ${data.sslMode === 'prefer' ? 'selected' : ''}>prefer</option>
                    <option value="require" ${data.sslMode === 'require' ? 'selected' : ''}>require</option>
                    <option value="verify-ca" ${data.sslMode === 'verify-ca' ? 'selected' : ''}>verify-ca</option>
                    <option value="verify-full" ${data.sslMode === 'verify-full' ? 'selected' : ''}>verify-full</option>
                </select>
                <span class="form-hint">Режим использования SSL для подключения</span>
            </div>
        </div>
    `;
}

/**
 * Отрисовка секции настроек сервера
 * 
 * @returns {string} HTML код секции
 */
function renderServerSection() {
    const data = settingsData.server || getDefaultSettings().server;
    
    return `
        <div class="settings-section server-section" data-section="server" style="display: none;">
            <h3 class="section-title">Параметры сервера</h3>
            
            <div class="form-group">
                <label for="server-host" class="form-label">Хост сервера</label>
                <input type="text" id="server-host" class="form-input" value="${data.host}" placeholder="0.0.0.0" required>
                <span class="form-hint">IP-адрес для прослушивания входящих соединений</span>
            </div>
            
            <div class="form-group">
                <label for="server-port" class="form-label">Порт</label>
                <input type="number" id="server-port" class="form-input" value="${data.port}" min="1" max="65535" required>
                <span class="form-hint">Порт для прослушивания (по умолчанию: 8080)</span>
            </div>
            
            <div class="form-group">
                <label for="server-read-timeout" class="form-label">Таймаут чтения (сек)</label>
                <input type="number" id="server-read-timeout" class="form-input" value="${data.readTimeout}" min="1" max="300" required>
                <span class="form-hint">Максимальное время чтения запроса (по умолчанию: 15)</span>
            </div>
            
            <div class="form-group">
                <label for="server-write-timeout" class="form-label">Таймаут записи (сек)</label>
                <input type="number" id="server-write-timeout" class="form-input" value="${data.writeTimeout}" min="1" max="300" required>
                <span class="form-hint">Максимальное время записи ответа (по умолчанию: 15)</span>
            </div>
            
            <div class="form-group">
                <label for="server-max-header-bytes" class="form-label">Макс. размер заголовков (байт)</label>
                <input type="number" id="server-max-header-bytes" class="form-input" value="${data.maxHeaderBytes}" min="0" max="10485760" required>
                <span class="form-hint">Максимальный размер HTTP-заголовков (по умолчанию: 1048576)</span>
            </div>
        </div>
    `;
}

/**
 * Отрисовка секции настроек логирования
 * 
 * @returns {string} HTML код секции
 */
function renderLoggingSection() {
    const data = settingsData.logging || getDefaultSettings().logging;
    
    const levels = ['debug', 'info', 'warn', 'error', 'fatal', 'panic'];
    const formats = ['text', 'json'];
    
    return `
        <div class="settings-section logging-section" data-section="logging" style="display: none;">
            <h3 class="section-title">Настройки логирования</h3>
            
            <div class="form-group">
                <label for="log-level" class="form-label">Уровень логирования</label>
                <select id="log-level" class="form-select">
                    ${levels.map(level => `<option value="${level}" ${data.level === level ? 'selected' : ''}>${level}</option>`).join('')}
                </select>
                <span class="form-hint">Минимальный уровень логов (по умолчанию: info)</span>
            </div>
            
            <div class="form-group">
                <label for="log-format" class="form-label">Формат логов</label>
                <select id="log-format" class="form-select">
                    ${formats.map(format => `<option value="${format}" ${data.format === format ? 'selected' : ''}>${format}</option>`).join('')}
                </select>
                <span class="form-hint">Формат вывода логов (по умолчанию: text)</span>
            </div>
            
            <div class="form-group">
                <label for="log-path" class="form-label">Путь к логам</label>
                <input type="text" id="log-path" class="form-input" value="${data.path}" placeholder="./logs" required>
                <span class="form-hint">Директория для хранения лог-файлов</span>
            </div>
            
            <div class="form-group">
                <label for="log-max-size" class="form-label">Макс. размер файла (МБ)</label>
                <input type="number" id="log-max-size" class="form-input" value="${data.maxSize}" min="1" max="1000" required>
                <span class="form-hint">Максимальный размер лог-файла перед ротацией (по умолчанию: 100)</span>
            </div>
            
            <div class="form-group">
                <label for="log-max-backups" class="form-label">Количество резервных копий</label>
                <input type="number" id="log-max-backups" class="form-input" value="${data.maxBackups}" min="0" max="100" required>
                <span class="form-hint">Количество архивных файлов для хранения (по умолчанию: 10)</span>
            </div>
            
            <div class="form-group">
                <label for="log-max-age" class="form-label">Макс. возраст файла (дни)</label>
                <input type="number" id="log-max-age" class="form-input" value="${data.maxAge}" min="1" max="365" required>
                <span class="form-hint">Максимальное время хранения логов (по умолчанию: 30)</span>
            </div>
            
            <div class="form-group">
                <label class="form-label">
                    <input type="checkbox" id="log-compress" ${data.compress ? 'checked' : ''}>
                    Сжимать архивные логи
                </label>
                <span class="form-hint">Сжимать старые лог-файлы в gzip формат</span>
            </div>
        </div>
    `;
}

/**
 * Отрисовка секции настроек уведомлений
 * 
 * @returns {string} HTML код секции
 */
function renderNotificationsSection() {
    const data = settingsData.notifications || getDefaultSettings().notifications;
    
    return `
        <div class="settings-section notifications-section" data-section="notifications" style="display: none;">
            <h3 class="section-title">Настройки уведомлений</h3>
            
            <div class="form-group">
                <label class="form-label">
                    <input type="checkbox" id="notif-enabled" ${data.enabled ? 'checked' : ''}>
                    Включить уведомления
                </label>
                <span class="form-hint">Отправлять уведомления о событиях системы</span>
            </div>
            
            <div class="form-group">
                <label for="notif-type" class="form-label">Тип уведомлений</label>
                <select id="notif-type" class="form-select">
                    <option value="email" ${data.type === 'email' ? 'selected' : ''}>Email</option>
                    <option value="webhook" ${data.type === 'webhook' ? 'selected' : ''}>Webhook</option>
                </select>
                <span class="form-hint">Метод доставки уведомлений</span>
            </div>
            
            <div class="notif-email-config">
                <h4 class="subsection-title">Email настройки</h4>
                
                <div class="form-group">
                    <label for="notif-email-host" class="form-label">SMTP хост</label>
                    <input type="text" id="notif-email-host" class="form-input" value="${data.email.host}" placeholder="smtp.example.com" required>
                    <span class="form-hint">SMTP-сервер для отправки писем</span>
                </div>
                
                <div class="form-group">
                    <label for="notif-email-port" class="form-label">SMTP порт</label>
                    <input type="number" id="notif-email-port" class="form-input" value="${data.email.port}" min="1" max="65535" required>
                    <span class="form-hint">Порт SMTP-сервера (обычно: 587 или 465)</span>
                </div>
                
                <div class="form-group">
                    <label for="notif-email-user" class="form-label">Пользователь</label>
                    <input type="text" id="notif-email-user" class="form-input" value="${data.email.user}" placeholder="noreply@example.com" required>
                    <span class="form-hint">Имя пользователя для SMTP аутентификации</span>
                </div>
                
                <div class="form-group">
                    <label for="notif-email-password" class="form-label">Пароль</label>
                    <input type="password" id="notif-email-password" class="form-input" value="${data.email.password}" placeholder="••••••">
                    <span class="form-hint">Пароль для SMTP аутентификации</span>
                </div>
                
                <div class="form-group">
                    <label for="notif-email-from" class="form-label">Отправитель</label>
                    <input type="email" id="notif-email-from" class="form-input" value="${data.email.from}" placeholder="noreply@example.com" required>
                    <span class="form-hint">Email отправителя</span>
                </div>
                
                <div class="form-group">
                    <label for="notif-email-to" class="form-label">Получатели (через запятую)</label>
                    <input type="text" id="notif-email-to" class="form-input" value="${data.email.to}" placeholder="admin@example.com" required>
                    <span class="form-hint">Email получателей уведомлений</span>
                </div>
            </div>
            
            <div class="notif-webhook-config" style="display: none;">
                <h4 class="subsection-title">Webhook настройки</h4>
                
                <div class="form-group">
                    <label for="notif-webhook-url" class="form-label">URL Webhook</label>
                    <input type="url" id="notif-webhook-url" class="form-input" value="${data.webhook.url}" placeholder="https://example.com/webhook" required>
                    <span class="form-hint">URL endpoint для отправки уведомлений</span>
                </div>
                
                <div class="form-group">
                    <label for="notif-webhook-method" class="form-label">HTTP метод</label>
                    <select id="notif-webhook-method" class="form-select">
                        <option value="POST" ${data.webhook.method === 'POST' ? 'selected' : ''}>POST</option>
                        <option value="PUT" ${data.webhook.method === 'PUT' ? 'selected' : ''}>PUT</option>
                        <option value="PATCH" ${data.webhook.method === 'PATCH' ? 'selected' : ''}>PATCH</option>
                    </select>
                    <span class="form-hint">HTTP метод для отправки уведомлений</span>
                </div>
            </div>
        </div>
    `;
}

/**
 * Отрисовка секции настроек процессов
 * 
 * @returns {string} HTML код секции
 */
function renderProcessesSection() {
    const data = settingsData.processes || getDefaultSettings().processes;
    
    return `
        <div class="settings-section processes-section" data-section="processes" style="display: none;">
            <h3 class="section-title">Настройки процессов</h3>
            
            <div class="form-group">
                <label for="proc-timeout" class="form-label">Таймаут процесса (сек)</label>
                <input type="number" id="proc-timeout" class="form-input" value="${data.timeout}" min="1" max="3600" required>
                <span class="form-hint">Максимальное время выполнения процесса (по умолчанию: 60)</span>
            </div>
            
            <div class="form-group">
                <label for="proc-priority" class="form-label">Приоритет</label>
                <select id="proc-priority" class="form-select">
                    <option value="low" ${data.priority === 'low' ? 'selected' : ''}>low</option>
                    <option value="normal" ${data.priority === 'normal' ? 'selected' : ''}>normal</option>
                    <option value="high" ${data.priority === 'high' ? 'selected' : ''}>high</option>
                </select>
                <span class="form-hint">Приоритет процессов обработки (по умолчанию: normal)</span>
            </div>
            
            <div class="form-group">
                <label for="proc-max-retries" class="form-label">Макс. количество повторов</label>
                <input type="number" id="proc-max-retries" class="form-input" value="${data.maxRetries}" min="0" max="10" required>
                <span class="form-hint">Количество попыток повторного выполнения при ошибке (по умолчанию: 3)</span>
            </div>
            
            <div class="form-group">
                <label for="proc-retry-delay" class="form-label">Задержка между повторами (сек)</label>
                <input type="number" id="proc-retry-delay" class="form-input" value="${data.retryDelay}" min="1" max="60" required>
                <span class="form-hint">Время ожидания перед повторной попыткой (по умолчанию: 5)</span>
            </div>
            
            <div class="form-group">
                <label for="proc-max-concurrent" class="form-label">Макс..concurrent процессы</label>
                <input type="number" id="proc-max-concurrent" class="form-input" value="${data.maxConcurrent}" min="1" max="100" required>
                <span class="form-hint">Максимальное количество одновременно выполняемых процессов (по умолчанию: 10)</span>
            </div>
            
            <div class="form-group">
                <label for="proc-queue-size" class="form-label">Размер очереди</label>
                <input type="number" id="proc-queue-size" class="form-input" value="${data.queueSize}" min="1" max="10000" required>
                <span class="form-hint">Максимальный размер очереди задач (по умолчанию: 1000)</span>
            </div>
        </div>
    `;
}

/**
 * Добавление обработчиков событий для формы настроек
 * 
 * @returns {void}
 */
function addSettingsEventListeners() {
    // П��реключение вкладок
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const tab = e.target.dataset.tab;
            
            // Скрыть все секции
            document.querySelectorAll('.settings-section').forEach(section => {
                section.style.display = 'none';
            });
            
            // Показать выбранную секцию
            document.querySelector(`.${tab}-section`).style.display = 'block';
            
            // Обновить активную вкладку
            document.querySelectorAll('.tab-btn').forEach(b => {
                b.classList.remove('active');
            });
            btn.classList.add('active');
        });
    });
    
    // Кнопка сохранения настроек
    document.getElementById('save-settings-btn').addEventListener('click', saveSettings);
    
    // Кнопка сброса настроек
    document.getElementById('reset-settings-btn').addEventListener('click', resetSettings);
    
    // Кнопка проверки настроек
    document.getElementById('test-settings-btn').addEventListener('click', testSettings);
    
    // Переключение типа уведомлений
    document.getElementById('notif-type').addEventListener('change', (e) => {
        if (e.target.value === 'email') {
            document.querySelector('.notif-email-config').style.display = 'block';
            document.querySelector('.notif-webhook-config').style.display = 'none';
        } else {
            document.querySelector('.notif-email-config').style.display = 'none';
            document.querySelector('.notif-webhook-config').style.display = 'block';
        }
    });
}

/**
 * Сохранение настроек через REST API
 * 
 * @returns {Promise<void>}
 */
async function saveSettings() {
    try {
        const settings = collectSettingsFromForm();
        
        const response = await fetch('/api/settings', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(settings)
        });
        
        if (!response.ok) {
            throw new Error(`Ошибка сохранения настроек: ${response.status} ${response.statusText}`);
        }
        
        settingsData = settings;
        
        // Отобразить сообщение об успехе
        showNotification('Настройки успешно сохранены', 'success');
        
        console.log('[Settings] Настройки успешно сохранены', settings);
    } catch (error) {
        console.error('[Settings] Ошибка при сохранении настроек:', error);
        showNotification('Ошибка при сохранении настроек', 'error');
    }
}

/**
 * Сбор настроек из формы
 * 
 * @returns {Object} Объект с настройками
 */
function collectSettingsFromForm() {
    return {
        database: {
            host: document.getElementById('db-host').value,
            port: parseInt(document.getElementById('db-port').value),
            name: document.getElementById('db-name').value,
            user: document.getElementById('db-user').value,
            password: document.getElementById('db-password').value,
            poolSize: parseInt(document.getElementById('db-pool-size').value),
            connectionTimeout: parseInt(document.getElementById('db-connection-timeout').value),
            sslMode: document.getElementById('db-ssl-mode').value
        },
        server: {
            host: document.getElementById('server-host').value,
            port: parseInt(document.getElementById('server-port').value),
            readTimeout: parseInt(document.getElementById('server-read-timeout').value),
            writeTimeout: parseInt(document.getElementById('server-write-timeout').value),
            maxHeaderBytes: parseInt(document.getElementById('server-max-header-bytes').value)
        },
        logging: {
            level: document.getElementById('log-level').value,
            format: document.getElementById('log-format').value,
            path: document.getElementById('log-path').value,
            maxSize: parseInt(document.getElementById('log-max-size').value),
            maxBackups: parseInt(document.getElementById('log-max-backups').value),
            maxAge: parseInt(document.getElementById('log-max-age').value),
            compress: document.getElementById('log-compress').checked
        },
        notifications: {
            enabled: document.getElementById('notif-enabled').checked,
            type: document.getElementById('notif-type').value,
            email: {
                host: document.getElementById('notif-email-host').value,
                port: parseInt(document.getElementById('notif-email-port').value),
                user: document.getElementById('notif-email-user').value,
                password: document.getElementById('notif-email-password').value,
                from: document.getElementById('notif-email-from').value,
                to: document.getElementById('notif-email-to').value
            },
            webhook: {
                url: document.getElementById('notif-webhook-url').value,
                method: document.getElementById('notif-webhook-method').value,
                headers: {}
            }
        },
        processes: {
            timeout: parseInt(document.getElementById('proc-timeout').value),
            priority: document.getElementById('proc-priority').value,
            maxRetries: parseInt(document.getElementById('proc-max-retries').value),
            retryDelay: parseInt(document.getElementById('proc-retry-delay').value),
            maxConcurrent: parseInt(document.getElementById('proc-max-concurrent').value),
            queueSize: parseInt(document.getElementById('proc-queue-size').value)
        }
    };
}

/**
 * Обновление формы настроек
 * 
 * @returns {void}
 */
function updateSettingsForm() {
    // Скрыть все секции
    document.querySelectorAll('.settings-section').forEach(section => {
        section.style.display = 'none';
    });
    
    // Показать активную секцию
    const activeTab = document.querySelector('.tab-btn.active');
    if (activeTab) {
        const tab = activeTab.dataset.tab;
        document.querySelector(`.${tab}-section`).style.display = 'block';
    }
}

/**
 * Сброс настроек к значениям по умолчанию
 * 
 * @returns {void}
 */
function resetSettings() {
    if (!confirm('Вы уверены, что хотите сбросить настройки к значениям по умолчанию?')) {
        return;
    }
    
    settingsData = getDefaultSettings();
    
    // Перерисовать форму
    renderSettingsForm();
    
    showNotification('Настройки сброшены к значениям по умолчанию', 'info');
}

/**
 * Проверка настроек подключения к базе данных
 * 
 * @returns {Promise<void>}
 */
async function testSettings() {
    try {
        const settings = collectSettingsFromForm();
        
        const response = await fetch('/api/settings/test', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(settings)
        });
        
        if (!response.ok) {
            throw new Error(`Ошибка проверки настроек: ${response.status} ${response.statusText}`);
        }
        
        const result = await response.json();
        
        if (result.success) {
            showNotification(result.message || 'Настройки успешно проверены', 'success');
        } else {
            showNotification(result.message || 'Ошибка проверки настроек', 'error');
        }
    } catch (error) {
        console.error('[Settings] Ошибка при проверке настроек:', error);
        showNotification('Ошибка при проверке настроек', 'error');
    }
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
        console.warn('[Settings] Элемент notifications-container не найден');
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
        initSettingsModule,
        connectSettingsWebSocket,
        loadSettings,
        getDefaultSettings,
        renderSettingsForm,
        saveSettings,
        resetSettings,
        testSettings,
        collectSettingsFromForm,
        updateSettingsForm,
        showNotification
    };
}
