// Generates a full API key + its prefix + sha-256 hash.
// Call in the browser; full key shown once and never stored.
export async function generateApiKey(): Promise<{
  fullKey: string;
  prefix: string;
  hash: string;
}> {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  const hex = Array.from(array)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");

  const fullKey = `rk_live_${hex}`;
  const prefix = fullKey.slice(0, 16); // "rk_live_" + 8 chars

  const encoder = new TextEncoder();
  const data = encoder.encode(fullKey);
  const hashBuf = await crypto.subtle.digest("SHA-256", data);
  const hashArr = Array.from(new Uint8Array(hashBuf));
  const hash = hashArr.map((b) => b.toString(16).padStart(2, "0")).join("");

  return { fullKey, prefix, hash };
}
