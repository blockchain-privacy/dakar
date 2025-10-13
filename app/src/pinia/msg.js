// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {defineStore} from 'pinia';

let msgCounter = 1;
export const useMsgStore = defineStore('msg', {
	state: () => ({
		messages: new Map(),
	}),
	getters: {
		getMessages: state => state.messages,
	},
	actions: {
		addMessage(payload) {
			if (!payload.text || payload.text.toString().trim() === '') {
				return;
			}

			// Convert potential error object to string
			payload.text = payload.text.toString();

			this.messages.set(msgCounter, payload);
			msgCounter += 1;
		},
		resetMessages() {
			this.messages.clear();
		},
		removeMessage(msgKey) {
			this.messages.delete(msgKey);
		},
		filterMessages(category) {
			for (const [key, value] of this.messages) {
				if (value.category === category) {
					this.messages.delete(key);
				}
			}
		},
	},
});
