const test = require("node:test");
const assert = require("node:assert/strict");
const { reconcile } = require("../web/bill-check.js");
test("compares certified measurements and printed bill independently", () => {
  assert.deepEqual(reconcile("12.5","10","40","510"), [50000n,40000n,1000n,11000n]);
});
test("rounds half a paisa upward for a positive line", () => {
  assert.deepEqual(reconcile("0.005","0","1","0.01"), [1n,0n,0n,1n]);
});
test("rejects silent truncation and malformed amounts", () => {
  for (const invalid of ["-2","1,40","1.005","1e5",""]) assert.throws(()=>reconcile("1","1",invalid,"1"));
});
