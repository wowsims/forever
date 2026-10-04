import type { FieldFrameProps } from './types';

/** A field renders a `Field` root only for its frame: a label, a description, a layout or a clear slot. Without any, the control renders bare. */
export const hasFieldFrame = ({ label, description, layout }: Pick<FieldFrameProps, 'label' | 'description' | 'layout'>, clear?: object): boolean =>
	!!(label || description || layout || clear);
