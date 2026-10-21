export interface BooleanPickerOptions {
	/** Label first, checkbox last. */
	reverse?: boolean;
	/** Draws the box with a danger border and marks the checkbox `aria-invalid`. */
	invalid?: boolean;
	/** Ids of extra text that describes the checkbox, such as an sr-only note; a field `description` is added to them. */
	'aria-describedby'?: string;
}
