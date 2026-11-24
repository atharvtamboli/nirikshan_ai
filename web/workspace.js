function renderPortfolio(){
  const projects=events.filter(e=>e.kind==="sanction");
  const cards=projects.map(e=>{
    const rows=events.filter(row=>row.project===e.project);
    const stage=stages.findLast(s=>rows.some(row=>row.kind===s));
    const billed=rows.filter(row=>row.kind==="bill").reduce((sum,row)=>sum+row.amountPaise,0);
    const certified=rows.filter(row=>row.kind==="certificate").reduce((sum,row)=>sum+row.amountPaise,0);
    const card=node("div","","metric"),button=node("button",e.title);
    button.type="button";
    button.addEventListener("click",()=>{$("project").value=e.project;selectedStage="";render();$("name").scrollIntoView();});
    card.append(button,node("span",labels[stage]));
    if(account.role!=="citizen")card.append(node("span",money(Math.max(0,billed-certified))+" awaiting certification"));
    return card;
  });
  $("portfolioProjects").replaceChildren(...(cards.length?cards:[node("p",account.role==="government"?"Record the first project sanction below.":"No assigned projects have a recorded sanction yet.")]));
}

function showAccount(next){
  account=next;
  sessionRevision++;
  pending=null;
  selectedStage="";
  events=[];
  $("event").reset();
  $("clearLocation").click();
  $("tenderText").value="";
  $("tenderFile").value="";
  $("tenderStatus").textContent="";
  $("result").textContent="";
  $("verifyFile").value="";
  $("trustedKey").value="";
  $("projectSearch").value="";
  $("verifyResult").textContent="";
  $("status").textContent="";
  $("project").replaceChildren();
  for(const id of ["discover","history","reports","newsList","qualityResults","money","fundFlow","riskSummary","stageDetail","routeMap","metrics"])$(id).replaceChildren();
  for(const id of ["name","route","current","schedule","progress","reportStats"])$(id).textContent="";
  $("project").closest("label").hidden=true;
  $("projectContent").hidden=true;
  $("portfolioProjects").replaceChildren();
  $("workspaceContent").hidden=!next;
  $("signIn").hidden=!!next;
  $("signedIn").hidden=!next;
  $("accountStatus").textContent=next?next.name+" · "+(next.role==="government"?"department":next.role):"";
  if(next){
    $("role").value=next.role;
    $("role").dispatchEvent(new Event("change"));
    render();
  }
}

async function loadWorkspace(){
  $("modeLabel").textContent="Authenticated workspace";
  document.querySelector(".demo-guide").hidden=true;
  $("workspaceAccount").hidden=false;
  $("portfolio").hidden=false;
  $("role").disabled=true;
  $("deskHeading").textContent="Record an update";
  $("deskDescription").textContent="Entries are attributed to your account and checked against the project history.";
  $("save").textContent="Save project record";
  showAccount(null);
  $("signIn").addEventListener("submit",signIntoWorkspace);
  $("signOut").addEventListener("click",signOutOfWorkspace);
  $("tenderFile").addEventListener("change",loadTenderFile);
  $("extractTender").addEventListener("click",populateTenderFields);
  try{
    const r=await fetch("./api/workspace/session",{cache:"no-store"});
    if(r.ok){showAccount(await r.json());await refresh();}
    else if(r.status===404){$("signIn").hidden=true;$("sessionStatus").textContent="Start the local server with go run . -workspace to enable sign-in.";}
    else if(r.status!==401)throw Error("Could not load the workspace session.");
  }catch(error){$("sessionStatus").textContent=error.message;}
}

async function signIntoWorkspace(event){
  event.preventDefault();
  const button=event.submitter||$("signIn").querySelector("button");button.disabled=true;
  $("sessionStatus").textContent="";
  try{
    const response=await fetch("./api/workspace/session",{method:"POST",headers:writeHeaders,
      body:JSON.stringify({name:$("accountName").value.trim(),password:$("accountPassword").value})});
    const data=await response.json();
    if(!response.ok)throw Error(data.error);
    showAccount(data);
    await refresh();
  }catch(error){$("sessionStatus").textContent=error.message;}finally{$("accountPassword").value="";button.disabled=false;}
}

async function signOutOfWorkspace(){
  $("signOut").disabled=true;
  try{
    const response=await fetch("./api/workspace/session",{method:"DELETE",headers:writeHeaders});
    if(!response.ok&&response.status!==401)throw Error("Sign-out failed. Try again.");
    showAccount(null);
    $("sessionStatus").textContent="Signed out.";
  }catch(error){$("sessionStatus").textContent=error.message;}finally{$("signOut").disabled=false;}
}

async function loadTenderFile(){
  const file=$("tenderFile").files[0];if(!file)return;
  try{
    if(file.size>150000)throw Error("Choose a text file under 150 KB.");
    $("tenderText").value=new TextDecoder("utf-8",{fatal:true}).decode(await file.arrayBuffer());
    $("tenderStatus").textContent="Text loaded. Select Extract fields.";
  }catch(error){$("tenderStatus").textContent=error.message;}
}

function populateTenderFields(){
  const draft=extractTenderFields($("tenderText").value);
  for(const [field,value] of Object.entries(draft))if(value&&!$(field).value)$(field).value=value;
  $("tenderStatus").textContent=Object.keys(draft).length?"Review the populated fields. Existing entries were preserved.":"No recognized fields. Enter the tender details manually.";
}
