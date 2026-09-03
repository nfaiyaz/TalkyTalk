import { useEffect, useRef, useState } from 'react';

export default function ChatWindow({ session, conversation, messages, onSend, onLoadOlderMessages, hasMoreMessages,
  loadingOlderMessages }) {
  const [text, setText] = useState('');

  const bottomRef = useRef(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  async function submit(event) {
    event.preventDefault();

    const body = text.trim();

    if (!body) return;

    setText('');

    await onSend(body);
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

      {conversation && hasMoreMessages && (
      <button onClick={onLoadOlderMessages}>
        {loadingOlderMessages ? 'Loading...' : 'Load older messages'}
      </button>
    )}

      <section className="messages">
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

      <form className="composer" onSubmit={submit}>
        <input
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Type a message..."
          maxLength={4000}
        />

        <button>Send</button>
      </form>
    </main>
  );
}
