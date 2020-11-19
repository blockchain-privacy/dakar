<template>
  <v-container class="fill-height" fluid v-if="data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar :color="data.privacytype?'purple':'primary'" dark flat>
            <v-toolbar-title>
              <v-icon>mdi-transfer</v-icon>
              Transaction {{ data.txhash }}
            </v-toolbar-title>
            <v-spacer></v-spacer>
            <v-tooltip bottom>
              <template v-slot:activator="{ on, attrs }">
              <v-btn  outlined icon @click="goToHeuristicPage" v-on="on" v-bind="attrs">
                <v-icon>mdi-graph</v-icon>
              </v-btn>
              </template>
              <span>Open the heuristic editor for this transaction.</span>
            </v-tooltip>
<!--            todo remove?-->
<!--            <v-tooltip bottom>-->
<!--              <template v-slot:activator="{ on, attrs }">-->
<!--                <v-btn :loading="isLoading" style="padding-left: 2px; padding-right: 4px" outlined v-on:click="getCSV"-->
<!--                       v-if="data.origincount > 0" v-on="on" v-bind="attrs"-->
<!--                       class="d-none d-sm-flex" :disabled="data.origincount > csvDownloadMaxOrigins || isLoading">-->
<!--                  <v-icon>mdi-download</v-icon>-->
<!--                  {{ data.origincount }} origins-->
<!--                </v-btn>-->
<!--                <v-btn :loading="isLoading" icon v-on:click="getCSV" v-if="data.origincount > 0" v-on="on"-->
<!--                       v-bind="attrs"-->
<!--                       class="d-flex d-sm-none" :disabled="data.origincount > csvDownloadMaxOrigins || isLoading">-->
<!--                  <v-icon large>mdi-download</v-icon>-->
<!--                </v-btn>-->
<!--              </template>-->
<!--              <span>Download potential origins of this transaction. {{ data.origincount }} origins found.</span>-->
<!--            </v-tooltip>-->
          </v-toolbar>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-format-list-numbered" title="Block Height">
                    <router-link :to="data.bid.toString()"> {{ data.bid }}</router-link>
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-calendar" title="Timestamp">
                    {{ new Date(data.bts).toLocaleString() }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col v-if="(data.fee || data.fee === 0) && data.fee >= 0">
                  <IconItem icon="mdi-cash" title="Fee">
                    {{ convertAmount(data.fee) }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-format-header-pound" title="Block">
                    <router-link :to="data.bhash">{{ shortenHash(data.bhash) }}</router-link>
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-pound" title="Number of outputs">
                    {{ !data.outputs ? 0 : data.outputs.length }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-pound" title="Number of inputs">
                    {{ !data.inputs ? 0 : data.inputs.length }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-divider v-if="outputs"></v-divider>
              <v-row v-if="outputs">
                <v-col v-for="i in outputs" v-bind:key="i.addresshash + i.outputindex">
                  <v-sheet min-height="50" class="fill-height" color="transparent">
                    <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 1}">
                      <IconItem icon="mdi-currency-usd-circle-outline" title="Output">
                        Address hash:
                        <router-link :to="i.addresshash">{{ i.addresshash }}</router-link>
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
              <v-divider v-if="inputs"></v-divider>
              <v-row v-if="inputs">
                <v-col v-for="i in inputs" v-bind:key="i.addresshash + i.inputindex">
                  <v-sheet min-height="50" class="fill-height" color="transparent">
                    <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 1}">
                      <IconItem icon="mdi-currency-usd-circle" title="Input">
                        Address hash:
                        <router-link :to="i.addresshash">{{ i.addresshash }}</router-link>
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
import {shortenHash, convertAmount} from "@/utilities";
import {PAGE_TITLE, ROUTE_PATHS, CSV_DOWNLOAD_MAX_ORIGINS, ROUTE_NAME_HEURISTIC_PAGE} from "@/constants";
import IconItem from "@/components/common/IconItem";

export default {
  name: 'TxLookup',
  components: {IconItem},
  data: function () {
    return {
      isLoading: false,
      csvDownloadMaxOrigins: CSV_DOWNLOAD_MAX_ORIGINS,
    }
  },
  computed: {
    data() {
      return this.$store.getters.getTransactionData;
    },
    inputs() {
      return this.sortByInput(this.data.inputs);
    },
    outputs() {
      return this.sortByOutput(this.data.outputs);
    }
  },
  methods: {
    shortenHash,
    convertAmount,
    goToHeuristicPage() {
      this.$router.push({name: ROUTE_NAME_HEURISTIC_PAGE})
    },
    sortByOutput(outputs) {
      if (outputs == null) return null;
      return outputs.sort((a, b) => {
        return a.outputindex > b.outputindex
      })
    },
    sortByInput(inputs) {
      if (inputs == null) return null;
      return inputs.sort((a, b) => {
        return a.inputindex > b.inputindex
      })
    },
    getCSV: function () {
      const options = {
        headers: {
          // header for pass through
          Accept: '*/*'
        }
      };
      this.isLoading = true;
      fetch(ROUTE_PATHS + this.data.txhash, options)
          .then(res => res.blob())
          .then(blob => {
            // looks hacky, but it is the only way with good UX
            const a = document.createElement("a");
            a.href = URL.createObjectURL(blob);
            a.setAttribute("download", `${this.data.txhash}.csv`);
            a.click();
            a.remove();
            this.isLoading = false;
          })
          .catch(error => {
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
      h = ` ${this.data.txhash} `
    }
    document.title = `Transaction${h}- ${PAGE_TITLE}`;
  },
}
</script>
