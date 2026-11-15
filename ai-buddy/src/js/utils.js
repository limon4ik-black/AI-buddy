const TODAY='2026-11-13',$=s=>document.querySelector(s),pad=n=>String(n).padStart(2,'0');
const esc=t=>String(t).replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]));
const fd=d=>d.split('-').reverse().slice(0,2).join('.');
const dd=d=>Math.round((new Date(d)-new Date(TODAY))/864e5);
const lab=n=>n<0?'просрочено':n===0?'сегодня':n===1?'завтра':'через '+n+' дн.';
const MN=['Январь','Февраль','Март','Апрель','Май','Июнь','Июль','Август','Сентябрь','Октябрь','Ноябрь','Декабрь'];
