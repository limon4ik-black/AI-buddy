function cal(mini){const[y,m]=S.cm,first=(new Date(y,m,1).getDay()+6)%7,days=new Date(y,m+1,0).getDate();
let h=['Пн','Вт','Ср','Чт','Пт','Сб','Вс'].map(d=>`<b class="wd">${d}</b>`).join('')+'<i></i>'.repeat(first);
for(let d=1;d<=days;d++){const k=`${y}-${pad(m+1)}-${pad(d)}`,ts=S.tasks.filter(t=>t.d===k);
h+=`<div class="cd ${k===TODAY?'td':''} ${ts.length?'has':''}"><span>${d}</span>${mini?(ts.some(t=>!t.done)?'<u></u>':''):ts.map(t=>`<em class="${t.done?'dn':''}">${esc(t.t)}</em>`).join('')}</div>`}
return `<div class="row sp" style="margin-bottom:10px"><button class="chip" onclick="cmv(-1)" aria-label="Назад">‹</button><b>${MN[m]} ${y}</b><button class="chip" onclick="cmv(1)" aria-label="Вперёд">›</button></div><div class="${mini?'':'big'}"><div class="cal ${mini?'mini':''}">${h}</div></div>`}
function cmv(n){let[y,m]=S.cm;m+=n;if(m<0){m=11;y--}if(m>11){m=0;y++}S.cm=[y,m];draw()}
