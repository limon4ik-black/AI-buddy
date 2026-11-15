function openBoard(){$('#bd').hidden=false;drawBoard()}
function drawBoard(){$('#bd').innerHTML=`<div class="bar2"><b>Доска</b><span class="mu">Перетаскивайте стикеры за верхнюю полоску</span><span style="flex:1"></span>${['#FFF1B8','#B8DAFB','#C8FFC0','#FFD6DA'].map(c=>`<button class="chip" style="background:${c};color:#141414" onclick="addSt('${c}')">＋ стикер</button>`).join('')}<button class="btn blk" onclick="$('#bd').hidden=true">Закрыть</button></div>`+S.st.map(s=>`<div class="st" style="left:${s.x}px;top:${s.y}px;background:${s.c}"><div class="sth" onpointerdown="drag(event,${s.id})">⠿<button onclick="delSt(${s.id})" aria-label="Удалить стикер">×</button></div><div class="stt" contenteditable oninput="S.st.find(z=>z.id==${s.id}).t=this.textContent">${esc(s.t)}</div></div>`).join('')}
function addSt(c){S.st.push({id:Date.now(),x:80+Math.random()*200,y:140+Math.random()*160,c,t:'Новая идея'});drawBoard()}
function delSt(id){S.st=S.st.filter(s=>s.id!==id);drawBoard()}
function drag(e,id){const s=S.st.find(z=>z.id==id),el=e.currentTarget.parentNode,sx=e.clientX-s.x,sy=e.clientY-s.y;
const mv=ev=>{s.x=ev.clientX-sx;s.y=Math.max(60,ev.clientY-sy);el.style.left=s.x+'px';el.style.top=s.y+'px'};
const up=()=>{removeEventListener('pointermove',mv);removeEventListener('pointerup',up)};addEventListener('pointermove',mv);addEventListener('pointerup',up);e.preventDefault()}
