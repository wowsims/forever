import './storybook.css';
import '@i18n/config';
import '../ui/shared/page_boot';

import type { Preview } from '@storybook/react-vite';
import { configureActions } from 'storybook/actions';
import { themes } from 'storybook/theming';

import { SPEC_THEMES } from '../ui/design-system/tokens/spec_themes';
import { withSimUi } from '../ui/testing/storybook/decorators';
import { DEFINITIONS } from '../ui/ui-kit/Icon/definitions';

(globalThis as unknown as { gtag: () => void }).gtag = () => {};
// The Actions panel logs an fn() call as its whole argument list, so a React event arrives unrecognised
// and its `view` (the window) reaches the channel serialiser, which throws on the window's cycles.
// Depth 3 stops the serialiser short of them; 5 and deeper throws.
configureActions({ depth: 3 });

const preview: Preview = {
	parameters: {
		layout: 'fullscreen',
		docs: { theme: themes.dark },
		controls: { matchers: { color: /(background|color)$/i } },
		options: { storySort: { method: 'alphabetical' } },
	},
	// A component that declares an `icon` prop picks it from the kit's FontAwesome set; a story whose `icon` is an object opts out with `control: false`.
	argTypesEnhancers: [
		context => {
			const icon = context.argTypes.icon;
			if (!icon || icon.control === false) return context.argTypes;
			return { ...context.argTypes, icon: { ...icon, control: { type: 'select' }, options: Object.keys(DEFINITIONS) } };
		},
	],
	globalTypes: {
		spec: {
			description: 'Which <spec>-sim-ui class frames the story',
			toolbar: {
				title: 'Spec',
				icon: 'paintbrush',
				dynamicTitle: true,
				items: [
					{ value: 'none', title: 'none (landing page)' },
					{ value: 'default', title: 'sim-ui default' },
					...SPEC_THEMES.map(theme => ({ value: theme.className, title: theme.name })),
				],
			},
		},
	},
	initialGlobals: { spec: 'default' },
	decorators: [withSimUi],
};

export default preview;
