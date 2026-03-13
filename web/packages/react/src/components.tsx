import React, { forwardRef, useRef, useCallback } from 'react';
import '@workfort/ui';

import type {
  WfPanel, WfButton, WfBadge, WfStatusDot, WfSkeleton,
  WfTextInput, WfList, WfListItem, WfScrollArea, WfErrorFallback,
} from '@workfort/ui';

type WfProps<E, P = {}> = P & React.HTMLAttributes<E> & {
  children?: React.ReactNode;
  /** Allow custom event handler props like onWfClick, onWfChange, etc. */
  [key: `on${Uppercase<string>}${string}`]: ((...args: unknown[]) => void) | undefined;
};

/**
 * Separates event props (onX) from attribute props, and attaches event listeners
 * via addEventListener on the Custom Element ref. This is needed because React 18
 * does not forward onX props to Custom Element addEventListener calls.
 * React 19+ handles this natively, but we support React 18.
 */
function useWcEvents<E extends HTMLElement>(
  forwardedRef: React.ForwardedRef<E>,
  props: Record<string, unknown>,
): { ref: React.RefCallback<E>; cleanProps: Record<string, unknown> } {
  const innerRef = useRef<E | null>(null);
  const listenersRef = useRef<Map<string, EventListener>>(new Map());

  const cleanProps: Record<string, unknown> = {};
  const eventProps: Record<string, EventListener> = {};

  for (const [key, val] of Object.entries(props)) {
    if (key.startsWith('on') && key.length > 2 && typeof val === 'function') {
      // Convert onWfClick -> wf-click (camelCase to kebab-case)
      const raw = key[2].toLowerCase() + key.slice(3);
      const eventName = raw.replace(/([A-Z])/g, '-$1').toLowerCase();
      eventProps[eventName] = val as EventListener;
    } else {
      cleanProps[key] = val;
    }
  }

  const refCallback = useCallback((node: E | null) => {
    // Clean up old listeners
    if (innerRef.current) {
      listenersRef.current.forEach((fn, name) => innerRef.current!.removeEventListener(name, fn));
      listenersRef.current.clear();
    }
    innerRef.current = node;
    // Attach new listeners
    if (node) {
      for (const [name, fn] of Object.entries(eventProps)) {
        node.addEventListener(name, fn);
        listenersRef.current.set(name, fn);
      }
    }
    // Forward ref
    if (typeof forwardedRef === 'function') forwardedRef(node);
    else if (forwardedRef) forwardedRef.current = node;
  }, [forwardedRef, ...Object.keys(eventProps)]);

  return { ref: refCallback, cleanProps };
}

function wrapWc<E extends HTMLElement, P extends Record<string, unknown>>(
  tag: string,
  displayName: string,
) {
  const Comp = forwardRef<E, WfProps<E, P>>(({ children, ...rest }, ref) => {
    const { ref: wcRef, cleanProps } = useWcEvents<E>(ref, rest as Record<string, unknown>);
    return React.createElement(tag, { ref: wcRef, ...cleanProps }, children as React.ReactNode);
  });
  Comp.displayName = displayName;
  return Comp;
}

export const Panel = wrapWc<WfPanel, { label?: string }>('wf-panel', 'Panel');
export const Button = wrapWc<WfButton, { variant?: 'text' | 'filled'; disabled?: boolean }>('wf-button', 'Button');
export const Badge = wrapWc<WfBadge, { count?: number }>('wf-badge', 'Badge');
export const StatusDot = wrapWc<WfStatusDot, { status?: string }>('wf-status-dot', 'StatusDot');
export const Skeleton = wrapWc<WfSkeleton, { width?: string; height?: string }>('wf-skeleton', 'Skeleton');
export const Divider = wrapWc<HTMLElement, {}>('wf-divider', 'Divider');
export const TextInput = wrapWc<WfTextInput, { placeholder?: string; value?: string; disabled?: boolean }>('wf-text-input', 'TextInput');
export const List = wrapWc<WfList, {}>('wf-list', 'List');
export const ListItem = wrapWc<WfListItem, { active?: boolean }>('wf-list-item', 'ListItem');
export const ScrollArea = wrapWc<WfScrollArea, {}>('wf-scroll-area', 'ScrollArea');
export const ErrorFallback = wrapWc<WfErrorFallback, { title?: string; message?: string }>('wf-error-fallback', 'ErrorFallback');
