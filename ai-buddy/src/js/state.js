// Общее состояние приложения (хранится в памяти)
const S={u:null,v:'home',mode:'in',chat:'ai',copen:0,nopen:null,q:'',tv:'list',cm:[2026,10],busy:0,leave:0,tasks:TASKS,notes:NOTES,st:STICKERS,chats:CHATS};
const done=()=>S.tasks.filter(t=>t.done).length;
const hot=()=>S.tasks.filter(t=>!t.done&&t.tag!=='Заявки'&&dd(t.d)<=3).sort((a,b)=>a.d<b.d?-1:1);
