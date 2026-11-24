const test=require("node:test");
const assert=require("node:assert/strict");
const {extractTenderFields}=require("../web/tender-import.js");

test("extracts an Indian tender reference, description, amount and source",()=>{
  const draft=extractTenderFields([
    "Tender ID: 2026_CEPWD_566165_4",
    "Work Title: Bridge work",
    "Work Description: Approach and bridge repair",
    "Tender Value in ₹: ₹1,97,55,000.50",
    "Official URL: https://example.gov.in/tender?id=4",
  ].join("\n"));
  assert.deepEqual(draft,{
    title:"Bridge work",
    detail:"Tender reference: 2026_CEPWD_566165_4\nApproach and bridge repair",
    amount:"19755000.50",
    source:"https://example.gov.in/tender?id=4",
  });
});

test("leaves unsupported amounts and non-HTTPS or credential-bearing sources for manual entry",()=>{
  for(const amount of ["20 crore","invalid","1,9,755","100.123","0","1000000000001"]){
    assert.deepEqual(extractTenderFields("Tender value: "+amount+"\nSource: https://user:pass@example.com"),{});
  }
});
