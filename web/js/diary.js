import { API } from './constants.js';
import { esc } from './utils.js';

'use strict';

let _currentDate = new Date().toISOString().substring(0, 10);

function init() {
    var el = document.getElementById('diaryDate');
    if (el) {
        el.value = _currentDate;
        el.addEventListener('change', function() {
            _currentDate = this.value;
            loadDiary();
        });
    }
    el = document.getElementById('diarySaveBtn');
    if (el) el.addEventListener('click', saveDiary);
    el = document.getElementById('insertLedgerBtn');
    if (el) el.addEventListener('click', insertLedgerBlock);
    document.querySelectorAll('.tu-diary-mood-btn').forEach(function(btn) {
        if (btn) btn.addEventListener('click', function() {
            document.querySelectorAll('.tu-diary-mood-btn').forEach(function(b) { b.classList.remove('active'); });
            this.classList.add('active');
        });
    });
}

function onTabEnter() { if (!document.hidden) loadDiary(); }
function onTabLeave() {}

async function loadDiary() {
    try {
        var resp = await fetch(API.DIARY + '/' + _currentDate);
        var json = await resp.json();
        var contentEl = document.getElementById('diaryContent');
        var previewEl = document.getElementById('diaryPreview');
        if (json.ok && json.data) {
            if (contentEl) contentEl.value = json.data.content_md || '';
            if (previewEl) previewEl.innerHTML = mdPreview(json.data.content_md || '');
            if (json.data.mood) {
                document.querySelectorAll('.tu-diary-mood-btn').forEach(function(b) {
                    b.classList.toggle('active', b.getAttribute('data-mood') === json.data.mood);
                });
            }
        } else {
            if (contentEl) contentEl.value = '';
            if (previewEl) previewEl.innerHTML = '<span class="tu-text-muted">今天还没写日记</span>';
        }
    } catch (e) {
        console.warn('Diary load:', e.message);
    }
}

async function saveDiary() {
    var contentEl = document.getElementById('diaryContent');
    if (!contentEl) return;
    var contentMD = contentEl.value;
    var moodBtn = document.querySelector('.tu-diary-mood-btn.active');
    var mood = moodBtn ? moodBtn.getAttribute('data-mood') : '';

    try {
        var resp = await fetch(API.DIARY + '/' + _currentDate, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ content_md: contentMD, mood: mood })
        });
        var json = await resp.json();
        if (json.ok) {
            var previewEl = document.getElementById('diaryPreview');
            if (previewEl) previewEl.innerHTML = mdPreview(contentMD);
        }
    } catch (e) { console.warn('Diary save:', e.message); }
}

function insertLedgerBlock() {
    var textarea = document.getElementById('diaryContent');
    if (!textarea) return;
    var template = '\n```ledger\ntype: expense\namount: 0\ncategory: \nnote: \n```\n';
    var pos = textarea.selectionStart;
    textarea.value = textarea.value.substring(0, pos) + template + textarea.value.substring(textarea.selectionEnd);
    textarea.focus();
    textarea.selectionStart = textarea.selectionEnd = pos + template.length - 5;
}

function mdPreview(md) {
    if (!md) return '';

    // Extract ledger blocks and replace with placeholders
    var blocks = [];
    var withoutLedger = md.replace(/```ledger\n([\s\S]*?)```/g, function(match, content) {
        blocks.push(renderLedgerBlock(content));
        return '<!--LEDGERBLOCK' + (blocks.length - 1) + '-->';
    });

    // Use marked.js for full Markdown rendering
    var html;
    if (typeof marked !== 'undefined' && marked.parse) {
        html = marked.parse(withoutLedger);
    } else {
        // Fallback basic markdown
        html = esc(withoutLedger)
            .replace(/^### (.+)$/gm, '<h3>$1</h3>')
            .replace(/^## (.+)$/gm, '<h2>$1</h2>')
            .replace(/^# (.+)$/gm, '<h1>$1</h1>')
            .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
            .replace(/`([^`]+)`/g, '<code>$1</code>')
            .replace(/\n\n/g, '</p><p>')
            .replace(/\n/g, '<br>');
    }

    // Restore ledger blocks
    blocks.forEach(function(blockHtml, i) {
        html = html.replace('<!--LEDGERBLOCK' + i + '-->', blockHtml);
    });

    return html;
}

function renderLedgerBlock(content) {
    var lines = content.trim().split('\n');
    var type = '', amount = '', category = '', note = '';
    lines.forEach(function(line) {
        var parts = line.split(':');
        if (parts.length < 2) return;
        var k = parts[0].trim().toLowerCase();
        var v = parts.slice(1).join(':').trim();
        if (k === 'type') type = v;
        else if (k === 'amount') amount = v;
        else if (k === 'category') category = v;
        else if (k === 'note') note = v;
    });
    var color = type === 'income' ? '#34c759' : '#ff3b30';
    return '<div class="tu-ledger-card">' +
        '<span class="tu-ledger-amount" style="color:' + color + '">' + esc(amount) + '</span>' +
        '<span class="tu-ledger-category">' + esc(category) + '</span>' +
        '<span class="tu-ledger-note">' + esc(note) + '</span></div>';
}

const Diary = { init, onTabEnter, onTabLeave };
export { Diary };
