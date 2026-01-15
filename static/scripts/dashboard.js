const state = {
  able: null,
  items: [],
  isUploading: false,
  observer: null,
  loadObserver: null,
  scrollHooked: false,
  uploads: [],
  albums: [],
  view: "recent",
  groups: [],
  renderIndex: 0,
};

function qs(id) {
  return document.getElementById(id);
}

async function apiJSON(url) {
  const r = await fetch(url, { credentials: "include" });
  if (!r.ok) return null;
  return r.json();
}

async function apiUpload(buffer) {
  const r = await fetch("/api/image", {
    method: "POST",
    body: buffer,
    headers: { "Content-Type": "application/octet-stream" },
    credentials: "include",
  });
  if (!r.ok) return null;
  return r.json();
}

async function apiAlbums() {
  const r = await fetch("/api/albums", { credentials: "include" });
  if (!r.ok) return [];
  return r.json();
}

async function apiAlbumCreate(name) {
  const r = await fetch("/api/albums?name=" + encodeURIComponent(name), {
    method: "POST",
    credentials: "include",
  });
  if (!r.ok) return [];
  return r.json();
}

async function apiAlbumImages(name) {
  const r = await fetch("/api/albums/" + encodeURIComponent(name), { credentials: "include" });
  if (!r.ok) return [];
  return r.json();
}

async function apiAlbumAdd(name, id) {
  const r = await fetch("/api/albums/" + encodeURIComponent(name) + "/add?id=" + encodeURIComponent(id), {
    method: "POST",
    credentials: "include",
  });
  return r.ok;
}

async function apiBinList() {
  const r = await fetch("/api/bin", { credentials: "include" });
  if (!r.ok) return [];
  return r.json();
}

async function apiBinRestore(id) {
  const r = await fetch("/api/bin/restore/" + encodeURIComponent(id), { method: "POST", credentials: "include" });
  return r.ok;
}

async function apiBinDelete(id) {
  const r = await fetch("/api/bin/" + encodeURIComponent(id), { method: "DELETE", credentials: "include" });
  return r.ok;
}

async function apiDelete(id) {
  const r = await fetch("/api/image/" + encodeURIComponent(id), {
    method: "DELETE",
    credentials: "include",
  });
  return r.ok;
}

function setStatus(msg) {
  const el = qs("status");
  if (el) el.textContent = msg || "";
}

function ensureObserver() {
  if (state.observer) return;
  state.observer = new IntersectionObserver(entries => {
    for (const e of entries) {
      if (e.isIntersecting) {
        const img = e.target;
        const src = img.dataset.src;
        if (src && !img.src) {
          img.src = src;
        }
        state.observer.unobserve(img);
      }
    }
  }, { rootMargin: "200px" });
}

function monthName(m) {
  return ["January","February","March","April","May","June","July","August","September","October","November","December"][m] || "";
}

function groupByMonth(items) {
  const groups = new Map();
  for (const it of items || []) {
    const ts = Number(it.timestamp || 0);
    const d = new Date(ts);
    const y = d.getFullYear();
    const m = d.getMonth();
    const key = y + "-" + m;
    if (!groups.has(key)) groups.set(key, { year: y, month: m, items: [] });
    groups.get(key).items.push(it);
  }
  const out = Array.from(groups.values());
  out.sort((a,b) => (b.year - a.year) || (b.month - a.month));
  for (const g of out) {
    g.items.sort((a,b) => Number(b.timestamp||0) - Number(a.timestamp||0));
    g.title = monthName(g.month) + " " + g.year;
  }
  return out;
}

function render(items) {
  state.items = (items || []).slice().sort((a,b) => Number(b.timestamp||0) - Number(a.timestamp||0));
  state.groups = groupByMonth(state.items);
  state.renderIndex = 0;
  const grid = qs("grid");
  if (!grid) return;
  grid.innerHTML = "";
  ensureObserver();
  // create or reuse sentinel
  let sentinel = document.getElementById("loadSentinel");
  if (!sentinel) {
    sentinel = document.createElement("div");
    sentinel.id = "loadSentinel";
    sentinel.style.height = "1px";
  }
  grid.appendChild(sentinel);
  ensureLoadObserver();
  renderMore();
}

const GROUP_BATCH = 6;

