import { Theme } from './theme.js';
import { Calendar } from './calendar.js';
import { Todo } from './todo.js';
import { Countdown } from './countdown.js';
import { Pomodoro } from './pomodoro.js';
import { Diary } from './diary.js';
import { Ledger } from './ledger.js';

'use strict';

const TABS = ['calendar', 'todo', 'pomodoro', 'diary', 'countdown'];
const MODULES = { Calendar, Todo, Pomodoro, Diary, Ledger, Countdown };
let _currentTab = 'calendar';

function loadTabData(tabId) {
    const m = MODULES[tabId.charAt(0).toUpperCase() + tabId.slice(1)];
    if (m && m.onTabEnter) m.onTabEnter();
    // diary tab also loads ledger data
    if (tabId === 'diary' && MODULES.Ledger) {
        MODULES.Ledger.onTabEnter();
    }
}

function switchTab(tabId) {
    if (tabId === _currentTab) return;

    const oldMod = MODULES[_currentTab.charAt(0).toUpperCase() + _currentTab.slice(1)];
    if (oldMod && oldMod.onTabLeave) oldMod.onTabLeave();
    if (_currentTab === 'diary' && MODULES.Ledger && MODULES.Ledger.onTabLeave) {
        MODULES.Ledger.onTabLeave();
    }

    document.querySelectorAll('.tu-tab-btn').forEach(function(btn) {
        btn.classList.toggle('active', btn.getAttribute('data-tab') === tabId);
    });
    document.querySelectorAll('.tu-bottom-nav .tu-tab-btn').forEach(function(btn) {
        btn.classList.toggle('active', btn.getAttribute('data-tab') === tabId);
    });
    document.querySelectorAll('.tu-section').forEach(function(sec) {
        sec.classList.toggle('active', sec.id === 'sec-' + tabId);
    });

    _currentTab = tabId;
    loadTabData(tabId);
    window.location.hash = tabId;
}

function onTabClick(e) {
    const btn = e.target.closest('.tu-tab-btn');
    if (!btn) return;
    const tabId = btn.getAttribute('data-tab');
    if (tabId) switchTab(tabId);
}

function init() {
    Theme.initTheme();

    Calendar.init();
    Todo.init();
    Countdown.init();
    Pomodoro.init();
    Diary.init();
    Ledger.init();

    document.querySelectorAll('.tu-tab-bar, .tu-bottom-nav').forEach(function(bar) {
        bar.addEventListener('click', onTabClick);
    });

    const themeBtn = document.getElementById('themeToggleBtn');
    if (themeBtn) themeBtn.addEventListener('click', function() { Theme.toggleTheme(); });

    const hash = window.location.hash.replace('#', '');
    const initialTab = (hash && TABS.indexOf(hash) !== -1) ? hash : 'calendar';

    if (initialTab !== 'calendar') {
        switchTab(initialTab);
    } else {
        loadTabData('calendar');
    }

    if (initialTab === 'calendar') {
        document.querySelectorAll('.tu-tab-btn[data-tab="calendar"]').forEach(function(btn) {
            btn.classList.add('active');
        });
    }

    window.addEventListener('hashchange', function() {
        const h = window.location.hash.replace('#', '');
        if (h && TABS.indexOf(h) !== -1) switchTab(h);
    });

    document.addEventListener('visibilitychange', function() {
        if (document.hidden) {
            const m = MODULES[_currentTab.charAt(0).toUpperCase() + _currentTab.slice(1)];
            if (m && m.onTabLeave) m.onTabLeave();
            if (_currentTab === 'diary' && MODULES.Ledger && MODULES.Ledger.onTabLeave) {
                MODULES.Ledger.onTabLeave();
            }
        } else {
            const m = MODULES[_currentTab.charAt(0).toUpperCase() + _currentTab.slice(1)];
            if (m && m.onTabEnter) m.onTabEnter();
            if (_currentTab === 'diary' && MODULES.Ledger && MODULES.Ledger.onTabEnter) {
                MODULES.Ledger.onTabEnter();
            }
        }
    });
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}
