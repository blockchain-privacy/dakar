<template>
  <v-dialog v-model="show" max-width="400px">
    <v-card class="mx-auto elevation-4">
      <v-card-title>
        <span class="text-h5">Delete Attribution</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Are you sure you want to delete the attribution <code>{{ tag }}</code>?
        </div>
        <v-row class="mt-4">
          <v-col class="d-flex justify-end align-center">
            <v-btn text :disabled="isLoading" class="mr-2" @click="show = false">
              Cancel
            </v-btn>
            <v-btn text :loading="isLoading" color="red" @click="deleteAttribution">
              Delete
            </v-btn>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import { doGet } from '../../utilities';
import { ROUTE_DELETE_ATTRIBUTION } from '../../constants';

export default {
  name: 'DeleteAttribution',
  props: {
    value: { type: Boolean, required: true },
    attributionUid: { type: String, required: true },
    tag: { type: String, required: true },
  },
  data() {
    return {
      isLoading: false,
    };
  },
  computed: {
    show: {
      get() {
        return this.value;
      },
      set(value) {
        this.$emit('input', value);
      },
    },
  },
  methods: {
    setPersistentErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: false });
    },
    deleteAttribution() {
      if (this.attributionUid === '' || this.numAddresses <= 0) {
        this.setPersistentErrorMessage('could not delete attribution');
        this.show = false;
        return;
      }

      this.isLoading = true;
      doGet(ROUTE_DELETE_ATTRIBUTION, this.$router, this.$store, this.attributionUid)
        .then((d) => {
          if (d.success === undefined || (!d.success && d.msg === undefined)) throw new Error('error deleting attribution');
          if (!d.success && d.msg !== undefined) throw new Error(d.msg);
          this.$emit('deleted', this.attributionUid);
        })
        .catch((e) => {
          this.setPersistentErrorMessage(e);
        })
        .finally(() => {
          this.isLoading = false;
          this.show = false;
        });
    },
  },
};
</script>

<style scoped>

</style>
