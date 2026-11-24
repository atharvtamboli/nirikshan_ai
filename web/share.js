(() => {
  'use strict';
  const isHindi = () => document.documentElement.lang === 'hi';
  const label = (english, hindi) => isHindi() ? hindi : english;
  let encoder;
  function loadEncoder() {
    if (!encoder) encoder = new Promise((resolve, reject) => {
      const script = document.createElement('script');
      script.src = './vendor/qrcode.js';
      script.onload = () => resolve(window.qrcode);
      script.onerror = () => { encoder = undefined; reject(new Error('QR library unavailable')); };
      document.head.append(script);
    });
    return encoder;
  }

  function node(tag, text) {
    const el = document.createElement(tag);
    if (text !== undefined) el.textContent = text;
    return el;
  }
  document.addEventListener('click', async event => {
    const launch = event.target.closest('[data-share-qr]');
    if (!launch) return;
    const url = new URL(location.href);
    url.hash = '';
    const modal = node('dialog');
    modal.style.cssText = 'border:1px solid #527568;border-radius:16px;padding:24px;width:min(90vw,420px);max-height:90vh;color:#123d35';
    const title = node('h2', label('Share this bridge', 'इस पुल को साझा करें'));
    modal.setAttribute('aria-label', title.textContent);
    const close = node('button', label('Close', 'बंद करें'));
    close.type = 'button';
    close.onclick = () => modal.close();
    modal.addEventListener('close', () => { modal.remove(); launch.focus(); });
    const codeBox = node('div');
    codeBox.style.textAlign = 'center';
    const link = node('a', url.href);
    link.href = url.href;
    link.style.overflowWrap = 'anywhere';
    const line = node('p');
    line.append(link);
    modal.append(title, codeBox, line, close);
    document.body.append(modal);
    modal.showModal();
    try {
      const makeQr = await loadEncoder();
      const qr = makeQr(0, 'M');
      qr.addData(url.href);
      qr.make();
      const pixels = (qr.getModuleCount() + 8) * 8;
      const canvas = node('canvas');
      canvas.width = canvas.height = pixels;
      canvas.style.cssText = 'width:min(100%,260px);height:auto';
      canvas.setAttribute('aria-label', label('Project link QR code', 'परियोजना लिंक का क्यूआर कोड'));
      canvas.setAttribute('role','img');
      const ctx = canvas.getContext('2d');
      ctx.fillStyle = 'white'; ctx.fillRect(0,0,pixels,pixels);
      ctx.fillStyle = '#102d26';
      for (let row = 0; row < qr.getModuleCount(); row++) {
        for (let col = 0; col < qr.getModuleCount(); col++) {
          if (qr.isDark(row,col)) ctx.fillRect((col+4)*8,(row+4)*8,8,8);
        }
      }
      const save = node('a',label('Download QR image','क्यूआर चित्र डाउनलोड करें'));
      save.download = 'nirikshan-' + (url.searchParams.get('project') || 'atlas') + '.png';
      save.href = canvas.toDataURL('image/png');
      codeBox.append(canvas,node('br'),save);
    } catch {
      codeBox.textContent = label('QR unavailable. Use the project link below.', 'क्यूआर उपलब्ध नहीं है। नीचे परियोजना लिंक का उपयोग करें।');
    }
  });

  let saved = '', cached = false;
  const offlineNote = node('div');
  offlineNote.setAttribute('role','status');
  offlineNote.style.cssText = 'background:#fff0bf;color:#422e04;text-align:center;padding:.65rem;font-size:.875rem';
  document.body.prepend(offlineNote);
  function status() {
    offlineNote.hidden = !(cached || !navigator.onLine);
    if (offlineNote.hidden) return;
    const last = saved && new Date(saved);
    const suffix = last && !isNaN(last) ? ' · '+last.toLocaleString(isHindi()?'hi-IN':'en-IN') : '';
    offlineNote.textContent = label('Offline or using saved records. Refresh online for current evidence.', 'ऑफ़लाइन या सहेजे गए रिकॉर्ड दिख रहे हैं। ताज़ा साक्ष्य के लिए ऑनलाइन रीफ़्रेश करें।') + suffix;
  }
  status();
  window.addEventListener('offline', status);
  window.addEventListener('online', status);
  new MutationObserver(status).observe(document.documentElement,{attributes:true,attributeFilter:['lang']});
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.addEventListener('message', event => {
      if (event.data?.type === 'atlas-cached') {
        cached = true; saved = event.data.saved || saved; status();
      }
    });
    navigator.serviceWorker.register('./sw.js').catch(() => {});
  }
})();
