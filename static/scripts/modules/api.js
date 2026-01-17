export async function apiJSON(url) {
    try {
        const res = await fetch(url);
        if (!res.ok) return null;
        return await res.json();
    } catch (e) {
        return null;
    }
}

export async function apiUpload(buffer, onProgress) {
    return new Promise((resolve) => {
        const xhr = new XMLHttpRequest();
        xhr.open("POST", "/api/image/upload");
        xhr.upload.onprogress = (event) => {
            if (event.lengthComputable) {
                const percent = (event.loaded / event.total) * 100;
                if (onProgress) onProgress(percent);
            }
        };
        xhr.onload = () => {
            if (xhr.status === 200) {
                resolve({ ok: true, data: JSON.parse(xhr.responseText) });
            } else {
                let err = "Upload failed";
                try { err = JSON.parse(xhr.responseText).error; } catch (e) { }
                resolve({ ok: false, error: err });
            }
        };
        xhr.onerror = () => resolve({ ok: false, error: "Network error" });
        xhr.send(buffer);
    });
}

export async function apiAlbums() {
    return await apiJSON("/api/albums");
}

export async function apiAlbumCreate(name) {
    const res = await fetch(`/api/albums?name=${encodeURIComponent(name)}`, {
        method: "POST"
    });
    return res.ok;
}

export async function apiAlbumImages(name) {
    return await apiJSON("/api/albums/" + encodeURIComponent(name));
}

export async function apiAlbumAdd(name, id) {
    const res = await fetch(`/api/albums/${encodeURIComponent(name)}/add?id=${encodeURIComponent(id)}`, {
        method: "POST"
    });
    return res.ok;
}

export async function apiAlbumRemove(name, id) {
    const res = await fetch(`/api/albums/${encodeURIComponent(name)}/remove?id=${encodeURIComponent(id)}`, {
        method: "POST"
    });
    return res.ok;
}

export async function apiBinList() {
    return await apiJSON("/api/bin");
}

export async function apiBinRestore(id) {
    const res = await fetch("/api/bin/restore/" + encodeURIComponent(id), { method: "POST" });
    return res.ok;
}

export async function apiBinDelete(id) {
    const res = await fetch("/api/bin/" + encodeURIComponent(id), { method: "DELETE" });
    return res.ok;
}

export async function apiBinEmpty() {
    const res = await fetch("/api/bin/empty", { method: "POST" });
    return res.ok;
}

export async function apiLogout() {
    await fetch("/api/logout");
}

export async function apiAlbumDelete(name) {
    const res = await fetch("/api/albums/" + encodeURIComponent(name), {
        method: "DELETE"
    });
    return res.ok;
}

export async function apiShareCreate(imageId, username) {
    const res = await fetch("/api/share", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ imageId, username })
    });
    return res.ok;
}

export async function apiShareOthers() {
    return await apiJSON("/api/share/others");
}

export async function apiShareMine() {
    return await apiJSON("/api/share/mine");
}

export async function apiShareInfo(id) {
    return await apiJSON("/api/share/info/" + encodeURIComponent(id));
}

export async function apiSharePatch(id, add, remove, isPublic) {
    const body = {};
    if (add) body.add = add;
    if (remove) body.remove = remove;
    if (isPublic !== undefined) body.isPublic = isPublic;

    const res = await fetch("/api/share/" + encodeURIComponent(id), {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body)
    });
    return res.ok;
}

export async function apiDelete(id) {
    const res = await fetch("/api/image/" + encodeURIComponent(id), { method: "DELETE" });
    return res.ok;
}

export async function apiRotateImage(id, angle) {
    const res = await fetch(`/api/image/${encodeURIComponent(id)}/rotate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ angle })
    });
    return res.ok;
}
