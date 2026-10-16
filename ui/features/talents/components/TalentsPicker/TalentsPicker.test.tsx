import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import type { TalentsConfig } from '@sim/talents/config';
import { newTalentsConfig } from '@sim/talents/config';
import { mageTalentsConfig } from '@sim/talents/mage';
import { warriorTalentsConfig } from '@sim/talents/warrior';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/hooks/useActionId', () => ({
	useActionId: () => ({ iconUrl: '', name: '', href: '', ready: true }),
}));

const { TalentsHeaderView, TalentsPicker } = await import('./TalentsPicker');

type Fake = Record<string, never>;

const COLS = 4;

const tree = (name: string) => ({
	name,
	backgroundUrl: `${name}.jpg`,
	talents: Array.from({ length: 5 * COLS }, (_, index) => ({
		fieldName: `talent${index}`,
		fancyName: `Talent ${index}`,
		location: { rowIdx: Math.floor(index / COLS), colIdx: index % COLS },
		spellId: 100 * index + 1,
		maxPoints: 5,
	})),
});

const trees: TalentsConfig<Fake> = newTalentsConfig<Fake>([tree('First'), tree('Second'), tree('Third')]);

let copied: Array<string>;
let talentsString: string;
const setValue = vi.fn((_player: Player<any>, next: string) => {
	talentsString = next;
});

const mount = (pickerTrees: TalentsConfig<any> = trees) => {
	const player = { getTalentsString: () => talentsString } as unknown as Player<any>;
	const host = { player } as never;
	const config = {
		id: 'talents-picker',
		trees: pickerTrees,
		getValue: () => talentsString,
		setValue,
	};
	return render(
		<SimHostProvider host={host}>
			<TalentsPicker config={config} />
		</SimHostProvider>,
	);
};

const talents = () => Array.from(document.querySelectorAll<HTMLAnchorElement>('[data-testid="talent-picker-root"]'));
const copyButton = () => document.querySelector<HTMLButtonElement>('[data-testid="talents-picker-actions"] button')!;
const remaining = () => document.querySelector('label')!.textContent!.match(/\d+$/)?.[0];

beforeEach(() => {
	vi.useFakeTimers();
	copied = [];
	talentsString = '';
	setValue.mockClear();
	vi.stubGlobal('navigator', {
		clipboard: {
			writeText: vi.fn((text: string) => {
				copied.push(text);
				return Promise.resolve();
			}),
		},
	});
});

afterEach(() => {
	vi.useRealTimers();
	vi.unstubAllGlobals();
});

describe('TalentsPicker container', () => {
	it('renders all three trees in one row', () => {
		mount();
		expect(document.querySelectorAll('[data-testid="talent-tree"]')).toHaveLength(3);
		expect(Array.from(document.querySelectorAll('[data-testid="talent-tree-title"]')).map(el => el.textContent)).toEqual(['First', 'Second', 'Third']);
	});

	it('counts down from the point cap across every tree', () => {
		talentsString = '5-5-5';
		mount();
		expect(remaining()).toBe('36');
	});

	it('writes the joined string when a talent in the last tree is clicked', () => {
		mount();
		const lastTreeTalents = talents().slice(2 * 5 * COLS);
		fireEvent.mouseDown(lastTreeTalents[0], { button: 0 });
		expect(setValue).toHaveBeenCalledWith(expect.anything(), '--1');
	});

	it('refuses the 52nd point anywhere', () => {
		talentsString = '5555555555-1';
		mount();
		expect(remaining()).toBe('0');

		const secondTree = talents().slice(5 * COLS);
		fireEvent.mouseDown(secondTree[2], { button: 0 });
		expect(setValue).toHaveBeenLastCalledWith(expect.anything(), '5555555555-1');
	});
});

describe('TalentsPicker round trip', () => {
	// A valid Forever mage build: 51 points, Fire left empty so the codec's empty-run
	// handling stays covered.
	const MAGE_DEFAULT = '25502252231221--2555';

	it('hands a real Forever default back byte-identically, empty middle tree and all', () => {
		talentsString = MAGE_DEFAULT;
		mount(mageTalentsConfig);

		expect(remaining()).toBe('0');
		fireEvent.mouseDown(talents()[0], { button: 0 });
		expect(setValue).toHaveBeenLastCalledWith(expect.anything(), MAGE_DEFAULT);
	});

	it('restores the original string after a point is taken back and re-spent', () => {
		talentsString = MAGE_DEFAULT;
		mount(mageTalentsConfig);

		// Most spent talents are load-bearing in a full build; this finds one the rules actually let go.
		const freeable = talents().findIndex((el, idx) => {
			if (Number(el.dataset.points) === 0) return false;
			fireEvent.mouseDown(talents()[idx], { button: 2 });
			return talentsString !== MAGE_DEFAULT;
		});
		expect(freeable).toBeGreaterThanOrEqual(0);

		fireEvent.mouseDown(talents()[freeable], { button: 0 });
		expect(talentsString).toBe(MAGE_DEFAULT);
	});
});

describe('TalentsPicker copy button', () => {
	it('copies the talents string as it reads at click time', () => {
		mount();
		talentsString = 'after';
		fireEvent.click(copyButton());
		expect(copied).toEqual(['after']);
	});
});

describe('TalentsPicker trees', () => {
	it('scrolls the middle tree into view on mount', () => {
		const scrollIntoView = vi.fn();
		vi.spyOn(Element.prototype, 'scrollIntoView').mockImplementation(scrollIntoView);

		mount();

		expect(scrollIntoView).toHaveBeenCalledTimes(1);
		expect(scrollIntoView.mock.instances[0]).toBe(screen.getAllByTestId('talent-tree')[1]);
	});

	it('builds the wowhead link from the generated trait data, def and rank included', () => {
		const talent = warriorTalentsConfig.flatMap(tree => tree.talents).find(candidate => candidate.definitionId)!;

		const href = ActionId.makeSpellUrl(talent.spellId, 2, talent.definitionId!);
		expect(href).toContain(`spell=${talent.spellId}`);
		expect(href).toContain(`def=${talent.definitionId}`);
		expect(href).toContain('rank=2');
	});
});

describe('TalentsHeaderView', () => {
	it('shows the points left to spend', () => {
		render(<TalentsHeaderView pointsRemaining={36} getTalentsString={() => ''} />);
		expect(document.querySelector('label')!.textContent!.match(/\d+$/)?.[0]).toBe('36');
	});
});
