// Compares two builds of the site, or checks one deployed build, in a real browser.
//
//   node .github/skills/wowsims-dependencies/scripts/compare_builds.mjs --base <checkout> --head <checkout>
//   node .github/skills/wowsims-dependencies/scripts/compare_builds.mjs --smoke https://<preview>/forever/
//
// Each checkout needs `dist` and `storybook-static` built. --specs druid/balance,mage/dps narrows the pages and --out
// names the screenshot folder. Playwright is not a dependency of the repository: install it outside it and point
// PLAYWRIGHT_MODULE at its index.mjs.
import fs from 'node:fs';
import http from 'node:http';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { parseArgs } from 'node:util';

const { values: args } = parseArgs({
	options: { base: { type: 'string' }, head: { type: 'string' }, smoke: { type: 'string' }, specs: { type: 'string' }, out: { type: 'string' } },
});
const { chromium } = await import(process.env.PLAYWRIGHT_MODULE ? pathToFileURL(process.env.PLAYWRIGHT_MODULE).href : 'playwright');

const OUT = args.out ?? fs.mkdtempSync(path.join(os.tmpdir(), 'compare-builds-'));
const SIM_TIMEOUT = 180_000;
const STILL = '*,*::before,*::after{animation:none!important;transition:none!important;caret-color:transparent!important}';
const MIME = {
	'.html': 'text/html',
	'.js': 'text/javascript',
	'.css': 'text/css',
	'.json': 'application/json',
	'.wasm': 'application/wasm',
	'.png': 'image/png',
	'.jpg': 'image/jpeg',
	'.svg': 'image/svg+xml',
	'.woff2': 'font/woff2',
};

const serve = root =>
	new Promise(resolve => {
		const server = http.createServer((request, response) => {
			let file = path.join(root, decodeURIComponent(request.url.split('?')[0]));
			if (fs.existsSync(file) && fs.statSync(file).isDirectory()) file = path.join(file, 'index.html');
			fs.readFile(file, (error, body) => {
				response.writeHead(error ? 404 : 200, { 'content-type': MIME[path.extname(file)] ?? 'application/octet-stream' });
				response.end(error ? undefined : body);
			});
		});
		server.listen(0, () => resolve(`http://localhost:${server.address().port}`));
	});

const specPages = (root, marker) =>
	args.specs?.split(',') ??
	fs.readdirSync(root).flatMap(group => {
		const folder = path.join(root, group);
		if (!fs.statSync(folder).isDirectory() || ['assets', 'bundle'].includes(group)) return [];
		return fs
			.readdirSync(folder)
			.filter(page => fs.existsSync(path.join(folder, page, marker)))
			.map(page => `${group}/${page}`);
	});

const browser = await chromium.launch();

const newPage = async ({ origin, errors }) => {
	const context = await browser.newContext({
		viewport: { width: 1600, height: 1000 },
		reducedMotion: 'reduce',
		timezoneId: 'UTC',
		locale: 'en-US',
	});
	// The tooltips and the analytics are third-party hosts, whose answers differ between two runs.
	await context.route(
		url => url.origin !== origin,
		route => route.abort(),
	);
	const page = await context.newPage();
	page.on('pageerror', error => errors.push(String(error).slice(0, 200)));
	return page;
};

const openSpec = async (page, url) => {
	await page.goto(url, { waitUntil: 'load', timeout: 60_000 });
	await page.addStyleTag({ content: STILL });
	await page.waitForSelector('[data-testid="sim-ui"]', { timeout: 60_000 });
	await page.waitForTimeout(2500);
	// The welcome tour covers the page on a first visit and swallows every click.
	const skip = page.getByText('Skip tour');
	if (await skip.count()) await skip.click();
	await page.waitForTimeout(400);
};

const simulate = async page => {
	const action = page.locator('[data-testid="dps-action"]:visible').first();
	// A gear planner page has no sim to run.
	if (!(await action.count())) return null;
	await action.click({ timeout: 8000 });
	// Partial numbers show while the sim runs; the unit joins the number when it ends.
	const result = page.locator('[data-testid="topline-result-avg"]', { hasText: /[A-Z]{3}$/ }).first();
	await result.waitFor({ timeout: SIM_TIMEOUT });
	return parseFloat((await result.textContent()).replace(/,/g, ''));
};

