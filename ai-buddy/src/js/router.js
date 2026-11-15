function go(v,keep){if(v==='chat'&&!keep)S.copen=0;if(v==='notes'&&!keep)S.nopen=null;S.v=v;history.replaceState(null,'','#'+v);
document.querySelectorAll('[data-v]').forEach(b=>b.classList.toggle('on',b.dataset.v===v));
const m=$('#main');m.className=(v==='chat'||v==='notes')?'full':'';m.scrollTop=0;m.innerHTML=({home,chat,notes,list,profile})[v]();if(v==='chat')drawMsgs()}
const draw=()=>go(S.v,1);
