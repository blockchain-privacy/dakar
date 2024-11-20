import pluginVue from 'eslint-plugin-vue';
import globals from 'globals';
import eslintJsPlugin from '@eslint/js';
import xo from 'eslint-config-xo';

export default [
	{
		rules: eslintJsPlugin.configs.recommended.rules,
	},
	...pluginVue.configs['flat/base'],
	...pluginVue.configs['flat/recommended'],
	...xo,
	{
		files: ['**/*.js', '**/*.vue'],
		ignores: ['**/*.gitignore'],
		rules: {
			'vue/prefer-true-attribute-shorthand': ['error', 'always'],
			'vue/multi-word-component-names': 'off',
			'vue/no-boolean-default': 'error',
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
