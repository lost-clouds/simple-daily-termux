'use strict';
// NOTE: paths are relative to work through nginx /simpledaily/ proxy.
// When SPA is served at /simpledaily/, relative 'api/todos' resolves to /simpledaily/api/todos
const API = {
    TODOS:      'api/todos',
    COUNTDOWN:  'api/countdown',
    POMODORO:   'api/pomodoro',
    POMO_START: 'api/pomodoro/start',
    POMO_REST:  'api/pomodoro/start-rest',
    POMO_TODAY: 'api/pomodoro/today',
    DIARY:      'api/diary',
    LEDGER:     'api/ledger',
    LEDGER_SUM: 'api/ledger/summary',
    CALENDAR:   'api/calendar',
    SUMMARY:    'api/summary'
};
export { API };
