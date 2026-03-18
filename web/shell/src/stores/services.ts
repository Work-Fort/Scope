import { createSignal } from 'solid-js';
import { fetchServices, checkSession, type ServiceInfo, type Conflict, type ServicesResponse } from '../lib/api';
import { registerNewRemotes } from '../lib/remotes';
import { addBanner, removeBanner, banners } from './banners';
import { addToast } from './toasts';
import { addNotification, fetchNotifications } from './notifications';

const POLL_INTERVAL = 30_000;
const WS_RECONNECT_DELAY = 5_000;

const [serviceList, setServiceList] = createSignal<ServiceInfo[]>([]);
const [conflictList, setConflictList] = createSignal<Conflict[]>([]);
const [currentFort, setCurrentFort] = createSignal('');
const [setupMode, setSetupMode] = createSignal(false);

// Auth state: starts true (assume unauthenticated), cleared after
// successful sign-in or if a BFF-protected probe succeeds.
const [needsAuth, setNeedsAuth] = createSignal(true);

/** Called by the sign-in/setup forms after successful authentication. */
export function clearAuthRequired(): void {
  setNeedsAuth(false);
}

let prevConnected = new Map<string, boolean>();

function handleServiceUpdate(services: ServiceInfo[], fort?: string): void {
  if (fort) setCurrentFort(fort);

  const nextConnected = new Map<string, boolean>();
  for (const svc of services) {
    nextConnected.set(svc.name, svc.connected);
    const was = prevConnected.get(svc.name);
    if (was !== undefined && was !== svc.connected) {
      if (svc.connected) {
        addToast('success', `${svc.label} reconnected`);
        removeBanner(`disconnected:${svc.name}`);
      } else {
        addToast('error', `${svc.label} disconnected`, { sticky: true });
        addBanner(
          `disconnected:${svc.name}`,
          'warning',
          `${svc.label} is not responding`,
          `Service "${svc.name}" is unreachable. This page will update when it recovers.`,
          'system',
        );
      }
    }
  }
  prevConnected = nextConnected;

  for (const svc of services) {
    if (svc.connected) {
      removeBanner(`disconnected:${svc.name}`);
    }
  }

  setServiceList(services);

  const authSvc = services.find((s) => s.name === 'auth');
  setSetupMode(authSvc?.setup_mode === true);

  if (authSvc?.setup_mode) {
    setNeedsAuth(false);
  }
}

function handlePollResult(res: ServicesResponse): void {
  setConflictList(res.conflicts ?? []);

  const activeConflictKeys = new Set((res.conflicts ?? []).map((c) => `conflict:${c.name}`));
  for (const conflict of res.conflicts ?? []) {
    addBanner(
      `conflict:${conflict.name}`,
      'error',
      `Service conflict: "${conflict.name}"`,
      `${conflict.reason}\nURL: ${conflict.url}`,
      'system',
    );
  }
  for (const b of banners()) {
    if (b.key.startsWith('conflict:') && !activeConflictKeys.has(b.key)) {
      removeBanner(b.key);
    }
  }

  // Map TrackedService fields to ServiceInfo shape
  const services: ServiceInfo[] = res.services.map((s) => ({
    ...s,
    enabled: s.connected || s.ui,
  }));

  registerNewRemotes(res.fort, services);
  handleServiceUpdate(services, res.fort);
}

// --- Shell WebSocket ---

let shellWs: WebSocket | null = null;
let wsReconnectTimer: ReturnType<typeof setTimeout> | null = null;
let wsActiveFort: string | null = null;

function connectShellWs(fort: string): void {
  if (shellWs) {
    shellWs.onclose = null;
    shellWs.close();
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/ws/shell`;

  shellWs = new WebSocket(wsUrl);

  shellWs.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);

      switch (msg.type) {
        case 'services_changed': {
          const services: ServiceInfo[] = (msg.data ?? []).map((s: any) => ({
            ...s,
            enabled: s.connected || s.ui,
          }));
          registerNewRemotes(fort, services);
          handleServiceUpdate(services);
          break;
        }
        case 'notification':
          addNotification(msg.data);
          break;
      }
    } catch {
      // Ignore malformed messages.
    }
  };

  shellWs.onclose = () => {
    shellWs = null;
    if (wsActiveFort === fort) {
      wsReconnectTimer = setTimeout(() => connectShellWs(fort), WS_RECONNECT_DELAY);
    }
  };

  shellWs.onerror = () => {
    // onclose fires after onerror, reconnect handled there.
  };
}

function disconnectShellWs(): void {
  wsActiveFort = null;
  if (wsReconnectTimer) {
    clearTimeout(wsReconnectTimer);
    wsReconnectTimer = null;
  }
  if (shellWs) {
    shellWs.onclose = null;
    shellWs.close();
    shellWs = null;
  }
}

// --- Polling (fallback for initial load + conflict detection) ---

let intervalId: ReturnType<typeof setInterval> | null = null;
let activeFort: string | null = null;

export function startPolling(fort: string): void {
  // If fort changed, reset state.
  if (activeFort !== fort) {
    stopPolling();
    prevConnected = new Map();
    setServiceList([]);
    setConflictList([]);
    setNeedsAuth(true);
  }
  activeFort = fort;

  // Initial HTTP fetch for services + conflicts
  fetchServices(fort).then((res) => {
    handlePollResult(res);

    // Probe for existing session if not in setup mode.
    if (!setupMode()) {
      checkSession(fort).then((authenticated) => {
        if (authenticated) setNeedsAuth(false);
      });
    }
  }).catch(console.error);

  // Fetch initial notifications
  fetchNotifications().catch(() => {});

  // Connect the shell WebSocket for live updates
  wsActiveFort = fort;
  connectShellWs(fort);

  // Keep HTTP polling at a slower interval as fallback for conflict detection
  intervalId = setInterval(() => {
    fetchServices(fort).then(handlePollResult).catch(console.error);
  }, POLL_INTERVAL);
}

export function stopPolling(): void {
  if (intervalId) {
    clearInterval(intervalId);
    intervalId = null;
  }
  activeFort = null;
  disconnectShellWs();
}

export const services = serviceList;
export const conflicts = conflictList;
export const fortName = currentFort;
export const isSetupMode = setupMode;
export const isAuthRequired = needsAuth;
