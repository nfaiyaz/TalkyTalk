import { useEffect, useRef, useState } from 'react';

export default function ChatWindow({
  session,
  conversation,
  messages,
  onSend,
  onLoadOlderMessages,
  hasMoreMessages,
  loadingOlderMessages,
  onTypingStart,
  onTypingStop,
  typingUser,
}) {
  const [text, setText] = useState('');

  const bottomRef = useRef(null);
  const typingTimeoutRef = useRef(null);
  const typingStartedRef = useRef(false);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }

      if (typingStartedRef.current) {
        onTypingStop?.();
      }
    };
  }, [onTypingStop]);

  async function submit(event) {
    event.preventDefault();

    const body = text.trim();

    if (!body) return;

    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    if (typingStartedRef.current) {
      onTypingStop?.();
      typingStartedRef.current = false;
    }

    setText('');

    await onSend(body);
  }

  function handleTyping(event) {
    const value = event.target.value;

    setText(value);

    if (!conversation) return;

    if (!typingStartedRef.current) {
      onTypingStart?.();
      typingStartedRef.current = true;
    }

    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    typingTimeoutRef.current = setTimeout(() => {
      onTypingStop?.();
      typingStartedRef.current = false;
    }, 1000);
  }

  if (!conversation) {
    return (
      <main className="empty-chat">
        <div>
          <h2>Your messages</h2>
          <p>Select a chat or choose a person to begin.</p>
        </div>
      </main>
    );
  }

  return (
    <main className="chat">
      <header className="chat-header">
        <span className="avatar large">
          {conversation.other_user.name[0]?.toUpperCase()}
        </span>

        <div>
          <strong>{conversation.other_user.name}</strong>
          <small>{conversation.other_user.email}</small>
        </div>
      </header>

      <section className="messages">
        {conversation && hasMoreMessages && (
          <button onClick={onLoadOlderMessages}>
            {loadingOlderMessages
              ? 'Loading...'
              : 'Load older messages'}
          </button>
        )}

        {messages.map((m) => {
          const mine = m.sender_id === session.user.id;

          return (
            <div
              key={m.id}
              className={`message-row ${mine ? 'mine' : ''}`}
            >
              <div className="bubble">
                <div>{m.body}</div>

                <small>
                  {new Date(m.created_at).toLocaleTimeString([], {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </small>
              </div>
            </div>
          );
        })}

        <div ref={bottomRef} />
      </section>

      {typingUser && (
        <div className="typing-indicator">
          {typingUser.name} is typing...
        </div>
      )}

      <form className="composer" onSubmit={submit}>
        <input
          value={text}
          onChange={handleTyping}
          placeholder="Type a message..."
          maxLength={4000}
        />

        <button>Send</button>
      </form>
    </main>
  );
}