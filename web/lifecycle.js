const $ = id => document.getElementById(id);
const stages = ["sanction","tender","award","work_order","construction","inspection","completion","maintenance"];
const roles = {government:["sanction","tender","award","work_order","payment","completion","maintenance","response"],contractor:["bid","construction","delivery","progress","bill"],engineer:["certificate","inspection"],citizen:["report"]};
const labels = {sanction:"Government sanction",tender:"Tender published",award:"Contract awarded",work_order:"Work order",construction:"Construction",inspection:"Inspection",completion:"Completion",maintenance:"Maintenance",bid:"Contractor quotation",delivery:"Material delivery",progress:"Progress update",bill:"Contractor bill",certificate:"Certified measurement",payment:"Payment released",report:"Citizen report",response:"Department response"};
roles.government.push("acknowledge","resolve");
roles.government.push("work_item","variation");
roles.citizen.push("reopen","news");
Object.assign(labels,{acknowledge:"Department acknowledgement",resolve:"Repair or resolution",reopen:"Report reopened",work_item:"Contract work item",variation:"Approved quantity increase",news:"News source for review"});
let events = [], pending = null, locationEvidence = {}, locationRequest = 0, selectedStage = "";
const money = n => new Intl.NumberFormat("en-IN",{style:"currency",currency:"INR",maximumFractionDigits:2}).format(n/100);
function node(tag,text,cls){const el=document.createElement(tag);el.textContent=text;if(cls)el.className=cls;return el;}
function option(value,text){const el=node("option",text);el.value=value;return el;}
function records(){return events.filter(e=>e.project===$("project").value);}
function record(e){const item=node("div","","record");item.append(node("strong",labels[e.kind]+" · "+e.title),node("p",e.detail),node("small",new Date(e.time).toLocaleString()+" · "+e.id));if(e.amountPaise)item.append(node("strong",money(e.amountPaise)));if(e.source){try{const url=new URL(e.source);if(url.protocol==="https:"){const link=node("a","Open supporting source");link.href=url.href;link.target="_blank";link.rel="noopener noreferrer";item.append(link);}}catch{}}if(e.photoHash){const img=node("img","");img.src="./api/demo/photo/"+e.photoHash;img.alt="Submitted photo for "+e.title;img.loading="lazy";img.className="report-photo";item.append(img,node("small","Image digest: "+e.photoHash));if(e.duplicate)item.append(node("p","Matching image already exists in the demo history","flag"));}if(e.location)item.append(node("small","Approximate device location: "+e.latitude.toFixed(3)+", "+e.longitude.toFixed(3)+" · reported accuracy "+Math.round(e.accuracy)+" m · unverified"));return item;}
function configure(){
  const kind=$("kind").value, ref=$("ref").value;
  $("projectIdLabel").hidden=kind!=="sanction";
  $("scopeFields").hidden=kind!=="sanction";
  $("progressField").hidden=kind!=="progress";
  $("amountLabel").hidden=!["sanction","tender","bid","award","bill","certificate","payment","work_item","variation"].includes(kind);
  $("refLabel").hidden=!["award","bill","variation","certificate","payment","response","acknowledge","resolve","reopen"].includes(kind);
  $("evidenceFields").hidden=!["report","resolve"].includes(kind);
  $("quantityFields").hidden=!["bill","certificate","delivery","work_item","variation"].includes(kind);
  const target=kind==="award"?"bid":["bill","variation"].includes(kind)?"work_item":["certificate","payment"].includes(kind)?"bill":"report";
  const options=records().filter(e=>e.kind===target).map(e=>option(e.id,e.title+" · "+e.id));
  if(kind==="bill")options.unshift(option("","Unlinked bill — review required"));
  $("ref").replaceChildren(...options);
  if([...$("ref").options].some(o=>o.value===ref))$("ref").value=ref;
}
function render(){
  const chosen=$("project").value || "demo-bridge";
  $("project").replaceChildren(...events.filter(e=>e.kind==="sanction").map(e=>option(e.project,e.title)));
  if([...$("project").options].some(o=>o.value===chosen))$("project").value=chosen;
  const rows=records(), first=rows[0];
  if(!first)return;
  $("name").textContent=first.title;
  $("route").textContent=first.detail;
  $("routeMap").replaceChildren();
  if(first.startPoint&&first.endPoint){
    const route=node("div","","route-track");route.append(node("strong",first.startPoint),node("span",""),node("strong",first.endPoint));$("routeMap").append(route,node("small","Illustrative route schematic · endpoints are fictional"));
  }else $("routeMap").append(node("p","Route endpoints awaiting record."));
  $("schedule").textContent=[first.state,first.district,first.category,first.startDate&&"Start "+first.startDate,first.endDate&&"Target "+first.endDate].filter(Boolean).join(" · ")||"Schedule awaiting record.";
  const latestProgress=rows.filter(e=>e.kind==="progress").at(-1);
  $("progress").textContent=latestProgress?"Contractor-reported physical completion: "+(latestProgress.progress||0)+"%":"Physical progress percentage awaiting record.";
  const current=stages.findLastIndex(s=>rows.some(e=>e.kind===s));
  $("current").textContent="Latest documented stage · "+labels[stages[current]];
  if(!selectedStage)selectedStage=stages[current];
  $("stages").replaceChildren(...stages.map((s,i)=>{const event=rows.find(e=>e.kind===s),li=node("li","",i===current?"active":i<current?"done":"");const button=node("button",labels[s]);button.type="button";button.setAttribute("aria-pressed",String(s===selectedStage));button.addEventListener("click",()=>{selectedStage=s;render();$("stageDetail").focus();});li.append(button,node("small",event?new Date(event.time).toLocaleDateString(): "Awaiting record"));return li;}));
  $("stageDetail").tabIndex=-1;
  const stageRecord=rows.find(e=>e.kind===selectedStage);$("stageDetail").replaceChildren(stageRecord?record(stageRecord):node("p",labels[selectedStage]+" · awaiting record"));
  const total=kind=>rows.filter(e=>e.kind===kind).reduce((sum,e)=>sum+e.amountPaise,0);
  $("metrics").replaceChildren(...[["Sanction",total("sanction")],["Award",total("award")],["Certified",total("certificate")],["Paid",total("payment")]].map(([label,num])=>{const card=node("div","","metric");card.append(node("span",label),node("strong",num?money(num):"Awaiting record"));return card;}));
  const signals=[];
  if(total("bill")>total("certificate"))signals.push(money(total("bill")-total("certificate"))+" billed and awaiting certification.");
  const approved=total("award")+total("variation");
  if(total("certificate")>approved&&approved)signals.push("Certified amounts exceed the recorded contract award plus approved increases.");
  if(approved>total("sanction")&&total("sanction"))signals.push("Contract and approved increases exceed the recorded sanction.");
  for(const item of rows.filter(e=>e.kind==="work_item")){
    const allowed=item.quantityMilli+rows.filter(e=>e.kind==="variation"&&e.ref===item.id).reduce((sum,e)=>sum+e.quantityMilli,0);
    const billed=rows.filter(e=>e.kind==="bill"&&e.ref===item.id).reduce((sum,e)=>sum+e.quantityMilli,0);
    if(billed>allowed)signals.push(item.title+": billed quantity exceeds approved quantity by "+(billed-allowed)/1000+" "+item.unit+".");
  }
  if(latestProgress&&total("award")&&total("payment")*100/total("award")>(latestProgress.progress||0)+20)signals.push("Payment share is more than 20 percentage points above reported work progress. Review milestone terms.");
  $("riskSummary").replaceChildren(...(signals.length?signals:["No aggregate discrepancy identified in the recorded demo data."]).map(s=>node("p",s)));
  const financial=rows.filter(e=>["tender","bid","award","work_item","variation","delivery","bill","certificate","payment"].includes(e.kind));
  $("money").replaceChildren(...financial.map(e=>{
    const card=record(e);
    if(e.ref){const linked=rows.find(x=>x.id===e.ref);if(linked)card.append(node("small","Related to: "+linked.title));}
    if(e.quantityMilli)card.append(node("p",e.quantityMilli/1000+" "+e.unit+(e.ratePaise?" × "+money(e.ratePaise):"")));
    if(e.invoice)card.append(node("small","Reference: "+e.invoice));
    if(e.kind==="bill"){
      const certificates=rows.filter(x=>x.kind==="certificate"&&x.ref===e.id),certified=certificates.reduce((sum,x)=>sum+x.amountPaise,0);
      if(certified<e.amountPaise)card.append(node("p",money(e.amountPaise-certified)+" awaiting certification","flag"));
      if(e.quantityMilli&&(BigInt(e.quantityMilli)*BigInt(e.ratePaise)+500n)/1000n!==BigInt(e.amountPaise))card.append(node("p","Printed amount differs from quantity × rate","flag"));
      if(e.invoice&&rows.filter(x=>x.kind==="bill"&&x.invoice?.trim().toLowerCase()===e.invoice.trim().toLowerCase()).length>1)card.append(node("p","Repeated invoice reference — review for duplication","flag"));
      if(e.quantityMilli)card.append(node("small","Certified quantity: "+certificates.reduce((sum,x)=>sum+(x.quantityMilli||0),0)/1000+" "+e.unit));
    }
    return card;
  }));
  $("history").replaceChildren(...rows.slice().reverse().map(record));
  $("reports").replaceChildren(...rows.filter(e=>e.kind==="report").map(e=>{const card=record(e),responses=rows.filter(x=>["response","acknowledge","resolve","reopen"].includes(x.kind)&&x.ref===e.id);const last=responses.at(-1)?.kind;card.append(node("strong",last==="resolve"?"Resolved":last==="reopen"?"Reopened":last?"Acknowledged":"Submitted · awaiting department"),...responses.map(record));return card;}));
  if(!$("reports").children.length)$("reports").append(node("p","No reports recorded."));
  configure();
  discover();
}
function discover(){
  const query=$("projectSearch").value.trim().toLowerCase();
  const matches=events.filter(e=>e.kind==="sanction"&&[e.title,e.state,e.district,e.category,e.detail].join(" ").toLowerCase().includes(query));
  $("discover").replaceChildren(...matches.map(e=>{const button=node("button",e.title+(e.district?" · "+e.district:""));button.type="button";button.addEventListener("click",()=>{$("project").value=e.project;render();$("name").scrollIntoView({behavior:"smooth",block:"center"});});return button;}));
  if(!matches.length)$("discover").append(node("p","No demo projects match."));
}
$("projectSearch").addEventListener("input",discover);
async function refresh(){
  try{const r=await fetch("./api/demo",{cache:"no-store"}), data=await r.json();if(!r.ok)throw Error(data.error);if(JSON.stringify(events)!==JSON.stringify(data)){events=data;render();}$("status").textContent="Live local demo · fictional data";}catch(e){$("status").textContent=e.message;}
}
$("project").addEventListener("change",()=>{selectedStage="";render();});
$("role").addEventListener("change",()=>{$("kind").replaceChildren(...roles[$("role").value].map(k=>option(k,labels[k])));configure();});
$("kind").addEventListener("change",configure);
$("role").dispatchEvent(new Event("change"));
$("locate").addEventListener("click",()=>{
  if(!navigator.geolocation){$("locStatus").textContent="Location unavailable.";return;}
  $("locStatus").textContent="Requesting location…";
  const request=++locationRequest;
  navigator.geolocation.getCurrentPosition(p=>{if(request!==locationRequest)return;locationEvidence={location:true,latitude:p.coords.latitude,longitude:p.coords.longitude,accuracy:p.coords.accuracy};$("locStatus").textContent="Approximate location ready. Accuracy "+Math.round(p.coords.accuracy)+" m.";},()=>{if(request!==locationRequest)return;$("locStatus").textContent="Location unavailable. You can still submit.";},{enableHighAccuracy:true,timeout:12000});
});
$("clearLocation").addEventListener("click",()=>{locationRequest++;locationEvidence={};$("locStatus").textContent="Location not shared.";});
const sha256=async bytes=>Array.from(new Uint8Array(await crypto.subtle.digest("SHA-256",bytes)),b=>b.toString(16).padStart(2,"0")).join("");
const fromBase64=s=>Uint8Array.from(atob(s),c=>c.charCodeAt(0));
function toBase64(bytes){let binary="";for(let i=0;i<bytes.length;i+=8192)binary+=String.fromCharCode(...bytes.subarray(i,i+8192));return btoa(binary);}
$("export").addEventListener("click",async()=>{
  $("export").disabled=true;
  try{
    const rows=records(),photos={},payload=JSON.stringify({fictional:true,exportedAt:new Date().toISOString(),events:rows});
    for(const hash of new Set(rows.map(e=>e.photoHash).filter(Boolean))){
      const response=await fetch("./api/demo/photo/"+hash);if(!response.ok)throw Error("Photo unavailable: "+hash);
      const bytes=new Uint8Array(await response.arrayBuffer());if(await sha256(bytes)!==hash)throw Error("Photo digest mismatch: "+hash);
      photos[hash]=toBase64(bytes);
    }
    const bundle={format:"nirikshan-demo-v1",payload,digest:await sha256(new TextEncoder().encode(payload)),photos};
    const url=URL.createObjectURL(new Blob([JSON.stringify(bundle,null,2)],{type:"application/json"}));
    const link=node("a","");link.href=url;link.download="nirikshan-demo-evidence.json";link.click();setTimeout(()=>URL.revokeObjectURL(url),1000);
    $("verifyResult").textContent="Export created with "+Object.keys(photos).length+" photos.";
  }catch(err){$("verifyResult").textContent=err.message;}finally{$("export").disabled=false;}
});
$("verifyFile").addEventListener("change",async()=>{
  const file=$("verifyFile").files[0];if(!file)return;
  try{
    if(file.size>50000000)throw Error("Bundle exceeds 50 MB.");
    const bundle=JSON.parse(await file.text());
    if(bundle.format!=="nirikshan-demo-v1"||await sha256(new TextEncoder().encode(bundle.payload))!==bundle.digest)throw Error("Record digest mismatch.");
    const data=JSON.parse(bundle.payload);
    for(const hash of new Set(data.events.map(e=>e.photoHash).filter(Boolean))){
      if(!bundle.photos?.[hash]||await sha256(fromBase64(bundle.photos[hash]))!==hash)throw Error("Missing or changed photo: "+hash);
    }
    $("verifyResult").textContent="Bundle matches its included digests. External authenticity requires a trusted published digest.";
  }catch(err){$("verifyResult").textContent="Verification failed: "+err.message;}
});
async function preparePhoto(file){
  if(!file)return {};
  if(file.size>15000000)throw Error("Choose an image under 15 MB.");
  const bitmap=await createImageBitmap(file);
  try{
    const scale=Math.min(1,1280/Math.max(bitmap.width,bitmap.height)),canvas=document.createElement("canvas");
    canvas.width=Math.max(1,Math.round(bitmap.width*scale));canvas.height=Math.max(1,Math.round(bitmap.height*scale));
    const ctx=canvas.getContext("2d");ctx.fillStyle="#fff";ctx.fillRect(0,0,canvas.width,canvas.height);ctx.drawImage(bitmap,0,0,canvas.width,canvas.height);
    for(const quality of [.8,.65,.45]){const data=canvas.toDataURL("image/jpeg",quality).split(",")[1];if(data.length<=530000)return {photo:data,capture:"camera_requested"};}
    throw Error("Photo is too large after compression.");
  }finally{bitmap.close();}
}
$("event").addEventListener("submit",async e=>{
  e.preventDefault();
  const kind=$("kind").value, amount=$("amount").value.trim();
  if(!/^(\d+)(\.\d{1,2})?$/.test(amount||"0")&&!$("amountLabel").hidden){$("result").textContent="Enter a valid rupee amount with up to two decimal places.";return;}
  $("save").disabled=true;
  try{const payload={project:kind==="sanction"?$("projectId").value.trim():$("project").value,kind,title:$("title").value.trim(),detail:$("detail").value,amountPaise:$("amountLabel").hidden?0:Math.round(Number(amount||0)*100),ref:$("refLabel").hidden?"":$("ref").value};
    if(!$("evidenceFields").hidden)Object.assign(payload,locationEvidence,await preparePhoto($("photo").files[0]));
    payload.source=$("source").value.trim();
    if(kind==="sanction")for(const field of ["state","district","category","startPoint","endPoint","startDate","endDate"])payload[field]=$(field).value.trim();
    if(kind==="progress")payload.progress=Number($("progressValue").value);
    if(!$("quantityFields").hidden){
      const quantity=$("quantity").value.trim(),rate=$("rate").value.trim(),unit=$("unit").value.trim(),invoice=$("invoice").value.trim();
      if(quantity||rate||unit||invoice){
        if(!/^\d+(\.\d{1,3})?$/.test(quantity)||!unit||(rate&&!/^\d+(\.\d{1,2})?$/.test(rate)))throw Error("Check the quantity, unit and rate.");
        Object.assign(payload,{invoice,unit,quantityMilli:Math.round(Number(quantity)*1000),ratePaise:Math.round(Number(rate||0)*100)});
      }
    }
    if(!pending||JSON.stringify(pending.payload)!==JSON.stringify(payload))pending={payload,id:crypto.randomUUID()};
    const r=await fetch("./api/demo",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({...payload,id:pending.id})}),data=await r.json();if(!r.ok)throw Error(data.error);pending=null;$("result").textContent="Saved · receipt "+data.id;for(const field of ["title","detail","amount","photo","invoice","unit","quantity","rate","source"])$(field).value="";$("clearLocation").click();await refresh();$("project").value=data.project;render();}catch(err){$("result").textContent=err.message;}finally{$("save").disabled=false;}
});
refresh();setInterval(refresh,3000);
