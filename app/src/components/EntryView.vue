<template>
  <v-layout row v-if="data">
    <v-flex xs12 sm8 offset-sm2>
      <v-card>
        <v-card-title>
          <v-icon>mdi-database</v-icon>
          Database Information
        </v-card-title>
        <v-list-item>
          <v-list-item-avatar>
            <v-icon>mdi-progress-wrench</v-icon>
          </v-list-item-avatar>
          <v-list-item-content>
            <v-list-item-title>Crawler status</v-list-item-title>
            <v-list-item-subtitle>
              {{ data.iscrawling?"Active":"Inactive" }}
            </v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
        <v-list two-line subheader>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-format-list-numbered</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Last block ID</v-list-item-title>
              <v-list-item-subtitle>
                <router-link :to="'search/' + data.lastblockid">{{ data.lastblockid }}</router-link>
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Global block count</v-list-item-title>
              <v-list-item-subtitle>
                {{ data.blkcount }}
              </v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-avatar>
              <v-icon>mdi-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Global transaction count</v-list-item-title>
              <v-list-item-subtitle>
                {{ data.txcount }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Global output count</v-list-item-title>
              <v-list-item-subtitle>
                {{ data.outputcount }}
              </v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-avatar>
              <v-icon>mdi-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Global address count</v-list-item-title>
              <v-list-item-subtitle>
                {{ data.addresscount }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-card>
      <v-card><!-- TODO hack, to leave space at the bottom, such that content is not hidden by the footer.-->
        <div style="height: 40pt">
        </div>
      </v-card>
    </v-flex>
  </v-layout>
</template>

<script>
export default {
  name: 'EntryView',
  computed: {
    data() {
      return this.$store.getters.getMetaData;
    },
  },
  created() {
    this.$store.dispatch('updateMetaData');
  }
}
</script>

<style scoped>

</style>