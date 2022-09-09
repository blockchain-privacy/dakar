<template>
  <div class="fill-height" style="padding: 12px 10px 0 10px">
    <v-row class="fill-height">
      <v-col cols="2" class="hidden-md-and-down pa-0">
        <v-navigation-drawer permanent>
          <v-list-item :to="{name: routeWikiRoot}" exact-path>
            <v-list-item-icon>
              <v-icon>{{ icons.mdiBookOpen }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title class="text-h6">
              Wiki
            </v-list-item-title>
          </v-list-item>
          <v-divider></v-divider>
          <v-list dense nav v-if="fileSet">
            <div v-for="fileItem in fileHierarchy"
                 :key="fileItem.name">
              <v-list-group
                  v-if="fileItem.items"
                  v-model="fileItem.active"
                  :prepend-icon="icons.mdiBook"
                  no-action>
                <template v-slot:activator>
                  <v-list-item-content>
                    <v-list-item-title>{{ fileItem.name }}</v-list-item-title>
                  </v-list-item-content>
                </template>
                <v-list-item
                    v-for="child in fileItem.items"
                    :key="child.title"
                    :to="{name: routeWiki, params: { file: child.path }}">
                  <v-icon>{{ icons.mdiBook }}</v-icon>
                  <v-list-item-title>{{ child.name }}</v-list-item-title>
                </v-list-item>
              </v-list-group>
              <v-list-item
                  v-else
                  :to="{name: routeWiki, params: { file: fileItem.path }}">
                <v-list-item-icon>
                  <v-icon>{{ icons.mdiBook }}</v-icon>
                </v-list-item-icon>
                <v-list-item-title>
                  {{ fileItem.name }}
                </v-list-item-title>
              </v-list-item>
            </div>
          </v-list>
          <v-skeleton-loader
              v-else
              type="list-item-three-line,list-item-three-line,list-item-three-line"/>
        </v-navigation-drawer>
      </v-col>
      <v-col class="fill-height">
        <transition name="component-fade" mode="out-in">
          <v-card v-if="showRootPage" flat>
            <v-card-text>
              <v-autocomplete
                  v-model="query"
                  :items="namePathPairs"
                  item-text="name"
                  item-value="path"
                  label="Search for wiki pages"
                  outlined
                  @change="navigateToWikiPage"
                  @keydown.enter="navigateToWikiPage"
                  @click:append-outer="navigateToWikiPage"
                  :append-outer-icon="icons.mdiMagnify">
              </v-autocomplete>
            </v-card-text>
          </v-card>
          <template v-else>
            <div v-if="fileHTML" v-html="fileHTML"></div>
            <v-skeleton-loader v-else type="article"/>
          </template>
        </transition>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { mdiBookOpen, mdiBook, mdiMagnify } from '@mdi/js';
import {
  PAGE_TITLE, WIKIAPI_PATH_PREFIX, ROUTE_NAME_WIKI, ROUTE_NAME_WIKI_ROOT,
} from '../../constants';
import { doGet } from '../../utilities';

// separateWords adds a space before each capitalized letter
function separateWords(string) {
  return string.replace(/([A-Z])/g, ' $1').trim();
}

// capitalize capitalizes the given string
function capitalize(string) {
  return string[0].toUpperCase() + string.slice(1);
}

// cleanName capitalizes the given string and remove the '.md' postfix
function cleanName(fileName) {
  return capitalize(separateWords(fileName)).replace('.md', '').replace('-', ' ');
}

// getFileHierarchy returns a file hierarchy based on the given directories.
// convert map to array of objects
// result:
// [
//   {
//     "name": "Index",
//     "items": null,
//     "path": "index.md"
//   },
//   {
//     "name": "TransactionTypes",
//     "items": [
//       {
//         "name": "Destination",
//         "path": "transactionTypes/destination.md"
//       },
//       {
//         "name": "Mixing",
//         "path": "transactionTypes/mixing.md"
//       },
//     ],
//     "path": "transactionTypes/destination.md"
//   }
// ]
function getFileHierarchy(fileSet) {
  if (fileSet === null) {
    return [];
  }

  const hierarchy = new Map();

  fileSet.forEach((d) => {
    const pathParts = d.split('/');

    if (pathParts.length > 2) {
      // only a depth of 2 is supported
      return;
    }

    let [directory, fileName] = pathParts;

    directory = cleanName(directory);

    const itemProps = { items: null, path: d };

    if (fileName) {
      fileName = cleanName(fileName);
      let props = itemProps;

      if (hierarchy.has(directory)) {
        props = hierarchy.get(directory);
      }

      if (!props.items) props.items = [];

      props.items.push({ name: fileName, path: d });
      hierarchy.set(directory, props);
    } else hierarchy.set(directory, itemProps);
  });

  const hierarchyArray = [];
  hierarchy.forEach((props, directory) => {
    const item = { name: directory, ...props };
    hierarchyArray.push(item);
  });

  return hierarchyArray;
}

export default {
  name: 'Wiki',
  data() {
    return {
      icons: {
        mdiBookOpen, mdiBook, mdiMagnify,
      },
      routeWiki: ROUTE_NAME_WIKI,
      routeWikiRoot: ROUTE_NAME_WIKI_ROOT,
      fileHTML: '',
      // fileSet will hold a set with all possible file paths
      fileSet: null,
      // isRoot determines of the root page of the wiki is shown
      showRootPage: true,
      query: null,
    };
  },
  computed: {
    fileHierarchy() {
      return getFileHierarchy(this.fileSet);
    },
    // namePathPairs returns an array of name and
    // path pairs [{name: filename, path: filename.md}, ...]
    namePathPairs() {
      const pairs = [];

      this.fileHierarchy.forEach((d) => {
        if (d.items) {
          d.items.forEach((l) => {
            pairs.push(l);
          });
        } else {
          pairs.push(d);
        }
      });

      return pairs;
    },
  },
  methods: {
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    getFileIndex() {
      return doGet(`${WIKIAPI_PATH_PREFIX}/index/`, this.$router, this.$store)
        .then((d) => {
          if (!d.success) return;
          if (d.data && d.data.index) this.fileSet = new Set(d.data.index);
        })
        .catch((e) => this.setErrorMessage(e));
    },
    getFile(filePath) {
      // only try to get file if it is in list of known files
      if (this.fileSet === null || !this.fileSet.has(filePath)) {
        return;
      }

      doGet(`${WIKIAPI_PATH_PREFIX}/file/${filePath}`, this.$router, this.$store)
        .then((d) => {
          if (!d.success) return;

          if (d.data && d.data.html) this.fileHTML = d.data.html;
        })
        .catch((e) => this.setErrorMessage(e));
    },
    navigateToWikiPage() {
      if (!this.query) return;
      this.$router.push({ name: this.routeWiki, params: { file: this.query } });
    },
  },
  async mounted() {
    document.title = `Wiki - ${PAGE_TITLE}`;

    if (this.$route.params.file) {
      this.showRootPage = false;
    }

    await this.getFileIndex();

    if (!this.showRootPage) {
      this.getFile(this.$route.params.file);
    }
  },
  watch: {
    $route() {
      if (this.$route.params.file) {
        this.showRootPage = false;
        this.getFile(this.$route.params.file);
      } else {
        this.showRootPage = true;
      }
    },
  },
};
</script>

<style scoped>

</style>
