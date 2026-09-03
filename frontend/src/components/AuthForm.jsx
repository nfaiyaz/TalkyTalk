import { useState } from 'react';
import { api, saveSession } from '../api';

export default function AuthForm({ onAuthenticated }) {
	const [mode, setMode] = useState('login');
	const [form, setForm] = useState({
		name: '',
		email: '',
		password: ''
	});
	const [error, setError] = useState('');
	const [loading, setLoading] = useState(false);

	async function submit(event) {
		event.preventDefault();
		setError('');
		setLoading(true);

		try {
			const path = mode === 'login' ? '/api/login' : '/api/register';

			const payload =
				mode === 'login'
					? {
							email: form.email,
							password: form.password
						}
					: form;

			const session = await api(path, {
				method: 'POST',
				body: JSON.stringify(payload)
			});

			saveSession(session);
			onAuthenticated(session);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	}

	return (
		<div className="auth-page">
			<form className="auth-card" onSubmit={submit}>
				<h1>TalkyTalk (Go Chat)</h1>

				<p>
					{mode === 'login'
						? 'Sign in to continue'
						: 'Create your account'}
				</p>

				{mode === 'register' && (
					<input
						placeholder="Name"
						value={form.name}
						onChange={(e) =>
							setForm({
								...form,
								name: e.target.value
							})
						}
						required
					/>
				)}

				<input
					type="email"
					placeholder="Email"
					value={form.email}
					onChange={(e) =>
						setForm({
							...form,
							email: e.target.value
						})
					}
					required
				/>

				<input
					type="password"
					placeholder="Password (8+ characters)"
					value={form.password}
					onChange={(e) =>
						setForm({
							...form,
							password: e.target.value
						})
					}
					required
				/>

				{error && <div className="error">{error}</div>}

				<button disabled={loading}>
					{loading
						? 'Please wait...'
						: mode === 'login'
							? 'Login'
							: 'Register'}
				</button>

				<button
					type="button"
					className="link-button"
					onClick={() => {
						setMode(mode === 'login' ? 'register' : 'login');
						setError('');
					}}
				>
					{mode === 'login'
						? 'Need an account? Register'
						: 'Already registered? Login'}
				</button>
			</form>
		</div>
	);
}