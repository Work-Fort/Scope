import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export interface WfTableColumn {
  key: string;
  header: string;
  sortable?: boolean;
  width?: string;
  render?: (value: unknown, row: Record<string, unknown>) => unknown;
}

/**
 * `<wf-table>` -- Data table with sorting, pagination, and row click events.
 * Set `columns` and `data` properties to render.
 *
 * @element wf-table
 * @fires wf-row-click -- When a row is clicked. Detail: { row, index }.
 * @fires wf-sort -- When a column sort is triggered. Detail: { key, direction }.
 * @fires wf-page-change -- When pagination changes. Detail: { page }.
 */
export class WfTable extends WfElement {
  @property({ type: Array }) columns: WfTableColumn[] = [];
  @property({ type: Array }) data: Array<Record<string, unknown>> = [];
  @property({ type: Boolean, reflect: true }) striped = false;
  @property({ type: String, attribute: 'sort-key' }) sortKey = '';
  @property({ type: String, attribute: 'sort-direction' }) sortDirection: 'asc' | 'desc' | '' = '';
  @property({ type: Number }) page = 1;
  @property({ type: Number, attribute: 'page-size' }) pageSize = 10;
  @property({ type: Boolean, reflect: true }) paginate = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-table');
    this._syncClasses();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-table--striped', this.striped);
  }

  /** Sort data by a column key. Toggles asc/desc. */
  sort(key: string): void {
    if (this.sortKey === key) {
      this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
      this.sortKey = key;
      this.sortDirection = 'asc';
    }
    this.requestUpdate();
    this.dispatchEvent(
      new CustomEvent('wf-sort', {
        detail: { key: this.sortKey, direction: this.sortDirection },
        bubbles: true,
        composed: true,
      }),
    );
  }

  /** Navigate to a page (1-indexed). */
  goToPage(page: number): void {
    this.page = Math.max(1, Math.min(page, this.totalPages));
    this.requestUpdate();
    this.dispatchEvent(
      new CustomEvent('wf-page-change', {
        detail: { page: this.page },
        bubbles: true,
        composed: true,
      }),
    );
  }

  get totalPages(): number {
    if (!this.paginate || this.pageSize <= 0) return 1;
    return Math.ceil(this.data.length / this.pageSize);
  }

  /** Get the data to display -- sorted and paginated. */
  get displayData(): Array<Record<string, unknown>> {
    let result = [...this.data];

    // Sort
    if (this.sortKey && this.sortDirection) {
      result.sort((a, b) => {
        const aVal = a[this.sortKey];
        const bVal = b[this.sortKey];
        const cmp =
          typeof aVal === 'string' && typeof bVal === 'string'
            ? aVal.localeCompare(bVal)
            : Number(aVal) - Number(bVal);
        return this.sortDirection === 'desc' ? -cmp : cmp;
      });
    }

    // Paginate
    if (this.paginate && this.pageSize > 0) {
      const start = (this.page - 1) * this.pageSize;
      result = result.slice(start, start + this.pageSize);
    }

    return result;
  }

  private _handleRowClick(row: Record<string, unknown>, index: number): void {
    this.dispatchEvent(
      new CustomEvent('wf-row-click', {
        detail: { row, index },
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _handleSort(key: string): void {
    this.sort(key);
  }

  render() {
    const rows = this.displayData;

    return html`
      <table class="wf-table__table">
        <thead>
          <tr>
            ${this.columns.map(
              (col) => html`
                <th
                  class="wf-table__th ${col.sortable ? 'wf-table__th--sortable' : ''}"
                  style=${col.width ? `width: ${col.width}` : ''}
                  @click=${col.sortable ? () => this._handleSort(col.key) : null}
                >
                  ${col.header}
                  <span
                    class="wf-table__sort-icon"
                    ?hidden=${!(col.sortable && this.sortKey === col.key)}
                  >${this.sortDirection === 'asc' ? '\u25B2' : '\u25BC'}</span>
                </th>
              `,
            )}
          </tr>
        </thead>
        <tbody>
          ${rows.length > 0
            ? rows.map(
                (row, i) => html`
                  <tr
                    class="wf-table__row"
                    @click=${() => this._handleRowClick(row, i)}
                  >
                    ${this.columns.map(
                      (col) => html`
                        <td class="wf-table__td">
                          ${col.render
                            ? col.render(row[col.key], row)
                            : row[col.key]}
                        </td>
                      `,
                    )}
                  </tr>
                `,
              )
            : html`
                <tr>
                  <td class="wf-table__empty" colspan=${this.columns.length}>
                    No data
                  </td>
                </tr>
              `}
        </tbody>
      </table>
      <div
        class="wf-table__pagination"
        ?hidden=${!(this.paginate && this.totalPages > 1)}
      >
        <button
          class="wf-table__page-btn"
          ?disabled=${this.page <= 1}
          @click=${() => this.goToPage(this.page - 1)}
        >
          Previous
        </button>
        <span class="wf-table__page-info">
          Page ${this.page} of ${this.totalPages}
        </span>
        <button
          class="wf-table__page-btn"
          ?disabled=${this.page >= this.totalPages}
          @click=${() => this.goToPage(this.page + 1)}
        >
          Next
        </button>
      </div>
    `;
  }
}

customElements.define('wf-table', WfTable);

declare global {
  interface HTMLElementTagNameMap {
    'wf-table': WfTable;
  }
}
