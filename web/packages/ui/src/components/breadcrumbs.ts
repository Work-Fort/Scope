import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export interface BreadcrumbItem {
  label: string;
  href?: string;
}

/**
 * `<wf-breadcrumbs>` — Navigation breadcrumb trail.
 * Renders a list of links separated by a configurable separator.
 * The last item is rendered as plain text (current page).
 *
 * @element wf-breadcrumbs
 */
export class WfBreadcrumbs extends WfElement {
  @property({ type: String }) separator = '/';

  private _items: BreadcrumbItem[] = [];
  private _container: HTMLElement | null = null;

  get items(): BreadcrumbItem[] {
    return this._items;
  }

  set items(value: BreadcrumbItem[]) {
    this._items = value;
    this._renderItems();
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-breadcrumbs');
    this.setAttribute('role', 'navigation');
    this.setAttribute('aria-label', 'Breadcrumb');
    this._ensureContainer();
    this._renderItems();
  }

  updated(): void {
    this._ensureContainer();
    this._renderItems();
  }

  private _ensureContainer(): void {
    if (!this._container) {
      this._container = document.createElement('div');
      this._container.style.display = 'contents';
      this.appendChild(this._container);
    }
  }

  private _renderItems(): void {
    if (!this._container) return;
    this._container.innerHTML = '';

    const ol = document.createElement('ol');
    ol.className = 'wf-breadcrumbs__list';

    this._items.forEach((item, index) => {
      const li = document.createElement('li');
      li.className = 'wf-breadcrumbs__item';

      const isLast = index === this._items.length - 1;

      if (isLast) {
        const span = document.createElement('span');
        span.className = 'wf-breadcrumbs__current';
        span.setAttribute('aria-current', 'page');
        span.textContent = item.label;
        li.appendChild(span);
      } else {
        if (item.href) {
          const link = document.createElement('a');
          link.className = 'wf-breadcrumbs__link';
          link.href = item.href;
          link.textContent = item.label;
          link.addEventListener('click', (e) => {
            e.preventDefault();
            this.dispatchEvent(
              new CustomEvent('wf-navigate', {
                bubbles: true,
                composed: true,
                detail: { href: item.href, label: item.label, index },
              }),
            );
          });
          li.appendChild(link);
        } else {
          const span = document.createElement('span');
          span.className = 'wf-breadcrumbs__link';
          span.textContent = item.label;
          li.appendChild(span);
        }

        const sep = document.createElement('span');
        sep.className = 'wf-breadcrumbs__separator';
        sep.setAttribute('aria-hidden', 'true');
        sep.textContent = this.separator;
        li.appendChild(sep);
      }

      ol.appendChild(li);
    });

    this._container.appendChild(ol);
  }
}

customElements.define('wf-breadcrumbs', WfBreadcrumbs);

declare global {
  interface HTMLElementTagNameMap {
    'wf-breadcrumbs': WfBreadcrumbs;
  }
}
