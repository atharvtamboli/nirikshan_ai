// Local browser journey. Build bin/nirikshan.exe first; Windows Edge is the default browser.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const net = require('node:net');
const {spawn} = require('node:child_process');
const wait = ms => new Promise(resolve => setTimeout(resolve, ms));
async function port() {
  const socket = net.createServer();
  await new Promise(resolve => socket.listen(0, '127.0.0.1', resolve));
  const result = socket.address().port;
  await new Promise(resolve => socket.close(resolve));
  return result;
}
async function poll(check) {
  for (let i=0;i<80;i++) {
    try { const value=await check(); if(value)return value; } catch {}
    await wait(150);
  }
  throw Error('Timed out waiting for local application or browser');
}
(async()=>{
  const dir=fs.mkdtempSync(path.join(os.tmpdir(),'nirikshan-browser-'));
  const project=path.resolve(__dirname,'..');
  const p=await port(), debug=await port(), base='http://127.0.0.1:'+p;
  const server=spawn(path.join(project,'bin','nirikshan.exe'),[],{cwd:project,env:{...process.env,PORT:String(p),DATA_DIR:path.join(dir,'data'),NIRIKSHAN_DEMO:'1'},windowsHide:true,stdio:'ignore'});
  let browser, ws;
  try {
    await poll(async()=> (await fetch(base+'/healthz')).ok);
    browser=spawn(process.env.NIRIKSHAN_BROWSER||'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',[
      '--headless=new','--disable-gpu','--no-first-run','--disable-extensions','--disable-background-networking','--remote-debugging-port='+debug,'--user-data-dir='+path.join(dir,'browser'),base+'/lifecycle.html'
    ],{windowsHide:true,stdio:'ignore'});
    const target=await poll(async()=> (await (await fetch('http://127.0.0.1:'+debug+'/json/list')).json()).find(t=>t.type==='page'));
    ws=new WebSocket(target.webSocketDebuggerUrl);
    await new Promise((resolve,reject)=>{ws.onopen=resolve;ws.onerror=reject;});
    let id=0;
    const pending=new Map();
    ws.onmessage=ev=>{const m=JSON.parse(ev.data);if(pending.has(m.id)){pending.get(m.id)(m);pending.delete(m.id);}};
    const call=(method,params={})=>new Promise((resolve,reject)=>{
      const key=++id, timer=setTimeout(()=>{pending.delete(key);reject(Error('Browser command timed out: '+method));},15000);
      pending.set(key,msg=>{clearTimeout(timer);if(msg.error)reject(Error(msg.error.message));else resolve(msg.result);});
      ws.send(JSON.stringify({id:key,method,params}));
    });
    async function query(expression){const result=await call('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(result.exceptionDetails)throw Error(JSON.stringify(result.exceptionDetails));return result.result.value;}
    await poll(()=>query("document.getElementById('name')?.textContent.includes('Aravalli')"));
    assert.match(await query("document.getElementById('riskSummary').innerText"),/awaiting certification/);
    await call('Emulation.setDeviceMetricsOverride',{width:390,height:844,deviceScaleFactor:1,mobile:true});
    assert.equal(await query("document.documentElement.scrollWidth <= innerWidth"),true);
    await query("document.getElementById('role').value='citizen';document.getElementById('role').dispatchEvent(new Event('change'));document.getElementById('title').value='Browser inspection report';document.getElementById('detail').value='Approach road needs attention'");
    const root=(await call('DOM.getDocument')).root.nodeId;
    const photo=(await call('DOM.querySelector',{nodeId:root,selector:'#photo'})).nodeId;
    await call('DOM.setFileInputFiles',{nodeId:photo,files:[path.join(project,'web','icon-192.png')]});
    await query("document.getElementById('event').requestSubmit()");
    await poll(()=>query("document.getElementById('reports').innerText.includes('Browser inspection report')"));
    await poll(()=>query("document.querySelector('#reports img')?.complete && document.querySelector('#reports img')?.naturalWidth>0"));
    await query("document.getElementById('role').value='government';document.getElementById('role').dispatchEvent(new Event('change'));document.getElementById('kind').value='acknowledge';document.getElementById('kind').dispatchEvent(new Event('change'));document.getElementById('title').value='Inspection scheduled';document.getElementById('event').requestSubmit()");
    await poll(()=>query("document.getElementById('reports').innerText.includes('Acknowledged')"));
    await call('Page.setDownloadBehavior',{behavior:'allow',downloadPath:dir});
    await query("document.getElementById('export').click()");
    const bundle=path.join(dir,'nirikshan-demo-evidence.json');
    await poll(()=>fs.existsSync(bundle));
    const verify=(await call('DOM.querySelector',{nodeId:root,selector:'#verifyFile'})).nodeId;
    await call('DOM.setFileInputFiles',{nodeId:verify,files:[bundle]});
    await poll(()=>query("document.getElementById('verifyResult').innerText.includes('matches its included digests')"));
    const changed=path.join(dir,'changed.json'), altered=JSON.parse(fs.readFileSync(bundle,'utf8'));
    altered.payload+=' ';fs.writeFileSync(changed,JSON.stringify(altered));
    await call('DOM.setFileInputFiles',{nodeId:verify,files:[changed]});
    await poll(()=>query("document.getElementById('verifyResult').innerText.includes('Verification failed')"));
    await call('Emulation.setDeviceMetricsOverride',{width:1440,height:900,deviceScaleFactor:1,mobile:false});
    assert.equal(await query("document.documentElement.scrollWidth <= innerWidth"),true);
    console.log('PASS: Mobile and desktop layout, photo report, acknowledgement, evidence export and tamper detection');
  } finally { if(ws)ws.close();if(browser)browser.kill();server.kill(); }
})().catch(error=>{console.error(error.message);process.exitCode=1;});
