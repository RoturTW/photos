import { state } from './state.js';
import { qs, formatDate } from './utils.js';

export function formatLightboxMetadata(item) {
    let meta = formatDate(item.timestamp);
    if (item.width && item.height) {
        meta += ` • ${item.width} x ${item.height}`;
    }
    if (item.filename) {
        meta += ` • ${item.filename}`;
    }
    return meta;
}

export function openLightbox(id) {
    const index = state.items.findIndex(item => item.id === id);
    if (index === -1) return;

    state.lightboxIndex = index;
    const lightbox = qs("lightbox");
    const image = qs("lightboxImage");
    const metadata = qs("lightboxMetadata");

    if (!lightbox || !image) return;

    const item = state.items[index];
    let url = "";
    if (item.owner) {
        url = "/api/shared/" + encodeURIComponent(item.owner) + "/";
    } else {
        url = state.view === "bin" ? "/api/bin/preview/" : "/api/image/";
    }
    image.src = url + encodeURIComponent(item.id);

    if (metadata) {
        metadata.textContent = formatLightboxMetadata(item);
    }

    updateLightboxFavoriteBtn();
    updateURLForImage(item);

    lucide.createIcons();

    lightbox.classList.add("active");
    document.body.style.overflow = "hidden";
}

export function closeLightbox() {
    const lightbox = qs("lightbox");
    if (lightbox) {
        lightbox.classList.remove("active");
    }
    document.body.style.overflow = "";
    state.lightboxIndex = -1;
    history.replaceState(null, "", "/");
}

export function navigateLightbox(direction) {
    const newIndex = state.lightboxIndex + direction;
    if (newIndex < 0 || newIndex >= state.items.length) return;

    state.lightboxIndex = newIndex;
    const item = state.items[newIndex];
    const image = qs("lightboxImage");
    const metadata = qs("lightboxMetadata");

    if (image) {
        let url = "";
        if (item.owner) {
            url = "/api/shared/" + encodeURIComponent(item.owner) + "/";
        } else {
            url = state.view === "bin" ? "/api/bin/preview/" : "/api/image/";
        }
        image.src = url + encodeURIComponent(item.id);
    }

    if (metadata) {
        metadata.textContent = formatLightboxMetadata(item);
    }

    updateLightboxFavoriteBtn();
    updateURLForImage(item);
}

export function updateLightboxFavoriteBtn() {
    const btn = qs("lightboxFavorite");
    if (!btn) return;

    const item = state.items[state.lightboxIndex];
    if (!item) return;

    if (state.favorites.has(item.id)) {
        btn.classList.add("active");
        btn.innerHTML = '<i data-lucide="heart" style="fill:currentColor"></i>';
    } else {
        btn.classList.remove("active");
        btn.innerHTML = '<i data-lucide="heart"></i>';
    }
    lucide.createIcons({
        attrs: { class: "lucide-icon" },
        nameAttr: 'data-lucide'
    });
}

function updateURLForImage(item) {
    if (!item) return;
    const myUser = document.querySelector(".user-dropdown-name")?.textContent?.trim();
    const owner = item.owner || myUser;
    if (owner && item.id) {
        history.replaceState(null, "", `/${owner}/${item.id}`);
    }
}

export async function refreshRotatedImage(id, angle) {
    const item = state.items.find(it => it.id === id);
    if (item) {
        if (angle === 90 || angle === 270) {
            const oldWidth = item.width;
            item.width = item.height;
            item.height = oldWidth;
        }
    }
    const img = document.querySelector(`.photo-card[data-id="${id}"] img`);
    if (img) {
        const base = state.view === "bin" ? "/api/bin/preview/" : "/api/image/preview/";
        img.src = base + encodeURIComponent(id) + "?t=" + new Date().getTime();
    }
    if (state.lightboxIndex !== -1 && state.items[state.lightboxIndex].id === id) {
        const lightboxImage = qs("lightboxImage");
        if (lightboxImage) {
            const url = state.view === "bin" ? "/api/bin/preview/" : "/api/image/";
            lightboxImage.src = url + encodeURIComponent(id) + "?t=" + new Date().getTime();
        }
        const metadata = qs("lightboxMetadata");
        if (metadata) {
            metadata.textContent = formatLightboxMetadata(item);
        }
    }
}
