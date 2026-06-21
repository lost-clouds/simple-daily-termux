import { API } from './constants.js';
import { esc } from './utils.js';

'use strict';

let _abortController=null;

function init() {
  var el=document.getElementById('ledgerAddBtn'); if(el) el.addEventListener('click',showForm);
  el=document.getElementById('ledgerSaveBtn'); if(el) el.addEventListener('click',saveEntry);
  el=document.getElementById('ledgerCancelBtn'); if(el) el.addEventListener('click',hideForm);
}

function abort() { if(_abortController){_abortController.abort();_abortController=null;} }
function refresh() { fetchSummary(); fetchList(); }

function showForm() {
  document.getElementById('ledgerForm').style.display='block';
  document.getElementById('ledgerEntryDate').value=new Date().toISOString().substring(0,10);
  document.getElementById('ledgerAmount').value=''; document.getElementById('ledgerNote').value='';
}
function hideForm() { document.getElementById('ledgerForm').style.display='none'; }

async function saveEntry() {
  var date=document.getElementById('ledgerEntryDate').value,
      type=document.getElementById('ledgerType').value,
      amount=parseFloat(document.getElementById('ledgerAmount').value),
      category=document.getElementById('ledgerCategory').value,
      note=document.getElementById('ledgerNote').value;
  if(!date||isNaN(amount)) return;
  try {
    var resp=await fetch(API.LEDGER,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({entry_date:date,type:type,amount:amount,category:category,note:note})});
    var json=await resp.json(); if(!json.ok){console.warn(json.error);return;}
    hideForm(); fetchSummary(); fetchList();
  } catch(e){console.warn(e);}
}

function fetchSummary() {
  abort(); _abortController=new AbortController();
  var now=new Date(), month=now.getFullYear()+'-'+String(now.getMonth()+1).padStart(2,'0');
  fetch(API.LEDGER_SUM+'?month='+month,{signal:_abortController.signal}).then(function(r){return r.json()}).then(function(j){
    if(!j.ok||!j.data) return; var d=j.data;
    setText('sumExpense','¥'+(d.expense||0).toFixed(2)); setText('sumIncome','¥'+(d.income||0).toFixed(2));
    setText('sumBalance','¥'+(d.balance||0).toFixed(2)); setText('sumSavings','¥'+(d.savings||0).toFixed(2));
  }).catch(function(e){if(e.name!=='AbortError')console.warn(e);});
}

function fetchList() {
  var now=new Date(), month=now.getFullYear()+'-'+String(now.getMonth()+1).padStart(2,'0');
  fetch(API.LEDGER+'?month='+month,{signal:_abortController?{signal:_abortController.signal}:{}}).then(function(r){return r.json()}).then(function(j){
    if(!j.ok) return; renderList(j.data);
  }).catch(function(e){if(e.name!=='AbortError')console.warn(e);});
}

function renderList(entries) {
  var el=document.getElementById('ledgerList'); if(!el) return;
  if(!entries||!entries.length){ el.innerHTML='<div class="tu-empty">本月暂无记录</div>'; return; }
  var h='';
  entries.forEach(function(e){
    var color=e.type==='income'?'#34c759':'#ff3b30', prefix=e.type==='income'?'+':'-';
    h+='<div class="tu-ledger-entry"><span class="tu-ledger-entry-cat">'+esc(e.category)+'</span>';
    if (e.note) h+='<span class="tu-text-xs tu-text-muted">'+esc(e.note)+'</span>';
    h+='<span class="tu-ledger-entry-amount" style="color:'+color+'">'+prefix+'¥'+e.amount.toFixed(2)+'</span></div>';
  });
  el.innerHTML=h;
}

function setText(id,t){ var el=document.getElementById(id); if(el) el.textContent=t||'--'; }

const Ledger = { init, fetchSummary, fetchList, abort, refresh };
export { Ledger };
