import {defineStore} from 'pinia';

export const useWorkspaceStore = defineStore('workspace', {
	state: () => ({
		// If the workspace is loaded, this variable is being watched. When the value changes the item is loaded into the workspace.
		workspaceNode: null,
		// If the workspace is loaded, this variable is being watched. Holds all node IDs which can be loaded into the workspace.
		workspaceNodes: new Set(),
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
		addNodeToSet(payload) {
			this.workspaceNodes.add(payload);
		},
		removeNodeFromSet(payload) {
			this.workspaceNodes.delete(payload);
		},
	},
});
