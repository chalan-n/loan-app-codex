/**
 * Connectivity Status Bar
 *
 * - Injects a slim fixed bar showing Online/Offline state
 * - When offline: disables submit button and turns it into "Save Draft"
 * - When back online: restores submit button and shows brief green banner
 * - Exposes window.ConnectivityBar.isOnline() for other modules
 */
(function () {
  'use strict';

  /* ── Injected CSS ────────────────────────────────────────────── */
  var style = document.createElement('style');
  style.textContent = [
    '#conn-bar {',
    '  position: fixed; top: 0; left: 0; right: 0; height: 36px;',
    '  display: flex; align-items: center; justify-content: center; gap: 8px;',
    '  font-family: "Kanit", sans-serif; font-size: 0.85rem; font-weight: 600;',
    '  z-index: 99999; transform: translateY(-100%);',
    '  transition: transform 0.35s ease; pointer-events: none;',
    '}',
    '#conn-bar.show  { transform: translateY(0); }',
    '#conn-bar.online  { background: #d1fae5; color: #065f46; }',
    '#conn-bar.offline { background: #fef3c7; color: #92400e; }',
    '#conn-bar .conn-dot {',
    '  width: 8px; height: 8px; border-radius: 50%;',
    '  display: inline-block; flex-shrink: 0;',
    '}',
    '#conn-bar.online  .conn-dot { background: #10b981; }',
    '#conn-bar.offline .conn-dot { background: #f59e0b; animation: connPulse 1s infinite; }',
    '@keyframes connPulse { 0%,100%{opacity:1} 50%{opacity:0.3} }',
    '#offlineToast {',
    '  position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%);',
    '  background: #1e3a8a; color: white; padding: 14px 24px; border-radius: 12px;',
    '  font-family: "Kanit", sans-serif; font-size: 0.95rem; font-weight: 600;',
    '  box-shadow: 0 8px 24px rgba(30,58,138,0.3); z-index: 99998;',
    '  display: flex; align-items: center; gap: 10px;',
    '  opacity: 0; transition: opacity 0.4s ease; pointer-events: none;',
    '}',
    '#offlineToast.show { opacity: 1; }'
  ].join('\n');
  document.head.appendChild(style);

  /* ── Status bar element ──────────────────────────────────────── */
  var bar = document.createElement('div');
  bar.id = 'conn-bar';
  bar.innerHTML = '<span class="conn-dot"></span><span id="conn-text"></span>';
  document.body.appendChild(bar);

  /* ── Offline toast element ───────────────────────────────────── */
  var toast = document.createElement('div');
  toast.id = 'offlineToast';
  toast.innerHTML = '<i class="fas fa-check-circle" style="color:#34d399;font-size:1.2rem;"></i>' +
    ' บันทึก Draft เรียบร้อยแล้ว — จะส่งอัตโนมัติเมื่อมีอินเทอร์เน็ต';
  document.body.appendChild(toast);

  var hideBarTimer = null;
  var hideToastTimer = null;

  function updateBar(online) {
    var text = document.getElementById('conn-text');
    clearTimeout(hideBarTimer);

    if (online) {
      bar.className = 'online show';
      text.textContent = 'ออนไลน์ — ข้อมูลจะส่งถึงเซิร์ฟเวอร์ทันที';
      hideBarTimer = setTimeout(function () { bar.classList.remove('show'); }, 2500);
    } else {
      bar.className = 'offline show';
      text.textContent = 'ออฟไลน์ — ข้อมูลจะถูกบันทึกเป็น Draft อัตโนมัติ';
    }

    guardSubmitButton(online);
  }

  function guardSubmitButton(online) {
    var btn = document.querySelector('button[type="submit"].btn-next');
    if (!btn) return;

    if (!online) {
      if (btn.getAttribute('data-offline-guard') === 'true') return; // already guarded
      btn.setAttribute('data-offline-guard', 'true');
      btn.setAttribute('data-original-html', btn.innerHTML);
      btn.innerHTML = '<i class="fas fa-save"></i> บันทึก Draft (Offline)';
      btn.type = 'button';
      btn.addEventListener('click', handleOfflineClick);
    } else {
      if (btn.getAttribute('data-offline-guard') !== 'true') return; // not guarded
      btn.removeAttribute('data-offline-guard');
      btn.type = 'submit';
      btn.removeEventListener('click', handleOfflineClick);
      var orig = btn.getAttribute('data-original-html');
      if (orig) btn.innerHTML = orig;
    }
  }

  function handleOfflineClick() {
    if (window.DraftAutoSave) window.DraftAutoSave.saveNow();
    showOfflineToast();
  }

  function showOfflineToast() {
    clearTimeout(hideToastTimer);
    toast.classList.add('show');
    hideToastTimer = setTimeout(function () { toast.classList.remove('show'); }, 3500);
  }

  /* ── Event wiring ────────────────────────────────────────────── */
  window.addEventListener('online',  function () { updateBar(true); });
  window.addEventListener('offline', function () { updateBar(false); });

  /* Initial state after DOM ready */
  function onReady() { updateBar(navigator.onLine); }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', onReady);
  } else {
    onReady();
  }

  /* ── Public API ──────────────────────────────────────────────── */
  window.ConnectivityBar = {
    isOnline: function () { return navigator.onLine; }
  };
})();
