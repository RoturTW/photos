import { state } from './state.js';
import { qs, monthName, formatBytes } from './utils.js';

const GROUP_BATCH = 6;

export function groupByMonth(items) {
    const groups = [];
    let currentMonth = "";
    let currentGroup = null;

    items.forEach(item => {
        const d = new Date(item.timestamp || Date.now());
        const month = isNaN(d.getTime()) ? "Unknown Date" : `${monthName(d.getMonth())} ${d.getFullYear()}`;
        if (month !== currentMonth) {
            currentMonth = month;
            currentGroup = { month, items: [] };
            groups.push(currentGroup);
        }
        currentGroup.items.push(item);
    });
    return groups;
}

export function render(items) {
    const grid = qs("grid");
    if (!grid) return;
    grid.innerHTML = "";
    if (!items || items.length === 0) {
        grid.innerHTML = '<div class="status" style="margin-top: 100px;">No photos found.</div>';
        return;
    }

    state.items = items;
    state.groups = groupByMonth(items);
    state.renderIndex = 0;

    const sentinel = document.createElement("div");
    sentinel.id = "loadSentinel";
    sentinel.style.height = "10px";
    grid.after(sentinel);

    renderMore();
    ensureObserver();
}

export function renderMore() {
    const grid = qs("grid");
    if (!grid || state.renderIndex >= state.groups.length) return;

    const end = Math.min(state.renderIndex + GROUP_BATCH, state.groups.length);
    for (let i = state.renderIndex; i < end; i++) {
        const group = state.groups[i];
        const section = document.createElement("div");
        section.className = "section";
        section.innerHTML = `<div class="section-header"><div class="section-title">${group.month}</div></div>`;

        const photoGrid = document.createElement("div");
        photoGrid.className = "photo-grid";
        group.items.forEach(item => {
            photoGrid.appendChild(createPhotoCard(item));
        });

        section.appendChild(photoGrid);
        grid.appendChild(section);
    }
    state.renderIndex = end;
    lucide.createIcons();
}

export function createPhotoCard(item) {
    const card = document.createElement("div");
    card.className = "photo-card";
    card.dataset.id = item.id;
    if (state.selectedItems.has(item.id)) card.classList.add("selected");

    const aspectRatio = (item.width && item.height) ? (item.width / item.height) : 1;
    const span = aspectRatio > 1.5 ? 2 : 1;
    card.style.gridColumnEnd = `span ${span}`;

    const isFavorite = state.favorites.has(item.id);
    const previewUrl = item.owner ? `/api/shared/${encodeURIComponent(item.owner)}/${encodeURIComponent(item.id)}` : (state.view === "bin" ? `/api/bin/preview/${encodeURIComponent(item.id)}` : `/api/image/preview/${encodeURIComponent(item.id)}`);

    card.innerHTML = `
    <img loading="lazy" src="${previewUrl}" alt="">
    <div class="photo-overlay">
      <div class="photo-overlay-btn zoom-btn" title="Open Lightbox">
        <i data-lucide="zoom-in"></i>
      </div>
      <button class="photo-overlay-btn favorite-btn ${isFavorite ? 'active' : ''}" title="Favorite">
        <i data-lucide="heart" ${isFavorite ? 'style="fill:currentColor"' : ''}></i>
      </button>
    </div>
  `;

    card.querySelector(".zoom-btn").onclick = (e) => {
        e.stopPropagation();
        window.dispatchEvent(new CustomEvent('gallery:openLightbox', { detail: { id: item.id } }));
    };

    card.querySelector(".favorite-btn").onclick = (e) => {
        e.stopPropagation();
        window.dispatchEvent(new CustomEvent('gallery:toggleFavorite', { detail: { id: item.id } }));
    };

    card.oncontextmenu = (e) => {
        e.preventDefault();
        window.dispatchEvent(new CustomEvent('gallery:contextMenu', { detail: { event: e, id: item.id } }));
    };

    card.onclick = (e) => {
        if (e.shiftKey || state.selectedItems.size > 0) {
            window.dispatchEvent(new CustomEvent('gallery:toggleSelection', { detail: { id: item.id } }));
        } else {
            window.dispatchEvent(new CustomEvent('gallery:openLightbox', { detail: { id: item.id } }));
        }
    };

    return card;
}

export function ensureObserver() {
    const rootEl = document.querySelector(".content") || null;
    if (!state.loadObserver) {
        state.loadObserver = new IntersectionObserver(entries => {
            for (const e of entries) {
                if (e.isIntersecting && state.renderIndex < state.groups.length) {
                    renderMore();
                }
            }
        }, { root: rootEl, rootMargin: "800px" });
    }
    const sentinel = qs("loadSentinel");
    if (sentinel) state.loadObserver.observe(sentinel);
}

export function renderStorage(stats) {
    const grid = qs("grid");
    if (!grid || !stats) return;

    grid.innerHTML = "";
    const totalUsed = stats.totalBytes + stats.binBytes;
    const quota = stats.quota || 0;
    const percent = quota > 0 ? Math.min((totalUsed / quota) * 100, 100) : 0;

    const overview = document.createElement("div");
    overview.className = "storage-overview";
    overview.innerHTML = `
    <div class="storage-card">
      <div class="storage-header">
        <i data-lucide="hard-drive" style="width: 32px; height: 32px;"></i>
        <div class="storage-title">Storage Usage</div>
      </div>
      <div class="storage-total">${formatBytes(totalUsed)}</div>
      <div class="storage-progress-container">
        <div class="storage-progress-bar" style="width: ${percent}%"></div>
      </div>
      <div class="storage-progress-text">${percent.toFixed(1)}% of ${formatBytes(quota)} used</div>
    </div>
  `;
    grid.appendChild(overview);
    lucide.createIcons();
}
