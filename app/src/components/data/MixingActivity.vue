<template>
  <div>
    <v-row>
      <v-col>
        <v-select
            item-value="id"
            item-text="text"
            label="Filter by Privacy Type"
            v-model="selectedPrivacyLabel"
            style="max-width: 300px; min-width: 200px;"
            :items="privacyLabels"
            @change="updateSvgData(false)">
          <template v-slot:item="{ item }">
            <span>
               <v-chip label :outlined="item.color === undefined"
                       :color="item.color?item.color:'black'" small/>
              {{ item.text }}
            </span>
          </template>
        </v-select>
      </v-col>
      <v-col>
        <v-menu
            v-model="menu"
            :close-on-content-click="false"
            :nudge-right="40"
            transition="scale-transition"
            offset-y
            min-width="auto"
            @input="handleMenuChange">
          <template v-slot:activator="{ on, attrs }">
            <v-text-field
                :value="dateRangeString"
                label="Filter by Date"
                :prepend-icon="icons.mdiCalendarRange"
                readonly
                v-bind="attrs"
                v-on="on"/>
          </template>
          <v-date-picker
              no-title
              range
              color="primary"
              v-model="dateRange"/>
        </v-menu>
      </v-col>
      <v-col>
        <v-switch
            label="Include cluster addresses"
            v-model="includeCusterAddresses"
            @click="updateSvgData(true)"/>
      </v-col>
    </v-row>
    <p v-if="showNotEnoughDataMessage && !isLoading" class="text-h6" style="text-align: center">
      No enough data available to draw chart
    </p>
    <p v-if="showEmptyResponseMessage && !isLoading" class="text-h6" style="text-align: center">
      No data available
    </p>
    <v-progress-linear v-if="isLoading" indeterminate/>
    <div v-if="showHistogram" class="text-subtitle-1" style="text-align: center">
      {{ selectedPrivacyLabel === 'all' ? 'All Privacy' : capitalize(selectedPrivacyLabel) }}
      Transactions
    </div>
    <svg id="mixing_activity_canvas" :class="{'hide': !showHistogram}"/>
    <v-expand-transition>
      <v-data-table
          v-if="barTable.transactions.length > 0"
          :headers="barTable.headers"
          :items="barTable.transactions">
        <template v-slot:top>
          <v-toolbar flat class="hidden-sm-and-up">
            <v-toolbar-title>
              Privacy Transactions from {{ barTable.transactions.startDate }}
              to {{ barTable.transactions.endDate }}
            </v-toolbar-title>
          </v-toolbar>
          <v-toolbar flat class="hidden-xs-only">
            <v-toolbar-title>
              Privacy Transactions from {{ barTable.transactions.startDate }}
              to {{ barTable.transactions.endDate }}
            </v-toolbar-title>
          </v-toolbar>
        </template>
        <template v-slot:[`item.txhash`]="{ item }">
          <router-link :to="{ name: txRoute, params: { id: item.txhash }}">
            {{ item.txhash }}
          </router-link>
        </template>
        <template v-slot:[`item.dateTime`]="{ item }">
          <span>{{ item.dateTime.toLocaleString() }}</span>
        </template>
      </v-data-table>
    </v-expand-transition>

  </div>
</template>

<script>
import { mdiCalendarRange } from '@mdi/js';
import Histogram from '../../d3Documents/histogram';
import { doPost, getPrivacyTypeLabel, handleError } from '../../utilities';
import { ROUTE_MIXING_ACTIVITY, ROUTE_NAME_TRANSACTION_PAGE } from '../../constants';

// capitalize returns the first letter of each word (separated by an space) in str capitalized
function capitalize(str) {
  return str.split(' ').map((d) => d[0].toUpperCase() + d.slice(1)).join(' ');
}

