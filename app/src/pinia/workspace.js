import {defineStore} from 'pinia';

export const useWorkspaceStore = defineStore('workspace', {
	state: () => ({
		// If the workspace is loaded, this variable is being watched. When the value changes the item is loaded into the workspace.
		workspaceNode: null,
		// Is set to true as soon as the workspace component is mounted and set to false when it is unmounted
		isWorkspaceActive: false,
	}),
	getters: {
		getIsWorkspaceActive: state => state.isWorkspaceActive,
	},
	actions: {
		setWorkspaceNode(payload) {
			this.workspaceNode = payload;
		},
		setWorkspaceActive(active) {
			this.isWorkspaceActive = active;
		},
	},
});
