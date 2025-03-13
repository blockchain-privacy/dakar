<template>
  <v-dialog
    v-model="model"
    max-width="600"
  >
    <v-card
      title="Add entities"
      :prepend-icon="mdiPlus"
    >
      <v-form
        id="queryForm"
        ref="queryForm"
        validate-on="submit"
      >
        <v-card-text>
          <p class="text-subtitle-1">
            Add one or multiple entities. Separate multiple entities by any special character.
          </p>
          <v-text-field
            v-model="graphQuery"
            class="mt-4"
            autofocus
            variant="outlined"
            density="compact"
            color="primary"
            :rules="inputRules"
            label="Add a transactions or address clusters"
            :disabled="!addEntityEnabled"
            :append-inner-icon="mdiMagnify"
            @click:append-inner="onAddEntities"
            @keydown.enter="onAddEntities"
          />
          <v-expand-transition>
            <div
              v-if="queryItemCount > 1"
              class="d-flex justify-center"
            >
              <v-btn
                variant="text"
                @click="showDetectedEntities = !showDetectedEntities"
              >
                {{ showDetectedEntities?'Hide':'Show' }} detected entities
              </v-btn>
            </div>
          </v-expand-transition>
          <v-expand-transition>
            <div v-if="queryItemCount > 1 && showDetectedEntities">
              <v-list
                v-for="entity in detectedEntities"
                :key="entity"
                density="compact"
              >
                <v-list-item class="ma-0 pa-0">
                  {{ entity }}
                </v-list-item>
              </v-list>
            </div>
          </v-expand-transition>
        </v-card-text>
        <v-card-actions>
          <v-btn
            class="ml-auto"
            text="Cancel"
            @click="model = false"
          />
          <v-btn
            :disabled="queryItemCount === 0"
            @click="onAddEntities"
          >
            Add {{ queryItemCount > 1?queryItemCount:'' }} {{ pluralIrregular('entity','entities', queryItemCount) }}
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {mdiMagnify, mdiPlus} from '@mdi/js';
import {extractEntities, pluralIrregular} from '@/utilities/index.js';
import {ref, computed} from 'vue';

const model = defineModel({type: Boolean});
const emit = defineEmits(['addEntities']);

defineProps({
	addEntityEnabled: {type: Boolean, required: true},
});

const graphQuery = ref('');
const queryForm = ref(null);
const showDetectedEntities = ref(false);

const inputRules = [
	q => extractEntities(q).length > 0 || 'query contains no valid entities',
];

// Computed
const detectedEntities = computed(() => extractEntities(graphQuery.value));

const queryItemCount = computed(() => detectedEntities.value.length);

// Functions
async function onAddEntities() {
	const {valid} = await queryForm.value.validate();
	if (!valid) {
		return;
	}

	model.value = false;
	emit('addEntities', extractEntities(graphQuery.value));
	graphQuery.value = '';
}
</script>

<style scoped>

</style>
