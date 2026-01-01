// popup.js — SAFE + Subject в шапке

const $ = (id) => document.getElementById(id);
const on = (el, evt, fn) => el && el.addEventListener(evt, fn);

const startBtn = $('start');
const stopBtn  = $('stop');
const copyBtn  = $('copy');
const clearBtn = $('clear');
const draftBtn = $('draft');

const hintEl = $('hint');
const subjectHintEl = $('subjectHint');

const micStateEl = $('micState');
const sttStateEl = $('sttState');

const toastEl = $('toast');
const textEl = $('text');

const themeBtn = $('themeBtn');
const themeIcon = $('themeIcon');

const helpBtn = $('help');
const helpModal = $('helpModal');
const helpClose = $('helpClose');
const helpX = $('helpX');
const helpOk = $('helpOk');

let mediaStream = null;
let mediaRecorder = null;
let recognition = null;
let finalText = "";

// ===== UI helpers =====
function setText(el, text) {
  if (!el) return;
  el.textContent = text ?? '';
}

function toastMsg(msg) {
  setText(toastEl, msg || '');
}

function setHint(msg) {
  setText(hintEl, msg || '');
}

function setSubjectHint(text) {
  if (!subjectHintEl) return;
  subjectHintEl.textContent = text || 'Тема появится здесь';
}

// ===== Subject parser =====
function splitSubject(emailText) {
  const text = (emailText || "").trim();

  const m = text.match(/^\s*subject\s*:\s*(.+)\s*$/im);
  if (!m) return { subject: "", body: text };

  const subject = m[1].trim();
  const body = text.replace(/^\s*subject\s*:\s*.*\r?\n/i, "").trim();
  return { subject, body };
}

// ===== Theme =====
async function getTheme() {
  const obj = await chrome.storage.local.get('theme');
  return obj?.theme || null;
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
on(themeBtn, 'click', async () => {
  const current = document.documentElement.getAttribute('data-theme') || 'dark';
  const next = current === 'dark' ? 'light' : 'dark';
  applyTheme(next);
  await setTheme(next);
});

// ===== Help modal =====
function openHelp() { helpModal?.classList.remove('hidden'); }
function closeHelp() { helpModal?.classList.add('hidden'); }
on(helpBtn, 'click', openHelp);
on(helpClose, 'click', closeHelp);
on(helpX, 'click', closeHelp);
on(helpOk, 'click', closeHelp);
document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape' && helpModal && !helpModal.classList.contains('hidden')) closeHelp();
});

// ===== Speech Recognition =====
function initSpeechRecognition() {
  const SR = window.SpeechRecognition || window.webkitSpeechRecognition;
  if (!SR) return null;

  const r = new SR();
  r.lang = 'ru-RU';
  r.continuous = true;
  r.interimResults = true;

  r.onresult = (event) => {
    let interim = "";
    for (let i = event.resultIndex; i < event.results.length; i++) {
      const res = event.results[i];
      const transcript = res[0]?.transcript ?? "";
      if (res.isFinal) finalText += transcript + " ";
      else interim += transcript;
    }
    if (textEl) textEl.value = (finalText + interim).trim();
  };

  r.onerror = (e) => toastMsg(`Ошибка распознавания: ${e.error || 'unknown'}`);
  return r;
}

// ===== Permissions =====
async function isMicGrantedFlag() {
  const obj = await chrome.storage.local.get('micPermissionGranted');
  return Boolean(obj?.micPermissionGranted);
}
async function setMicGrantedFlag(value) {
  await chrome.storage.local.set({ micPermissionGranted: Boolean(value) });
}
async function openPermissionWindow() {
  const url = chrome.runtime.getURL('permission.html');
  await chrome.windows.create({ url, type: 'popup', width: 320, height: 200 });
}
function waitForMicGrantedMessage(timeoutMs = 60000) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      chrome.runtime.onMessage.removeListener(onMsg);
      reject(new Error('timeout'));
    }, timeoutMs);

    function onMsg(msg) {
      if (msg?.type === 'mic_permission_granted') {
        clearTimeout(timer);
        chrome.runtime.onMessage.removeListener(onMsg);
        resolve(true);
      }
    }
    chrome.runtime.onMessage.addListener(onMsg);
  });
}
async function ensureMicPermission() {
  const flag = await isMicGrantedFlag();
  if (flag) return true;

  setHint('Разрешите микрофон в появившемся окне.');
  toastMsg('Запрашиваю доступ к микрофону…');

  await openPermissionWindow();

  try {
    await waitForMicGrantedMessage();
    toastMsg('Доступ к микрофону получен.');
    return true;
  } catch {
    toastMsg('Разрешение не получено.');
    return false;
  }
}

