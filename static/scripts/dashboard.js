import { state } from './modules/state.js';
import { qs } from './modules/utils.js';
import {
  apiJSON, apiLogout, apiUpload, apiDelete, apiBinRestore,
  apiAlbumCreate, apiAlbumAdd, apiRotateImage
} from './modules/api.js';
import {
  modalConfirm, modalPrompt, modalSelect,
  showLoader, hideLoader, showToast,
  setStatus, showUploadsPanel, copyLink
} from './modules/ui.js';
import { render } from './modules/gallery.js';
import {
  loadRecent, loadAll, loadFavorites, loadBin, loadSharedWithMe, loadSharedByMe,
  loadAlbumsSidebar, refreshCurrentView, doSearch,
  saveFavorites, loadFavoritesState
} from './modules/actions.js';
import {
  openLightbox, closeLightbox, navigateLightbox, updateLightboxFavoriteBtn, refreshRotatedImage
} from './modules/lightbox.js';
import { loadStorage } from './modules/storage.js';
import { openShareManagement } from './modules/sharing.js';

window.toggleSidebar = toggleSidebar;

function toggleSelection(id) {
  if (state.selectedItems.has(id)) {
    state.selectedItems.delete(id);
  } else {
    state.selectedItems.add(id);
  }
  updateSelectionUI();
  updateSelectionBar();
}

function updateSelectionUI() {
  document.querySelectorAll(".photo-card").forEach(card => {
    const id = card.dataset.id;
    if (state.selectedItems.has(id)) card.classList.add("selected");
    else card.classList.remove("selected");
  });
}

async function updateSelectionBar() {
  const bar = qs("selectionBar");
  const count = qs("selectionCount");
  if (!bar || !count) return;

  const num = state.selectedItems.size;
  count.textContent = num === 1 ? "1 item selected" : `${num} items selected`;

  if (num > 0) {
    bar.classList.add("active");
    const dropdown = qs("selectAddToAlbumDropdown");
    if (dropdown) {
      const albums = await apiJSON("/api/albums");
      while (dropdown.options.length > 2) dropdown.remove(2);
      (albums || []).forEach(a => {
        const opt = document.createElement("option");
        opt.value = opt.textContent = a;
        dropdown.appendChild(opt);
      });
    }
  } else {
    bar.classList.remove("active");
  }
}

function clearSelection() {
  state.selectedItems.clear();
  updateSelectionUI();
  updateSelectionBar();
}

function toggleFavorite(id) {
  if (state.favorites.has(id)) state.favorites.delete(id);
  else state.favorites.add(id);
  saveFavorites();
  updateFavoriteUI(id);
}

function updateFavoriteUI(id) {
  const card = document.querySelector(`.photo-card[data-id="${id}"]`);
  if (card) {
    const btn = card.querySelector(".favorite-btn");
    if (btn) {
      if (state.favorites.has(id)) btn.classList.add("active");
      else btn.classList.remove("active");
    }
  }
}

function updateUploadsList() {
  const list = qs("uploadsList");
  if (!list) return;
  list.innerHTML = "";
  for (const u of state.uploads) {
    const item = document.createElement("div");
    item.className = "upload-item";
    item.innerHTML = `
      <div class="upload-status ${u.status}"></div>
      <div class="upload-details">
        <div class="upload-name">${u.name}</div>
        <div class="upload-progress-bar">
          <div class="upload-progress-bar-inner" style="width: ${u.progress || 0}%"></div>
        </div>
      </div>
      <div class="upload-progress">${u.status === 'uploading' ? Math.round(u.progress) + '%' : u.status.toUpperCase()}</div>
    `;
    list.appendChild(item);
  }
}

