import { API } from './constants.js';
import { esc } from './utils.js';

'use strict';

let _abortController = null;

function init() {
    var el = document.getElementById('cdAddBtn');
    if (el) el.addEventListener('click', showForm);
    el = document.getElementById('cdSaveBtn');
    if (el) el.addEventListener('click', saveCD);
    el = document.getElementById('cdCancelBtn');
    if (el) el.addEventListener('click', hideForm);
}

function onTabEnter() { if (!document.hidden) fetchList(); }
function onTabLeave() {
    if (_abortController) { _abortController.abort(); _abortController = null; }
}

function showForm() {
    var form = document.getElementById('cdForm');
    if (form) form.style.display = 'block';
    var titleEl = document.getElementById('cdTitle');
    if (titleEl) titleEl.value = '';
    var targetEl = document.getElementById('cdTarget');
    if (targetEl) targetEl.value = '';
}

function hideForm() {
    var form = document.getElementById('cdForm');
    if (form) form.style.display = 'none';
}

async function saveCD() {
    var titleEl = document.getElementById('cdTitle');
    var targetEl = document.getElementById('cdTarget');
    if (!titleEl || !targetEl) return;
    var title = titleEl.value.trim();
    var targetAt = targetEl.value;
    if (!title || !targetAt) return;

    try {
        await fetch(API.COUNTDOWN, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                title: title,
                target_at: new Date(targetAt).toISOString()
            })
        });
        hideForm();
        fetchList();
    } catch (e) { console.warn('Countdown save:', e.message); }
}

async function fetchList() {
    if (_abortController) _abortController.abort();
    _abortController = new AbortController();
    try {
        var resp = await fetch(API.COUNTDOWN, { signal: _abortController.signal });
        var json = await resp.json();
        if (!json.ok) return;
        renderList(json.data);
    } catch (e) {
        if (e.name !== 'AbortError') console.warn('Countdown:', e.message);
    }
}

function renderList(events) {
    var el = document.getElementById('cdList');
    if (!el) return;

    if (!events || events.length === 0) {
        el.innerHTML = '<div class="tu-empty">暂无倒计时</div>';
        return;
    }

    var html = '';
    events.forEach(function(e) {
        var urgent = e.days_left <= 7 ? ' urgent' : '';
        var label = e.days_left < 0 ? '已过去' + Math.abs(e.days_left) + '天' : '还有' + e.days_left + '天';
        html += '<div class="tu-countdown-item' + urgent + '">';
        html += '<div class="tu-countdown-days">' + e.days_left + '</div>';
        html += '<div class="tu-countdown-info">';
        html += '<div class="tu-countdown-title">' + esc(e.title) + '</div>';
        html += '<div class="tu-countdown-date">' + label + ' · ' + esc(e.target_at.substring(0, 10)) + '</div>';
        html += '</div>';
        if (e.source === 'manual') {
            html += '<button class="tu-btn tu-btn-xs tu-btn-danger tu-cd-del" data-id="' + esc(e.id) + '">删除</button>';
        }
        html += '</div>';
    });
    el.innerHTML = html;

    el.querySelectorAll('.tu-cd-del').forEach(function(btn) {
        btn.addEventListener('click', function() { deleteCD(this.getAttribute('data-id')); });
    });
}

async function deleteCD(id) {
    if (!confirm('确定删除？')) return;
    try {
        await fetch(API.COUNTDOWN + '/' + id, { method: 'DELETE' });
        fetchList();
    } catch (e) { console.warn('Countdown delete:', e.message); }
}

const Countdown = { init, onTabEnter, onTabLeave };
export { Countdown };
