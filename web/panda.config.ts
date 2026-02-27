import { defineConfig } from '@pandacss/dev';
import { buttonRecipe } from './src/lib/recipes/button';
import { badgeRecipe } from './src/lib/recipes/badge';
import { toggleRecipe } from './src/lib/recipes/toggle';

export default defineConfig({
	preflight: true,

	include: ['./src/**/*.{js,ts,svelte}'],
	exclude: [],

	outdir: 'styled-system',

	jsxFramework: undefined,

	theme: {
		extend: {
			recipes: {
				button: buttonRecipe,
				badge: badgeRecipe,
			},
			slotRecipes: {
				toggle: toggleRecipe,
			},
			tokens: {
				colors: {
					orange: {
						50: { value: 'oklch(0.98 0.016 73)' },
						100: { value: 'oklch(0.954 0.038 75.164)' },
						200: { value: 'oklch(0.901 0.076 70.697)' },
						300: { value: 'oklch(0.837 0.128 66.29)' },
						400: { value: 'oklch(0.75 0.183 55.934)' },
						500: { value: 'oklch(0.705 0.191 47.604)' },
						600: { value: 'oklch(0.646 0.222 41.116)' },
						700: { value: 'oklch(0.553 0.195 38.402)' },
						800: { value: 'oklch(0.47 0.157 37.304)' },
						900: { value: 'oklch(0.408 0.123 38.172)' },
						950: { value: 'oklch(0.266 0.079 36.259)' },
					},
					neutral: {
						0: { value: 'oklch(1 0 0)' },
						50: { value: 'oklch(0.985 0 0)' },
						100: { value: 'oklch(0.97 0 0)' },
						200: { value: 'oklch(0.922 0 0)' },
						300: { value: 'oklch(0.87 0 0)' },
						400: { value: 'oklch(0.708 0 0)' },
						500: { value: 'oklch(0.556 0 0)' },
						600: { value: 'oklch(0.439 0 0)' },
						700: { value: 'oklch(0.371 0 0)' },
						800: { value: 'oklch(0.269 0 0)' },
						850: { value: 'oklch(0.205 0 0)' },
						900: { value: 'oklch(0.145 0 0)' },
						950: { value: 'oklch(0.09 0 0)' },
					},
				},
				radii: {
					sm: { value: '0.25rem' },
					md: { value: '0.375rem' },
					lg: { value: '0.5rem' },
					xl: { value: '0.75rem' },
					'2xl': { value: '1rem' },
					full: { value: '9999px' },
				},
				fontSizes: {
					xs: { value: '0.75rem' },
					sm: { value: '0.875rem' },
					md: { value: '1rem' },
					lg: { value: '1.125rem' },
					xl: { value: '1.25rem' },
					'2xl': { value: '1.5rem' },
					'3xl': { value: '1.875rem' },
					'4xl': { value: '2.25rem' },
				},
				fontWeights: {
					normal: { value: '400' },
					medium: { value: '500' },
					semibold: { value: '600' },
					bold: { value: '700' },
				},
				lineHeights: {
					tight: { value: '1.25' },
					snug: { value: '1.375' },
					normal: { value: '1.5' },
					relaxed: { value: '1.625' },
				},
				spacing: {
					'0': { value: '0' },
					'1': { value: '0.25rem' },
					'2': { value: '0.5rem' },
					'3': { value: '0.75rem' },
					'4': { value: '1rem' },
					'5': { value: '1.25rem' },
					'6': { value: '1.5rem' },
					'8': { value: '2rem' },
					'10': { value: '2.5rem' },
					'12': { value: '3rem' },
					'14': { value: '3.5rem' },
					'16': { value: '4rem' },
					'20': { value: '5rem' },
					'24': { value: '6rem' },
				},
				shadows: {
					sm: { value: '0 1px 2px 0 rgb(0 0 0 / 0.05)' },
					md: { value: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)' },
					lg: { value: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)' },
				},
			},
			semanticTokens: {
				colors: {
					bg: {
						DEFAULT: { value: { base: '{colors.neutral.0}', _dark: '{colors.neutral.900}' } },
						subtle: { value: { base: '{colors.neutral.50}', _dark: '{colors.neutral.850}' } },
						muted: { value: { base: '{colors.neutral.100}', _dark: '{colors.neutral.800}' } },
						emphasis: { value: { base: '{colors.neutral.200}', _dark: '{colors.neutral.700}' } },
					},
					fg: {
						DEFAULT: { value: { base: '{colors.neutral.900}', _dark: '{colors.neutral.50}' } },
						muted: { value: { base: '{colors.neutral.500}', _dark: '{colors.neutral.400}' } },
						subtle: { value: { base: '{colors.neutral.400}', _dark: '{colors.neutral.600}' } },
					},
					border: {
						DEFAULT: { value: { base: '{colors.neutral.200}', _dark: 'oklch(1 0 0 / 10%)' } },
						strong: { value: { base: '{colors.neutral.300}', _dark: 'oklch(1 0 0 / 20%)' } },
					},
					primary: {
						DEFAULT: { value: { base: '{colors.orange.500}', _dark: '{colors.orange.500}' } },
						hover: { value: { base: '{colors.orange.600}', _dark: '{colors.orange.400}' } },
						fg: { value: { base: '{colors.neutral.0}', _dark: '{colors.neutral.900}' } },
						subtle: { value: { base: '{colors.orange.50}', _dark: '{colors.orange.950}' } },
						subtleFg: { value: { base: '{colors.orange.700}', _dark: '{colors.orange.300}' } },
					},
					danger: {
						DEFAULT: { value: { base: 'oklch(0.577 0.245 27.325)', _dark: 'oklch(0.704 0.191 22.216)' } },
						fg: { value: { base: '{colors.neutral.0}', _dark: '{colors.neutral.0}' } },
					},
					success: {
						DEFAULT: { value: { base: 'oklch(0.166 0.044 156.743)', _dark: 'oklch(0.266 0.065 156.743)' } },
						fg: { value: { base: 'oklch(0.696 0.17 162.48)', _dark: 'oklch(0.696 0.17 162.48)' } },
					},
				},
			},
			breakpoints: {
				sm: '640px',
				md: '768px',
				lg: '1024px',
				xl: '1280px',
			},
		},
	},

	globalCss: {
		'*': {
			boxSizing: 'border-box',
			borderColor: 'border',
		},
		body: {
			bg: 'bg',
			color: 'fg',
			fontFamily: 'system-ui, -apple-system, sans-serif',
			lineHeight: 'normal',
		},
		a: {
			color: 'inherit',
			textDecoration: 'none',
		},
	},
});
