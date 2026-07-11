// main.js - Главный JavaScript файл для веб-интерфейса Venera
//
// Этот файл обеспечивает:
// - Инициализацию приложения
// - Навигацию между разделами
// - Управление темой (темная/светлая)
// - Хэш-роутинг
// - Защиту от дублирования скриптов
//
// Использование:
// Подключается в index.html

$(document).ready(function() {
    'use strict';

    // Инициализация приложения
    initApp();
});

// Инициализация приложения
function initApp() {
    console.log('Venera Web UI initializing...');

    // Настройка хэш-роутинга
    setupHashRouting();

    // Инициализация темы
    initTheme();

    // Инициализация навигации
    initNavigation();

    // Инициализация горячих клавиш
    initHotkeys();

    // Загрузка конфигурации при старте
    loadInitialData();

    console.log('Venera Web UI initialized');
}

// Настройка хэш-роутинга
function setupHashRouting() {
    // Загрузка раздела из хэша при старте
    if (window.location.hash) {
        var page = window.location.hash.substring(1);
        loadPage(page);
    }

    // Обработка изменения хэша
    $(window).on('hashchange', function() {
        if (window.location.hash) {
            var page = window.location.hash.substring(1);
            loadPage(page);
        }
    });
}

// Инициализация темы
function initTheme() {
    // Получение сохраненной темы из localStorage
    var savedTheme = localStorage.getItem('theme');
    var theme = savedTheme || 'light';

    // Применение темы
    applyTheme(theme);

    // Обработчик переключения темы
    $('#theme-toggle').on('click', function() {
        var newTheme = theme === 'light' ? 'dark' : 'light';
        applyTheme(newTheme);
        localStorage.setItem('theme', newTheme);
        theme = newTheme;
    });
}

// Применение темы
function applyTheme(theme) {
    if (theme === 'dark') {
        document.documentElement.setAttribute('data-theme', 'dark');
        $('#theme-toggle').text('☀️');
    } else {
        document.documentElement.setAttribute('data-theme', 'light');
        $('#theme-toggle').text('🌙');
    }
}

// Инициализация навигации
function initNavigation() {
    $('.nav-item').on('click', function(e) {
        e.preventDefault();

        var page = $(this).data('page');
        if (!page) return;

        // Обновление активного элемента навигации
        $('.nav-item').removeClass('active');
        $(this).addClass('active');

        // Обновление хэша URL
        window.location.hash = page;

        // Загрузка страницы
        loadPage(page);
    });
}

// Загрузка страницы
function loadPage(page) {
    // Скрытие всех страниц
    $('.page-content').hide().removeClass('active');

    // Показ целевой страницы
    var pageElement = $('#page-' + page);
    if (pageElement.length) {
        pageElement.show().addClass('active');
    }

    // Инициализация конкретной страницы
    if (typeof pageInit === 'function') {
        pageInit(page);
    }

    console.log('Page loaded:', page);
}

// Инициализация горячих клавиш
function initHotkeys() {
    // Ctrl+F - поиск
    $(document).on('keydown', function(e) {
        if (e.ctrlKey && e.key === 'f') {
            e.preventDefault();
            var filterInput = $('#log-filter, #db-filter');
            if (filterInput.length) {
                filterInput.focus();
            }
        }

        // Esc - очистка поиска
        if (e.key === 'Escape') {
            var filterInput = $('#log-filter, #db-filter');
            if (filterInput.length) {
                filterInput.val('');
                // Вызов функции фильтрации
                applyFilter();
            }
        }
    });
}

// Загрузка начальных данных
function loadInitialData() {
    // Загрузка конфигурации
    $.ajax({
        url: '/api/config',
        method: 'GET',
        success: function(config) {
            console.log('Config loaded:', config);
            // Заполнение форм конфигурацией
            fillConfigForm(config);
        },
        error: function(xhr, status, error) {
            console.error('Failed to load config:', error);
        }
    });

    // Загрузка процессов
    $.ajax({
        url: '/api/processes',
        method: 'GET',
        success: function(processes) {
            console.log('Processes loaded:', processes);
            // Обновление списка процессов
            updateProcessesList(processes);
        },
        error: function(xhr, status, error) {
            console.error('Failed to load processes:', error);
        }
    });
}

