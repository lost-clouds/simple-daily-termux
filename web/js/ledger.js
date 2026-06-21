import { API } from './constants.js';
import { esc } from './utils.js';

'use strict';

function init() {
    var el = document.getElementById('ledgerAddBtn');
    if (el) el.addEventListener('click', showForm);
    el = document.getElementById('ledgerSaveBtn');
    if (el) el.addEventListener('click', saveEntry);
    el = document.getElementById('ledgerCancelBtn');
    if (el) el.addEventListener('click', hideForm);
}

function onTabEnter() { if (!document.hidden) { fetchSummary(); fetchList(); } }
function onTabLeave() {}

function showForm() {
    var form = document.getElementById('ledgerForm');
    if (form) form.style.display = 'block';
    var dateEl = document.getElementById('ledgerEntryDate');
    if (dateEl) dateEl.value = new Date().toISOString().substring(0, 10);
    var amountEl = document.getElementById('ledgerAmount');
    if (amountEl) amountEl.value = '';
    var catEl = document.getElementById('ledgerCategory');
    if (catEl) catEl.value = '';
}

function hideForm() {
    var form = document.getElementById('ledgerForm');
    if (form) form.style.display = 'none';
}

async function saveEntry() {
    var dateEl = document.getElementById('ledgerEntryDate');
    var typeEl = document.getElementById('ledgerType');
    var amountEl = document.getElementById('ledgerAmount');
    var catEl = document.getElementById('ledgerCategory');
    var noteEl = document.getElementById('ledgerNote');
    if (!dateEl || !typeEl || !amountEl || !catEl) return;

    var entryDate = dateEl.value;
    var type = typeEl.value;
    var amount = parseFloat(amountEl.value);
    var category = catEl.value.trim();
    var note = noteEl ? noteEl.value.trim() : '';

    if (!entryDate || !category || isNaN(amount)) return;

    try {
        var resp = await fetch(API.LEDGER, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                entry_date: entryDate, type: type, amount: amount,
                category: category, note: note
            })
        });
        var json = await resp.json();
        if (!json.ok) { console.warn('Ledger save failed:', json.error); return; }
        hideForm();
        fetchSummary();
        fetchList();
    } catch (e) { console.warn('Ledger save:', e.message); }
}

async function fetchSummary() {
    var now = new Date();
    var month = now.getFullYear() + '-' + String(now.getMonth() + 1).padStart(2, '0');
    try {
        var resp = await fetch(API.LEDGER_SUM + '?month=' + month);
        var json = await resp.json();
        if (!json.ok || !json.data) return;
        var d = json.data;
        setText('sumExpense', '¥' + (d.expense || 0).toFixed(2));
        setText('sumIncome', '¥' + (d.income || 0).toFixed(2));
        setText('sumBalance', '¥' + (d.balance || 0).toFixed(2));
        setText('sumSavings', '¥' + (d.savings || 0).toFixed(2));
    } catch (e) { console.warn('Ledger summary:', e.message); }
}

async function fetchList() {
    var now = new Date();
    var month = now.getFullYear() + '-' + String(now.getMonth() + 1).padStart(2, '0');
    try {
        var resp = await fetch(API.LEDGER + '?month=' + month);
        var json = await resp.json();
        if (!json.ok) return;
        renderList(json.data);
    } catch (e) { console.warn('Ledger list:', e.message); }
}

function renderList(entries) {
    var el = document.getElementById('ledgerList');
    if (!el) return;
    if (!entries || entries.length === 0) {
        el.innerHTML = '<div class="tu-empty">本月暂无记录</div>';
        return;
    }

    var html = '';
    entries.forEach(function(e) {
        var color = e.type === 'income' ? '#34c759' : '#ff3b30';
        var prefix = e.type === 'income' ? '+' : '-';
        html += '<div class="tu-ledger-entry">';
        html += '<span class="tu-ledger-entry-date">' + esc(e.entry_date) + '</span>';
        html += '<span class="tu-ledger-entry-category">' + esc(e.category) + '</span>';
        html += '<span class="tu-ledger-entry-amount" style="color:' + color + '">' + prefix + '¥' + e.amount.toFixed(2) + '</span>';
        html += '</div>';
    });
    el.innerHTML = html;
}

function setText(id, text) {
    var el = document.getElementById(id);
    if (el) el.textContent = text || '--';
}

const Ledger = { init, onTabEnter, onTabLeave };
export { Ledger };
