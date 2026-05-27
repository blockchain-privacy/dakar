// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {defineStore} from 'pinia';

export const useWorkspaceStore = defineStore('workspace', {
	state: () => ({
		// If the workspace is loaded, this variable is being watched. When the value changes the item is loaded into the workspace.
		workspaceNode: null,
		// If the workspace is loaded, this variable is being watched. Holds all node IDs which can be loaded into the workspace.
		workspaceNodes: new Map(),
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
		addNodeToMap(payload) {
			this.workspaceNodes.set(payload.id, payload);
		},
		// Expects an array
		setWorkspaceNodes(payload) {
			payload.forEach(d => this.workspaceNodes.set(d.id, d));
		},
		removeNodeFromMap(payload) {
			this.workspaceNodes.delete(payload);
		},
		// Expects an array
		removeNodesFromMap(payload) {
			payload.forEach(d => this.workspaceNodes.delete(d));
		},
	},
});