function renderMore() {
  const grid = qs("grid");
  if (!grid || !state.groups) return;
  const end = Math.min(state.renderIndex + GROUP_BATCH, state.groups.length);
  for (let gi = state.renderIndex; gi < end; gi++) {
    const g = state.groups[gi];
    const sec = document.createElement("section");
    sec.className = "section";
    const title = document.createElement("div");
    title.className = "section-title";
    title.textContent = g.title;
    const sg = document.createElement("div");
    sg.className = "section-grid";
    for (const it of g.items) {
      const wrap = document.createElement("div");
      wrap.className = "tile";
      if (it.width && it.height) {
        wrap.style.aspectRatio = it.width + " / " + it.height;
      }
      const img = document.createElement("img");
      img.loading = "lazy";
      img.style.width = "100%";
      img.style.height = "100%";
      img.style.objectFit = "cover";
      const base = state.view === "bin" ? "/api/bin/preview/" : "/api/image/preview/";
      img.dataset.src = base + encodeURIComponent(it.id);
      if (!state.isUploading) {
        state.observer.observe(img);
      }
      wrap.appendChild(img);
      wrap.onclick = () => {
        const url = state.view === "bin" ? ("/api/bin/preview/" + encodeURIComponent(it.id)) : ("/api/image/" + encodeURIComponent(it.id));
        window.open(url, "_blank");
      };
      wrap.oncontextmenu = e => openContextMenu(e, it.id);
      sg.appendChild(wrap);
    }
    sec.appendChild(title);
    sec.appendChild(sg);
    if (gi === end - 1) {
      const next = document.createElement("div");
      next.style.height = "1px";
      next.id = "groupSentinel-" + gi;
      sec.appendChild(next);
      if (state.loadObserver) state.loadObserver.observe(next);
    }
    const sentinel = document.getElementById("loadSentinel");
    grid.insertBefore(sec, sentinel);
  }
  state.renderIndex = end;
}

function ensureLoadObserver() {
  const rootEl = document.querySelector(".content") || null;
  if (!state.loadObserver) {
    state.loadObserver = new IntersectionObserver(entries => {
      for (const e of entries) {
        if (e.isIntersecting) {
          if (state.renderIndex < state.groups.length) {
            renderMore();
          }
        }
      }
    }, { root: rootEl, rootMargin: "800px" });
  }
  const sentinel = document.getElementById("loadSentinel");
  if (sentinel && state.loadObserver) state.loadObserver.observe(sentinel);
  if (!state.scrollHooked) {
    const content = document.querySelector(".content");
    const target = content || window;
    target.addEventListener("scroll", checkNearBottom, { passive: true });
    state.scrollHooked = true;
  }
}

function checkNearBottom() {
  const grid = qs("grid");
  if (!grid || !state.groups) return;
  const content = document.querySelector(".content");
  if (content) {
    const scrolled = content.scrollTop + content.clientHeight;
    const total = content.scrollHeight;
    if (total - scrolled < 1200 && state.renderIndex < state.groups.length) {
      renderMore();
    }
  } else {
    const scrolled = window.scrollY + window.innerHeight;
    const total = document.documentElement.scrollHeight;
    if (total - scrolled < 1200 && state.renderIndex < state.groups.length) {
      renderMore();
    }
  }
}

async function loadRecent() {
  setStatus("Loading recent...");
  const items = await apiJSON("/api/images/recent");
  setStatus("");
  if (items) {
    state.view = "recent";
    render(items);
  }
}

async function loadCurrentYear() {
  const year = new Date().getFullYear();
  setStatus("Loading " + year + "...");
  const items = await apiJSON("/api/images/" + encodeURIComponent(year));
  setStatus("");
  if (items) {
    state.view = "year:" + year;
    render(items);
  }
}

async function loadAll() {
  setStatus("Loading photos...");
  const items = await apiJSON("/api/search?q=");
  setStatus("");
  if (items) {
    state.view = "all";
    render(items);
  }
}

async function loadBin() {
  setStatus("Loading bin...");
  const items = await apiBinList();
  setStatus("");
  if (items) {
    state.view = "bin";
    render(items);
  }
}

async function doSearch() {
  const q = qs("searchInput").value.trim();
  setStatus("Searching...");
  const items = await apiJSON("/api/search?q=" + encodeURIComponent(q));
  setStatus("");
  if (items) render(items);
}

async function doUploadAll() {
  const list = qs("fileInput").files;
  const files = Array.from(list || []).filter(f => f && f.type && f.type.startsWith("image/"));
  if (!files.length) return;
  const max = state.able && state.able.maxUpload ? (parseInt(state.able.maxUpload, 10) || 0) : 0;
  const total = files.length;
  let done = 0;
  let okCount = 0;
  let failCount = 0;
  const CONCURRENCY = 8;
  let idx = 0;
  state.isUploading = true;
  setStatus("Uploading 0/" + total);
  state.uploads = files.map(f => ({ name: f.name, status: "pending" }));
  renderUploads();
  showUploadsPanel(true);
  async function worker() {
    while (idx < total) {
      const i = idx++;
      const f = files[i];
      if (max && f.size > max) {
        failCount++;
        done++;
        setStatus("Uploading " + done + "/" + total + " • ok " + okCount + " • fail " + failCount);
        continue;
      }
      try {
        const buf = await f.arrayBuffer();
        state.uploads[i].status = "uploading";
        renderUploads();
        const res = await apiUpload(buf);
        if (res && res.ok) {
          okCount++;
          state.uploads[i].status = "ok";
        } else {
          failCount++;
          state.uploads[i].status = "fail";
        }
      } catch (e) {
        failCount++;
        state.uploads[i].status = "fail";
      }
      done++;
      setStatus("Uploading " + done + "/" + total + " • ok " + okCount + " • fail " + failCount);
      renderUploads();
    }
  }
  const workers = [];
  for (let i = 0; i < Math.min(CONCURRENCY, total); i++) {
    workers.push(worker());
  }
  await Promise.all(workers);
  state.isUploading = false;
  setStatus("Uploaded " + okCount + "/" + total + " • failed " + failCount);
  showUploadsPanel(false);
  await loadRecent();
}

