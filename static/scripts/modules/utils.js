export function qs(id) {
    return document.getElementById(id);
}

export function monthName(m) {
    const months = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
    return months[m];
}

export function formatDate(timestamp) {
    const d = new Date(timestamp);
    const now = new Date();
    if (d.getFullYear() === now.getFullYear()) {
        return `${monthName(d.getMonth())} ${d.getDate()}`;
    }
    return `${monthName(d.getMonth())} ${d.getDate()}, ${d.getFullYear()}`;
}

export function formatBytes(bytes) {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}
