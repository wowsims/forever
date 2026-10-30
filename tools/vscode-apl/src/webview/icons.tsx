// Line icons drawn at 16px in the editor's foreground colour.
const PATHS = {
	grip: 'M6 3.5h.01M10 3.5h.01M6 8h.01M10 8h.01M6 12.5h.01M10 12.5h.01',
	eye: 'M1.5 8s2.5-4.5 6.5-4.5S14.5 8 14.5 8 12 12.5 8 12.5 1.5 8 1.5 8zM8 10a2 2 0 1 0 0-4 2 2 0 0 0 0 4z',
	eyeOff: 'M2 2l12 12M6.5 6.6A2 2 0 0 0 9.4 9.5M4.2 4.4C2.5 5.6 1.5 8 1.5 8s2.5 4.5 6.5 4.5c1.3 0 2.4-.4 3.4-1M7 3.6c.3 0 .7-.1 1-.1 4 0 6.5 4.5 6.5 4.5s-.5 1-1.5 2.1',
	copy: 'M5.5 5.5h8v8h-8zM10.5 5.5v-3h-8v8h3',
	trash: 'M2.5 4h11M6 4V2.5h4V4M4 4l.7 9.5h6.6L12 4',
	close: 'M4 4l8 8M12 4l-8 8',
	up: 'M8 13V3M4 7l4-4 4 4',
	down: 'M4 6l4 4 4-4',
	right: 'M6 4l4 4-4 4',
	jump: 'M9 3h4v4M13 3L7 9M11 9.5V13H3V5h3.5',
	check: 'M3 8.5l3.5 3.5L13 4.5',
	expand: 'M2.5 8h11M5 5.5L2.5 8 5 10.5M11 5.5l2.5 2.5-2.5 2.5',
	extract: 'M3 3.5h5M3 8h5M3 12.5h5M10.5 5.5l3 2.5-3 2.5',
	block: 'M3 3v10M3 3h3M3 13h3M8 6h5M8 10h5',
	unblock: 'M3 3v10M3 3h3M3 13h3M9 6l4 4M13 6l-4 4',
};

export function Icon({ name }: { name: keyof typeof PATHS }) {
	return (
		<svg
			className="icon"
			width="16"
			height="16"
			viewBox="0 0 16 16"
			fill="none"
			stroke="currentColor"
			strokeWidth={name === 'grip' ? 2.4 : 1.3}
			strokeLinecap="round"
			strokeLinejoin="round"
			aria-hidden="true">
			<path d={PATHS[name]} />
		</svg>
	);
}
