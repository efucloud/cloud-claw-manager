export const toNumber = (value?: number) => Number(value || 0);

export const normalizeProvider = (value?: string) => {
  const text = String(value || '').trim();
  if (!text || text.toLowerCase() === 'unknown') {
    return '';
  }
  return text;
};

export const normalizeUrlHost = (value?: string) => {
  const text = String(value || '').trim();
  if (!text) {
    return '';
  }
  try {
    return new URL(text).host || text;
  } catch (_error) {
    return text;
  }
};

export const normalizeURL = (raw?: string) => {
  const value = String(raw || '').trim();
  if (!value) {
    return '';
  }
  if (/^https?:\/\//i.test(value)) {
    return value;
  }
  return `https://${value}`;
};

export const isPreviewable = (raw?: string) => {
  const value = normalizeURL(raw);
  if (!value) {
    return false;
  }
  try {
    const u = new URL(value);
    return u.protocol === 'http:' || u.protocol === 'https:';
  } catch (_error) {
    return false;
  }
};

export const formatBytes = (value?: number) => {
  const num = Number(value || 0);
  if (!Number.isFinite(num) || num <= 0) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let n = num;
  let idx = 0;
  while (n >= 1024 && idx < units.length - 1) {
    n /= 1024;
    idx += 1;
  }
  const fixed = n >= 100 ? 0 : n >= 10 ? 1 : 2;
  return `${n.toFixed(fixed)} ${units[idx]}`;
};

export const formatNumber = (value?: number, digits = 2) => {
  const num = Number(value || 0);
  if (!Number.isFinite(num)) {
    return '0';
  }
  return num.toFixed(digits);
};

export const decodePart = (value?: string) => {
  const text = String(value || '').trim();
  if (!text) {
    return '';
  }
  try {
    return decodeURIComponent(text);
  } catch (_error) {
    return text;
  }
};
