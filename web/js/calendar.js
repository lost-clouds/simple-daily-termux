import { API } from './constants.js';
import { esc } from './utils.js';

'use strict';

let _currentYear, _currentMonth;
let _abortController = null;

function init() {
    const now = new Date();
    _currentYear = now.getFullYear();
    _currentMonth = now.getMonth() + 1;

    const prevBtn = document.getElementById('calPrev');
    const nextBtn = document.getElementById('calNext');
    if (prevBtn) prevBtn.addEventListener('click', function() {
        _currentMonth--; if (_currentMonth < 1) { _currentMonth = 12; _currentYear--; }
        fetchMonth();
    });
    if (nextBtn) nextBtn.addEventListener('click', function() {
        _currentMonth++; if (_currentMonth > 12) { _currentMonth = 1; _currentYear++; }
        fetchMonth();
    });
}

function onTabEnter() {
    if (!document.hidden) fetchMonth();
}

function onTabLeave() {
    if (_abortController) { _abortController.abort(); _abortController = null; }
}

async function fetchMonth() {
    if (_abortController) _abortController.abort();
    _abortController = new AbortController();

    const titleEl = document.getElementById('calTitle');
    if (titleEl) titleEl.textContent = _currentYear + '年' + _currentMonth + '月';

    try {
        const m = String(_currentMonth).padStart(2, '0');
        const resp = await fetch(API.CALENDAR + '?month=' + _currentYear + '-' + m, {
            signal: _abortController.signal
        });
        const json = await resp.json();
        if (!json.ok) return;
        renderCalendar(json.data);
    } catch (e) {
        if (e.name !== 'AbortError') console.warn('Calendar:', e.message);
    }
}

function renderCalendar(data) {
    const grid = document.getElementById('calGrid');
    if (!grid) return;

    const year = _currentYear, month = _currentMonth;
    const firstDay = new Date(year, month - 1, 1).getDay();
    const daysInMonth = new Date(year, month, 0).getDate();
    const diaryDates = new Set(data.diary_dates || []);

    const today = new Date();
    const todayStr = today.getFullYear() + '-' +
        String(today.getMonth() + 1).padStart(2, '0') + '-' +
        String(today.getDate()).padStart(2, '0');

    let html = '';
    var headers = ['日', '一', '二', '三', '四', '五', '六'];
    headers.forEach(function(h) {
        html += '<div class="tu-calendar-day-header">' + h + '</div>';
    });

    for (var i = 0; i < firstDay; i++) {
        html += '<div class="tu-calendar-day other-month"></div>';
    }

    for (var d = 1; d <= daysInMonth; d++) {
        var dateStr = year + '-' + String(month).padStart(2, '0') + '-' + String(d).padStart(2, '0');
        var cls = 'tu-calendar-day';
        if (dateStr === todayStr) cls += ' today';
        if (diaryDates.has(dateStr)) cls += ' has-diary';
        html += '<div class="' + cls + '" data-date="' + dateStr + '">' + d + '</div>';
    }

    grid.innerHTML = html;

    var eventsList = document.getElementById('calEventsList');
    if (eventsList) {
        var sections = [];

        // Calendar events
        var events = data.events || [];
        if (events.length > 0) {
            var ehtml = '';
            events.forEach(function(ev) {
                var dateLabel = (ev.start_at || '').substring(0, 10);
                ehtml += '<div class="tu-calendar-event-item">' +
                    '<span class="tu-calendar-event-date">' + esc(dateLabel) + '</span>' +
                    '<span>' + esc(ev.title || '') + '</span></div>';
            });
            sections.push(ehtml);
        }

        // Todo deadlines
        var deadlines = data.todo_deadlines || [];
        if (deadlines.length > 0) {
            var dhtml = '<div class="tu-card-header"><span class="tu-card-title">📌 待办截止</span></div>';
            deadlines.forEach(function(td) {
                var dl = td.deadline_at ? td.deadline_at.substring(0, 10) : '';
                dhtml += '<div class="tu-calendar-event-item">' +
                    '<span class="tu-calendar-event-date">' + esc(dl) + '</span>' +
                    '<span>' + esc(td.title || '') + '</span>' +
                    '<span class="tu-badge tu-badge-' + (td.status || 'pending') + '">' + esc(td.status || 'pending') + '</span>' +
                    '</div>';
            });
            sections.push(dhtml);
        }

        // Countdown targets
        var targets = data.countdown_targets || [];
        if (targets.length > 0) {
            var chtml = '<div class="tu-card-header"><span class="tu-card-title">⏳ 倒计时</span></div>';
            targets.forEach(function(ct) {
                var ctDate = ct.target_at ? ct.target_at.substring(0, 10) : '';
                chtml += '<div class="tu-calendar-event-item">' +
                    '<span class="tu-calendar-event-date">' + esc(ctDate) + '</span>' +
                    '<span>' + esc(ct.title || '') + '</span>' +
                    '<span style="color:var(--accent);margin-left:auto">' + ct.days_left + '天</span>' +
                    '</div>';
            });
            sections.push(chtml);
        }

        eventsList.innerHTML = sections.length > 0 ? sections.join('') :
            '<div class="tu-empty">本月无事件</div>';
    }
}

const Calendar = { init, onTabEnter, onTabLeave };
export { Calendar };
