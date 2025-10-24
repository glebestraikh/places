const API = '/api';

document.addEventListener('DOMContentLoaded', init);

function init() {
    document.getElementById('searchInput').addEventListener('keypress', e => {
        if (e.key === 'Enter') search();
    });
    document.getElementById('searchButton').addEventListener('click', search);
    document.getElementById('backButton').addEventListener('click', goBack);
}

async function search() {
    const query = document.getElementById('searchInput').value.trim();
    if (!query) {
        showError('Введите название места');
        return;
    }

    const btn = document.getElementById('searchButton');
    btn.disabled = true;
    btn.textContent = 'Поиск...';

    try {
        const res = await fetch(`${API}/search`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({query})
        });

        const locations = await res.json();
        if (!locations?.length) {
            showError('Локации не найдены');
            return;
        }

        showLocations(locations);
    } catch (err) {
        showError('Ошибка: ' + err.message);
    } finally {
        btn.disabled = false;
        btn.textContent = 'Поиск';
    }
}

function showLocations(locations) {
    const list = document.getElementById('locationsList');
    list.innerHTML = locations.map(loc => `
        <div class="location-item" onclick='selectLocation(${JSON.stringify(loc).replace(/'/g, "&apos;")})'>
            <h3>${esc(loc.name)}</h3>
            <p>${[loc.state, loc.country].filter(Boolean).map(esc).join(', ')}</p>
            <p style="font-size: 12px; color: #80868b;">📍 ${loc.lat.toFixed(4)}, ${loc.lon.toFixed(4)}</p>
        </div>
    `).join('');
}

async function selectLocation(location) {
    show('loadingSection');

    try {
        const res = await fetch(`${API}/location/details`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({location})
        });

        const data = await res.json();
        showResults(data);
    } catch (err) {
        showError('Ошибка: ' + err.message);
        goBack();
    }
}

function showResults(data) {
    showWeather(data.location, data.weather);
    showPlaces(data.places);
    show('resultsSection');
}

function showWeather(location, weather) {
    const card = document.getElementById('weatherCard');
    if (!weather) {
        card.innerHTML = '';
        return;
    }

    card.innerHTML = `
        <div class="weather-card">
            <div class="weather-header">
                <div>
                    <div class="weather-title">${esc(location.name)}</div>
                    <div class="weather-desc">${esc(weather.description)}</div>
                </div>
                <img src="https://openweathermap.org/img/wn/${weather.icon}@2x.png" alt="${esc(weather.description)}" style="width: 80px;">
            </div>
            <div class="weather-info">
                <div class="weather-item">
                    <label>Температура</label>
                    <value>${Math.round(weather.temp)}°C</value>
                </div>
                <div class="weather-item">
                    <label>Ощущается</label>
                    <value>${Math.round(weather.feels_like)}°C</value>
                </div>
                <div class="weather-item">
                    <label>Влажность</label>
                    <value>${weather.humidity}%</value>
                </div>
                <div class="weather-item">
                    <label>Ветер</label>
                    <value>${weather.wind_speed} м/с</value>
                </div>
            </div>
        </div>
    `;
}

function showPlaces(places) {
    const card = document.getElementById('placesCard');
    if (!places?.length) {
        card.innerHTML = '<p style="color: #5f6368;">Интересные места не найдены</p>';
        return;
    }

    card.innerHTML = `
        <div class="places-header">Интересные места (${places.length})</div>
        <div class="places-list">
            ${places.map((p, i) => `
                <div class="place-item" onclick="showModal(${i})">
                    <div class="place-item-header">
                        <div class="place-item-title">${esc(p.name || 'Без названия')}</div>
                        ${p.kinds ? `<span class="place-category">${esc(p.kinds.split(',')[0].trim())}</span>` : ''}
                    </div>
                    <div class="place-item-description">
                        ${formatPlaceShortInfo(p)}
                    </div>
                </div>
            `).join('')}
        </div>
    `;

    window.currentPlaces = places;
}

function formatPlaceShortInfo(place) {
    if (!place.description) {
        return '<span style="color: #80868b;">Нажмите для просмотра деталей</span>';
    }

    const lines = place.description.split('\n').filter(line => line.trim());
    const preview = lines.slice(0, 2).join(' • ');
    return esc(truncate(preview, 120));
}

function showModal(index) {
    const place = window.currentPlaces[index];
    const modal = document.getElementById('placeModal');
    const body = document.getElementById('modalBody');

    let html = '';

    if (place.image) {
        html += `<img src="${esc(place.image)}" class="modal-image" onerror="this.style.display='none'">`;
    }

    html += `<div class="modal-title">${esc(place.name || 'Без названия')}</div>`;

    if (place.kinds) {
        html += `<div class="place-category" style="margin-bottom: 20px;">${esc(place.kinds.split(',')[0].trim())}</div>`;
    }

    if (place.description) {
        html += `
            <div class="modal-section">
                <h3>Информация</h3>
                <p>${esc(place.description)}</p>
            </div>
        `;
    }

    const links = [];

    // Website link
    if (place.website) {
        const url = place.website.startsWith('http') ? place.website : `https://${place.website}`;
        links.push(`<a href="${esc(url)}" target="_blank" class="modal-link">🌐 Сайт</a>`);
    }

    // Wikipedia link
    if (place.wikipedia) {
        const url = place.wikipedia.startsWith('http') ? place.wikipedia : `https://${place.wikipedia}`;
        links.push(`<a href="${esc(url)}" target="_blank" class="modal-link">📖 Wikipedia</a>`);
    }

    // Google Maps link
    links.push(`<a href="https://www.google.com/maps/search/?api=1&query=${place.lat},${place.lon}" target="_blank" class="modal-link">🗺️ Карта</a>`);

    if (links.length > 0) {
        html += `
            <div class="modal-section">
                <h3>Ссылки</h3>
                <div class="modal-links">${links.join('')}</div>
            </div>
        `;
    }

    body.innerHTML = html;
    modal.style.display = 'block';
    modal.onclick = e => { if (e.target === modal) closeModal(); };
}

function closeModal() {
    document.getElementById('placeModal').style.display = 'none';
}

function goBack() {
    show('searchSection');
    document.getElementById('locationsList').innerHTML = '';
    document.getElementById('errorMessage').innerHTML = '';
}

function show(section) {
    ['searchSection', 'resultsSection', 'loadingSection'].forEach(id => {
        document.getElementById(id).style.display = id === section ? 'block' : 'none';
    });
}

function showError(msg) {
    document.getElementById('errorMessage').innerHTML = `<div class="error-message">${esc(msg)}</div>`;
}

function esc(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function truncate(text, len) {
    return text.length > len ? text.substr(0, len) + '...' : text;
}

document.addEventListener('keydown', e => {
    if (e.key === 'Escape') closeModal();
});

