<script lang="ts">
import { css } from 'styled-system/css';
import Button from '$lib/components/ui/button.svelte';
import { LogOut, Plus } from '@lucide/svelte';
import type { LayoutData } from './$types';

const { data }: { data: LayoutData } = $props();

const layout = css({
	minH: 'screen',
	bg: 'bg',
	display: 'flex',
	flexDirection: 'column',
});

const header = css({
	bg: 'bg.subtle',
	borderBottomWidth: '1px',
	borderBottomColor: 'border',
	px: '6',
	py: '4',
});

const headerContent = css({
	maxW: '7xl',
	mx: 'auto',
	display: 'flex',
	justifyContent: 'space-between',
	alignItems: 'center',
});

const headerTitle = css({
	fontSize: 'xl',
	fontWeight: 'semibold',
});

const headerRight = css({
	display: 'flex',
	gap: '2',
	alignItems: 'center',
});

const userInfo = css({
	fontSize: 'sm',
	color: 'fg.muted',
});

const content = css({
	flex: '1',
	px: '6',
	py: '8',
});

const contentMax = css({
	maxW: '7xl',
	mx: 'auto',
});
</script>

<div class={layout}>
	<header class={header}>
		<div class={headerContent}>
			<h1 class={headerTitle}>Photographer Dashboard</h1>
			<div class={headerRight}>
				<span class={userInfo}>{data.user.email}</span>
				<form action="/logout" method="POST">
					<Button variant="ghost" size="sm" type="submit">
						<LogOut />
					</Button>
				</form>
			</div>
		</div>
	</header>

	<main class={content}>
		<div class={contentMax}>
			<slot />
		</div>
	</main>
</div>
