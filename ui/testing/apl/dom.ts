import { createEvent, fireEvent } from '@testing-library/react';
import { vi } from 'vitest';

/** happy-dom has no `IntersectionObserver`; this one records every instance so a test can fire its callback. */
export class FakeIntersectionObserver {
	static instances: Array<FakeIntersectionObserver> = [];
	constructor(
		readonly callback: IntersectionObserverCallback,
		readonly options?: IntersectionObserverInit,
	) {
		FakeIntersectionObserver.instances.push(this);
	}
	observe = vi.fn();
	disconnect = vi.fn();
	unobserve = vi.fn();
	fire(...entries: Array<Partial<IntersectionObserverEntry>>) {
		this.callback(entries as Array<IntersectionObserverEntry>, this as unknown as IntersectionObserver);
	}
}

// happy-dom lays nothing out, so each row is given a box: `index` rows of 20px from the top.
export const place = (rows: Array<HTMLElement>, offset = 0) =>
	rows.forEach((element, index) => vi.spyOn(element, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, (offset + index) * 20, 200, 20)));

// happy-dom's DragEvent is a plain Event, which drops `clientY` from its init.
export const dragEvent = (type: 'dragOver' | 'drop', target: HTMLElement, y: number) => {
	const event = createEvent[type](target, { dataTransfer: new DataTransfer() });
	Object.defineProperty(event, 'clientY', { value: y });
	return fireEvent(target, event);
};
