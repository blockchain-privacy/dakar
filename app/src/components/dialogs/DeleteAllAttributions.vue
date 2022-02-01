<template>
  <v-dialog v-model="show" max-width="400px">
    <v-card class="mx-auto elevation-4">
      <v-card-title>
        <span class="text-h5">Delete All Attributions</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Are you sure you want to delete all attributions?
        </div>
        <v-row class="mt-4">
          <v-col class="d-flex justify-end align-center">
            <v-btn text :disabled="isLoading" class="mr-2" @click="show = false">
              Cancel
            </v-btn>
            <v-btn text color="red" :loading="isLoading" @click="deleteAllClusters">
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
import { ROUTE_DELETE_ALL_PRIVATE_ATTRIBUTIONS } from '../../constants';

export default {
  name: 'DeleteAllAttributions',
  props: {
    value: { type: Boolean, required: true },
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
    deleteAllClusters() {
      this.isLoading = true;
      doGet(ROUTE_DELETE_ALL_PRIVATE_ATTRIBUTIONS, this.$router, this.$store, this.clusterUid)
        .then((d) => {
          if (d.success === undefined || (!d.success && d.msg === undefined)) throw new Error('error deleting attributions');
          if (!d.success && d.msg !== undefined) throw new Error(d.msg);
          this.$emit('deleted');
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
