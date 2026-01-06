// popup.js — SAFE + Subject в шапке + Profile/Paywall + Login(OTP) + /api/me gating

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

// ===== New UI: Profile + Paywall + Draft Status =====
const profileBtn = $('profile');
const profileModal = $('profileModal');
const profileClose = $('profileClose');
const profileX = $('profileX');
const profileEmailEl = $('profileEmail');
const profileStatusEl = $('profileStatus');
const profileAttemptsRow = $('profileAttemptsRow');
const profileAttemptsEl = $('profileAttempts');
const profileBuyBtn = $('profileBuy');
const logoutBtn = $('logoutBtn');

const paywallModal = $('paywallModal');
const paywallClose = $('paywallClose');
const paywallX = $('paywallX');
const paywallBuyBtn = $('paywallBuy');
const paywallLaterBtn = $('paywallLater');

const draftStatusEl = $('draftStatus');

// ===== Login Modal (OTP) =====
const loginModal = $('loginModal');
const loginClose = $('loginClose');
const loginX = $('loginX');
const loginCancel = $('loginCancel');
const loginEmail = $('loginEmail');
const loginCode = $('loginCode');
const loginSendCode = $('loginSendCode');
const loginVerify = $('loginVerify');
const loginHint = $('loginHint');

let mediaStream = null;
let mediaRecorder = null;
let recognition = null;
let finalText = "";

// ===== API =====
const API_BASE = 'http://localhost:8080';

async function getToken() {
  const obj = await chrome.storage.local.get('access_token');
  return obj?.access_token || null;
}
async function setToken(token) {
  await chrome.storage.local.set({ access_token: token });
}
async function clearToken() {
  await chrome.storage.local.remove('access_token');
}

async function apiGetMe() {
  const token = await getToken();
  if (!token) throw new Error('no_token');

  const resp = await fetch(`${API_BASE}/api/me`, {
    method: 'GET',
    headers: { 'Authorization': `Bearer ${token}` }
  });

  if (!resp.ok) {
    const err = await resp.json().catch(() => ({}));
    throw new Error(err.error || `me error: ${resp.status}`);
  }
  return await resp.json();
}

async function apiRequestOTP(email) {
  const resp = await fetch(`${API_BASE}/api/auth/otp/request`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email })
  });
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({}));
    throw new Error(err.error || `otp request error: ${resp.status}`);
  }
  return await resp.json().catch(() => ({}));
}

async function apiVerifyOTP(email, code) {
  const resp = await fetch(`${API_BASE}/api/auth/otp/verify`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, code })
  });
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({}));
    throw new Error(err.error || `otp verify error: ${resp.status}`);
  }
  return await resp.json();
}

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

// ===== Login modal =====
function openLogin() {
  loginModal?.classList.remove('hidden');
  setText(loginHint, "");
  // удобнее сразу поставить фокус на email
  setTimeout(() => loginEmail?.focus(), 0);
}
function closeLogin() {
  loginModal?.classList.add('hidden');
  setText(loginHint, "");
}
on(loginClose, 'click', closeLogin);
on(loginX, 'click', closeLogin);
on(loginCancel, 'click', closeLogin);

// ===== Profile + Paywall modals =====
function openProfile() { profileModal?.classList.remove('hidden'); }
function closeProfile() { profileModal?.classList.add('hidden'); }
function openPaywall() { paywallModal?.classList.remove('hidden'); }
function closePaywall() { paywallModal?.classList.add('hidden'); }

on(profileBtn, 'click', async () => {
  const token = await getToken();
  if (!token) {
    openLogin();
    return;
  }
  await refreshMe();
  openProfile();
});
on(profileClose, 'click', closeProfile);
on(profileX, 'click', closeProfile);

on(paywallClose, 'click', closePaywall);
on(paywallX, 'click', closePaywall);
on(paywallLaterBtn, 'click', closePaywall);

on(profileBuyBtn, 'click', () => {
  closeProfile();
  openPaywall();
});

on(paywallBuyBtn, 'click', async () => {
  try {
    paywallBuyBtn.disabled = true;
    toastMsg('Создаю платеж…');

    const token = await getToken();
    if (!token) {
      toastMsg('Нужно войти в аккаунт.');
      closePaywall();
      openLogin();
      return;
    }

    const p = await apiCreatePayment();

    // поддержим разные названия полей на всякий
    const paymentId = p.payment_id || p.id || p.paymentID || p.paymentId;
    const confirmationUrl = p.confirmation_url || p.confirmationUrl || p.url;

    if (!paymentId) throw new Error('server did not return payment_id');

    // DEV: сразу подтверждаем оплату
    toastMsg('Подтверждаю платеж (DEV)…');
    await apiMarkPaymentSucceeded(paymentId);

    // обновляем /me, чтобы paid_until подтянулся
    toastMsg('Активирую подписку…');
    await refreshMe();

    closePaywall();
    toastMsg('Подписка активирована ✅');

    // (опционально) если хочешь — можно открыть confirmationUrl в реальном сценарии:
    // if (confirmationUrl) chrome.tabs.create({ url: confirmationUrl });
  } catch (e) {
    console.error(e);
    toastMsg('Ошибка оплаты: ' + (e.message || 'unknown'));
  } finally {
    paywallBuyBtn.disabled = false;
  }
});

on(logoutBtn, 'click', async () => {
  await clearToken();
  meCache = null;
  closeProfile();
  renderAccessStatus(null);
  toastMsg('Вы вышли из аккаунта.');
});

