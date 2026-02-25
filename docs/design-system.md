# Sky Flux CMS Design System

> Design reference: [stagehand.dev](https://www.stagehand.dev/) warm amber style
> Theme engine: shadcn/ui + Tailwind CSS V4 + OKLCH CSS Variables

## Color System

Based on shadcn **Orange** theme with custom amber tint, adapted for CMS admin panel.
All colors use **OKLCH** color space for perceptually uniform lightness.

### CSS Variables (global.css)

```css
@import "tailwindcss";
@import "tw-animate-css";
@import "shadcn/tailwind.css";

@custom-variant dark (&:is(.dark *));

/* ─── Tailwind V4 Bridge ─── */
@theme inline {
    --radius-sm: calc(var(--radius) - 4px);
    --radius-md: calc(var(--radius) - 2px);
    --radius-lg: var(--radius);
    --radius-xl: calc(var(--radius) + 4px);
    --radius-2xl: calc(var(--radius) + 8px);

    /* Semantic Colors → Tailwind Utilities */
    --color-background: var(--background);
    --color-foreground: var(--foreground);
    --color-card: var(--card);
    --color-card-foreground: var(--card-foreground);
    --color-popover: var(--popover);
    --color-popover-foreground: var(--popover-foreground);
    --color-primary: var(--primary);
    --color-primary-foreground: var(--primary-foreground);
    --color-secondary: var(--secondary);
    --color-secondary-foreground: var(--secondary-foreground);
    --color-muted: var(--muted);
    --color-muted-foreground: var(--muted-foreground);
    --color-accent: var(--accent);
    --color-accent-foreground: var(--accent-foreground);
    --color-destructive: var(--destructive);
    --color-border: var(--border);
    --color-input: var(--input);
    --color-ring: var(--ring);
    --color-chart-1: var(--chart-1);
    --color-chart-2: var(--chart-2);
    --color-chart-3: var(--chart-3);
    --color-chart-4: var(--chart-4);
    --color-chart-5: var(--chart-5);
    --color-sidebar: var(--sidebar);
    --color-sidebar-foreground: var(--sidebar-foreground);
    --color-sidebar-primary: var(--sidebar-primary);
    --color-sidebar-primary-foreground: var(--sidebar-primary-foreground);
    --color-sidebar-accent: var(--sidebar-accent);
    --color-sidebar-accent-foreground: var(--sidebar-accent-foreground);
    --color-sidebar-border: var(--sidebar-border);
    --color-sidebar-ring: var(--sidebar-ring);

    /* Custom: Warning & Success */
    --color-warning: var(--warning);
    --color-warning-foreground: var(--warning-foreground);
    --color-success: var(--success);
    --color-success-foreground: var(--success-foreground);
}

/* ─── Light Mode ─── */
:root {
    --radius: 0.625rem;

    /* Base */
    --background: oklch(1 0 0);
    --foreground: oklch(0.141 0.005 285.823);

    /* Card */
    --card: oklch(1 0 0);
    --card-foreground: oklch(0.141 0.005 285.823);

    /* Popover */
    --popover: oklch(1 0 0);
    --popover-foreground: oklch(0.141 0.005 285.823);

    /* Primary — Amber/Orange (Stagehand-inspired) */
    --primary: oklch(0.646 0.222 41.116);
    --primary-foreground: oklch(0.98 0.016 73.684);

    /* Secondary — Neutral gray */
    --secondary: oklch(0.967 0.001 286.375);
    --secondary-foreground: oklch(0.21 0.006 285.885);

    /* Muted */
    --muted: oklch(0.967 0.001 286.375);
    --muted-foreground: oklch(0.552 0.016 285.938);

    /* Accent */
    --accent: oklch(0.967 0.001 286.375);
    --accent-foreground: oklch(0.21 0.006 285.885);

    /* Destructive */
    --destructive: oklch(0.577 0.245 27.325);

    /* Borders & Input */
    --border: oklch(0.92 0.004 286.32);
    --input: oklch(0.92 0.004 286.32);
    --ring: oklch(0.75 0.183 55.934);

    /* Charts — Orange gradient scale */
    --chart-1: oklch(0.837 0.128 66.29);
    --chart-2: oklch(0.705 0.213 47.604);
    --chart-3: oklch(0.646 0.222 41.116);
    --chart-4: oklch(0.553 0.195 38.402);
    --chart-5: oklch(0.47 0.157 37.304);

    /* Sidebar */
    --sidebar: oklch(0.985 0 0);
    --sidebar-foreground: oklch(0.141 0.005 285.823);
    --sidebar-primary: oklch(0.646 0.222 41.116);
    --sidebar-primary-foreground: oklch(0.98 0.016 73.684);
    --sidebar-accent: oklch(0.967 0.001 286.375);
    --sidebar-accent-foreground: oklch(0.21 0.006 285.885);
    --sidebar-border: oklch(0.92 0.004 286.32);
    --sidebar-ring: oklch(0.75 0.183 55.934);

    /* Custom Semantic */
    --warning: oklch(0.852 0.199 91.936);
    --warning-foreground: oklch(0.421 0.095 57.708);
    --success: oklch(0.723 0.191 149.579);
    --success-foreground: oklch(0.985 0 0);
}

/* ─── Dark Mode ─── */
.dark {
    /* Base */
    --background: oklch(0.141 0.005 285.823);
    --foreground: oklch(0.985 0 0);

    /* Card */
    --card: oklch(0.21 0.006 285.885);
    --card-foreground: oklch(0.985 0 0);

    /* Popover */
    --popover: oklch(0.21 0.006 285.885);
    --popover-foreground: oklch(0.985 0 0);

    /* Primary — Slightly lighter orange in dark mode */
    --primary: oklch(0.705 0.213 47.604);
    --primary-foreground: oklch(0.98 0.016 73.684);

    /* Secondary */
    --secondary: oklch(0.274 0.006 286.033);
    --secondary-foreground: oklch(0.985 0 0);

    /* Muted */
    --muted: oklch(0.274 0.006 286.033);
    --muted-foreground: oklch(0.705 0.015 286.067);

    /* Accent */
    --accent: oklch(0.274 0.006 286.033);
    --accent-foreground: oklch(0.985 0 0);

    /* Destructive */
    --destructive: oklch(0.704 0.191 22.216);

    /* Borders & Input */
    --border: oklch(1 0 0 / 10%);
    --input: oklch(1 0 0 / 15%);
    --ring: oklch(0.408 0.123 38.172);

    /* Charts — Orange gradient scale (adjusted for dark bg) */
    --chart-1: oklch(0.837 0.128 66.29);
    --chart-2: oklch(0.705 0.213 47.604);
    --chart-3: oklch(0.646 0.222 41.116);
    --chart-4: oklch(0.553 0.195 38.402);
    --chart-5: oklch(0.47 0.157 37.304);

    /* Sidebar */
    --sidebar: oklch(0.21 0.006 285.885);
    --sidebar-foreground: oklch(0.985 0 0);
    --sidebar-primary: oklch(0.705 0.213 47.604);
    --sidebar-primary-foreground: oklch(0.98 0.016 73.684);
    --sidebar-accent: oklch(0.274 0.006 286.033);
    --sidebar-accent-foreground: oklch(0.985 0 0);
    --sidebar-border: oklch(1 0 0 / 10%);
    --sidebar-ring: oklch(0.408 0.123 38.172);

    /* Custom Semantic */
    --warning: oklch(0.795 0.184 86.047);
    --warning-foreground: oklch(0.421 0.095 57.708);
    --success: oklch(0.723 0.191 149.579);
    --success-foreground: oklch(0.985 0 0);
}

@layer base {
    * {
        @apply border-border outline-ring/50;
    }
    body {
        @apply bg-background text-foreground;
    }
}
```

---

## Color Palette Reference

### Semantic Roles

| Role | Light | Dark | Usage |
|------|-------|------|-------|
| `primary` | Amber-orange | Lighter amber | CTA buttons, active nav, focus rings |
| `secondary` | Light gray | Dark gray | Secondary buttons, tags |
| `muted` | Light gray | Dark gray | Disabled states, subtle backgrounds |
| `accent` | Light gray | Dark gray | Hover states, highlights |
| `destructive` | Red | Lighter red | Delete buttons, error states |
| `warning` | Yellow | Lighter yellow | Warning alerts, caution badges |
| `success` | Green | Green | Success messages, publish status |

### Chart Colors (Dashboard)

5-step amber gradient from light to dark, consistent across both modes:

| Token | OKLCH | Visual Use |
|-------|-------|------------|
| `chart-1` | `oklch(0.837 0.128 66.29)` | Lightest — area fills |
| `chart-2` | `oklch(0.705 0.213 47.604)` | Medium-light |
| `chart-3` | `oklch(0.646 0.222 41.116)` | Primary series (= primary color) |
| `chart-4` | `oklch(0.553 0.195 38.402)` | Medium-dark |
| `chart-5` | `oklch(0.47 0.157 37.304)` | Darkest — text/labels on charts |

---

## Typography

### Font Stack

```css
/* System font stack — no web font loading overhead */
--font-sans: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont,
             "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
--font-mono: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas,
             "Liberation Mono", monospace;
```

### Type Scale

| Level | Class | Size | Weight | Usage |
|-------|-------|------|--------|-------|
| H1 | `text-3xl font-bold` | 30px | 700 | Page titles |
| H2 | `text-2xl font-semibold` | 24px | 600 | Section headers |
| H3 | `text-xl font-semibold` | 20px | 600 | Card titles |
| H4 | `text-lg font-medium` | 18px | 500 | Sub-headers |
| Body | `text-sm` | 14px | 400 | Default text (admin UI) |
| Small | `text-xs` | 12px | 400 | Labels, timestamps |
| Mono | `font-mono text-sm` | 14px | 400 | Code, API keys |

---

## Spacing & Layout

### Border Radius

```
--radius: 0.625rem (10px)

Derived tokens:
  --radius-sm:  6px   (buttons, inputs)
  --radius-md:  8px   (cards, dropdowns)
  --radius-lg:  10px  (dialogs, large cards)
  --radius-xl:  14px  (modals)
```

### Layout Breakpoints

| Breakpoint | Width | Layout |
|------------|-------|--------|
| `sm` | 640px | Mobile |
| `md` | 768px | Tablet (sidebar collapses) |
| `lg` | 1024px | Desktop (sidebar visible) |
| `xl` | 1280px | Wide desktop |
| `2xl` | 1536px | Ultra-wide |

### Dashboard Shell

```
┌──────────────────────────────────────────┐
│ Sidebar (240px)  │  Main Content         │
│                  │  ┌─────────────────┐  │
│  Logo            │  │ Header (h-14)   │  │
│  Nav Items       │  ├─────────────────┤  │
│  ...             │  │ Page Content    │  │
│                  │  │ (p-6)           │  │
│  User Menu       │  │                 │  │
│                  │  └─────────────────┘  │
└──────────────────────────────────────────┘
```

- Sidebar width: `w-60` (240px), collapsible to `w-14` (56px, icons only)
- Header height: `h-14` (56px)
- Content padding: `p-6` (24px)
- Max content width: `max-w-7xl` (1280px) for form pages

---

## Component Patterns

### Buttons

```tsx
// Primary CTA — amber background
<Button>Create Post</Button>

// Secondary — gray outline
<Button variant="secondary">Cancel</Button>

// Destructive — red
<Button variant="destructive">Delete</Button>

// Ghost — transparent hover
<Button variant="ghost">More Options</Button>

// Icon button
<Button variant="ghost" size="icon">
  <Plus className="h-4 w-4" />
</Button>
```

### Cards (Dashboard Stats)

```tsx
<Card>
  <CardHeader className="flex flex-row items-center justify-between pb-2">
    <CardTitle className="text-sm font-medium text-muted-foreground">
      Total Posts
    </CardTitle>
    <FileText className="h-4 w-4 text-muted-foreground" />
  </CardHeader>
  <CardContent>
    <div className="text-2xl font-bold">1,234</div>
    <p className="text-xs text-muted-foreground">+12% from last month</p>
  </CardContent>
</Card>
```

### Data Tables

```tsx
<Table>
  <TableHeader>
    <TableRow>
      <TableHead>Title</TableHead>
      <TableHead>Status</TableHead>
      <TableHead>Date</TableHead>
      <TableHead className="text-right">Actions</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    <TableRow>
      <TableCell className="font-medium">My Post</TableCell>
      <TableCell><Badge>Published</Badge></TableCell>
      <TableCell className="text-muted-foreground">2026-02-25</TableCell>
      <TableCell className="text-right">
        <Button variant="ghost" size="icon">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </TableCell>
    </TableRow>
  </TableBody>
</Table>
```

### Status Badges

```tsx
// Post status
<Badge>Published</Badge>                                  // primary (amber)
<Badge variant="secondary">Draft</Badge>                  // gray
<Badge variant="outline">Scheduled</Badge>                // outline
<Badge className="bg-destructive text-white">Trashed</Badge> // red
```

### Alerts & Toasts

```tsx
// Success toast
toast.success("Post published successfully")

// Error toast
toast.error("Failed to save changes")

// Warning alert
<Alert className="bg-warning/10 border-warning text-warning-foreground">
  <AlertTriangle className="h-4 w-4" />
  <AlertDescription>This action cannot be undone.</AlertDescription>
</Alert>
```

---

## Sidebar Navigation

### Structure

```tsx
// Nav item — active state uses sidebar-primary
<SidebarMenuItem>
  <SidebarMenuButton isActive={true}>
    <LayoutDashboard className="h-4 w-4" />
    <span>Dashboard</span>
  </SidebarMenuButton>
</SidebarMenuItem>
```

### Nav Groups

| Group | Items |
|-------|-------|
| Main | Dashboard |
| Content | Posts, Categories, Tags, Media, Comments |
| Structure | Menus, Redirects |
| Admin | Users, Roles, API Keys, Settings, Audit Log |

---

## Icons

Using **Lucide React** (`lucide-react`) for consistency with shadcn/ui.

### Common Icons

| Context | Icon | Import |
|---------|------|--------|
| Dashboard | `LayoutDashboard` | `lucide-react` |
| Posts | `FileText` | `lucide-react` |
| Categories | `FolderTree` | `lucide-react` |
| Tags | `Tag` | `lucide-react` |
| Media | `Image` | `lucide-react` |
| Comments | `MessageSquare` | `lucide-react` |
| Menus | `Menu` | `lucide-react` |
| Redirects | `ArrowRightLeft` | `lucide-react` |
| Users | `Users` | `lucide-react` |
| Roles | `Shield` | `lucide-react` |
| API Keys | `Key` | `lucide-react` |
| Settings | `Settings` | `lucide-react` |
| Audit Log | `ScrollText` | `lucide-react` |
| Add/Create | `Plus` | `lucide-react` |
| Edit | `Pencil` | `lucide-react` |
| Delete | `Trash2` | `lucide-react` |
| Search | `Search` | `lucide-react` |
| Filter | `Filter` | `lucide-react` |
| Sort | `ArrowUpDown` | `lucide-react` |
| Publish | `Send` | `lucide-react` |
| Draft | `FileEdit` | `lucide-react` |
| Schedule | `Clock` | `lucide-react` |
| Light mode | `Sun` | `lucide-react` |
| Dark mode | `Moon` | `lucide-react` |
| Language | `Languages` | `lucide-react` |
| Logout | `LogOut` | `lucide-react` |
| External link | `ExternalLink` | `lucide-react` |
| Copy | `Copy` | `lucide-react` |
| Check | `Check` | `lucide-react` |
| Warning | `AlertTriangle` | `lucide-react` |
| Info | `Info` | `lucide-react` |
| Close | `X` | `lucide-react` |
| Chevron | `ChevronRight` | `lucide-react` |
| More | `MoreHorizontal` | `lucide-react` |

### Icon Sizing

| Context | Class | Size |
|---------|-------|------|
| Inline with text | `h-4 w-4` | 16px |
| Button icon | `h-4 w-4` | 16px |
| Sidebar nav | `h-4 w-4` | 16px |
| Empty state | `h-12 w-12` | 48px |
| Page header | `h-5 w-5` | 20px |

---

## Dark Mode Implementation

### Toggle Mechanism

Theme stored in `localStorage` under key `theme`. Values: `"light"` | `"dark"` | `"system"`.

```tsx
// ThemeToggle component — already implemented in Batch 9
// Adds/removes .dark class on <html> element
// Respects system preference via prefers-color-scheme
```

### Design Considerations

- **Light mode**: White backgrounds, dark text, amber primary accents
- **Dark mode**: Near-black backgrounds (#1a1a1e), light text, slightly brighter amber accents
- **Borders**: Light uses solid gray, dark uses `oklch(1 0 0 / 10%)` (white at 10% opacity)
- **Cards**: Light = white, Dark = slightly lighter than background (#2a2a2e)
- **Shadows**: Light = subtle gray shadows, Dark = no visible shadows (use borders instead)

---

## Design Principles (from Stagehand.dev)

### Borrowed Patterns

1. **Warm neutral palette** — Amber/orange primary instead of cold blue, creates friendly CMS feel
2. **Clean stat cards** — Large numbers + small labels for dashboard KPIs
3. **Minimal data tables** — Thin borders, no zebra stripes, clean typography
4. **Full-screen mobile nav** — Hamburger menu expands to full-screen overlay on mobile

### CMS Adaptations

1. **Solid borders over dashed** — Admin panels need stability, not blueprint aesthetics
2. **Compact information density** — CMS users need to scan lots of data
3. **Consistent icon system** — Lucide icons at 16px for all nav and actions
4. **Status color coding** — Amber (active), Gray (draft), Green (published), Red (trashed)

---

## File Organization

```
web/src/
├── styles/
│   └── global.css          # Theme variables + Tailwind imports
├── components/
│   ├── ui/                 # shadcn/ui primitives (auto-generated)
│   ├── layout/             # Shell, Sidebar, Header, ThemeToggle
│   └── providers/          # QueryProvider, ThemeProvider, I18nProvider
├── lib/
│   └── utils.ts            # cn() helper for class merging
└── hooks/                  # Custom React hooks
```
