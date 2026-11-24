const $ = id => document.getElementById(id);
let projects = [];
const requests = [
  ["Administrative sanction and technical sanction with approved estimate and scope.","स्वीकृत लागत और कार्यक्षेत्र सहित प्रशासनिक एवं तकनीकी स्वीकृति की प्रति।"],
  ["Tender award, accepted quotation, signed agreement and work order.","निविदा आवंटन, स्वीकृत दर, हस्ताक्षरित अनुबंध और कार्यादेश की प्रति।"],
  ["Approved bill of quantities and all sanctioned variations.","स्वीकृत मात्रा सूची और सभी अनुमोदित परिवर्तनों की प्रति।"],
  ["Certified running and final bills with relevant measurement book pages.","प्रमाणित चालू एवं अंतिम बिल तथा संबंधित माप पुस्तिका के पृष्ठ।"],
  ["Project-wise payments, dates, amounts and corresponding voucher references.","परियोजना के भुगतान, तारीख, राशि और संबंधित वाउचर संदर्भ।"],
  ["Inspection reports, material quality test results and recorded corrective actions.","निरीक्षण रिपोर्ट, सामग्री गुणवत्ता परीक्षण के परिणाम और दर्ज सुधारात्मक कार्रवाई।"],
  ["Latest documented physical progress, expected completion date and approved extensions.","नवीनतम दर्ज भौतिक प्रगति, अपेक्षित पूर्णता तिथि और स्वीकृत समय विस्तार।"]
];
function project(){return projects.find(p=>p.id===$("requestProject").value);}
function renderOptions(){
  const hi=$("requestLanguage").value==="hi";
  const checked=new Set([...$("requestOptions").querySelectorAll("input:checked")].map(e=>e.value));
  const initial=$("requestOptions").children.length===0;
  $("requestOptions").replaceChildren(...requests.map((r,index)=>{
    const label=document.createElement("label"),box=document.createElement("input"),span=document.createElement("span");
    box.type="checkbox";box.value=String(index);box.checked=initial||checked.has(box.value);box.style.width="auto";
    label.style.display="flex";label.style.alignItems="start";label.style.margin="12px 0";
    span.textContent=r[hi?1:0];label.append(box,span);return label;
  }));
}
function generate(){
  const p=project();if(!p)return;
  const hi=$("requestLanguage").value==="hi",selected=[...$("requestOptions").querySelectorAll("input:checked")].map(e=>requests[Number(e.value)][hi?1:0]);
  if(!selected.length){$("requestStatus").textContent="Select at least one record.";return;}
  $("requestAuthority").textContent=p.authority;
  $("requestSource").href=p.source;
  const header=hi?["सेवा में,","जन सूचना अधिकारी",p.authority,"","विषय: सार्वजनिक निर्माण कार्य के अभिलेख उपलब्ध कराने हेतु सूचना का अनुरोध।","परियोजना: "+p.name,"निविदा आईडी: "+p.id,"स्थान: "+p.place,"","कृपया निम्न अभिलेखों की प्रमाणित प्रतियां अथवा उपलब्ध इलेक्ट्रॉनिक प्रतियां प्रदान करें।"]:
    ["To","The Public Information Officer",p.authority,"","Subject: Request for records concerning a public works project","Project: "+p.name,"Tender ID: "+p.id,"Location: "+p.place,"","Please provide certified copies or available electronic copies of the following records:"];
  const footer=hi?["","यदि कोई अभिलेख उपलब्ध नहीं है, कृपया उसकी स्थिति बताएं।","आवेदक का नाम: [भरें]","संपर्क / पता: [भरें]","दिनांक: [भरें]"]:["","If a record is unavailable, please indicate its recorded status.","Applicant name: [complete]","Contact / address: [complete]","Date: [complete]"];
  $("requestDraft").value=[...header,...selected.map((s,i)=>(i+1)+". "+s),...footer].join("\n");$("requestStatus").textContent="Draft ready for your review.";
}
$("requestProject").addEventListener("change",generate);
$("requestLanguage").addEventListener("change",()=>{renderOptions();generate();});
$("generateRequest").addEventListener("click",generate);
$("copyRequest").addEventListener("click",async()=>{try{await navigator.clipboard.writeText($("requestDraft").value);$("requestStatus").textContent="Draft copied.";}catch{$("requestDraft").select();$("requestStatus").textContent="Draft selected. Use your device's copy command.";}});
$("downloadRequest").addEventListener("click",()=>{const url=URL.createObjectURL(new Blob([$("requestDraft").value],{type:"text/plain;charset=utf-8"}));const a=document.createElement("a");a.href=url;a.download="nirikshan-records-request.txt";a.click();setTimeout(()=>URL.revokeObjectURL(url),1000);});
fetch("./projects.json").then(r=>{if(!r.ok)throw Error("Catalog unavailable");return r.json();}).then(data=>{
  projects=data;for(const p of projects){const option=document.createElement("option");option.value=p.id;option.textContent=p.name;$("requestProject").append(option);}renderOptions();generate();
}).catch(err=>{$("requestStatus").textContent=err.message;});
