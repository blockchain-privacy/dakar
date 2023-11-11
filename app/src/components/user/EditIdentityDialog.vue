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

<script>
import {emailRules, handleError} from '@/utilities';

export default {
	name: 'EditIdentityDialog',
	props: {
		modelValue: {type: Boolean, required: true},
		identity: {type: Object, required: true},
		createNewUser: {type: Boolean, required: true},
	},
	emits: ['update:modelValue', 'saved'],
	data() {
		return {
			isLoading: false,
			shadowIdentity: {
				id: '',
				email: '',
				roles: [],
				state: '',
			},
			roles: ['admin', 'user', 'privileged'],
			states: ['active', 'inactive'],
			rules: {
				roleRules: [
					v => v.length > 0 || 'At least one role is required',
				],
				stateRules: [
					v => v.length > 0 || 'State must be set',
				],
				emailRules,
			},
		};
	},
	computed: {
		formTitle() {
			return this.createNewUser ? 'Create Identity' : 'Edit Identity';
		},
		show: {
			get() {
				return this.modelValue;
			},
			set(value) {
				this.$emit('update:modelValue', value);
			},
		},
	},
	mounted() {
		this.shadowIdentity = this.identity;
	},
	methods: {
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
		setInfoMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'info', temporary: true, category: this.$route.name});
		},
		async saveIdentity() {
			const {valid} = await this.$refs.modifyIdentityForm.validate();

			if (!valid) {
				return;
			}

			this.isLoading = true;
			if (this.createNewUser) {
				try {
					const response = await this.dakar.authentication.createIdentityPost({
						identity: {
							email: this.shadowIdentity.email,
							roles: this.shadowIdentity.roles,
							state: this.shadowIdentity.state,
						},
					});
					if (response.msg) {
						this.setInfoMessage(response.msg);
					}

					this.$emit('saved');
				} catch (e) {
					this.setErrorMessage(e);
				}
			} else {
				try {
					const response = await this.dakar.authentication.modifyIdentityPost({
						identity: {
							uid: this.shadowIdentity.id,
							email: this.shadowIdentity.email,
							state: this.shadowIdentity.state,
							roles: this.shadowIdentity.roles.map(d => ({name: d})),
						},
					});

					if (response.msg) {
						this.setInfoMessage(response.msg);
					}

					this.$emit('saved');
				} catch (e) {
					handleError(this, e);
				}
			}

			this.isLoading = false;
			this.show = false;
		},
	},
};
</script>

<style scoped>

</style>
