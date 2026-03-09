<script lang="ts">
/* eslint-disable @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-unsafe-assignment -- superforms zod4 adapter types not resolved by eslint */
import { superForm } from 'sveltekit-superforms';
import { zod4Client } from 'sveltekit-superforms/adapters';
import { registerSchema } from '$lib/schemas/auth';
import { resolve } from '$app/paths';
import { css, cx } from 'styled-system/css';
import { button } from 'styled-system/recipes';
import { Mail, Lock, User } from '@lucide/svelte';
import UiSelect from '$lib/components/ui/select.svelte';

let { data } = $props();

// svelte-ignore state_referenced_locally
const { form, errors, constraints, enhance, delayed, message } = superForm(data.form, {
	validators: zod4Client(registerSchema),
});

const roleItems = [
	{ label: 'Photographer', value: 'photographer' },
	{ label: 'Customer', value: 'customer' },
];

let roleValue = $state<string[]>($form.role ? [$form.role] : []);

$effect(() => {
	$form.role = (roleValue[0] ?? '') as typeof $form.role;
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

const titleStyle = css({
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

const selectTrigger = css({
	display: 'flex',
	alignItems: 'center',
	gap: '2',
	bg: 'bg',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'md',
	px: '3',
	h: '10',
	w: 'full',
	cursor: 'pointer',
	transition: 'border-color',
	transitionDuration: '150ms',
	_focusVisible: {
		outline: 'none',
		borderColor: 'primary',
		outlineWidth: '1px',
		outlineColor: 'primary',
		outlineStyle: 'solid',
	},
	_open: {
		borderColor: 'primary',
		outlineWidth: '1px',
		outlineColor: 'primary',
		outlineStyle: 'solid',
	},
});

const selectTriggerError = css({
	borderColor: 'danger',
	_focusVisible: {
		borderColor: 'danger',
		outlineColor: 'danger',
	},
	_open: {
		borderColor: 'danger',
		outlineColor: 'danger',
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
  <title>Register &mdash; MotoPhoto</title>
</svelte:head>

<div class={container}>
  <div class={card}>
    <h1 class={titleStyle}>Create an account</h1>

    {#if $message}
      <div class={formMessage}>{$message}</div>
    {/if}

    <form method="POST" use:enhance>
      <div class={formFields}>
        <div class={fieldGroup}>
          <label for="display_name" class={label}>Display name</label>
          <div class={cx(inputWrapper, $errors.display_name && inputWrapperError)}>
            <User size={16} class={iconStyle} />
            <input
              id="display_name"
              name="display_name"
              type="text"
              placeholder="Your name"
              class={input}
              aria-invalid={$errors.display_name ? 'true' : undefined}
              bind:value={$form.display_name}
              {...$constraints.display_name}
            />
          </div>
          {#if $errors.display_name}<span class={errorText}>{$errors.display_name}</span>{/if}
        </div>

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
              placeholder="At least 8 characters"
              class={input}
              aria-invalid={$errors.password ? 'true' : undefined}
              bind:value={$form.password}
              {...$constraints.password}
            />
          </div>
          {#if $errors.password}<span class={errorText}>{$errors.password}</span>{/if}
        </div>

        <div class={fieldGroup}>
          <label for="confirm_password" class={label}>Confirm password</label>
          <div class={cx(inputWrapper, $errors.confirm_password && inputWrapperError)}>
            <Lock size={16} class={iconStyle} />
            <input
              id="confirm_password"
              name="confirm_password"
              type="password"
              placeholder="Repeat your password"
              class={input}
              aria-invalid={$errors.confirm_password ? 'true' : undefined}
              bind:value={$form.confirm_password}
              {...$constraints.confirm_password}
            />
          </div>
          {#if $errors.confirm_password}<span class={errorText}>{$errors.confirm_password}</span>{/if}
        </div>

        <div class={fieldGroup}>
          <label for="role" class={label}>I am a...</label>
          <UiSelect
            items={roleItems}
            value={roleValue}
            onValueChange={(v: string[]) => (roleValue = v)}
            placeholder="Select your role"
            name="role"
            id="role"
            invalid={!!$errors.role}
            triggerClass={cx(selectTrigger, $errors.role && selectTriggerError)}
          />
          {#if $errors.role}<span class={errorText}>{$errors.role}</span>{/if}
        </div>

        <button type="submit" class={button({ variant: 'solid', size: 'lg' })} disabled={$delayed}>
          {#if $delayed}Creating account...{:else}Create account{/if}
        </button>
      </div>
    </form>

    <p class={footerText}>
      Already have an account? <a href={resolve("/login")} class={footerLink}>Login</a>
    </p>
  </div>
</div>
