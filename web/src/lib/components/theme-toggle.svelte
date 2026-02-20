<script lang="ts">
import { themeStore } from '$lib/stores/theme.svelte';
import Sun from '@lucide/svelte/icons/sun';
import Moon from '@lucide/svelte/icons/moon';
import { Button } from '$lib/components/ui/button/index.js';
import { tick } from 'svelte';

/**
 * Theme toggle with View Transitions API circular reveal animation.
 * The clip-path circle expands from the click point to cover the viewport.
 */
function handleToggle(event: MouseEvent) {
	const supportsViewTransition =
		typeof document !== 'undefined' &&
		'startViewTransition' in document &&
		!window.matchMedia('(prefers-reduced-motion: reduce)').matches;

	if (!supportsViewTransition) {
		themeStore.toggle();
		return;
	}

	const x = event.clientX;
	const y = event.clientY;
	const endRadius = Math.hypot(Math.max(x, innerWidth - x), Math.max(y, innerHeight - y));

	document.documentElement.classList.add('theme-transitioning');

	const transition = document.startViewTransition(async () => {
		themeStore.toggle();
		await tick();
	});

	void transition.finished.finally(() => {
		document.documentElement.classList.remove('theme-transitioning');
	});

	void transition.ready.then(() => {
		document.documentElement.animate(
			{
				clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${endRadius}px at ${x}px ${y}px)`],
			},
			{
				duration: 500,
				easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
				pseudoElement: '::view-transition-new(root)',
			},
		);
	});
}
</script>

<Button onclick={handleToggle} variant="outline" size="icon">
	<Sun
		class="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0"
	/>
	<Moon
		class="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100"
	/>
	<span class="sr-only">Toggle theme</span>
</Button>
