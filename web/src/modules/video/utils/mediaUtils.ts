// Convert S3 key or relative URL to browser-accessible URL
export function resolveMediaUrl(key: string): string {
  if (!key) return "";
  if (key.startsWith("http://") || key.startsWith("https://") || key.startsWith("blob:") || key.startsWith("data:")) {
    return key;
  }
  if (key.startsWith("/")) {
    return key;
  }
  return `/files/${key}`;
}
