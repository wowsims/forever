import { MAX_MELEE_RANGE, MIN_RANGED_RANGE } from '@sim/proto/sim_tables_auto_gen';

import { type MovementRun, pathOf } from '../../model/movement/movement';

export interface Palette {
	target: string;
	player: string;
	pet: string;
	ring: string;
	deadZone: string;
	text: string;
	muted: string;
	background: string;
	cast: string;
	deadZoneAlpha: number;
	tetherAlpha: number;
	petPathAlpha: number;
	font: string;
	/** The canvas's font size in pixels, which every size below is a multiple of. */
	em: number;
}

// A canvas has no CSS cascade, so each colour is read off the theme where the canvas sits, which keeps
// the spec's own primary colour. The font comes from the canvas's own text utility.
export function readPalette(canvas: HTMLCanvasElement): Palette {
	const style = getComputedStyle(canvas);
	const color = (name: string) => style.getPropertyValue(name).trim();
	const alpha = (name: string) => parseFloat(style.getPropertyValue(name));
	return {
		target: color('--color-danger'),
		player: color('--color-primary'),
		pet: color('--color-info'),
		ring: color('--color-muted'),
		deadZone: color('--color-danger'),
		text: color('--color-foreground'),
		muted: color('--color-muted'),
		background: color('--color-background'),
		cast: color('--color-timeline-cast'),
		deadZoneAlpha: alpha('--movement-dead-zone-opacity'),
		tetherAlpha: alpha('--movement-tether-opacity'),
		petPathAlpha: alpha('--movement-pet-path-opacity'),
		font: `${style.fontSize} ${style.fontFamily}`,
		em: parseFloat(style.fontSize),
	};
}

export interface FieldScene {
	run: MovementRun;
	/** Yards from the target of each of the run's tracks, in the same order. */
	yards: ReadonlyArray<number>;
	/** The label beside each track's circle, in the same order. */
	unitLabels: ReadonlyArray<string>;
	targetLabel: string;
	meleeLabel: string;
	rangedLabel: string;
	/** The player's casts under way, each a bar under its circle: the share of the cast done, and its label. */
	casting: ReadonlyArray<{ readonly done: number; readonly label: string }>;
	/** The icons of the player's casts that just went off, the latest first, each at the opacity it has faded to. */
	goneOff: ReadonlyArray<{ readonly icon: CanvasImageSource; readonly alpha: number }>;
}

// The rings worth drawing: melee range, and the ranged minimum only for a weapon that cannot fire inside it.
const breakpoints = (run: MovementRun): Array<number> => (run.deadZone ? [MAX_MELEE_RANGE, MIN_RANGED_RANGE] : [MAX_MELEE_RANGE]);

// The player points straight up and the pets fan out to either side of it, a fifth of a half turn apart.
const bearing = (index: number): number => -Math.PI / 2 + (index % 2 === 1 ? 1 : -1) * Math.ceil(index / 2) * (Math.PI / 5);

/** The target in the middle, the melee breakpoint and any ranged one as rings around it, and each unit at its distance on its own bearing. */
export function paintField(ctx: CanvasRenderingContext2D, width: number, height: number, palette: Palette, scene: FieldScene) {
	const { em } = palette;
	const cx = width / 2;
	const cy = height / 2;
	const radius = Math.max(0, Math.min(width, height) / 2 - 3 * em);
	const scale = radius / scene.run.reach;
	const melee = MAX_MELEE_RANGE * scale;
	const ranged = MIN_RANGED_RANGE * scale;

	const labels = [scene.meleeLabel, scene.rangedLabel];
	const rings = breakpoints(scene.run).map((yards, index): [number, string] => [yards * scale, labels[index]]);

	ctx.font = palette.font;
	ctx.lineWidth = em / 10;

	if (scene.run.deadZone) {
		ctx.save();
		ctx.globalAlpha = palette.deadZoneAlpha;
		ctx.fillStyle = palette.deadZone;
		ctx.beginPath();
		ctx.arc(cx, cy, ranged, 0, 2 * Math.PI);
		ctx.arc(cx, cy, melee, 0, 2 * Math.PI, true);
		ctx.fill();
		ctx.restore();
	}

	ctx.save();
	ctx.setLineDash([em / 2, em / 3]);
	ctx.strokeStyle = palette.ring;
	for (const [ring] of rings) {
		ctx.beginPath();
		ctx.arc(cx, cy, ring, 0, 2 * Math.PI);
		ctx.stroke();
	}
	ctx.restore();

	ctx.fillStyle = palette.muted;
	ctx.textAlign = 'left';
	ctx.textBaseline = 'top';
	for (const [ring, label] of rings) {
		const offset = (ring + em / 4) * Math.SQRT1_2;
		ctx.fillText(label, cx + offset, cy + offset);
	}

	ctx.fillStyle = palette.target;
	ctx.beginPath();
	ctx.arc(cx, cy, 0.8 * em, 0, 2 * Math.PI);
	ctx.fill();
	ctx.fillStyle = palette.text;
	ctx.textAlign = 'center';
	ctx.fillText(scene.targetLabel, cx, cy + 1.2 * em);

	scene.run.tracks.forEach((track, index) => {
		const angle = bearing(index);
		const distance = scene.yards[index] * scale;
		const x = cx + Math.cos(angle) * distance;
		const y = cy + Math.sin(angle) * distance;
		const dot = (track.isPet ? 0.45 : 0.6) * em;

		ctx.save();
		ctx.globalAlpha = palette.tetherAlpha;
		ctx.strokeStyle = palette.muted;
		ctx.beginPath();
		ctx.moveTo(cx, cy);
		ctx.lineTo(x, y);
		ctx.stroke();
		ctx.restore();

		ctx.fillStyle = track.isPet ? palette.pet : palette.player;
		ctx.beginPath();
		ctx.arc(x, y, dot, 0, 2 * Math.PI);
		ctx.fill();

		ctx.fillStyle = palette.text;
		ctx.textAlign = 'left';
		ctx.textBaseline = 'middle';
		ctx.fillText(scene.unitLabels[index], x + dot + em / 3, y);

		if (index === 0) paintPlayerCasts(ctx, palette, scene, x, y, dot);
	});
}

