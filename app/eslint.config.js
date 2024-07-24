import pluginVue from 'eslint-plugin-vue';
import pluginVuetify from 'eslint-plugin-vuetify';
import {FlatCompat} from '@eslint/eslintrc';
import globals from 'globals';
import eslintJsPlugin from '@eslint/js';

const compat = new FlatCompat();

export default [
	{
		rules: eslintJsPlugin.configs.recommended.rules,
	},
	...compat.extends('xo'),
	...pluginVue.configs['flat/base'],
	...pluginVue.configs['flat/recommended'],
	{
		files: ['**/*.js', '**/*.vue'],
		ignores: ['**/*.gitignore'],
		plugins: {
			pluginVuetify,
		},
		rules: {
			'vue/multi-word-component-names': 'off',
			'no-mixed-operators': 'off',
			'vue/component-name-in-template-casing': ['error', 'kebab-case'],
			'no-return-await': 'off',
			'vue/valid-v-slot': ['error', {allowModifiers: true}],
		},
		languageOptions: {
			globals: {
				...globals.browser,
			},
		},
	},
];
