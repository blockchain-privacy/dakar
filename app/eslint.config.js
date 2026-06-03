import pluginVue from 'eslint-plugin-vue';
import globals from 'globals';
import eslintJsPlugin from '@eslint/js';
import xo from 'eslint-config-xo';

export default [
	...pluginVue.configs['flat/base'],
	...pluginVue.configs['flat/recommended'],
	...xo(),
	{
		files: ['**/*.js', '**/*.vue'],
		ignores: ['**/*.gitignore'],
		rules: {
			'regexp/prefer-named-capture-group': 'off',
			'unicorn/prefer-query-selector': 'off',
			'unicorn/no-this-assignment': 'off',
			'unicorn/prefer-global-this': 'off',
			'unicorn/filename-case': 'off',
			'import-x/no-anonymous-default-export': 'off',
			'n/file-extension-in-import': 'off',
			'import-x/extensions': 'off',
			'unicorn/no-array-for-each': 'off',
			'regexp/sort-character-class-elements': 'off',
			'unicorn/prevent-abbreviations': 'off',
			'vue/prefer-true-attribute-shorthand': ['error', 'always'],
			'vue/multi-word-component-names': 'off',
			'vue/no-boolean-default': 'error',
			'vue/component-name-in-template-casing': ['error', 'kebab-case'],
			'no-return-await': 'off',
			'vue/valid-v-slot': ['error', {allowModifiers: true}],
			...eslintJsPlugin.configs.recommended.rules,
		},
		languageOptions: {
			globals: {
				...globals.browser,
			},
			ecmaVersion: 2025,
		},
		linterOptions: {
			reportUnusedInlineConfigs: 'error',
			reportUnusedDisableDirectives: 'error',
		},
	},
];
