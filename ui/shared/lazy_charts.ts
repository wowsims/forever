import { lazy } from 'react';

// chart.js and its setup load with the first chart drawn, not with the page; each one renders inside a Suspense.
const charts = () => import('./charts');

export const Bar = lazy(() => charts().then(module => ({ default: module.Bar })));

export const Chart = lazy(() => charts().then(module => ({ default: module.Chart })));

export const Line = lazy(() => charts().then(module => ({ default: module.Line })));
