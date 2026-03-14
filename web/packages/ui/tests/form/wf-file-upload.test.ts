import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/form/wf-file-upload.js';
import type { WfFileUpload } from '../../src/form/wf-file-upload.js';

describe('WfFileUpload', () => {
  afterEach(cleanup);

  it('renders with wf-file-upload class', async () => {
    const el = await fixture<WfFileUpload>('wf-file-upload');
    expect(el.classList.contains('wf-file-upload')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfFileUpload>('wf-file-upload');
    expect(el.shadowRoot).toBeNull();
  });

  it('has hidden file input', async () => {
    const el = await fixture<WfFileUpload>('wf-file-upload');
    const input = el.querySelector('input[type="file"]') as HTMLInputElement;
    expect(input).not.toBeNull();
    expect(input.classList.contains('wf-file-upload__input')).toBe(true);
  });

  it('accept attribute passes through', async () => {
    const el = await fixture<WfFileUpload>('wf-file-upload', { accept: '.png,.jpg' });
    const input = el.querySelector('input[type="file"]') as HTMLInputElement;
    expect(input.accept).toBe('.png,.jpg');
  });

  it('multiple attribute passes through', async () => {
    const el = await fixture<WfFileUpload>('wf-file-upload', { multiple: true });
    const input = el.querySelector('input[type="file"]') as HTMLInputElement;
    expect(input.multiple).toBe(true);
  });

  it('has drop zone element', async () => {
    const el = await fixture<WfFileUpload>('wf-file-upload');
    const dropZone = el.querySelector('.wf-file-upload__drop-zone');
    expect(dropZone).not.toBeNull();
  });

  it('disabled state applies class and disables input', async () => {
    const el = await fixture<WfFileUpload>('wf-file-upload');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-file-upload--disabled')).toBe(true);
    const input = el.querySelector('input[type="file"]') as HTMLInputElement;
    expect(input.disabled).toBe(true);
  });
});