// Under the player's circle a bar for each cast under way with its label below, and to the circle's left the icons of
// the casts that just went off, the latest nearest.
function paintPlayerCasts(ctx: CanvasRenderingContext2D, palette: Palette, scene: FieldScene, x: number, y: number, dot: number) {
	const { em } = palette;
	const barWidth = 6 * em;
	const barHeight = 0.4 * em;
	const left = x - barWidth / 2;
	let top = y + dot + em / 3;
	ctx.save();
	ctx.textAlign = 'center';
	ctx.textBaseline = 'top';
	for (const { done, label } of scene.casting) {
		// A solid track and a label outlined in the background colour keep the bar readable over the rings and tether.
		ctx.fillStyle = palette.background;
		ctx.fillRect(left, top, barWidth, barHeight);
		ctx.fillStyle = palette.cast;
		ctx.fillRect(left, top, barWidth * Math.min(1, Math.max(0, done)), barHeight);
		ctx.lineWidth = em / 10;
		ctx.strokeStyle = palette.muted;
		ctx.strokeRect(left, top, barWidth, barHeight);
		ctx.lineWidth = em / 4;
		ctx.strokeStyle = palette.background;
		ctx.strokeText(label, x, top + barHeight + em / 6);
		ctx.fillStyle = palette.text;
		ctx.fillText(label, x, top + barHeight + em / 6);
		top += barHeight + 1.5 * em;
	}
	ctx.restore();

	const size = 1.6 * em;
	const gap = em / 4;
	scene.goneOff.forEach(({ icon, alpha }, index) => {
		ctx.save();
		ctx.globalAlpha = alpha;
		ctx.drawImage(icon, x - dot - em / 2 - (index + 1) * size - index * gap, y - size / 2, size, size);
		ctx.restore();
	});
}

/** Each track's distance over the whole run, with any dead zone banded and the pull marked. */
export function paintStrip(ctx: CanvasRenderingContext2D, width: number, height: number, palette: Palette, run: MovementRun) {
	const { em } = palette;
	const inset = em / 4;
	const x = (t: number) => ((t - run.start) / (run.end - run.start)) * width;
	const y = (yards: number) => inset + (1 - yards / run.reach) * (height - 2 * inset);
	ctx.lineWidth = em / 10;

	if (run.deadZone) {
		ctx.save();
		ctx.globalAlpha = palette.deadZoneAlpha;
		ctx.fillStyle = palette.deadZone;
		ctx.fillRect(0, y(MIN_RANGED_RANGE), width, y(MAX_MELEE_RANGE) - y(MIN_RANGED_RANGE));
		ctx.restore();
	}

	ctx.save();
	ctx.setLineDash([em / 2, em / 3]);
	ctx.strokeStyle = palette.ring;
	for (const yards of breakpoints(run)) {
		ctx.beginPath();
		ctx.moveTo(0, y(yards));
		ctx.lineTo(width, y(yards));
		ctx.stroke();
	}
	if (run.start < 0) {
		ctx.beginPath();
		ctx.moveTo(x(0), 0);
		ctx.lineTo(x(0), height);
		ctx.stroke();
	}
	ctx.restore();

	// The pets first, so the player's line stays on top.
	[...run.tracks].reverse().forEach(track => {
		ctx.save();
		ctx.strokeStyle = track.isPet ? palette.pet : palette.player;
		if (track.isPet) ctx.globalAlpha = palette.petPathAlpha;
		ctx.lineWidth = (track.isPet ? 1 : 1.5) * (em / 10);
		ctx.beginPath();
		pathOf(track, run.start, run.end).forEach((point, index) => {
			if (index === 0) ctx.moveTo(x(point.time), y(point.yards));
			else ctx.lineTo(x(point.time), y(point.yards));
		});
		ctx.stroke();
		ctx.restore();
	});
}
