<template>
  <v-card
      class="mx-auto elevation-12"
      max-width="900">
    <v-toolbar color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icon.mdiChartTimelineVariant }}</v-icon>
        Shortest Path
      </v-toolbar-title>
    </v-toolbar>
    <v-card-text>
      <div class="text-subtitle-1">
        Find the shortest path between two transactions.
        <v-hover v-slot:default="{ hover }" open-delay="0">
          <v-icon small id="shortest_path_tooltip">
            {{ hover ? icon.mdiInformation : icon.mdiInformationOutline }}
          </v-icon>
        </v-hover>
        <v-tooltip right activator="#shortest_path_tooltip">
          <span>The result is nondeterministic, because multiple shortest paths can exist.</span>
        </v-tooltip>
      </div>
      <v-row>
        <v-col>
          <v-text-field label="From"/>
        </v-col>
        <v-col>
          <v-text-field label="To"/>
        </v-col>
      </v-row>
      <v-row>
        <v-col class="d-flex justify-end align-center">
          <v-switch label="Include private transactions" class="mx-5"/>
          <v-btn color="primary">
            Search
          </v-btn>
        </v-col>
      </v-row>
      <v-timeline :dense="this.$vuetify.breakpoint.smAndDown">
        <v-timeline-item v-for="(tx, i) in transactions" :key="i" small>
          <template v-slot:opposite><span class="headline" v-text="tx.year"></span></template>
          <div class="py-4">
            <div class="text-h5 font-weight-bold mb-4">{{ tx.hash }}</div>
            <div>
              Lorem ipsum dolor sit amet, no nam oblique veritus. Commune scaevola
              imperdiet nec ut, sed euismod convenire principes at. Est et nobis iisque
              percipit, an vim zril disputando voluptatibus, vix an salutandi sententiae.
            </div>
          </div>
        </v-timeline-item>
      </v-timeline>
    </v-card-text>
  </v-card>
</template>

<script>
import {
  mdiChartTimelineVariant, mdiInformation, mdiInformationOutline,
} from '@mdi/js';
import { PAGE_TITLE } from '../../constants';

export default {
  name: 'ShortestPath',
  data() {
    return {
      icon: {
        mdiChartTimelineVariant, mdiInformation, mdiInformationOutline,
      },
      transactions: [
        {
          hash: 'a84676b18321c62ed3aad252a667c62c3afa89c6ff07a464eccb22bba7005d31',
          year: '01/02/2021, 22:03:08 ',
        },
        {
          hash: '307e4d59d205e8ffa158715d3b9975fd4a6381c86d4ce64b210a0f3c517bfd9c',
          year: '01/02/2021, 22:03:08 ',
        },
        {
          hash: 'bb6c491addbb8f1414a81c991684cf6a160e3ac59432341a00d76e50aefe600a',
          year: '01/02/2021, 22:03:08 ',
        },
        {
          hash: '6e8da7f496b02086f6da34ea50fe35d3dd181972cb8f9f8b5e07a4565b5759c6',
          year: '01/02/2021, 22:03:08 ',
        },
        {
          hash: 'b700aa661deb845b1659687f92b36f4591d9bbce7a80ef76c934d3cb151ebfa2',
          year: '01/02/2021, 22:03:08 ',
        },
      ],
    };
  },
  mounted() {
    document.title = `Shortest Path - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

</style>
