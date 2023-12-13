<template>
  <v-dialog
    v-model="show"
    max-width="500px"
    transition="fade-transition"
  >
    <v-card>
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
                label="E-mail"
                type="email"
                :rules="rules.emailRules"
                style="min-width: 250px"
                :autofocus="true"
              />
              <v-select
                v-model="shadowIdentity.roles"
                :rules="rules.roleRules"
                :items="roles"
                label="Roles"
                :multiple="true"
              />
              <v-select
                v-model="shadowIdentity.state"
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
          @click="show = false"
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
import {computed, inject, onMounted, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const route = useRoute();
const dakar = inject('dakar');
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

const emit = defineEmits(['update:modelValue', 'saved']);

const props = defineProps({
	modelValue: {type: Boolean, required: true},
	identity: {type: Object, required: true},
	createNewUser: {type: Boolean, required: true},
});

const isLoading = ref(false);
const shadowIdentity = ref({id: '', email: '', roles: [], state: ''});
// Template ref
const modifyIdentityForm = ref(null);

const roles = ['admin', 'user', 'privileged'];
const states = ['active', 'inactive'];
const rules = {
	roleRules: [
		v => v.length > 0 || 'At least one role is required',
	],
	stateRules: [
		v => v.length > 0 || 'State must be set',
	],
	emailRules,
};

// Computed
const formTitle = computed(() => props.createNewUser ? 'Create Identity' : 'Edit Identity');

const show = computed({
	get() {
		return props.modelValue;
	},
	set(value) {
		emit('update:modelValue', value);
	},
});

onMounted(() => {
	shadowIdentity.value = props.identity;
});

function setErrorMessage(msg) {
	msgStore.addMessage({text: msg, type: 'error', temporary: true, category: route.name});
}

function setInfoMessage(msg) {
	msgStore.addMessage({text: msg, type: 'info', temporary: true, category: route.name});
}

async function saveIdentity() {
	const {valid} = await modifyIdentityForm.value.validate();

	if (!valid) {
		return;
	}

	isLoading.value = true;
	if (props.createNewUser) {
		try {
			const response = await dakar.authentication.createIdentityPost({
				identity: {
					email: shadowIdentity.value.email,
					roles: shadowIdentity.value.roles,
					state: shadowIdentity.value.state,
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
			const response = await dakar.authentication.modifyIdentityPost({
				identity: {
					uid: shadowIdentity.value.id,
					email: shadowIdentity.value.email,
					state: shadowIdentity.value.state,
					roles: shadowIdentity.value.roles,
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
	show.value = false;
}

</script>

<style scoped>

</style>
