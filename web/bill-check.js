"use strict";
// Convert decimal text to integers before multiplying rupees and quantities.
function scaled(value, digits) {
  if (!/^\d{1,12}(\.\d+)?$/.test(value)) throw new Error("decimal");
  const [whole, fraction = ""] = value.split(".");
  if (fraction.length > digits) throw new Error("precision");
  return BigInt(whole + fraction.padEnd(digits, "0"));
}
function reconcile(billed, measured, rate, subtotal) {
  const claim = scaled(billed, 3), measuredQty = scaled(measured, 3);
  const price = scaled(rate, 2), printed = scaled(subtotal, 2);
  const line = qty => (qty * price + 500n) / 1000n;
  return [line(claim), line(measuredQty), printed - line(claim), printed - line(measuredQty)];
}
if (typeof module !== "undefined") module.exports = { reconcile };
if (typeof document !== "undefined") {
  const copy = {
    en: {title:"Check a bill line", intro:"Choose one item from a bill, the matching measurement-book entry and its contract rate. Use the same unit for both quantities. Enter the line subtotal before tax or deductions.", inputs:"Figures from the records", fields:["Billed quantity","Measured quantity","Contract rate per unit (₹)","Printed line subtotal (₹)"], calculate:"Compare figures", results:["Billed quantity × contract rate","Measured quantity × contract rate","Printed subtotal minus billed calculation","Printed subtotal minus measured calculation"], privacy:"This worksheet runs on your device. A difference needs review of the units, approved variations and supporting records.", error:"Enter nonnegative figures, with up to three quantity decimals and two money decimals."},
    hi: {title:"बिल की एक मद जाँचें", intro:"बिल की एक मद, माप पुस्तिका की संबंधित प्रविष्टि और अनुबंध की दर चुनें। दोनों मात्राओं की इकाई समान रखें। कर या कटौती से पहले का उप-योग दर्ज करें।", inputs:"अभिलेखों में दर्ज आँकड़े", fields:["बिल में मात्रा","मापी गई मात्रा","अनुबंध दर प्रति इकाई (₹)","बिल में दर्ज उप-योग (₹)"], calculate:"तुलना करें", results:["बिल की मात्रा × अनुबंध दर","मापी गई मात्रा × अनुबंध दर","दर्ज उप-योग और बिल की गणना का अंतर","दर्ज उप-योग और माप की गणना का अंतर"],privacy:"गणना आपके उपकरण पर होती है। अंतर मिलने पर इकाइयों, स्वीकृत बदलावों और संबंधित अभिलेखों की जाँच करें।",error:"धनराशि में दो और मात्रा में तीन दशमलव स्थान तक गैर-ऋणात्मक अंक दर्ज करें।"}
  };
  const langButton=document.querySelector("#language"), result=document.querySelector("#result");
  let lang=new URLSearchParams(location.search).get("lang")==="hi"?"hi":"en";
  const inputs=Array.from({length:4},()=>{const label=document.createElement("label"),span=document.createElement("span"),field=document.createElement("input");field.inputMode="decimal";field.required=true;field.maxLength=16;label.append(span,field);document.querySelector("#fields").append(label);return field;});
  function paint(){document.documentElement.lang=lang;langButton.textContent=lang==="en"?"हिन्दी":"English";document.querySelectorAll("[data-i]").forEach(el=>el.textContent=copy[lang][el.dataset.i]);inputs.forEach((el,i)=>el.previousSibling.textContent=copy[lang].fields[i]);result.replaceChildren();}
  function money(value){const abs=value<0n?-value:value;return (value<0n?"−":"")+"₹"+(abs/100n).toLocaleString("en-IN")+"."+(abs%100n).toString().padStart(2,"0");}
  document.querySelector("form").addEventListener("submit",event=>{event.preventDefault();result.replaceChildren();try{const dl=document.createElement("dl");reconcile(...inputs.map(el=>el.value.trim())).forEach((value,i)=>{const key=document.createElement("dt"),amount=document.createElement("dd");key.textContent=copy[lang].results[i];amount.textContent=money(value);dl.append(key,amount);});result.append(dl);}catch{result.textContent=copy[lang].error;}});
  langButton.onclick=()=>{lang=lang==="en"?"hi":"en";paint();};paint();
}
