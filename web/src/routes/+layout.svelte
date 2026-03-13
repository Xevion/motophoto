<script lang="ts">
import '../app.css';
import favicon from '$lib/assets/favicon.svg';
import { resolve } from '$app/paths';
import { themeStore } from '$lib/stores/theme.svelte';
import ThemeToggle from '$lib/components/theme-toggle.svelte';
import Button from '$lib/components/ui/button.svelte';
import type { Snippet } from 'svelte';
import { css } from 'styled-system/css';

let { children }: { children: Snippet } = $props();

themeStore.init();

const navLinks = css({
	display: 'flex',
	alignItems: 'center',
	gap: '1',
});

const navLink = css({
	fontSize: 'sm',
	fontWeight: 'medium',
	color: 'fg.muted',
	textDecoration: 'none',
	px: '3',
	py: '1.5',
	borderRadius: 'md',
	transition: 'colors',
	transitionDuration: '150ms',
	_hover: { color: 'fg', bg: 'bg.muted' },
});

const layout = css({
	minH: '100vh',
	bg: 'bg',
	color: 'fg',
});

const header = css({
	position: 'sticky',
	top: '0',
	zIndex: '50',
	w: 'full',
	borderBottomWidth: '1px',
	borderColor: 'border',
	bg: 'bg/95',
	backdropFilter: 'blur(8px)',
});

const headerInner = css({
	mx: 'auto',
	display: 'flex',
	h: '14',
	maxW: '5xl',
	alignItems: 'center',
	justifyContent: 'space-between',
	px: '4',
});

const logoLink = css({
	display: 'flex',
	alignItems: 'center',
	gap: '2',
	fontWeight: 'bold',
	fontSize: 'lg',
	textDecoration: 'none',
});

const logoText = css({
	background: 'linear-gradient(to right, {colors.orange.500}, {colors.orange.700})',
	backgroundClip: 'text',
	color: 'transparent',
});

const main = css({
	mx: 'auto',
	maxW: '5xl',
	px: '4',
	py: '8',
});

const navActions = css({
	display: 'flex',
	alignItems: 'center',
	gap: '2',
});
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

<div class={layout}>
  <header class={header}>
    <div class={headerInner}>
      <a href={resolve("/")} class={logoLink}>
        <span class={logoText}>MotoPhoto</span>
      </a>
      <nav class={navLinks}>
        <a href={resolve('/events')} class={navLink}>Browse Events</a>
        <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- route not yet created -->
        <a href="/for-photographers" class={navLink}>For Photographers</a>
      </nav>
      <div class={navActions}>
        <Button href={resolve('/login')} variant="ghost" size="sm">Log In</Button>
        <Button href={resolve('/register')} size="sm">Sign Up</Button>
        <ThemeToggle />
      </div>
    </div>
  </header>

  <main class={main}>
    {@render children()}
  </main>
</div>
