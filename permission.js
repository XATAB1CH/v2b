// permission.js
const statusEl = document.getElementById('status');
const grantBtn = document.getElementById('grant');

const themeBtn = document.getElementById('themeBtn');
const themeIcon = document.getElementById('themeIcon');

function setStatus(t) { if (statusEl) statusEl.textContent = t; }

async function getTheme() {
  const { theme } = await chrome.storage.local.get('theme');
  return theme || null;
}
async function setTheme(theme) {
  await chrome.storage.local.set({ theme });
}
function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  if (themeIcon) themeIcon.textContent = theme === 'dark' ? '🌙' : '☀️';
}
async function initTheme() {
  const saved = await getTheme();
  if (saved) return applyTheme(saved);

  const prefersDark = window.matchMedia?.('(prefers-color-scheme: dark)')?.matches;
  const initial = prefersDark ? 'dark' : 'light';
  applyTheme(initial);
  await setTheme(initial);
}

themeBtn?.addEventListener('click', async () => {
  const current = document.documentElement.getAttribute('data-theme') || 'dark';
  const next = current === 'dark' ? 'light' : 'dark';
  applyTheme(next);
  await setTheme(next);
});

grantBtn?.addEventListener('click', async () => {
  if (grantBtn) grantBtn.disabled = true;
  setStatus('Открываю запрос доступа…');

  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    stream.getTracks().forEach(t => t.stop());

    await chrome.storage.local.set({ micPermissionGranted: true });
    chrome.runtime.sendMessage({ type: 'mic_permission_granted' });

    setStatus('Разрешение получено.');
    setTimeout(() => window.close(), 600);
  } catch (e) {
    console.error(e);
    setStatus('Доступ не получен. Нажмите «Разрешить» ещё раз.');
    if (grantBtn) grantBtn.disabled = false;
  }
});

initTheme();
