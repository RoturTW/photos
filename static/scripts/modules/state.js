export const state = {
  able: null,
  items: [],
  selectedItems: new Set(),
  favorites: new Set(),
  isUploading: false,
  imageObserver: null,
  scrollHooked: false,
  uploads: [],
  albums: [],
  view: "recent",
  groups: [],
  renderIndex: 0,
  lightboxIndex: -1,
  sidebarCollapsed: false,
  sidebarOpen: false,
  storageStats: null
};
