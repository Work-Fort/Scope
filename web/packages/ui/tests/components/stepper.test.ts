import { describe, it, expect, afterEach, vi } from 'vitest';
import { cleanup } from '../helpers.js';
import '../../src/components/stepper.js';
import type { WfStepper, WfStep } from '../../src/components/stepper.js';

async function createStepper(current = 0): Promise<WfStepper> {
  const stepper = document.createElement('wf-stepper') as WfStepper;
  stepper.setAttribute('current', String(current));

  const step1 = document.createElement('wf-step') as WfStep;
  step1.label = 'Account';
  const step2 = document.createElement('wf-step') as WfStep;
  step2.label = 'Details';
  const step3 = document.createElement('wf-step') as WfStep;
  step3.label = 'Confirm';

  stepper.appendChild(step1);
  stepper.appendChild(step2);
  stepper.appendChild(step3);

  document.body.appendChild(stepper);
  await stepper.updateComplete;
  // Wait for steps to connect
  await Promise.all([step1.updateComplete, step2.updateComplete, step3.updateComplete]);
  return stepper;
}

describe('WfStepper', () => {
  afterEach(cleanup);

  it('renders with wf-stepper class', async () => {
    const el = await createStepper();
    expect(el.classList.contains('wf-stepper')).toBe(true);
  });

  it('finds child steps', async () => {
    const el = await createStepper();
    expect(el.steps.length).toBe(3);
  });

  it('marks first step as active by default', async () => {
    const el = await createStepper();
    expect(el.steps[0].status).toBe('active');
    expect(el.steps[1].status).toBe('untouched');
    expect(el.steps[2].status).toBe('untouched');
  });

  it('marks previous steps as complete', async () => {
    const el = await createStepper(1);
    expect(el.steps[0].status).toBe('complete');
    expect(el.steps[1].status).toBe('active');
    expect(el.steps[2].status).toBe('untouched');
  });

  it('navigates via next() and previous()', async () => {
    const el = await createStepper();
    el.next();
    expect(el.current).toBe(1);
    expect(el.steps[0].status).toBe('complete');
    expect(el.steps[1].status).toBe('active');

    el.previous();
    expect(el.current).toBe(0);
    expect(el.steps[0].status).toBe('active');
  });

  it('does not go below 0 or above max', async () => {
    const el = await createStepper();
    el.previous();
    expect(el.current).toBe(0);

    el.goto(2);
    el.next();
    expect(el.current).toBe(2);
  });

  it('fires wf-step-change event', async () => {
    const el = await createStepper();
    const handler = vi.fn();
    el.addEventListener('wf-step-change', handler);
    el.next();
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ from: 0, to: 1 });
  });

  it('renders step indicators', async () => {
    const el = await createStepper();
    const indicators = el.querySelectorAll('.wf-step__indicator');
    expect(indicators.length).toBe(3);
  });

  it('renders connectors between steps', async () => {
    const el = await createStepper();
    const connectors = el.querySelectorAll('.wf-step__connector');
    expect(connectors.length).toBe(2);
  });

  it('applies vertical orientation', async () => {
    const el = await createStepper();
    el.orientation = 'vertical';
    await el.updateComplete;
    expect(el.classList.contains('wf-stepper--vertical')).toBe(true);
  });
});

describe('WfStep', () => {
  afterEach(cleanup);

  it('renders with wf-step class', async () => {
    const step = document.createElement('wf-step') as WfStep;
    document.body.appendChild(step);
    await step.updateComplete;
    expect(step.classList.contains('wf-step')).toBe(true);
  });

  it('applies status classes', async () => {
    const step = document.createElement('wf-step') as WfStep;
    document.body.appendChild(step);
    await step.updateComplete;
    expect(step.classList.contains('wf-step--untouched')).toBe(true);

    step.status = 'active';
    await step.updateComplete;
    expect(step.classList.contains('wf-step--active')).toBe(true);
    expect(step.classList.contains('wf-step--untouched')).toBe(false);
  });
});
