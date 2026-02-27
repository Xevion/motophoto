import { defineRecipe } from '@pandacss/dev';

export const buttonRecipe = defineRecipe({
	className: 'button',
	description: 'Button component styles',
	base: {
		display: 'inline-flex',
		alignItems: 'center',
		justifyContent: 'center',
		gap: '2',
		borderRadius: 'md',
		fontSize: 'sm',
		fontWeight: 'medium',
		whiteSpace: 'nowrap',
		transition: 'all',
		transitionDuration: '150ms',
		cursor: 'pointer',
		userSelect: 'none',
		outline: 'none',
		border: 'none',
		textDecoration: 'none',
		_disabled: { opacity: 0.5, pointerEvents: 'none' },
		_focusVisible: {
			outlineWidth: '2px',
			outlineColor: 'primary',
			outlineOffset: '2px',
			outlineStyle: 'solid',
		},
		'& svg': {
			pointerEvents: 'none',
			flexShrink: 0,
			width: '1em',
			height: '1em',
		},
	},
	variants: {
		variant: {
			solid: {
				bg: 'primary',
				color: 'primary.fg',
				_hover: { bg: 'primary.hover' },
			},
			outline: {
				bg: 'transparent',
				color: 'fg',
				borderWidth: '1px',
				borderColor: 'border',
				_hover: { bg: 'bg.muted' },
			},
			ghost: {
				bg: 'transparent',
				color: 'fg',
				_hover: { bg: 'bg.muted' },
			},
			danger: {
				bg: 'danger',
				color: 'danger.fg',
				_hover: { opacity: '0.9' },
			},
			link: {
				bg: 'transparent',
				color: 'primary',
				_hover: { textDecoration: 'underline' },
				p: '0',
				h: 'auto',
			},
		},
		size: {
			sm: { h: '8', px: '3', fontSize: 'xs' },
			md: { h: '9', px: '4', fontSize: 'sm' },
			lg: { h: '10', px: '6', fontSize: 'md' },
			icon: { h: '9', w: '9', p: '0' },
			'icon-sm': { h: '8', w: '8', p: '0' },
		},
	},
	defaultVariants: {
		variant: 'solid',
		size: 'md',
	},
});
