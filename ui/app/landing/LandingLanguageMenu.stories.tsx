import { supportedLanguages } from '@i18n/locale_service';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';

import { LandingLanguageMenu } from './LandingLanguageMenu';

const meta = {
	title: 'App/Landing/Components/LandingLanguageMenu',
	component: LandingLanguageMenu,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
	decorators: [
		Story => (
			<div className="flex justify-end">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof LandingLanguageMenu>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ canvasElement }) => {
		await userEvent.click(canvasElement.querySelector<HTMLElement>('#languageDropdown')!);
		const items = within(canvasElement.ownerDocument.body).getAllByTestId('dropdown-item');
		await expect(items).toHaveLength(Object.keys(supportedLanguages).length);
	},
};
