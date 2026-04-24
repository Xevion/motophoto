<script lang="ts">
import { superForm } from 'sveltekit-superforms';
import { zod4Client } from 'sveltekit-superforms/adapters';
import { eventSchema } from '$lib/schemas/event';
import { css } from 'styled-system/css';
import Button from '$lib/components/ui/button.svelte';
import UiSelect from '$lib/components/ui/select.svelte';
import FormAlert from '$lib/components/ui/form-alert.svelte';
import ArrowLeft from '@lucide/svelte/icons/arrow-left';
import type { PageData } from './$types';

interface Props {
	data: PageData;
	title: string;
	submitLabel?: string;
}

let { data, title, submitLabel = 'Create Event' }: Props = $props();

const { form, errors, constraints, enhance, delayed, message } = superForm(data.form, {
	validators: zod4Client(eventSchema),
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

const formRow = css({
	display: 'grid',
	gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
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
	minH: '120px',
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

const helpText = css({
	fontSize: 'xs',
	color: 'fg.muted',
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

const buttonGroup = css({
	display: 'flex',
	gap: '2',
	justifyContent: 'flex-start',
});

const statusOptions = [
	{ label: 'Draft', value: 'draft' },
	{ label: 'Published', value: 'published' },
	{ label: 'Archived', value: 'archived' },
];
</script>

<div class={pageContainer}>
	<div class={pageTitle}>
		<a href="/dashboard">
			<Button variant="ghost" size="sm">
				<ArrowLeft />
			</Button>
		</a>
		<span>{title}</span>
	</div>

	{#if $message}
		<FormAlert variant="error" message={$message} />
	{/if}

	<div class={formSection}>
		<form method="POST" use:enhance class={fieldGroup}>
			<div class={formRow}>
				<div class={field}>
					<label class={label} for="name">Event Name *</label>
					<input
						class={input}
						type="text"
						id="name"
						name="name"
						placeholder="e.g., Summer Motocross Championship"
						bind:value={$form.name}
						{...$constraints.name}
					/>
					{#if $errors.name}
						<span class={errorText}>{$errors.name}</span>
					{/if}
				</div>

				<div class={field}>
					<label class={label} for="slug">URL Slug *</label>
					<input
						class={input}
						type="text"
						id="slug"
						name="slug"
						placeholder="e.g., summer-motocross-championship"
						bind:value={$form.slug}
						{...$constraints.slug}
					/>
					{#if $errors.slug}
						<span class={errorText}>{$errors.slug}</span>
					{/if}
					<span class={helpText}>Lowercase letters, numbers, and hyphens only</span>
				</div>
			</div>

			<div class={formRow}>
				<div class={field}>
					<label class={label} for="sport">Sport *</label>
					<input
						class={input}
						type="text"
						id="sport"
						name="sport"
						placeholder="e.g., Motocross"
						bind:value={$form.sport}
						{...$constraints.sport}
					/>
					{#if $errors.sport}
						<span class={errorText}>{$errors.sport}</span>
					{/if}
				</div>

				<div class={field}>
					<label class={label} for="date">Date</label>
					<input
						class={input}
						type="date"
						id="date"
						name="date"
						bind:value={$form.date}
						{...$constraints.date}
					/>
					{#if $errors.date}
						<span class={errorText}>{$errors.date}</span>
					{/if}
				</div>
			</div>

			<div class={formRow}>
				<div class={field}>
					<label class={label} for="location">Location</label>
					<input
						class={input}
						type="text"
						id="location"
						name="location"
						placeholder="e.g., Daytona International Speedway"
						bind:value={$form.location}
						{...$constraints.location}
					/>
					{#if $errors.location}
						<span class={errorText}>{$errors.location}</span>
					{/if}
				</div>

				<div class={field}>
					<label class={label} for="status">Status *</label>
					<UiSelect
						items={statusOptions}
						value={[$form.status]}
						onValueChange={(v: string[]) => ($form.status = v[0] as 'draft' | 'published' | 'archived')}
						triggerClass={selectTrigger}
					/>
					{#if $errors.status}
						<span class={errorText}>{$errors.status}</span>
					{/if}
				</div>
			</div>

			<div class={field}>
				<label class={label} for="description">Description</label>
				<textarea
					class={textarea}
					id="description"
					name="description"
					placeholder="Event details and highlights..."
					bind:value={$form.description}
					{...$constraints.description}
				/>
				{#if $errors.description}
					<span class={errorText}>{$errors.description}</span>
				{/if}
			</div>

			<div class={buttonGroup}>
				<Button variant="solid" type="submit" disabled={$delayed}>
					{submitLabel}
				</Button>
				<a href="/dashboard">
					<Button variant="outline" type="button">Cancel</Button>
				</a>
			</div>
		</form>
	</div>
</div>