function showUploadsPanel(show) {
  const panel = qs("uploadsPanel");
  if (!panel) return;
  panel.style.display = show ? "" : "none";
}

function renderUploads() {
  const list = qs("uploadsList");
  if (!list) return;
  list.innerHTML = "";
  for (const u of state.uploads) {
    const chip = document.createElement("div");
    chip.className = "upload-chip";
    chip.textContent = u.name + " • " + u.status;
    list.appendChild(chip);
  }
}

async function loadAlbumsSidebar() {
  const names = await apiAlbums();
  state.albums = names || [];
  const list = qs("albumList");
  if (!list) return;
  list.innerHTML = "";
  for (const name of state.albums) {
    const btn = document.createElement("button");
    btn.textContent = name;
    btn.onclick = async () => {
      state.view = "album:" + name;
      setStatus("Loading album...");
      const items = await apiAlbumImages(name);
      setStatus("");
      render(items || []);
    };
    list.appendChild(btn);
  }
}

function openContextMenu(e, id) {
  e.preventDefault();
  const menu = qs("contextMenu");
  if (!menu) return;
  menu.innerHTML = "";
  const actions = [];
  actions.push({ text: "Add to album…", fn: async () => {
    const name = prompt("Album name");
    if (name && name.trim()) {
      await apiAlbumCreate(name.trim());
      await apiAlbumAdd(name.trim(), id);
      await loadAlbumsSidebar();
    }
  }});
  if (state.view && state.view.startsWith("album:")) {
    const name = state.view.slice(6);
    actions.push({ text: "Remove from album", fn: async () => {
      await apiAlbumRemove(name, id);
      setStatus("Loading album...");
      const items = await apiAlbumImages(name);
      setStatus("");
      render(items || []);
    }});
  }
  if (state.view !== "bin") {
    actions.push({ text: "Move to bin", fn: async () => {
      setStatus("Moving to bin...");
      const ok = await apiDelete(id);
      setStatus("");
      if (!ok) return;
      await refreshAfterDelete();
    }});
  }
  for (const a of actions) {
    const btn = document.createElement("button");
    btn.textContent = a.text;
    btn.onclick = async () => {
      hideContextMenu();
      await a.fn();
    };
    menu.appendChild(btn);
  }
  menu.style.left = e.clientX + "px";
  menu.style.top = e.clientY + "px";
  menu.style.display = "";
  document.addEventListener("click", onDocClickOnce, { once: true });
  window.addEventListener("scroll", hideContextMenu, { once: true });
}

function hideContextMenu() {
  const menu = qs("contextMenu");
  if (menu) menu.style.display = "none";
}

function onDocClickOnce() {
  hideContextMenu();
}

async function apiAlbumRemove(name, id) {
  const r = await fetch("/api/albums/" + encodeURIComponent(name) + "/remove?id=" + encodeURIComponent(id), {
    method: "POST",
    credentials: "include",
  });
  return r.ok;
}

async function refreshAfterDelete() {
  if (state.view === "recent") return loadRecent();
  if (state.view === "all") return loadAll();
  if (state.view && state.view.startsWith("album:")) {
    const name = state.view.slice(6);
    setStatus("Loading album...");
    const items = await apiAlbumImages(name);
    setStatus("");
    return render(items || []);
  }
  if (state.view && state.view.startsWith("year:")) {
    const y = parseInt(state.view.slice(5), 10);
    setStatus("Loading " + y + "...");
    const items = await apiJSON("/api/images/" + encodeURIComponent(y));
    setStatus("");
    return render(items || []);
  }
  if (state.view === "bin") return loadBin();
  return loadRecent();
}

window.addEventListener("load", async () => {
  const able = await apiJSON("/api/able");
  state.able = able;
  if (!able || !able.canAccess) {
    setStatus("Not authorized");
    return;
  }
  const fi = qs("fileInput");
  if (fi) fi.onchange = doUploadAll;
  const uploadBtn = qs("uploadBtn");
  if (uploadBtn) uploadBtn.onclick = () => {
    const fi = qs("fileInput");
    if (fi) fi.click();
  };
  qs("searchBtn").onclick = doSearch;
  const navAllPhotos = qs("navAllPhotos");
  if (navAllPhotos) navAllPhotos.onclick = loadAll;
  const navBin = qs("navBin");
  if (navBin) navBin.onclick = loadBin;
  const addAlbumBtn = qs("addAlbumBtn");
  if (addAlbumBtn) addAlbumBtn.onclick = async () => {
    const name = prompt("Album name");
    if (name && name.trim()) {
      await apiAlbumCreate(name.trim());
      await loadAlbumsSidebar();
    }
  };
  await loadAlbumsSidebar();
  await loadCurrentYear();
});
