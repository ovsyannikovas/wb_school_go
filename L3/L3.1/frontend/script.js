// Конфигурация
const API_BASE = 'http://localhost:8080/api';
let pollInterval = null;
let currentTrackingId = null;

// Инициализация
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    setDefaultDateTime();
    loadRecentNotifications();
});

// Настройка обработчиков
function setupEventListeners() {
    document.getElementById('notificationForm').addEventListener('submit', createNotification);
    document.getElementById('trackBtn').addEventListener('click', trackNotification);
    document.getElementById('clearBtn').addEventListener('click', clearTracking);

    document.getElementById('trackingId').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') trackNotification();
    });
}

// Установка времени по умолчанию (+1 минута для теста)
function setDefaultDateTime() {
    const sendAt = document.getElementById('sendAt');
    const date = new Date();
    date.setMinutes(date.getMinutes() + 1);
    sendAt.value = date.toISOString().slice(0, 16);
}

// Создание уведомления
async function createNotification(e) {
    e.preventDefault();

    const button = document.getElementById('createBtn');
    button.disabled = true;
    button.textContent = '⏳ Создание...';

    const formData = {
        user_id: document.getElementById('userId').value,
        channel: document.getElementById('channel').value,
        recipient: document.getElementById('recipient').value,
        content: document.getElementById('content').value,
        send_at: new Date(document.getElementById('sendAt').value).toISOString()
    };

    try {
        const response = await fetch(`${API_BASE}/notify`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });

        if (!response.ok) {
            throw new Error(`HTTP error ${response.status}`);
        }

        const result = await response.json();

        // Получаем ID из ответа (может быть в разных форматах)
        const notificationId = result.data?.id || result.id;

        // Показываем уведомление с ID
        showNotificationId(notificationId);

        // Очищаем форму
        document.getElementById('notificationForm').reset();
        setDefaultDateTime();

        // Обновляем список последних
        loadRecentNotifications();

    } catch (error) {
        console.error('Creation error:', error);
        showToast('❌ Ошибка создания: ' + error.message, 'error');
    } finally {
        button.disabled = false;
        button.textContent = '📨 Создать уведомление';
    }
}

// Показываем ID созданного уведомления
function showNotificationId(id) {
    // Создаем временное уведомление
    const notificationDiv = document.createElement('div');
    notificationDiv.className = 'notification-card pending';
    notificationDiv.innerHTML = `
        <div style="text-align: center; padding: 20px;">
            <div style="font-size: 48px; margin-bottom: 20px;">✅</div>
            <h3 style="margin-bottom: 15px;">Уведомление создано!</h3>
            <div style="background: #edf2f7; padding: 15px; border-radius: 8px; margin-bottom: 15px;">
                <div style="font-size: 12px; color: #666; margin-bottom: 5px;">ID уведомления:</div>
                <div style="font-family: monospace; font-size: 14px; word-break: break-all; background: white; padding: 10px; border-radius: 5px; border: 1px solid #ddd;">
                    ${id}
                </div>
            </div>
            <div style="display: flex; gap: 10px; justify-content: center;">
                <button onclick="copyToClipboard('${id}')" style="width: auto; padding: 8px 20px; background: #48bb78;">📋 Копировать ID</button>
                <button onclick="document.getElementById('trackingId').value='${id}'; trackNotification();" style="width: auto; padding: 8px 20px;">🔍 Отследить</button>
            </div>
        </div>
    `;

    // Показываем во временном контейнере
    const trackedDiv = document.getElementById('trackedNotification');
    trackedDiv.style.display = 'block';
    trackedDiv.innerHTML = '';
    trackedDiv.appendChild(notificationDiv);

    // Автоматически вставляем ID в поле отслеживания
    document.getElementById('trackingId').value = id;

    showToast('✅ Уведомление создано! ID скопирован в буфер?', 'success');
}

// Копирование в буфер обмена
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        showToast('📋 ID скопирован в буфер обмена', 'success');
    }).catch(() => {
        // Fallback
        const textarea = document.createElement('textarea');
        textarea.value = text;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
        showToast('📋 ID скопирован в буфер обмена', 'success');
    });
}

// Отслеживание уведомления
async function trackNotification() {
    const id = document.getElementById('trackingId').value.trim();

    if (!id) {
        showToast('Введите ID уведомления', 'error');
        return;
    }

    currentTrackingId = id;

    // Останавливаем предыдущий пуллинг
    if (pollInterval) {
        clearInterval(pollInterval);
    }

    // Загружаем и начинаем пуллинг
    await fetchAndDisplayNotification(id);
    pollInterval = setInterval(() => fetchAndDisplayNotification(id), 2000);
}

// Загрузка и отображение уведомления
async function fetchAndDisplayNotification(id) {
    try {
        const response = await fetch(`${API_BASE}/notify/${id}`);

        if (response.status === 404) {
            document.getElementById('trackedNotification').innerHTML = `
                <div class="error-message" style="text-align: center; padding: 40px;">
                    <div style="font-size: 48px; margin-bottom: 20px;">🔍</div>
                    <h3 style="color: #666;">Уведомление не найдено</h3>
                    <p style="color: #999; margin-top: 10px;">ID: ${id}</p>
                </div>
            `;
            return;
        }

        if (!response.ok) {
            throw new Error('Ошибка загрузки');
        }

        const data = await response.json();
        const notification = data.data || data;

        displayTrackedNotification(notification);

        // Если уведомление в конечном статусе, останавливаем пуллинг
        if (['sent', 'failed', 'cancelled'].includes(notification.status)) {
            if (pollInterval) {
                clearInterval(pollInterval);
                pollInterval = null;
            }
        }
    } catch (error) {
        console.error('Error fetching notification:', error);
    }
}

