// Local browser journey. Build bin/nirikshan.exe first; Windows Edge is the default browser.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const net = require('node:net');
const {spawn, spawnSync} = require('node:child_process');
const wait = ms => new Promise(resolve => setTimeout(resolve, ms));
async function port() {
  const socket = net.createServer();
  await new Promise(resolve => socket.listen(0, '127.0.0.1', resolve));
  const result = socket.address().port;
  await new Promise(resolve => socket.close(resolve));
  return result;
}
async function poll(check, label="local application or browser") {
  for (let i=0;i<80;i++) {
    try { const value=await check(); if(value)return value; } catch {}
    await wait(150);
  }
  throw Error('Timed out waiting for '+label);
}
(async()=>{
  const dir=fs.mkdtempSync(path.join(os.tmpdir(),'nirikshan-browser-'));
  const project=path.resolve(__dirname,'..');
  const p=await port(), debug=await port(), base='http://127.0.0.1:'+p;
  const binary=path.join(project,'bin',process.platform==='win32'?'nirikshan.exe':'nirikshan');
  const password='browser test password';
  const env={...process.env,PORT:String(p),DATA_DIR:path.join(dir,'data'),NIRIKSHAN_DEMO:'',NIRIKSHAN_WORKSPACE:'1',NIRIKSHAN_WORKSPACE_HTTPS:''};
  const provision=spawnSync(binary,['account-add','test-department','government'],{cwd:project,env,input:password+'\n',timeout:15000,encoding:'utf8',windowsHide:true});
  if(provision.status!==0)throw Error(provision.error?.message||provision.stderr||'Account provisioning failed');
  const server=spawn(binary,['-demo'],{cwd:project,env,windowsHide:true,stdio:'ignore'});
  let browser, ws;
  try {
    await poll(async()=> (await fetch(base+'/healthz',{signal:AbortSignal.timeout(1000)})).ok,"health check");
    browser=spawn(process.env.NIRIKSHAN_BROWSER||'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',[
      '--headless=new','--disable-gpu','--no-first-run','--disable-extensions','--disable-background-networking','--remote-debugging-port='+debug,'--user-data-dir='+path.join(dir,'browser'),base+'/lifecycle.html'
    ],{windowsHide:true,stdio:'ignore'});
    const target=await poll(async()=> (await (await fetch('http://127.0.0.1:'+debug+'/json/list',{signal:AbortSignal.timeout(1000)})).json()).find(t=>t.type==='page'&&t.url.includes('/lifecycle.html')),"browser startup");
    ws=new WebSocket(target.webSocketDebuggerUrl);
    await new Promise((resolve,reject)=>{
      const timer=setTimeout(()=>reject(Error('Timed out connecting to the browser debugger')),10000);
      ws.onopen=()=>{clearTimeout(timer);resolve();};
      ws.onerror=()=>{clearTimeout(timer);reject(Error('Could not connect to the browser debugger'));};
    });
    let id=0;
    const pending=new Map();
    ws.onmessage=ev=>{const m=JSON.parse(ev.data);if(pending.has(m.id)){pending.get(m.id)(m);pending.delete(m.id);}};
    const call=(method,params={})=>new Promise((resolve,reject)=>{
      const key=++id, timer=setTimeout(()=>{pending.delete(key);reject(Error('Browser command timed out: '+method));},15000);
      pending.set(key,msg=>{clearTimeout(timer);if(msg.error)reject(Error(msg.error.message));else resolve(msg.result);});
      ws.send(JSON.stringify({id:key,method,params}));
    });
    async function query(expression){const result=await call('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(result.exceptionDetails)throw Error(JSON.stringify(result.exceptionDetails));return result.result.value;}
    async function screenshot(name) {
      if(process.env.NIRIKSHAN_SCREENSHOTS!=='1')return;
      const destination=path.join(project,'docs','screenshots');
      fs.mkdirSync(destination,{recursive:true});
      const result=await call('Page.captureScreenshot',{format:'jpeg',quality:83,captureBeyondViewport:false});
      fs.writeFileSync(path.join(destination,name+'.jpg'),Buffer.from(result.data,'base64'));
    }
    // Edge can expose the page target before committing its initial navigation.
    try{await poll(()=>query("document.getElementById('name')?.textContent.includes('Aravalli')"),"project render");}
    catch {
      await call('Page.navigate',{url:base+'/lifecycle.html'}).catch(()=>{});
      await poll(()=>query("document.getElementById('name')?.textContent.includes('Aravalli')"),"project render after retry");
    }
    assert.match(await query("document.getElementById('riskSummary').innerText"),/awaiting certification/);
    await call('Emulation.setDeviceMetricsOverride',{width:1440,height:1050,deviceScaleFactor:1,mobile:false});
    await query("window.scrollTo(0,document.querySelector('.project-banner').offsetTop-70);true");
    await screenshot('project-tracker');
    await query("document.getElementById('project').value='demo-school';document.getElementById('project').dispatchEvent(new Event('change'))");
    assert.match(await query("document.getElementById('current').innerText"),/Maintenance/);
    await query("document.getElementById('project').value='demo-bridge';document.getElementById('project').dispatchEvent(new Event('change'))");
    await query("document.getElementById('role').value='engineer';document.getElementById('role').dispatchEvent(new Event('change'));document.getElementById('kind').value='quality_test';document.getElementById('kind').dispatchEvent(new Event('change'));Object.entries({title:'Concrete test',detail:'Illustrative lab report',sample:'Q-17',testUnit:'MPa',testLimit:'30',testValue:'28'}).forEach(([id,value])=>document.getElementById(id).value=value);document.getElementById('event').requestSubmit()");
    await poll(()=>query("document.getElementById('qualityResults').innerText.includes('Q-17')"),"quality test result");
    await call('Emulation.setDeviceMetricsOverride',{width:390,height:844,deviceScaleFactor:1,mobile:true});
    assert.equal(await query("document.documentElement.scrollWidth <= innerWidth"),true);
    await query("document.getElementById('role').value='citizen';document.getElementById('role').dispatchEvent(new Event('change'));document.getElementById('title').value='Browser inspection report';document.getElementById('detail').value='Approach road needs attention'");
    const root=(await call('DOM.getDocument')).root.nodeId;
    const photo=(await call('DOM.querySelector',{nodeId:root,selector:'#photo'})).nodeId;
    await call('DOM.setFileInputFiles',{nodeId:photo,files:[path.join(project,'web','icon-192.png')]});
    await query("document.getElementById('event').requestSubmit()");
    await poll(()=>query("document.getElementById('reports').innerText.includes('Browser inspection report')"),"report submission");
    await poll(()=>query("document.querySelector('#reports img')?.complete && document.querySelector('#reports img')?.naturalWidth>0"),"photo render");
    await query("[...document.querySelectorAll('#reports > .record')].find(e=>e.innerText.includes('Browser inspection report')).querySelector('button').click();document.getElementById('title').value='Inspection scheduled';document.getElementById('event').requestSubmit()");
    await poll(()=>query("[...document.querySelectorAll('#reports > .record')].find(e=>e.innerText.includes('Browser inspection report')).innerText.includes('Acknowledged')"),"department acknowledgement");
    await call('Page.setDownloadBehavior',{behavior:'allow',downloadPath:dir});
    await query("document.getElementById('export').click()");
    const bundle=path.join(dir,'nirikshan-demo-evidence.json');
    await poll(()=>fs.existsSync(bundle),"signed bundle download");
    const verify=(await call('DOM.querySelector',{nodeId:root,selector:'#verifyFile'})).nodeId;
    const original=JSON.parse(fs.readFileSync(bundle,'utf8'));
    assert.equal(original.format,'nirikshan-demo-v2');
    await query("document.getElementById('trustedKey').value="+JSON.stringify(original.fingerprint));
    await call('DOM.setFileInputFiles',{nodeId:verify,files:[bundle]});
    await poll(()=>query("document.getElementById('verifyResult').innerText.includes('Signature matches trusted fingerprint')"),"signature verification");
    const changed=path.join(dir,'changed.json'), altered=JSON.parse(fs.readFileSync(bundle,'utf8'));
    altered.payload+=' ';altered.digest=require('node:crypto').createHash('sha256').update(altered.payload).digest('hex');fs.writeFileSync(changed,JSON.stringify(altered));
    await call('DOM.setFileInputFiles',{nodeId:verify,files:[changed]});
    await poll(()=>query("document.getElementById('verifyResult').innerText.includes('Verification failed')"),"tamper rejection");
    await call('Emulation.setDeviceMetricsOverride',{width:1440,height:900,deviceScaleFactor:1,mobile:false});
    assert.equal(await query("document.documentElement.scrollWidth <= innerWidth"),true);
    await query("location.href='"+base+"/records-request.html';true");
    await poll(()=>query("document.getElementById('requestDraft')?.value.includes('2026_CEPWD_566165_4')"),"records request draft");
    await query("document.getElementById('requestLanguage').value='hi';document.getElementById('requestLanguage').dispatchEvent(new Event('change'))");
    assert.match(await query("document.getElementById('requestDraft').value"),/निविदा आईडी/);
    await query("location.href='"+base+"/';true");
    await poll(()=>query("document.getElementById('detail')?.innerText.includes('Boramcha')"),"public atlas");
    await screenshot('public-atlas');
    await call('Page.navigate',{url:base+'/lifecycle.html?mode=workspace'});
    await poll(()=>query("document.getElementById('workspaceAccount')?.hidden===false"),"workspace sign-in page");
    const ui=(expression)=>query("(async ($)=>{"+expression+"})(id=>document.getElementById(id))");
    await ui("$('accountName').value='test-department';$('accountPassword').value="+JSON.stringify(password)+";$('signIn').requestSubmit()");
    await poll(()=>ui("return $('accountStatus')?.innerText.includes('test-department')"),"workspace login");
    await ui("Object.entries({projectId:'browser-road',title:'Browser road',amount:'250000'}).forEach(([id,value])=>$(id).value=value);$('event').requestSubmit()");
    await poll(()=>ui("return $('name')?.innerText==='Browser road'"),"recorded project sanction");
    await ui("$('kind').value='tender';$('kind').dispatchEvent(new Event('change'));$('tenderText').value="+JSON.stringify("Tender ID: TEST-001\nTitle: Road repairs\nTender Value: Rs. 2,30,000")+";$('extractTender').click()");
    assert.equal(await ui("return $('amount').value"),'230000');
    await ui("$('event').requestSubmit()");
    await poll(()=>ui("return $('current')?.innerText.includes('Tender published')"),"imported tender");
    assert.match(await ui("return await fetch($('auditDownload').href).then(r=>r.text())"),/Road repairs/);
    await ui("$('signOut').click()");
    await poll(()=>ui("return $('workspaceContent').hidden&&$('signIn').hidden===false"),"workspace sign-out");
    console.log('PASS: demo and authenticated workspace browser journeys');
  } finally { if(ws)ws.close();if(browser)browser.kill();server.kill(); }
})().catch(error=>{console.error(error.message);process.exitCode=1;});
