<template>
  <div class="flex-column d-flex" style="height: 100%;">
    <v-toolbar
        dense dark
        color="primary"
        style="z-index: 10; box-shadow: 0 2px 4px -1px rgba(0, 0, 0, 0.2);">
      <v-toolbar-title class="hidden-md-and-up">
        {{ this.addressHash }}
      </v-toolbar-title>
      <v-toolbar-title class="hidden-sm-and-down">
        <v-icon>{{ icon.mdiCardBulletedOutline }}</v-icon>
        Address {{ this.addressHash }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-menu bottom>
        <v-list>
          <v-list-item :to="{ name: routeAddress, params: { id: addressHash }}">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiOpenInNew }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Transaction Page</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>
    <svg id="cluster_view_canvas"/>
  </div>
</template>

<script>
import { mdiCardBulletedOutline } from '@mdi/js';
import {
  APPLICATION_NAME, ROUTE_CLUSTER_HMI_LOOKUP, ROUTE_NAME_ADDRESS_PAGE,
} from '../../constants';
import { doGet, handleError } from '../../utilities';
import ClusterTree from '../../d3Documents/clusterTree';

function newRouting(context) {
  const { id } = context.$route.params;
  if (id === undefined || context.$route.name !== ROUTE_CLUSTER_HMI_LOOKUP) {
    return;
  }

  context.onMounted();
}

export default {
  name: 'HMIView',
  data() {
    return {
      svgCanvasId: 'cluster_view_canvas',
      addressHash: '',
      routeAddress: ROUTE_NAME_ADDRESS_PAGE,
      icon: {
        mdiCardBulletedOutline,
      },
      hmiData: {
        clusters: [],
        addressCluster: '',
      },
      ct: new ClusterTree(150),
    };
  },
  methods: {
    loadClusterData(addressHash) {
      return doGet(ROUTE_CLUSTER_HMI_LOOKUP, this.$router, this.$store, addressHash)
        .then((data) => {
          if (!data.success) return;
          this.hmiData.clusters = data.clusters;
          this.hmiData.addressCluster = data.address_cluster;
          this.$store.dispatch('resetMessages');
        }).catch((e) => {
          handleError(this.$store, e);
          return e;
        });
    },
    updateGraph() {
      // maps the node data to the tree layout
      // this.ct.processGraphData(this.hmiData.clusters);

      this.ct.drawForce(this.hmiData.clusters);
    },
    async refreshData() {
      await this.loadClusterData(this.addressHash);

      if (!this.hmiData.addressCluster) {
        return false;
      }

      this.updateGraph();

      return true;
    },
    clusterClickHandler() {
    },
    async onMounted() {
      // remove previous svg children
      document.getElementById(this.svgCanvasId).innerHTML = '';

      // set transaction hashes for this page view
      this.addressHash = this.$route.params.id;

      // set page title
      document.title = `Hierarchical Cluster View ${this.addressHash} - ${APPLICATION_NAME}`;

      if (!this.ct.setNodeClickHandler(this.clusterClickHandler)) {
        this.setErrorMessage('error setting heuristic click handler');
        return false;
      }

      this.ct.setupSvg(this, this.svgCanvasId);

      if (!await this.refreshData()) {
        return false;
      }

      await this.ct.centerGraph();

      return true;
    },
  },
  mounted() {
    this.onMounted();
  },
  watch: {
    $route() {
      newRouting(this);
    },
  },
};
</script>

<style scoped>

>>> .node text {
  font: 12px sans-serif;
  cursor: pointer;
}

>>> .link {
  fill: none;
  stroke: darkslategrey;
  stroke-width: 2px;
}

>>> .rect {
  stroke: #008ee5;
  fill-opacity: 0;
  cursor: pointer;
}

>>> .clicked {
  stroke: #FDD835;
}

>>> .modified {
  stroke-dasharray: 5;
}

>>> .selected {
  fill: #9CCC65;
  fill-opacity: 1;
}

>>> #cluster_view_canvas {
  height: 100%;
}
</style>
