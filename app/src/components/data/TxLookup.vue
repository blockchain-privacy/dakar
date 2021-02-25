<template>
  <v-container fluid v-if="data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8" v-for="tx in data" :key="tx.txhash+tx.bid">
        <v-card class="elevation-4">
          <v-toolbar :color="tx.privacytype?'purple':'primary'" dark flat>
            <v-toolbar-title>
              <v-icon>{{ icon.mdiTransfer }}</v-icon>
              Transaction {{ tx.txhash }}
            </v-toolbar-title>
            <v-spacer></v-spacer>
            <v-btn
                id="btn_open_heuristic_editor"
                v-if="tx.privacytype === 'destination' && showHeuristicEditor"
                style="margin-right: 0" outlined icon
                @click="goToHeuristicPage">
              <v-icon>{{ icon.mdiGraph }}</v-icon>
            </v-btn>
            <v-tooltip bottom activator="#btn_open_heuristic_editor">
              <span>Open the heuristic editor for this transaction.</span>
            </v-tooltip>
            <!--            todo remove?-->
            <!--  <v-tooltip bottom>-->
            <!--    <template v-slot:activator="{ on, attrs }">-->
            <!--      <v-btn :loading="isLoading" style="padding-left: 2px; padding-right: 4px"-->
            <!--             outlined v-on:click="getCSV" v-if="data.origincount > 0" -->
            <!--             v-on="on" v-bind="attrs"-->
            <!--             class="d-none d-sm-flex"-->
            <!--             :disabled="data.origincount > csvDownloadMaxOrigins || isLoading">-->
            <!--        <v-icon>mdi-download</v-icon>-->
            <!--        {{ data.origincount }} origins-->
            <!--      </v-btn>-->
            <!--      <v-btn :loading="isLoading" icon v-on:click="getCSV"-->
            <!--             v-if="data.origincount > 0" v-on="on" v-bind="attrs"-->
            <!--             class="d-flex d-sm-none"-->
            <!--             :disabled="data.origincount > csvDownloadMaxOrigins || isLoading">-->
            <!--        <v-icon large>mdi-download</v-icon>-->
            <!--      </v-btn>-->
            <!--    </template>-->
            <!--    <span>Download potential origins of this transaction.-->
            <!--      {{ data.origincount }} origins found.</span>-->
            <!--  </v-tooltip>-->
          </v-toolbar>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col>
                  <IconItem :icon="icon.mdiFormatListNumbered" title="Block Height">
                    <router-link :to="{ name: blockRoute, params: { id: tx.bid }}">
                      {{ tx.bid }}
                    </router-link>
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem :icon="icon.mdiCalendar" title="Timestamp">
                    {{ new Date(tx.bts).toLocaleString() }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col v-if="(tx.fee || tx.fee === 0) && tx.fee >= 0">
                  <IconItem :icon="icon.mdiCash" title="Fee">
                    {{ convertAmount(tx.fee) }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem :icon="icon.mdiFormatHeaderPound" title="Block">
                    <router-link :to="{ name: blockRoute, params: { id: tx.bhash }}">
                      {{ shortenHash(tx.bhash) }}
                    </router-link>
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem :icon="icon.mdiPound" title="Number of outputs">
                    {{ !tx.outputs ? 0 : tx.outputs.length }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem :icon="icon.mdiPound" title="Number of inputs">
                    {{ !tx.inputs ? 0 : tx.inputs.length }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-divider v-if="tx.outputs"></v-divider>
              <v-row v-if="tx.outputs">
                <v-col v-for="i in sortByOutput(tx.outputs)"
                       v-bind:key="i.addresshash + i.outputindex">
                  <v-sheet min-height="50" class="fill-height" color="transparent">
                    <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 1}">
                      <IconItem :icon="icon.mdiCurrencyUsdCircleOutline" title="Output">
                        Address hash:
                        <router-link :to="{ name: addressRoute, params: { id: i.addresshash }}">
                          {{ i.addresshash }}
                        </router-link>
                        <br>
                        Amount: {{ convertAmount(i.amount) }}<br>
                        Spent: {{ i.inputindex != null }}<br>
                        Index: {{ i.outputindex }}<br>
                        Coinbase: {{ i.iscoinbase }}
                      </IconItem>
                    </v-lazy>
                  </v-sheet>
                </v-col>
              </v-row>
              <v-divider v-if="tx.inputs"></v-divider>
              <v-row v-if="tx.inputs">
                <v-col v-for="i in sortByInput(tx.inputs)"
                       v-bind:key="i.addresshash + i.inputindex">
                  <v-sheet min-height="50" class="fill-height" color="transparent">
                    <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 1}">
                      <IconItem :icon="icon.mdiCurrencyUsdCircle" title="Input">
                        Address hash:
                        <router-link :to="{ name: addressRoute, params: { id: i.addresshash }}">
                          {{ i.addresshash }}
                        </router-link>
                        <br>
                        Amount: {{ convertAmount(i.amount) }}<br>
                        Index: {{ i.inputindex }}<br>
                        Coinbase: {{ i.iscoinbase }}
                      </IconItem>
                    </v-lazy>
                  </v-sheet>
                </v-col>
              </v-row>
            </v-container>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {
  mdiTransfer, mdiGraph, mdiFormatListNumbered, mdiCalendar,
  mdiCash, mdiFormatHeaderPound, mdiPound, mdiCurrencyUsdCircleOutline,
  mdiCurrencyUsdCircle,
} from '@mdi/js';
import { shortenHash, convertAmount, getCurrentDate } from '../../utilities';
import {
  PAGE_TITLE, ROUTE_PATHS, CSV_DOWNLOAD_MAX_ORIGINS, ROUTE_NAME_HEURISTIC_PAGE,
  ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_ADDRESS_PAGE,
} from '../../constants';
import IconItem from '../common/IconItem.vue';

export default {
  name: 'TxLookup',
  components: { IconItem },
  data() {
    return {
      icon: {
        mdiTransfer,
        mdiGraph,
        mdiFormatListNumbered,
        mdiCalendar,
        mdiCash,
        mdiFormatHeaderPound,
        mdiPound,
        mdiCurrencyUsdCircleOutline,
        mdiCurrencyUsdCircle,
      },
      blockRoute: ROUTE_NAME_BLOCK_PAGE,
      addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      isLoading: false,
      csvDownloadMaxOrigins: CSV_DOWNLOAD_MAX_ORIGINS,
    };
  },
  computed: {
    data() {
      return this.$store.getters.getTransactionData;
    },
    userData: {
      get() {
        return this.$store.getters.getActiveUser;
      },
      set(value) {
        this.$store.dispatch('setActiveUser', value);
      },
    },
    showHeuristicEditor() {
      return this.userData && this.userData.roles
          && this.userData.roles.some((d) => d.role_name === 'admin' || d.role_name === 'privileged');
    },
  },
  methods: {
    shortenHash,
    convertAmount,
    goToHeuristicPage() {
      this.$router.push({ name: ROUTE_NAME_HEURISTIC_PAGE });
    },
    sortByOutput(outputs) {
      if (outputs == null) return null;
      const copiedOutputs = JSON.parse(JSON.stringify(outputs));
      return copiedOutputs.sort((a, b) => a.outputindex > b.outputindex);
    },
    sortByInput(inputs) {
      if (inputs == null) return null;
      const copiedInputs = JSON.parse(JSON.stringify(inputs));
      return copiedInputs.sort((a, b) => a.inputindex > b.inputindex);
    },
    getCSV() {
      const options = {
        headers: {
          // header for pass through
          Accept: '*/*',
        },
      };
      this.isLoading = true;
      fetch(ROUTE_PATHS + this.data.txhash, options)
        .then((res) => res.blob())
        .then((blob) => {
          // looks hacky, but it is the only way with good UX
          const a = document.createElement('a');
          a.href = URL.createObjectURL(blob);
          a.setAttribute('download', `paths_${getCurrentDate}_${this.data.txhash}.csv`);
          a.click();
          a.remove();
          this.isLoading = false;
        })
        .catch((error) => {
          this.errorMsg = error;
          this.isLoading = false;
        });
    },
  },
  mounted() {
    document.title = `Transaction - ${PAGE_TITLE}`;
  },
  updated() {
    let h = ' ';
    if (this.data && this.data.txhash) {
      h = ` ${this.data.txhash} `;
    }
    document.title = `Transaction${h}- ${PAGE_TITLE}`;
  },
};
</script>
