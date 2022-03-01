<template>
  <div>
    <v-card class="mx-auto" elevation="4" max-width="1200">
      <v-toolbar dark flat color="primary" class="mb-1">
        <v-toolbar-title>
          <v-icon>{{ icon.mdiPlaylistRemove }}</v-icon>
          Address Exclusions
        </v-toolbar-title>
        <v-spacer></v-spacer>
        <v-menu bottom left>
          <template v-slot:activator="{ on, attrs }">
            <v-btn dark icon v-bind="attrs" v-on="on">
              <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="addAddressExclusions = true">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiFileImport }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>Import address exclusions</v-list-item-title>
            </v-list-item>
            <v-list-item :disabled="items.length === 0" @click="deleteAllExclusionsDialog = true">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiDelete }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>Delete all address exclusions</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </v-toolbar>
      <v-card-text v-if="items.length > 0">
        <div class="d-flex">
          <v-icon>{{ icon.mdiInformationOutline }}</v-icon>
          <div class="ml-2">
            The addresses part of this list can be excluded from processing by heuristics.
            This list contains {{ Number(addressCount).toLocaleString() }} address exclusions.
            The list below is limited to 30 addresses.
          </div>
        </div>
      </v-card-text>
      <v-card-text v-else>
        <div class="d-flex justify-center">
          <v-btn @click="addAddressExclusions = true" text>
            <v-icon>{{ icon.mdiFileImport }}</v-icon>
            Import address exclusions
          </v-btn>
        </div>
      </v-card-text>
    </v-card>
    <v-row v-if="items.length > 0" class="mt-2 mx-auto"
           style="max-width: 1200px; background-color: transparent">
      <v-col v-for="addressHash in items" :key="addressHash" cols="12" sm="6" md="4" lg="4">
        <v-card elevation="4">
          <div class="d-flex" style="flex-wrap: nowrap;">
            <div style="min-width: 100px" class="flex-grow-0 flex-shrink-1">
              <v-list-item :to="{ name: routes.addressRoute, params: { id: addressHash }}">
                <v-list-item-title>
                  {{ addressHash }}
                </v-list-item-title>
              </v-list-item>
            </div>
            <div class="align-self-center ml-auto">
              <v-menu bottom left>
                <template v-slot:activator="{ on, attrs }">
                  <v-btn icon v-bind="attrs" v-on="on">
                    <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
                  </v-btn>
                </template>
                <v-list>
                  <v-list-item @click="deleteItem(addressHash)">
                    <v-list-item-icon>
                      <v-icon>{{ icon.mdiDelete }}</v-icon>
                    </v-list-item-icon>
                    <v-list-item-title>Delete</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </div>
          </div>
        </v-card>
      </v-col>
    </v-row>
    <import-address-exclusions v-model="addAddressExclusions" @added="loadData"/>
    <delete-all-address-exclusions v-model="deleteAllExclusionsDialog"
                                   :count="addressCount" @deleted="loadData"/>
    <delete-address-exclusion v-model="deleteExclusionDialog"
                              :address-hash="deleteAddressHash"
                              @deleted="handleExclusionDeletion"/>
  </div>
</template>

<script>
import {
  mdiPlaylistRemove, mdiDelete, mdiDotsVertical, mdiFileImport, mdiInformationOutline,
} from '@mdi/js';
import { PAGE_TITLE, ROUTE_ADDRESS_EXCLUSION_OVERVIEW, ROUTE_NAME_ADDRESS_PAGE } from '../../constants';
import { doGet, handleError } from '../../utilities';
import ImportAddressExclusions from '../dialogs/ImportAddressExclusions.vue';
import DeleteAddressExclusion from '../dialogs/DeleteAddressExclusion.vue';
import DeleteAllAddressExclusions from '../dialogs/DeleteAllAddressExclusions.vue';

export default {
  name: 'AddressExclusions',
  components: {
    DeleteAllAddressExclusions, DeleteAddressExclusion, ImportAddressExclusions,
  },
  data() {
    return {
      icon: {
        mdiPlaylistRemove, mdiDelete, mdiDotsVertical, mdiFileImport, mdiInformationOutline,
      },
      routes: {
        addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      },
      addAddressExclusions: false,
      deleteExclusionDialog: false,
      deleteAllExclusionsDialog: false,
      deleteAddressHash: '',
      items: [],
      addressCount: -1,
    };
  },
  methods: {
    loadData() {
      this.items = [];
      doGet(ROUTE_ADDRESS_EXCLUSION_OVERVIEW, this.$router, this.$store)
        .then((data) => {
          if (!data.success || data.addresses === undefined) throw new Error('could not get address exclusion data');

          if (data.addresses === null) {
            this.items = [];
            return;
          }

          this.addressCount = data.addressCount;

          this.items = data.addresses;
        })
        .catch((e) => {
          handleError(this.$store, e);
        });
    },
    deleteItem(addressHash) {
      this.deleteAddressHash = addressHash;
      this.deleteExclusionDialog = true;
    },
    handleExclusionDeletion(addressHash) {
      this.addressCount -= 1;
      this.items = this.items.filter((d) => d !== addressHash);
    },
  },
  mounted() {
    document.title = `Address Exclusions - ${PAGE_TITLE}`;

    this.loadData();
  },
};
</script>

<style scoped>

</style>
