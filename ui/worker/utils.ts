// Copies of shared utils, kept here so the worker stays isolated and chunk building hits no circular dependency.
// eslint-disable-next-line @typescript-eslint/no-empty-function
export const noop = () => {};
export const sleep = (ms: number) => new Promise(r => setTimeout(r, ms));
