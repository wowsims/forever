export interface BooleanPickerOptions {
	/** Label first, checkbox last. */
	reverse?: boolean;
	/** Draws the box with a danger border and marks the checkbox `aria-invalid`. */
	invalid?: boolean;
	/** Ids of extra text that describes the checkbox, such as an sr-only note; a field `description` is added to them. */
	'aria-describedby'?: string;
	/**
	 * Keeps a disabled box in the tab order, still marked disabled and not toggled by a click, Space or Enter, and opens
	 * the label tooltip on its focus. The tooltip needs the field `id`, which names it.
	 */
	focusableWhenDisabled?: boolean;
}
