// Countdown events: manual creation, days-left display, urgency styling.
import { API } from './constants.js';
import { esc } from './utils.js';

'use strict';

let _abortController=null;

function init() {
  var el=document.getElementById('cdAddBtn'); if(el) el.addEventListener('click',showForm);
  el=document.getElementById('cdSaveBtn'); if(el) el.addEventListener('click',saveCD);
  el=document.getElementById('cdCancelBtn'); if(el) el.addEventListener('click',hideForm);
}

function abort() { if(_abortController){_abortController.abort();_abortController=null;} }
function refresh() { fetchList(); }

function showForm() { document.getElementById('cdForm').style.display='block'; document.getElementById('cdTitle').value=''; document.getElementById('cdTarget').value=''; }
function hideForm() { document.getElementById('cdForm').style.display='none'; }

async function saveCD() {
  var title=document.getElementById('cdTitle').value.trim(), targetAt=document.getElementById('cdTarget').value;
  if(!title||!targetAt) return;
  try {
    await fetch(API.COUNTDOWN,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({title:title,target_at:new Date(targetAt).toISOString()})});
    hideForm(); fetchList();
  } catch(e){console.warn(e);}
}

function fetchList() {
  abort();
  _abortController=new AbortController();
  fetch(API.COUNTDOWN,{signal:_abortController.signal}).then(function(r){return r.json()}).then(function(j){
    if(!j.ok) return; renderList(j.data);
  }).catch(function(e){if(e.name!=='AbortError')console.warn(e);});
}

function renderList(events) {
  var el=document.getElementById('cdList'); if(!el) return;
  if(!events||!events.length){ el.innerHTML='<div class="tu-empty">暂无倒计时</div>'; return; }
  var h='';
  events.forEach(function(e){
    var urgent=e.days_left<=7?' urgent':'', label=e.days_left<0?'已过去'+Math.abs(e.days_left)+'天':'还有'+e.days_left+'天';
    h+='<div class="tu-cd-item'+urgent+'"><div class="tu-cd-days">'+e.days_left+'</div><div class="tu-cd-info"><div class="tu-cd-title">'+esc(e.title)+'</div><div class="tu-cd-date">'+label+' · '+e.target_at.substring(0,10)+'</div></div>';
    if(e.source==='manual') h+='<button class="tu-btn tu-btn-xs tu-btn-danger tu-cd-del" data-id="'+e.id+'">✕</button>';
    h+='</div>';
  });
  el.innerHTML=h;
  el.querySelectorAll('.tu-cd-del').forEach(function(b){b.addEventListener('click',function(){deleteCD(this.getAttribute('data-id'));});});
}

async function deleteCD(id) { if(!confirm('删除?'))return; await fetch(API.COUNTDOWN+'/'+id,{method:'DELETE'}); fetchList(); }

const Countdown = { init, fetchList, abort, refresh };
export { Countdown };
