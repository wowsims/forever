import { Chart, registerables } from 'chart.js';

import { applyThemeFontDefaults } from './chart_fonts';

Chart.register(...registerables);
Chart.defaults.color = 'white';
applyThemeFontDefaults();

export { Bar, Chart, Line } from 'react-chartjs-2';
