<template>
  <v-dialog
    v-model="model"
    max-width="300px"
    transition="fade-transition"
  >
    <v-card class="mx-auto">
      <v-card-title>
        <span class="text-h5">{{ formTitle }}</span>
      </v-card-title>
      <v-card-text>
        <v-container>
          <v-row>
            <v-form
              ref="modifyIdentityForm"
              validate-on="submit"
            >
              <v-text-field
                v-model="shadowIdentity.email"
                class="my-1"
                label="E-mail"
                type="email"
                :rules="rules.emailRules"
                style="min-width: 250px"
                autofocus
              />
              <v-select
                v-model="shadowIdentity.roles"
                class="my-1"
                :rules="rules.roleRules"
                :items="roles"
                label="Roles"
                multiple
              />
              <v-select
                v-model="shadowIdentity.services"
                class="my-1"
                :rules="rules.serviceRules"
                :items="services"
                label="Services"
                multiple
              />
              <v-select
                v-model="shadowIdentity.state"
                class="my-1"
                :rules="rules.stateRules"
                :items="states"
                label="State"
              />
            </v-form>
          </v-row>
        </v-container>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          color="red"
          @click="model = false"
        >
          Cancel
        </v-btn>
        <v-btn @click="saveIdentity">
          Save
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {emailRules, handleError} from '@/utilities';
import {
	computed, inject, onMounted, ref,
} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const route = useRoute();
const kratosAdmin = inject('kratosadmin');
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

const emit = defineEmits(['saved']);
const model = defineModel({type: Boolean});
const props = defineProps({
	identity: {type: Object, required: true},
	createNewUser: {type: Boolean, required: true},
});

const isLoading = ref(false);
const shadowIdentity = ref({
	id: '', email: '', roles: [], state: '', services: [],
});
// Template ref
const modifyIdentityForm = ref(null);

const roles = [{title: 'Admin', value: 'admin'}, {title: 'Privileged', value: 'privileged'}];
const services = [{title: 'Dakar Dash', value: 'dakarDash'}, {title: 'Dakar BTC', value: 'dakarBTC'}];
const states = [{title: 'Active', value: 'active'}, {title: 'Inactive', value: 'inactive'}];
const rules = {
	roleRules: [
		v => v.length > 0 || 'At least one role is required',
	],
	serviceRules: [
		v => v.length > 0 || 'At least one service is required',
	],
	stateRules: [
		v => v.length > 0 || 'State must be set',
	],
	emailRules,
};

// Computed
const formTitle = computed(() => props.createNewUser ? 'Create Identity' : 'Edit Identity');

onMounted(() => {
	shadowIdentity.value = props.identity;
	shadowIdentity.value.services = [];
	if (props.identity.metadata_public?.dakar_dash_user) {
		shadowIdentity.value.services.push('dakarDash');
	}

	if (props.identity.metadata_public?.dakar_btc_user) {
		shadowIdentity.value.services.push('dakarBTC');
	}
});

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

function setInfoMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'info', temporary: true, category: route.name,
	});
}

async function saveIdentity() {
	const {valid} = await modifyIdentityForm.value.validate();

	if (!valid) {
		return;
	}

	isLoading.value = true;
	if (props.createNewUser) {
		try {
			const response = await kratosAdmin.identitiesPost({
				identity: {
					email: shadowIdentity.value.email,
					roles: shadowIdentity.value.roles,
					state: shadowIdentity.value.state,
					services: shadowIdentity.value.services,
				},
			});
			if (response.msg) {
				setInfoMessage(response.msg);
			}

			emit('saved');
		} catch (e) {
			setErrorMessage(e);
		}
	} else {
		try {
			const response = await kratosAdmin.identitiesPut({
				identity: {
					uid: shadowIdentity.value.id,
					email: shadowIdentity.value.email,
					state: shadowIdentity.value.state,
					roles: shadowIdentity.value.roles,
					services: shadowIdentity.value.services,
				},
			});

			if (response.msg) {
				setInfoMessage(response.msg);
			}

			emit('saved');
		} catch (e) {
			handleError(context, e);
		}
	}

	isLoading.value = false;
	model.value = false;
}

</script>

<style scoped>

</style>
