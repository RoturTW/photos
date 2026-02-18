import { qs } from './utils.js';
import { apiShareInfo, apiShareCreate, apiSharePatch } from './api.js';
import { showModal, showToast, copyLink } from './ui.js';

export async function openShareManagement(id) {
  const info = await apiShareInfo(id);
  if (!info) return;

  const sharedWith = info.sharedWith || [];
  const isPublic = info.isPublic || false;

  const ownerId = info.ownerId || document.body.getAttribute("data-user-id");
  const path = `/${ownerId}/${id}`;
  const fullUrl = window.location.origin + path;

  const renderContent = () => {
    let html = `
      <div class="share-public-toggle">
        <div class="share-public-label">
          <i data-lucide="${isPublic ? 'globe' : 'lock'}"></i>
          <span>Public Access</span>
        </div>
        <button class="btn ${isPublic ? 'btn-danger' : 'btn-secondary'}" id="togglePublicBtn" style="padding: 6px 12px; font-size: 13px;">
          ${isPublic ? 'Disable' : 'Enable'}
        </button>
      </div>

      <div class="share-link-box">
        <i data-lucide="link" style="width: 16px; min-width: 16px; opacity: 0.6;"></i>
        <div class="share-link-text">${fullUrl}</div>
        <button class="btn btn-secondary" id="modalCopyBtn" style="padding: 4px 10px; font-size: 12px;">Copy</button>
      </div>

      <div style="text-align: left; margin-bottom: 8px; font-weight: 500; font-size: 14px;">Shared with</div>
      <div class="share-list">
    `;

    if (sharedWith.length === 0) {
      html += '<p style="opacity: 0.6; padding: 12px; font-size: 14px; background: var(--bg-secondary); border-radius: 8px;">Not shared with any specific users.</p>';
    } else {
      sharedWith.forEach(u => {
        const displayName = u.username || u.userId;
        html += `
          <div class="share-item">
            <img class="share-user-avatar" src="https://avatars.rotur.dev/${displayName}" onerror="this.src='https://avatars.rotur.dev/default'">
            <div class="share-user-info">
              <span class="share-user-name">${displayName}</span>
            </div>
            <button class="btn-remove-user" data-user="${u.userId}">
              <i data-lucide="user-minus"></i>
            </button>
          </div>
        `;
      });
    }
    html += '</div>';
    return html;
  };

  showModal({
    title: "Share image",
    message: renderContent(),
    inputPlaceholder: "Add a person by username...",
    onConfirm: async (newUser) => {
      if (newUser && newUser.trim()) {
        const u = newUser.trim();
        const ok = await apiShareCreate(id, u);
        if (ok) {
          showToast(`Added ${u}`, "success");
          openShareManagement(id); // Reload
        } else {
          showToast("Failed to add user", "error");
        }
      }
    }
  });

  const modal = qs("genericModal");

  const toggleBtn = modal.querySelector("#togglePublicBtn");
  if (toggleBtn) toggleBtn.onclick = async () => {
    const ok = await apiSharePatch(id, [], [], !isPublic);
    if (ok) {
      showToast(isPublic ? "Access is now private" : "Image is now public", "success");
      openShareManagement(id);
    }
  };

  const copyBtn = modal.querySelector("#modalCopyBtn");
  if (copyBtn) copyBtn.onclick = () => copyLink(path);

  modal.querySelectorAll(".btn-remove-user").forEach(btn => {
    btn.onclick = async (e) => {
      e.stopPropagation();
      const user = btn.dataset.user;
      const ok = await apiSharePatch(id, [], [user]);
      if (ok) {
        showToast(`Removed ${user}`, "success");
        openShareManagement(id);
      }
    };
  });

  lucide.createIcons();
}
