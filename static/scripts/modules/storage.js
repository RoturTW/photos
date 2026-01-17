import { state } from './state.js';
import { qs, formatBytes } from './utils.js';
import { showToast, showLoader, hideLoader, modalConfirm, setViewTitle, setActiveNav, setStatus } from './ui.js';
import { apiJSON, apiDelete } from './api.js';
import { openLightbox } from './lightbox.js';

export async function loadStorage() {
    setActiveNav("navStorage");
    setStatus("Loading storage stats...");
    const [stats, items] = await Promise.all([
        apiJSON("/api/storage"),
        apiJSON("/api/images/all")
    ]);
    setStatus("");
    if (!stats) return;

    state.items = items || [];
    state.storageStats = stats;
    state.view = "storage";
    setViewTitle("Storage");

    renderStorage();
}

export function patchStorageState(id) {
    if (!state.storageStats) return;
    const size = (state.storageStats.fileSizes && state.storageStats.fileSizes[id]) ? state.storageStats.fileSizes[id] : 0;
    state.storageStats.totalBytes -= size;
    state.storageStats.imageCount--;

    if (state.storageStats.largestImages) {
        state.storageStats.largestImages = state.storageStats.largestImages.filter(li => li.id !== id);
    }

    if (state.storageStats.duplicateGroups) {
        state.storageStats.duplicateGroups = state.storageStats.duplicateGroups.reduce((acc, group) => {
            const newIds = group.ids.filter(gid => gid !== id);
            if (newIds.length > 1) {
                group.ids = newIds;
                acc.push(group);
            }
            return acc;
        }, []);
    }

    state.items = state.items.filter(it => it.id !== id);
}

export async function cleanUpAllDuplicates() {
    if (!state.storageStats || !state.storageStats.duplicateGroups || state.storageStats.duplicateGroups.length === 0) return;

    const toDelete = [];
    state.storageStats.duplicateGroups.forEach(group => {
        for (let i = 1; i < group.ids.length; i++) {
            toDelete.push(group.ids[i]);
        }
    });

    if (toDelete.length === 0) return;

    if (!(await modalConfirm("Clean Up All", `This will delete ${toDelete.length} duplicate images, keeping one unique copy of each. Continue?`))) return;

    showLoader(`Cleaning up ${toDelete.length} duplicates...`);
    try {
        for (const id of toDelete) {
            await apiDelete(id);
            patchStorageState(id);
        }
        showToast(`Successfully removed ${toDelete.length} duplicates`, "success");
    } catch (err) {
        console.error("Cleanup error:", err);
        showToast("An error occurred during cleanup", "error");
    } finally {
        hideLoader();
        renderStorage();
    }
}

