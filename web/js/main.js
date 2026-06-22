// Application shell: tab navigation, home page orchestration, calendar rendering.
import { Theme } from './theme.js';
import { esc } from './utils.js';
import { API } from './constants.js';
import { Todo } from './todo.js';
import { Countdown } from './countdown.js';
import { Pomodoro } from './pomodoro.js';
import { Diary } from './diary.js';
import { Ledger } from './ledger.js';

'use strict';

const TABS = ['home','calendar','todo','pomodoro','diary','countdown'];
let _currentTab = 'home';
let _selectedDate = new Date().toISOString().substring(0,10);
let _homeCalYear, _homeCalMonth, _pageCalYear, _pageCalMonth;

const MODULES = { Todo, Countdown, Pomodoro, Diary, Ledger };

/* === navigation === */
function switchTab(tabId) {
  if (tabId === _currentTab) return;
  // abort previous module
  var prevMod = MODULES[_currentTab.charAt(0).toUpperCase()+_currentTab.slice(1)];
  if (prevMod && prevMod.abort) prevMod.abort();

  _currentTab = tabId;
  document.querySelectorAll('.tu-section').forEach(function(s){ s.classList.toggle('active', s.id==='sec-'+tabId); });
  document.querySelectorAll('.tu-nav-btn').forEach(function(b){ b.classList.toggle('active', b.getAttribute('data-tab')===tabId); });
  document.querySelectorAll('.tu-bottom-nav .tu-nav-btn').forEach(function(b){ b.classList.toggle('active', b.getAttribute('data-tab')===tabId); });
  window.location.hash = tabId;

  // enter new module
  var mod = MODULES[tabId.charAt(0).toUpperCase()+tabId.slice(1)];
  if (mod && mod.refresh) { mod.refresh(); }
  if (tabId==='home') { loadHome(); }
  if (tabId==='calendar') { CalendarPage.init(); }
  if (tabId==='diary') { Diary.loadDiary(); }
}

function onNavClick(e) {
  var btn = e.target.closest('.tu-nav-btn');
  if (!btn) return;
  switchTab(btn.getAttribute('data-tab'));
}

/* === Home page === */
function loadHome() {
  var today = new Date().toISOString().substring(0,10);
  fetch(API.TODOS+'/ensure-daily', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({date:today})})
    .then(function(){ return loadHomeTodos(); })
    .then(function(){
      // calendar, diary, stats in parallel
      return Promise.all([initCalendarHomeAsync(), loadHomeDiaryAsync(), loadHomeStatsAsync()]);
    }).catch(function(e){ console.warn('Home load:', e); });
}

function loadHomeTodos() {
  return fetch(API.TODOS+'?entry_date='+_selectedDate).then(function(r){return r.json()}).then(function(j){
    if (!j.ok) return;
    var todos = j.data || [];
    var el = document.getElementById('homeTodoList');
    if (!el) return;
    if (!todos.length) { el.innerHTML='<div class="tu-empty">今日无待办</div>'; return; }
    var h='';
    todos.forEach(function(t){
      if (t.status==='done') return;
      h+='<div class="tu-todo-item" style="font-size:0.75rem;padding:0.2rem 0.3rem;" data-id="'+esc(t.id)+'">';
      h+='<span class="tu-pri-dot tu-pri-'+roman(t.priority)+'"></span>';
      h+='<span style="flex:1">'+esc(t.title)+'</span>';
      if (t.deadline_at) h+='<span class="tu-text-xs tu-text-muted">'+t.deadline_at.substring(0,10)+'</span>';
      h+='</div>';
    });
    el.innerHTML = h || '<div class="tu-empty">全部完成!</div>';
    el.querySelectorAll('.tu-todo-item').forEach(function(item){
      item.addEventListener('click', function(){
        var id = this.getAttribute('data-id');
        fetch(API.TODOS+'/'+id, {method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:'done'})})
          .then(function(){ loadHomeTodos(); });
      });
    });
  });
}

function initCalendarHomeAsync() {
  var now = new Date(); _homeCalYear=now.getFullYear(); _homeCalMonth=now.getMonth()+1;
  return fetchCalendar('calGrid','calTitle','calEventsList', function(date){
    _selectedDate=date;
    document.getElementById('homeDateLabel').textContent=date+' 待办';
    loadHomeTodos(); loadHomeDiaryAsync();
  }, _homeCalYear, _homeCalMonth);
}

