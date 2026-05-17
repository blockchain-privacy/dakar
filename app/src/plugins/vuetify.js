// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

// eslint-disable-next-line import-x/no-unassigned-import
import 'vuetify/styles';
import {createVuetify} from 'vuetify';
import {aliases, mdi} from 'vuetify/iconsets/mdi-svg';
// Custom icons
import {md3} from 'vuetify/blueprints';
import graph from '../assets/graph.svg';

const darkTheme = {
	dark: true,
	colors: {
		primary: '#008EE5',
		'on-primary': '#fff',
		secondary: '#8592A3',
		'on-secondary': '#fff',
		success: '#71DD37',
		'on-success': '#fff',
		info: '#03C3EC',
		'on-info': '#fff',
		warning: '#FFAB00',
		'on-warning': '#fff',
		error: '#FF3E1D',
		background: '#232333',
		'on-background': '#DBDBEB',
		surface: '#2B2C40',
		'on-surface': '#DBDBEB',
		'grey-50': '#2A2E42',
		'grey-100': '#444463',
		'grey-200': '#4A5072',
		'grey-300': '#5E6692',
		'grey-400': '#7983BB',
		'grey-500': '#8692D0',
		'grey-600': '#AAB3DE',
		'grey-700': '#B6BEE3',
		'grey-800': '#CFD3EC',
		'grey-900': '#E7E9F6',
		'perfect-scrollbar-thumb': '#4A5072',
		'skin-bordered-background': '#2b2c40',
		'skin-bordered-surface': '#2b2c40',
	},
	variables: {
		'code-color': '#d400ff',
		'overlay-scrim-background': '#0D0E24',
		'overlay-scrim-opacity': 0.6,
		'border-color': '#DBDBEB',
		'snackbar-background': '#DBDBEB',
		'snackbar-color': '#2B2C40',
		'tooltip-background': '#464A65',
		'tooltip-opacity': 0.9,
		'table-header-background': '#3A3E5B',

		// Shadows
		'shadow-key-umbra-opacity': 'rgba(20, 21, 33, 0.06)',
		'shadow-key-penumbra-opacity': 'rgba(20, 21, 33, 0.04)',
		'shadow-key-ambient-opacity': 'rgba(20, 21, 33, 0.02)',
	},
};

const lightTheme = {
	dark: false,
	colors: {
		primary: '#008EE5',
	},
};

export default createVuetify({
	blueprint: md3,
	defaults: {
		VAppBar: {
			VBtn: {
				color: undefined,
			},
		},
		VSwitch: {
			color: 'primary',
		},
		VSlider: {
			color: 'primary',
		},
		VRadioGroup: {
			color: 'primary',
		},
		VRangeSlider: {
			color: 'primary',
		},
		VCheckbox: {
			color: 'primary',
		},
		VCard: {
			VCard: {
				VToolbar: {
					color: undefined,
				},
			},
			VForm: {
				VTextField: {
					variant: undefined,
				},
			},
			VSelect: {
				variant: 'outlined',
			},
			VTextField: {
				variant: 'outlined',
			},
			VTextarea: {
				variant: 'outlined',
			},
			VBtn: {
				color: 'primary',
			},
		},
	},
	theme: {
		defaultTheme: 'system',
		variations: {
			colors: ['primary', 'secondary'],
			lighten: 4,
			darken: 4,
		},
		themes: {
			light: lightTheme,
			dark: darkTheme,
		},
	},
	icons: {
		defaultSet: 'mdi',
		aliases: {
			...aliases,
			graphIcon: graph,
		},
		sets: {
			mdi,
		},
	},
});