async function handleUploadAll(files) {
  const list = files || (qs("fileInput") ? qs("fileInput").files : null);
  if (!list || list.length === 0) return;

  const filesToUpload = Array.from(list).filter(f => f.type.startsWith("image/"));
  if (!filesToUpload.length) return;

  const total = filesToUpload.length;
  state.uploads = filesToUpload.map(f => ({ name: f.name, status: "pending", progress: 0 }));
  showUploadsPanel(true);
  updateUploadsList();

  const CONCURRENCY = 8;
  let idx = 0;

  async function worker() {
    while (idx < total) {
      const i = idx++;
      const f = filesToUpload[i];
      state.uploads[i].status = "uploading";
      updateUploadsList();

      try {
        const buf = await f.arrayBuffer();
        const res = await apiUpload(buf, (p) => {
          state.uploads[i].progress = p;
          updateUploadsList();
        });

        if (res.ok) {
          state.uploads[i].status = "ok";
          state.uploads[i].progress = 100;

          // Add to current album if in album view
          if (state.view && state.view.startsWith("album:")) {
            const albumName = state.view.slice(6);
            if (res.data && res.data.id) {
              await apiAlbumAdd(albumName, res.data.id);
            }
          }
        } else {
          state.uploads[i].status = "fail";
          state.uploads[i].error = res.error;
        }
      } catch (e) {
        state.uploads[i].status = "fail";
        state.uploads[i].error = e.message;
      }
      updateUploadsList();
    }
  }

  const workers = [];
  for (let i = 0; i < Math.min(CONCURRENCY, total); i++) {
    workers.push(worker());
  }
  await Promise.all(workers);

  setTimeout(() => showUploadsPanel(false), 3000);
  refreshCurrentView();
}

function openContextMenu(e, id) {
  const menu = qs("contextMenu");
  if (!menu) return;
  menu.innerHTML = "";

  const item = state.items.find(i => i.id === id);
  const isMine = item && !item.owner;

  const actions = [];

  actions.push({
    text: isMine ? "Share & Access" : "Copy Link",
    icon: isMine ? "share-2" : "link",
    fn: async () => {
      if (isMine) openShareManagement(id);
      else {
        const myName = document.querySelector(".user-dropdown-name")?.textContent?.trim();
        const path = item.owner ? `/${item.owner}/${id}` : `/${myName}/${id}`;
        await copyLink(path);
      }
    }
  });

  actions.push({
    text: "Add to album...",
    icon: "plus-square",
    fn: async () => {
      const albums = await apiJSON("/api/albums");
      const options = (albums || []).map(a => ({ value: a, label: a }));
      if (options.length === 0) {
        const name = await modalPrompt("New Album", "Create one:", "Album Name");
        if (name) { await apiAlbumCreate(name); await apiAlbumAdd(name, id); await loadAlbumsSidebar(); }
        return;
      }
      const name = await modalSelect("Add to Album", "Select:", options);
      if (name) { await apiAlbumAdd(name, id); showToast(`Added to ${name}`, "success"); }
    }
  });

  actions.push({
    text: state.favorites.has(id) ? "Remove Favorite" : "Favorite",
    icon: "heart",
    fn: () => toggleFavorite(id)
  });

  actions.push({
    text: "Rotate Right",
    icon: "rotate-cw",
    fn: async () => {
      setStatus("Rotating...");
      const ok = await apiRotateImage(id, 90);
      setStatus("");
      if (ok) { showToast("Rotated", "success"); await refreshRotatedImage(id, 90); }
    }
  });

  if (state.view !== "bin") {
    actions.push({
      text: "Move to Bin",
      icon: "trash-2",
      fn: async () => { if (await apiDelete(id)) { showToast("Moved to bin", "success"); refreshCurrentView(); } }
    });
  } else {
    actions.push({
      text: "Restore",
      icon: "archive-restore",
      fn: async () => { if (await apiBinRestore(id)) { showToast("Restored", "success"); refreshCurrentView(); } }
    });
  }

  actions.forEach(a => {
    const btn = document.createElement("button");
    btn.innerHTML = `<i data-lucide="${a.icon || 'circle'}"></i><span>${a.text}</span>`;
    btn.onclick = () => { a.fn(); menu.style.display = "none"; };
    menu.appendChild(btn);
  });

  menu.style.display = "block";
  menu.style.left = e.clientX + "px";
  menu.style.top = e.clientY + "px";

  lucide.createIcons();

  const hide = (ev) => { if (!menu.contains(ev.target)) { menu.style.display = "none"; document.removeEventListener("mousedown", hide); } };
  setTimeout(() => document.addEventListener("mousedown", hide), 10);
}

