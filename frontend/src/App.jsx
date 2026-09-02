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

    try {
      const conversationMessages = await api(
        `/api/conversations/${conversation.id}/messages`
      );

      setMessages(conversationMessages);
    } catch (error) {
      console.error(error);
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
      />
    </div>
  );
}