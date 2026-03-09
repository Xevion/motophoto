<script lang="ts">
import { css, cx } from 'styled-system/css';
import { CircleAlert, CircleCheck } from '@lucide/svelte';
import { slide } from 'svelte/transition';

interface Props {
	variant?: 'error' | 'success';
	message: string;
}

const { variant = 'error', message }: Props = $props();

const root = css({
	display: 'flex',
	alignItems: 'flex-start',
	gap: '2.5',
	p: '3',
	borderRadius: 'lg',
	borderWidth: '1px',
	fontSize: 'sm',
	lineHeight: 'snug',
});

const errorVariant = css({
	bg: 'danger.subtle',
	color: 'danger.subtleFg',
	borderColor: 'danger.border',
});

const successVariant = css({
	bg: 'success.subtle',
	color: 'success.subtleFg',
	borderColor: 'success.border',
});

const icon = css({
	flexShrink: 0,
	mt: '0.5',
});
</script>

<div
	transition:slide={{ duration: 200 }}
	class={cx(root, variant === 'error' ? errorVariant : successVariant)}
	role="alert"
>
	{#if variant === 'error'}
		<CircleAlert size={16} class={icon} />
	{:else}
		<CircleCheck size={16} class={icon} />
	{/if}
	<span>{message}</span>
</div>