// Заполнение формы конфигурацией
function fillConfigForm(config) {
    if (!config) return;

    // Generic
    $('#setting-mode').val(config.Generic.Mode);
    $('#setting-auto-start').prop('checked', config.Generic.AutoStartProcesses);
    $('#setting-max-processes').val(config.Generic.MaxProcesses);
    $('#setting-web-port').val(config.Generic.WebServerPort);

    // Paths
    $('#setting-podman-path').val(config.Paths.PodmanPath);
    $('#setting-tshark-path').val(config.Paths.TsharkPath);
    $('#setting-filter-file').val(config.Paths.FilterFile);
    $('#setting-control-file').val(config.Paths.ControlFile);
    $('#setting-alerts-file').val(config.Paths.AlertsFile);
    $('#setting-dragonfly-image').val(config.Paths.DragonflyImage);
    $('#setting-dragonfly-backup').val(config.Paths.DragonflyBackupPath);

    // DragonflyDB
    $('#setting-df-host').val(config.DragonflyDB.Host);
    $('#setting-df-port').val(config.DragonflyDB.Port);
    $('#setting-df-password').val(config.DragonflyDB.Password);
    $('#setting-df-database').val(config.DragonflyDB.Database);
    $('#setting-df-batch-size').val(config.DragonflyDB.BatchSize);
    $('#setting-df-timeout').val(config.DragonflyDB.Timeout);

    // PostgreSQL
    $('#setting-pg-host').val(config.PostgreSQL.Host);
    $('#setting-pg-port').val(config.PostgreSQL.Port);
    $('#setting-pg-database').val(config.PostgreSQL.Database);
    $('#setting-pg-user').val(config.PostgreSQL.User);
    $('#setting-pg-password').val(config.PostgreSQL.Password);
    $('#setting-pg-max-connections').val(config.PostgreSQL.MaxConnections);
}

// Обновление списка процессов
function updateProcessesList(processes) {
    var tbody = $('#processes-body');
    tbody.empty();

    if (!processes || processes.length === 0) {
        tbody.append('<tr><td colspan="6" class="empty">Нет активных процессов</td></tr>');
        return;
    }

    $.each(processes, function(index, process) {
        var statusClass = process.Status === 'running' ? 'status-running' : 'status-stopped';
        var statusText = process.Status === 'running' ? 'Запущен' : 'Остановлен';
        var typeText = getTypeText(process.Type);

        var row = '<tr data-id="' + process.ID + '" class="' + statusClass + '">'
            + '<td>' + process.ID + '</td>'
            + '<td>' + process.Name + '</td>'
            + '<td>' + typeText + '</td>'
            + '<td>' + getSourceText(process) + '</td>'
            + '<td>' + statusText + '</td>'
            + '<td class="actions">'
            + (process.Status === 'stopped' 
                ? '<button class="btn btn-success btn-sm" onclick="startProcess(\'' + process.ID + '\')">Старт</button>'
                : '<button class="btn btn-warning btn-sm" onclick="stopProcess(\'' + process.ID + '\')">Стоп</button>')
            + '<button class="btn btn-danger btn-sm" onclick="deleteProcess(\'' + process.ID + '\')">Удалить</button>'
            + '</td>'
            + '</tr>';

        tbody.append(row);
    });
}

// Получение текста типа процесса
function getTypeText(type) {
    var types = {
        'network': 'Сетевой поток',
        'folder': 'Папка с файлами',
        'file': 'Отдельный файл'
    };
    return types[type] || type;
}

// Получение текста источника
function getSourceText(process) {
    if (process.Type === 'network') {
        return process.IP + ':' + process.Port;
    } else if (process.Type === 'folder') {
        return process.Path + (process.ScanSubfolders ? ' (с подпапками)' : '');
    } else if (process.Type === 'file') {
        return process.Path;
    }
    return '';
}

// Запуск процесса
function startProcess(id) {
    $.ajax({
        url: '/api/processes/' + id + '/start',
        method: 'POST',
        success: function() {
            updateProcessesListFromServer();
        },
        error: function(xhr, status, error) {
            alert('Ошибка запуска процесса: ' + error);
        }
    });
}

// Остановка процесса
function stopProcess(id) {
    $.ajax({
        url: '/api/processes/' + id + '/stop',
        method: 'POST',
        success: function() {
            updateProcessesListFromServer();
        },
        error: function(xhr, status, error) {
            alert('Ошибка остановки процесса: ' + error);
        }
    });
}

// Удаление процесса
function deleteProcess(id) {
    if (!confirm('Вы уверены, что хотите удалить процесс?')) {
        return;
    }

    $.ajax({
        url: '/api/processes/' + id,
        method: 'DELETE',
        success: function() {
            updateProcessesListFromServer();
        },
        error: function(xhr, status, error) {
            alert('Ошибка удаления процесса: ' + error);
        }
    });
}