export default {
  name: 'MixingActivity',
  props: {
    addressHash: { type: String, required: true },
    includeCluster: { type: Boolean, required: false, default: true },
  },
  data() {
    return {
      icons: {
        mdiCalendarRange,
      },
      txRoute: ROUTE_NAME_TRANSACTION_PAGE,
      selectedPrivacyLabel: 'all',
      showHistogram: false,
      colorMap: new Map(),
      svgHistogram: null,
      isLoading: false,
      showEmptyResponseMessage: false,
      showNotEnoughDataMessage: false,
      activities: null,
      initialLoadDone: false,
      includeCusterAddresses: true,
      dateRange: null,
      menu: null,
      oldDateRange: null,
      barTable: {
        headers: [{
          text: 'Transaction', align: 'start', value: 'txhash',
        },
        { text: 'Timestamp', value: 'dateTime' },
        { text: 'Privacy Type', value: 'privacytype' },
        ],
        transactions: [],
      },
    };
  },
  computed: {
    privacyLabels() {
      const labels = [{ text: 'All', id: 'all' }];

      this.colorMap.forEach((v, k) => {
        labels.push({ text: capitalize(k), color: v, id: k });
      });

      return labels;
    },
    dateRangeString() {
      if (!this.dateRange) {
        return [];
      }

      return this.dateRange.map((d) => new Date(d).toLocaleDateString());
    },
  },
  methods: {
    capitalize,
    onBarClick(data) {
      if (data.x0.getHours() === data.x1.getHours()
          && data.x0.getMinutes() === data.x1.getMinutes()) {
        data.startDate = data.x0.toLocaleDateString();
        data.endDate = data.x1.toLocaleDateString();
      } else {
        data.startDate = data.x0.toLocaleString();
        data.endDate = data.x1.toLocaleString();
      }

      this.barTable.transactions = data;
    },
    handleMenuChange(open) {
      if (open) {
        this.oldDateRange = this.dateRange;
      } else if (this.oldDateRange !== this.dateRange) {
        this.updateSvgData(false);
      }
    },
    getMixingActivity() {
      return doPost(ROUTE_MIXING_ACTIVITY, this.$router, this.$store,
        { addressHash: this.addressHash, isClusterLookup: this.includeCusterAddresses })
        .then((data) => data)
        .catch((e) => {
          handleError(this.$store, e);
          return [];
        });
    },
    async updateSvgData(pullNewData) {
      this.showHistogram = false;
      this.isLoading = true;

      if (pullNewData || !this.initialLoadDone) {
        const mixingActivity = await this.getMixingActivity();
        if (mixingActivity.activities === undefined) {
          this.showEmptyResponseMessage = true;
          this.isLoading = false;
          return;
        }

        this.activities = mixingActivity.activities.map((d) => {
          d.privacytype = getPrivacyTypeLabel(d.privacytype);
          d.parsedDate = new Date(d.ts);
          return d;
        });

        this.initialLoadDone = true;
      }

      this.showEmptyResponseMessage = false;

      let fromDate = null;
      let toDate = null;
      let considerDate = true;

      if (this.dateRange !== null) {
        fromDate = new Date(this.dateRange[0]);
        if (this.dateRange.length === 2) {
          toDate = new Date(this.dateRange[1]);
        } else {
          toDate = new Date(this.dateRange[0]);
        }

        // check if dates are in the right order
        if (toDate < fromDate) {
          const mem = toDate;
          toDate = fromDate;
          fromDate = mem;
        }
      } else {
        considerDate = false;
      }

      const filtered = this.activities.filter((d) => {
        if (this.selectedPrivacyLabel !== 'all' && d.privacytype !== this.selectedPrivacyLabel) return false;

        if (!considerDate) return true;

        return d.parsedDate <= toDate && d.parsedDate >= fromDate;
      });

      let categories = [];

      if (this.selectedPrivacyLabel !== 'all') {
        categories = [this.selectedPrivacyLabel];
      } else {
        // find categories
        categories = [...new Set(filtered.map((d) => d.privacytype))];
      }

      this.svgHistogram.reset();
      this.svgHistogram.drawStacked(filtered, categories, this.colorMap);

      this.showHistogram = !this.svgHistogram.empty;
      this.showNotEnoughDataMessage = this.svgHistogram.empty;
      this.isLoading = false;
    },
  },
  beforeMount() {
    this.includeCusterAddresses = this.includeCluster;
    this.colorMap.set('destination', '#0072B2');
    this.colorMap.set('collateral creation', '#E69F00');
    this.colorMap.set('collateral payment', '#009E73');
    this.colorMap.set('origin', '#D55E00');
    this.colorMap.set('mixing', '#56B4E9');

    this.svgHistogram = new Histogram('mixing_activity_canvas',
      1200, 300, 'Privacy transactions');
    this.svgHistogram.setClickHandler(this.onBarClick);
  },
  mounted() {
    this.updateSvgData();
  },
};
</script>

<style scoped>
>>> .overlay {
  stroke-width: 2px;
  stroke: red;
  fill: red;
  cursor: pointer;
}

.hide {
  display: none;
}
</style>
