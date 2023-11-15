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
	},
};
