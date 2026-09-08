import { getSession } from './api';

const WS_URL = 'ws://localhost:8080/ws';

export function connectSocket(onEvent) {
  let socket = null;
  let stopped = false;
  let reconnectTimer = null;
  let reconnectAttempt = 0;

  function connect() {
    if (stopped) return;

    const session = getSession();

    if (!session?.token) {
      return;
    }

    socket = new WebSocket(
      `${WS_URL}?token=${encodeURIComponent(session.token)}`
    );

    socket.onopen = () => {
      console.log('WebSocket connected');

      reconnectAttempt = 0;
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        onEvent(data);
      } catch (error) {
        console.error('Invalid WebSocket message:', error);
      }
    };

    socket.onclose = () => {
      console.log('WebSocket disconnected');

      if (stopped) return;

      scheduleReconnect();
    };

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  function scheduleReconnect() {
    if (stopped || reconnectTimer) return;

    reconnectAttempt += 1;

    const delay = Math.min(
      1000 * 2 ** (reconnectAttempt - 1),
      30000
    );

    console.log(
      `WebSocket reconnecting in ${delay / 1000} seconds...`
    );

    reconnectTimer = setTimeout(() => {
      reconnectTimer = null;
      connect();
    }, delay);
  }

  function send(data) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }

    socket.send(JSON.stringify(data));
  }

  connect();

  return {
    send,

    close() {
      stopped = true;

      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }

      if (socket) {
        socket.close();
        socket = null;
      }
    },
  };
}