<script lang="ts">
/* eslint-disable @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-argument -- superforms zod4 adapter types not resolved by eslint */
import { superForm } from 'sveltekit-superforms';
import { zod4Client } from 'sveltekit-superforms/adapters';
import { loginSchema } from '$lib/schemas/auth';
import { resolve } from '$app/paths';
import { cx } from 'styled-system/css';
import { button } from 'styled-system/recipes';
import { Mail, Lock, Eye, EyeOff } from '@lucide/svelte';
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
	validators: zod4Client(loginSchema),
});

let showPassword = $state(false);
</script>

<svelte:head>
  <title>Login &mdash; MotoPhoto</title>
</svelte:head>

<div class={authContainer}>
  <div class={authCard}>
    <h1 class={authTitle}>Login</h1>

    {#if $message}
      <FormAlert message={$message} />
    {/if}

    <form method="POST" use:enhance>
      <div class={authFormFields}>
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
              placeholder="Enter your password"
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

        <button type="submit" class={button({ variant: 'solid', size: 'lg' })} disabled={$delayed}>
          {#if $delayed}Logging in...{:else}Login{/if}
        </button>
      </div>
    </form>

    <p class={authFooterText}>
      Don't have an account? <a href={resolve("/register")} class={authFooterLink}>Register</a>
    </p>
  </div>
</div>