export function renderStorage() {
    const stats = state.storageStats;
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
      <div class="storage-progress-text">
        ${percent.toFixed(1)}% of ${formatBytes(quota)} used
      </div>
      <div class="storage-breakdown">
        <div class="storage-item">
          <span class="storage-label">Photos (${stats.imageCount})</span>
          <span class="storage-value">${formatBytes(stats.totalBytes)}</span>
        </div>
        <div class="storage-item">
          <span class="storage-label">Bin</span>
          <span class="storage-value">${formatBytes(stats.binBytes)}</span>
        </div>
      </div>
    </div>
  `;
    grid.appendChild(overview);

    if (stats.largestImages && stats.largestImages.length > 0) {
        const section = document.createElement("div");
        section.className = "section";
        section.innerHTML = `<div class="section-header"><div class="section-title">Largest Files</div></div>`;

        const largeGrid = document.createElement("div");
        largeGrid.className = "photo-grid";

        for (const img of stats.largestImages) {
            const card = document.createElement("div");
            card.className = "photo-card storage-photo";
            card.innerHTML = `
        <img loading="lazy" src="/api/image/preview/${encodeURIComponent(img.id)}" style="width:100%;height:100%;object-fit:cover;">
        <div class="storage-photo-size">${formatBytes(img.bytes)}</div>
        <div class="photo-overlay">
          <button class="photo-overlay-btn storage-delete-btn" title="Delete">
            <i data-lucide="trash-2"></i>
          </button>
        </div>
      `;
            card.querySelector(".storage-delete-btn").onclick = async (e) => {
                e.stopPropagation();
                if (await modalConfirm("Delete Photo", `Delete this ${formatBytes(img.bytes)} image?`)) {
                    await apiDelete(img.id);
                    patchStorageState(img.id);
                    renderStorage();
                    showToast("Photo deleted", "success");
                }
            };
            card.onclick = () => openLightbox(img.id);
            largeGrid.appendChild(card);
        }

        section.appendChild(largeGrid);
        grid.appendChild(section);
    }

    if (stats.duplicateGroups && stats.duplicateGroups.length > 0) {
        const section = document.createElement("div");
        section.className = "section";
        section.style.marginTop = "40px";
        section.innerHTML = `
      <div class="section-header" style="display: flex; justify-content: space-between; align-items: flex-start;">
        <div>
          <div class="section-title">Duplicate Images</div>
          <p style="opacity: 0.6; margin-top: 4px; font-size: 14px;">Images with identical content found in your storage.</p>
        </div>
        <button class="btn btn-secondary" id="cleanUpDuplicatesBtn" style="padding: 6px 12px; font-size: 13px;">
          <i data-lucide="sparkles" style="width:14px;height:14px;margin-right:6px;"></i>Clean up all
        </button>
      </div>
    `;

        const cleanBtn = section.querySelector("#cleanUpDuplicatesBtn");
        if (cleanBtn) cleanBtn.onclick = cleanUpAllDuplicates;

        const dupContainer = document.createElement("div");
        dupContainer.style.display = "flex";
        dupContainer.style.flexDirection = "column";
        dupContainer.style.gap = "24px";

        for (const group of stats.duplicateGroups) {
            const groupDiv = document.createElement("div");
            groupDiv.className = "duplicate-group";
            groupDiv.style.cssText = "background: var(--bg-secondary); padding: 16px; border-radius: var(--border-radius); border: 1px solid var(--border-color);";

            const groupHeader = document.createElement("div");
            groupHeader.style.cssText = "margin-bottom: 12px; font-size: 13px; font-weight: 600; opacity: 0.8;";
            groupHeader.innerHTML = `<i data-lucide="layers" style="width:14px;height:14px;vertical-align:middle;margin-right:6px;opacity:0.6;"></i>${group.ids.length} identical copies`;
            groupDiv.appendChild(groupHeader);

            const groupGrid = document.createElement("div");
            groupGrid.style.display = "grid";
            groupGrid.style.gridTemplateColumns = "repeat(auto-fill, minmax(140px, 1fr))";
            groupGrid.style.gap = "12px";

            for (const id of group.ids) {
                const card = document.createElement("div");
                card.className = "photo-card storage-photo";
                card.style.aspectRatio = "1";
                card.innerHTML = `
          <img loading="lazy" src="/api/image/preview/${encodeURIComponent(id)}" style="width:100%;height:100%;object-fit:cover;">
          <div class="photo-overlay">
            <button class="photo-overlay-btn storage-delete-btn" title="Delete">
              <i data-lucide="trash-2"></i>
            </button>
          </div>
        `;
                card.querySelector(".storage-delete-btn").onclick = async (e) => {
                    e.stopPropagation();
                    if (await modalConfirm("Delete Duplicate", "Delete this duplicate copy?")) {
                        await apiDelete(id);
                        patchStorageState(id);
                        renderStorage();
                        showToast("Duplicate copy deleted", "success");
                    }
                };
                card.onclick = () => openLightbox(id);
                groupGrid.appendChild(card);
            }
            groupDiv.appendChild(groupGrid);
            dupContainer.appendChild(groupDiv);
        }

        section.appendChild(dupContainer);
        grid.appendChild(section);
    }

    lucide.createIcons();
}
