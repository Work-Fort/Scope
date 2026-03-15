import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

/**
 * `<wf-pagination>` — Page navigation with prev/next and numbered buttons.
 *
 * Custom implementation instead of extending LionPagination because:
 * - LionPagination uses Shadow DOM styles (`static get styles`)
 * - LionPagination uses LocalizeMixin/msgLit for i18n which requires Lion's localize infrastructure
 * - Our light DOM approach (createRenderRoot → this) is incompatible with Lion's slot-based rendering
 *
 * @element wf-pagination
 * @fires current-changed - When the current page changes
 */
export class WfPagination extends WfElement {
  @property({ type: Number, reflect: true }) current = 1;
  @property({ type: Number, reflect: true }) count = 0;

  private _visiblePages = 5;
  private _container: HTMLElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-pagination');
    this.setAttribute('role', 'navigation');
    this.setAttribute('aria-label', 'Pagination');
    this._ensureContainer();
    this._renderPages();
  }

  updated(): void {
    this._ensureContainer();
    this._renderPages();
  }

  next(): void {
    if (this.current < this.count) {
      this._goTo(this.current + 1);
    }
  }

  previous(): void {
    if (this.current > 1) {
      this._goTo(this.current - 1);
    }
  }

  first(): void {
    if (this.count >= 1) this._goTo(1);
  }

  last(): void {
    if (this.count >= 1) this._goTo(this.count);
  }

  goto(page: number): void {
    if (page >= 1 && page <= this.count) {
      this._goTo(page);
    }
  }

  private _goTo(page: number): void {
    if (page !== this.current) {
      this.current = page;
      this.dispatchEvent(new Event('current-changed'));
    }
  }

  private _ensureContainer(): void {
    if (!this._container) {
      this._container = document.createElement('div');
      this._container.style.display = 'contents';
      this.appendChild(this._container);
    }
  }

  private _calculateNavList(): (number | '...')[] {
    const start = 1;
    const finish = this.count;

    if (this.count > this._visiblePages + 2) {
      const pos = this.current;

      if (pos <= 4) {
        const list: (number | '...')[] = Array.from(
          { length: this._visiblePages },
          (_, i) => start + i,
        );
        list.push('...', this.count);
        return list;
      }

      if (finish - pos <= 3) {
        const list: (number | '...')[] = [1, '...'];
        for (let i = this.count - this._visiblePages + 1; i <= this.count; i++) {
          list.push(i);
        }
        return list;
      }

      return [start, '...', pos - 1, pos, pos + 1, '...', finish];
    }

    return Array.from({ length: finish }, (_, i) => i + 1);
  }

  private _renderPages(): void {
    if (!this._container) return;
    this._container.innerHTML = '';

    if (this.count <= 0) return;

    // Previous button
    const prevBtn = this._createButton('\u2039', 'Previous page', () => this.previous());
    if (this.current <= 1) prevBtn.disabled = true;
    this._container.appendChild(prevBtn);

    // Page buttons
    for (const page of this._calculateNavList()) {
      if (page === '...') {
        const ellipsis = document.createElement('span');
        ellipsis.className = 'wf-pagination__ellipsis';
        ellipsis.textContent = '\u2026';
        this._container.appendChild(ellipsis);
      } else {
        const btn = this._createButton(String(page), `Page ${page}`, () =>
          this._goTo(page as number),
        );
        if (page === this.current) {
          btn.setAttribute('aria-current', 'true');
        }
        this._container.appendChild(btn);
      }
    }

    // Next button
    const nextBtn = this._createButton('\u203A', 'Next page', () => this.next());
    if (this.current >= this.count) nextBtn.disabled = true;
    this._container.appendChild(nextBtn);
  }

  private _createButton(
    text: string,
    ariaLabel: string,
    handler: () => void,
  ): HTMLButtonElement {
    const btn = document.createElement('button');
    btn.className = 'wf-pagination__btn';
    btn.textContent = text;
    btn.setAttribute('aria-label', ariaLabel);
    btn.addEventListener('click', handler);
    return btn;
  }
}

customElements.define('wf-pagination', WfPagination);

declare global {
  interface HTMLElementTagNameMap {
    'wf-pagination': WfPagination;
  }
}
