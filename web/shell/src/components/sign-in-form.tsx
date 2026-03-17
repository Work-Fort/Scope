import { createSignal, Show, type Component } from 'solid-js';

interface SignInFormProps {
  fort: string;
  onComplete: () => void;
}

const SignInForm: Component<SignInFormProps> = (props) => {
  const [email, setEmail] = createSignal('');
  const [password, setPassword] = createSignal('');
  const [error, setError] = createSignal('');
  const [loading, setLoading] = createSignal(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    setError('');

    if (!email() || !password()) {
      setError('Email and password are required');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(`/forts/${props.fort}/api/auth/v1/sign-in/email`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: email(),
          password: password(),
        }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        setError(body.message ?? body.error ?? `Sign-in failed (${res.status})`);
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
        Sign In
      </h1>

      <Show when={error()}>
        <wf-banner variant="error" headline={error()} style="margin-bottom: var(--wf-space-md);" />
      </Show>

      <form on:submit={handleSubmit} style="display: flex; flex-direction: column; gap: var(--wf-space-md);">
        <label style={labelStyle}>
          Email
          <input type="email" required value={email()} on:input={(e: Event) => setEmail((e.target as HTMLInputElement).value)} style={inputStyle} />
        </label>
        <label style={labelStyle}>
          Password
          <input type="password" required value={password()} on:input={(e: Event) => setPassword((e.target as HTMLInputElement).value)} style={inputStyle} />
        </label>
        <button type="submit" disabled={loading()} style={`margin-top: var(--wf-space-sm); padding: var(--wf-space-sm) var(--wf-space-md); border-radius: var(--wf-radius-sm); border: 1px solid var(--wf-color-border); background: var(--wf-color-bg-elevated); color: var(--wf-color-text); font-family: inherit; font-size: var(--wf-text-sm); cursor: pointer;`}>
          {loading() ? 'Signing in...' : 'Sign In'}
        </button>
      </form>
    </div>
  );
};

export default SignInForm;