const eachTab = async (page, visit) => {
	const tabs = page.locator('[data-testid="sim-tabs"] [role="tab"]');
	const count = await tabs.count();
	for (let index = 0; index < count; index++) {
		const name = (await tabs.nth(index).innerText()).trim().replace(/\W+/g, '-');
		if (!name) continue;
		await tabs.nth(index).click({ timeout: 4000 });
		await page.waitForTimeout(900);
		if ((await tabs.nth(index).getAttribute('aria-selected')) !== 'true') throw new Error(`the ${name} tab did not open`);
		await visit(name);
		await page.keyboard.press('Escape');
	}
	await tabs.first().click();
};

const shoot = async (page, file) => {
	fs.mkdirSync(path.dirname(file), { recursive: true });
	await page.screenshot({ path: file });
};

const sameFile = (a, b) => fs.existsSync(a) && fs.existsSync(b) && fs.readFileSync(a).equals(fs.readFileSync(b));

const shootStories = async (origin, ids, folder, settle) => {
	const queue = [...ids];
	await Promise.all(
		Array.from({ length: 6 }, async () => {
			const page = await newPage({ origin, errors: [] });
			await page.addInitScript(() => {
				Date.now = () => 1700000000000;
				Math.random = () => 0.42;
			});
			while (queue.length) {
				const id = queue.shift();
				await page.goto(`${origin}/iframe.html?id=${id}&viewMode=story`, { waitUntil: 'load', timeout: 30_000 });
				await page.addStyleTag({ content: STILL });
				await page.waitForTimeout(settle);
				await shoot(page, path.join(folder, `${id}.png`));
			}
			await page.context().close();
		}),
	);
};

const compareStories = async (base, head) => {
	const origins = { base: await serve(path.join(base, 'storybook-static')), head: await serve(path.join(head, 'storybook-static')) };
	const index = JSON.parse(fs.readFileSync(path.join(head, 'storybook-static/index.json'), 'utf8'));
	const ids = Object.values(index.entries)
		.filter(entry => entry.type === 'story')
		.map(entry => entry.id);
	await shootStories(origins.base, ids, `${OUT}/stories/base`, 800);
	await shootStories(origins.head, ids, `${OUT}/stories/head`, 800);
	const differing = ids.filter(id => !sameFile(`${OUT}/stories/base/${id}.png`, `${OUT}/stories/head/${id}.png`));
	// A story that differs from its own base on a second shot is unstable, not changed.
	await shootStories(origins.base, differing, `${OUT}/stories/base2`, 3000);
	await shootStories(origins.head, differing, `${OUT}/stories/head2`, 3000);
	const unstable = differing.filter(id => !sameFile(`${OUT}/stories/base/${id}.png`, `${OUT}/stories/base2/${id}.png`));
	const changed = differing.filter(id => !unstable.includes(id) && !sameFile(`${OUT}/stories/base2/${id}.png`, `${OUT}/stories/head2/${id}.png`));
	console.log(`stories: ${ids.length}, ${ids.length - differing.length} identical, ${unstable.length} unstable on base, ${changed.length} changed`);
	for (const id of [...changed, ...unstable]) console.log(`  ${changed.includes(id) ? 'changed ' : 'unstable'} ${id}`);
	return changed.length;
};

