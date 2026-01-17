import { qs } from './utils.js';

export function showModal({ title, message, inputPlaceholder, options, onConfirm, onCancel }) {
    const modal = qs("genericModal");
    const mTitle = qs("modalTitle");
    const mMsg = qs("modalMessage");
    const mInputContainer = qs("modalInputContainer");
    const mInput = qs("modalInput");
    const mSelectContainer = qs("modalSelectContainer");
    const mSelect = qs("modalSelect");
    const mConfirm = qs("modalConfirm");
    const mCancel = qs("modalCancel");

    if (!modal) return;

    mTitle.textContent = title || "";
    mMsg.innerHTML = message || "";

    mInputContainer.style.display = inputPlaceholder ? "block" : "none";
    if (inputPlaceholder) {
        mInput.placeholder = inputPlaceholder;
        mInput.value = "";
    }

    mSelectContainer.style.display = options ? "block" : "none";
    if (options) {
        mSelect.innerHTML = "";
        options.forEach(opt => {
            const o = document.createElement("option");
            o.value = opt.value;
            o.textContent = opt.label;
            mSelect.appendChild(o);
        });
    }

    modal.classList.add("active");

    const close = () => {
        modal.classList.remove("active");
        mConfirm.onclick = null;
        mCancel.onclick = null;
        document.removeEventListener("keydown", onKeyDown);
    };

    mConfirm.onclick = () => {
        let val = null;
        if (inputPlaceholder) val = mInput.value;
        else if (options) val = mSelect.value;
        if (onConfirm) onConfirm(val);
        close();
    };

    mCancel.onclick = () => {
        if (onCancel) onCancel();
        close();
    };

    const onKeyDown = (e) => {
        if (e.key === "Escape") mCancel.click();
        if (e.key === "Enter" && !options) mConfirm.click();
    };
    document.addEventListener("keydown", onKeyDown);
}

export function modalSelect(title, message, options) {
    return new Promise(resolve => {
        showModal({
            title, message, options,
            onConfirm: (val) => resolve(val),
            onCancel: () => resolve(null)
        });
    });
}

export function modalPrompt(title, message, placeholder) {
    return new Promise(resolve => {
        showModal({
            title, message, inputPlaceholder: placeholder,
            onConfirm: (val) => resolve(val),
            onCancel: () => resolve(null)
        });
    });
}

export function modalConfirm(title, message) {
    return new Promise(resolve => {
        showModal({
            title, message,
            onConfirm: () => resolve(true),
            onCancel: () => resolve(false)
        });
    });
}

export function showLoader(text = "Processing...") {
    const l = qs("globalLoader");
    const t = qs("loaderText");
    if (l && t) {
        t.textContent = text;
        l.classList.add("active");
    }
}

export function hideLoader() {
    const l = qs("globalLoader");
    if (l) l.classList.remove("active");
}

export function showToast(message, type = "info") {
    const container = qs("toastContainer");
    if (!container) return;

    const toast = document.createElement("div");
    toast.className = `toast ${type}`;
    toast.textContent = message;

    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.remove("show");
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

export function setStatus(msg) {
    const el = qs("status");
    if (el) {
        el.textContent = msg;
        if (msg) el.classList.add("active");
        else el.classList.remove("active");
    }
}

export function setViewTitle(title) {
    const el = qs("viewTitle");
    if (el) el.textContent = title;
    document.title = (title ? title + " - " : "") + "roturPhotos";
}

export function setActiveNav(id) {
    document.querySelectorAll(".nav-item").forEach(btn => btn.classList.remove("active"));
    const active = qs(id);
    if (active) active.classList.add("active");
}

export function clearActiveNav() {
    document.querySelectorAll(".nav-item").forEach(btn => btn.classList.remove("active"));
}

export function showUploadsPanel(show) {
    const panel = qs("uploadsPanel");
    if (!panel) return;
    if (show) panel.classList.remove("hidden");
    else panel.classList.add("hidden");
}

export function copyLink(url) {
    const fullUrl = window.location.origin + url;
    navigator.clipboard.writeText(fullUrl).then(() => {
        showToast("Link copied to clipboard", "success");
    }).catch(err => {
        console.error("Failed to copy link", err);
        showToast("Failed to copy link", "error");
    });
}
