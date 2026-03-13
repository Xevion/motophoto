# PandaCSS

This project uses [PandaCSS](https://panda-css.com) for styling. It generates CSS at build time via static extraction -- no runtime overhead, no style injection. Understanding how it works prevents silent styling failures.

## How It Works

PandaCSS scans source files at build time using a static AST parser. It reads literal values from `css()`, `cva()`, and recipe calls, then generates atomic CSS classes into `styled-system/`. The PostCSS plugin (`@pandacss/dev/postcss`) injects the output into the `@layer` declarations in `src/app.css`. No CSS files from `styled-system/` need to be manually imported.

**Nothing is executed at scan time -- only source text is read.** Values that are only known at runtime (component props, state variables, computed expressions) are invisible to the extractor.

## Token System

Design tokens are defined in `panda.config.ts` under `theme.tokens` (base values) and `theme.semanticTokens` (context-aware aliases like dark mode variants).

**Keys that don't exist in the token system are passed through as raw CSS values with no units and no warning.** The result is a broken style with no error.

```ts
css({ px: 'gutter' })  // 'gutter' not defined -- outputs: padding-inline: gutter (broken)
css({ px: '4' })       // outputs: padding-inline: var(--spacing-4) (correct)
```

This project uses the default `@pandacss/preset-panda` (Tailwind-like scale) plus custom additions in `panda.config.ts`. The full available token set is in the generated `styled-system/tokens/tokens.d.ts` -- check it when unsure. Key categories:

| Category | Source | Example keys |
|---|---|---|
| `spacing` | preset + config | Full Tailwind scale: `0 0.5 1 1.5 2 2.5 3 3.5 4 5 6 7 8 9 10 11 12 14 16 20 24 ...` |
| `radii` | config override | `sm md lg xl 2xl full` |
| `fontSizes` | config override | `xs sm md lg xl 2xl 3xl 4xl` |
| `fontWeights` | config override | `normal medium semibold bold` |
| `shadows` | config override | `sm md lg` |
| `colors` | config | `orange.*` and `neutral.*` scales |
| Semantic colors | config | `bg bg.muted bg.subtle bg.emphasis fg fg.muted fg.subtle border border.strong primary primary.hover primary.fg primary.subtle danger.* success.*` |

## Static Extraction and Dynamic Recipe Calls

**The most common source of missing styles:** calling a recipe with props that are runtime variables.

```svelte
<!-- button.svelte -->
<script lang="ts">
  let { variant, size } = $props()
  const classes = button({ variant, size })  // PandaCSS cannot see the values
</script>
```

PandaCSS sees `button({ variant, size })` but cannot determine what `variant` or `size` will be at runtime. Only variant combinations it has seen with literal values elsewhere get CSS generated. Any combination not seen in source is silently missing.

**Fix: `staticCss` in `panda.config.ts`.** Any recipe used through a wrapper component with dynamic props must be listed:

```ts
staticCss: {
  recipes: {
    button: ['*'],  // generate all variant combinations unconditionally
    badge: ['*'],
  }
}
```

The `['*']` wildcard generates every variant combination. When adding a new recipe that will be called with dynamic props, add it to `staticCss` at the same time.

Recipes defined with `cva()` (inline) always generate all variants and do not need `staticCss`. Config recipes (`defineRecipe`) use just-in-time generation by default -- `staticCss` overrides this.

## Recipe Patterns

Recipes live in `src/lib/recipes/`. Each file exports a `defineRecipe` call registered in `panda.config.ts`.

```ts
// src/lib/recipes/button.ts
import { defineRecipe } from '@pandacss/dev'

export const buttonRecipe = defineRecipe({
  className: 'button',
  base: {
    display: 'inline-flex',
    borderRadius: 'lg',
    fontWeight: 'semibold',
  },
  variants: {
    variant: {
      solid:   { bg: 'primary', color: 'primary.fg' },
      ghost:   { bg: 'transparent', color: 'fg', _hover: { bg: 'bg.muted' } },
      outline: { borderWidth: '1px', borderColor: 'border', _hover: { bg: 'bg.muted' } },
    },
    size: {
      sm: { h: '8', px: '3', fontSize: 'xs' },
      md: { h: '10', px: '5', fontSize: 'sm' },
    },
  },
  defaultVariants: { variant: 'solid', size: 'md' },
})
```

Then register it in `panda.config.ts`:

```ts
theme: {
  extend: {
    recipes: {
      button: buttonRecipe,
    }
  }
}
```

**Slot recipes** (`defineSlotRecipe`) are for components with multiple coordinated parts (e.g. a menu with `root`, `item`, `content` parts). See `src/lib/recipes/menu.ts` for an example.

## The `token()` CSS Function

Use `token()` only inside composite CSS string values where a bare token key is not valid syntax:

```ts
// composite shadow value -- token() resolves to the raw CSS variable value
css({ boxShadow: '0 2px 8px token(colors.primary)' })

// shorthand {} syntax is equivalent and slightly shorter
css({ boxShadow: '0 2px 8px {colors.primary}' })
```

**Do not use `token()` with `/opacity` syntax.** The `/opacity` modifier (`red.500/40`) is only valid in style object property values, not inside composite strings:

```ts
// good -- /opacity modifier in property value position
css({ bg: 'primary/30' })
css({ borderColor: 'danger/50' })

// broken -- /opacity is not valid inside token() composite string
css({ boxShadow: '0 1px 4px token(colors.primary/0.3)' })  // does nothing
```

## `css()` Usage Patterns

All `css()` calls must be **named constants in `<script>`**, not inline in templates. PandaCSS extracts from `<script>` blocks; inline calls in templates may not be extracted and create noise.

```svelte
<!-- good -->
<script lang="ts">
  const wrapper = css({ display: 'flex', gap: '4' })
</script>
<div class={wrapper}>...</div>

<!-- avoid -->
<div class={css({ display: 'flex', gap: '4' })}>...</div>
```

Use `cx()` for conditional class merging:

```svelte
<script lang="ts">
  import { css, cx } from 'styled-system/css'
  const base = css({ borderColor: 'border' })
  const errored = css({ borderColor: 'danger' })
</script>

<input class={cx(base, hasError && errored)} />
```

## Common Mistakes

| Mistake | Symptom | Fix |
|---|---|---|
| Token key not in `panda.config.ts` | Style silently ignored | Check config before using any key |
| Recipe called with dynamic props, no `staticCss` | Component renders unstyled | Add recipe to `staticCss.recipes` |
| Spacing value like `'3.5'` or `'7'` not in config | Unitless raw value in CSS | Use a defined key or explicit units like `'14px'` |
| `token(colors.x)/50` in composite string | Opacity modifier ignored | Use `borderColor: 'x/50'` property instead |
| Hand-editing `styled-system/` | Overwritten on next codegen | Never edit; change source recipe or config |
| Adding a recipe to config without `panda codegen` | Old output used | Run `just check` which triggers codegen |