function loadHomeDiaryAsync() {
  return fetch(API.DIARY+'/'+_selectedDate).then(function(r){return r.json()}).then(function(j){
    var el = document.getElementById('homeDiaryPreview');
    if (!el) return;
    if (!j.data) { el.innerHTML='<span class="tu-text-muted">今天还没写日记</span>'; return; }
    var md = j.data.content_md || '';
    var clean = md.replace(/```ledger[\s\S]*?```/g,'');
    var preview = Array.from(clean).slice(0,300).join('');
    if (clean.length > 300) preview += '…';
    el.innerHTML = Diary.mdPreview(preview);
    loadHomeLedger();
  });
}

function loadHomeLedger() {
  fetch(API.LEDGER+'?month='+_selectedDate.substring(0,7)).then(function(r){return r.json()}).then(function(j){
    if (!j.ok) return;
    var entries = j.data || [];
    var exp=0, inc=0;
    entries.forEach(function(e){
      if (e.entry_date!==_selectedDate) return;
      if (e.type==='expense') exp+=e.amount_cents;
      else inc+=e.amount_cents;
    });
    var ee=document.getElementById('homeExpense'), ei=document.getElementById('homeIncome');
    if (ee) ee.innerHTML = '¥'+(exp/100).toFixed(2);
    if (ei) ei.innerHTML = '¥'+(inc/100).toFixed(2);
  }).catch(function(){});
}

function loadHomeStatsAsync() {
  return fetch(API.POMO_TODAY).then(function(r){return r.json()}).then(function(j){
    if (!j.ok) return;
    var f=document.getElementById('homeFocusMin'), r=document.getElementById('homeRestMin');
    if (f) f.textContent = j.data.total_minutes||0;
    if (r) r.textContent = j.data.rest_minutes||0;
  }).catch(function(){});
}

/* === Calendar rendering === */
function fetchCalendar(gridId, titleId, eventsId, clickCb, calYear, calMonth) {
  var m = String(calMonth).padStart(2,'0');
  return fetch(API.CALENDAR+'?month='+calYear+'-'+m).then(function(r){return r.json()}).then(function(j){
    if (!j.ok) return;
    renderCalendarGrid(gridId, titleId, eventsId, j.data, clickCb, calYear, calMonth);
  });
}

function renderCalendarGrid(gridId, titleId, eventsId, data, clickCb, calYear, calMonth) {
  var grid=document.getElementById(gridId), title=document.getElementById(titleId);
  if (!grid) return;
  if (title) title.textContent=calYear+'年'+calMonth+'月';
  var firstDay = new Date(calYear,calMonth-1,1).getDay();
  var daysInMonth = new Date(calYear,calMonth,0).getDate();
  var diaryDates = new Set((data.diary_dates||[]).map(function(d){return d.entry_date||d;}));
  var todayStr = new Date().toISOString().substring(0,10);
  var eventMap = {};
  (data.todo_deadlines||[]).forEach(function(td){
    if (!td.deadline_at) return; var d=td.deadline_at.substring(0,10);
    if (!eventMap[d]) eventMap[d]=[]; eventMap[d].push({title:td.title,type:'todo'});
  });
  (data.countdown_targets||[]).forEach(function(ct){
    var d=ct.target_at.substring(0,10); if (!eventMap[d]) eventMap[d]=[];
    eventMap[d].push({title:ct.title,type:'cd'});
  });
  var html='', headers=['日','一','二','三','四','五','六'];
  headers.forEach(function(h){ html+='<div class="tu-calendar-day-header">'+h+'</div>'; });
  for (var i=0;i<firstDay;i++) html+='<div class="tu-calendar-day other-month"></div>';
  for (var d=1;d<=daysInMonth;d++){
    var ds=calYear+'-'+String(calMonth).padStart(2,'0')+'-'+String(d).padStart(2,'0');
    var cls='tu-calendar-day'; if (ds===todayStr) cls+=' today';
    html+='<div class="'+cls+'" data-date="'+ds+'">';
    html+='<div class="tu-calendar-day-num">'+d+'</div>';
    var evts=eventMap[ds]||[];
    for (var ei=0;ei<Math.min(evts.length,3);ei++){
      var text=evts[ei].title, trunc=text.length>8?text.substring(0,7)+'…':text;
      html+='<div class="tu-calendar-event-preview has-event">'+esc(trunc)+'</div>';
    }
    if (diaryDates.has(ds)) html+='<div class="tu-calendar-dot diary"></div>';
    html+='</div>';
  }
  grid.innerHTML=html;
  if (clickCb) {
    grid.querySelectorAll('.tu-calendar-day').forEach(function(cell){
      cell.addEventListener('click',function(){
        var date=this.getAttribute('data-date');
        if (date) clickCb(date);
      });
    });
  }
  var evList=document.getElementById(eventsId);
  if (evList && data.events && data.events.length){
    var ehtml=''; data.events.forEach(function(ev){
      ehtml+='<div class="tu-calendar-event-item"><span class="tu-calendar-event-date">'+(ev.start_at||'').substring(0,10)+'</span><span>'+esc(ev.title)+'</span></div>';
    });
    evList.innerHTML=ehtml;
  } else if (evList) { evList.innerHTML=''; }
}

