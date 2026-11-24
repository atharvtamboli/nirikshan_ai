/* Local, reproducible checks for approved records in a single project. */
(function (root) {
  "use strict";
  function normalizeRef(value) {
    return String(value || "").normalize("NFKC").trim().toUpperCase().replace(/\s+/g, " ");
  }

  function analyze(project, records, language = "en") {
    const hi = language === "hi";
    const rows = (Array.isArray(records) ? records : []).filter(row => row && row.project === project.id);
    const findings = [];
    const groups = new Map();
    for (const row of rows) {
      const ref = normalizeRef(row.reference);
      if (!ref || !["bill", "payment"].includes(row.kind)) continue;
      if (!groups.has(ref)) groups.set(ref, { bill: [], payment: [] });
      groups.get(ref)[row.kind].push(row);
    }
    for (const [ref, group] of groups) {
      if (group.bill.length > 1) {
        findings.push({
          type: "review",
          title: hi ? "बिल का संदर्भ दोहराया गया" : "Repeated bill reference",
          detail: hi ? ref + " कई बिल अभिलेखों में है। संशोधित बिल और मूल दस्तावेज़ की जाँच करें।" : ref + " appears on multiple bill records. Check revisions and original documents before counting the amount twice.",
          ids: group.bill.map(x => x.id)
        });
      }
      if (group.bill.length === 1 && Number.isSafeInteger(group.bill[0].amountPaise)) {
        const bill = group.bill[0];
        for (const payment of group.payment) {
          if (Number.isSafeInteger(payment.amountPaise) && payment.amountPaise > bill.amountPaise) {
            findings.push({
              type: "review",
              title: hi ? "भुगतान दर्ज बिल से अधिक है" : "Payment exceeds the linked bill",
              detail: hi ? ref + " के दस्तावेज़ों में राशि और कार्य का दायरा जाँचें।" : "Compare the original amounts and scope for reference " + ref + ". A payment can cover several bills.",
              ids: [bill.id, payment.id]
            });
          }
        }
      }
    }
    for (const row of rows) {
      if (row.kind === "award" && Number.isSafeInteger(project.tenderPaise) && Number.isSafeInteger(row.amountPaise) && row.amountPaise > project.tenderPaise) {
        findings.push({
          type: "review",
          title: hi ? "अनुबंध राशि निविदा अनुमान से अधिक है" : "Award exceeds the advertised estimate",
          detail: hi ? "मूल निविदा और अनुबंध में कर, बदलाव और कार्य का दायरा देखें।" : "Check taxes, amendments and package scope in the tender and award before interpreting the difference.",
          ids: [row.id]
        });
      }
    }
    const news = rows.filter(row => row.kind === "news");
    if (news.length) findings.push({
      type: "context",
      title: hi ? "समाचार स्रोत" : "News coverage",
      detail: hi ? "खबर के दावे को मूल सरकारी दस्तावेज़ अथवा स्वतंत्र निरीक्षण से मिलाएँ।" : "Check each reported claim against a primary document or independent inspection. Review approval records the source of the story.",
      ids: news.map(x => x.id)
    });
    return findings;
  }
  root.NirikshanAnalysis = Object.freeze({ analyze });
})(globalThis);
