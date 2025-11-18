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
        <v-btn
          variant="outlined"
          class="my-2"
          @click="handleApplyDeviceAuthPreset"
        >
          Apply Device Auth Preset
        </v-btn>
        <v-text-field
          v-model="clientDetails.client_name"
          label="Name"
        />
        <v-select
          v-model="clientDetails.scope"
          multiple
          chips
          label="Scope"
          :items="scopeModel"
        />
        <v-select
          v-model="clientDetails.grant_types"
          multiple
          chips
          label="Grant Types"
          :items="grantTypesModel"
        />
        <v-text-field
          v-model="clientDetails.redirect_uris"
          label="Redirect URIs"
          hint="Separate multiple URIs by comma"
        />
        <v-select
          v-model="clientDetails.response_types"
          multiple
          chips
          label="Response Types"
          :items="responseTypesModel"
        />
        <v-select
          v-model="clientDetails.token_endpoint_auth_method"
          label="Token Endpoint Auth Method"
          :items="tokenEndPointAuthModel"
        />
        <v-checkbox
          v-model="clientDetails.skip_consent"
          label="Skip Consent"
          hide-details
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

const tokenEndPointAuthModel = ref(['client_secret_post', 'client_secret_basic', 'none']);
const responseTypesModel = ref(['code', 'id_token', 'token']);
const grantTypesModel = ref(['authorization_code', 'implicit', 'client_credentials', 'refresh_token', 'urn:ietf:params:oauth:grant-type:device_code']);
const scopeModel = ref(['offline_access', 'offline', 'openid']);
const clientDetails = ref({
	// eslint-disable-next-line camelcase
	client_name: '',
	scope: [],
	// eslint-disable-next-line camelcase
	grant_types: [],
	// eslint-disable-next-line camelcase
	redirect_uris: '',
	// eslint-disable-next-line camelcase
	response_types: [],
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
	// Scope must be separated by space
	clone.scope = clone.scope.join(' ');
	// eslint-disable-next-line camelcase
	clone.redirect_uris = clone.redirect_uris.replaceAll(' ', '').split(',').filter(d => d);

	return clone;
}

async function createClient() {
	const params = getParams();
	if (!params) {
		return;
	}

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

function handleApplyDeviceAuthPreset() {
	clientDetails.value.scope = ['openid'];
	// eslint-disable-next-line camelcase
	clientDetails.value.grant_types = ['authorization_code', 'refresh_token', 'urn:ietf:params:oauth:grant-type:device_code'];
	// eslint-disable-next-line camelcase
	clientDetails.value.response_types = ['code', 'id_token'];
	// eslint-disable-next-line camelcase
	clientDetails.value.token_endpoint_auth_method = 'none';
	// eslint-disable-next-line camelcase
	clientDetails.value.skip_consent = true;
}

</script>

<style scoped>

</style>