// Отображение отслеживаемого уведомления
function displayTrackedNotification(notification) {
    const container = document.getElementById('trackedNotification');
    container.style.display = 'block';

    // Вычисляем прогресс
    const now = new Date();
    const sendAt = new Date(notification.send_at);
    const createdAt = new Date(notification.created_at);

    let progress = 0;
    let timeInfo = '';
    let countdown = '';

    if (notification.status === 'pending') {
        if (now < sendAt) {
            const total = sendAt - createdAt;
            const elapsed = now - createdAt;
            progress = Math.min((elapsed / total) * 100, 99);

            const timeLeft = sendAt - now;
            const minutes = Math.floor(timeLeft / 60000);
            const seconds = Math.floor((timeLeft % 60000) / 1000);
            countdown = `${minutes}м ${seconds}с`;
            timeInfo = '⏳ До отправки';
        } else {
            progress = 99;
            timeInfo = '⏳ Отправляется...';
        }
    } else if (notification.status === 'sent') {
        progress = 100;
        timeInfo = '✅ Отправлено';
    } else if (notification.status === 'failed') {
        progress = 100;
        timeInfo = '❌ Ошибка';
    } else if (notification.status === 'cancelled') {
        progress = 100;
        timeInfo = '🚫 Отменено';
    }

    container.innerHTML = `
        <div class="notification-card ${notification.status}">
            <div class="notification-header">
                <span class="notification-id">📋 ID: ${notification.id}</span>
                <span class="notification-status status-${notification.status}">${notification.status}</span>
            </div>
            
            <div class="notification-content">${notification.content}</div>
            
            <div class="notification-meta">
                <div>👤 Пользователь: ${notification.user_id}</div>
                <div>📱 Канал: ${notification.channel}</div>
                <div>📧 Получатель: ${notification.recipient}</div>
                <div>⏰ Создано: ${new Date(notification.created_at).toLocaleString()}</div>
                <div>📅 Отправка: ${new Date(notification.send_at).toLocaleString()}</div>
                <div>🔄 Попыток: ${notification.retry_count || 0}</div>
            </div>
            
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${progress}%"></div>
            </div>
            
            <div class="time-info">
                <span>${timeInfo}</span>
                ${countdown ? `<span class="countdown">⏱️ ${countdown}</span>` : ''}
            </div>
            
            ${notification.last_error ? `
                <div class="error-details">
                    ❌ Ошибка: ${notification.last_error}
                </div>
            ` : ''}
            
            ${notification.status === 'pending' ? `
                <div style="margin-top: 15px; display: flex; gap: 10px;">
                    <button onclick="copyToClipboard('${notification.id}')" style="width: auto; padding: 8px; background: #718096;">📋 Копировать ID</button>
                    <button onclick="cancelNotification('${notification.id}')" class="btn-delete" style="width: auto; padding: 8px; flex: 1;">
                        🚫 Отменить
                    </button>
                </div>
            ` : `
                <div style="margin-top: 15px;">
                    <button onclick="copyToClipboard('${notification.id}')" style="width: 100%; padding: 8px; background: #718096;">📋 Копировать ID</button>
                </div>
            `}
        </div>
    `;
}

// Отмена уведомления
async function cancelNotification(id) {
    if (!confirm('Отменить уведомление?')) return;

    try {
        const response = await fetch(`${API_BASE}/notify/${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) throw new Error('Ошибка отмены');

        showToast('✅ Уведомление отменено', 'success');
        await fetchAndDisplayNotification(id);
        loadRecentNotifications();
    } catch (error) {
        showToast('❌ ' + error.message, 'error');
    }
}

// Очистка отслеживания
function clearTracking() {
    document.getElementById('trackingId').value = '';
    document.getElementById('trackedNotification').style.display = 'none';
    document.getElementById('trackedNotification').innerHTML = '';

    if (pollInterval) {
        clearInterval(pollInterval);
        pollInterval = null;
    }

    currentTrackingId = null;
}

// Загрузка последних уведомлений
async function loadRecentNotifications() {
    try {
        const response = await fetch(`${API_BASE}/notifications`);
        if (!response.ok) return;

        const result = await response.json();
        let notifications = result.data || result || [];

        if (!Array.isArray(notifications)) {
            notifications = [];
        }

        // Сортируем по дате создания
        notifications.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        notifications = notifications.slice(0, 5);

        const container = document.getElementById('recentNotifications');

        if (notifications.length === 0) {
            container.innerHTML = '<div class="loading">Нет уведомлений</div>';
            return;
        }

        container.innerHTML = notifications.map(n => `
            <div class="notification-card ${n.status}" style="cursor: pointer; margin-bottom: 10px;" 
                 onclick="document.getElementById('trackingId').value='${n.id}'; trackNotification();">
                <div class="notification-header">
                    <span class="notification-id">📋 ${n.id.slice(0, 8)}...</span>
                    <span class="notification-status status-${n.status}">${n.status}</span>
                </div>
                <div class="notification-content">${n.content.substring(0, 50)}${n.content.length > 50 ? '...' : ''}</div>
                <div style="font-size: 11px; color: #999; margin-top: 5px;">
                    ${new Date(n.send_at).toLocaleString()}
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading recent notifications:', error);
    }
}

// Показ toast уведомлений
function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => toast.remove(), 3000);
}

// Очистка при уходе со страницы
window.addEventListener('beforeunload', () => {
    if (pollInterval) {
        clearInterval(pollInterval);
    }
});