// Обновление списка процессов с сервера
function updateProcessesListFromServer() {
    $.ajax({
        url: '/api/processes',
        method: 'GET',
        success: function(processes) {
            updateProcessesList(processes);
        },
        error: function(xhr, status, error) {
            console.error('Failed to reload processes:', error);
        }
    });
}

// Применение фильтра
function applyFilter() {
    var filterValue = $('#log-filter, #db-filter').val().toLowerCase();
    var table = $('#log-messages, #db-data-table tbody');

    if (table.length) {
        table.find('tr').each(function() {
            var rowText = $(this).text().toLowerCase();
            if (rowText.indexOf(filterValue) > -1) {
                $(this).show();
            } else {
                $(this).hide();
            }
        });
    }
}

// Инициализация конкретной страницы
var pageInit = function(page) {
    switch(page) {
        case 'statistics':
            initStatistics();
            break;
        case 'db':
            initDB();
            break;
        case 'settings':
            initSettings();
            break;
        case 'logs':
            initLogs();
            break;
        case 'diagnose':
            initDiagnose();
            break;
    }
};

// Инициализация страницы статистики
function initStatistics() {
    console.log('Initializing statistics page');

    // Подключение к WebSocket для получения метрик в реальном времени
    var socket = io();

    socket.on('metrics', function(data) {
        updateMetricsDisplay(data);
    });

    socket.on('connect', function() {
        console.log('Connected to WebSocket for metrics');
    });

    socket.on('disconnect', function() {
        console.log('Disconnected from WebSocket for metrics');
    });
}

// Обновление отображения метрик
function updateMetricsDisplay(data) {
    if (!data) return;

    // Системные метрики
    if (data.System) {
        $('#stat-ram').text(data.System.RAMUsage + ' MB');
        $('#stat-ram-bar').css('width', data.System.RAMPercent + '%');
        $('#stat-ram-bar').attr('class', 'stat-bar')
            .addClass(getBarClass(data.System.RAMPercent));

        $('#stat-disk').text(data.System.DiskFree + ' GB');
        $('#stat-disk-bar').css('width', data.System.DiskPercent + '%');
        $('#stat-disk-bar').attr('class', 'stat-bar')
            .addClass(getBarClass(data.System.DiskPercent));

        $('#stat-cpu').text(data.System.CPUUsage + '%');
        $('#stat-cpu-bar').css('width', data.System.CPUPercent + '%');
        $('#stat-cpu-bar').attr('class', 'stat-bar')
            .addClass(getBarClass(data.System.CPUPercent));
    }

    // Метрики процессов
    if (data.Processes) {
        var tbody = $('#stats-processes-body');
        tbody.empty();

        $.each(data.Processes, function(id, metrics) {
            var row = '<tr>'
                + '<td>' + id + '</td>'
                + '<td>' + (metrics.JSONPairs || 0) + '</td>'
                + '<td>' + (metrics.Selected || 0) + '</td>'
                + '<td>' + (metrics.PostgreSQL || 0) + '</td>'
                + '<td>' + (metrics.Speed || 0) + ' pkt/s</td>'
                + '</tr>';

            tbody.append(row);
        });
    }
}

// Получение класса для полоски прогресса
function getBarClass(percent) {
    if (percent < 50) return 'stat-bar-success';
    if (percent < 80) return 'stat-bar-warning';
    return 'stat-bar-danger';
}

// Инициализация страницы базы данных
function initDB() {
    console.log('Initializing DB page');

    // Загрузка данных при старте
    loadDBData();

    // Обработчики кнопок
    $('#btn-db-query').on('click', function() {
        loadDBData();
    });

    $('#btn-db-export').on('click', function() {
        exportDBData();
    });
}

// Загрузка данных из базы данных
function loadDBData() {
    var filter = $('#db-filter').val();

    $.ajax({
        url: '/api/db?filter=' + encodeURIComponent(filter),
        method: 'GET',
        success: function(data) {
            updateDBTable(data);
        },
        error: function(xhr, status, error) {
            alert('Ошибка загрузки данных из базы: ' + error);
        }
    });
}

// Обновление таблицы базы данных
function updateDBTable(data) {
    var tbody = $('#db-data-body');
    tbody.empty();

    if (!data || data.length === 0) {
        tbody.append('<tr><td colspan="6" class="empty">Нет данных</td></tr>');
        return;
    }

    $.each(data, function(index, record) {
        var row = '<tr>'
            + '<td>' + record.ID + '</td>'
            + '<td>' + record.Source + '</td>'
            + '<td>' + record.Key + '</td>'
            + '<td>' + record.Value + '</td>'
            + '<td>' + formatDate(record.DateFirst) + '</td>'
            + '<td>' + formatDate(record.DateLast) + '</td>'
            + '</tr>';

        tbody.append(row);
    });
}

