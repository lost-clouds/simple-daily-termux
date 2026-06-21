import { API } from './constants.js';

'use strict';

let _timer = null;
let _secondsLeft = 0;
let _running = false;
let _sessionId = '';

function init() {
    var el = document.getElementById('pomoStart');
    if (el) el.addEventListener('click', start);
    el = document.getElementById('pomoPause');
    if (el) el.addEventListener('click', pause);
    el = document.getElementById('pomoReset');
    if (el) el.addEventListener('click', reset);

    var presets = document.querySelectorAll('.tu-pomo-preset-btn');
    presets.forEach(function(btn) {
        btn.addEventListener('click', function() {
            var mins = parseInt(this.getAttribute('data-minutes'));
            if (!_running) {
                _secondsLeft = mins * 60;
                updateDisplay();
            }
        });
    });
    _secondsLeft = 25 * 60;
    updateDisplay();
}

function onTabEnter() { if (!document.hidden) fetchStats(); }
function onTabLeave() {
    if (_running) {
        pause();
    }
}

function updateDisplay() {
    var m = Math.floor(_secondsLeft / 60);
    var s = _secondsLeft % 60;
    var el = document.getElementById('pomoTimer');
    if (el) el.textContent = String(m).padStart(2, '0') + ':' + String(s).padStart(2, '0');
}

function tick() {
    _secondsLeft--;
    updateDisplay();
    if (_secondsLeft <= 0) {
        finish('completed');
    }
}

async function start() {
    if (_running) return;
    try {
        var resp = await fetch(API.POMO_START, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ planned_minutes: Math.ceil(_secondsLeft / 60) })
        });
        var json = await resp.json();
        if (!json.ok) return;
        _sessionId = json.data.id;
        _running = true;
        var startEl = document.getElementById('pomoStart');
        var pauseEl = document.getElementById('pomoPause');
        if (startEl) startEl.style.display = 'none';
        if (pauseEl) pauseEl.style.display = '';
        _timer = setInterval(tick, 1000);
    } catch (e) { console.warn('Pomodoro start:', e.message); }
}

async function pause() {
    _running = false;
    if (_timer) { clearInterval(_timer); _timer = null; }
    // Finish the session on server as aborted
    if (_sessionId) {
        try {
            await fetch(API.POMODORO + '/' + _sessionId + '/finish', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ status: 'aborted' })
            });
        } catch (e) { console.warn(e); }
        _sessionId = '';
    }
    var startEl = document.getElementById('pomoStart');
    var pauseEl = document.getElementById('pomoPause');
    if (startEl) startEl.style.display = '';
    if (pauseEl) pauseEl.style.display = 'none';
    fetchStats();
}

async function finish(status) {
    _running = false;
    if (_timer) { clearInterval(_timer); _timer = null; }
    if (_sessionId) {
        try {
            await fetch(API.POMODORO + '/' + _sessionId + '/finish', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ status: status })
            });
        } catch (e) { console.warn(e); }
        _sessionId = '';
    }
    var startEl = document.getElementById('pomoStart');
    var pauseEl = document.getElementById('pomoPause');
    if (startEl) startEl.style.display = '';
    if (pauseEl) pauseEl.style.display = 'none';
    fetchStats();
}

async function reset() {
    if (_running) {
        // Finish as aborted before resetting
        if (_sessionId) {
            try {
                await fetch(API.POMODORO + '/' + _sessionId + '/finish', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ status: 'aborted' })
                });
            } catch (e) { console.warn(e); }
        }
    }
    _running = false;
    if (_timer) { clearInterval(_timer); _timer = null; }
    _secondsLeft = 25 * 60;
    _sessionId = '';
    updateDisplay();

    var startEl = document.getElementById('pomoStart');
    var pauseEl = document.getElementById('pomoPause');
    if (startEl) startEl.style.display = '';
    if (pauseEl) pauseEl.style.display = 'none';
    fetchStats();
}

async function fetchStats() {
    try {
        var resp = await fetch(API.POMO_TODAY);
        if (!resp.ok) return;
        var json = await resp.json();
        if (!json.ok) return;
        var el = document.getElementById('pomoTodayMinutes');
        if (el) el.textContent = json.data.total_minutes;
    } catch (e) { console.warn(e); }
}

const Pomodoro = { init, onTabEnter, onTabLeave };
export { Pomodoro };
