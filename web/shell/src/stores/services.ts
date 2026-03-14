import { createSignal } from 'solid-js';
import { fetchServices, type ServiceInfo, type Conflict, type ServicesResponse } from '../lib/api';
import { registerNewRemotes } from '../lib/remotes';
import { addBanner, removeBanner, banners } from './banners';
import { addToast } from './toasts';

const POLL_INTERVAL = 30_000;

const [serviceList, setServiceList] = createSignal<ServiceInfo[]>([]);
const [conflictList, setConflictList] = createSignal<Conflict[]>([]);
const [currentFort, setCurrentFort] = createSignal('');

let prevConnected = new Map<string, boolean>();

function handlePollResult(res: ServicesResponse): void {
  setCurrentFort(res.fort);
  setConflictList(res.conflicts ?? []);

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

  registerNewRemotes(res.fort, res.services);

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
let activeFort: string | null = null;

export function startPolling(fort: string): void {
  // If fort changed, reset state.
  if (activeFort !== fort) {
    stopPolling();
    prevConnected = new Map();
    setServiceList([]);
    setConflictList([]);
  }
  activeFort = fort;
  fetchServices(fort).then(handlePollResult).catch(console.error);
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
}

export const services = serviceList;
export const conflicts = conflictList;
export const fortName = currentFort;
