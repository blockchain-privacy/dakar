<template>
  <v-dialog v-model="show" max-width="400px">
    <v-card class="mx-auto elevation-4">
      <v-card-title>
        <span class="text-h5">Delete Cluster</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Are you sure you want to delete this cluster?
          It is attached to <strong>{{ numAddresses }}</strong> addresses.
        </div>
        <v-row class="mt-4">
          <v-col class="d-flex justify-end align-center">
            <v-btn text :disabled="isLoading" class="mr-2" @click="show = false">
              Cancel
            </v-btn>
            <v-btn text :loading="isLoading" @click="deleteCluster">
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
import { ROUTE_DELETE_CLUSTER } from '../../constants';

export default {
  name: 'DeleteCluster',
  props: {
    value: { type: Boolean, required: true },
    clusterUid: { type: String, required: true },
    numAddresses: { type: Number, required: true },
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
    deleteCluster() {
      if (this.clusterUid === '' || this.numAddresses <= 0) {
        this.setPersistentErrorMessage('could not delete cluster');
        this.show = false;
        return;
      }

      this.isLoading = true;
      doGet(ROUTE_DELETE_CLUSTER, this.$router, this.$store, this.clusterUid)
        .then((d) => {
          if (d.success === undefined || (!d.success && d.msg === undefined)) throw new Error('error deleting cluster');
          if (!d.success && d.msg !== undefined) throw new Error(d.msg);
          this.$emit('deleted', this.clusterUid);
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