function toggleSidebar() {
  const app = qs("app");
  if (window.innerWidth <= 768) {
    state.sidebarOpen = !state.sidebarOpen;
    app.classList.toggle("sidebar-open", state.sidebarOpen);
  } else {
    state.sidebarCollapsed = !state.sidebarCollapsed;
    app.classList.toggle("sidebar-collapsed", state.sidebarCollapsed);
  }
}

async function checkDeepLink() {
  const path = window.location.pathname;
  const parts = path.split("/").filter(p => p);
  if (parts.length === 2 && !["auth", "static", "api"].includes(parts[0])) {
    const owner = parts[0];
    const imageId = parts[1];
    try {
      const info = await apiJSON(`/api/shared/info/${encodeURIComponent(owner)}/${encodeURIComponent(imageId)}`);
      if (info && info.id) {
        if (!state.items.find(i => i.id === info.id)) { state.items.unshift(info); render(state.items); }
        openLightbox(info.id);
      }
    } catch (err) { }
  }
}

// Global Event Listeners for Module communication
window.addEventListener('gallery:toggleFavorite', (e) => toggleFavorite(e.detail.id));
window.addEventListener('gallery:contextMenu', (e) => openContextMenu(e.detail.event, e.detail.id));
window.addEventListener('gallery:toggleSelection', (e) => toggleSelection(e.detail.id));
window.addEventListener('gallery:openLightbox', (e) => openLightbox(e.detail.id));