var CalendarPage = {
  init: function(){
    var now = new Date(); _pageCalYear=now.getFullYear(); _pageCalMonth=now.getMonth()+1;
    var m = String(_pageCalMonth).padStart(2,'0');
    fetch(API.CALENDAR+'?month='+_pageCalYear+'-'+m).then(function(r){return r.json()}).then(function(j){
      if (!j.ok) return;
      renderCalendarGrid('calPageGrid','calPageTitle','calPageEventsList', j.data, null, _pageCalYear, _pageCalMonth);
    });
  },
  abort: function(){}
};

/* === Init === */
function init() {
  Theme.initTheme();
  document.getElementById('themeToggleBtn').addEventListener('click',function(){Theme.toggleTheme();});

  // Wire diary save callback
  Diary.onSaveCallbacks.push(function(){ Ledger.fetchSummary(); Ledger.fetchList(); });

  document.querySelectorAll('.tu-nav-bar,.tu-bottom-nav').forEach(function(b){ b.addEventListener('click',onNavClick); });

  // Home calendar nav
  var cp=document.getElementById('calPrev'), cn=document.getElementById('calNext');
  if(cp) cp.addEventListener('click',function(){ _homeCalMonth--; if(_homeCalMonth<1){_homeCalMonth=12;_homeCalYear--;} fetchCalendar('calGrid','calTitle','calEventsList',function(d){_selectedDate=d;document.getElementById('homeDateLabel').textContent=d+'待办';loadHomeTodos();loadHomeDiaryAsync();},_homeCalYear,_homeCalMonth); });
  if(cn) cn.addEventListener('click',function(){ _homeCalMonth++; if(_homeCalMonth>12){_homeCalMonth=1;_homeCalYear++;} fetchCalendar('calGrid','calTitle','calEventsList',function(d){_selectedDate=d;document.getElementById('homeDateLabel').textContent=d+'待办';loadHomeTodos();loadHomeDiaryAsync();},_homeCalYear,_homeCalMonth); });

  // Calendar page nav
  var cpp=document.getElementById('calPagePrev'), cpn=document.getElementById('calPageNext');
  if(cpp) cpp.addEventListener('click',function(){ _pageCalMonth--; if(_pageCalMonth<1){_pageCalMonth=12;_pageCalYear--;} CalendarPage.init(); });
  if(cpn) cpn.addEventListener('click',function(){ _pageCalMonth++; if(_pageCalMonth>12){_pageCalMonth=1;_pageCalYear++;} CalendarPage.init(); });

  // Home buttons
  var hta=document.getElementById('homeTodoAddBtn'), hde=document.getElementById('homeDiaryEditBtn');
  if(hta) hta.addEventListener('click',function(){ switchTab('todo'); });
  if(hde) hde.addEventListener('click',function(){ switchTab('diary'); });

  Todo.init();
  Countdown.init();
  Pomodoro.init();
  Diary.init();
  Ledger.init();

  // visibilitychange
  document.addEventListener('visibilitychange', function(){
    if (document.hidden) {
      var mod = MODULES[_currentTab.charAt(0).toUpperCase()+_currentTab.slice(1)];
      if (mod && mod.abort) mod.abort();
    } else {
      var mod = MODULES[_currentTab.charAt(0).toUpperCase()+_currentTab.slice(1)];
      if (mod && mod.refresh) mod.refresh();
      if (_currentTab==='home') loadHome();
      if (_currentTab==='calendar') CalendarPage.init();
      if (_currentTab==='diary') Diary.loadDiary();
    }
  });

  var hash=window.location.hash.replace('#','');
  var initialTab=(hash&&TABS.indexOf(hash)!==-1)?hash:'home';
  if (initialTab!=='home') switchTab(initialTab); else loadHome();

  window.addEventListener('hashchange',function(){
    var h=window.location.hash.replace('#','');
    if (h&&TABS.indexOf(h)!==-1) switchTab(h);
  });
}

function roman(n){ return ['','I','II','III','IV'][n]||'IV'; }

if (document.readyState==='loading') document.addEventListener('DOMContentLoaded',init); else init();
