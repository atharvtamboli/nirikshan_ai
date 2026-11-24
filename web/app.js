'use strict';
const el = id => document.getElementById(id);
const state = { language:'en', projects:[], project:null, records:[], section:'funds', online:false, loading:false, error:false, generation:0 };
try { state.language = localStorage.getItem('nirikshan-language') === 'hi' ? 'hi' : 'en'; } catch {}
const t = (en,hi) => state.language === 'hi' ? hi : en;
const kinds = {
  award:['Award','अनुबंध'], bill:['Bill','बिल'], payment:['Payment','भुगतान'],
  inspection:['Inspection','निरीक्षण'], news:['News report','समाचार'], observation:['Observation','अवलोकन']
};
const sections = {funds:['Money trail','वित्तीय विवरण'],records:['Sources','स्रोत'],analysis:['Review signals','समीक्षा संकेत'],submit:['Contribute','जानकारी दें']};
function node(tag, text, cls) {
  const n=document.createElement(tag);
  if (text != null) n.textContent=text;
  if (cls) n.className=cls;
  return n;
}
function source(url,label) {
  try {
    const u=new URL(url);
    if (u.protocol!=='https:' && u.protocol!=='http:') throw new Error();
    const a=node('a',label+' ↗'); a.href=u.href; a.target='_blank'; a.rel='noopener noreferrer'; return a;
  } catch { return node('span',t('Source link unavailable','स्रोत लिंक उपलब्ध नहीं'),'muted'); }
}
function money(paise) {
  return new Intl.NumberFormat(state.language==='hi'?'hi-IN':'en-IN',{style:'currency',currency:'INR',minimumFractionDigits:paise%100?2:0}).format(paise/100);
}
function name(p){return state.language==='hi'&&p.hi?p.hi:p.name;}
function type(k){return kinds[k]?t(...kinds[k]):String(k);}
function staticText(){
  document.documentElement.lang=state.language;
  document.querySelectorAll('[data-en]').forEach(n=>n.textContent=t(n.dataset.en,n.dataset.hi));
  el('language').textContent=t('हिन्दी','English');
  el('language').lang=t('hi','en');
  el('query').placeholder=t(el('query').dataset.placeholderEn,el('query').dataset.placeholderHi);
  el('query').setAttribute('aria-label',el('query').placeholder);
}
function renderList() {
  const q=el('query').value.trim().toLocaleLowerCase();
  const list=state.projects.filter(p=>[p.name,p.hi,p.id,p.place,p.authority].join(' ').toLocaleLowerCase().includes(q));
  el('projects').replaceChildren();
  el('count').textContent=list.length+' '+t('projects','परियोजनाएँ');
  for(const p of list){
    const b=node('button',null,'project'); b.type='button';b.setAttribute('aria-current',String(state.project?.id===p.id));
    b.append(node('strong',name(p)),node('span',p.place));b.onclick=()=>choose(p,true);el('projects').append(b);
  }
  if (!list.length) el('projects').append(node('p',t('No matching projects. Try a place name.','कोई परिणाम नहीं। स्थान का नाम खोजें।'),'muted'));
}
function moneyTrail(box,p){
  const featured=node('div',null,'feature');
  featured.append(node('span',t('Advertised tender value','प्रकाशित निविदा राशि'),'pill'),node('p',Number.isSafeInteger(p.tenderPaise)?money(p.tenderPaise):'—','amount'),source(p.source,t('Read tender notice','निविदा सूचना खोलें')));
  box.append(featured);
  for(const k of ['award','bill','payment']){
    const entries=state.records.filter(r=>r.kind===k&&Number.isSafeInteger(r.amountPaise));
    const row=node('div',null,'moneyrow');
    row.append(node('strong',type(k)),node('span',entries.length?entries.map(r=>money(r.amountPaise)).join(' · '):state.loading?t('Loading…','लोड हो रहा है…'):state.error?t('Records unavailable','रिकॉर्ड उपलब्ध नहीं'):t('No published record','प्रकाशित रिकॉर्ड नहीं मिला'),'muted'));box.append(row);
  }
  box.append(node('p',t('Amounts refer to their individual records. Open Sources to check dates and references.','प्रत्येक राशि अलग दस्तावेज़ की है। तारीख और संदर्भ के लिए स्रोत खोलें।'),'small'));
}
function records(box,p){
  const items=[{title:p.name,date:p.date,url:p.source,kind:'tender',reference:p.authority},...state.records];
  for(const r of items){
    const card=node('section',null,'record');
    card.append(node('p',(r.kind==='tender'?t('Tender notice','निविदा सूचना'):type(r.kind))+' · '+(r.date||'—'),'pill'),node('strong',r.title));
    if(r.reference)card.append(node('p',r.reference,'muted'));
    if(Number.isSafeInteger(r.amountPaise))card.append(node('p',money(r.amountPaise)));
    card.append(source(r.url,t('Open original source','मूल स्रोत खोलें')));box.append(card);
  }
}
function signals(box,p){
  box.append(node('p',t('Review prompts compare the published records. Each signal needs an independent check.','प्रकाशित रिकॉर्ड की तुलना से ये संकेत बनते हैं। हर संकेत की स्वतंत्र जाँच ज़रूरी है।'),'muted'));
  if(state.error||state.loading)return;
  const findings=window.NirikshanAnalysis?.analyze(p,state.records,state.language)||[];
  if(!findings.length)box.append(node('p',t('No review signals in the available records.','उपलब्ध रिकॉर्ड में कोई समीक्षा संकेत नहीं मिला।'),'status'));
  for(const f of findings){
    const card=node('section',null,'record');card.append(node('strong',f.title),node('p',f.detail,'muted'));
    for(const id of f.ids||[]){
      const r=state.records.find(r=>r.id===id);
      if(r){const link=node('p');link.append(source(r.url,r.title));card.append(link);}
    }
    box.append(card);
  }
}
function contribute(box,p){
  box.append(node('p',t('Share a public document or a published report. A reviewer checks the source before it appears here. Leave out private contact details.','सार्वजनिक दस्तावेज़ या प्रकाशित रिपोर्ट साझा करें। प्रकाशन से पहले स्रोत की समीक्षा होगी। निजी संपर्क विवरण न दें।'),'muted'));
  if(!state.online){box.append(node('p',t('This edition is read-only. Submissions require the connected service.','यह संस्करण केवल पढ़ने के लिए है। जानकारी भेजने के लिए ऑनलाइन सेवा ज़रूरी है।'),'status'));return;}
  const form=node('form',null,'form');
  function field(key,label,tag='input',inputType='text',required=true){
    const wrap=node('label',label),input=node(tag);input.name=key;input.required=required;
    if(tag==='input')input.type=inputType;wrap.append(input);form.append(wrap);return input;
  }
  const kind=field('kind',t('Record type','रिकॉर्ड प्रकार'),'select');
  for(const k of Object.keys(kinds)){const opt=node('option',type(k));opt.value=k;kind.append(opt);}
  const title=field('title',t('Record title or observation','रिकॉर्ड शीर्षक या अवलोकन'));title.maxLength=180;
  const url=field('url',t('Public HTTPS source URL','सार्वजनिक HTTPS स्रोत URL'),'input','url');url.pattern='https://.+';
  field('reference',t('Document or page reference','दस्तावेज़ या पृष्ठ संदर्भ'));
  field('date',t('Record date','रिकॉर्ड की तारीख'),'input','date');
  const amt=field('amount',t('Amount in rupees (optional)','राशि रुपये में (वैकल्पिक)'),'input','number',false);amt.min='0';amt.step='.01';amt.max='90071992547409.90';
  function toggleFields(){url.required=kind.value!=='observation';amt.disabled=!['bill','payment','award'].includes(kind.value);if(amt.disabled)amt.value='';}
  kind.onchange=toggleFields;toggleFields();
  const submit=node('button',t('Send for review','समीक्षा के लिए भेजें'),'primary');
  const msg=node('p',null,'muted');msg.setAttribute('role','status');form.append(submit,msg);box.append(form);
  const receipt=node('details'),summary=node('summary',t('Track a contribution','अपने योगदान की स्थिति देखें'));
  const lookup=node('form',null,'form'),lbl=node('label',t('Private tracking reference','निजी ट्रैकिंग संदर्भ')),ref=node('input');
  ref.required=true;lbl.append(ref);const check=node('button',t('Check status','स्थिति देखें'),'action'),result=node('p',null,'muted');result.setAttribute('role','status');
  lookup.append(lbl,check,result);receipt.append(summary,lookup);box.append(receipt);
  lookup.onsubmit=async e=>{e.preventDefault();check.disabled=true;try{const status=(await getJSON('/api/receipts/'+encodeURIComponent(ref.value.trim()))).status;const labels={pending:t('Awaiting review','समीक्षा लंबित'),approve:t('Approved','स्वीकृत'),reject:t('Returned','लौटाया गया')};result.textContent=labels[status]||status;}catch{result.textContent=t('Receipt unavailable. Check the reference and try again.','रसीद उपलब्ध नहीं। संदर्भ जाँचकर फिर प्रयास करें।');}finally{check.disabled=false;}};
  form.onsubmit=async e=>{
    e.preventDefault();const data=Object.fromEntries(new FormData(form));data.project=p.id;
    if(data.amount != null && data.amount !== '')data.amountPaise=Math.round(Number(data.amount)*100);delete data.amount;
    submit.disabled=true;msg.textContent=t('Sending…','भेजा जा रहा है…');
    try{
      const res=await fetch('/api/evidence',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});
      if(!res.ok)throw Error(String(res.status));
      const result=await res.json();msg.textContent=t('Submitted for review.','समीक्षा के लिए भेजा गया।')+' '+(result.id||'');ref.value=result.id||'';form.reset();toggleFields();
    }catch{msg.textContent=t('Submission could not be confirmed. Your entries are still here.','भेजने की पुष्टि नहीं हुई। आपके दर्ज विवरण यहीं हैं।');}
    finally{submit.disabled=false;}
  };
}
function renderDetail(focus=false){
  const p=state.project;if(!p)return;
  const article=el('detail'),heading=node('h2',name(p));heading.id='projectName';heading.tabIndex=-1;
  article.replaceChildren(node('p',p.place,'eyebrow'),heading,node('p',p.authority||'','muted'));
  const rail=node('div',null,'rail');rail.setAttribute('role','group');rail.setAttribute('aria-label',t('Project views','परियोजना अनुभाग'));
  for(const [k,label] of Object.entries(sections)){
    const b=node('button',t(...label));b.type='button';b.id='view-'+k;b.setAttribute('aria-pressed',String(k===state.section));b.onclick=()=>{state.section=k;renderDetail();el('view-'+k).focus();};rail.append(b);
  }
  article.append(rail);
  const tools=node('div',null,'project-tools');
  const bill=node('a',t('Bill worksheet','बिल जाँच'));bill.href='./bill-check.html';
  const share=node('button',t('Share QR','QR साझा करें'),'action');share.type='button';share.setAttribute('data-share-qr','');
  tools.append(bill,share);article.append(tools);
  if(state.error)article.append(node('p',t('Published updates could not be loaded. Please refresh to retry.','प्रकाशित जानकारी लोड नहीं हुई। फिर से पेज खोलें।'),'status'));
  const box=node('div');article.append(box);
  ({funds:moneyTrail,records,analysis:signals,submit:contribute})[state.section](box,p);
  document.title=name(p)+' · Nirikshan';if(focus)heading.focus();
}
async function choose(p,writeURL=false){
  const generation=++state.generation;state.project=p;state.records=[];state.section='funds';state.error=false;state.loading=state.online;
  if(!writeURL && !new URL(location.href).searchParams.get('project')){const u=new URL(location.href);u.searchParams.set('project',p.id);history.replaceState(null,'',u);}
  if(writeURL){const url=new URL(location.href);url.searchParams.set('project',p.id);history.pushState(null,'',url);}
  renderList();renderDetail(writeURL);
  if(!state.online)return;
  try{
    const rows=await getJSON('/api/evidence?project='+encodeURIComponent(p.id));
    if(generation!==state.generation)return;state.records=rows||[];
  }catch{if(generation===state.generation)state.error=true;}
  finally{if(generation===state.generation){state.loading=false;if(state.section!=='submit')renderDetail();}}
}
async function getJSON(url){const r=await fetch(url,{headers:{Accept:'application/json'}});if(!r.ok)throw Error(String(r.status));return r.json();}
function openFromURL(){
  const id=new URL(location.href).searchParams.get('project'),p=id?state.projects.find(p=>p.id===id):state.projects[0];
  if(p)return choose(p);
  ++state.generation;state.project=null;renderList();el('detail').replaceChildren(node('p',t('Project reference not found. Select a project from the atlas.','परियोजना संदर्भ नहीं मिला। सूची से एक परियोजना चुनें।'),'status'));
}
async function init(){
  staticText();
  try{
    try{state.projects=await getJSON('/api/projects');if(!Array.isArray(state.projects))throw Error();state.online=true;}
    catch{state.projects=await getJSON('./projects.json');el('review').hidden=true;}
    if(!state.projects.length)throw Error();await openFromURL();
  }catch{el('detail').replaceChildren(node('p',t('Project records could not be loaded. Refresh to retry.','परियोजना रिकॉर्ड लोड नहीं हुए। फिर से पेज खोलें।'),'status'));}
}
el('language').onclick=()=>{state.language=t('hi','en');try{localStorage.setItem('nirikshan-language',state.language);}catch{}staticText();renderList();renderDetail();};
el('query').oninput=renderList;
window.addEventListener('popstate',()=>{if(state.projects.length)openFromURL();});
init();