// Форматирование даты
function formatDate(timestamp) {
    if (!timestamp) return '';
    var date = new Date(timestamp * 1000);
    return date.toLocaleString('ru-RU');
}

// Экспорт данных из базы данных
function exportDBData() {
    window.location.href = '/api/db/export';
}

// Инициализация страницы настроек
function initSettings() {
    console.log('Initializing settings page');

    // Сохранение настроек
    $('#btn-save-settings').on('click', function() {
        saveSettings();
    });

    // Проверка подключения к DragonflyDB
    $('#btn-check-dragonfly').on('click', function() {
        checkDragonflyConnection();
    });

    // Проверка подключения к PostgreSQL
    $('#btn-check-postgres').on('click', function() {
        checkPostgreSQLConnection();
    });
}

// Сохранение настроек
function saveSettings() {
    var config = {
        Generic: {
            Mode: $('#setting-mode').val(),
            AutoStartProcesses: $('#setting-auto-start').prop('checked'),
            MaxProcesses: parseInt($('#setting-max-processes').val()),
            WebServerPort: parseInt($('#setting-web-port').val())
        },
        Paths: {
            PodmanPath: $('#setting-podman-path').val(),
            TsharkPath: $('#setting-tshark-path').val(),
            FilterFile: $('#setting-filter-file').val(),
            ControlFile: $('#setting-control-file').val(),
            AlertsFile: $('#setting-alerts-file').val(),
            DragonflyImage: $('#setting-dragonfly-image').val(),
            DragonflyBackupPath: $('#setting-dragonfly-backup').val()
        },
        DragonflyDB: {
            Host: $('#setting-df-host').val(),
            Port: parseInt($('#setting-df-port').val()),
            Password: $('#setting-df-password').val(),
            Database: parseInt($('#setting-df-database').val()),
            BatchSize: parseInt($('#setting-df-batch-size').val()),
            Timeout: parseInt($('#setting-df-timeout').val())
        },
        PostgreSQL: {
            Host: $('#setting-pg-host').val(),
            Port: parseInt($('#setting-pg-port').val()),
            Database: $('#setting-pg-database').val(),
            User: $('#setting-pg-user').val(),
            Password: $('#setting-pg-password').val(),
            MaxConnections: parseInt($('#setting-pg-max-connections').val())
        }
    };

    $.ajax({
        url: '/api/config',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(config),
        success: function() {
            alert('Настройки сохранены');
        },
        error: function(xhr, status, error) {
            alert('Ошибка сохранения настроек: ' + error);
        }
    });
}

// Проверка подключения к DragonflyDB
function checkDragonflyConnection() {
    $.ajax({
        url: '/api/config/check-dragonfly',
        method: 'GET',
        success: function(data) {
            alert('Подключение к DragonflyDB: ' + (data.Success ? 'Успешно' : 'Не удалось'));
        },
        error: function() {
            alert('Подключение к DragonflyDB: Не удалось');
        }
    });
}

// Проверка подключения к PostgreSQL
function checkPostgreSQLConnection() {
    $.ajax({
        url: '/api/config/check-postgres',
        method: 'GET',
        success: function(data) {
            alert('Подключение к PostgreSQL: ' + (data.Success ? 'Успешно' : 'Не удалось'));
        },
        error: function() {
            alert('Подключение к PostgreSQL: Не удалось');
        }
    });
}

// Инициализация страницы логов
function initLogs() {
    console.log('Initializing logs page');

    // Загрузка файлов логов
    loadLogFiles();

    // Обработчики фильтрации
    $('#log-filter').on('input', function() {
        applyFilter();
    });

    $('#log-level-filter').on('change', function() {
        applyFilter();
    });

    $('#log-file-select').on('change', function() {
        loadLogFile($(this).val());
    });

    // Обработчики кнопок
    $('#btn-clear-search').on('click', function() {
        $('#log-filter').val('');
        applyFilter();
    });

    $('#btn-export-html').on('click', function() {
        exportLogsHTML();
    });

    $('#btn-export-pdf').on('click', function() {
        exportLogsPDF();
    });

    // Подключение к WebSocket для получения логов в реальном времени
    var socket = io('/logs');

    socket.on('log', function(data) {
        addLogMessage(data);
    });

    socket.on('connect', function() {
        console.log('Connected to WebSocket for logs');
    });
}

