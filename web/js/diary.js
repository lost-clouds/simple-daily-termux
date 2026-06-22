// Diary editor: Markdown with live preview, mood picker, embedded ledger block rendering.
import { API } from './constants.js';
import { esc } from './utils.js';

'use strict';

let _currentDate=new Date().toISOString().substring(0,10);
let _abortController=null;
const onSaveCallbacks = [];

function init() {
  var el=document.getElementById('diaryDate'); if(el){el.value=_currentDate;el.addEventListener('change',function(){_currentDate=this.value;loadDiary();});}
  el=document.getElementById('diarySaveBtn'); if(el) el.addEventListener('click',saveDiary);
  el=document.getElementById('diaryExportBtn'); if(el) el.addEventListener('click',_exportMD);
  el=document.getElementById('diaryImportBtn'); if(el) el.addEventListener('click',_onImportClick);
  document.querySelectorAll('.tu-diary-mood-btn').forEach(function(b){b.addEventListener('click',function(){document.querySelectorAll('.tu-diary-mood-btn').forEach(function(x){x.classList.remove('active');});this.classList.add('active');});});
}

// -- export/import --
async function _exportMD() {
  var m=_currentDate.substring(0,7);
  var resp=await fetch(API.DIARY_EXPORT+'?month='+m);
  if(!resp.ok) return;
  var blob=await resp.blob();
  var a=document.createElement('a');
  a.href=URL.createObjectURL(blob);
  a.download='diary-'+m+'.md';
  a.click();
  URL.revokeObjectURL(a.href);
}
function _onImportClick() {
  var input=document.createElement('input');
  input.type='file'; input.accept='.md,.txt';
  input.onchange=function(){ if(this.files[0]) _importMD(this.files[0]); };
  input.click();
}
async function _importMD(file) {
  var form=new FormData(); form.append('file',file);
  var resp=await fetch(API.DIARY_IMPORT,{method:'POST',body:form});
  var json=await resp.json();
  if(json.ok){ alert('导入 '+json.data.imported+' 条日记'); loadDiary(); }
  else alert('导入失败: '+(json.error||'未知错误'));
}

function abort() { if(_abortController){_abortController.abort();_abortController=null;} }
function refresh() { loadDiary(); }

function loadDiary() {
  abort(); _abortController=new AbortController();
  fetch(API.DIARY+'/'+_currentDate,{signal:_abortController.signal}).then(function(r){return r.json()}).then(function(j){
    var ce=document.getElementById('diaryContent'), pe=document.getElementById('diaryPreview');
    if(j.ok&&j.data){ if(ce) ce.value=j.data.content_md||''; if(pe) pe.innerHTML=mdPreview(j.data.content_md||'');
      if(j.data.mood){ document.querySelectorAll('.tu-diary-mood-btn').forEach(function(b){b.classList.toggle('active',b.getAttribute('data-mood')===j.data.mood);}); }
    } else { if(ce) ce.value=''; if(pe) pe.innerHTML='<span class="tu-text-muted">无</span>'; }
  }).catch(function(e){if(e.name!=='AbortError')console.warn(e);});
}

async function saveDiary() {
  var ce=document.getElementById('diaryContent'); if(!ce) return;
  var md=ce.value, moodBtn=document.querySelector('.tu-diary-mood-btn.active'), mood=moodBtn?moodBtn.getAttribute('data-mood'):'';
  try {
    var resp=await fetch(API.DIARY+'/'+_currentDate,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({content_md:md,mood:mood})});
    var json=await resp.json();
    if(json.ok){ var pe=document.getElementById('diaryPreview'); if(pe) pe.innerHTML=mdPreview(md); onSaveCallbacks.forEach(function(cb){cb();}); }
  } catch(e){console.warn(e);}
}

function mdPreview(md) {
  if(!md) return '';
  var blocks=[];
  var withoutLedger=md.replace(/```ledger\n([\s\S]*?)```/g,function(m,c){blocks.push(renderLedgerBlock(c));return'<!--LB'+blocks.length+'-->';});
  var html;
  if(typeof marked!=='undefined'&&marked.parse){ html=marked.parse(withoutLedger); }
  else { html=esc(withoutLedger).replace(/^### (.+)$/gm,'<h3>$1</h3>').replace(/^## (.+)$/gm,'<h2>$1</h2>').replace(/^# (.+)$/gm,'<h1>$1</h1>').replace(/\*\*(.+?)\*\*/g,'<strong>$1</strong>').replace(/`([^`]+)`/g,'<code>$1</code>').replace(/\n\n/g,'</p><p>').replace(/\n/g,'<br>'); }
  blocks.forEach(function(bh,i){html=html.replace('<!--LB'+(i+1)+'-->',bh);});
  return html;
}

function renderLedgerBlock(content) {
  var lines=content.trim().split('\n'), type='',amount='',category='',note='';
  lines.forEach(function(l){var p=l.split(':');if(p.length<2)return;var k=p[0].trim().toLowerCase(),v=p.slice(1).join(':').trim();if(k==='type')type=v;else if(k==='amount')amount=v;else if(k==='category')category=v;else if(k==='note')note=v;});
  var color=type==='income'?'#34c759':'#ff3b30';
  return'<div class="tu-ledger-card"><span class="tu-ledger-amount" style="color:'+color+'">'+esc(amount)+'</span><span class="tu-ledger-category">'+esc(category)+'</span><span class="tu-ledger-note">'+esc(note)+'</span></div>';
}

const Diary = { init, loadDiary, mdPreview, abort, refresh, onSaveCallbacks };
export { Diary };
