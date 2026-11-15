function list(){const ts=[...S.tasks].sort((a,b)=>(a.done?1:0)-(b.done?1:0)||(a.d<b.d?-1:1));
return `<div class="wrap"><div class="row sp wrapf"><div><h1 style="font-size:30px">Чек-лист</h1><p class="mu">Выполнено ${done()} из ${S.tasks.length}</p></div><div class="sw"><button class="${S.tv==='list'?'on':''}" onclick="S.tv='list';draw()">Список</button><button class="${S.tv==='cal'?'on':''}" onclick="S.tv='cal';draw()">Календарь</button></div></div>
<div class="card"><div class="row wrapf"><input id="nt" placeholder="Новая задача" style="flex:2;min-width:180px"><input id="nd" type="date" value="${TODAY}" style="flex:1;min-width:140px"><button class="btn" onclick="addTask()">Добавить</button></div></div>
${S.tv==='list'?`<div class="col" style="gap:8px">${ts.map(tkHtml).join('')}</div>`:`<div class="card">${cal(0)}</div>`}</div>`}
function addTask(){const t=$('#nt').value.trim(),d=$('#nd').value;if(!t||!d)return;S.tasks.push({id:Date.now(),t,d,tag:'Мои'});draw()}
