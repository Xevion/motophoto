import { defineRecipe } from '@pandacss/dev';

export const badgeRecipe = defineRecipe({
	className: 'badge',
	description: 'Badge / pill component styles',
	base: {
		display: 'inline-flex',
		alignItems: 'center',
		borderRadius: 'full',
		fontSize: 'xs',
		fontWeight: 'medium',
		lineHeight: 'tight',
		px: '2.5',
		py: '0.5',
		whiteSpace: 'nowrap',
	},
	variants: {
		variant: {
			default: {
				bg: 'primary',
				color: 'primary.fg',
			},
			secondary: {
				bg: 'bg.muted',
				color: 'fg.muted',
			},
			outline: {
				bg: 'transparent',
				color: 'fg',
				borderWidth: '1px',
				borderColor: 'border',
			},
			success: {
				bg: 'success',
				color: 'success.fg',
			},
			danger: {
				bg: 'danger',
				color: 'danger.fg',
			},
		},
	},
	defaultVariants: {
		variant: 'default',
	},
});
