import { useCallback, useEffect, useRef, useState } from 'react';
import { api, clearSession, getSession } from './api';
import { connectSocket } from './socket';
import AuthForm from './components/AuthForm';
import Sidebar from './components/Sidebar';
import ChatWindow from './components/ChatWindow';
import './styles.css';

export default function App() {
  const [session, setSession] = useState(getSession());
  const [users, setUsers] = useState([]);
  const [conversations, setConversations] = useState([]);
  const [selected, setSelected] = useState(null);
  const [messages, setMessages] = useState([]);
  const [hasMoreMessages, setHasMoreMessages] = useState(true);
  const [loadingOlderMessages, setLoadingOlderMessages] = useState(false);

  const selectedRef = useRef(null);

  useEffect(() => {
    selectedRef.current = selected;
  }, [selected]);

  const loadConversations = useCallback(async () => {
    try {
      const chats = await api('/api/conversations');
      setConversations(chats);
    } catch (error) {
      console.error(error);
    }
  }, []);

  const logout = useCallback(() => {
    clearSession();
    setSession(null);
    setSelected(null);
    setMessages([]);
  }, []);

  useEffect(() => {
    if (!session) return;

    async function loadInitial() {
      try {
        const [people, chats] = await Promise.all([
          api('/api/users'),
          api('/api/conversations'),
        ]);

        setUsers(people);
        setConversations(chats);
      } catch (error) {
        if (error.message === 'unauthorized') {
          logout();
        } else {
          console.error(error);
        }
      }
    }

    loadInitial();

    const socket = connectSocket((event) => {
      if (event.type !== 'new_message') return;

      const message = event.data;

      if (selectedRef.current?.id === message.conversation_id) {
        setMessages((current) =>
          current.some((m) => m.id === message.id)
            ? current
            : [...current, message]
        );
      }

      loadConversations();
    });

    return () => {
      socket?.close();
    };
  }, [session, loadConversations, logout]);

  async function selectConversation(conversation) {
  setSelected(conversation);
  setHasMoreMessages(true);

  try {
    const loadedMessages = await api(
      `/api/conversations/${conversation.id}/messages`
    );

    setMessages(loadedMessages);

    if (loadedMessages.length < 50) {
      setHasMoreMessages(false);
    }
  } catch (error) {
    console.error(error);
  }
}

async function loadOlderMessages() {
  if (!selected || loadingOlderMessages || !hasMoreMessages) {
    return;
  }

  if (messages.length === 0) {
    return;
  }

  const oldestMessage = messages[0];

  setLoadingOlderMessages(true);

  try {
    const olderMessages = await api(
      `/api/conversations/${selected.id}/messages?before=${oldestMessage.id}`
    );

    if (olderMessages.length === 0) {
      setHasMoreMessages(false);
      return;
    }

    setMessages((current) => [...olderMessages, ...current]);

    if (olderMessages.length < 50) {
      setHasMoreMessages(false);
    }
  } catch (error) {
    console.error(error);
  } finally {
    setLoadingOlderMessages(false);
  }
}

  async function startChat(user) {
    try {
      const conversation = await api('/api/conversations', {
        method: 'POST',
        body: JSON.stringify({
          user_id: user.id,
        }),
      });

      await loadConversations();
      await selectConversation(conversation);
    } catch (error) {
      alert(error.message);
    }
  }

  async function sendMessage(body) {
    if (!selected) return;

    try {
      await api(`/api/conversations/${selected.id}/messages`, {
        method: 'POST',
        body: JSON.stringify({
          body,
        }),
      });

      // The WebSocket event will append the saved message.
    } catch (error) {
      alert(error.message);
    }
  }

  if (!session) {
    return <AuthForm onAuthenticated={setSession} />;
  }

  return (
    <div className="app-shell">
      <Sidebar
        session={session}
        conversations={conversations}
        users={users}
        selectedId={selected?.id}
        onSelectConversation={selectConversation}
        onStartChat={startChat}
        onLogout={logout}
      />

      <ChatWindow
        session={session}
        conversation={selected}
        messages={messages}
        onSend={sendMessage}
        onLoadOlderMessages={loadOlderMessages}
  hasMoreMessages={hasMoreMessages}
  loadingOlderMessages={loadingOlderMessages}
      />
    </div>
  );
}