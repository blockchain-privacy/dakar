<template>
  <v-dialog
    v-model="model"
    max-width="500px"
    transition="fade-transition"
  >
    <v-card class="mx-auto">
      <v-card-title>
        {{ title }}
      </v-card-title>
      <v-card-text>
        <alert :text="errorMsg" />
        <alert
          :text="infoMsg"
          type="info"
        />
        <v-btn
          variant="outlined"
          class="my-2"
          @click="handleApplyDeviceAuthPreset"
        >
          Apply Device Auth Preset
        </v-btn>
        <v-btn
          variant="outlined"
          class="my-2"
          @click="handleApplyAuthCodeFlowPreset"
        >
          Apply Auth Code Preset
        </v-btn>
        <div
          v-if="isEdit && client.client_id"
          class="text-body-small my-2 text-center"
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
          :loading="isLoading"
          @click="model = false"
        >
          Cancel
        </v-btn>
        <v-btn
          :loading="isLoading"
          @click="createClient"
        >
          {{ submitButtonTitle }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script setup>
import {
	computed,
	inject,
	onMounted,
	onUpdated,
	ref,
	toRaw,
} from 'vue';
import Alert from '@/components/common/Alert.vue';

const model = defineModel({type: Boolean});
const kratosAdmin = inject('kratosadmin');
const emit = defineEmits(['created']);
const isLoading = ref(false);
const errorMsg = ref('');
const infoMsg = ref('');

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
const scopeModel = ref(['offline_access', 'offline', 'openid', 'dakar']);
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

function getParams() {
	if (!clientDetails.value.clientName || clientDetails.value.scope.length === 0
		|| clientDetails.value.responseTypes.length === 0 || clientDetails.value.grantTypes.length === 0) {
		return undefined;
	}

	const clone = structuredClone(toRaw(clientDetails.value));
	clone.clientName = clone.clientName.trim();
	// Scope must be separated by space
	clone.scope = clone.scope.join(' ');
	clone.redirectURIs = clone.redirectURIs.replaceAll(' ', '').split(',').filter(Boolean);

	return clone;
}

async function createClient() {
	errorMsg.value = '';

	const params = getParams();
	if (!params) {
		errorMsg.value = 'invalid parameters';
		return;
	}

	isLoading.value = true;
	infoMsg.value = '';

	let response;
	try {
		if (props.isEdit) {
			params.clientID = props.client.client_id;
			response = await kratosAdmin.oauth.clientsPut({client: params});
		} else {
			response = await kratosAdmin.oauth.clientsPost({client: params});
		}

		if (response.msg) {
			infoMsg.value = response.msg;
			isLoading.value = false;
			return;
		}

		emit('created');
		model.value = false;
	} catch (error) {
		errorMsg.value = error.message;
	}

	isLoading.value = false;
}

function handleApplyDeviceAuthPreset() {
	clientDetails.value.scope = ['openid', 'offline_access'];
	clientDetails.value.grantTypes = [
		'authorization_code',
		'refresh_token',
		'urn:ietf:params:oauth:grant-type:device_code',
	];
	clientDetails.value.responseTypes = ['code', 'id_token'];
	clientDetails.value.tokenEndpointAuthMethod = 'none';
	clientDetails.value.skipConsent = true;
}

function handleApplyAuthCodeFlowPreset() {
	clientDetails.value.scope = ['offline_access'];
	clientDetails.value.grantTypes = ['authorization_code', 'refresh_token'];
	clientDetails.value.responseTypes = ['code', 'id_token'];
	clientDetails.value.tokenEndpointAuthMethod = 'none';
	clientDetails.value.skipConsent = true;
}

</script>

<style scoped>

</style>
