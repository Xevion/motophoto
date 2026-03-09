<script lang="ts">
/* eslint-disable @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-argument -- superforms zod4 adapter types not resolved by eslint */
import { superForm } from 'sveltekit-superforms';
import { zod4Client } from 'sveltekit-superforms/adapters';
import { registerSchema } from '$lib/schemas/auth';
import { resolve } from '$app/paths';
import { css, cx } from 'styled-system/css';
import { button } from 'styled-system/recipes';
import { Mail, Lock, User, Eye, EyeOff } from '@lucide/svelte';
import UiSelect from '$lib/components/ui/select.svelte';
import FormAlert from '$lib/components/ui/form-alert.svelte';
import {
	authContainer,
	authCard,
	authTitle,
	authFieldGroup,
	authLabel,
	authInputWrapper,
	authInputWrapperError,
	authInput,
	authIcon,
	authErrorText,
	authFooterText,
	authFooterLink,
	authFormFields,
	authVisibilityToggle,
} from '$lib/styles/auth-form';

let { data } = $props();

// svelte-ignore state_referenced_locally
const { form, errors, constraints, enhance, delayed, message } = superForm(data.form, {
	validators: zod4Client(registerSchema),
});

const roleItems = [
	{ label: 'Photographer', value: 'photographer' },
	{ label: 'Customer', value: 'customer' },
];

function handleRoleChange(values: string[]) {
	const selected = values[0];
	if (selected === 'photographer' || selected === 'customer') {
		$form.role = selected;
	}
}

let showPassword = $state(false);
let showConfirmPassword = $state(false);

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
</script>

<svelte:head>
  <title>Register &mdash; MotoPhoto</title>
</svelte:head>

<div class={authContainer}>
  <div class={authCard}>
    <h1 class={authTitle}>Create an account</h1>

    {#if $message}
      <FormAlert message={$message} />
    {/if}

    <form method="POST" use:enhance>
      <div class={authFormFields}>
        <div class={authFieldGroup}>
          <label for="display_name" class={authLabel}>Display name</label>
          <div class={cx(authInputWrapper, $errors.display_name && authInputWrapperError)}>
            <User size={16} class={authIcon} />
            <input
              id="display_name"
              name="display_name"
              type="text"
              placeholder="Your name"
              class={authInput}
              aria-invalid={$errors.display_name ? 'true' : undefined}
              aria-describedby={$errors.display_name ? 'display_name-error' : undefined}
              bind:value={$form.display_name}
              {...$constraints.display_name}
            />
          </div>
          {#if $errors.display_name}<span id="display_name-error" class={authErrorText}>{$errors.display_name}</span>{/if}
        </div>

        <div class={authFieldGroup}>
          <label for="email" class={authLabel}>Email</label>
          <div class={cx(authInputWrapper, $errors.email && authInputWrapperError)}>
            <Mail size={16} class={authIcon} />
            <input
              id="email"
              name="email"
              type="email"
              placeholder="you@example.com"
              class={authInput}
              aria-invalid={$errors.email ? 'true' : undefined}
              aria-describedby={$errors.email ? 'email-error' : undefined}
              bind:value={$form.email}
              {...{ ...$constraints.email, pattern: undefined }}
            />
          </div>
          {#if $errors.email}<span id="email-error" class={authErrorText}>{$errors.email}</span>{/if}
        </div>

        <div class={authFieldGroup}>
          <label for="password" class={authLabel}>Password</label>
          <div class={cx(authInputWrapper, $errors.password && authInputWrapperError)}>
            <Lock size={16} class={authIcon} />
            <input
              id="password"
              name="password"
              type={showPassword ? 'text' : 'password'}
              placeholder="At least 8 characters"
              class={authInput}
              aria-invalid={$errors.password ? 'true' : undefined}
              aria-describedby={$errors.password ? 'password-error' : undefined}
              bind:value={$form.password}
              {...$constraints.password}
            />
            <button
              type="button"
              class={authVisibilityToggle}
              onclick={() => (showPassword = !showPassword)}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
            >
              {#if showPassword}<EyeOff size={16} />{:else}<Eye size={16} />{/if}
            </button>
          </div>
          {#if $errors.password}<span id="password-error" class={authErrorText}>{$errors.password}</span>{/if}
        </div>

        <div class={authFieldGroup}>
          <label for="confirm_password" class={authLabel}>Confirm password</label>
          <div class={cx(authInputWrapper, $errors.confirm_password && authInputWrapperError)}>
            <Lock size={16} class={authIcon} />
            <input
              id="confirm_password"
              name="confirm_password"
              type={showConfirmPassword ? 'text' : 'password'}
              placeholder="Repeat your password"
              class={authInput}
              aria-invalid={$errors.confirm_password ? 'true' : undefined}
              aria-describedby={$errors.confirm_password ? 'confirm_password-error' : undefined}
              bind:value={$form.confirm_password}
              {...$constraints.confirm_password}
            />
            <button
              type="button"
              class={authVisibilityToggle}
              onclick={() => (showConfirmPassword = !showConfirmPassword)}
              aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
            >
              {#if showConfirmPassword}<EyeOff size={16} />{:else}<Eye size={16} />{/if}
            </button>
          </div>
          {#if $errors.confirm_password}<span id="confirm_password-error" class={authErrorText}>{$errors.confirm_password}</span>{/if}
        </div>

        <div class={authFieldGroup}>
          <label for="role" class={authLabel}>I am a...</label>
          <UiSelect
            items={roleItems}
            value={$form.role ? [$form.role] : []}
            onValueChange={handleRoleChange}
            placeholder="Select your role"
            name="role"
            id="role"
            invalid={!!$errors.role}
            triggerClass={cx(selectTrigger, $errors.role && selectTriggerError)}
          />
          {#if $errors.role}<span id="role-error" class={authErrorText}>{$errors.role}</span>{/if}
        </div>

        <button type="submit" class={button({ variant: 'solid', size: 'lg' })} disabled={$delayed}>
          {#if $delayed}Creating account...{:else}Create account{/if}
        </button>
      </div>
    </form>

    <p class={authFooterText}>
      Already have an account? <a href={resolve("/login")} class={authFooterLink}>Login</a>
    </p>
  </div>
</div>
