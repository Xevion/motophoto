<script lang="ts">
import { superForm } from 'sveltekit-superforms';
import { zod4Client } from 'sveltekit-superforms/adapters';
import { gallerySchema } from '$lib/schemas/event';
import { css } from 'styled-system/css';
import Button from '$lib/components/ui/button.svelte';
import FormAlert from '$lib/components/ui/form-alert.svelte';
import ArrowLeft from '@lucide/svelte/icons/arrow-left';
import type { PageData } from './$types';

const { data }: { data: PageData } = $props();

const { form, errors, constraints, enhance, delayed, message } = superForm(data.form, {
	validators: zod4Client(gallerySchema),
});

const pageContainer = css({
	maxW: '2xl',
	display: 'flex',
	flexDirection: 'column',
	gap: '6',
});

const pageTitle = css({
	fontSize: '2xl',
	fontWeight: 'bold',
	display: 'flex',
	alignItems: 'center',
	gap: '2',
});

const formSection = css({
	bg: 'bg.subtle',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'lg',
	p: '6',
});

const fieldGroup = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '4',
});

const field = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '2',
});

const label = css({
	fontSize: 'sm',
	fontWeight: 'semibold',
	color: 'fg',
});

const input = css({
	w: 'full',
	px: '3',
	py: '2',
	bg: 'bg',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'md',
	fontSize: 'sm',
	transition: 'border-color 150ms',
	_focus: {
		outline: 'none',
		borderColor: 'primary',
	},
	_placeholder: {
		color: 'fg.muted',
	},
});

const textarea = css({
	w: 'full',
	px: '3',
	py: '2',
	bg: 'bg',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'md',
	fontSize: 'sm',
	minH: '100px',
	fontFamily: 'inherit',
	resize: 'vertical',
	transition: 'border-color 150ms',
	_focus: {
		outline: 'none',
		borderColor: 'primary',
	},
	_placeholder: {
		color: 'fg.muted',
	},
});

const errorText = css({
	fontSize: 'xs',
	color: 'danger',
});

const buttonGroup = css({
	display: 'flex',
	gap: '2',
	justifyContent: 'flex-start',
});
</script>

<div class={pageContainer}>
	<div class={pageTitle}>
		<a href="/dashboard">
			<Button variant="ghost" size="sm">
				<ArrowLeft />
			</Button>
		</a>
		<span>Create Gallery</span>
	</div>

	{#if $message}
		<FormAlert variant="error" message={$message} />
	{/if}

	<div class={formSection}>
		<form method="POST" use:enhance class={fieldGroup}>
			<div class={field}>
				<label class={label} for="name">Gallery Name *</label>
				<input
					class={input}
					type="text"
					id="name"
					name="name"
					placeholder="e.g., Heats"
					bind:value={$form.name}
					{...$constraints.name}
				/>
				{#if $errors.name}
					<span class={errorText}>{$errors.name}</span>
				{/if}
			</div>

			<div class={field}>
				<label class={label} for="description">Description</label>
				<textarea
					class={textarea}
					id="description"
					name="description"
					placeholder="Optional gallery details..."
					bind:value={$form.description}
					{...$constraints.description}
				/>
				{#if $errors.description}
					<span class={errorText}>{$errors.description}</span>
				{/if}
			</div>

			<div class={buttonGroup}>
				<Button variant="solid" type="submit" disabled={$delayed}>
					Create Gallery
				</Button>
				<a href="..">
					<Button variant="outline" type="button">Cancel</Button>
				</a>
			</div>
		</form>
	</div>
</div>
