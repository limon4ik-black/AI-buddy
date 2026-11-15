function chips(){return ['Как оформить отгул?','Какие доступы мне нужны?','Что горит по срокам?','Одобри мой отгул'].map(q=>`<button class="chip" onclick="ask('${q}')">${q}</button>`).join('')}
