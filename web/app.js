// genspark: tiny client-side script for the demo
const $ = (sel) => document.querySelector(sel);
const url = $('#url');
const metaEl = $('#meta');
const textEl = $('#text');
const headingsEl = $('#headings');
const linksEl = $('#links');
const pricesEl = $('#prices');
const summaryEl = $('#summary');

$('#btnFetch').addEventListener('click', async () => {
  if (!url.value.trim()) { alert('Enter URL'); return; }
  const res = await fetch('/api/fetch?url=' + encodeURIComponent(url.value));
  const data = await res.json();
  renderExtract(data);
});

$('#btnAutopilot').addEventListener('click', async () => {
  if (!url.value.trim()) { alert('Enter URL'); return; }
  const res = await fetch('/api/autopilot', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url: url.value })
  });
  const data = await res.json();
  if (data.meta) renderExtract(data.meta);
  if (data.summary) renderSummary(data.summary);
});

$('#btnSummarize').addEventListener('click', async () => {
  const res = await fetch('/api/summarize', {
    method: 'POST',
    body: textEl.value || ''
  });
  const data = await res.json();
  renderSummary(data.summary || []);
});

function renderExtract(d) {
  metaEl.textContent = JSON.stringify({ url: d.url, title: d.title }, null, 2);
  textEl.value = d.text || '';
  renderList(headingsEl, d.headings || []);
  renderList(linksEl, (d.links || []).slice(0, 50));
  renderList(pricesEl, d.prices || []);
}

function renderSummary(lines) {
  summaryEl.innerHTML = '';
  (lines || []).forEach(s => {
    const li = document.createElement('li');
    li.textContent = s;
    summaryEl.appendChild(li);
  });
}

function renderList(el, arr) {
  el.innerHTML = '';
  (arr || []).forEach(v => {
    const li = document.createElement('li');
    li.textContent = v;
    el.appendChild(li);
  });
}
