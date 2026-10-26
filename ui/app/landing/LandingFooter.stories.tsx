import type { Meta, StoryObj } from '@storybook/react-vite';
import { VERSIONS_MANIFEST, withVersionsManifest } from '@testing/versions_manifest';
import { within } from 'storybook/test';

import { LandingFooter } from './LandingFooter';

const meta = {
	title: 'App/Landing/Components/LandingFooter',
	component: LandingFooter,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
	beforeEach: () => withVersionsManifest(VERSIONS_MANIFEST),
	// The footer renders once the manifest fetch resolves
	play: async ({ canvasElement }) => {
		await within(canvasElement).findByRole('contentinfo');
	},
} satisfies Meta<typeof LandingFooter>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Content from the shared manifest, with a version that is not live yet and a two-column group. */
export const Default: Story = {};
