const base = process.env.NEXT_PUBLIC_API_URL ?? "http://127.0.0.1:4000";

export async function apiFetch(
  path: string,
  token: string | null,
  init?: RequestInit
): Promise<Response> {
  const headers = new Headers(init?.headers);
  headers.set("Content-Type", "application/json");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  return fetch(`${base}${path}`, { ...init, headers });
}