// Payments
async function apiCreatePayment() {
  const token = await getToken();
  if (!token) throw new Error('no_token');

  const resp = await fetch(`${API_BASE}/api/billing/payments`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({
      // если твой handler требует тело — оставь.
      // если не требует — можно {}.
      plan: 'subscription_30d'
    }),
  });

  if (!resp.ok) {
    const err = await resp.json().catch(() => ({}));
    throw new Error(err.error || `billing error: ${resp.status}`);
  }

  return await resp.json();
}

async function apiMarkPaymentSucceeded(paymentId) {
  const token = await getToken();
  if (!token) throw new Error('no_token');

  const resp = await fetch(`${API_BASE}/api/dev/payments/${encodeURIComponent(paymentId)}/mark-succeeded`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({}),
  });

  if (!resp.ok) {
    const err = await resp.json().catch(() => ({}));
    throw new Error(err.error || `mark-succeeded error: ${resp.status}`);
  }

  return await resp.json().catch(() => ({}));
}


// ESC closes the top-most modal (simple)
document.addEventListener('keydown', (e) => {
  if (e.key !== 'Escape') return;

  if (loginModal && !loginModal.classList.contains('hidden')) return closeLogin();
  if (paywallModal && !paywallModal.classList.contains('hidden')) return closePaywall();
  if (profileModal && !profileModal.classList.contains('hidden')) return closeProfile();
  if (helpModal && !helpModal.classList.contains('hidden')) return closeHelp();
});

// ===== Entitlements render (/me) =====
let meCache = null;

function renderAccessStatus(me) {
  if (!draftStatusEl) return;
  if (!me) { draftStatusEl.textContent = '—'; return; }

  if (me.subscription_active) {
    // если захочешь — можно добавить дату paid_until
    draftStatusEl.textContent = 'Подписка активна';
  } else {
    const n = (me.free_attempts_left ?? 0);
    draftStatusEl.textContent = `Осталось попыток: ${n}`;
  }
}

function renderProfile(me) {
  if (!me) return;

  setText(profileEmailEl, me.email || '—');

  if (me.subscription_active) {
    setText(profileStatusEl, 'Подписка активна');
    profileAttemptsRow?.classList.add('hidden');
    if (profileBuyBtn) profileBuyBtn.textContent = 'Продлить подписку';
  } else {
    setText(profileStatusEl, 'Нет подписки');
    profileAttemptsRow?.classList.remove('hidden');
    setText(profileAttemptsEl, String(me.free_attempts_left ?? 0));
    if (profileBuyBtn) profileBuyBtn.textContent = 'Купить подписку';
  }
}

async function refreshMe() {
  try {
    const me = await apiGetMe();
    meCache = me;
    renderAccessStatus(me);
    renderProfile(me);
    return me;
  } catch (e) {
    renderAccessStatus(null);
    return null;
  }
}

// ===== Server call (draft-email) =====
async function requestBusinessEmailFromServer(rawText) {
  const token = await getToken();
  if (!token) throw new Error('no_token');

  const resp = await fetch(`${API_BASE}/api/draft-email`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
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

// ===== Login logic =====
on(loginSendCode, 'click', async () => {
  const email = (loginEmail?.value || '').trim();
  if (!email || !email.includes('@')) {
    setText(loginHint, 'Введите корректный email.');
    return;
  }

  loginSendCode.disabled = true;
  setText(loginHint, 'Отправляю код…');

  try {
    await apiRequestOTP(email);
    setText(loginHint, 'Код отправлен. Проверьте почту и введите код.');
    setTimeout(() => loginCode?.focus(), 0);
  } catch (e) {
    setText(loginHint, 'Ошибка: ' + (e.message || 'unknown'));
  } finally {
    loginSendCode.disabled = false;
  }
});

on(loginVerify, 'click', async () => {
  const email = (loginEmail?.value || '').trim();
  const code = (loginCode?.value || '').trim();

  if (!email || !email.includes('@')) {
    setText(loginHint, 'Введите корректный email.');
    return;
  }
  if (!code || code.length < 4) {
    setText(loginHint, 'Введите код из письма.');
    return;
  }

  loginVerify.disabled = true;
  setText(loginHint, 'Проверяю код…');

  try {
    const tok = await apiVerifyOTP(email, code);
    if (!tok?.access_token) throw new Error('server did not return access_token');

    await setToken(tok.access_token);

    await refreshMe();
    closeLogin();
    toastMsg('Вы вошли в аккаунт.');
  } catch (e) {
    setText(loginHint, 'Ошибка: ' + (e.message || 'unknown'));
  } finally {
    loginVerify.disabled = false;
  }
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

  // if no token -> open login
  const token = await getToken();
  if (!token) {
    toastMsg('Нужно войти в аккаунт.');
    openLogin();
    return;
  }

  if (draftBtn) draftBtn.disabled = true;
  if (startBtn) startBtn.disabled = true;
  if (stopBtn) stopBtn.disabled = true;

  setHint('Формирую деловое письмо…');
  toastMsg('Проверяю доступ…');

  try {
    const me = await refreshMe();
    if (!me) {
      toastMsg('Нужно войти в аккаунт.');
      openLogin();
      return;
    }

    if (!me.subscription_active && (me.free_attempts_left ?? 0) <= 0) {
      setHint('Закончились бесплатные попытки.');
      toastMsg('Откройте подписку для продолжения.');
      openPaywall();
      return;
    }

    toastMsg('Отправляю на сервер…');

    const emailRaw = await requestBusinessEmailFromServer(raw);
    const { subject, body } = splitSubject(emailRaw);

    setSubjectHint(subject || 'Без темы');
    if (textEl) textEl.value = body;

    setHint('Готово.');
    toastMsg('Письмо готово.');

    await refreshMe();
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

  // initial access status (if token already stored)
  refreshMe();
})();
