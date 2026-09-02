const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export function getSession() {
const raw = localStorage.getItem('chat_session');
return raw ? JSON.parse(raw) : null;
}

export function saveSession(session) {
localStorage.setItem('chat_session', JSON.stringify(session));
}

export function clearSession() {
localStorage.removeItem('chat_session');
}

export async function api(path, options = {}) {
const session = getSession();
const headers = {
'Content-Type': 'application/json',
...(options.headers || {}),
};
if (session?.token) {
headers.Authorization = `Bearer ${session.token}`;
}
const response = await fetch(`${API_URL}${path}`, { ...options, headers });
const data = await response.json().catch(() => ({}));
if (!response.ok) {
throw new Error(data.error || `Request failed (${response.status})`);
}
return data;
}