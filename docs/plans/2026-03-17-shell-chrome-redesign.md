# Shell Chrome Redesign — Plan 9

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Redesign the shell navigation bar for horizontal tabs, hamburger collapse on narrow viewports, configurable button positioning for handedness, and responsive layouts.

**Architecture:** New `wf-nav-bar` and `wf-hamburger` web components in `@workfort/ui`. The shell replaces its current nav-bar.tsx with these components. Responsive CSS handles the grid collapse. A `handedness` preference controls button positions.

**Tech Stack:** Lit, CSS `@media` queries, `ResizeObserver`, `@workfort/ui`

**Repo:** `scope/lead`

**Prerequisites:** Plans 5-6 (extraction) should be complete for consistency, but this plan is technically independent.

---

### Task 1: `wf-hamburger` Web Component

**Files:**
- Create: `web/packages/ui/src/layout/wf-hamburger.ts`
- Create: `web/packages/ui/src/styles/hamburger.css`
- Create: `web/packages/ui/tests/layout/hamburger.test.ts`
- Modify: `web/packages/ui/src/index.ts`
- Modify: `web/packages/ui/src/styles/index.css`

A hamburger button that opens a slide-out panel. Position is configurable.

Properties:
- `position`: `'top-left' | 'top-right' | 'bottom-left' | 'bottom-right'` (default: `'top-right'`)
- `open`: `boolean`

Events:
- `wf-toggle`: fired when button is clicked (detail: `{ open: boolean }`)

The button renders a ☰ icon. When open, a full-screen overlay panel slides in from the relevant edge. Content goes in the default slot.

CSS uses `position: fixed` with the corner determined by the `position` attribute. The panel slides in from the nearest edge.

**Tests:**
- Renders hamburger icon
- Dispatches wf-toggle on click
- Positions in correct corner based on position prop
- Shows/hides overlay panel based on open prop

**Commit:** `feat(@workfort/ui): add wf-hamburger component with configurable position`

---

### Task 2: `wf-nav-bar` Web Component

**Files:**
- Create: `web/packages/ui/src/layout/wf-nav-bar.ts`
- Create: `web/packages/ui/src/styles/nav-bar.css`
- Create: `web/packages/ui/tests/layout/nav-bar.test.ts`
- Modify: `web/packages/ui/src/index.ts`
- Modify: `web/packages/ui/src/styles/index.css`

A responsive navigation bar with overflow detection.

Properties:
- `breakpoint`: `number` (default: `640`) — width at which to collapse to hamburger
- `hamburger-position`: `string` (default: `'top-right'`) — forwarded to internal `wf-hamburger`

Slots:
- `brand` — left-aligned brand content
- Default slot — tab items
- `actions` — right-aligned action buttons (theme toggle, etc.)

Behavior:
- At full width: brand + horizontal tabs + actions
- Uses `ResizeObserver` on the tabs container to detect overflow
- When tabs overflow: visible tabs + "more" dropdown for the rest
- Below `breakpoint`: hamburger only, all content in the hamburger panel
- Brand hides below a second, narrower breakpoint

**Tests:**
- Renders brand, tabs, and actions slots
- Shows hamburger below breakpoint
- Hides brand on very narrow viewport

**Commit:** `feat(@workfort/ui): add wf-nav-bar component with overflow collapse`

---

### Task 3: Update Shell to Use New Components

**Files:**
- Modify: `web/shell/src/components/nav-bar.tsx`
- Modify: `web/shell/src/global.css`
- Modify: `web/shell/src/stores/services.ts` (filter non-UI services)

**nav-bar.tsx:**

Replace the current implementation with `<wf-nav-bar>`:

```tsx
const NavBar: Component = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams<{ fort: string }>();
  const theme = useTheme();

  // Only show services with UI.
  const visibleServices = () => services().filter((s) => s.enabled && s.ui);

  return (
    <wf-nav-bar hamburger-position={handedness() === 'left' ? 'top-left' : 'top-right'}>
      <span slot="brand" class="shell-nav__brand">{fortName() || 'WorkFort'}</span>

      <For each={visibleServices()}>
        {(svc) => (
          <wf-list-item
            active={location.pathname.includes(svc.route)}
            on:wf-select={() => navigate(`/forts/${params.fort}${svc.route}`)}
          >
            <wf-status-dot status={svc.connected ? 'online' : 'offline'} />
            {svc.label}
          </wf-list-item>
        )}
      </For>

      <div slot="actions">
        <wf-button variant="text" on:wf-click={() => toggleTheme()}>
          {theme() === 'dark' ? 'Light' : 'Dark'}
        </wf-button>
      </div>
    </wf-nav-bar>
  );
};
```

Key change: `services().filter((s) => s.enabled && s.ui)` — Auth and other non-UI services are hidden.

**global.css:**

Update the grid for responsive behavior:

```css
@media (max-width: 640px) {
  .shell-layout {
    grid-template-columns: 1fr;
    grid-template-areas:
      "banners"
      "nav"
      "content";
  }
  .shell-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: 280px;
    z-index: 100;
    background: var(--wf-color-bg);
    transform: translateX(-100%);
    transition: transform 0.2s ease;
  }
  .shell-sidebar--open {
    transform: translateX(0);
  }
  .shell-content {
    padding: var(--wf-space-md);
  }
}
```

**Commit:** `feat(shell): use wf-nav-bar, hide non-UI services, add responsive grid`

---

### Task 4: Handedness Preference

**Files:**
- Modify: `web/shell/src/stores/theme.ts` (add handedness)
- Modify: `web/shell/src/components/nav-bar.tsx` (use handedness)

Add to the theme store:

```typescript
const [handedness, setHandedness] = createSignal<'left' | 'right'>(
  (localStorage.getItem('wf-handedness') as 'left' | 'right') || 'right',
);

function toggleHandedness(): void {
  const next = handedness() === 'right' ? 'left' : 'right';
  setHandedness(next);
  localStorage.setItem('wf-handedness', next);
}

export { handedness, toggleHandedness };
```

Add a "Switch hand" option in the hamburger menu.

The `wf-hamburger` position follows the handedness. The Sharkfin sidebar toggle (future) will also follow it.

**Commit:** `feat(shell): add handedness preference for navigation controls`

---

### Task 5: Sharkfin Sidebar Toggle on Mobile

**Files:**
- Modify: `web/src/components/chat.tsx` (sharkfin repo)
- Modify: `web/src/index.tsx` (sharkfin repo)

On mobile (when the shell sidebar is an overlay), Sharkfin needs a toggle button to open/close it.

This is a SolidJS component that renders a small floating button. Its position follows the handedness preference (read from a CSS custom property set by the shell, or from the same localStorage key).

```tsx
function SidebarToggle(props: { onClick: () => void }) {
  const pos = localStorage.getItem('wf-handedness') === 'left' ? 'left' : 'right';
  return (
    <button
      class="sf-sidebar-toggle"
      style={`position: fixed; bottom: var(--wf-space-lg); ${pos}: var(--wf-space-lg); z-index: 99;`}
      on:click={props.onClick}
    >
      ☰
    </button>
  );
}
```

Only rendered on mobile (via CSS media query or a signal that checks viewport width).

**Commit:** `feat(sharkfin): add mobile sidebar toggle with handedness support`
