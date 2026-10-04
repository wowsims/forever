import { act, render } from '@testing-library/react';
import { StrictMode } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { useHideZeroRestore } from './useHideZeroRestore';

interface ProbeProps {
	hidden: boolean;
	revision: number;
	value: number;
	setValue: (next: number) => void;
	restoreWhen?: (current: number) => boolean;
	onSettled?: (current: number) => void;
}

const Probe = ({ hidden, revision, value, setValue, restoreWhen, onSettled }: ProbeProps) => {
	useHideZeroRestore({ value, setValue, hidden, revision }, 0, restoreWhen, onSettled);
	return null;
};

describe('useHideZeroRestore', () => {
	it.each([2, 1])('stores %i and zeroes the source when it goes hidden, then restores exactly that, leaving normalisation to the caller', value => {
		const setValue = vi.fn();
		const { rerender } = render(<Probe hidden={false} revision={0} value={value} setValue={setValue} />);

		rerender(<Probe hidden setValue={setValue} revision={1} value={value} />);
		expect(setValue).toHaveBeenCalledTimes(1);
		expect(setValue).toHaveBeenLastCalledWith(0);

		rerender(<Probe hidden={false} setValue={setValue} revision={2} value={0} />);
		expect(setValue).toHaveBeenCalledTimes(2);
		expect(setValue).toHaveBeenLastCalledWith(value);
	});

	it('zeroes a picker that mounts hidden at the first notification, not at mount', () => {
		const setValue = vi.fn();
		const { rerender } = render(<Probe hidden revision={0} value={2} setValue={setValue} />);
		expect(setValue).not.toHaveBeenCalled();

		rerender(<Probe hidden revision={1} value={2} setValue={setValue} />);
		expect(setValue).toHaveBeenCalledTimes(1);
		expect(setValue).toHaveBeenLastCalledWith(0);

		rerender(<Probe hidden={false} revision={2} value={0} setValue={setValue} />);
		expect(setValue).toHaveBeenLastCalledWith(2);
	});

	it('still writes nothing at mount under StrictMode', () => {
		const setValue = vi.fn();
		const { rerender } = render(
			<StrictMode>
				<Probe hidden revision={0} value={2} setValue={setValue} />
			</StrictMode>,
		);
		expect(setValue).not.toHaveBeenCalled();

		act(() => {
			rerender(
				<StrictMode>
					<Probe hidden revision={1} value={2} setValue={setValue} />
				</StrictMode>,
			);
		});
		expect(setValue).toHaveBeenCalledTimes(1);
		expect(setValue).toHaveBeenLastCalledWith(0);
	});

	it('settles on the value each notification leaves behind: the zero, the restored value, or the value as it stands', () => {
		const setValue = vi.fn();
		const onSettled = vi.fn();
		const probe = (props: Pick<ProbeProps, 'hidden' | 'revision' | 'value'>) => (
			<Probe {...props} setValue={setValue} onSettled={onSettled} restoreWhen={current => current === 0} />
		);
		const { rerender } = render(probe({ hidden: false, revision: 0, value: 2 }));
		expect(onSettled).not.toHaveBeenCalled();

		rerender(probe({ hidden: true, revision: 1, value: 2 }));
		expect(onSettled).toHaveBeenLastCalledWith(0);

		rerender(probe({ hidden: false, revision: 2, value: 0 }));
		expect(onSettled).toHaveBeenLastCalledWith(2);

		rerender(probe({ hidden: false, revision: 3, value: 5 }));
		expect(onSettled).toHaveBeenLastCalledWith(5);
		expect(onSettled).toHaveBeenCalledTimes(3);
	});

	it('settles on the value the source moved to when restoreWhen refuses the restore', () => {
		const onSettled = vi.fn();
		const { rerender } = render(
			<Probe hidden={false} revision={0} value={2} setValue={vi.fn()} onSettled={onSettled} restoreWhen={current => current === 0} />,
		);

		rerender(<Probe hidden revision={1} value={2} setValue={vi.fn()} onSettled={onSettled} restoreWhen={current => current === 0} />);
		rerender(<Probe hidden={false} revision={2} value={3} setValue={vi.fn()} onSettled={onSettled} restoreWhen={current => current === 0} />);
		expect(onSettled).toHaveBeenLastCalledWith(3);
	});

	it('does not settle at mount or on a StrictMode replay of the same revision', () => {
		const onSettled = vi.fn();
		const { rerender } = render(
			<StrictMode>
				<Probe hidden revision={0} value={2} setValue={vi.fn()} onSettled={onSettled} />
			</StrictMode>,
		);
		expect(onSettled).not.toHaveBeenCalled();

		act(() => {
			rerender(
				<StrictMode>
					<Probe hidden revision={1} value={2} setValue={vi.fn()} onSettled={onSettled} />
				</StrictMode>,
			);
		});
		expect(onSettled).toHaveBeenCalledOnce();
	});
});