window.addEventListener("load", async () => {
  loadFavoritesState();
  const able = await apiJSON("/api/able");
  state.able = able;

  if (able) {
    if (!able.canAccess) {
      if (able.hasImages) {
        qs("readonlyBanner")?.classList.remove("hidden");
        qs("uploadBtn")?.classList.add("hidden");
      } else {
        qs("subscriptionOverlay")?.classList.add("active");
      }
    }
  }

  if (qs("logoutBtn")) qs("logoutBtn").onclick = async () => { await apiLogout(); window.location.href = "/auth"; };

  if (qs("userAvatar")) {
    const avatar = qs("userAvatar");
    const dropdown = qs("userDropdown");
    avatar.onclick = (e) => {
      e.stopPropagation();
      dropdown.classList.toggle("active");
    };
    document.addEventListener("mousedown", (e) => {
      if (!avatar.contains(e.target) && !dropdown.contains(e.target)) {
        dropdown.classList.remove("active");
      }
    });
  }
  if (qs("navAllPhotos")) qs("navAllPhotos").onclick = loadAll;
  if (qs("navFavorites")) qs("navFavorites").onclick = loadFavorites;
  if (qs("navBin")) qs("navBin").onclick = loadBin;
  if (qs("navSharedWithMe")) qs("navSharedWithMe").onclick = loadSharedWithMe;
  if (qs("navSharedByMe")) qs("navSharedByMe").onclick = loadSharedByMe;
  if (qs("navStorage")) qs("navStorage").onclick = loadStorage;
  if (qs("uploadBtn")) qs("uploadBtn").onclick = () => qs("fileInput").click();
  if (qs("fileInput")) qs("fileInput").onchange = () => handleUploadAll();

  if (qs("hamburgerBtn")) qs("hamburgerBtn").onclick = toggleSidebar;

  if (qs("lightboxClose")) qs("lightboxClose").onclick = closeLightbox;
  if (qs("lightboxPrev")) qs("lightboxPrev").onclick = () => navigateLightbox(-1);
  if (qs("lightboxNext")) qs("lightboxNext").onclick = () => navigateLightbox(1);
  if (qs("lightboxFavorite")) qs("lightboxFavorite").onclick = () => {
    const item = state.items[state.lightboxIndex];
    if (item) { toggleFavorite(item.id); updateLightboxFavoriteBtn(); }
  };
  if (qs("lightboxRotateRight")) qs("lightboxRotateRight").onclick = async () => {
    const item = state.items[state.lightboxIndex];
    if (item) { showLoader("Rotating..."); if (await apiRotateImage(item.id, 90)) await refreshRotatedImage(item.id, 90); hideLoader(); }
  };
  if (qs("lightboxRotateLeft")) qs("lightboxRotateLeft").onclick = async () => {
    const item = state.items[state.lightboxIndex];
    if (item) { showLoader("Rotating..."); if (await apiRotateImage(item.id, 270)) await refreshRotatedImage(item.id, 270); hideLoader(); }
  };
  if (qs("lightboxShare")) qs("lightboxShare").onclick = () => {
    const item = state.items[state.lightboxIndex];
    if (item) { if (item.owner) copyLink(`/${item.owner}/${item.id}`); else openShareManagement(item.id); }
  };
  if (qs("lightboxDelete")) qs("lightboxDelete").onclick = async () => {
    const item = state.items[state.lightboxIndex];
    if (item && await modalConfirm("Delete", "Delete photo?")) {
      await apiDelete(item.id); closeLightbox(); refreshCurrentView();
    }
  };

  if (qs("searchInput")) {
    qs("searchInput").onkeypress = (e) => { if (e.key === 'Enter') doSearch(e.target.value); };
  }

  if (qs("selectDeselect")) qs("selectDeselect").onclick = clearSelection;
  if (qs("selectAllBtn")) qs("selectAllBtn").onclick = () => {
    if (state.selectedItems.size === state.items.length) clearSelection();
    else { state.items.forEach(it => state.selectedItems.add(it.id)); updateSelectionUI(); updateSelectionBar(); }
  };

  if (qs("selectDelete")) qs("selectDelete").onclick = async () => {
    if (state.selectedItems.size === 0) return;
    if (await modalConfirm("Delete", `Delete ${state.selectedItems.size} items?`)) {
      showLoader();
      for (const id of state.selectedItems) await apiDelete(id);
      hideLoader(); clearSelection(); refreshCurrentView();
    }
  };

  if (qs("selectAddToAlbumDropdown")) {
    qs("selectAddToAlbumDropdown").onchange = async (e) => {
      const val = e.target.value;
      if (!val) return;
      if (val === "NEW_ALBUM") {
        const name = await modalPrompt("New Album", "Name:", "Album Name");
        if (name) {
          await apiAlbumCreate(name);
          for (const id of state.selectedItems) await apiAlbumAdd(name, id);
          await loadAlbumsSidebar(); clearSelection();
        }
      } else {
        for (const id of state.selectedItems) await apiAlbumAdd(val, id);
        clearSelection();
      }
      e.target.value = "";
    };
  }

  if (qs("addAlbumBtn")) qs("addAlbumBtn").onclick = async () => {
    const name = await modalPrompt("New Album", "Enter album name:", "Album Name");
    if (name && name.trim()) {
      const ok = await apiAlbumCreate(name.trim());
      if (ok) {
        showToast("Album created", "success");
        await loadAlbumsSidebar();
      } else {
        showToast("Failed to create album", "error");
      }
    }
  };

  await loadAlbumsSidebar();
  await loadAll();
  await checkDeepLink();
  lucide.createIcons();
});

document.addEventListener("keydown", (e) => {
  const lb = qs("lightbox");
  if (lb && lb.classList.contains("active")) {
    if (e.key === "Escape") closeLightbox();
    if (e.key === "ArrowLeft") navigateLightbox(-1);
    if (e.key === "ArrowRight") navigateLightbox(1);
  }
});