const comparePages = async (base, head) => {
	const origins = { base: await serve(path.join(base, 'dist')), head: await serve(path.join(head, 'dist')) };
	let changed = 0;
	for (const spec of specPages(path.join(head, 'dist/forever'), 'index.html')) {
		const tag = spec.replace('/', '_');
		const shots = [];
		const dps = {};
		for (const side of ['base', 'head']) {
			const errors = [];
			const page = await newPage({ origin: origins[side], errors });
			await openSpec(page, `${origins[side]}/forever/${spec}/`);
			await eachTab(page, async name => {
				if (side === 'base') shots.push(name);
				await shoot(page, `${OUT}/pages/${side}/${tag}__${name}.png`);
			});
			dps[side] = await simulate(page);
			if (side === 'base' && dps.base !== null) {
				// A sim seeds itself at random: a second base run is the size of the noise the head run is read against.
				await openSpec(page, `${origins[side]}/forever/${spec}/`);
				dps.base2 = await simulate(page);
			}
			await page.context().close();
			if (errors.length) console.log(`  ${side} ${spec} page error: ${errors[0]}`);
			changed += errors.length;
		}
		const different = shots.filter(name => !sameFile(`${OUT}/pages/base/${tag}__${name}.png`, `${OUT}/pages/head/${tag}__${name}.png`));
		changed += different.length;
		const percent = (a, b) => `${((Math.abs(a - b) / a) * 100).toFixed(2)}%`;
		const sim =
			dps.base === null
				? 'planner'
				: `base ${dps.base} ${dps.base2} head ${dps.head} noise ${percent(dps.base, dps.base2)} delta ${percent(dps.base, dps.head)}`;
		console.log(`${spec}: ${shots.length} tabs, ${different.length} differ${different.length ? ` (${different.join(', ')})` : ''}, ${sim}`);
	}
	return changed;
};

// A link carries the settings through pako, and a bad one fails on the page that opens it.
const smokeLink = async (page, url, errors) => {
	await page.getByText('Export', { exact: true }).first().click();
	await page.getByRole('menuitem', { name: /link/i }).first().click({ timeout: 4000 });
	const link = await page.locator('textarea:visible').first().inputValue({ timeout: 8000 });
	if (!link.startsWith(url)) throw new Error(`the link is ${link.slice(0, 80)}`);
	const opened = await newPage({ origin: new URL(url).origin, errors });
	await openSpec(opened, link);
	await opened.context().close();
	await page.keyboard.press('Escape');
	return 'link';
};

// The generated `create` reads `this`, so adding a row fails when a list holds it unbound.
const smokeRotation = async page => {
	await page.locator('[data-testid="sim-tabs"] [role="tab"]', { hasText: 'Rotation' }).click();
	// Some specs open on the Simple rotation, and its picker offers the action list.
	const type = page.locator('#rotation-tab-rotation-type');
	if ((await type.innerText()).trim() !== 'APL') {
		await type.click();
		await page.getByRole('option', { name: 'APL', exact: true }).click({ timeout: 4000 });
		await page.waitForTimeout(600);
	}
	const pane = page.locator('[data-testid="tab-pane"]:visible').first();
	const size = () => pane.evaluate(element => element.querySelectorAll('*').length);
	const before = await size();
	await page.getByRole('button', { name: 'New Action', exact: true }).last().click();
	await page.waitForTimeout(300);
	if ((await size()) <= before) throw new Error('the add button added no action');
	return 'rotation';
};

const smoke = async root => {
	const url = root.endsWith('/') ? root : `${root}/`;
	let failed = 0;
	for (const spec of specPages('ui/specs', 'spec.ts')) {
		const errors = [];
		const page = await newPage({ origin: new URL(url).origin, errors });
		const steps = [];
		try {
			await openSpec(page, `${url}${spec}/`);
			await eachTab(page, async name => steps.push(name));
			const dps = await simulate(page);
			steps.push(dps === null ? 'planner' : `sim ${dps}`);
			steps.push(await smokeLink(page, `${url}${spec}/`, errors));
			if (dps !== null) steps.push(await smokeRotation(page));
		} catch (error) {
			errors.push(String(error).slice(0, 200));
		}
		await page.context().close();
		failed += errors.length;
		console.log(`${errors.length ? 'FAIL' : 'ok  '} ${spec}: ${steps.join(', ')}${errors.length ? ` -- ${errors[0]}` : ''}`);
	}
	return failed;
};

const problems = args.smoke ? await smoke(args.smoke) : (await compareStories(args.base, args.head)) + (await comparePages(args.base, args.head));
await browser.close();
console.log(`screenshots in ${OUT}`);
process.exit(problems ? 1 : 0);
