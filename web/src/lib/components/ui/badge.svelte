<script lang="ts">
import type { Snippet } from 'svelte';
import { badge } from 'styled-system/recipes';
import { cx } from '$lib/utils';

type BadgeVariant = 'default' | 'secondary' | 'outline' | 'success' | 'danger';

let {
	variant = 'default',
	class: className = '',
	children,
}: {
	variant?: BadgeVariant;
	class?: string;
	children?: Snippet;
} = $props();

const classes = $derived(cx(badge({ variant }), className));
</script>

<span class={classes}>
  {@render children?.()}
</span>
