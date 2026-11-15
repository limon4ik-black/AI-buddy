// Имитация сервиса отсутствий (заявка на отгул)
function submitLeave(){const a=$('#d1').value,b=$('#d2').value,r=$('#rs').value.trim();
if(!a||!b||b<a){$('#fe').textContent='Проверьте даты: окончание не может быть раньше начала';return}
const bt=$('#fb');bt.disabled=true;bt.textContent='Отправляем в сервис отсутствий…';
setTimeout(()=>{S.leave={id:'OFF-1042',a,b,r};S.tasks.push({id:99,t:'Отгул: заявка на согласовании',d:a,tag:'Заявки'});
const m=S.chats.ai.m,i=m.findIndex(x=>x[2]==='form');if(i>-1)m[i]=['bot','Заявка отправлена. Сервис вернул номер и статус, я сохранил их.','result'];
m.push(['bot','Руководитель Мария Петрова рассмотрит заявку. Дата отгула добавлена в календарь чек-листа.']);drawMsgs()},1100)}
