function drawAuth(){const reg=S.mode==='up';
$('#auth').innerHTML=`<section class="hero"><div class="logo"><div class="dots"><i></i><i></i><i></i><i></i></div>Avito Start</div>
<div><h1>Больше, чем просто <span>AI-buddy</span></h1><p style="margin-top:16px;font-size:18px;max-width:520px">Наставник, заметки, доска и чек-лист адаптации в одном месте.</p></div>
<div class="stk"><div>💬 Ответы на вопросы</div><div>🔑 Доступы</div><div>📋 Заявки в один клик</div><div>📝 Заметки и доска</div><div>✅ Чек-лист и календарь</div></div></section>
<section class="form"><div class="tabs"><button class="${reg?'':'on'}" onclick="S.mode='in';drawAuth()">Вход</button><button class="${reg?'on':''}" onclick="S.mode='up';drawAuth()">Регистрация</button></div>
<h2 style="font-size:28px">${reg?'Создать аккаунт':'С возвращением'}</h2>${reg?'<label>Имя<input id="nm" placeholder="Анна Смирнова"></label>':''}
<label>Рабочая почта<input id="em" type="email" placeholder="name@company.test"></label><label>Пароль<input id="pw" type="password" placeholder="••••••••"></label>
<p class="err" id="ae" role="alert"></p><button class="btn" onclick="auth()">${reg?'Зарегистрироваться':'Войти'}</button>
<button class="btn sec" onclick="enter('Анна Смирнова','anna@company.test')">Войти как демо-стажёр</button><p class="mu">Прототип: данные синтетические, авторизация Авито подключается снаружи.</p></section>`}
function auth(){const em=$('#em').value.trim(),pw=$('#pw').value,nm=$('#nm')?$('#nm').value.trim():'';
if(!/^\S+@\S+\.\S+$/.test(em)){$('#ae').textContent='Введите почту в формате name@company.test';return}
if(pw.length<4){$('#ae').textContent='Пароль должен быть не короче 4 символов';return}enter(nm||em.split('@')[0],em)}
function enter(name,email){S.u={name,email};S.chats.ai.m=[['bot',`Привет, ${esc(name.split(' ')[0])}! Я AI-buddy. Подскажу по процессам, доступам и срокам. С чего начнём?`,'chips']];
$('#auth').hidden=true;$('#app').hidden=false;
$('.side').innerHTML=`<div class="logo"><div class="dots"><i></i><i></i><i></i><i></i></div>AI-buddy</div>`+NAV.map(n=>`<button class="nv" data-v="${n[0]}" onclick="go('${n[0]}')"><span>${n[1]}</span>${n[2]}</button>`).join('')+`<div class="me"><b>${esc(name)}</b><div class="mu">Стажёр · Backend</div></div>`;
$('.nav').innerHTML=NAV.map(n=>`<button data-v="${n[0]}" onclick="go('${n[0]}')"><b>${n[1]}</b>${n[2]}</button>`).join('');
const h=location.hash.slice(1);go(NAV.some(n=>n[0]===h)?h:'home')}
