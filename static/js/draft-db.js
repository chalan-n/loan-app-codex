/**
 * LoanApp Draft Database — IndexedDB via Dexie.js
 *
 * Schema (drafts store):
 *   draft_id     : UUID string (primary key)
 *   mo_id        : staff ID decoded from JWT cookie
 *   loan_id      : loan record ID (from cookie / URL param)
 *   step         : step number 1-7
 *   step_data    : JSON string of all form field values
 *   last_updated : Unix timestamp (ms)
 *
 * Requires Dexie CDN to be loaded before this script.
 */
(function () {
  'use strict';

  if (typeof Dexie === 'undefined') {
    console.warn('[DraftDB] Dexie.js not found — offline drafts disabled');
    return;
  }

  var db = new Dexie('LoanAppDB');

  db.version(1).stores({
    drafts: 'draft_id, mo_id, loan_id, step, last_updated'
  });

  db.open().catch(function (err) {
    console.error('[DraftDB] Failed to open database:', err);
  });

  window.LoanDraftDB = db;
})();
