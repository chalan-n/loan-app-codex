/**
 * Draft Auto-Save System
 *
 * - Detects current loan_id (cookie / URL param) and step (form action)
 * - Debounced save (300ms) to IndexedDB on every input/change event
 * - On page load: checks for existing draft and shows restore banner
 * - On form submit: deletes draft (server will have the data)
 * - Exposes window.DraftAutoSave.saveNow() for connectivity.js
 *
 * Requires: draft-db.js (window.LoanDraftDB) loaded before this script.
 */
(function () {
  'use strict';

  /* ── Restore banner CSS ──────────────────────────────────────── */
  var style = document.createElement('style');
  style.textContent = [
    '#draftRestoreBanner {',
    '  background: #eff6ff; border: 2px solid #3b82f6; border-radius: 14px;',
    '  padding: 14px 18px; margin-bottom: 18px;',
    '  display: flex; align-items: center; justify-content: space-between;',
    '  gap: 12px; font-family: "Kanit", sans-serif; flex-wrap: wrap;',
    '}',
    '#draftRestoreBanner .draft-banner-info {',
    '  display: flex; align-items: center; gap: 10px;',
    '}',
    '#draftRestoreBanner .draft-banner-info i { color: #3b82f6; font-size: 1.3rem; }',
    '#draftRestoreBanner .draft-banner-title { font-weight: 600; color: #1e3a8a; }',
    '#draftRestoreBanner .draft-banner-sub { font-size: 0.85rem; color: #475569; }',
    '#draftRestoreBanner .draft-banner-btns { display: flex; gap: 8px; flex-shrink: 0; }',
    '.draft-btn-yes {',
    '  background: #1e3a8a; color: white; border: none;',
    '  padding: 8px 18px; border-radius: 8px;',
    '  font-family: "Kanit", sans-serif; font-size: 0.9rem;',
    '  cursor: pointer; font-weight: 600;',
    '}',
    '.draft-btn-no {',
    '  background: #e5e7eb; color: #374151; border: none;',
    '  padding: 8px 18px; border-radius: 8px;',
    '  font-family: "Kanit", sans-serif; font-size: 0.9rem; cursor: pointer;',
    '}'
  ].join('\n');
  document.head.appendChild(style);

  /* ── Helpers ─────────────────────────────────────────────────── */
  function getLoanId() {
    var params = new URLSearchParams(window.location.search);
    if (params.get('id')) return params.get('id');
    var match = document.cookie.match(/(?:^|;\s*)loan_id=([^;]*)/);
    return match ? decodeURIComponent(match[1]) : null;
  }

  function getMoId() {
    try {
      var match = document.cookie.match(/(?:^|;\s*)token=([^;]*)/);
      if (!match) return 'unknown';
      var payload = JSON.parse(atob(match[1].split('.')[1]));
      return payload.username || payload.sub || 'unknown';
    } catch (e) {
      return 'unknown';
    }
  }

  function getStep() {
    var form = document.querySelector('form[id]');
    if (!form) return null;
    var m = (form.getAttribute('action') || '').match(/\/step(\d)/);
    return m ? parseInt(m[1], 10) : null;
  }

  function getDraftId(loanId, step) {
    var key = 'draft_uuid:' + loanId + ':' + step;
    var id = localStorage.getItem(key);
    if (!id) {
      id = typeof crypto !== 'undefined' && crypto.randomUUID
        ? crypto.randomUUID()
        : Date.now().toString(36) + Math.random().toString(36).slice(2);
      localStorage.setItem(key, id);
    }
    return id;
  }

  function formatLastUpdated(ts) {
    var d = new Date(ts);
    var pad = function (n) { return n < 10 ? '0' + n : '' + n; };
    return pad(d.getDate()) + '/' + pad(d.getMonth() + 1) + '/' + (d.getFullYear() + 543) +
      ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes());
  }

  /* ── Form data collection ────────────────────────────────────── */
  function collectFormData(form) {
    var data = {};
    var els = form.querySelectorAll(
      'input:not([type=file]):not([type=submit]):not([type=button]), select, textarea'
    );
    els.forEach(function (el) {
      if (!el.name) return;
      if (el.type === 'radio' || el.type === 'checkbox') {
        if (el.checked) data[el.name] = el.value;
      } else {
        data[el.name] = el.value;
      }
    });
    return data;
  }

  function restoreFormData(form, data) {
    Object.keys(data).forEach(function (name) {
      var value = data[name];
      var els = form.querySelectorAll('[name="' + name + '"]');
      els.forEach(function (el) {
        if (el.type === 'radio') {
          el.checked = (el.value === value);
          if (el.checked && typeof el.onclick === 'function') el.onclick();
        } else if (el.type === 'checkbox') {
          el.checked = (el.value === value);
        } else {
          el.value = value;
          el.dispatchEvent(new Event('change', { bubbles: true }));
        }
      });
    });
  }

  /* ── Restore banner ──────────────────────────────────────────── */
  function showRestoreBanner(form, draft, onYes, onNo) {
    var banner = document.createElement('div');
    banner.id = 'draftRestoreBanner';
    var timeLabel = draft.last_updated ? formatLastUpdated(draft.last_updated) : '';
    banner.innerHTML =
      '<div class="draft-banner-info">' +
        '<i class="fas fa-history"></i>' +
        '<div>' +
          '<div class="draft-banner-title">พบข้อมูลที่กรอกไว้ก่อนหน้านี้</div>' +
          '<div class="draft-banner-sub">บันทึกล่าสุด: ' + timeLabel + ' — ต้องการกู้คืนข้อมูล Draft หรือไม่?</div>' +
        '</div>' +
      '</div>' +
      '<div class="draft-banner-btns">' +
        '<button class="draft-btn-yes" id="draftRestoreYes"><i class="fas fa-redo"></i> กู้คืน</button>' +
        '<button class="draft-btn-no"  id="draftRestoreNo">ไม่ต้องการ</button>' +
      '</div>';

    var anchor = form.closest('.form-body') || form.parentElement;
    anchor.insertBefore(banner, form);

    document.getElementById('draftRestoreYes').addEventListener('click', function () {
      onYes();
      banner.remove();
    });
    document.getElementById('draftRestoreNo').addEventListener('click', function () {
      onNo();
      banner.remove();
    });
  }

  /* ── Debounce ────────────────────────────────────────────────── */
  var saveTimer = null;
  function debounce(fn, delay) {
    clearTimeout(saveTimer);
    saveTimer = setTimeout(fn, delay);
  }

  /* ── Core init ───────────────────────────────────────────────── */
  function init() {
    if (!window.LoanDraftDB) return;

    var loanId = getLoanId();
    var step   = getStep();
    if (!loanId || !step) return;

    var form = document.querySelector('form#step' + step + 'Form');
    if (!form) return;

    var moId    = getMoId();
    var draftId = getDraftId(loanId, step);

    /* ── Try restore on load ── */
    window.LoanDraftDB.drafts.get(draftId).then(function (existing) {
      if (!existing || !existing.step_data) return;
      var data = {};
      try { data = JSON.parse(existing.step_data); } catch (e) { return; }
      if (!Object.keys(data).length) return;

      showRestoreBanner(
        form,
        existing,
        function () { restoreFormData(form, data); },
        function () {
          window.LoanDraftDB.drafts.delete(draftId);
          localStorage.removeItem('draft_uuid:' + loanId + ':' + step);
        }
      );
    }).catch(function (e) {
      console.warn('[Draft] restore check failed', e);
    });

    /* ── Save function ── */
    function doSave() {
      if (!window.LoanDraftDB) return;
      window.LoanDraftDB.drafts.put({
        draft_id:     draftId,
        mo_id:        moId,
        loan_id:      loanId,
        step:         step,
        step_data:    JSON.stringify(collectFormData(form)),
        last_updated: Date.now()
      }).catch(function (e) {
        console.warn('[Draft] save failed', e);
      });
    }

    /* ── Auto-save on every input/change ── */
    form.addEventListener('input',  function () { debounce(doSave, 300); });
    form.addEventListener('change', function () { debounce(doSave, 300); });

    /* ── Clear draft on successful form submit ── */
    form.addEventListener('submit', function () {
      window.LoanDraftDB.drafts.delete(draftId).catch(function () {});
      localStorage.removeItem('draft_uuid:' + loanId + ':' + step);
    });

    /* ── Public API ── */
    window.DraftAutoSave = {
      saveNow: doSave,
      draftId: draftId,
      loanId:  loanId,
      step:    step
    };
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
