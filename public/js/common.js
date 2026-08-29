/* ============================================================
   Huz CCTV — Shared front-end helpers
   - Header shown on every page (includes the language switcher
     and the Sign out button)
   - Auth check / guard
   - fetch API wrapper + toast + utilities
   ============================================================ */
(function () {
  'use strict';

  var LOGIN_PATH = '/login.html';
  var STORAGE_KEY = 'huz_authed';

  var NAV_ITEMS = [
    { href: '/index.html', tKey: 'nav.dashboard', icon: 'grid' },
    { href: '/devices.html', tKey: 'nav.network', icon: 'network' },
    { href: '/camera.html', tKey: 'nav.camera', icon: 'video' },
  ];

  var ICONS = {
    grid:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/></svg>',
    network:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="9" width="6" height="6" rx="1.5"/><rect x="16" y="9" width="6" height="6" rx="1.5"/><rect x="9" y="2" width="6" height="6" rx="1.5"/><rect x="9" y="16" width="6" height="6" rx="1.5"/><path d="M5 15v2a2 2 0 0 0 2 2h4"/><path d="M19 15v2a2 2 0 0 1-2 2h-4"/></svg>',
    video:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="6" width="14" height="12" rx="2"/><path d="m22 8-6 4 6 4V8Z"/></svg>',
    logout:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><path d="m16 17 5-5-5-5"/><path d="M21 12H9"/></svg>',
    refresh:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-2.64-6.36"/><path d="M21 3v6h-6"/></svg>',
    camera:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 4h-5L7 7H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-3l-2.5-3Z"/><circle cx="12" cy="13" r="3"/></svg>',
    server:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="7" rx="2"/><rect x="2" y="14" width="20" height="7" rx="2"/><path d="M6 6.5h.01"/><path d="M6 17.5h.01"/></svg>',
    shield:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-3.5 8-10V5l-8-3-8 3v7c0 6.5 8 10 8 10Z"/><path d="m9 12 2 2 4-4"/></svg>',
    user:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-4 3.5-6 8-6s8 2 8 6"/></svg>',
  };

  /* ---------- Utilities ---------- */
  function t(key, params) {
    return window.I18N ? window.I18N.t(key, params) : key;
  }

  function esc(value) {
    return String(value === null || value === undefined ? '' : value).replace(
      /[&<>"']/g,
      function (c) {
        return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
      }
    );
  }

  function icon(name) {
    return ICONS[name] || '';
  }

  /* ---------- fetch wrapper ---------- */
  async function api(path, options) {
    var opts = Object.assign({ headers: {} }, options || {});
    if (opts.body && typeof opts.body !== 'string') {
      opts.headers['Content-Type'] = 'application/json';
      opts.body = JSON.stringify(opts.body);
    }
    var resp = await fetch(path, opts);
    var data = null;
    try {
      data = await resp.json();
    } catch (_e) {
      /* body is not JSON */
    }
    if (!resp.ok) {
      var code = data && data.code;
      var localized = code ? t('error.' + code) : null;
      var msg =
        localized && localized !== 'error.' + code
          ? localized
          : (data && data.message) || t('error.request_failed');
      var err = new Error(msg);
      err.status = resp.status;
      err.code = code;
      err.data = data;
      if (resp.status === 401 && location.pathname !== LOGIN_PATH) {
        redirectToLogin();
      }
      throw err;
    }
    return data;
  }

  /* ---------- Auth ---------- */
  async function getMe() {
    try {
      return await api('/api/auth/me');
    } catch (_e) {
      return null;
    }
  }

  function redirectToLogin() {
    var next = location.pathname + location.search;
    if (next === '/index.html' || next === '/' || next === '/login.html') {
      next = '';
    }
    location.href = LOGIN_PATH + (next ? '?next=' + encodeURIComponent(next) : '');
  }

  function loginRedirect() {
    var next = new URLSearchParams(location.search).get('next');
    if (next && next.startsWith('/') && !next.startsWith('//')) {
      return next;
    }
    return '/index.html';
  }

  /* ---------- Header ---------- */
  function currentNav() {
    var path = location.pathname;
    for (var i = 0; i < NAV_ITEMS.length; i++) {
      if (path === NAV_ITEMS[i].href) {
        return NAV_ITEMS[i].href;
      }
    }
    return '';
  }

  function languageSelect() {
    var current = window.I18N ? window.I18N.lang() : 'en';
    return (
      '<select class="lang-select" id="langSelect" aria-label="' +
      t('header.language') +
      '">' +
      '<option value="en"' +
      (current === 'en' ? ' selected' : '') +
      '>English</option>' +
      '<option value="vi"' +
      (current === 'vi' ? ' selected' : '') +
      '>Tiếng Việt</option>' +
      '</select>'
    );
  }

  function renderHeader(user) {
    var mount = document.getElementById('app-header');
    if (!mount) return;

    var active = currentNav();
    var navHtml = NAV_ITEMS.map(function (item) {
      return (
        '<a class="nav-link' +
        (item.href === active ? ' active' : '') +
        '" href="' +
        item.href +
        '">' +
        icon(item.icon) +
        '<span>' +
        t(item.tKey) +
        '</span></a>'
      );
    }).join('');

    var actionsHtml;
    if (user) {
      actionsHtml =
        '<div class="header-user">' +
        '<span class="avatar">' +
        esc((user.user && user.user.username) || 'U').charAt(0).toUpperCase() +
        '</span>' +
        '<span class="username">' +
        esc((user.user && user.user.username) || '') +
        '</span></div>' +
        '<button class="btn btn-danger btn-sm" id="logoutBtn" type="button">' +
        icon('logout') +
        '<span>' +
        t('header.logout') +
        '</span></button>';
    } else {
      actionsHtml =
        '<a class="btn btn-primary btn-sm" href="/login.html">' +
        t('header.login') +
        '</a>';
    }

    mount.innerHTML =
      '<div class="header-inner">' +
      '<a class="brand" href="/index.html">' +
      '<span class="brand-mark"></span>' +
      '<span class="brand-text">' +
      '<span class="brand-name">Huz CCTV</span>' +
      '<span class="brand-sub">' +
      t('header.brand.sub') +
      '</span></span></a>' +
      '<nav class="header-nav">' +
      navHtml +
      '</nav>' +
      '<div class="header-actions">' +
      languageSelect() +
      actionsHtml +
      '</div></div>';

    var langSelect = document.getElementById('langSelect');
    if (langSelect) {
      langSelect.addEventListener('change', function () {
        if (window.I18N) window.I18N.setLang(langSelect.value);
      });
    }

    var logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
      logoutBtn.addEventListener('click', async function () {
        logoutBtn.disabled = true;
        logoutBtn.innerHTML = '<span>' + t('header.loggingOut') + '</span>';
        try {
          await api('/api/auth/logout', { method: 'POST' });
        } catch (_e) {
          /* redirect to login even on error */
        }
        localStorage.removeItem(STORAGE_KEY);
        location.href = LOGIN_PATH;
      });
    }
  }

  /* ---------- Toast ---------- */
  function toast(message, type) {
    var root = document.getElementById('toast-root');
    if (!root) {
      root = document.createElement('div');
      root.id = 'toast-root';
      document.body.appendChild(root);
    }
    var el = document.createElement('div');
    el.className = 'toast toast-' + (type || 'info');
    el.textContent = message;
    root.appendChild(el);
    requestAnimationFrame(function () {
      requestAnimationFrame(function () {
        el.classList.add('show');
      });
    });
    setTimeout(function () {
      el.classList.remove('show');
      setTimeout(function () {
        el.remove();
      }, 300);
    }, 3400);
  }

  /* ---------- Uptime formatting ---------- */
  function formatUptime(seconds) {
    if (typeof seconds !== 'number' || isNaN(seconds) || seconds < 0) return '–';
    seconds = Math.floor(seconds);
    var d = Math.floor(seconds / 86400);
    var h = Math.floor((seconds % 86400) / 3600);
    var m = Math.floor((seconds % 3600) / 60);
    var s = seconds % 60;
    var parts = [];
    if (d) parts.push(t('dash.uptime.day', { n: d }));
    if (h) parts.push(t('dash.uptime.hour', { n: h }));
    if (m) parts.push(t('dash.uptime.min', { n: m }));
    if (!parts.length && s) parts.push(t('dash.uptime.sec', { n: s }));
    return parts.length ? parts.join(' ') : t('dash.uptime.sec', { n: 0 });
  }

  /* ---------- Page init ---------- */
  var lastUser = null;

  async function init(options) {
    var opts = options || {};
    var requireAuth = opts.requireAuth !== false;
    var user = await getMe();

    if (user) {
      localStorage.setItem(STORAGE_KEY, '1');
    } else {
      localStorage.removeItem(STORAGE_KEY);
      if (requireAuth) {
        redirectToLogin();
        return null;
      }
    }

    lastUser = user;
    renderHeader(user);
    document.body.classList.toggle('authed', !!user);
    if (typeof opts.onReady === 'function') {
      opts.onReady(user);
    }
    return user;
  }

  /* Re-render the header when the language changes. */
  if (document.addEventListener) {
    document.addEventListener('huz:langchange', function () {
      renderHeader(lastUser);
    });
  }

  window.HuzApp = {
    init: init,
    api: api,
    getMe: getMe,
    toast: toast,
    esc: esc,
    icon: icon,
    t: function (key, params) {
      return t(key, params);
    },
    lang: function () {
      return window.I18N ? window.I18N.lang() : 'en';
    },
    locale: function () {
      return window.I18N ? window.I18N.locale() : 'en-US';
    },
    setLang: function (lang) {
      if (window.I18N) window.I18N.setLang(lang);
    },
    formatUptime: formatUptime,
    redirectToLogin: redirectToLogin,
    loginRedirect: loginRedirect,
  };
})();