// Загрузка файлов логов
function loadLogFiles() {
    $.ajax({
        url: '/api/logs/files',
        method: 'GET',
        success: function(files) {
            var select = $('#log-file-select');
            select.empty();

            $.each(files, function(index, file) {
                select.append('<option value="' + file + '">' + file + '</option>');
            });

            if (files.length > 0) {
                loadLogFile(files[0]);
            }
        },
        error: function(xhr, status, error) {
            console.error('Failed to load log files:', error);
        }
    });
}

// Загрузка файла логов
function loadLogFile(filename) {
    $.ajax({
        url: '/api/logs/file?name=' + encodeURIComponent(filename),
        method: 'GET',
        success: function(data) {
            updateLogMessages(data);
        },
        error: function(xhr, status, error) {
            console.error('Failed to load log file:', error);
        }
    });
}

// Обновление сообщений логов
function updateLogMessages(messages) {
    var container = $('#log-messages');
    container.empty();

    if (!messages || messages.length === 0) {
        container.append('<p class="empty">Нет сообщений</p>');
        return;
    }

    $.each(messages, function(index, msg) {
        var levelClass = 'log-' + msg.Level.toLowerCase();
        var row = '<div class="log-message ' + levelClass + '">'
            + '<span class="log-time">' + msg.Time + '</span>'
            + '<span class="log-level">' + msg.Level + '</span>'
            + '<span class="log-message-text">' + msg.Message + '</span>'
            + '</div>';

        container.append(row);
    });
}

// Добавление сообщения лога
function addLogMessage(msg) {
    var container = $('#log-messages');
    var levelClass = 'log-' + msg.Level.toLowerCase();
    var row = '<div class="log-message ' + levelClass + '">'
        + '<span class="log-time">' + msg.Time + '</span>'
        + '<span class="log-level">' + msg.Level + '</span>'
        + '<span class="log-message-text">' + msg.Message + '</span>'
        + '</div>';

    container.prepend(row);

    // Ограничение количества сообщений
    if (container.find('.log-message').length > 1000) {
        container.find('.log-message:last').remove();
    }
}

// Экспорт логов в HTML
function exportLogsHTML() {
    window.location.href = '/api/logs/export/html';
}

// Экспорт логов в PDF
function exportLogsPDF() {
    window.location.href = '/api/logs/export/pdf';
}

// Инициализация страницы диагностики
function initDiagnose() {
    console.log('Initializing diagnose page');

    // Запуск диагностики
    $('#btn-run-diagnose').on('click', function() {
        runDiagnose();
    });

    // Экспорт отчета
    $('#btn-export-report').on('click', function() {
        exportReport();
    });

    // Создание архива
    $('#btn-create-archive').on('click', function() {
        createArchive();
    });
}

// Запуск диагностики
function runDiagnose() {
    $('#diagnose-results').html('<p>Запуск диагностики...</p>');

    $.ajax({
        url: '/api/diagnose',
        method: 'GET',
        success: function(data) {
            updateDiagnoseResults(data);
        },
        error: function(xhr, status, error) {
            $('#diagnose-results').html('<p class="error">Ошибка диагностики: ' + error + '</p>');
        }
    });
}

// Обновление результатов диагностики
function updateDiagnoseResults(data) {
    var container = $('#diagnose-results');
    container.empty();

    if (!data || !data.Results || data.Results.length === 0) {
        container.html('<p>Нет результатов диагностики</p>');
        return;
    }

    $.each(data.Results, function(index, result) {
        var statusClass = 'diagnose-' + result.Status;
        var statusText = result.Status === 'success' ? 'ОК' : (result.Status === 'warning' ? 'Предупреждение' : 'Ошибка');

        var row = '<div class="diagnose-result ' + statusClass + '">'
            + '<span class="diagnose-name">' + result.Name + '</span>'
            + '<span class="diagnose-status ' + statusClass + '">' + statusText + '</span>'
            + '<span class="diagnose-message">' + result.Message + '</span>'
            + '</div>';

        container.append(row);
    });
}

// Экспорт отчета диагностики
function exportReport() {
    window.location.href = '/api/diagnose/export';
}

// Создание архива с логами
function createArchive() {
    $.ajax({
        url: '/api/diagnose/archive',
        method: 'POST',
        success: function(data) {
            alert('Архив создан: ' + data.ArchivePath);
            // Автоматическая загрузка архива
            window.location.href = '/api/diagnose/download?file=' + encodeURIComponent(data.ArchivePath);
        },
        error: function(xhr, status, error) {
            alert('Ошибка создания архива: ' + error);
        }
    });
}
