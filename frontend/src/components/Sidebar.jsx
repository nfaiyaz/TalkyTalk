export default function Sidebar({
	session,
	conversations,
	users,
	selectedId,
	onSelectConversation,
	onStartChat,
	onLogout
}) {
	return (
		<aside className="sidebar">
			<div className="sidebar-header">
				<div>
					<strong>{session.user.name}</strong>
					<small>{session.user.email}</small>
				</div>

				<button className="small-button" onClick={onLogout}>
					Logout
				</button>
			</div>

			<h3>Chats</h3>

			<div className="list">
				{conversations.map((c) => (
					<button
						key={c.id}
						className={`list-item ${
							selectedId === c.id ? 'active' : ''
						}`}
						onClick={() => onSelectConversation(c)}
					>
						<span className="avatar">
							{c.other_user.name[0]?.toUpperCase()}
						</span>

						<span>
							<strong>{c.other_user.name}</strong>
							<small>{c.other_user.email}</small>
						</span>
					</button>
				))}
			</div>

			<h3>People</h3>

			<div className="list people-list">
				{users.map((user) => (
					<button
						key={user.id}
						className="list-item"
						onClick={() => onStartChat(user)}
					>
						<span className="avatar">
							{user.name[0]?.toUpperCase()}
						</span>

						<span>
							<strong>{user.name}</strong>
							<small>Start chat</small>
						</span>
					</button>
				))}
			</div>
		</aside>
	);
}