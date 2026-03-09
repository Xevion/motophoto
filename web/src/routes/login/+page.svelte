<script lang="ts">
/* eslint-disable @typescript-eslint/no-unsafe-member-access -- superforms zod4 adapter types not resolved by eslint */
import { superForm } from 'sveltekit-superforms';
import { zod4Client } from 'sveltekit-superforms/adapters';
import { loginSchema } from '$lib/schemas/auth';
import { resolve } from '$app/paths';
import { css, cx } from 'styled-system/css';
import { button } from 'styled-system/recipes';
import { Mail, Lock } from '@lucide/svelte';

let { data } = $props();

// svelte-ignore state_referenced_locally
const { form, errors, constraints, enhance, delayed, message } = superForm(data.form, {
	validators: zod4Client(loginSchema),
});

const container = css({
	display: 'flex',
	justifyContent: 'center',
	py: '12',
});

const card = css({
	w: 'full',
	maxW: '400px',
	bg: 'bg.subtle',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'xl',
	p: '8',
	display: 'flex',
	flexDirection: 'column',
	gap: '6',
});

const title = css({
	fontSize: '2xl',
	fontWeight: 'bold',
	textAlign: 'center',
});

const fieldGroup = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '1.5',
});

const label = css({
	fontSize: 'sm',
	fontWeight: 'medium',
});

const inputWrapper = css({
	display: 'flex',
	alignItems: 'center',
	gap: '2',
	bg: 'bg',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'md',
	px: '3',
	h: '10',
	transition: 'border-color',
	transitionDuration: '150ms',
	_focusWithin: {
		borderColor: 'primary',
		outlineWidth: '1px',
		outlineColor: 'primary',
		outlineStyle: 'solid',
	},
});

const inputWrapperError = css({
	borderColor: 'danger',
	_focusWithin: {
		borderColor: 'danger',
		outlineColor: 'danger',
	},
});

const input = css({
	flex: '1',
	bg: 'transparent',
	border: 'none',
	outline: 'none',
	color: 'fg',
	fontSize: 'sm',
	h: 'full',
	'&::placeholder': {
		color: 'fg.subtle',
	},
});

const iconStyle = css({
	color: 'fg.muted',
	flexShrink: 0,
});

const errorText = css({
	fontSize: 'xs',
	color: 'danger',
});

const formMessage = css({
	fontSize: 'sm',
	textAlign: 'center',
	p: '3',
	borderRadius: 'md',
	bg: 'danger',
	color: 'danger.fg',
});

const footerText = css({
	fontSize: 'sm',
	color: 'fg.muted',
	textAlign: 'center',
});

const footerLink = css({
	color: 'primary',
	fontWeight: 'medium',
	_hover: { textDecoration: 'underline' },
});

const formFields = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '4',
});
</script>

<svelte:head>
  <title>Login &mdash; MotoPhoto</title>
</svelte:head>

<div class={container}>
  <div class={card}>
    <h1 class={title}>Login</h1>

    {#if $message}
      <div class={formMessage}>{$message}</div>
    {/if}

    <form method="POST" use:enhance>
      <div class={formFields}>
        <div class={fieldGroup}>
          <label for="email" class={label}>Email</label>
          <div class={cx(inputWrapper, $errors.email && inputWrapperError)}>
            <Mail size={16} class={iconStyle} />
            <input
              id="email"
              name="email"
              type="email"
              placeholder="you@example.com"
              class={input}
              aria-invalid={$errors.email ? 'true' : undefined}
              bind:value={$form.email}
              {...{ ...$constraints.email, pattern: undefined }}
            />
          </div>
          {#if $errors.email}<span class={errorText}>{$errors.email}</span>{/if}
        </div>

        <div class={fieldGroup}>
          <label for="password" class={label}>Password</label>
          <div class={cx(inputWrapper, $errors.password && inputWrapperError)}>
            <Lock size={16} class={iconStyle} />
            <input
              id="password"
              name="password"
              type="password"
              placeholder="Enter your password"
              class={input}
              aria-invalid={$errors.password ? 'true' : undefined}
              bind:value={$form.password}
              {...$constraints.password}
            />
          </div>
          {#if $errors.password}<span class={errorText}>{$errors.password}</span>{/if}
        </div>

        <button type="submit" class={button({ variant: 'solid', size: 'lg' })} disabled={$delayed}>
          {#if $delayed}Logging in...{:else}Login{/if}
        </button>
      </div>
    </form>

    <p class={footerText}>
      Don't have an account? <a href={resolve("/register")} class={footerLink}>Register</a>
    </p>
  </div>
</div>
