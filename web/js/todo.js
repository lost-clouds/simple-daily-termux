import { API } from './constants.js';
import { esc } from './utils.js';

'use strict';

function init() {
    var el = document.getElementById('todoAddBtn');
    if (el) el.addEventListener('click', showForm);
    el = document.getElementById('todoSaveBtn');
    if (el) el.addEventListener('click', saveTodo);
    el = document.getElementById('todoCancelBtn');
    if (el) el.addEventListener('click', hideForm);
    el = document.getElementById('todoFilter');
    if (el) el.addEventListener('change', fetchList);
}

function onTabEnter() { if (!document.hidden) fetchList(); }
function onTabLeave() {}

function showForm() {
    document.getElementById('todoFormWrap').style.display = 'block';
    document.getElementById('todoId').value = '';
    document.getElementById('todoTitle').value = '';
    document.getElementById('todoDeadline').value = '';
    document.getElementById('todoPriority').value = '0';
}

function hideForm() {
    document.getElementById('todoFormWrap').style.display = 'none';
}

async function saveTodo() {
    const id = document.getElementById('todoId').value;
    const title = document.getElementById('todoTitle').value.trim();
    if (!title) return;

    const deadline = document.getElementById('todoDeadline').value;
    const priority = parseInt(document.getElementById('todoPriority').value) || 0;

    const body = {
        title: title,
        priority: priority
    };
    if (deadline) body.deadline_at = new Date(deadline).toISOString();

    try {
        const url = id ? API.TODOS + '/' + id : API.TODOS;
        const method = id ? 'PUT' : 'POST';
        if (id) body.status = document.getElementById('todoStatus') ? document.getElementById('todoStatus').value : 'pending';
        const resp = await fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        });
        if (!resp.ok) return;
        hideForm();
        fetchList();
    } catch (e) { console.warn('Todo save:', e.message); }
}

async function fetchList() {
    try {
        const filter = document.getElementById('todoFilter');
        let qs = '';
        if (filter && filter.value) qs += '?status=' + filter.value;

        const resp = await fetch(API.TODOS + qs);
        const json = await resp.json();
        if (!json.ok) return;
        renderList(json.data);
    } catch (e) { console.warn('Todo list:', e.message); }
}

function renderList(todos) {
    const el = document.getElementById('todoList');
    if (!el) return;

    if (!todos || todos.length === 0) {
        el.innerHTML = '<div class="tu-empty">暂无待办事项</div>';
        return;
    }

    let html = '';
    todos.forEach(function(t) {
        const done = t.status === 'done' ? ' done' : '';
        const checked = t.status === 'done' ? ' checked' : '';
        const badgeClass = t.status === 'done' ? 'tu-badge-done' : (t.status === 'doing' ? 'tu-badge-doing' : 'tu-badge-pending');
        const statusLabel = t.status === 'done' ? '完成' : (t.status === 'doing' ? '进行中' : '待办');

        html += '<div class="tu-todo-item' + done + '" data-id="' + t.id + '">';
        html += '<div class="tu-todo-checkbox' + checked + '" data-id="' + t.id + '" onclick="this.checked_=true"></div>';
        html += '<div class="tu-todo-content">';
        html += '<div class="tu-todo-title">' + esc(t.title) + '</div>';
        html += '<div class="tu-todo-meta">';
        html += '<span class="tu-badge ' + badgeClass + '">' + statusLabel + '</span>';
        if (t.priority > 0) html += ' P' + t.priority;
        if (t.deadline_at) html += ' 截止: ' + t.deadline_at.substring(0, 10);
        html += '</div></div>';
        html += '<div class="tu-todo-actions">';
        html += '<button class="tu-btn tu-btn-xs tu-btn-ghost tu-todo-edit" data-id="' + t.id + '">编辑</button>';
        html += '<button class="tu-btn tu-btn-xs tu-btn-danger tu-todo-del" data-id="' + t.id + '">删除</button>';
        html += '</div></div>';
    });
    el.innerHTML = html;

    // bind events
    el.querySelectorAll('.tu-todo-checkbox').forEach(function(cb) {
        cb.addEventListener('click', function() { toggleDone(this.getAttribute('data-id')); });
    });
    el.querySelectorAll('.tu-todo-del').forEach(function(btn) {
        btn.addEventListener('click', function() { deleteTodo(this.getAttribute('data-id')); });
    });
    el.querySelectorAll('.tu-todo-edit').forEach(function(btn) {
        btn.addEventListener('click', function() { editTodo(this.getAttribute('data-id')); });
    });
}

async function toggleDone(id) {
    try {
        await fetch(API.TODOS + '/' + id, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: 'done' })
        });
        fetchList();
    } catch (e) { console.warn(e); }
}

async function deleteTodo(id) {
    if (!confirm('确定删除？')) return;
    try {
        await fetch(API.TODOS + '/' + id, { method: 'DELETE' });
        fetchList();
    } catch (e) { console.warn(e); }
}

async function editTodo(id) {
    try {
        const resp = await fetch(API.TODOS + '/' + id);
        const json = await resp.json();
        if (!json.ok || !json.data) return;
        const t = json.data;
        document.getElementById('todoId').value = t.id;
        document.getElementById('todoTitle').value = t.title;
        document.getElementById('todoPriority').value = t.priority;
        var st = document.getElementById('todoStatus');
        if (st) st.value = t.status || 'pending';
        if (t.deadline_at) {
            document.getElementById('todoDeadline').value = t.deadline_at.substring(0, 16);
        } else {
            document.getElementById('todoDeadline').value = '';
        }
        document.getElementById('todoFormWrap').style.display = 'block';
    } catch (e) { console.warn(e); }
}

const Todo = { init, fetchList, onTabEnter, onTabLeave };
export { Todo };
