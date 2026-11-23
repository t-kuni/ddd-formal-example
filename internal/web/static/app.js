const $=s=>document.querySelector(s), esc=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])), yen=n=>'¥'+Number(n||0).toLocaleString('ja-JP');
const emptyOrders=()=>({orders:[],metrics:{total:0,active:0,delivered:0,exceptions:0,points:0}});
let boot, data={inventory:[],commands:[],orderPage:emptyOrders(),now:0}, actor, page='catalog',
    search='', statusFilter='', detailID='', detail, waiting=new Map(), draft={};
const pendingAdds=[];
let submitting=false;
let cursors=[null], revision=0, lastState='', lastRefresh=0, refreshing=false, polling=false;
const labels={customer:'顧客',warehouse:'倉庫作業者',carrier:'運送業者',operator:'オペレーター'};
const pages={catalog:'商品カタログ',cart:'カート・発注',orders:'受注・配送追跡',warehouse:'倉庫作業',delivery:'配送管理',inventory:'在庫管理',commands:'処理履歴'};
function notice(s,error=false){$('#notice').textContent=s;$('#notice').className=error?'error':''}
async function api(path,options={}){const r=await fetch(path,{...options,headers:{'Content-Type':'application/json','X-Actor-ID':actor?.id||'',...options.headers}});if(!r.ok){const t=await r.text();let m=t;try{m=JSON.parse(t).error||t}catch{}throw Error(m)}return r.headers.get('content-type')?.includes('json')?r.json():r.text()}
function stock(id){return data.inventory.find(x=>x.id===id)}
function cart(){return data.cart?.body||{items:{},expires:0}}
function product(id){return boot.products.find(x=>x.id===id)}function who(id){return boot.actors.find(x=>x.id===id)?.name||'システム'}
function badge(s){return '<span class="badge s'+s+'">'+esc(boot.statuses[s])+'</span>'}function date(s){return new Date(s).toLocaleString('ja-JP',{month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit'})}function canRender(){return !['INPUT','TEXTAREA','SELECT'].includes(document.activeElement?.tagName)}
function nav(){let items=actor.role==='customer'?['catalog','cart','orders','commands']:actor.role==='warehouse'?['warehouse','orders','commands']:actor.role==='carrier'?['delivery','orders','commands']:['orders','inventory','commands'];if(!items.includes(page))page=items[0];$('#nav').innerHTML=items.map(p=>'<button data-page="'+p+'" class="'+(page===p?'active':'')+'">'+pages[p]+'</button>').join('');$('#title').textContent=pages[page]}
async function refresh(force=false){
 const current=revision;
 if(refreshing&&!force)return;
 refreshing=true;
 try{
  const query=new URLSearchParams();
  let path='/api/catalog', key='catalog';
  if(['orders','warehouse','delivery'].includes(page)){
   key='orderPage';
   if(page!=='orders')query.set('queue',page);
   if(statusFilter!=='')query.set('status',statusFilter);
   if(cursors.at(-1))query.set('before',cursors.at(-1));
   path='/api/orders?'+query;
  }else if(page==='commands'){key='commands';path='/api/commands'}
  const [next,nextDetail]=await Promise.all([api(path),detailID?api('/api/orders/'+detailID):null]);
  if(current!==revision)return;
  if(key==='catalog')Object.assign(data,next,{cart:next.cart});
  else data[key]=next;
  detail=nextDetail;
  lastRefresh=Date.now();
  const signature=JSON.stringify([page,key==='catalog'?[data.inventory,data.cart,cart().expires<=data.now]:next,detail]);
  if(force||(signature!==lastState&&canRender())){
   lastState=signature;
   render();
   if(detailID&&$('#detail').open)renderDetail();
  }
 }catch(e){if(current===revision)notice(e.message,true)}
 finally{if(current===revision)refreshing=false}
}
async function poll(){
 if(polling)return;
 if(!waiting.size){
  if(Date.now()-lastRefresh>=5000)await refresh();
  return;
 }
 polling=true;
 const owner=actor.id;
 try{
  for(const [id,kind] of waiting){
   const [c]=await api('/api/commands?id='+id);
   if(owner!==actor.id)return;
   if(!c||c.status==='queued')continue;
   if(c.status==='succeeded'){
    if(kind==='order.checkout'&&c.result?.orderId){
     page='orders';statusFilter='';cursors=[null];
     detailID=c.result.orderId;detail=null;
     renderDetail();if(!$('#detail').open)$('#detail').showModal();
    }
    notice('操作が完了しました。');
   }else notice(c.error||'処理に失敗しました',true);
   revision++;
   await refresh(true);
   waiting.delete(id);
  }
 }catch(e){if(owner===actor.id)notice(e.message,true)}
 finally{polling=false;void runNextAdd()}
}
async function runNextAdd(){
 if(submitting||waiting.size||!pendingAdds.length)return;
 submitting=true;
 const id=pendingAdds.shift();
 try{
  const {cart:current,now}=await api('/api/catalog');
  const quantity=current?.body.expires>now?(current.body.items[id]||0):0;
  await submitCommand('cart.set',{productId:id,quantity:quantity+1,expectedVersion:current?.version||0});
 }catch(e){notice(product(id).name+'の追加に失敗しました。'+e.message,true)}
 finally{submitting=false}
 if(waiting.size)await poll();
 else void runNextAdd();
}
async function submitCommand(type,payload){
 const id=crypto.randomUUID();
 await api('/api/commands',{method:'POST',headers:{'Idempotency-Key':id},body:JSON.stringify({type,payload})});
 waiting.set(id,type);
 notice('受付済みです。処理結果を確認しています…');
}
async function command(type,payload={}){
 if(submitting||waiting.size||pendingAdds.length){
  notice('前の操作の完了を待っています…');return;
 }
 submitting=true;
 try{await submitCommand(type,payload)}
 catch(e){notice(e.message,true)}
 finally{submitting=false}
 await poll();
}
function render(){nav();const content=$('#content');if(page==='catalog')content.innerHTML=catalog();else if(page==='cart')content.innerHTML=cartPage();else if(page==='inventory')content.innerHTML=inventory();else if(page==='commands')content.innerHTML=commands();else content.innerHTML=orders();}
function catalog(){const list=boot.products.filter(p=>(p.name+p.category+p.sku).toLowerCase().includes(search.toLowerCase()));return `<section class="hero"><div><span class="eyebrow">THE EVERYDAY ESSENTIALS</span><h2>日々のビジネスに、確かな供給を。</h2><p>商品をカートに追加すると在庫を確保します。<br>カートを変更してから1時間、まとめてお取り置き。</p></div><div class="hero-mark">01—06</div></section><div class="toolbar"><input id="search" aria-label="商品検索" placeholder="商品名・SKU・カテゴリーで検索" value="${esc(search)}"><span class="count">${list.length} PRODUCTS</span><button class="secondary" data-page="cart">カートを見る (${Object.values(cart().items).reduce((a,b)=>a+b,0)})</button></div><section class="grid">${list.map(p=>{const st=stock(p.id)?.body||{};return `<article class="product"><img src="${p.image}" alt="${esc(p.name)}"><div class="product-body"><div class="category">${esc(p.category)}</div><h3>${esc(p.name)}</h3><div class="sku">${p.sku} · ${p.unit}</div><div class="price">${yen(p.price)} <small>/ ${p.unit}・税抜</small></div><div class="product-actions"><span class="stock">購入可能 ${st.onHand-st.reserved} ${p.unit}</span><button data-add="${p.id}">＋ カートに追加</button></div></div></article>`}).join('')}</section>`}
function cartPage(){const c=cart(),ids=Object.keys(c.items),expired=c.expires<=data.now,addr=draft.address||actor.address;return `<section class="panel"><div class="panel-heading"><h2>お取り置き中の商品</h2><span class="badge">${ids.length?expired?'有効期限切れ':'有効期限 '+date(c.expires*1000):'カートは空です'}</span></div>${ids.length?`<div class="table-wrap"><table><thead><tr><th>商品</th><th>単価（税抜）</th><th>数量</th><th>小計</th><th></th></tr></thead><tbody>${ids.map(id=>{const p=product(id);return `<tr><td>${esc(p.name)}<br><small class="muted">${p.sku}</small></td><td>${yen(p.price)}</td><td><input aria-label="${esc(p.name)}の数量" id="qty-${id}" type="number" min="0" max="10000" value="${c.items[id]}"> <button class="tiny secondary" data-set="${id}">更新</button></td><td>${yen(p.price*c.items[id])}</td><td><button class="tiny danger" data-remove="${id}">削除</button></td></tr>`}).join('')}</tbody></table></div><div class="summary"><span class="muted">価格・割引・税は発注時に確定します。</span><span class="total">${yen(ids.reduce((s,id)=>s+product(id).price*c.items[id],0))} <small>税抜</small></span></div>`:'<div class="empty">まだ商品がありません。<br><button class="secondary" data-page="catalog">商品を探す</button></div>'}</section>${ids.length?`<section class="panel"><h2>配送先・発注内容</h2><form id="checkout"><div class="form-grid">${[['company','会社名'],['recipient','担当者名'],['postal','郵便番号'],['phone','電話番号'],['text','住所']].map(([k,l])=>`<label class="${k==='text'?'wide':''}">${l}<input name="${k}" required maxlength="250" value="${esc(addr[k])}"></label>`).join('')}<label>希望納品日<input type="date" name="requestedDate" value="${esc(draft.requestedDate||'')}"></label><label>クーポン<input name="coupon" placeholder="WELCOME10 / PARTNER5" value="${esc(draft.coupon||'')}"></label><label class="wide">配送に関する連絡事項<textarea name="note" maxlength="1000" placeholder="納品口、受取可能な時間など">${esc(draft.note||'')}</textarea></label></div><div class="actions"><button ${expired?'disabled':''}>発注する</button><span class="footnote">発注後、擬似決済の成功・失敗を選択できます。</span></div></form></section>`:''}`}
function orders(){const {orders:list,metrics,next}=data.orderPage;return `<section class="metrics"><div class="metric"><span>受注件数</span><b>${metrics.total}<small class="muted"> 件</small></b></div><div class="metric"><span>対応中</span><b>${metrics.active}</b></div><div class="metric"><span>配送完了</span><b>${metrics.delivered}</b></div><div class="metric"><span>${actor.role==='customer'?'保有ポイント':'要確認'}</span><b>${actor.role==='customer'?metrics.points:metrics.exceptions}</b></div></section><section class="panel"><div class="panel-heading"><h2>${page==='warehouse'?'作業キュー':page==='delivery'?'配送キュー':'受注一覧'}</h2><select id="status-filter" aria-label="ステータス絞り込み"><option value="">すべてのステータス</option>${boot.statuses.map((s,i)=>`<option value="${i}" ${statusFilter===String(i)?'selected':''}>${s}</option>`).join('')}</select></div>${list.length?`<div class="table-wrap"><table><thead><tr><th>受注番号 / 発注日時</th><th>届け先</th><th>状態</th><th>請求額</th><th>希望納品日</th><th></th></tr></thead><tbody>${list.map(o=>{return `<tr><td><strong>${esc(o.number)}</strong><br><small class="muted">${date(o.createdAt)}</small></td><td>${esc(o.company)}<br><small class="muted">${esc(o.recipient)}</small></td><td>${badge(o.status)}</td><td>${yen(o.total)}</td><td>${esc(o.requestedDate)||'指定なし'}</td><td><button class="tiny secondary" data-detail="${o.id}">詳細・操作</button></td></tr>`}).join('')}</tbody></table></div>`:'<div class="empty">対象の受注はありません。</div>'}<div class="actions">${cursors.length>1?'<button class="secondary" data-order-page="previous">前へ</button>':''}${next?'<button class="secondary" data-order-page="next">次へ</button>':''}</div></section>`}
function inventory(){return `<section class="hero"><div><h2>引当を守り、在庫を整える。</h2><p>手持ち数量が引当済み数量を下回る調整は拒否されます。</p></div></section><section class="panel"><table><thead><tr><th>商品</th><th>手持ち</th><th>引当</th><th>販売可能</th><th>増減数</th><th>調整理由</th><th></th></tr></thead><tbody>${boot.products.map(p=>{const v=stock(p.id),s=v?.body||{};return `<tr><td>${esc(p.name)}<br><span class="muted">${p.sku}</span></td><td>${s.onHand}</td><td>${s.reserved}</td><td>${s.onHand-s.reserved}</td><td><input type="number" id="delta-${p.id}" value="0" aria-label="${p.sku}の増減数"></td><td><input id="reason-${p.id}" aria-label="${p.sku}の調整理由" placeholder="棚卸差異・破損・返品検品"></td><td><button class="tiny" data-adjust="${p.id}">調整</button></td></tr>`}).join('')}</tbody></table></section>`}
function commands(){return `<section class="panel"><div class="panel-heading"><h2>直近50件の処理履歴</h2>${actor.role==='operator'?'<button class="secondary tiny" id="maintenance">期限・再送・メール処理を実行</button>':''}</div><p class="muted">受付と処理完了は別です。拒否された操作では業務データは変更されません。</p><div class="table-wrap"><table><thead><tr><th>受付日時 / ID</th><th>操作</th><th>状態</th><th>結果</th></tr></thead><tbody>${data.commands.map(c=>`<tr><td>${date(c.createdAt)}<br><small class="muted">${c.id.slice(0,8)}</small></td><td>${esc(c.type)}</td><td><span class="badge ${c.status}">${({queued:'受付済み',succeeded:'完了',rejected:'拒否',dead:'要再投入'})[c.status]||c.status}</span></td><td>${esc(c.error)}${c.status==='dead'&&actor.role==='operator'?`<button class="tiny" data-retry="${c.id}">再投入</button>`:''}</td></tr>`).join('')}</tbody></table></div></section>`}
function orderActions(o){let a=[];const r=actor.role,s=o.status;if(r==='customer'){if(s===0)a.push(['order.pay:true','決済成功'],['order.pay:false','決済失敗']);if([0,1].includes(s))a.push(['order.cancel','キャンセル'])}if(r==='operator'&&[0,1,2,3].includes(s))a.push(['order.cancel','キャンセル']);if(r==='warehouse'){if(s===1||(s===3&&o.workerId===actor.id))a.push(['order.pick',s===1?'作業開始':'作業再開']);if(s===2&&o.workerId===actor.id)a.push(['order.block','欠品・破損で保留'],['order.pack','梱包完了'])}if(r==='carrier'){if(s===4)a.push(['order.dispatch','運送開始']);if(o.carrierId===actor.id){if(s===5)a.push(['order.deliver','配送完了'],['order.delivery_failed','配送失敗']);if(s===6)a.push(['order.redeliver','再配送開始'],['order.return','倉庫へ返送'])}}return a}
function renderDetail(){if(!detail){$('#detail-content').innerHTML='<div class="empty">受注情報を取得しています…</div>';return}const o=detail.order,actions=orderActions(o),invoice=detail.invoice;$('#detail-content').innerHTML=`<span class="eyebrow">ORDER DETAIL</span><h2>${esc(o.number)} ${badge(o.status)}</h2><p class="muted">${o.id}</p><div class="detail-meta"><div><small>お届け先</small>${esc(o.address.company)} / ${esc(o.address.recipient)}<br>${esc(o.address.postal)} ${esc(o.address.text)}<br>${esc(o.address.phone)}</div><div><small>配送・担当</small>希望日: ${esc(o.requestedDate)||'指定なし'}<br>倉庫: ${o.workerId?esc(who(o.workerId)):'未割当'}<br>運送: ${o.carrierId?esc(who(o.carrierId)):'未割当'}<br>追跡番号: ${esc(o.tracking)||'未発行'}</div></div><p>${esc(o.note)}</p><table><thead><tr><th>商品</th><th>数量</th><th>税抜金額</th></tr></thead><tbody>${o.lines.map(l=>`<tr><td>${esc(l.name)}</td><td>${l.quantity} ${esc(l.unit)}</td><td>${yen(l.quantity*l.price)}</td></tr>`).join('')}</tbody></table><div class="summary"><span class="muted">割引 ${yen(o.discount)} / 消費税10% ${yen(o.tax)}<br>配送完了で ${o.points}ポイント付与</span><span class="total">${yen(o.total)} <small>税込</small></span></div>${o.boxes.length?'<h3>配送する箱</h3>'+o.boxes.map(b=>`<div class="box-tag">${esc(b.label)} · ${b.weightGrams}g<br><span class="muted">${b.id}</span></div>`).join(''):''}${actions.length?`<section class="panel"><h3>次の操作</h3><div class="form-grid"><label class="wide">作業メモ・異常理由<input id="action-note" maxlength="1000" placeholder="保留・配送失敗・返送では理由を入力"></label>${actions.some(a=>a[0]==='order.pack')?'<label>箱数<input id="box-count" type="number" min="1" max="100" value="1"></label><label>1箱の重量（g）<input id="box-weight" type="number" min="1" value="1000"></label>':''}${actions.some(a=>a[0]==='order.dispatch')?'<label class="wide">追跡番号<input id="tracking" placeholder="TT-2026-0001"></label>':''}</div><div class="actions">${actions.map(([cmd,label])=>`<button data-order-action="${cmd}" class="${/cancel|false|block|failed|return/.test(cmd)?'secondary':''}">${label}</button>`).join('')}</div></section>`:''}${invoice?`<div class="panel"><h3>請求書</h3><a class="button secondary" target="_blank" href="/api/invoices/${invoice.id}?actor=${actor.id}">請求書PDFを開く</a><p class="muted">メール: ${invoice.sent?'送信済み':'送信待ち'} ${esc(invoice.error)}</p></div>`:''}<h3>状態の履歴</h3><ol class="timeline">${[...o.history].reverse().map(h=>`<li>${badge(h.to)} ${esc(h.note)}<small>${date(h.at)} · ${esc(who(h.actor))}</small></li>`).join('')}</ol>`}
document.addEventListener('click',async e=>{const b=e.target.closest('button');if(!b)return;try{if(b.dataset.page){page=b.dataset.page;statusFilter='';cursors=[null];data.orderPage=emptyOrders();revision++;await refresh(true)}else if(b.dataset.add){pendingAdds.push(b.dataset.add);void runNextAdd()}else if(b.dataset.set){const id=b.dataset.set;await command('cart.set',{productId:id,quantity:Number($('#qty-'+id).value),expectedVersion:data.cart?.version||0})}else if(b.dataset.remove){await command('cart.set',{productId:b.dataset.remove,quantity:0,expectedVersion:data.cart?.version||0})}else if(b.dataset.detail){detailID=b.dataset.detail;detail=null;renderDetail();$('#detail').showModal();revision++;await refresh(true)}else if(b.dataset.orderPage){if(b.dataset.orderPage==='next')cursors.push(data.orderPage.next);else cursors.pop();revision++;await refresh(true)}else if(b.dataset.adjust){const id=b.dataset.adjust;await command('stock.adjust',{productId:id,delta:Number($('#delta-'+id).value),note:$('#reason-'+id).value,expectedVersion:stock(id).version})}else if(b.dataset.orderAction){if(detail?.order.id!==detailID)return;const raw=b.dataset.orderAction;const type=raw.split(':')[0];const payload={orderId:detailID,expectedVersion:detail.version,note:$('#action-note').value};if(type==='order.pay')payload.success=raw.endsWith(':true');if(type==='order.pack'){payload.boxCount=Number($('#box-count').value);payload.weightGrams=Number($('#box-weight').value)}if(type==='order.dispatch')payload.tracking=$('#tracking').value;await command(type,payload)}else if(b.dataset.retry){await api('/api/commands/'+b.dataset.retry+'/retry',{method:'POST'});notice('再投入しました');revision++;await refresh(true)}else if(b.id==='maintenance'){await api('/api/maintenance',{method:'POST'});notice('定期処理を受け付けました')}else if(b.id==='close-detail')$('#detail').close()}catch(e){notice(e.message,true)}});
document.addEventListener('input',e=>{if(e.target.id==='search'){const pos=e.target.selectionStart;search=e.target.value;render();$('#search').focus();$('#search').setSelectionRange(pos,pos)}if(e.target.closest('#checkout')){const f=new FormData($('#checkout'));draft={address:Object.fromEntries(['company','recipient','postal','phone','text'].map(k=>[k,f.get(k)])),coupon:f.get('coupon'),note:f.get('note'),requestedDate:f.get('requestedDate')}}});
document.addEventListener('change',async e=>{if(e.target.id==='status-filter'){statusFilter=e.target.value;cursors=[null];revision++;await refresh(true)}});
$('#detail').addEventListener('close',()=>{if(!$('#detail').open&&detailID){detailID='';detail=null;$('#detail-content').replaceChildren();revision++;refreshing=false}});
document.addEventListener('submit',async e=>{if(e.target.id!=='checkout')return;e.preventDefault();const f=new FormData(e.target);await command('order.checkout',{address:Object.fromEntries(['company','recipient','postal','phone','text'].map(k=>[k,f.get(k)])),coupon:f.get('coupon'),note:f.get('note'),requestedDate:f.get('requestedDate'),expectedVersion:data.cart?.version||0})});
$('#actor').addEventListener('change',async e=>{if(submitting||waiting.size||pendingAdds.length){e.target.value=actor.id;notice('処理が完了してから利用者を切り替えてください。');return}actor=boot.actors.find(x=>x.id===e.target.value);localStorage.setItem('orbit-actor',actor.id);draft={};waiting.clear();detailID='';detail=null;$('#detail-content').replaceChildren();$('#detail').close();$('#content').replaceChildren();notice('');statusFilter='';cursors=[null];data={inventory:[],commands:[],orderPage:emptyOrders(),now:0};page=actor.role==='customer'?'catalog':actor.role==='warehouse'?'warehouse':actor.role==='carrier'?'delivery':'orders';revision++;await refresh(true)});
try{boot=await api('/api/bootstrap');actor=boot.actors.find(x=>x.id===localStorage.getItem('orbit-actor'))||boot.actors.find(x=>x.role==='customer');$('#actor').innerHTML=boot.actors.map(a=>`<option value="${a.id}">${labels[a.role]} · ${esc(a.name)}</option>`).join('');$('#actor').value=actor.id;revision++;await refresh(true);setInterval(()=>poll(),1500)}catch(e){notice(e.message,true)}
