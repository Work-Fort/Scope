import { createSignal } from 'solid-js';
import { fetchServices, type ServiceInfo, type Conflict, type ServicesResponse } from '../lib/api';
import { registerNewRemotes } from '../lib/remotes';
import { addBanner, removeBanner, banners } from './banners';
import { addToast } from './toasts';

const POLL_INTERVAL = 30_000;

const [serviceList, setServiceList] = createSignal<ServiceInfo[]>([]);
const [conflictList, setConflictList] = createSignal<Conflict[]>([]);
const [fort, setFort] = createSignal('');

let prevConnected = new Map<string, boolean>();

function handlePollResult(res: ServicesResponse): void {
  setFort(res.fort);
  setConflictList(res.conflicts ?? []);

  // Detect state transitions for toasts.
  const nextConnected = new Map<string, boolean>();
  for (const svc of res.services) {
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

  // Register any newly discovered remotes.
  registerNewRemotes(res.services);

  // Update conflict banners — add new, remove stale.
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

  for (const svc of res.services) {
    if (svc.connected) {
      removeBanner(`disconnected:${svc.name}`);
    }
  }

  setServiceList(res.services);
}

let intervalId: ReturnType<typeof setInterval> | null = null;

export function startPolling(): void {
  fetchServices().then(handlePollResult).catch(console.error);
  intervalId = setInterval(() => {
    fetchServices().then(handlePollResult).catch(console.error);
  }, POLL_INTERVAL);
}

export function stopPolling(): void {
  if (intervalId) {
    clearInterval(intervalId);
    intervalId = null;
  }
}

export const services = serviceList;
export const conflicts = conflictList;
export const fortName = fort;