// ===== Server call =====
async function requestBusinessEmailFromServer(rawText) {
  const resp = await fetch('http://localhost:8080/api/draft-email', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text: rawText, lang: 'ru', tone: 'neutral', recipient: '', sender: '' })
  });

  if (!resp.ok) {
    const err = await resp.json().catch(() => ({}));
    throw new Error(err.error || `Server error: ${resp.status}`);
  }

  const data = await resp.json();
  if (!data?.email) throw new Error('Empty email in response');
  return data.email;
}

// ===== Recording =====
async function startRecordingAndRecognition() {
  try {
    toastMsg('');
    setHint('Говорите — текст будет появляться ниже.');

    mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true });
    setText(micStateEl, 'разрешён');

    const mimeTypeCandidates = ["audio/webm;codecs=opus","audio/webm","audio/ogg;codecs=opus"];
    const mimeType = mimeTypeCandidates.find(t => MediaRecorder.isTypeSupported(t)) || "";
    mediaRecorder = new MediaRecorder(mediaStream, mimeType ? { mimeType } : undefined);

    mediaRecorder.onstop = () => {
      setHint('Текст готов. Нажмите «Получить письмо».');
      toastMsg('Запись остановлена.');
    };

    mediaRecorder.start();

    recognition = initSpeechRecognition();
    finalText = "";

    if (!recognition) {
      setText(sttStateEl, 'не поддерживается');
      toastMsg('Распознавание речи недоступно.');
    } else {
      setText(sttStateEl, 'активно');
      recognition.start();
    }

    if (startBtn) startBtn.disabled = true;
    if (stopBtn) stopBtn.disabled = false;
  } catch (e) {
    console.error(e);
    await setMicGrantedFlag(false);
    setText(micStateEl, 'нет доступа');
    toastMsg('Нет доступа к микрофону.');
    setHint('Проверьте разрешения и попробуйте снова.');
  }
}

function stopAll() {
  try {
    if (recognition) { try { recognition.stop(); } catch {} recognition = null; }
    if (mediaRecorder && mediaRecorder.state !== "inactive") mediaRecorder.stop();
    if (mediaStream) { mediaStream.getTracks().forEach(t => t.stop()); mediaStream = null; }

    if (startBtn) startBtn.disabled = false;
    if (stopBtn) stopBtn.disabled = true;
  } catch (e) {
    console.error(e);
    toastMsg('Ошибка при остановке.');
  }
}

// ===== Handlers =====
on(startBtn, 'click', async () => {
  if (mediaRecorder && mediaRecorder.state !== "inactive") return;
  const ok = await ensureMicPermission();
  if (!ok) return;
  await startRecordingAndRecognition();
});
on(stopBtn, 'click', stopAll);

on(clearBtn, 'click', () => {
  if (textEl) textEl.value = "";
  finalText = "";
  setSubjectHint('Тема появится здесь');
  toastMsg('Очищено.');
});

on(copyBtn, 'click', async () => {
  const subject = (subjectHintEl?.textContent || '').trim();
  const body = (textEl?.value || '').trim();

  const hasSubject = subject && !subject.toLowerCase().includes('тема появится');
  const combined = hasSubject ? `Тема: ${subject}\n\n${body}` : body;

  try {
    await navigator.clipboard.writeText(combined);
    toastMsg('Скопировано.');
  } catch {
    toastMsg('Не удалось скопировать.');
  }
});

on(draftBtn, 'click', async () => {
  const raw = (textEl?.value || "").trim();
  if (raw.length < 3) {
    toastMsg('Введите или надиктуйте текст.');
    return;
  }

  if (draftBtn) draftBtn.disabled = true;
  if (startBtn) startBtn.disabled = true;
  if (stopBtn) stopBtn.disabled = true;

  setHint('Формирую деловое письмо…');
  toastMsg('Отправляю на сервер…');

  try {
    const emailRaw = await requestBusinessEmailFromServer(raw);
    const { subject, body } = splitSubject(emailRaw);

    setSubjectHint(subject || 'Без темы');
    if (textEl) textEl.value = body;

    setHint('Готово.');
    toastMsg('Письмо готово.');
  } catch (e) {
    console.error(e);
    setHint('Ошибка.');
    toastMsg('Ошибка: ' + (e.message || 'unknown'));
  } finally {
    if (draftBtn) draftBtn.disabled = false;
    if (startBtn) startBtn.disabled = false;
    if (stopBtn) stopBtn.disabled = true;
  }
});

// ===== Init =====
(function init() {
  initTheme();
  setHint('Нажмите «Начать» и говорите.');
  setSubjectHint('Тема появится здесь');

  setText(micStateEl, 'не проверен');
  const SR = window.SpeechRecognition || window.webkitSpeechRecognition;
  setText(sttStateEl, SR ? 'доступно' : 'не поддерживается');
})();
