import { createSignal, Show, type Component } from 'solid-js';

interface SetupFormProps {
  fort: string;
  onComplete: () => void;
}

const SetupForm: Component<SetupFormProps> = (props) => {
  const [email, setEmail] = createSignal('');
  const [username, setUsername] = createSignal('');
  const [password, setPassword] = createSignal('');
  const [confirm, setConfirm] = createSignal('');
  const [error, setError] = createSignal('');
  const [loading, setLoading] = createSignal(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    setError('');

    if (password() !== confirm()) {
      setError('Passwords do not match');
      return;
    }
    if (!email() || !username() || !password()) {
      setError('All fields are required');
      return;
    }

    setLoading(true);
    try {
      const signUpRes = await fetch(`/forts/${props.fort}/api/auth/v1/sign-up/email`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: email(),
          password: password(),
          name: username(),
          username: username(),
          displayName: username(),
        }),
      });

      if (!signUpRes.ok) {
        const body = await signUpRes.json().catch(() => ({}));
        setError(body.message ?? body.error ?? `Sign-up failed (${signUpRes.status})`);
        return;
      }

      const signInRes = await fetch(`/forts/${props.fort}/api/auth/v1/sign-in/email`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: email(),
          password: password(),
        }),
      });

      if (!signInRes.ok) {
        setError('Account created but sign-in failed. Please refresh and sign in.');
        return;
      }

      props.onComplete();
    } catch (err: any) {
      setError(err.message ?? 'Network error');
    } finally {
      setLoading(false);
    }
  }

  const inputStyle = "padding: var(--wf-space-sm); border-radius: var(--wf-radius-sm); border: 1px solid var(--wf-color-border); background: var(--wf-color-bg); color: var(--wf-color-text); font-family: inherit; font-size: var(--wf-text-sm); outline: none; box-sizing: border-box; width: 100%;";
  const labelStyle = "display: flex; flex-direction: column; gap: var(--wf-space-xs); font-size: var(--wf-text-sm); color: var(--wf-color-text-secondary);";

  return (
    <div style="max-width: 24rem; margin: 4rem auto; padding: var(--wf-space-lg);">
      <h1 style="font-size: var(--wf-text-lg); font-weight: var(--wf-weight-semibold); margin-bottom: var(--wf-space-md); color: var(--wf-color-text);">
        Create Admin Account
      </h1>
      <p style="font-size: var(--wf-text-sm); color: var(--wf-color-text-secondary); margin-bottom: var(--wf-space-lg);">
        Set up the first account to get started with WorkFort.
      </p>

      <Show when={error()}>
        <wf-banner variant="error" headline={error()} style="margin-bottom: var(--wf-space-md);" />
      </Show>

      <form on:submit={handleSubmit} style="display: flex; flex-direction: column; gap: var(--wf-space-md);">
        <label style={labelStyle}>
          Email
          <input type="email" required value={email()} on:input={(e: Event) => setEmail((e.target as HTMLInputElement).value)} style={inputStyle} />
        </label>
        <label style={labelStyle}>
          Username
          <input type="text" required value={username()} on:input={(e: Event) => setUsername((e.target as HTMLInputElement).value)} style={inputStyle} />
        </label>
        <label style={labelStyle}>
          Password
          <input type="password" required value={password()} on:input={(e: Event) => setPassword((e.target as HTMLInputElement).value)} style={inputStyle} />
        </label>
        <label style={labelStyle}>
          Confirm Password
          <input type="password" required value={confirm()} on:input={(e: Event) => setConfirm((e.target as HTMLInputElement).value)} style={inputStyle} />
        </label>
        <wf-button type="submit" disabled={loading()} style="margin-top: var(--wf-space-sm);">
          {loading() ? 'Creating...' : 'Create Account'}
        </wf-button>
      </form>
    </div>
  );
};

export default SetupForm;
