function home(){const n=S.tasks.length,p=Math.round(done()/n*100),h=hot(),ns=[...S.notes].slice(0,3);
return `<div class="wrap"><div><h1 style="font-size:34px;letter-spacing:-.02em">Привет, ${esc(S.u.name.split(' ')[0])} 👋</h1><p class="mu">Пятница, 13 ноября</p></div>
<div class="hg"><div class="col">
<div class="card tg"><div class="row sp"><b>Процесс адаптации</b><b>${p}%</b></div><div class="bar" style="margin:10px 0;background:rgba(255,255,255,.55)"><i style="width:${p}%"></i></div><p>Выполнено ${done()} из ${n} шагов. <a href="#list" onclick="go('list');return false" style="color:inherit;font-weight:700">Открыть чек-лист</a></p></div>
<div class="card"><h3 style="margin-bottom:12px">🔥 Горят сроки</h3><div class="col" style="gap:8px">${h.length?h.map(tkHtml).join(''):'<p class="mu">Горящих задач нет, отлично!</p>'}</div></div>
<div class="card"><h3 style="margin-bottom:10px">Быстрые вопросы</h3><div class="chips">${chips()}</div></div></div>
<div class="col"><div class="card">${cal(1)}</div>
<div class="card"><div class="row sp" style="margin-bottom:10px"><h3>Последние заметки</h3><button class="chip" onclick="go('notes')">Все</button></div><div class="col" style="gap:8px">${ns.map(x=>`<button class="nt ${x.c}" onclick="S.nopen=${x.id};go('notes',1)"><b>${esc(x.t)}</b><span>${esc(x.b)}</span></button>`).join('')}</div></div></div></div></div>`}
