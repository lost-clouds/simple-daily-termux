// Todo CRUD: list with filters, create/edit/delete, daily task auto-generation.
import { API } from './constants.js';
import { esc } from './utils.js';

'use strict';

let _abortController = null;

function init() {
  var el=document.getElementById('todoAddBtn'); if(el) el.addEventListener('click',showForm);
  el=document.getElementById('todoSaveBtn'); if(el) el.addEventListener('click',saveTodo);
  el=document.getElementById('todoCancelBtn'); if(el) el.addEventListener('click',hideForm);
  el=document.getElementById('todoFilter'); if(el) el.addEventListener('change',fetchList);
  el=document.getElementById('todoTypeFilter'); if(el) el.addEventListener('change',fetchList);
}

function abort() { if(_abortController){_abortController.abort();_abortController=null;} }
function refresh() { fetchList(); }

function showForm() {
  document.getElementById('todoFormWrap').style.display='block';
  document.getElementById('todoId').value=''; document.getElementById('todoTitle').value='';
  document.getElementById('todoNotes').value=''; document.getElementById('todoTaskType').value='one_time';
  document.getElementById('todoStatus').value='pending'; document.getElementById('todoDeadline').value='';
}
function hideForm() { document.getElementById('todoFormWrap').style.display='none'; }

async function saveTodo() {
  var id=document.getElementById('todoId').value;
  var title=document.getElementById('todoTitle').value.trim(); if(!title) return;
  var taskType=document.getElementById('todoTaskType').value;
  var notes=document.getElementById('todoNotes').value;
  var deadline=document.getElementById('todoDeadline').value;
  var body={title:title, task_type:taskType, notes:notes};
  if (deadline) body.deadline_at=new Date(deadline).toISOString();
  try {
    var url=id?API.TODOS+'/'+id:API.TODOS, method=id?'PUT':'POST';
    if (id) body.status=document.getElementById('todoStatus').value;
    var resp=await fetch(url,{method:method,headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
    if (!resp.ok) return; hideForm(); fetchList();
  } catch(e){console.warn(e);}
}

function fetchList() {
  abort();
  _abortController=new AbortController();
  var status=document.getElementById('todoFilter'),type=document.getElementById('todoTypeFilter');
  var qs=''; if(status&&status.value) qs+='&status='+status.value;
  if(type&&type.value) qs+='&task_type='+type.value;
  fetch(API.TODOS+'?1=1'+qs,{signal:_abortController.signal}).then(function(r){return r.json()}).then(function(j){
    if(!j.ok) return; renderList(j.data);
  }).catch(function(e){if(e.name!=='AbortError')console.warn(e);});
}

function renderList(todos) {
  var el=document.getElementById('todoList'); if(!el) return;
  if(!todos||!todos.length){ el.innerHTML='<div class="tu-empty">暂无任务</div>'; return; }
  var h='';
  todos.forEach(function(t){
    var done=t.status==='done'?' done':'';
    var typeLabel={'daily':'每','long_term':'长','one_time':'一'}[t.task_type]||'一';
    var typeCls='tu-todo-type-'+({'daily':'daily','long_term':'long','one_time':'one'}[t.task_type]||'one');
    var statusBadge=t.status==='done'?'tu-badge-done':(t.status==='doing'?'tu-badge-doing':'tu-badge-pending');
    var statusLabel=t.status==='done'?'完成':(t.status==='doing'?'进行中':'待办');
    h+='<div class="tu-todo-item'+done+'">';
    h+='<span class="tu-pri-dot tu-pri-'+['','I','II','III','IV'][t.priority]+'"></span>';
    h+='<span class="tu-todo-type-tag '+typeCls+'">'+typeLabel+'</span>';
    h+='<div class="tu-todo-content"><div class="tu-todo-title">'+esc(t.title)+'</div>';
    h+='<div class="tu-todo-meta"><span class="tu-badge '+statusBadge+'">'+statusLabel+'</span>';
    if(t.deadline_at) h+=' 截止:'+t.deadline_at.substring(0,10);
    if(t.entry_date) h+=' '+t.entry_date;
    h+='</div></div><div class="tu-todo-actions">';
    h+='<button class="tu-btn tu-btn-xs tu-btn-ghost tu-todo-done" data-id="'+t.id+'">✓</button>';
    h+='<button class="tu-btn tu-btn-xs tu-btn-ghost tu-todo-edit" data-id="'+t.id+'">✎</button>';
    h+='<button class="tu-btn tu-btn-xs tu-btn-danger tu-todo-del" data-id="'+t.id+'">✕</button>';
    h+='</div></div>';
  });
  el.innerHTML=h;
  el.querySelectorAll('.tu-todo-done').forEach(function(b){ b.addEventListener('click',function(){ toggleDone(this.getAttribute('data-id')); }); });
  el.querySelectorAll('.tu-todo-del').forEach(function(b){ b.addEventListener('click',function(){ deleteTodo(this.getAttribute('data-id')); }); });
  el.querySelectorAll('.tu-todo-edit').forEach(function(b){ b.addEventListener('click',function(){ editTodo(this.getAttribute('data-id')); }); });
}

async function toggleDone(id) {
  await fetch(API.TODOS+'/'+id,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:'done'})});
  fetchList();
}
async function deleteTodo(id) { if(!confirm('删除?'))return; await fetch(API.TODOS+'/'+id,{method:'DELETE'}); fetchList(); }
async function editTodo(id) {
  var resp=await fetch(API.TODOS+'/'+id), j=await resp.json();
  if(!j.ok||!j.data) return; var t=j.data;
  document.getElementById('todoId').value=t.id; document.getElementById('todoTitle').value=t.title;
  document.getElementById('todoNotes').value=t.notes||'';
  document.getElementById('todoTaskType').value=t.task_type||'one_time';
  var st=document.getElementById('todoStatus'); if(st) st.value=t.status||'pending';
  if(t.deadline_at){
    var d=new Date(t.deadline_at), pad=function(n){return String(n).padStart(2,'0');};
    document.getElementById('todoDeadline').value=d.getFullYear()+'-'+pad(d.getMonth()+1)+'-'+pad(d.getDate())+'T'+pad(d.getHours())+':'+pad(d.getMinutes());
  } else document.getElementById('todoDeadline').value='';
  document.getElementById('todoFormWrap').style.display='block';
}

const Todo = { init, fetchList, abort, refresh };
export { Todo };
