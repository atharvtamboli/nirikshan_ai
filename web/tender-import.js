function extractTenderFields(text) {
  const lines=String(text).slice(0,150000).split(/\r?\n/).map(line=>line.trim()).filter(Boolean);
  const find=pattern=>{
    for (const line of lines){
      const match=line.match(pattern);
      if(match)return match[1].trim();
    }
    return "";
  };
  const draft={};
  const title=find(/^(?:work title|title|name of work|work name)\s*[:\t–-]\s*(.+)$/i);
  const id=find(/^(?:tender id|tender reference(?: number)?|reference number)\s*[:\t–-]\s*(.+)$/i);
  const description=find(/^(?:work description|description|scope)\s*[:\t–-]\s*(.+)$/i);
  const raw=find(/^(?:tender value(?: in (?:₹|inr|rs\.?))?|estimated (?:cost|value)|advertised value)\s*[:\t–-]\s*(.+)$/i);
  const amount=raw.replace(/^(?:₹|INR|Rs\.?)\s*/i,"")
    .replace(/\s*(?:₹|INR|Rs\.?|\/-)$/i,"").trim();
  const source=find(/^(?:official (?:url|link)|source(?: url)?|notice url)\s*[:\t–-]\s*(.+)$/i);
  if(title)draft.title=title.slice(0,160);
  if(id||description)draft.detail=[id&&"Tender reference: "+id,description].filter(Boolean).join("\n").slice(0,4000);
  if(/^(?:\d+|\d{1,3}(?:,\d{3})+|\d{1,2}(?:,\d{2})*,\d{3})(?:\.\d{1,2})?$/.test(amount)){
    const numeric=Number(amount.replaceAll(",",""));
    if(numeric>0&&numeric<=1e12)draft.amount=amount.replaceAll(",","");
  }
  try{
    const url=new URL(source);
    if(url.protocol==="https:"&&!url.username&&!url.password&&url.href.length<=2048)draft.source=url.href;
  }catch{}
  return draft;
}

if(typeof module!=="undefined")module.exports={extractTenderFields};
