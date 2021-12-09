<template>
  <v-card
      class="mx-auto elevation-4" max-width="1000">
    <v-toolbar color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icon.mdiMerge }}</v-icon>
        Add Cluster
      </v-toolbar-title>
    </v-toolbar>
    <v-card-text>
      <div class="text-subtitle-1">
        Add custom address clusters by uploading a CSV-file.
        The file must have two columns, where the first column contains an
        identifier for each cluster and the second column the addresses.
      </div>
      <v-expansion-panels flat>
        <v-expansion-panel>
          <v-expansion-panel-header>
            Example CSV-file
          </v-expansion-panel-header>
          <v-expansion-panel-content>
            <p>The following file content would generate two clusters with two addresses each.</p>
            <pre><code>cluster-id,address
1,XgG6Nosmei5woQ2VTDzwmLX7SzdNYKHdiz
1,Xf36MqBkoK8G5wBbjUSwDRy6XTjdNq8hgB
2,XatWuw7BhTxHvjPLbnvPArWgW9r6hjpt8o
2,XcsCPgY67TqW9CpsJLCbizDw2Yq2zFoh74</code></pre>
          </v-expansion-panel-content>
        </v-expansion-panel>
      </v-expansion-panels>
      <v-form ref="csvForm" id="csvForm">
        <v-file-input
            v-model="csv.file"
            :rules="rules.file"
            show-size
            accept="text/csv"
            label="Click here to select a file"
            truncate-length="15"/>
        <v-row>
          <v-col>
            <v-switch v-model="csv.firstRowContainsHeader"
                      label="First row of file contains headers" :disabled="isLoading"/>
          </v-col>
          <v-col>
            <v-select
                v-model="csv.separator"
                :items="separatorItems"
                item-text="text"
                item-value="value"
                label="Separator">
            </v-select>
          </v-col>
          <v-col class="d-flex justify-end align-center">
            <v-btn
                color="primary"
                :loading="isLoading"
                @click="handleCSVUpload">
              Upload
            </v-btn>
          </v-col>
        </v-row>
      </v-form>
    </v-card-text>
  </v-card>
</template>

<script>
import { mdiFileDownloadOutline, mdiMerge } from '@mdi/js';
import { PAGE_TITLE, ROUTE_ADD_CLUSTER } from '../../constants';
import { doPostUpload } from '../../utilities';

export default {
  name: 'AddCluster.vue',
  data() {
    return {
      icon: {
        mdiMerge, mdiFileDownloadOutline,
      },
      isLoading: false,
      separatorItems: [
        { text: 'Colon (,)', value: ',' },
        { text: 'Semicolon (;)', value: ';' },
      ],
      csv: {
        valid: false,
        file: null,
        separator: ',',
        firstRowContainsHeader: false,
      },
      rules: {
        file: [(v) => !!v || 'File is required'],
        separator: [
          (v) => !!v || 'Separator is required',
          (v) => (v && v.length <= 10) || 'Separator must not greater than 10 characters',
        ],
      },
    };
  },
  methods: {
    handleCSVUpload() {
      if (!this.$refs.csvForm.validate()) return;

      // create and fill form data object
      const newForm = new FormData();
      newForm.append('file', this.csv.file);
      newForm.append('separator', this.csv.separator);
      newForm.append('hasHeader', this.csv.firstRowContainsHeader ? '1' : '0');

      // upload to server
      doPostUpload(ROUTE_ADD_CLUSTER, this.$router, this.$store, newForm)
        .then((response) => {
          if (!response.success) {
            console.log('error processing inputs');

            if (response.msg) {
              console.log(response.msg);
            }
          } else {
            console.log('success');
          }
        });
    },
  },
  mounted() {
    document.title = `Add Cluster - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

</style>
