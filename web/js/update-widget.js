/* update-widget.js —— simple-daily-termux 集成卡片
   fetch /api/summary → render into Blog-termux dashboard 9th card */

'use strict';

let _timer = null;
let _fetching = false;

function set(id, text) {
    const el = document.getElementById(id);
    if (el) el.textContent = text || '--';
}

async function fetchSummary() {
    if (_fetching) return;
    _fetching = true;
    try {
        const resp = await fetch('/api/summary');
        if (!resp.ok) throw new Error('HTTP ' + resp.status);
        const json = await resp.json();
        if (!json.ok || !json.data) return;

        const d = json.data;
        const led = d.ledger || {};

        set('sumExpense', '¥' + (led.expense || 0).toFixed(2));
        set('sumIncome', '¥' + (led.income || 0).toFixed(2));
        set('sumBalance', '¥' + (led.balance || 0).toFixed(2));
        set('sumSavings', '¥' + (led.savings || 0).toFixed(2));
        set('sumFocus', (d.focus_today_minutes || 0) + ' 分钟');

        const cds = d.countdown || [];
        if (cds.length > 0) {
            set('sumCountdown', cds[0].title + ' ' + cds[0].days_left + '天');
        } else {
            set('sumCountdown', '无');
        }
    } catch (err) {
        console.warn('UpdateWidget:', err.message);
    } finally {
        _fetching = false;
    }
}

function start() {
    fetchSummary();
    _timer = setInterval(fetchSummary, 30000);
}

function stop() {
    if (_timer) { clearInterval(_timer); _timer = null; }
}

function init() {
    start();
    document.addEventListener('visibilitychange', function() {
        if (document.hidden) {
            stop();
        } else {
            start();
        }
    });
}

const UpdateWidget = { init, start, stop, fetchSummary };
export { UpdateWidget };
