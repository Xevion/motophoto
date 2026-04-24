<script lang="ts">
	import { css } from 'styled-system/css';
	import Button from '$lib/components/ui/button.svelte';
	import X from '@lucide/svelte/icons/x';

	interface Props {
		minTime?: string;
		maxTime?: string;
		onFilterChange?: (takenAfter: string | null, takenBefore: string | null) => void;
	}

	let { minTime, maxTime, onFilterChange }: Props = $props();

	let startDateTime = $state<string>('');
	let endDateTime = $state<string>('');

	function formatDateTimeLocal(isoString: string | undefined): string {
		if (!isoString) return '';
		const date = new Date(isoString);
		// Format as YYYY-MM-DDTHH:mm for datetime-local input
		return date.toISOString().slice(0, 16);
	}

	function handleStartChange() {
		applyFilters();
	}

	function handleEndChange() {
		applyFilters();
	}

	function applyFilters() {
		const takenAfter = startDateTime ? new Date(startDateTime).toISOString() : null;
		const takenBefore = endDateTime ? new Date(endDateTime).toISOString() : null;
		onFilterChange?.(takenAfter, takenBefore);
	}

	function clearFilters() {
		startDateTime = '';
		endDateTime = '';
		onFilterChange?.(null, null);
	}

	const container = css({
		display: 'flex',
		flexDirection: 'column',
		gap: '4',
		p: '4',
		bg: 'bg.muted',
		borderRadius: 'lg',
	});

	const inputGroup = css({
		display: 'flex',
		flexDirection: 'column',
		gap: '2',
		md: { flexDirection: 'row', gap: '4' },
	});

	const inputWrapper = css({
		display: 'flex',
		flexDirection: 'column',
		gap: '1',
		flex: 1,
	});

	const label = css({
		fontSize: 'sm',
		fontWeight: 'semibold',
		color: 'fg.muted',
	});

	const input = css({
		px: '3',
		py: '2',
		bg: 'bg',
		border: '1px solid',
		borderColor: 'border',
		borderRadius: 'md',
		fontSize: 'sm',
		colorScheme: 'light dark',
	});

	const buttonGroup = css({
		display: 'flex',
		gap: '2',
		flexDirection: 'column',
		md: { flexDirection: 'row' },
	});

	const info = css({
		fontSize: 'xs',
		color: 'fg.muted',
		mt: '2',
	});
</script>

<style>
	:global(.dark) :global(input[type='datetime-local']::webkit-calendar-picker-indicator) {
		filter: invert(0.8);
	}
</style>

<div class={container}>
	<div class={inputGroup}>
		<div class={inputWrapper}>
			<label class={label} for="start-time">From</label>
			<input
				id="start-time"
				type="datetime-local"
				bind:value={startDateTime}
				onchange={handleStartChange}
				max={endDateTime || formatDateTimeLocal(maxTime)}
				class={input}
			/>
		</div>
		<div class={inputWrapper}>
			<label class={label} for="end-time">To</label>
			<input
				id="end-time"
				type="datetime-local"
				bind:value={endDateTime}
				onchange={handleEndChange}
				min={startDateTime || formatDateTimeLocal(minTime)}
				class={input}
			/>
		</div>
	</div>

	{#if startDateTime || endDateTime}
		<div class={buttonGroup}>
			<Button onclick={clearFilters} variant="outline" size="sm">
				<X size={16} />
				Clear filters
			</Button>
		</div>
	{/if}

	{#if minTime && maxTime}
		<div class={info}>
			Available: {new Date(minTime).toLocaleDateString()} to {new Date(maxTime).toLocaleDateString()}
		</div>
	{/if}
</div>
