<template>
  <v-dialog
    v-model="model"
    max-width="500px"
    transition="fade-transition"
  >
    <v-card class="mx-auto">
      <v-card-title class="text-h5">
        Create OAuth 2.0 Client
      </v-card-title>
      <v-card-text>
        <v-text-field
          v-model="clientDetails.client_name"
          label="Name"
        />
        <v-text-field
          v-model="clientDetails.scope"
          label="Scope"
          hint="Separate multiple scopes by comma"
        />
        <v-text-field
          v-model="clientDetails.grant_types"
          label="Grant Types"
          hint="Separate multiple types by comma"
        />
        <v-text-field
          v-model="clientDetails.redirect_uris"
          label="Redirect URIs"
          hint="Separate multiple URIs by comma"
        />
        <v-text-field
          v-model="clientDetails.response_types"
          label="Response Types"
          hint="Separate multiple types by comma"
        />
        <v-text-field
          v-model="clientDetails.token_endpoint_auth_method"
          label="Token Endpoint Auth Method"
        />
        <v-checkbox
          v-model="clientDetails.skip_consent"
          label="Skip Consent"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          color="red"
          @click="model = false"
        >
          Cancel
        </v-btn>
        <v-btn @click="createClient">
          Create
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script setup>
import {inject, ref, toRaw} from 'vue';
import {useMsgStore} from '@/pinia/msg.js';
import {useRoute} from 'vue-router';

const route = useRoute();
const msgStore = useMsgStore();
const model = defineModel({type: Boolean});
const kratosAdmin = inject('kratosadmin');
const emit = defineEmits(['created']);
const isLoading = ref(false);
const errorMsg = ref('');

const clientDetails = ref({
	// eslint-disable-next-line camelcase
	client_name: '',
	scope: '',
	// eslint-disable-next-line camelcase
	grant_types: '',
	// eslint-disable-next-line camelcase
	redirect_uris: '',
	// eslint-disable-next-line camelcase
	response_types: '',
	// eslint-disable-next-line camelcase
	token_endpoint_auth_method: '',
	// eslint-disable-next-line camelcase
	skip_consent: false,
});

// Functions

function setInfoMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'info', temporary: true, category: route.name,
	});
}

function getParams() {
	if (!clientDetails.value.client_name || !clientDetails.value.scope
		|| !clientDetails.value.response_types || !clientDetails.value.grant_types) {
		return undefined;
	}

	const clone = structuredClone(toRaw(clientDetails.value));
	// eslint-disable-next-line camelcase
	clone.client_name = clone.client_name.trim();
	// Scope must be separated by space: remove all spaces, split by comma, remove empty items and join by space
	clone.scope = clone.scope.replaceAll(' ', '').split(',').filter(d => d).join(' ');
	// eslint-disable-next-line camelcase
	clone.grant_types = clone.grant_types.replaceAll(' ', '').split(',').filter(d => d);
	// eslint-disable-next-line camelcase
	clone.redirect_uris = clone.redirect_uris.replaceAll(' ', '').split(',').filter(d => d);
	// eslint-disable-next-line camelcase
	clone.response_types = clone.response_types.replaceAll(' ', '').split(',').filter(d => d);

	return clone;
}

async function createClient() {
	const params = getParams();
	if (!params) {
		return;
	}

	console.log(params);

	isLoading.value = true;
	try {
		const response = await kratosAdmin.oauth.clientsPost({client: params});
		if (response.msg) {
			setInfoMessage(response.msg);
		}

		emit('created');
	} catch (e) {
		errorMsg.value = e.message;
	}

	isLoading.value = false;
	model.value = false;
}

</script>

<style scoped>

</style>
