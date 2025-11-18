<template>
  <v-dialog
    v-model="model"
    max-width="500px"
    transition="fade-transition"
  >
    <v-card class="mx-auto">
      <v-card-title class="text-h5">
        {{ title }}
      </v-card-title>
      <v-card-text>
        <v-btn
          variant="outlined"
          class="my-2"
          @click="handleApplyDeviceAuthPreset"
        >
          Apply Device Auth Preset
        </v-btn>
        <div
          v-if="isEdit && client.client_id"
          class="text-caption my-2 text-center"
        >
          ID: {{ client.client_id }}
        </div>
        <v-text-field
          v-model="clientDetails.clientName"
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
          v-model="clientDetails.grantTypes"
          multiple
          chips
          label="Grant Types"
          :items="grantTypesModel"
        />
        <v-text-field
          v-model="clientDetails.redirectURIs"
          label="Redirect URIs"
          hint="Separate multiple URIs by comma"
        />
        <v-select
          v-model="clientDetails.responseTypes"
          multiple
          chips
          label="Response Types"
          :items="responseTypesModel"
        />
        <v-select
          v-model="clientDetails.tokenEndpointAuthMethod"
          label="Token Endpoint Auth Method"
          :items="tokenEndPointAuthModel"
        />
        <v-checkbox
          v-model="clientDetails.skipConsent"
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
          {{ submitButtonTitle }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script setup>
import {
	computed, inject, onMounted, onUpdated, ref, toRaw,
} from 'vue';
import {useMsgStore} from '@/pinia/msg.js';
import {useRoute} from 'vue-router';

const route = useRoute();
const msgStore = useMsgStore();
const model = defineModel({type: Boolean});
const kratosAdmin = inject('kratosadmin');
const emit = defineEmits(['created']);
const isLoading = ref(false);
const errorMsg = ref('');

const tokenEndPointAuthModel = ref([
	'client_secret_post',
	'client_secret_basic',
	'none',
]);
const responseTypesModel = ref(['code', 'id_token', 'token']);
const grantTypesModel = ref([
	'authorization_code',
	'implicit',
	'client_credentials',
	'refresh_token',
	'urn:ietf:params:oauth:grant-type:device_code',
]);
const scopeModel = ref(['offline_access', 'offline', 'openid']);
const clientDetails = ref({
	clientName: '',
	scope: [],
	grantTypes: [],
	redirectURIs: '',
	responseTypes: [],
	tokenEndpointAuthMethod: '',
	skipConsent: false,
});

const props = defineProps({
	isEdit: {type: Boolean, required: false},
	client: {
		type: Object, required: false, default() {
			return {};
		},
	},
});

// Computed

const title = computed(() => props.isEdit ? 'Update OAuth 2.0 Client' : 'Create OAuth 2.0 Client');
const submitButtonTitle = computed(() => props.isEdit ? 'Update' : 'Create');

// Hooks

onMounted(() => {
	updateFromProps();
});

onUpdated(() => {
	updateFromProps();
});

// Functions

function updateFromProps() {
	if (!props.isEdit) {
		return;
	}

	clientDetails.value.clientName = props.client.client_name;
	clientDetails.value.scope = props.client.scope.split(' ');
	clientDetails.value.grantTypes = props.client.grant_types;
	clientDetails.value.redirectURIs = props.client.redirect_uris.join(',');
	clientDetails.value.responseTypes = props.client.response_types;
	clientDetails.value.tokenEndpointAuthMethod = props.client.token_endpoint_auth_method;
	clientDetails.value.skipConsent = props.client.skip_consent;
}

function setInfoMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'info', temporary: true, category: route.name,
	});
}

function getParams() {
	if (!clientDetails.value.clientName || !clientDetails.value.scope
		|| !clientDetails.value.responseTypes || !clientDetails.value.grantTypes) {
		return undefined;
	}

	const clone = structuredClone(toRaw(clientDetails.value));
	clone.clientName = clone.clientName.trim();
	// Scope must be separated by space
	clone.scope = clone.scope.join(' ');
	clone.redirectURIs = clone.redirectURIs.replaceAll(' ', '').split(',').filter(d => d);

	return clone;
}

async function createClient() {
	const params = getParams();
	if (!params) {
		return;
	}

	isLoading.value = true;

	if (props.isEdit) {
		params.clientID = props.client.client_id;
		try {
			const response = await kratosAdmin.oauth.clientsPut({client: params});
			if (response.msg) {
				setInfoMessage(response.msg);
			}

			emit('created');
		} catch (e) {
			errorMsg.value = e.message;
		}
	} else {
		try {
			const response = await kratosAdmin.oauth.clientsPost({client: params});
			if (response.msg) {
				setInfoMessage(response.msg);
			}

			emit('created');
		} catch (e) {
			errorMsg.value = e.message;
		}
	}

	isLoading.value = false;
	model.value = false;
}

function handleApplyDeviceAuthPreset() {
	clientDetails.value.scope = ['openid'];
	clientDetails.value.grantTypes = [
		'authorization_code',
		'refresh_token',
		'urn:ietf:params:oauth:grant-type:device_code',
	];
	clientDetails.value.responseTypes = ['code', 'id_token'];
	clientDetails.value.tokenEndpointAuthMethod = 'none';
	clientDetails.value.skipConsent = true;
}

</script>

<style scoped>

</style>
