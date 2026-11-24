"use strict";
const byId = id => document.getElementById(id);
const queue = byId("queue");
const message = byId("message");
let token = "";
let epoch = 0;

function element(tag, text, className) {
  const node = document.createElement(tag);
  if (text !== undefined) node.textContent = text;
  if (className) node.className = className;
  return node;
}

function resetSession() {
  epoch++;
  token = "";
  byId("token").value = "";
  byId("connect").hidden = false;
  byId("session").hidden = true;
  queue.replaceChildren();
}

async function request(method, body) {
  const response = await fetch("api/review", {
    method,
    headers: {Authorization: "Bearer " + token, ...(body ? {"Content-Type":"application/json"} : {})},
    body: body ? JSON.stringify(body) : undefined, cache: "no-store", redirect: "error"
  });
  if (response.status === 401 || response.status === 403) throw new Error("Reviewer access was denied. Check the token and try again.");
  if (response.status === 409) throw new Error("This submission has already been decided. Refresh the queue.");
  if (!response.ok || !(response.headers.get("Content-Type") || "").includes("application/json")) throw new Error("The review service is unavailable. Open the server-backed pilot and try again.");
  return response.json();
}

function card(row) {
  const panel = element("article");
  panel.append(element("span", row.kind + " · " + row.project + " · " + row.date, "meta"), element("h3", row.title));
  if (row.reference) panel.append(element("p", "Document reference · " + row.reference));
  if (Number.isSafeInteger(row.amountPaise)) panel.append(element("p", "Reported amount · " + new Intl.NumberFormat("en-IN", {style:"currency",currency:"INR"}).format(row.amountPaise / 100)));
  try {
    const url = new URL(row.url);
    if (["https:", "http:"].includes(url.protocol)) {
      const link = element("a", "Open source at " + url.hostname + " ↗", "url");
      link.href = url.href;
      link.target = "_blank";
      link.rel = "noopener noreferrer";
      panel.append(link);
    }
  } catch { /* Observations may have no source URL. */ }
  const form = element("form");
  const label = element("label", "Decision reason");
  const note = element("textarea");
  note.required = true;
  note.maxLength = 600;
  note.rows = 2;
  label.append(note);
  const actions = element("div", undefined, "actions");
  const approve = element("button", "Approve publication");
  approve.value = "approve";
  const reject = element("button", "Reject submission", "alt");
  reject.value = "reject";
  actions.append(approve, reject);
  const status = element("p");
  status.setAttribute("role", "status");
  form.append(label, actions, status);
  form.addEventListener("submit", async event => {
    event.preventDefault();
    if (!note.value.trim() || !event.submitter) return;
    approve.disabled = reject.disabled = true;
    const version = epoch;
    try {
      await request("POST", {id:row.id,decision:event.submitter.value,reason:note.value.trim()});
      if (version !== epoch) return;
      await refresh("Decision saved. ");
      byId("refresh").focus();
    } catch (error) {
      status.textContent = error.message;
      approve.disabled = reject.disabled = false;
    }
  });
  panel.append(form);
  return panel;
}

async function refresh(prefix = "") {
  const version = ++epoch;
  message.textContent = prefix + "Loading queue…";
  try {
    const rows = await request("GET");
    if (version !== epoch) return;
    if (!Array.isArray(rows)) throw new Error("Unexpected response from the review service.");
    queue.replaceChildren(...rows.map(card));
    byId("connect").hidden = true;
    byId("session").hidden = false;
    message.textContent = prefix + (rows.length ? rows.length + " awaiting review." : "All caught up. There are no pending submissions.");
  } catch (error) {
    if (version !== epoch) return;
    resetSession();
    message.textContent = prefix + error.message;
  }
}
byId("connect").addEventListener("submit", event => {
  event.preventDefault();
  token = byId("token").value.trim();
  byId("token").value = "";
  refresh();
});
byId("refresh").addEventListener("click", () => refresh());
byId("lock").addEventListener("click", () => {resetSession(); message.textContent = "Workspace locked."; byId("token").focus();});
window.addEventListener("pagehide", resetSession);
