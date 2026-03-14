import { createSignal } from 'solid-js';

export interface BannerEntry {
  key: string;
  variant: 'error' | 'warning' | 'info';
  headline: string;
  details?: string;
  source: 'system' | 'app';
}

const [banners, setBanners] = createSignal<BannerEntry[]>([]);
const dismissed = new Set<string>();

export { banners };

export function addBanner(
  key: string,
  variant: BannerEntry['variant'],
  headline: string,
  details?: string,
  source: BannerEntry['source'] = 'app',
): void {
  setBanners((prev) => {
    if (prev.find((b) => b.key === key)) return prev;
    if (dismissed.has(key)) return prev;
    return [...prev, { key, variant, headline, details, source }];
  });
}

export function removeBanner(key: string): void {
  dismissed.delete(key);
  setBanners((prev) => prev.filter((b) => b.key !== key));
}

export function dismissBanner(key: string): void {
  dismissed.add(key);
  setBanners((prev) => prev.filter((b) => b.key !== key));
}

export function isBannerDismissed(key: string): boolean {
  return dismissed.has(key);
}

/** System banners first, then app banners. Errors before warnings. */
export function sortedBanners(): BannerEntry[] {
  const variantOrder = { error: 0, warning: 1, info: 2 };
  return banners()
    .filter((b) => !dismissed.has(b.key))
    .sort((a, b) => {
      if (a.source !== b.source) return a.source === 'system' ? -1 : 1;
      return variantOrder[a.variant] - variantOrder[b.variant];
    });
}
