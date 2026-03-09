import { css } from 'styled-system/css';

export const authContainer = css({
	display: 'flex',
	justifyContent: 'center',
	py: '12',
});

export const authCard = css({
	w: 'full',
	maxW: '400px',
	bg: 'bg.subtle',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'xl',
	p: '8',
	display: 'flex',
	flexDirection: 'column',
	gap: '6',
});

export const authTitle = css({
	fontSize: '2xl',
	fontWeight: 'bold',
	textAlign: 'center',
});

export const authFieldGroup = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '1.5',
});

export const authLabel = css({
	fontSize: 'sm',
	fontWeight: 'medium',
});

export const authInputWrapper = css({
	display: 'flex',
	alignItems: 'center',
	gap: '2',
	bg: 'bg',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'md',
	px: '3',
	h: '10',
	transition: 'border-color',
	transitionDuration: '150ms',
	_focusWithin: {
		borderColor: 'primary',
		outlineWidth: '1px',
		outlineColor: 'primary',
		outlineStyle: 'solid',
	},
});

export const authInputWrapperError = css({
	borderColor: 'danger',
	_focusWithin: {
		borderColor: 'danger',
		outlineColor: 'danger',
	},
});

export const authInput = css({
	flex: '1',
	bg: 'transparent',
	border: 'none',
	outline: 'none',
	color: 'fg',
	fontSize: 'sm',
	h: 'full',
	'&::placeholder': {
		color: 'fg.subtle',
	},
});

export const authIcon = css({
	color: 'fg.muted',
	flexShrink: 0,
});

export const authErrorText = css({
	fontSize: 'xs',
	color: 'danger',
});

export const authFooterText = css({
	fontSize: 'sm',
	color: 'fg.muted',
	textAlign: 'center',
});

export const authFooterLink = css({
	color: 'primary',
	fontWeight: 'medium',
	_hover: { textDecoration: 'underline' },
});

export const authFormFields = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '4',
});

export const authVisibilityToggle = css({
	display: 'flex',
	alignItems: 'center',
	justifyContent: 'center',
	bg: 'transparent',
	border: 'none',
	borderRadius: 'sm',
	p: '0.5',
	cursor: 'pointer',
	color: 'fg.muted',
	flexShrink: 0,
	transition: 'color 150ms',
	_hover: { color: 'fg' },
	_focusVisible: {
		outlineWidth: '2px',
		outlineColor: 'primary',
		outlineOffset: '1px',
		outlineStyle: 'solid',
	},
	'& svg': {
		width: '1em',
		height: '1em',
	},
});
