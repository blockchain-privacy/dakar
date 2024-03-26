module.exports = {
	root: true,
	env: {
		node: true,
	},
	extends: [
		'plugin:vue/base',
		'plugin:vue/vue3-recommended',
		'eslint:recommended',
		'plugin:vuetify/base',
		'xo',
	],
	rules: {
		'vue/multi-word-component-names': 'off',
		'no-mixed-operators': 'off',
		'vue/component-name-in-template-casing': ['error', 'kebab-case'],
		'no-return-await': 'off',
	},
	// Todo review if this is needed in the future
	globals: {
		defineModel: 'readonly',
	},
};
