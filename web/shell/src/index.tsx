import 'virtual:uno.css';
import '@workfort/ui';
import '@workfort/ui/style.css';
import './global.css';
import { render } from 'solid-js/web';
import App from './app';

render(() => <App />, document.getElementById('app')!);
