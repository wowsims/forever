import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ArmoryImporterDialog } from './ArmoryImporterDialog';

const importArmoryCharacter = vi.hoisted(() => vi.fn());
vi.mock('../../importers/armory', async importOriginal => ({
	...(await importOriginal<typeof import('../../importers/armory')>()),
	importArmoryCharacter,
}));

const toasts = vi.hoisted(() => [] as Array<{ variant: string; body: unknown }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { variant: string; body: unknown }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));

const field = (testId: string) => screen.getByTestId(testId).querySelector('input') ?? (screen.getByTestId(testId) as HTMLInputElement);
const enter = (testId: string, text: string) => {
	const input = field(testId);
	fireEvent.input(input, { target: { value: text } });
	fireEvent.change(input, { target: { value: text } });
};
const clickImport = async () => {
	await act(async () => {
		fireEvent.click(screen.getByTestId('import-button'));
	});
};

describe('ArmoryImporterDialog', () => {
	beforeEach(() => {
		toasts.length = 0;
		importArmoryCharacter.mockReset().mockResolvedValue(undefined);
		localStorage.clear();
	});

	it('looks the character up and closes', async () => {
		const onOpenChange = vi.fn();
		renderWithHost(<ArmoryImporterDialog open onOpenChange={onOpenChange} />);
		enter('armory-importer-realm', 'ClassicBetaPvE2');
		enter('armory-importer-name', 'Erlyn Farthing');

		await clickImport();

		expect(importArmoryCharacter).toHaveBeenCalledWith(expect.anything(), { region: 'us', realm: 'ClassicBetaPvE2', name: 'Erlyn-Farthing' });
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});

	it('fills the realm from a name pasted as the game writes it in logs', async () => {
		renderWithHost(<ArmoryImporterDialog open onOpenChange={() => {}} />);
		enter('armory-importer-name', 'Erlyn-Farthing-ClassicBetaPvE2');

		expect(field('armory-importer-realm').value).toBe('ClassicBetaPvE2');
		expect(field('armory-importer-name').value).toBe('Erlyn-Farthing');
	});

	it('asks for the missing fields without looking anything up', async () => {
		renderWithHost(<ArmoryImporterDialog open onOpenChange={() => {}} />);
		enter('armory-importer-name', 'Erlyn Farthing');

		await clickImport();

		expect(importArmoryCharacter).not.toHaveBeenCalled();
		expect(toasts).toEqual([{ variant: 'error', body: 'import.armory.errors.missing_fields' }]);
	});

	it('shows a failed lookup as an error and stays open', async () => {
		importArmoryCharacter.mockRejectedValue(new Error('No character'));
		const onOpenChange = vi.fn();
		renderWithHost(<ArmoryImporterDialog open onOpenChange={onOpenChange} />);
		enter('armory-importer-realm', 'Normal');
		enter('armory-importer-name', 'Erlyn Farthing');

		await clickImport();

		expect(toasts).toEqual([{ variant: 'error', body: 'import.error No character' }]);
		expect(onOpenChange).not.toHaveBeenCalled();
	});
});
