import { API } from './constants.js';

'use strict';

let _timer=null, _secondsLeft=0, _running=false, _sessionId='', _sessionType='focus', _tabHidden=false;

function init() {
  var el=document.getElementById('pomoStart'); if(el) el.addEventListener('click',start);
  el=document.getElementById('pomoPause'); if(el) el.addEventListener('click',pause);
  el=document.getElementById('pomoReset'); if(el) el.addEventListener('click',reset);
  el=document.getElementById('pomoRestOverlay'); if(el) el.addEventListener('click',startRest);
  document.querySelectorAll('.tu-pomo-preset-btn').forEach(function(b){
    b.addEventListener('click',function(){ if(!_running){_secondsLeft=parseInt(this.getAttribute('data-minutes'))*60;updateDisplay();} });
  });
  _secondsLeft=25*60; updateDisplay();
}

function abort() {}
function refresh() { fetchStats(); if(_tabHidden){_tabHidden=false;} }

function updateDisplay() {
  var m=Math.floor(_secondsLeft/60), s=_secondsLeft%60;
  var el=document.getElementById('pomoTimer'); if(el) el.textContent=String(m).padStart(2,'0')+':'+String(s).padStart(2,'0');
}
function tick() { _secondsLeft--; updateDisplay(); if(_secondsLeft<=0) finish('completed'); }

async function start() {
  if(_running) return;
  _sessionType='focus';
  try {
    var resp=await fetch(API.POMO_START,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({planned_minutes:Math.ceil(_secondsLeft/60)})});
    var json=await resp.json(); if(!json.ok) return;
    _sessionId=json.data.id; _running=true;
    document.getElementById('pomoStart').style.display='none';
    document.getElementById('pomoPause').style.display='';
    document.getElementById('pomoRestOverlay').style.display='none';
    _timer=setInterval(tick,1000);
  } catch(e){console.warn(e);}
}

async function pause() {
  _running=false; if(_timer){clearInterval(_timer);_timer=null;}
  if(_sessionId){
    await fetch(API.POMODORO+'/'+_sessionId+'/finish',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:'aborted'})});
    _sessionId='';
  }
  document.getElementById('pomoStart').style.display='';
  document.getElementById('pomoPause').style.display='none';
  document.getElementById('pomoRestOverlay').style.display='none';
  fetchStats();
}

async function finish(status) {
  _running=false; if(_timer){clearInterval(_timer);_timer=null;}
  if(_sessionId){
    await fetch(API.POMODORO+'/'+_sessionId+'/finish',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:status})});
    _sessionId='';
  }
  document.getElementById('pomoStart').style.display='';
  document.getElementById('pomoPause').style.display='none';
  fetchStats();
  if (_sessionType==='focus' && status==='completed') {
    var overlay=document.getElementById('pomoRestOverlay');
    if(overlay) overlay.style.display='flex';
  }
}

async function startRest() {
  _sessionType='rest'; _secondsLeft=5*60; updateDisplay();
  document.getElementById('pomoRestOverlay').style.display='none';
  try {
    var resp=await fetch(API.POMO_REST,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({planned_minutes:5})});
    var json=await resp.json(); if(!json.ok) return;
    _sessionId=json.data.id; _running=true;
    document.getElementById('pomoStart').style.display='none';
    document.getElementById('pomoPause').style.display='';
    _timer=setInterval(tick,1000);
  } catch(e){console.warn(e);}
}

async function reset() {
  if(_running) await pause();
  _secondsLeft=25*60; _sessionId=''; updateDisplay();
  document.getElementById('pomoRestOverlay').style.display='none';
  fetchStats();
}

async function fetchStats() {
  try {
    var resp=await fetch(API.POMO_TODAY); if(!resp.ok) return;
    var json=await resp.json(); if(!json.ok) return;
    var f=document.getElementById('pomoFocusMin'),r=document.getElementById('pomoRestMin');
    if(f) f.textContent=json.data.total_minutes||0;
    if(r) r.textContent=json.data.rest_minutes||0;
  } catch(e){console.warn(e);}
}

const Pomodoro = { init, fetchStats, abort, refresh };
export { Pomodoro };
