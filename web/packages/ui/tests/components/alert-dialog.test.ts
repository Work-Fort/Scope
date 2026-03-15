import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/alert-dialog.js';
import type { WfAlertDialog } from '../../src/components/alert-dialog.js';

describe('WfAlertDialog', () => {
  afterEach(() => {
    cleanup();
    // Clean up any leftover event listeners
  });

  it('renders with wf-dialog class', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    expect(el.classList.contains('wf-dialog')).toBe(true);
  });

  it('starts closed (no overlay)', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    expect(el.open).toBe(false);
    expect(el.querySelector('.wf-dialog-overlay')).toBeNull();
  });

  it('show() opens the dialog and returns a promise', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({ message: 'Are you sure?' });
    expect(el.open).toBe(true);
    expect(el.querySelector('.wf-dialog-overlay')).not.toBeNull();

    // Close it to resolve the promise
    el.close(true);
    const result = await promise;
    expect(result.confirmed).toBe(true);
  });

  it('renders title when provided', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({ title: 'Confirm', message: 'Delete this?' });

    const title = el.querySelector('.wf-dialog__title');
    expect(title).not.toBeNull();
    expect(title!.textContent).toBe('Confirm');

    el.close(false);
    await promise;
  });

  it('renders message', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({ message: 'This is important' });

    const msg = el.querySelector('.wf-dialog__message');
    expect(msg!.textContent).toBe('This is important');

    el.close(false);
    await promise;
  });

  it('renders confirm and cancel buttons', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({
      message: 'Proceed?',
      confirmLabel: 'Yes',
      cancelLabel: 'No',
    });

    const buttons = el.querySelectorAll('.wf-dialog__btn');
    expect(buttons.length).toBe(2);
    expect(buttons[0].textContent).toBe('No');
    expect(buttons[1].textContent).toBe('Yes');

    el.close(false);
    await promise;
  });

  it('confirm button resolves with confirmed: true', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({ message: 'OK?' });

    const confirmBtn = el.querySelector('.wf-dialog__btn--primary') as HTMLButtonElement;
    confirmBtn.click();

    const result = await promise;
    expect(result.confirmed).toBe(true);
    expect(el.open).toBe(false);
  });

  it('cancel button resolves with confirmed: false', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({ message: 'OK?' });

    const buttons = el.querySelectorAll('.wf-dialog__btn');
    // First button is cancel
    (buttons[0] as HTMLButtonElement).click();

    const result = await promise;
    expect(result.confirmed).toBe(false);
  });

  it('alert() shows no cancel button', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.alert({ message: 'Info' });

    const buttons = el.querySelectorAll('.wf-dialog__btn');
    expect(buttons.length).toBe(1);
    expect(buttons[0].textContent).toBe('OK');

    el.close(true);
    await promise;
  });

  it('Escape key closes the dialog', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({ message: 'Test' });

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));

    const result = await promise;
    expect(result.confirmed).toBe(false);
    expect(el.open).toBe(false);
  });

  it('clicking overlay background closes dialog', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({ message: 'Test' });

    const overlay = el.querySelector('.wf-dialog-overlay') as HTMLElement;
    // Simulate click on overlay (not panel)
    overlay.dispatchEvent(new MouseEvent('click', { bubbles: true }));

    const result = await promise;
    expect(result.confirmed).toBe(false);
  });

  it('has alertdialog role and aria-modal', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({ title: 'Test', message: 'Content' });

    const panel = el.querySelector('.wf-dialog__panel');
    expect(panel!.getAttribute('role')).toBe('alertdialog');
    expect(panel!.getAttribute('aria-modal')).toBe('true');

    el.close(false);
    await promise;
  });

  it('applies danger variant', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show({ message: 'Delete?', variant: 'danger', confirmLabel: 'Delete' });

    const confirmBtn = el.querySelector('.wf-dialog__btn--danger');
    expect(confirmBtn).not.toBeNull();
    expect(confirmBtn!.textContent).toBe('Delete');

    el.close(false);
    await promise;
  });
});
