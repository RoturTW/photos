import { state } from './state.js';
import { apiJSON, apiBinList, apiBinEmpty, apiShareOthers, apiShareMine, apiAlbums, apiAlbumImages, apiAlbumDelete } from './api.js';
import { setStatus, setViewTitle, setActiveNav, clearActiveNav, modalConfirm } from './ui.js';
import { render, renderStorage } from './gallery.js';
import { qs } from './utils.js';

export async function loadRecent() {
    state.view = "recent";
    setViewTitle("Photos");
    setStatus("Loading recent...");
    const items = await apiJSON("/api/images/recent");
    setStatus("");

    const upBtn = qs("uploadBtn");
    if (upBtn) upBtn.classList.remove("hidden");
    const emptyBtn = qs("emptyBinBtn");
    if (emptyBtn) emptyBtn.style.display = "none";

    render(items || []);
    setActiveNav("navAllPhotos");
}

export async function loadAll() {
    state.view = "all";
    setViewTitle("All Photos");
    setStatus("Loading all...");
    const items = await apiJSON("/api/images/all");
    setStatus("");

    const upBtn = qs("uploadBtn");
    if (upBtn) upBtn.classList.remove("hidden");
    const emptyBtn = qs("emptyBinBtn");
    if (emptyBtn) emptyBtn.style.display = "none";

    render(items || []);
    setActiveNav("navAllPhotos");
}

export async function loadFavorites() {
    setActiveNav("navFavorites");
    state.view = "favorites";
    setViewTitle("Favorites");
    setStatus("Loading favorites...");
    const allItems = await apiJSON("/api/images/all");
    const items = (allItems || []).filter(it => state.favorites.has(it.id));
    setStatus("");
    render(items || []);
}

export async function loadBin() {
    setActiveNav("navBin");
    setStatus("Loading bin...");
    const items = await apiBinList();
    setStatus("");
    if (items) {
        state.view = "bin";
        setViewTitle("Bin");
        const upBtn = qs("uploadBtn");
        if (upBtn) upBtn.classList.add("hidden");

        const emptyBtn = qs("emptyBinBtn");
        if (emptyBtn) {
            emptyBtn.style.display = items.length > 0 ? "inline-flex" : "none";
            emptyBtn.onclick = async () => {
                if (await modalConfirm("Empty Bin", "Permanently delete all items in Bin? This cannot be undone.")) {
                    setStatus("Emptying bin...");
                    await apiBinEmpty();
                    setStatus("");
                    await loadBin();
                }
            };
        }
        render(items);
    }
}

export async function loadSharedWithMe() {
    setActiveNav("navSharedWithMe");
    state.view = "sharedWith";
    setViewTitle("Shared with me");
    setStatus("Loading shared photos...");
    const items = await apiShareOthers();
    setStatus("");
    render(items || []);
}

export async function loadSharedByMe() {
    setActiveNav("navSharedByMe");
    state.view = "sharedMy";
    setViewTitle("Shared by me");
    setStatus("Loading shared photos...");
    const items = await apiShareMine();
    setStatus("");
    render(items || []);
}

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
    renderStorage(stats);
}

export async function loadAlbumsSidebar() {
    const names = await apiAlbums();
    state.albums = names || [];
    const list = qs("albumList");
    if (!list) return;
    list.innerHTML = "";

    for (const name of state.albums) {
        const wrapper = document.createElement("div");
        wrapper.className = "album-item";
        wrapper.style.cssText = "display: flex; align-items: center; gap: 4px;";

        const btn = document.createElement("button");
        btn.className = "nav-item";
        btn.style.flex = "1";
        btn.innerHTML = `<i data-lucide="album" style="width: 16px; height: 16px; margin-right: 8px;"></i><span>${name}</span>`;
        btn.onclick = async () => {
            state.view = "album:" + name;
            setStatus("Loading album...");
            const items = await apiAlbumImages(name);
            setStatus("");
            setViewTitle(name);
            clearActiveNav();
            btn.classList.add("active");
            render(items || []);
        };

        const deleteBtn = document.createElement("button");
        deleteBtn.className = "album-delete-btn";
        deleteBtn.title = "Delete album";
        deleteBtn.innerHTML = '<i data-lucide="trash-2"></i>';
        deleteBtn.onclick = async (e) => {
            e.stopPropagation();
            if (await modalConfirm("Delete Album", `Delete album "${name}"? Photos will not be deleted.`)) {
                await apiAlbumDelete(name);
                await loadAlbumsSidebar();
                if (state.view === "album:" + name) loadAll();
            }
        };

        wrapper.appendChild(btn);
        wrapper.appendChild(deleteBtn);
        list.appendChild(wrapper);
    }
    lucide.createIcons();
}

export async function refreshCurrentView() {
    if (state.view === "recent") return loadRecent();
    if (state.view === "all") return loadAll();
    if (state.view === "favorites") return loadFavorites();
    if (state.view && state.view.startsWith("album:")) {
        const name = state.view.slice(6);
        setStatus("Loading album...");
        const items = await apiAlbumImages(name);
        setStatus("");
        return render(items || []);
    }
    if (state.view === "bin") return loadBin();
    if (state.view === "storage") return loadStorage();
    return loadAll();
}

export async function doSearch(q) {
    if (!q) return;
    setStatus("Searching...");
    const items = await apiJSON("/api/search?q=" + encodeURIComponent(q));
    setStatus("");
    if (items) {
        state.view = "search:" + q;
        setViewTitle('Search: "' + q + '"');
        clearActiveNav();
        render(items);
    }
}

export function saveFavorites() {
    try {
        localStorage.setItem("favorites", JSON.stringify([...state.favorites]));
    } catch (e) {
        console.error("Failed to save favorites", e);
    }
}

export function loadFavoritesState() {
    const saved = localStorage.getItem("favorites");
    if (saved) {
        try {
            state.favorites = new Set(JSON.parse(saved));
        } catch (e) {
            console.error("Failed to load favorites", e);
        }
    }
}
