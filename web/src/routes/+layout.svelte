<script lang="ts">
/* eslint-disable @typescript-eslint/no-unsafe-call -- ArkUI asChild callback types unresolved */
import '../app.css';
import favicon from '$lib/assets/favicon.svg';
import { resolve } from '$app/paths';
import { themeStore } from '$lib/stores/theme.svelte';
import ThemeToggle from '$lib/components/theme-toggle.svelte';
import type { Snippet } from 'svelte';
import { css } from 'styled-system/css';
import { Menu } from '@ark-ui/svelte/menu';
import { Avatar } from '@ark-ui/svelte/avatar';
import { menu } from 'styled-system/recipes';

import LogIn from '@lucide/svelte/icons/log-in';
import UserPlus from '@lucide/svelte/icons/user-plus';

let { children }: { children: Snippet } = $props();

themeStore.init();

const menuClasses = menu();

const avatarRoot = css({
	display: 'inline-flex',
	alignItems: 'center',
	justifyContent: 'center',
	fontWeight: 'medium',
	flexShrink: 0,
	borderRadius: 'full',
	bg: 'bg.muted',
	color: 'fg.muted',
	fontSize: 'xs',
	w: '8',
	h: '8',
	cursor: 'pointer',
});

const avatarTrigger = css({
	display: 'inline-flex',
	alignItems: 'center',
	gap: '1.5',
	bg: 'transparent',
	border: 'none',
	cursor: 'pointer',
	borderRadius: 'full',
	p: '0',
	_focusVisible: {
		outlineWidth: '2px',
		outlineColor: 'primary',
		outlineOffset: '2px',
		outlineStyle: 'solid',
		borderRadius: 'full',
	},
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
      <div class={navActions}>
        <Menu.Root positioning={{ placement: 'bottom-end' }} closeOnSelect>
          <Menu.Trigger class={avatarTrigger}>
            <Avatar.Root class={avatarRoot}>
              <Avatar.Fallback>?</Avatar.Fallback>
            </Avatar.Root>
          </Menu.Trigger>
          <Menu.Positioner class={menuClasses.positioner}>
            <Menu.Content class={menuClasses.content}>
              <Menu.Item class={menuClasses.item} value="login">
                {#snippet asChild(itemProps)}
                  <a href={resolve('/login')} {...itemProps()}>
                    <LogIn />
                    Login
                  </a>
                {/snippet}
              </Menu.Item>
              <Menu.Item class={menuClasses.item} value="register">
                {#snippet asChild(itemProps)}
                  <a href={resolve('/register')} {...itemProps()}>
                    <UserPlus />
                    Register
                  </a>
                {/snippet}
              </Menu.Item>
            </Menu.Content>
          </Menu.Positioner>
        </Menu.Root>
        <ThemeToggle />
      </div>
    </div>
  </header>

  <main class={main}>
    {@render children()}
  </main>
</div>
