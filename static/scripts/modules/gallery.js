import { state } from './state.js';
import { qs, monthName, formatBytes } from './utils.js';

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

    const sorted = [...items].sort((a, b) => (b.timestamp || 0) - (a.timestamp || 0));

    state.items = sorted;
    state.groups = groupByMonth(sorted);

    state.groups.forEach(group => {
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
    });

    lucide.createIcons();
    ensureImageObserver();
}

export function createPhotoCard(item) {
    const card = document.createElement("div");
    card.className = "photo-card";
    card.dataset.id = item.id;
    if (state.selectedItems.has(item.id)) card.classList.add("selected");

    const isFavorite = state.favorites.has(item.id);
    const previewUrl = item.owner ? `/api/shared/${encodeURIComponent(item.owner)}/${encodeURIComponent(item.id)}` : (state.view === "bin" ? `/api/bin/preview/${encodeURIComponent(item.id)}` : `/api/image/preview/${encodeURIComponent(item.id)}`);

    card.dataset.src = previewUrl;

    card.innerHTML = `
    <img loading="lazy" src="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7" alt="">
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

export function ensureImageObserver() {
    if (state.imageObserver) {
        state.imageObserver.disconnect();
    }

    const rootEl = document.querySelector(".content") || null;

    state.imageObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            const card = entry.target;
            const img = card.querySelector('img');
            if (entry.isIntersecting) {
                if (img.src !== card.dataset.src) {
                    img.src = card.dataset.src;
                }
            } else {
                if (img.src && !img.src.startsWith("data:")) {
                    img.src = "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7";
                }
            }
        });
    }, {
        root: rootEl,
        rootMargin: "1000px 0px"
    });

    document.querySelectorAll('.photo-card').forEach(card => {
        state.imageObserver.observe(card);
    });
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
