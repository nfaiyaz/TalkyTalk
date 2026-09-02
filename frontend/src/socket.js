import { getSession } from './api';
export function connectSocket(onEvent) {
const session = getSession();
if (!session?.token) return null;
const base = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';
const socket = new WebSocket(`${base}?token=${encodeURIComponent(session.token)}`);
socket.onmessage = (event) => {
try {
onEvent(JSON.parse(event.data));
} catch (error) {
console.error('Bad WebSocket event', error);
}
};
socket.onerror = (error) => console.error('WebSocket error', error);
return socket;
}