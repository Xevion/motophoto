<script lang="ts">
import '../app.css';
import favicon from '$lib/assets/favicon.svg';
import { resolve } from '$app/paths';
import { themeStore } from '$lib/stores/theme.svelte';
import ThemeToggle from '$lib/components/theme-toggle.svelte';
import type { Snippet } from 'svelte';
import { css } from 'styled-system/css';

let { children }: { children: Snippet } = $props();

themeStore.init();

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
      <ThemeToggle />
    </div>
  </header>

  <main class={main}>
    {@render children()}
  </main>
</div>
